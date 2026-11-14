package ztetelnet

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"billing-backend/internal/domain"
)

// Client defines the contract for communicating with ZTE C300/C320 OLTs via Telnet CLI.
type Client interface {
	TestConnection(ctx context.Context, host string, port int, username, password string) (string, error)
	GetONUs(ctx context.Context, host string, port int, username, password string, board, pon int) ([]domain.ZTEONUInfo, error)
	GetPaginatedONUs(ctx context.Context, host string, port int, username, password string, board, pon int, page, limit int) (*domain.ZTEPaginatedONUs, error)
	GetONUDetail(ctx context.Context, host string, port int, username, password string, board, pon, onuID int) (*domain.ZTEONUDetail, error)
	GetEmptyONUIDs(ctx context.Context, host string, port int, username, password string, board, pon int) ([]int, error)
	GetUplinks(ctx context.Context, host string, port int, username, password string) (*domain.ZTEUplinksData, error)
	ClearCache(host string)
}

type zteTelnetClient struct {
	// hostLocks enforces strict VTY session discipline:
	// AT MOST 1 concurrent Telnet session per physical ZTE OLT host.
	// This prevents exhausting ZTE C320's limited VTY pool (5-8 sessions)
	// and guarantees operational engineers can always access PuTTY/CLI.
	hostLocks sync.Map // map[string]*sync.Mutex

	cacheMu sync.RWMutex
	cache   map[string]cacheItem
}

type cacheItem struct {
	data      []domain.ZTEONUInfo
	expiresAt time.Time
}

// NewClient initializes a new ZTE Telnet OLT client.
func NewClient() Client {
	return &zteTelnetClient{
		cache: make(map[string]cacheItem),
	}
}

func (c *zteTelnetClient) getHostLock(host string) *sync.Mutex {
	actual, _ := c.hostLocks.LoadOrStore(host, &sync.Mutex{})
	return actual.(*sync.Mutex)
}

func (c *zteTelnetClient) ClearCache(host string) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	for k := range c.cache {
		if strings.HasPrefix(k, host) {
			delete(c.cache, k)
		}
	}
}

// TelnetSession wraps a TCP connection with Telnet command handling.
type TelnetSession struct {
	conn   net.Conn
	reader *bufio.Reader
}

func dialTelnet(ctx context.Context, host string, port int, timeout time.Duration) (*TelnetSession, error) {
	if port <= 0 {
		port = 23
	}
	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, fmt.Errorf("gagal terhubung ke ZTE OLT Telnet (%s:%d): %w", host, port, err)
	}

	return &TelnetSession{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}, nil
}

// CloseWithLogout sends exit commands to clean up any open VTY sessions, then closes TCP socket.
func (s *TelnetSession) CloseWithLogout() error {
	if s.conn == nil {
		return nil
	}
	_ = s.conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
	_, _ = s.conn.Write([]byte("exit\r\n"))
	time.Sleep(50 * time.Millisecond)
	_, _ = s.conn.Write([]byte("exit\r\n"))
	return s.conn.Close()
}

func (s *TelnetSession) readUntil(timeout time.Duration, delimiters ...string) (string, error) {
	_ = s.conn.SetReadDeadline(time.Now().Add(timeout))
	var buf bytes.Buffer
	b := make([]byte, 1)

	for {
		n, err := s.reader.Read(b)
		if err != nil {
			return buf.String(), err
		}
		if n == 0 {
			continue
		}

		// Handle Telnet IAC negotiation (0xFF)
		if b[0] == 0xFF {
			cmdByte, err := s.reader.ReadByte()
			if err != nil {
				return buf.String(), err
			}
			optByte, err := s.reader.ReadByte()
			if err != nil {
				return buf.String(), err
			}
			if cmdByte == 253 { // DO -> WONT
				_, _ = s.conn.Write([]byte{0xFF, 252, optByte})
			} else if cmdByte == 251 { // WILL -> DONT
				_, _ = s.conn.Write([]byte{0xFF, 254, optByte})
			}
			continue
		}

		buf.WriteByte(b[0])
		current := buf.String()

		for _, delim := range delimiters {
			if strings.Contains(strings.ToLower(current), strings.ToLower(delim)) {
				return current, nil
			}
		}
	}
}

func (s *TelnetSession) writeLine(line string) error {
	_ = s.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := s.conn.Write([]byte(line + "\r\n"))
	return err
}

func (s *TelnetSession) login(username, password string) error {
	// 1. Wait for username / login prompt
	out, err := s.readUntil(8*time.Second, "username:", "login:", "user:")
	if err != nil {
		return fmt.Errorf("timeout menunggu prompt username ZTE: %w (output: %s)", err, out)
	}

	// 2. Send username
	if err := s.writeLine(username); err != nil {
		return fmt.Errorf("gagal mengirim username: %w", err)
	}

	// 3. Wait for password prompt
	out, err = s.readUntil(8*time.Second, "password:")
	if err != nil {
		return fmt.Errorf("timeout menunggu prompt password ZTE: %w (output: %s)", err, out)
	}

	// 4. Send password
	if err := s.writeLine(password); err != nil {
		return fmt.Errorf("gagal mengirim password: %w", err)
	}

	// 5. Wait for shell prompt (ZXAN> or ZXAN#)
	out, err = s.readUntil(8*time.Second, ">", "#")
	if err != nil {
		return fmt.Errorf("gagal login ke ZTE OLT: %w (output: %s)", err, out)
	}

	// 6. If in user EXEC mode (>), enter privileged EXEC mode (enable)
	if strings.Contains(out, ">") {
		if err := s.writeLine("enable"); err != nil {
			return err
		}
		res, err := s.readUntil(5*time.Second, "password:", "#")
		if err == nil && strings.Contains(strings.ToLower(res), "password:") {
			_ = s.writeLine(password)
			_, _ = s.readUntil(5*time.Second, "#")
		}
	}

	// 7. CRITICAL: disable pagination so output doesn't halt at --More--
	_ = s.writeLine("terminal length 0")
	_, _ = s.readUntil(4*time.Second, "#")

	return nil
}

func (s *TelnetSession) executeCommand(cmd string, timeout time.Duration) (string, error) {
	if err := s.writeLine(cmd); err != nil {
		return "", fmt.Errorf("gagal mengirim command '%s': %w", cmd, err)
	}
	out, err := s.readUntil(timeout, "#")
	if err != nil {
		return out, fmt.Errorf("timeout membaca output command '%s': %w", cmd, err)
	}
	return out, nil
}

// TestConnection checks Telnet accessibility and returns basic card/version info.
func (c *zteTelnetClient) TestConnection(ctx context.Context, host string, port int, username, password string) (string, error) {
	lock := c.getHostLock(host)
	lock.Lock()
	defer lock.Unlock()

	session, err := dialTelnet(ctx, host, port, 6*time.Second)
	if err != nil {
		return "", err
	}
	defer session.CloseWithLogout()

	if err := session.login(username, password); err != nil {
		return "", err
	}

	// 1. Try unambiguous command: "show card" (standard on all ZTE C300/C320)
	out, err := session.executeCommand("show card", 6*time.Second)
	if err != nil || strings.Contains(out, "%Error") {
		// 2. Fallback to "show version-running" (never use ambiguous "show version")
		out, err = session.executeCommand("show version-running", 6*time.Second)
		if err != nil || strings.Contains(out, "%Error") {
			// 3. Fallback to basic clock check
			out, _ = session.executeCommand("show clock", 4*time.Second)
		}
	}

	// Extract meaningful summary from output
	lines := strings.Split(out, "\n")
	var activeCards []string
	var versionStr string

	for _, l := range lines {
		lTrim := strings.TrimSpace(l)
		if lTrim == "" || strings.HasPrefix(lTrim, "%Error") || strings.HasPrefix(lTrim, "Rack") || strings.HasPrefix(lTrim, "---") {
			continue
		}
		// Match card line: e.g. "1 1 1 GTGH GTGHG 16 120400 V1.2.5P3 INSERVICE"
		if strings.Contains(lTrim, "INSERVICE") {
			parts := strings.Fields(lTrim)
			if len(parts) >= 4 {
				slot := parts[2]
				cardType := parts[3]
				activeCards = append(activeCards, fmt.Sprintf("Slot %s: %s", slot, cardType))
			}
		} else if strings.Contains(strings.ToLower(lTrim), "zte") || strings.Contains(strings.ToLower(lTrim), "zxros") {
			if versionStr == "" {
				versionStr = lTrim
			}
		}
	}

	var summary string
	if len(activeCards) > 0 {
		summary = fmt.Sprintf("Card Terdeteksi (%s)", strings.Join(activeCards, ", "))
	} else if versionStr != "" {
		summary = versionStr
	} else {
		summary = "ZTE C320 Control System Siap"
	}

	return fmt.Sprintf("Berhasil terhubung ke ZTE OLT via Telnet (%s:%d). %s", host, port, summary), nil
}

// GetUplinks retrieves detected line cards and uplink ports for ZTE C320.
func (c *zteTelnetClient) GetUplinks(ctx context.Context, host string, port int, username, password string) (*domain.ZTEUplinksData, error) {
	lock := c.getHostLock(host)
	lock.Lock()
	defer lock.Unlock()

	session, err := dialTelnet(ctx, host, port, 6*time.Second)
	if err != nil {
		return nil, err
	}
	defer session.CloseWithLogout()

	if err := session.login(username, password); err != nil {
		return nil, err
	}

	outCards, _ := session.executeCommand("show card", 8*time.Second)
	outPorts, _ := session.executeCommand("show interface brief", 8*time.Second)

	cards := parseZTECardsOutput(outCards)
	ports := parseZTEPortsOutput(outPorts)

	return &domain.ZTEUplinksData{
		Cards: cards,
		Ports: ports,
	}, nil
}

// GetONUs retrieves ONUs on a specified board and PON port.
func (c *zteTelnetClient) GetONUs(ctx context.Context, host string, port int, username, password string, board, pon int) ([]domain.ZTEONUInfo, error) {
	cacheKey := fmt.Sprintf("%s:%d:%d", host, board, pon)

	c.cacheMu.RLock()
	item, found := c.cache[cacheKey]
	c.cacheMu.RUnlock()
	if found && time.Now().Before(item.expiresAt) {
		return item.data, nil
	}

	lock := c.getHostLock(host)
	lock.Lock()
	defer lock.Unlock()

	// Double check cache after acquiring lock
	c.cacheMu.RLock()
	item, found = c.cache[cacheKey]
	c.cacheMu.RUnlock()
	if found && time.Now().Before(item.expiresAt) {
		return item.data, nil
	}

	session, err := dialTelnet(ctx, host, port, 6*time.Second)
	if err != nil {
		return nil, err
	}
	defer session.CloseWithLogout()

	if err := session.login(username, password); err != nil {
		return nil, err
	}

	// 1. Fetch ONU state on specified PON (e.g. gpon-olt_1/1/1)
	cmdState := fmt.Sprintf("show gpon onu state gpon-olt_1/%d/%d", board, pon)
	outState, err := session.executeCommand(cmdState, 8*time.Second)
	if err != nil || strings.Contains(outState, "%Error") || strings.Contains(outState, "Unknown") || !strings.Contains(outState, ":") {
		// Fallback to global show gpon onu state (ZTE C320 CLI output: 1/1/1:1 enable enable working 1(GPON))
		outState, _ = session.executeCommand("show gpon onu state", 10*time.Second)
	}

	// 2. Fetch ONU running configuration or baseinfo for names and serial numbers
	cmdConfig := fmt.Sprintf("show running-config interface gpon-olt_1/%d/%d", board, pon)
	outConfig, _ := session.executeCommand(cmdConfig, 8*time.Second)

	cmdBase := fmt.Sprintf("show gpon onu baseinfo gpon-olt_1/%d/%d", board, pon)
	outBase, _ := session.executeCommand(cmdBase, 8*time.Second)

	// 3. Optional batch optical power for the entire PON port
	cmdRx := fmt.Sprintf("show pon power onu-rx gpon-olt_1/%d/%d", board, pon)
	outRx, _ := session.executeCommand(cmdRx, 6*time.Second)

	onus := parseZTEONUs(board, pon, outState, outConfig, outBase, outRx)

	// Cache results for 15 seconds
	c.cacheMu.Lock()
	c.cache[cacheKey] = cacheItem{
		data:      onus,
		expiresAt: time.Now().Add(15 * time.Second),
	}
	c.cacheMu.Unlock()

	return onus, nil
}

// GetPaginatedONUs returns a paginated slice of ONUs.
func (c *zteTelnetClient) GetPaginatedONUs(ctx context.Context, host string, port int, username, password string, board, pon int, page, limit int) (*domain.ZTEPaginatedONUs, error) {
	allOnus, err := c.GetONUs(ctx, host, port, username, password, board, pon)
	if err != nil {
		return nil, err
	}

	total := len(allOnus)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	start := (page - 1) * limit
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	totalPages := 0
	if limit > 0 {
		totalPages = (total + limit - 1) / limit
	}

	return &domain.ZTEPaginatedONUs{
		Data: allOnus[start:end],
		Meta: domain.ZTEPaginationMeta{
			Page:      page,
			Limit:     limit,
			PageCount: totalPages,
			TotalRows: total,
		},
	}, nil
}

// GetONUDetail retrieves detailed optical power and distance telemetry for a single ONU.
func (c *zteTelnetClient) GetONUDetail(ctx context.Context, host string, port int, username, password string, board, pon, onuID int) (*domain.ZTEONUDetail, error) {
	lock := c.getHostLock(host)
	lock.Lock()
	defer lock.Unlock()

	session, err := dialTelnet(ctx, host, port, 6*time.Second)
	if err != nil {
		return nil, err
	}
	defer session.CloseWithLogout()

	if err := session.login(username, password); err != nil {
		return nil, err
	}

	onuTarget := fmt.Sprintf("gpon-onu_1/%d/%d:%d", board, pon, onuID)
	altTarget := fmt.Sprintf("1/%d/%d:%d", board, pon, onuID)

	// 1. Optical attenuation (Rx/Tx Power)
	cmdPower := fmt.Sprintf("show pon power attenuation %s", onuTarget)
	outPower, _ := session.executeCommand(cmdPower, 6*time.Second)
	if strings.Contains(outPower, "%Error") || strings.Contains(outPower, "Unrecognized") {
		outPower, _ = session.executeCommand(fmt.Sprintf("show pon power attenuation %s", altTarget), 6*time.Second)
	}

	// 1b. Fallback optical power if attenuation had error or missing Rx
	var outRxSingle string
	if strings.Contains(outPower, "%Error") || !strings.Contains(strings.ToLower(outPower), "rx") {
		cmdRxSingle := fmt.Sprintf("show pon power onu-rx %s", onuTarget)
		outRxSingle, _ = session.executeCommand(cmdRxSingle, 5*time.Second)
		if strings.Contains(outRxSingle, "%Error") {
			outRxSingle, _ = session.executeCommand(fmt.Sprintf("show pon power onu-rx %s", altTarget), 5*time.Second)
		}
	}

	// 2. Optical distance
	cmdDist := fmt.Sprintf("show gpon onu distance %s", onuTarget)
	outDist, _ := session.executeCommand(cmdDist, 6*time.Second)
	if strings.Contains(outDist, "%Error") || strings.Contains(outDist, "Unrecognized") {
		outDist, _ = session.executeCommand(fmt.Sprintf("show gpon onu distance %s", altTarget), 6*time.Second)
	}

	// 3. ONU detail info & state
	cmdDetail := fmt.Sprintf("show gpon onu detail-info %s", onuTarget)
	outDetail, _ := session.executeCommand(cmdDetail, 6*time.Second)
	if strings.Contains(outDetail, "%Error") || strings.Contains(outDetail, "Unrecognized") {
		outDetail, _ = session.executeCommand(fmt.Sprintf("show gpon onu detail-info %s", altTarget), 6*time.Second)
	}

	detail := parseZTEONUDetailOutput(board, pon, onuID, outPower, outDist, outDetail, outRxSingle)
	return detail, nil
}

// GetEmptyONUIDs calculates unallocated ONU IDs (1 to 128) on a PON port.
func (c *zteTelnetClient) GetEmptyONUIDs(ctx context.Context, host string, port int, username, password string, board, pon int) ([]int, error) {
	onus, err := c.GetONUs(ctx, host, port, username, password, board, pon)
	if err != nil {
		return nil, err
	}

	used := make(map[int]bool)
	for _, o := range onus {
		used[o.ONUID] = true
	}

	var emptyIDs []int
	for id := 1; id <= 128; id++ {
		if !used[id] {
			emptyIDs = append(emptyIDs, id)
		}
	}
	return emptyIDs, nil
}

// ===== PARSERS =====

func parseZTECardsOutput(output string) []domain.ZTECardInfo {
	var cards []domain.ZTECardInfo
	lines := strings.Split(output, "\n")
	// Header: Rack Shelf Slot CfgType RealType Port HardVer SoftVer Status
	// Row:    1    1     1    GTGH    GTGHG    16   120400  V1.2.5P3 INSERVICE
	re := regexp.MustCompile(`^\s*(\d+)\s+(\d+)\s+(\d+)\s+([A-Za-z0-9_-]+)\s+([A-Za-z0-9_-]+)\s+(\d+)`)
	for _, line := range lines {
		m := re.FindStringSubmatch(strings.TrimSpace(line))
		if len(m) >= 7 {
			slot, _ := strconv.Atoi(m[3])
			cardType := m[4]
			role := "Line Card"
			if strings.Contains(strings.ToUpper(cardType), "SMX") || strings.Contains(strings.ToUpper(cardType), "SCX") {
				role = "Control & Switch Card"
			}
			cards = append(cards, domain.ZTECardInfo{
				Slot: slot,
				Type: cardType,
				Role: role,
			})
		}
	}
	return cards
}

func parseZTEPortsOutput(output string) []domain.ZTEPortInfo {
	var ports []domain.ZTEPortInfo
	lines := strings.Split(output, "\n")
	// Match port patterns like gei_1/3/1, xgei_1/4/1, etc.
	re := regexp.MustCompile(`([a-z0-9_]+_1/(\d+)/(\d+))\s+([a-z]+)\s+([a-z]+)`)
	for _, line := range lines {
		m := re.FindStringSubmatch(strings.ToLower(strings.TrimSpace(line)))
		if len(m) >= 6 {
			slot, _ := strconv.Atoi(m[2])
			portNum, _ := strconv.Atoi(m[3])
			admin := m[4]
			oper := m[5]
			kind := "GE"
			speed := 1000
			if strings.Contains(m[1], "xgei") {
				kind = "10GE"
				speed = 10000
			}
			ports = append(ports, domain.ZTEPortInfo{
				Name:        m[1],
				Shelf:       1,
				Slot:        slot,
				Port:        portNum,
				Kind:        kind,
				AdminStatus: admin,
				OperStatus:  oper,
				SpeedMbps:   speed,
			})
		}
	}
	return ports
}

func parseZTEONUs(board, pon int, outState, outConfig, outBase string, outRx ...string) []domain.ZTEONUInfo {
	onusMap := make(map[int]*domain.ZTEONUInfo)

	// 1. Parse show gpon onu state:
	// Output formats:
	//   1/1/1:1     enable       enable      working      1(GPON)
	//   1/1/10:5    enable       enable      working      1(GPON)
	//   gpon-onu_1/1/1:1   enable  enable  working  1
	reState := regexp.MustCompile(`(?:gpon-onu_)?(?:(\d+)/)?(?:(\d+)/)?(\d+):(\d+)\s+(\S+)\s+(\S+)\s+(\S+)`)
	for _, line := range strings.Split(outState, "\n") {
		m := reState.FindStringSubmatch(strings.TrimSpace(line))
		if len(m) >= 8 {
			lineBoard, _ := strconv.Atoi(m[2])
			linePon, _ := strconv.Atoi(m[3])
			// If lineBoard is present and board > 0, filter by board
			if board > 0 && lineBoard > 0 && lineBoard != board {
				continue
			}
			// If linePon is present and pon > 0, filter by pon
			if pon > 0 && linePon > 0 && linePon != pon {
				continue
			}
			id, _ := strconv.Atoi(m[4])
			phaseState := strings.ToLower(m[7])
			status := "Offline"
			if phaseState == "working" || phaseState == "online" {
				status = "Online"
			}
			onusMap[id] = &domain.ZTEONUInfo{
				Board:   board,
				PON:     pon,
				ONUID:   id,
				Status:  status,
				ONUType: "GPON",
			}
		}
	}

	// 2. Parse show running-config interface gpon-olt_1/x/y:
	// Matches:
	//   onu 1 type ZTE-F660 sn ZTEGC1234567
	//   onu 1 name CustomerName
	reTypeSN := regexp.MustCompile(`onu\s+(\d+)\s+type\s+([^\s]+)\s+sn\s+([^\s]+)`)
	reName := regexp.MustCompile(`onu\s+(\d+)\s+name\s+([^\r\n]+)`)

	for _, line := range strings.Split(outConfig, "\n") {
		line = strings.TrimSpace(line)
		if m := reTypeSN.FindStringSubmatch(line); len(m) >= 4 {
			id, _ := strconv.Atoi(m[1])
			if onu, ok := onusMap[id]; ok {
				onu.ONUType = m[2]
				onu.SerialNumber = m[3]
			} else {
				onusMap[id] = &domain.ZTEONUInfo{
					Board:        board,
					PON:          pon,
					ONUID:        id,
					Status:       "Offline",
					ONUType:      m[2],
					SerialNumber: m[3],
				}
			}
		}
		if m := reName.FindStringSubmatch(line); len(m) >= 3 {
			id, _ := strconv.Atoi(m[1])
			if onu, ok := onusMap[id]; ok {
				onu.Name = strings.TrimSpace(m[2])
			}
		}
	}

	// 3. Fallback parse show gpon onu baseinfo if serial number is missing:
	// Format: 1/1/1:1  ZTEG-F660  ZTEGC1234567  ready
	// or:     gpon-onu_1/1/1:1  ZTEG-F660  ZTEGC1234567  ready
	reBase := regexp.MustCompile(`(?:gpon-onu_)?(?:(?:\d+)/)*(\d+):(\d+)\s+([^\s]+)\s+([^\s]+)`)
	for _, line := range strings.Split(outBase, "\n") {
		m := reBase.FindStringSubmatch(strings.TrimSpace(line))
		if len(m) >= 5 {
			id, _ := strconv.Atoi(m[2])
			if onu, ok := onusMap[id]; ok {
				if onu.ONUType == "" || onu.ONUType == "GPON" {
					onu.ONUType = m[3]
				}
				if onu.SerialNumber == "" {
					onu.SerialNumber = m[4]
				}
			}
		}
	}

	// 4. Batch parse Rx Power if available:
	if len(outRx) > 0 && outRx[0] != "" && !strings.Contains(outRx[0], "%Error") {
		reRxLine := regexp.MustCompile(`(?:gpon-onu_)?(?:(?:\d+)/)*(\d+):(\d+)\s+([-\d.]+)`)
		for _, line := range strings.Split(outRx[0], "\n") {
			m := reRxLine.FindStringSubmatch(strings.TrimSpace(line))
			if len(m) >= 4 {
				id, _ := strconv.Atoi(m[2])
				if onu, ok := onusMap[id]; ok {
					if val, err := strconv.ParseFloat(m[3], 64); err == nil {
						onu.RxPower = fmt.Sprintf("%.2f dBm", val)
					}
				}
			}
		}
	}

	// Flatten map to sorted slice
	var result []domain.ZTEONUInfo
	for id := 1; id <= 128; id++ {
		if onu, ok := onusMap[id]; ok {
			result = append(result, *onu)
		}
	}
	return result
}

func parseZTEONUDetailOutput(board, pon, onuID int, outPower, outDist, outDetail string, outRxSingle ...string) *domain.ZTEONUDetail {
	detail := &domain.ZTEONUDetail{
		Board:         board,
		PON:           pon,
		ONUID:         onuID,
		Status:        "Offline",
		OfflineReason: "Normal / Tidak Ada Error",
	}

	// 1. Parse optical power attenuation:
	// Format 1 (Standard ZTE C300/C320 attenuation output):
	//            OLT               ONU             Attenuation
	// --------------------------------------------------------------------------
	// up         Rx :-27.596(dbm)   Tx:2.362(dbm)   29.958(dB)
	// down       Tx :6.895(dbm)    Rx:-32.216(dbm)  39.111(dB)
	//
	// Format 2 (Tabular attenuation):
	//            OLT Rx(dBm)        ONU Rx(dBm)
	// up         -21.432            -
	// down       -                  -22.150

	// Check downstream line for ONU Rx power:
	reDownLine := regexp.MustCompile(`(?i)\bdown\b[^\r\n]+`)
	if mDown := reDownLine.FindString(outPower); mDown != "" {
		// Look for Rx on down line: e.g. Rx:-32.216(dbm) or Rx :-32.216(dbm) or Rx : -32.216
		reRx := regexp.MustCompile(`(?i)Rx\s*:\s*([-\d.]+)`)
		if m := reRx.FindStringSubmatch(mDown); len(m) >= 2 {
			if val, err := strconv.ParseFloat(m[1], 64); err == nil {
				detail.RxPower = fmt.Sprintf("%.2f dBm", val)
			} else {
				detail.RxPower = m[1] + " dBm"
			}
		} else {
			// Tabular: down - -22.150
			reTable := regexp.MustCompile(`(?i)down\s+([-\d.]+|-)\s+([-\d.]+)`)
			if mTable := reTable.FindStringSubmatch(mDown); len(mTable) >= 3 && mTable[2] != "-" {
				if val, err := strconv.ParseFloat(mTable[2], 64); err == nil {
					detail.RxPower = fmt.Sprintf("%.2f dBm", val)
				} else {
					detail.RxPower = mTable[2] + " dBm"
				}
			}
		}
	}

	// Fallback to single Rx power command if empty: e.g. "gpon-onu_1/1/2:1 -18.060(dbm)" or "-18.060"
	if detail.RxPower == "" && len(outRxSingle) > 0 && outRxSingle[0] != "" {
		reRxSingle := regexp.MustCompile(`([-\d.]+)\s*(?:\(dbm\)|dbm)?`)
		if m := reRxSingle.FindStringSubmatch(outRxSingle[0]); len(m) >= 2 {
			if val, err := strconv.ParseFloat(m[1], 64); err == nil {
				detail.RxPower = fmt.Sprintf("%.2f dBm", val)
			}
		}
	}

	// Final fallback for Rx power anywhere in outPower
	if detail.RxPower == "" {
		reRxAny := regexp.MustCompile(`(?i)(?:onu\s*rx|rx\s*power)[^\r\n:]*:\s*([-\d.]+)`)
		if m := reRxAny.FindStringSubmatch(outPower); len(m) >= 2 {
			if val, err := strconv.ParseFloat(m[1], 64); err == nil {
				detail.RxPower = fmt.Sprintf("%.2f dBm", val)
			} else {
				detail.RxPower = m[1] + " dBm"
			}
		}
	}

	// Check upstream line for ONU Tx power:
	// up Rx :-27.596(dbm) Tx:2.362(dbm) 29.958(dB)
	reUpLine := regexp.MustCompile(`(?i)\bup\b[^\r\n]+`)
	if mUp := reUpLine.FindString(outPower); mUp != "" {
		reTx := regexp.MustCompile(`(?i)Tx\s*:\s*([-\d.]+)`)
		if m := reTx.FindStringSubmatch(mUp); len(m) >= 2 {
			if val, err := strconv.ParseFloat(m[1], 64); err == nil {
				detail.TxPower = fmt.Sprintf("%.2f dBm", val)
			} else {
				detail.TxPower = m[1] + " dBm"
			}
		} else {
			// Tabular: up -21.432 -
			reTable := regexp.MustCompile(`(?i)up\s+([-\d.]+)\s+([-\d.]+|-)`)
			if mTable := reTable.FindStringSubmatch(mUp); len(mTable) >= 3 && mTable[1] != "-" {
				if val, err := strconv.ParseFloat(mTable[1], 64); err == nil {
					detail.TxPower = fmt.Sprintf("%.2f dBm", val)
				}
			}
		}
	}

	// 2. Parse optical distance (from show gpon onu distance or detail-info):
	// Formats:
	//   Distance: 1420m or 1420 m
	//   ONU Distance: 2450m
	//   Distance : 2450(m)
	reDist := regexp.MustCompile(`(?i)(?:onu\s*)?dist(?:ance)?[^\d]*(\d+)\s*(?:m|\(m\)|meter)?`)
	if m := reDist.FindStringSubmatch(outDist); len(m) >= 2 && m[1] != "" {
		detail.GponOpticalDistance = m[1]
	}
	if detail.GponOpticalDistance == "" {
		if m := reDist.FindStringSubmatch(outDetail); len(m) >= 2 && m[1] != "" {
			detail.GponOpticalDistance = m[1]
		}
	}

	// 3. Parse Uptime / Online duration from detail-info:
	// Format: Online Duration: 5d 12h 30m 15s
	reUptime := regexp.MustCompile(`(?i)(?:online\s*duration|uptime)[^\r\n:]*:\s*([^\r\n]+)`)
	if m := reUptime.FindStringSubmatch(outDetail); len(m) >= 2 {
		detail.Uptime = strings.TrimSpace(m[1])
	}

	// 4. Parse detail info: Name, SN, Model, State
	for _, line := range strings.Split(outDetail, "\n") {
		l := strings.TrimSpace(line)
		lLow := strings.ToLower(l)
		if strings.HasPrefix(lLow, "name:") {
			detail.Name = strings.TrimSpace(strings.TrimPrefix(l, "Name:"))
		} else if strings.HasPrefix(lLow, "sn:") || strings.HasPrefix(lLow, "serial number:") {
			parts := strings.SplitN(l, ":", 2)
			if len(parts) >= 2 {
				detail.SerialNumber = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(lLow, "type:") || strings.HasPrefix(lLow, "model:") {
			parts := strings.SplitN(l, ":", 2)
			if len(parts) >= 2 {
				detail.ONUType = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(lLow, "phase state:") || strings.HasPrefix(lLow, "state:") {
			parts := strings.SplitN(l, ":", 2)
			if len(parts) >= 2 {
				st := strings.ToLower(strings.TrimSpace(parts[1]))
				if st == "working" || st == "online" || st == "ready" {
					detail.Status = "Online"
					detail.OfflineReason = "Normal / Tidak Ada Error"
				} else {
					detail.Status = "Offline"
					if strings.Contains(st, "los") {
						detail.OfflineReason = "LOS (Kabel Optik Putus / Signal Loss)"
					} else if strings.Contains(st, "dyinggasp") {
						detail.OfflineReason = "Dying Gasp (Modem Mati Listrik / Power Off)"
					} else {
						detail.OfflineReason = "State: " + st
					}
				}
			}
		}
	}

	// If history section has cause for offline ONU:
	// 1    2026-05-10 08:12:00   2026-05-12 14:20:11  DyingGasp
	if detail.Status == "Offline" {
		if strings.Contains(outDetail, "DyingGasp") {
			detail.OfflineReason = "Dying Gasp (Modem Mati Listrik / Power Off)"
		} else if strings.Contains(outDetail, "LOS") {
			detail.OfflineReason = "LOS (Kabel Optik Putus / Signal Loss)"
		}
	}

	return detail
}
