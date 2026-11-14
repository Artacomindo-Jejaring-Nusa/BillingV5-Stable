package hsgqclient

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

// Client defines the contract for communicating with HSGQ EPON/GPON OLTs.
type Client interface {
	TestConnection(ctx context.Context, host string, port int, username, password string) (string, error)
	GetONUs(ctx context.Context, host string, port int, username, password string, ponFilter int) ([]domain.ZTEONUInfo, error)
	GetPaginatedONUs(ctx context.Context, host string, port int, username, password string, ponFilter int, page, limit int) (*domain.ZTEPaginatedONUs, error)
	GetONUDetail(ctx context.Context, host string, port int, username, password string, pon int, onuID int) (*domain.ZTEONUDetail, error)
	GetEmptyONUIDs(ctx context.Context, host string, port int, username, password string, pon int) ([]int, error)
	GetUplinks(ctx context.Context, host string, port int, username, password string) (*domain.ZTEUplinksData, error)
	ClearCache(host string)
}

type hsgqClient struct {
	cacheMu sync.RWMutex
	cache   map[string]cacheItem
}

type cacheItem struct {
	data      []domain.ZTEONUInfo
	expiresAt time.Time
}

// NewClient initializes a new HSGQ OLT client.
func NewClient() Client {
	return &hsgqClient{
		cache: make(map[string]cacheItem),
	}
}

func (c *hsgqClient) ClearCache(host string) {
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
		return nil, fmt.Errorf("gagal terhubung ke OLT Telnet (%s:%d): %w", host, port, err)
	}

	return &TelnetSession{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}, nil
}

func (s *TelnetSession) Close() error {
	return s.conn.Close()
}

// readUntil reads data until one of the expected prompt substrings is encountered.
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
			// Respond to DO (253) with WONT (252), and WILL (251) with DONT (254)
			if cmdByte == 253 {
				_, _ = s.conn.Write([]byte{0xFF, 252, optByte})
			} else if cmdByte == 251 {
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
	// 1. Wait for username/login prompt
	out, err := s.readUntil(8*time.Second, "username:", "login:", "user:")
	if err != nil {
		return fmt.Errorf("timeout menunggu prompt username OLT: %w (output: %s)", err, out)
	}

	// 2. Send username
	if err := s.writeLine(username); err != nil {
		return fmt.Errorf("gagal mengirim username: %w", err)
	}

	// 3. Wait for password prompt
	out, err = s.readUntil(8*time.Second, "password:")
	if err != nil {
		return fmt.Errorf("timeout menunggu prompt password OLT: %w (output: %s)", err, out)
	}

	// 4. Send password
	if err := s.writeLine(password); err != nil {
		return fmt.Errorf("gagal mengirim password: %w", err)
	}

	// 5. Wait for prompt (> or #)
	out, err = s.readUntil(8*time.Second, ">", "#")
	if err != nil {
		return fmt.Errorf("login gagal atau timeout: %w (output: %s)", err, out)
	}

	// 6. If in user EXEC mode (ends with >), elevate to enable mode
	if strings.HasSuffix(strings.TrimSpace(out), ">") {
		if err := s.writeLine("enable"); err != nil {
			return err
		}
		elevateOut, err := s.readUntil(5*time.Second, "password:", "#")
		if err == nil && strings.Contains(strings.ToLower(elevateOut), "password:") {
			_ = s.writeLine(password)
			_, _ = s.readUntil(5*time.Second, "#")
		}
	}

	// 7. Enter config mode
	_ = s.writeLine("config")
	_, _ = s.readUntil(5*time.Second, "(config)#", "#")

	return nil
}

func (s *TelnetSession) execCommand(cmd string) (string, error) {
	if err := s.writeLine(cmd); err != nil {
		return "", err
	}
	return s.readUntil(15*time.Second, "(config)#", "#")
}

func (c *hsgqClient) TestConnection(ctx context.Context, host string, port int, username, password string) (string, error) {
	sess, err := dialTelnet(ctx, host, port, 10*time.Second)
	if err != nil {
		return "", err
	}
	defer sess.Close()

	if err := sess.login(username, password); err != nil {
		return "", fmt.Errorf("autentikasi OLT HSGQ gagal: %w", err)
	}

	out, err := sess.execCommand("show onu-info all")
	if err != nil {
		return "Berhasil terhubung ke OLT HSGQ, namun pembacaan data awal timeout", nil
	}

	onus := parseONUInfoOutput(out)
	onlineCount := 0
	offlineCount := 0
	for _, onu := range onus {
		if strings.EqualFold(onu.Status, "Online") {
			onlineCount++
		} else {
			offlineCount++
		}
	}

	return fmt.Sprintf("Berhasil terhubung ke OLT HSGQ di %s:%d. Terdeteksi %d total ONU (%d Online, %d Offline)",
		host, port, len(onus), onlineCount, offlineCount), nil
}

func (c *hsgqClient) GetONUs(ctx context.Context, host string, port int, username, password string, ponFilter int) ([]domain.ZTEONUInfo, error) {
	cacheKey := fmt.Sprintf("%s:%d", host, port)

	c.cacheMu.RLock()
	item, found := c.cache[cacheKey]
	c.cacheMu.RUnlock()

	var allONUs []domain.ZTEONUInfo
	if found && time.Now().Before(item.expiresAt) {
		allONUs = item.data
	} else {
		sess, err := dialTelnet(ctx, host, port, 10*time.Second)
		if err != nil {
			return nil, err
		}
		defer sess.Close()

		if err := sess.login(username, password); err != nil {
			return nil, err
		}

		out, err := sess.execCommand("show onu-info all")
		if err != nil {
			return nil, fmt.Errorf("gagal mengeksekusi show onu-info all: %w", err)
		}

		allONUs = parseONUInfoOutput(out)

		// Cache for 20 seconds
		c.cacheMu.Lock()
		c.cache[cacheKey] = cacheItem{
			data:      allONUs,
			expiresAt: time.Now().Add(20 * time.Second),
		}
		c.cacheMu.Unlock()
	}

	if ponFilter <= 0 {
		return allONUs, nil
	}

	var filtered []domain.ZTEONUInfo
	for _, o := range allONUs {
		if o.PON == ponFilter {
			filtered = append(filtered, o)
		}
	}
	return filtered, nil
}

func (c *hsgqClient) GetPaginatedONUs(ctx context.Context, host string, port int, username, password string, ponFilter int, page, limit int) (*domain.ZTEPaginatedONUs, error) {
	onus, err := c.GetONUs(ctx, host, port, username, password, ponFilter)
	if err != nil {
		return nil, err
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	total := len(onus)
	pageCount := (total + limit - 1) / limit

	start := (page - 1) * limit
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	return &domain.ZTEPaginatedONUs{
		Data: onus[start:end],
		Meta: domain.ZTEPaginationMeta{
			Page:      page,
			Limit:     limit,
			PageCount: pageCount,
			TotalRows: total,
		},
	}, nil
}

func (c *hsgqClient) GetONUDetail(ctx context.Context, host string, port int, username, password string, pon int, onuID int) (*domain.ZTEONUDetail, error) {
	sess, err := dialTelnet(ctx, host, port, 10*time.Second)
	if err != nil {
		return nil, err
	}
	defer sess.Close()

	if err := sess.login(username, password); err != nil {
		return nil, err
	}

	cmd := fmt.Sprintf("show optical-rssi %d %d", pon, onuID)
	out, err := sess.execCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("gagal mengeksekusi %s: %w", cmd, err)
	}

	detail := parseOpticalRSSIOutput(out, pon, onuID)

	// Fetch ONU Name & MAC from cached or fresh table
	onus, _ := c.GetONUs(ctx, host, port, username, password, pon)
	for _, o := range onus {
		if o.ONUID == onuID {
			detail.Name = o.Name
			detail.SerialNumber = o.SerialNumber
			detail.Status = o.Status
			break
		}
	}

	return detail, nil
}

func (c *hsgqClient) GetEmptyONUIDs(ctx context.Context, host string, port int, username, password string, pon int) ([]int, error) {
	onus, err := c.GetONUs(ctx, host, port, username, password, pon)
	if err != nil {
		return nil, err
	}

	used := make(map[int]bool)
	for _, o := range onus {
		used[o.ONUID] = true
	}

	// HSGQ EPON supports up to 64 ONUs per PON port
	var emptyIDs []int
	for id := 1; id <= 64; id++ {
		if !used[id] {
			emptyIDs = append(emptyIDs, id)
		}
	}

	return emptyIDs, nil
}

func (c *hsgqClient) GetUplinks(ctx context.Context, host string, port int, username, password string) (*domain.ZTEUplinksData, error) {
	// HSGQ-E04ID has 1 fixed master board, 4 EPON ports, and 8 GE uplink ports
	cards := []domain.ZTECardInfo{
		{
			EntIndex: 1,
			Slot:     1,
			Type:     "HSGQ-E04ID",
			Role:     "Master",
		},
	}

	var ports []domain.ZTEPortInfo
	for i := 1; i <= 4; i++ {
		ports = append(ports, domain.ZTEPortInfo{
			Name:        fmt.Sprintf("EPON 0/%d", i),
			Shelf:       1,
			Slot:        1,
			Port:        i,
			Kind:        "EPON",
			AdminStatus: "up",
			OperStatus:  "up",
			SpeedMbps:   1250,
		})
	}

	for i := 1; i <= 8; i++ {
		ports = append(ports, domain.ZTEPortInfo{
			Name:        fmt.Sprintf("GE %d", i),
			Shelf:       1,
			Slot:        1,
			Port:        i,
			Kind:        "GE Uplink",
			AdminStatus: "up",
			OperStatus:  "up",
			SpeedMbps:   1000,
		})
	}

	return &domain.ZTEUplinksData{
		Cards: cards,
		Ports: ports,
	}, nil
}

// Regex for line in "show onu-info all"
// 1/1     70:2e:22:0a:59:47 Online    TRUE  TRUE  2026/10/01 02:06:04      PRM-A2-26-Ar NO-DESCRIPTI
var onuLineRegex = regexp.MustCompile(`^\s*(\d+)/(\d+)\s+([0-9a-fA-F:]{17})\s+(\S+)\s+(\S+)\s+(\S+)\s+([-\d/ :]+)\s+(\S+)(?:\s+(.*))?`)

func parseONUInfoOutput(output string) []domain.ZTEONUInfo {
	var list []domain.ZTEONUInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "---") || strings.HasPrefix(trimmed, "PON/ONU") || strings.HasPrefix(trimmed, "Total:") {
			continue
		}

		matches := onuLineRegex.FindStringSubmatch(line)
		if len(matches) >= 9 {
			pon, _ := strconv.Atoi(matches[1])
			onuID, _ := strconv.Atoi(matches[2])
			mac := strings.ToLower(matches[3])
			status := matches[4]
			name := matches[8]

			list = append(list, domain.ZTEONUInfo{
				Board:        1,
				PON:          pon,
				ONUID:        onuID,
				Name:         name,
				ONUType:      "EPON",
				SerialNumber: mac,
				Status:       status,
			})
		}
	}

	return list
}

var (
	rxPowerRegex = regexp.MustCompile(`(?i)Reci?eved\s+Power\s*:\s*([-\d\.]+)`)
	txPowerRegex = regexp.MustCompile(`(?i)Transmit\s+Power\s*:\s*([-\d\.]+)`)
	tempRegex    = regexp.MustCompile(`(?i)Temperature\s*:\s*([-\d\.]+)`)
	voltageRegex = regexp.MustCompile(`(?i)Voltage\s*:\s*([-\d\.]+)`)
	biasRegex    = regexp.MustCompile(`(?i)Bias\s*:\s*([-\d\.]+)`)
)

func parseOpticalRSSIOutput(output string, pon, onuID int) *domain.ZTEONUDetail {
	detail := &domain.ZTEONUDetail{
		Board:   1,
		PON:     pon,
		ONUID:   onuID,
		ONUType: "EPON",
		Status:  "Unknown",
	}

	if m := rxPowerRegex.FindStringSubmatch(output); len(m) > 1 {
		val, err := strconv.ParseFloat(m[1], 64)
		if err == nil {
			detail.RxPower = fmt.Sprintf("%.2f dBm", val)
		} else {
			detail.RxPower = m[1] + " dBm"
		}
	}

	if m := txPowerRegex.FindStringSubmatch(output); len(m) > 1 {
		val, err := strconv.ParseFloat(m[1], 64)
		if err == nil {
			detail.TxPower = fmt.Sprintf("%.2f dBm", val)
		} else {
			detail.TxPower = m[1] + " dBm"
		}
	}

	var descParts []string
	if m := tempRegex.FindStringSubmatch(output); len(m) > 1 {
		val, err := strconv.ParseFloat(m[1], 64)
		if err == nil {
			descParts = append(descParts, fmt.Sprintf("Suhu: %.1f°C", val))
		}
	}
	if m := voltageRegex.FindStringSubmatch(output); len(m) > 1 {
		val, err := strconv.ParseFloat(m[1], 64)
		if err == nil {
			descParts = append(descParts, fmt.Sprintf("Voltase: %.2fV", val))
		}
	}
	if m := biasRegex.FindStringSubmatch(output); len(m) > 1 {
		val, err := strconv.ParseFloat(m[1], 64)
		if err == nil {
			descParts = append(descParts, fmt.Sprintf("Arus Bias: %.2fmA", val))
		}
	}

	if len(descParts) > 0 {
		detail.Description = strings.Join(descParts, " | ")
	}

	return detail
}

