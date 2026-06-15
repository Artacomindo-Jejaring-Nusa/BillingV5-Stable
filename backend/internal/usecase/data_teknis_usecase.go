package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"billing-backend/internal/domain"
	"billing-backend/internal/websocket"
	"billing-backend/pkg/mikrotik"
	"billing-backend/pkg/utils"

	"github.com/go-routeros/routeros"
	"github.com/xuri/excelize/v2"
)

type dataTeknisUsecase struct {
	dataTeknisRepo   domain.DataTeknisRepository
	mikrotikRepo     domain.MikrotikRepository
	pelangganRepo    domain.PelangganRepository
	paketLayananRepo domain.PaketLayananRepository
}

// NewDataTeknisUsecase creates a new instance of DataTeknisUsecase
func NewDataTeknisUsecase(
	dataTeknisRepo domain.DataTeknisRepository,
	mikrotikRepo domain.MikrotikRepository,
	pelangganRepo domain.PelangganRepository,
	paketLayananRepo domain.PaketLayananRepository,
) domain.DataTeknisUsecase {
	return &dataTeknisUsecase{
		dataTeknisRepo:   dataTeknisRepo,
		mikrotikRepo:     mikrotikRepo,
		pelangganRepo:    pelangganRepo,
		paketLayananRepo: paketLayananRepo,
	}
}

func (u *dataTeknisUsecase) executeRouterOS(ctx context.Context, serverID uint64, op func(*routeros.Client) error) error {
	server, err := u.mikrotikRepo.GetByID(ctx, serverID)
	if err != nil {
		return fmt.Errorf("failed to fetch Mikrotik server: %w", err)
	}

	if !server.IsActive {
		return fmt.Errorf("mikrotik server %s is inactive", server.Name)
	}

	decryptedPassword := ""
	if server.Password != "" {
		decryptedPassword = utils.GlobalEncryptionService.Decrypt(server.Password)
	}

	client, err := mikrotik.GlobalPool.GetConnection(server.HostIP, server.Port, server.Username, decryptedPassword)
	if err != nil {
		return fmt.Errorf("failed to connect to Mikrotik router: %w", err)
	}
	defer mikrotik.GlobalPool.ReturnConnection(client, server.HostIP, server.Port)

	return op(client)
}

func (u *dataTeknisUsecase) triggerMikrotikCreate(ctx context.Context, data *domain.DataTeknis) error {
	if data.MikrotikServerID == nil {
		return errors.New("mikrotik_server_id is nil")
	}
	return u.executeRouterOS(ctx, *data.MikrotikServerID, func(client *routeros.Client) error {
		profile := ""
		if data.ProfilePppoe != nil {
			profile = *data.ProfilePppoe
		}
		ip := ""
		if data.IPPelanggan != nil {
			ip = *data.IPPelanggan
		}
		return mikrotik.CreatePPPoESecret(client, data.IDPelanggan, data.PasswordPppoe, profile, ip)
	})
}

func (u *dataTeknisUsecase) triggerMikrotikUpdate(ctx context.Context, oldName string, data *domain.DataTeknis, newStatus string) error {
	if data.MikrotikServerID == nil {
		return errors.New("mikrotik_server_id is nil")
	}
	return u.executeRouterOS(ctx, *data.MikrotikServerID, func(client *routeros.Client) error {
		profile := ""
		disabled := "no"
		if newStatus == "Aktif" {
			if data.ProfilePppoe != nil {
				profile = *data.ProfilePppoe
			}
			disabled = "no"
		} else if newStatus == "Suspended" {
			profile = "SUSPENDED"
			disabled = "yes"
		} else {
			if data.ProfilePppoe != nil {
				profile = *data.ProfilePppoe
			}
		}

		ip := ""
		if data.IPPelanggan != nil {
			ip = *data.IPPelanggan
		}

		err := mikrotik.UpdatePPPoESecret(client, oldName, data.IDPelanggan, data.PasswordPppoe, profile, ip, disabled)
		if err != nil {
			return err
		}

		if newStatus == "Suspended" {
			_ = mikrotik.RemoveActiveConnection(client, data.IDPelanggan)
		}
		return nil
	})
}

func (u *dataTeknisUsecase) FetchAll(ctx context.Context, skip, limit int, search string, olt string, profile string, vlan string, onuPowerMin, onuPowerMax *int) ([]domain.DataTeknis, int64, error) {
	return u.dataTeknisRepo.GetAll(ctx, skip, limit, search, olt, profile, vlan, onuPowerMin, onuPowerMax)
}

func (u *dataTeknisUsecase) GetByID(ctx context.Context, id uint64) (*domain.DataTeknis, error) {
	return u.dataTeknisRepo.GetByID(ctx, id)
}

func (u *dataTeknisUsecase) GetByPelangganID(ctx context.Context, pelangganID uint64) (*domain.DataTeknis, error) {
	return u.dataTeknisRepo.GetByPelangganID(ctx, pelangganID)
}

func (u *dataTeknisUsecase) Store(ctx context.Context, data *domain.DataTeknis) error {
	// 1. Check if Pelanggan exists
	cust, err := u.pelangganRepo.GetByID(ctx, data.PelangganID)
	if err != nil {
		return err
	}
	if cust == nil {
		return errors.New("pelanggan not found")
	}

	// 2. Check if customer already has DataTeknis
	existing, _ := u.dataTeknisRepo.GetByPelangganID(ctx, data.PelangganID)
	if existing != nil {
		return errors.New("pelanggan sudah memiliki data teknis")
	}

	// 3. IP validation
	if data.IPPelanggan != nil && *data.IPPelanggan != "" {
		isTaken, err := u.CheckIPAddress(ctx, *data.IPPelanggan, nil)
		if err != nil {
			return err
		}
		if isTaken {
			return fmt.Errorf("IP address %s is already in use", *data.IPPelanggan)
		}
	}

	// Save to DB
	err = u.dataTeknisRepo.Create(ctx, data)
	if err != nil {
		return err
	}

	// Trigger WebSocket notification
	if websocket.GlobalHub != nil {
		websocket.GlobalHub.BroadcastNotification("new_technical_data", map[string]interface{}{
			"pelanggan_nama": cust.Nama,
		})
	}

	// Sync with Mikrotik
	if data.MikrotikServerID != nil {
		syncErr := u.triggerMikrotikCreate(ctx, data)
		if syncErr != nil {
			data.MikrotikSyncPending = true
			_ = u.dataTeknisRepo.Update(ctx, data)
		}
	}

	websocket.InvalidateDashboardCache(ctx)
	return nil
}

func (u *dataTeknisUsecase) Update(ctx context.Context, id uint64, data *domain.DataTeknis) error {
	existing, err := u.dataTeknisRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("data teknis not found")
	}

	oldIDPelanggan := existing.IDPelanggan

	// IP validation (if IP is changed)
	if data.IPPelanggan != nil && *data.IPPelanggan != "" {
		if existing.IPPelanggan == nil || *existing.IPPelanggan != *data.IPPelanggan {
			isTaken, err := u.CheckIPAddress(ctx, *data.IPPelanggan, &id)
			if err != nil {
				return err
			}
			if isTaken {
				return fmt.Errorf("IP address %s is already in use", *data.IPPelanggan)
			}
		}
	}

	// Update fields
	existing.IDPelanggan = data.IDPelanggan
	existing.PasswordPppoe = data.PasswordPppoe
	existing.ProfilePppoe = data.ProfilePppoe
	existing.IPPelanggan = data.IPPelanggan
	existing.IDVlan = data.IDVlan
	existing.Olt = data.Olt
	existing.OltCustom = data.OltCustom
	existing.Pon = data.Pon
	existing.Otb = data.Otb
	existing.Odc = data.Odc
	existing.OdpID = data.OdpID
	existing.PortOdp = data.PortOdp
	existing.Sn = data.Sn
	existing.OnuPower = data.OnuPower
	existing.SpeedtestProof = data.SpeedtestProof
	existing.MikrotikServerID = data.MikrotikServerID

	err = u.dataTeknisRepo.Update(ctx, existing)
	if err != nil {
		return err
	}

	// Sync with Mikrotik
	if existing.MikrotikServerID != nil {
		status := "Aktif"
		if existing.Pelanggan != nil && len(existing.Pelanggan.Langganan) > 0 {
			status = existing.Pelanggan.Langganan[0].Status
		}
		syncErr := u.triggerMikrotikUpdate(ctx, oldIDPelanggan, existing, status)
		if syncErr != nil {
			existing.MikrotikSyncPending = true
			_ = u.dataTeknisRepo.Update(ctx, existing)
		} else {
			if existing.MikrotikSyncPending {
				existing.MikrotikSyncPending = false
				_ = u.dataTeknisRepo.Update(ctx, existing)
			}
		}
	}

	websocket.InvalidateDashboardCache(ctx)
	return nil
}

func (u *dataTeknisUsecase) Delete(ctx context.Context, id uint64) error {
	existing, err := u.dataTeknisRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("data teknis not found")
	}

	if existing.MikrotikServerID != nil {
		_ = u.executeRouterOS(ctx, *existing.MikrotikServerID, func(client *routeros.Client) error {
			_, err := mikrotik.DeletePPPoESecret(client, existing.IDPelanggan)
			return err
		})
	}

	err = u.dataTeknisRepo.Delete(ctx, id)
	if err == nil {
		websocket.InvalidateDashboardCache(ctx)
	}
	return err
}

func (u *dataTeknisUsecase) GetAvailableOLT(ctx context.Context) ([]string, error) {
	return u.dataTeknisRepo.GetAvailableOLT(ctx)
}

func (u *dataTeknisUsecase) GetAvailableProfiles(ctx context.Context) ([]string, error) {
	return u.dataTeknisRepo.GetAvailableProfiles(ctx)
}

func (u *dataTeknisUsecase) GetAvailableVlans(ctx context.Context) ([]string, error) {
	return u.dataTeknisRepo.GetAvailableVlans(ctx)
}

func (u *dataTeknisUsecase) GetOnuPowerRanges(ctx context.Context) (map[string]int, error) {
	min, max, err := u.dataTeknisRepo.GetOnuPowerRanges(ctx)
	if err != nil {
		return nil, err
	}
	res := make(map[string]int)
	if min != nil {
		res["min"] = *min
	} else {
		res["min"] = 0
	}
	if max != nil {
		res["max"] = *max
	} else {
		res["max"] = 0
	}
	return res, nil
}

func (u *dataTeknisUsecase) CheckIPAddress(ctx context.Context, ip string, excludeID *uint64) (bool, error) {
	// 1. Check in database
	existsInDB, err := u.dataTeknisRepo.CheckIPAddress(ctx, ip, excludeID)
	if err != nil {
		return false, err
	}
	if existsInDB {
		return true, nil
	}

	// 2. Check in all active Mikrotik servers
	servers, err := u.mikrotikRepo.GetAll(ctx)
	if err != nil {
		return false, err
	}

	for _, server := range servers {
		if !server.IsActive {
			continue
		}

		decryptedPassword := ""
		if server.Password != "" {
			decryptedPassword = utils.GlobalEncryptionService.Decrypt(server.Password)
		}

		client, err := mikrotik.GlobalPool.GetConnection(server.HostIP, server.Port, server.Username, decryptedPassword)
		if err != nil {
			continue
		}

		owner, err := mikrotik.CheckIPInSecrets(client, ip)
		mikrotik.GlobalPool.ReturnConnection(client, server.HostIP, server.Port)
		if err == nil && owner != "" {
			if excludeID != nil {
				existing, err := u.dataTeknisRepo.GetByID(ctx, *excludeID)
				if err == nil && existing != nil && existing.IDPelanggan == owner {
					continue
				}
			}
			return true, nil
		}
	}

	return false, nil
}

func (u *dataTeknisUsecase) GetAvailableProfilesForPackage(ctx context.Context, packageID uint64, pelangganID uint64, mikrotikServerID *uint64) ([]map[string]interface{}, error) {
	paket, err := u.paketLayananRepo.GetByID(ctx, packageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch package details: %w", err)
	}
	if paket == nil {
		return nil, errors.New("paket layanan tidak ditemukan")
	}

	var serverID uint64
	if mikrotikServerID != nil && *mikrotikServerID != 0 {
		serverID = *mikrotikServerID
	} else {
		dt, _ := u.dataTeknisRepo.GetByPelangganID(ctx, pelangganID)
		if dt != nil && dt.MikrotikServerID != nil {
			serverID = *dt.MikrotikServerID
		} else {
			cust, _ := u.pelangganRepo.GetByID(ctx, pelangganID)
			if cust != nil && cust.MikrotikServerID != nil {
				serverID = *cust.MikrotikServerID
			} else {
				return nil, errors.New("mikrotik server id is required")
			}
		}
	}

	var result []map[string]interface{}

	err = u.executeRouterOS(ctx, serverID, func(client *routeros.Client) error {
		profiles, err := mikrotik.GetAllPPPProfiles(client)
		if err != nil {
			return err
		}

		secrets, err := mikrotik.GetAllPPPSecrets(client)
		if err != nil {
			return err
		}

		usageCount := make(map[string]int)
		for _, secret := range secrets {
			if prof, ok := secret["profile"]; ok {
				usageCount[prof]++
			}
		}

		speedStr := fmt.Sprintf("%dMbps", paket.Kecepatan)
		for _, prof := range profiles {
			if strings.Contains(strings.ToLower(prof), strings.ToLower(speedStr)) {
				result = append(result, map[string]interface{}{
					"profile_name": prof,
					"usage_count":  usageCount[prof],
				})
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (u *dataTeknisUsecase) GetLastUsedIP(ctx context.Context, mikrotikServerID uint64) (map[string]interface{}, error) {
	var lastIP string
	var lastOctet int
	source := "mikrotik"

	err := u.executeRouterOS(ctx, mikrotikServerID, func(client *routeros.Client) error {
		secrets, err := mikrotik.GetAllPPPSecrets(client)
		if err != nil {
			return err
		}

		maxOctet := 0
		bestIP := ""
		for _, secret := range secrets {
			ipStr, ok := secret["remote-address"]
			if !ok || ipStr == "" {
				continue
			}

			ip := net.ParseIP(ipStr)
			if ip == nil {
				continue
			}

			ipv4 := ip.To4()
			if ipv4 == nil {
				continue
			}

			octet := int(ipv4[3])
			if octet > maxOctet {
				maxOctet = octet
				bestIP = ipStr
			}
		}

		if bestIP != "" {
			lastIP = bestIP
			lastOctet = maxOctet
		}
		return nil
	})

	if err != nil {
		source = "database"
		list, _, repoErr := u.dataTeknisRepo.GetAll(ctx, 0, 1000, "", "", "", "", nil, nil)
		if repoErr == nil {
			maxOctet := 0
			bestIP := ""
			for _, dt := range list {
				if dt.MikrotikServerID != nil && *dt.MikrotikServerID == mikrotikServerID && dt.IPPelanggan != nil && *dt.IPPelanggan != "" {
					ip := net.ParseIP(*dt.IPPelanggan)
					if ip != nil {
						ipv4 := ip.To4()
						if ipv4 != nil {
							octet := int(ipv4[3])
							if octet > maxOctet {
								maxOctet = octet
								bestIP = *dt.IPPelanggan
							}
						}
					}
				}
			}
			if bestIP != "" {
				lastIP = bestIP
				lastOctet = maxOctet
			}
		}
	}

	if lastIP == "" {
		return map[string]interface{}{
			"last_ip":    nil,
			"last_octet": 0,
			"source":     source,
		}, nil
	}

	return map[string]interface{}{
		"last_ip":    lastIP,
		"last_octet": lastOctet,
		"source":     source,
	}, nil
}

func (u *dataTeknisUsecase) ImportFromCSV(ctx context.Context, csvContent string) (int, error) {
	delimiter := ','
	firstLine := csvContent
	if idx := strings.Index(csvContent, "\n"); idx != -1 {
		firstLine = csvContent[:idx]
	}
	if strings.Count(firstLine, ";") > strings.Count(firstLine, ",") {
		delimiter = ';'
	}

	reader := csv.NewReader(strings.NewReader(csvContent))
	reader.Comma = delimiter
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return 0, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) < 2 {
		return 0, errors.New("CSV file is empty or only has headers")
	}

	headers := records[0]
	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	required := []string{"email_pelanggan", "id_pelanggan", "password_pppoe"}
	for _, req := range required {
		if _, ok := headerMap[req]; !ok {
			return 0, fmt.Errorf("missing required column: %s", req)
		}
	}

	var emails []string
	var serverNames []string
	var odpCodes []string

	type RowData struct {
		email          string
		idPelanggan    string
		passwordPppoe  string
		profilePppoe   string
		ipPelanggan    string
		idVlan         string
		olt            string
		oltCustom      string
		pon            string
		otb            string
		odc            string
		kodeOdp        string
		portOdp        string
		sn             string
		onuPower       string
		speedtestProof string
		serverName     string
	}

	var rows []RowData
	for i := 1; i < len(records); i++ {
		row := records[i]
		if len(row) < len(headers) {
			continue
		}

		getVal := func(key string) string {
			if idx, ok := headerMap[key]; ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		rd := RowData{
			email:          getVal("email_pelanggan"),
			idPelanggan:    getVal("id_pelanggan"),
			passwordPppoe:  getVal("password_pppoe"),
			profilePppoe:   getVal("profile_pppoe"),
			ipPelanggan:    getVal("ip_pelanggan"),
			idVlan:         getVal("id_vlan"),
			olt:            getVal("olt"),
			oltCustom:      getVal("olt_custom"),
			pon:            getVal("pon"),
			otb:            getVal("otb"),
			odc:            getVal("odc"),
			kodeOdp:        getVal("kode_odp"),
			portOdp:        getVal("port_odp"),
			sn:             getVal("sn"),
			onuPower:       getVal("onu_power"),
			speedtestProof: getVal("speedtest_proof"),
			serverName:     getVal("mikrotik_server"),
		}

		if rd.email == "" {
			continue
		}

		emails = append(emails, rd.email)
		if rd.serverName != "" {
			serverNames = append(serverNames, rd.serverName)
		}
		if rd.kodeOdp != "" {
			odpCodes = append(odpCodes, rd.kodeOdp)
		}
		rows = append(rows, rd)
	}

	customers, err := u.pelangganRepo.GetByEmails(ctx, emails)
	if err != nil {
		return 0, fmt.Errorf("failed to prefetch customers: %w", err)
	}
	customerMap := make(map[string]domain.Pelanggan)
	for _, c := range customers {
		customerMap[strings.ToLower(c.Email)] = c
	}

	var servers []domain.MikrotikServer
	if len(serverNames) > 0 {
		servers, err = u.mikrotikRepo.GetByNames(ctx, serverNames)
		if err != nil {
			return 0, fmt.Errorf("failed to prefetch mikrotik servers: %w", err)
		}
	}
	serverMap := make(map[string]domain.MikrotikServer)
	for _, s := range servers {
		serverMap[strings.ToLower(s.Name)] = s
	}

	var odps []domain.ODP
	if len(odpCodes) > 0 {
		odps, err = u.dataTeknisRepo.GetOdpByCodes(ctx, odpCodes)
		if err != nil {
			return 0, fmt.Errorf("failed to prefetch odps: %w", err)
		}
	}
	odpMap := make(map[string]domain.ODP)
	for _, o := range odps {
		odpMap[strings.ToLower(o.KodeOdp)] = o
	}

	var validationErrors []string
	var successCount int

	for idx, rd := range rows {
		rowNum := idx + 2
		cust, exists := customerMap[strings.ToLower(rd.email)]
		if !exists {
			validationErrors = append(validationErrors, fmt.Sprintf("Baris %d: Pelanggan dengan email '%s' tidak ditemukan", rowNum, rd.email))
			continue
		}

		existingDt, _ := u.dataTeknisRepo.GetByPelangganID(ctx, cust.ID)
		if existingDt != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("Baris %d: Pelanggan '%s' sudah memiliki data teknis", rowNum, rd.email))
			continue
		}

		var serverID *uint64
		if rd.serverName != "" {
			srv, srvExists := serverMap[strings.ToLower(rd.serverName)]
			if !srvExists {
				validationErrors = append(validationErrors, fmt.Sprintf("Baris %d: Mikrotik server '%s' tidak ditemukan", rowNum, rd.serverName))
				continue
			}
			serverID = &srv.ID
		}

		var odpID *uint64
		if rd.kodeOdp != "" {
			od, odExists := odpMap[strings.ToLower(rd.kodeOdp)]
			if !odExists {
				validationErrors = append(validationErrors, fmt.Sprintf("Baris %d: ODP '%s' tidak ditemukan", rowNum, rd.kodeOdp))
				continue
			}
			odpID = &od.ID
		}

		var profile *string
		if rd.profilePppoe != "" {
			profile = &rd.profilePppoe
		}
		var ip *string
		if rd.ipPelanggan != "" {
			ip = &rd.ipPelanggan
		}
		var vlan *string
		if rd.idVlan != "" {
			vlan = &rd.idVlan
		}
		var olt *string
		if rd.olt != "" {
			olt = &rd.olt
		}
		var oltCustom *string
		if rd.oltCustom != "" {
			oltCustom = &rd.oltCustom
		}
		var sn *string
		if rd.sn != "" {
			sn = &rd.sn
		}
		var speedtestProof *string
		if rd.speedtestProof != "" {
			speedtestProof = &rd.speedtestProof
		}

		var pon *int
		if rd.pon != "" {
			if v, err := strconv.Atoi(rd.pon); err == nil {
				pon = &v
			}
		}
		var otb *int
		if rd.otb != "" {
			if v, err := strconv.Atoi(rd.otb); err == nil {
				otb = &v
			}
		}
		var odc *int
		if rd.odc != "" {
			if v, err := strconv.Atoi(rd.odc); err == nil {
				odc = &v
			}
		}
		var portOdp *int
		if rd.portOdp != "" {
			if v, err := strconv.Atoi(rd.portOdp); err == nil {
				portOdp = &v
			}
		}
		var onuPower *int
		if rd.onuPower != "" {
			if v, err := strconv.Atoi(rd.onuPower); err == nil {
				onuPower = &v
			}
		}

		dt := &domain.DataTeknis{
			PelangganID:      cust.ID,
			IDPelanggan:      rd.idPelanggan,
			PasswordPppoe:    rd.passwordPppoe,
			ProfilePppoe:     profile,
			IPPelanggan:      ip,
			IDVlan:           vlan,
			Olt:              olt,
			OltCustom:        oltCustom,
			Pon:              pon,
			Otb:              otb,
			Odc:              odc,
			OdpID:            odpID,
			PortOdp:          portOdp,
			Sn:               sn,
			OnuPower:         onuPower,
			SpeedtestProof:   speedtestProof,
			MikrotikServerID: serverID,
		}

		err = u.Store(ctx, dt)
		if err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("Baris %d: Gagal menyimpan data: %v", rowNum, err))
			continue
		}
		successCount++
	}

	if len(validationErrors) > 0 {
		return successCount, errors.New(strings.Join(validationErrors, "; "))
	}

	return successCount, nil
}

func (u *dataTeknisUsecase) Export(ctx context.Context, format string) ([]byte, string, error) {
	headers := []string{
		"ID", "ID Pelanggan", "Password PPPoE", "Profile PPPoE", "IP Address", 
		"VLAN", "OLT", "OLT Custom", "PON", "OTB", "ODC", "ODP ID", "ODP Code", "Port ODP", "SN", "ONU Power",
	}
	limit := 1000
	offset := 0

	if format == "excel" {
		f := excelize.NewFile()
		sheet := "Data Teknis"
		f.SetSheetName("Sheet1", sheet)
		for i, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue(sheet, cell, h)
		}

		row := 2
		for {
			data, _, err := u.dataTeknisRepo.GetAll(ctx, offset, limit, "", "", "", "", nil, nil)
			if err != nil {
				return nil, "", err
			}
			if len(data) == 0 {
				break
			}

			for _, d := range data {
				profile := ""
				if d.ProfilePppoe != nil {
					profile = *d.ProfilePppoe
				}
				ip := ""
				if d.IPPelanggan != nil {
					ip = *d.IPPelanggan
				}
				vlan := ""
				if d.IDVlan != nil {
					vlan = *d.IDVlan
				}
				olt := ""
				if d.Olt != nil {
					olt = *d.Olt
				}
				oltCustom := ""
				if d.OltCustom != nil {
					oltCustom = *d.OltCustom
				}
				pon := 0
				if d.Pon != nil {
					pon = *d.Pon
				}
				otb := 0
				if d.Otb != nil {
					otb = *d.Otb
				}
				odc := 0
				if d.Odc != nil {
					odc = *d.Odc
				}
				odpID := uint64(0)
				if d.OdpID != nil {
					odpID = *d.OdpID
				}
				odpCode := ""
				if d.Odp != nil {
					odpCode = d.Odp.KodeOdp
				}
				portOdp := 0
				if d.PortOdp != nil {
					portOdp = *d.PortOdp
				}
				sn := ""
				if d.Sn != nil {
					sn = *d.Sn
				}
				op := 0
				if d.OnuPower != nil {
					op = *d.OnuPower
				}

				vals := []interface{}{
					d.ID, d.IDPelanggan, d.PasswordPppoe, profile, ip, 
					vlan, olt, oltCustom, pon, otb, odc, odpID, odpCode, portOdp, sn, op,
				}
				for c, v := range vals {
					cell, _ := excelize.CoordinatesToCellName(c+1, row)
					f.SetCellValue(sheet, cell, v)
				}
				row++
			}

			offset += limit
			if len(data) < limit {
				break
			}
		}

		buf, _ := f.WriteToBuffer()
		return buf.Bytes(), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", nil
	} else {
		buf := new(bytes.Buffer)
		w := csv.NewWriter(buf)
		w.Comma = ';'
		w.Write(headers)

		for {
			data, _, err := u.dataTeknisRepo.GetAll(ctx, offset, limit, "", "", "", "", nil, nil)
			if err != nil {
				return nil, "", err
			}
			if len(data) == 0 {
				break
			}

			for _, d := range data {
				profile := ""
				if d.ProfilePppoe != nil {
					profile = *d.ProfilePppoe
				}
				ip := ""
				if d.IPPelanggan != nil {
					ip = *d.IPPelanggan
				}
				vlan := ""
				if d.IDVlan != nil {
					vlan = *d.IDVlan
				}
				olt := ""
				if d.Olt != nil {
					olt = *d.Olt
				}
				oltCustom := ""
				if d.OltCustom != nil {
					oltCustom = *d.OltCustom
				}
				pon := "0"
				if d.Pon != nil {
					pon = fmt.Sprintf("%d", *d.Pon)
				}
				otb := "0"
				if d.Otb != nil {
					otb = fmt.Sprintf("%d", *d.Otb)
				}
				odc := "0"
				if d.Odc != nil {
					odc = fmt.Sprintf("%d", *d.Odc)
				}
				odpID := "0"
				if d.OdpID != nil {
					odpID = fmt.Sprintf("%d", *d.OdpID)
				}
				odpCode := ""
				if d.Odp != nil {
					odpCode = d.Odp.KodeOdp
				}
				portOdp := "0"
				if d.PortOdp != nil {
					portOdp = fmt.Sprintf("%d", *d.PortOdp)
				}
				sn := ""
				if d.Sn != nil {
					sn = *d.Sn
				}
				op := "0"
				if d.OnuPower != nil {
					op = fmt.Sprintf("%d", *d.OnuPower)
				}

				w.Write([]string{
					fmt.Sprintf("%d", d.ID), d.IDPelanggan, d.PasswordPppoe, profile, ip,
					vlan, olt, oltCustom, pon, otb, odc, odpID, odpCode, portOdp, sn, op,
				})
			}

			offset += limit
			if len(data) < limit {
				break
			}
		}

		w.Flush()
		return buf.Bytes(), "text/csv", nil
	}
}
