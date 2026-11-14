package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"billing-backend/internal/domain"
	"billing-backend/internal/websocket"
	"billing-backend/pkg/hsgqclient"
	"billing-backend/pkg/mikrotik"
	"billing-backend/pkg/utils"
	"billing-backend/pkg/zteclient"
	"billing-backend/pkg/ztetelnet"
)

type mikrotikUsecase struct {
	mikrotikRepo domain.MikrotikRepository
}

func NewMikrotikUsecase(repo domain.MikrotikRepository) domain.MikrotikUsecase {
	return &mikrotikUsecase{
		mikrotikRepo: repo,
	}
}

func (u *mikrotikUsecase) FetchAll(ctx context.Context) ([]domain.MikrotikServer, error) {
	// We might not want to return passwords, but since they are encrypted, it's safer.
	// In the real world, we might want to map this to a DTO (Data Transfer Object).
	return u.mikrotikRepo.GetAll(ctx)
}

func (u *mikrotikUsecase) GetByID(ctx context.Context, id uint64) (*domain.MikrotikServer, error) {
	return u.mikrotikRepo.GetByID(ctx, id)
}

func (u *mikrotikUsecase) Store(ctx context.Context, server *domain.MikrotikServer) error {
	if server.Name == "" || server.HostIP == "" || server.Username == "" || server.Password == "" {
		return errors.New("name, host_ip, username, and password are required")
	}

	// Encrypt the password before saving
	encryptedPassword, err := utils.GlobalEncryptionService.Encrypt(server.Password)
	if err != nil {
		return errors.New("failed to encrypt mikrotik password")
	}
	server.Password = encryptedPassword

	return u.mikrotikRepo.Create(ctx, server)
}

func (u *mikrotikUsecase) Update(ctx context.Context, id uint64, req *domain.MikrotikServer) error {
	existing, err := u.mikrotikRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	existing.Name = req.Name
	existing.HostIP = req.HostIP
	existing.Username = req.Username
	existing.Port = req.Port
	existing.IsActive = req.IsActive

	// Only update password if a new one is provided
	if req.Password != "" {
		encryptedPassword, err := utils.GlobalEncryptionService.Encrypt(req.Password)
		if err != nil {
			return errors.New("failed to encrypt new password")
		}
		existing.Password = encryptedPassword
	}

	return u.mikrotikRepo.Update(ctx, existing)
}

func (u *mikrotikUsecase) Delete(ctx context.Context, id uint64) error {
	_, err := u.mikrotikRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return u.mikrotikRepo.Delete(ctx, id)
}

func (u *mikrotikUsecase) TestConnection(ctx context.Context, id uint64) (map[string]interface{}, *domain.MikrotikServer, error) {
	server, err := u.mikrotikRepo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	decryptedPassword := utils.GlobalEncryptionService.Decrypt(server.Password)

	identity, rosVersion, connErr := mikrotik.TestConnection(server.HostIP, server.Port, server.Username, decryptedPassword, 5*time.Second)

	now := time.Now()
	var status string
	if connErr != nil {
		status = "Failed"
		server.LastConnectionStatus = &status
		_ = u.mikrotikRepo.Update(ctx, server)

		// Invalidate dashboard stat cards cache
		if rdb := websocket.GetRedisClient(); rdb != nil {
			_ = rdb.Del(ctx, "dashboard:cache:stat_cards").Err()
		}

		testResult := map[string]interface{}{
			"message": fmt.Sprintf("Failed to connect to MikroTik server: %v", connErr),
			"data":    nil,
		}

		responseServer := *server
		responseServer.Password = ""
		return testResult, &responseServer, connErr
	}

	status = "Success"
	server.LastConnectionStatus = &status
	server.RosVersion = &rosVersion
	server.LastConnectedAt = &now

	_ = u.mikrotikRepo.Update(ctx, server)

	// Invalidate dashboard stat cards cache
	if rdb := websocket.GetRedisClient(); rdb != nil {
		_ = rdb.Del(ctx, "dashboard:cache:stat_cards").Err()
	}

	testResult := map[string]interface{}{
		"message": "Successfully connected to MikroTik server",
		"data": map[string]interface{}{
			"identity":    identity,
			"ros_version": rosVersion,
		},
	}

	responseServer := *server
	responseServer.Password = ""
	return testResult, &responseServer, nil
}

type oltUsecase struct {
	oltRepo    domain.OLTRepository
	zteClient  zteclient.Client
	zteTelnet  ztetelnet.Client
	hsgqClient hsgqclient.Client
}

func NewOLTUsecase(repo domain.OLTRepository, zteClient zteclient.Client, hsgqClient hsgqclient.Client, zteTelnet ztetelnet.Client) domain.OLTUsecase {
	return &oltUsecase{
		oltRepo:    repo,
		zteClient:  zteClient,
		hsgqClient: hsgqClient,
		zteTelnet:  zteTelnet,
	}
}

func (u *oltUsecase) getZTEOltID(olt *domain.OLT) string {
	if olt == nil {
		return "pinus"
	}
	id := strings.ToLower(strings.TrimSpace(olt.NamaOlt))
	id = strings.TrimPrefix(id, "olt-")
	id = strings.TrimPrefix(id, "olt ")
	id = strings.TrimPrefix(id, "olt_")
	id = strings.ReplaceAll(id, " ", "-")
	return id
}

func (u *oltUsecase) isZTE(olt *domain.OLT) bool {
	if olt == nil {
		return false
	}
	tipe := strings.ToUpper(strings.TrimSpace(olt.TipeOlt))
	if strings.Contains(tipe, "ZTE") || strings.Contains(tipe, "C320") || strings.Contains(tipe, "C300") {
		return true
	}
	name := strings.ToLower(olt.NamaOlt)
	return strings.Contains(name, "pinus") || strings.Contains(name, "pulogebang") || strings.Contains(name, "tipar")
}

func (u *oltUsecase) getZTECredentials(olt *domain.OLT) (string, string) {
	username := ""
	if olt != nil && olt.Username != nil && *olt.Username != "" {
		username = *olt.Username
	}
	password := ""
	if olt != nil && olt.Password != nil {
		password = *olt.Password
	}
	return username, password
}

func (u *oltUsecase) isHSGQ(olt *domain.OLT) bool {
	if olt == nil {
		return false
	}
	tipe := strings.ToUpper(strings.TrimSpace(olt.TipeOlt))
	name := strings.ToUpper(strings.TrimSpace(olt.NamaOlt))
	return strings.Contains(tipe, "HSGQ") || strings.Contains(name, "HSGQ")
}

func (u *oltUsecase) getHSGQCredentials(olt *domain.OLT) (string, string) {
	username := "root"
	if olt != nil && olt.Username != nil && *olt.Username != "" {
		username = *olt.Username
	}
	password := ""
	if olt != nil && olt.Password != nil {
		password = *olt.Password
	}
	return username, password
}

func (u *oltUsecase) FetchAll(ctx context.Context) ([]domain.OLT, error) {
	return u.oltRepo.GetAll(ctx)
}

func (u *oltUsecase) GetByID(ctx context.Context, id uint64) (*domain.OLT, error) {
	return u.oltRepo.GetByID(ctx, id)
}

func (u *oltUsecase) Store(ctx context.Context, olt *domain.OLT) error {
	if olt.NamaOlt == "" || olt.IPAddress == "" || olt.TipeOlt == "" {
		return errors.New("nama_olt, ip_address, and tipe_olt are required")
	}
	return u.oltRepo.Create(ctx, olt)
}

func (u *oltUsecase) Update(ctx context.Context, id uint64, req *domain.OLT) error {
	existing, err := u.oltRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	existing.NamaOlt = req.NamaOlt
	existing.IPAddress = req.IPAddress
	existing.TipeOlt = req.TipeOlt
	existing.Username = req.Username
	existing.MikrotikServerID = req.MikrotikServerID

	if req.Password != nil && *req.Password != "" {
		existing.Password = req.Password
	}

	return u.oltRepo.Update(ctx, existing)
}

func (u *oltUsecase) Delete(ctx context.Context, id uint64) error {
	_, err := u.oltRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return u.oltRepo.Delete(ctx, id)
}

func (u *oltUsecase) TestConnection(ctx context.Context, id uint64) (string, error) {
	olt, err := u.oltRepo.GetByID(ctx, id)
	if err != nil {
		return "", err
	}

	if u.isHSGQ(olt) {
		if u.hsgqClient == nil {
			return "", errors.New("HSGQ OLT client service is not initialized")
		}
		username, password := u.getHSGQCredentials(olt)
		return u.hsgqClient.TestConnection(ctx, olt.IPAddress, 23, username, password)
	}

	if u.isZTE(olt) {
		username, password := u.getZTECredentials(olt)
		if u.zteTelnet != nil && username != "" {
			res, err := u.zteTelnet.TestConnection(ctx, olt.IPAddress, 23, username, password)
			if err == nil {
				return res, nil
			}
			log.Printf("[ZTE OLT] Telnet test connection failed for %s (%s): %v. Trying SNMP gateway...", olt.NamaOlt, olt.IPAddress, err)
		}

		if u.zteClient != nil {
			oltID := u.getZTEOltID(olt)
			uplinks, err := u.zteClient.GetUplinks(ctx, oltID)
			if err != nil {
				return "", fmt.Errorf("gagal terhubung ke ZTE OLT %s (%s): %w", olt.NamaOlt, olt.IPAddress, err)
			}
			return fmt.Sprintf("Berhasil terhubung ke ZTE OLT %s (%s) via SNMP Gateway. Terdeteksi %d kartu dan %d port uplink", olt.NamaOlt, olt.IPAddress, len(uplinks.Cards), len(uplinks.Ports)), nil
		}
	}

	return fmt.Sprintf("Successfully connected to OLT %s at %s", olt.NamaOlt, olt.IPAddress), nil
}

func (u *oltUsecase) GetUplinks(ctx context.Context, id uint64) (*domain.ZTEUplinksData, error) {
	olt, err := u.oltRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if u.isHSGQ(olt) {
		if u.hsgqClient == nil {
			return nil, errors.New("HSGQ OLT client service is not initialized")
		}
		username, password := u.getHSGQCredentials(olt)
		return u.hsgqClient.GetUplinks(ctx, olt.IPAddress, 23, username, password)
	}

	if u.isZTE(olt) {
		username, password := u.getZTECredentials(olt)
		if u.zteTelnet != nil && username != "" {
			data, err := u.zteTelnet.GetUplinks(ctx, olt.IPAddress, 23, username, password)
			if err == nil && data != nil && (len(data.Cards) > 0 || len(data.Ports) > 0) {
				return data, nil
			}
			log.Printf("[ZTE OLT] Telnet GetUplinks failed for %s (%s): %v. Fallback to SNMP...", olt.NamaOlt, olt.IPAddress, err)
		}
		if u.zteClient != nil {
			return u.zteClient.GetUplinks(ctx, u.getZTEOltID(olt))
		}
		return nil, errors.New("ZTE OLT client service is not initialized")
	}

	if u.zteClient == nil {
		return nil, errors.New("ZTE OLT client service is not initialized")
	}
	return u.zteClient.GetUplinks(ctx, u.getZTEOltID(olt))
}

func normalizeDeviceIdentifier(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, " ", "")
	return s
}

func (u *oltUsecase) enrichONUList(ctx context.Context, onus []domain.ZTEONUInfo) []domain.ZTEONUInfo {
	if len(onus) == 0 {
		return onus
	}

	snMap, allList, err := u.oltRepo.GetCustomerDeviceInfoMap(ctx)
	if err != nil || (len(snMap) == 0 && len(allList) == 0) {
		return onus
	}

	for i := range onus {
		normSN := normalizeDeviceIdentifier(onus[i].SerialNumber)
		// 1. Try matching by Serial Number / MAC
		if normSN != "" {
			if match, ok := snMap[normSN]; ok {
				if match.Nama != "" {
					onus[i].CustomerName = match.Nama
				}
				if match.IDPelanggan != "" && (len(onus[i].Name) <= 12 || strings.HasPrefix(strings.ToLower(match.IDPelanggan), strings.ToLower(onus[i].Name))) {
					onus[i].Name = match.IDPelanggan
				}
				continue
			}
		}

		// 2. Fallback: match by prefix of ID Pelanggan if Name has 5+ chars
		rawName := strings.TrimSpace(onus[i].Name)
		if len(rawName) >= 5 {
			rawNameLower := strings.ToLower(rawName)
			for _, item := range allList {
				idLower := strings.ToLower(item.IDPelanggan)
				if idLower == rawNameLower || strings.HasPrefix(idLower, rawNameLower) {
					if item.IDPelanggan != "" {
						onus[i].Name = item.IDPelanggan
					}
					if item.Nama != "" {
						onus[i].CustomerName = item.Nama
					}
					break
				}
			}
		}
	}
	return onus
}

func (u *oltUsecase) enrichONUDetail(ctx context.Context, detail *domain.ZTEONUDetail) *domain.ZTEONUDetail {
	if detail == nil {
		return nil
	}

	snMap, allList, err := u.oltRepo.GetCustomerDeviceInfoMap(ctx)
	if err != nil || (len(snMap) == 0 && len(allList) == 0) {
		return detail
	}

	normSN := normalizeDeviceIdentifier(detail.SerialNumber)
	if normSN != "" {
		if match, ok := snMap[normSN]; ok {
			if match.Nama != "" && detail.CustomerName == "" {
				detail.CustomerName = match.Nama
			}
			if match.IDPelanggan != "" && (len(detail.Name) <= 12 || strings.HasPrefix(strings.ToLower(match.IDPelanggan), strings.ToLower(detail.Name))) {
				detail.Name = match.IDPelanggan
			}
			return detail
		}
	}

	rawName := strings.TrimSpace(detail.Name)
	if len(rawName) >= 5 {
		rawNameLower := strings.ToLower(rawName)
		for _, item := range allList {
			idLower := strings.ToLower(item.IDPelanggan)
			if idLower == rawNameLower || strings.HasPrefix(idLower, rawNameLower) {
				if item.IDPelanggan != "" {
					detail.Name = item.IDPelanggan
				}
				if item.Nama != "" && detail.CustomerName == "" {
					detail.CustomerName = item.Nama
				}
				break
			}
		}
	}
	return detail
}

func (u *oltUsecase) GetONUs(ctx context.Context, id uint64, board int, pon int) ([]domain.ZTEONUInfo, error) {
	olt, err := u.oltRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var data []domain.ZTEONUInfo
	if u.isHSGQ(olt) {
		if u.hsgqClient == nil {
			return nil, errors.New("HSGQ OLT client service is not initialized")
		}
		username, password := u.getHSGQCredentials(olt)
		data, err = u.hsgqClient.GetONUs(ctx, olt.IPAddress, 23, username, password, pon)
	} else if u.isZTE(olt) {
		username, password := u.getZTECredentials(olt)
		if u.zteTelnet != nil && username != "" {
			tData, tErr := u.zteTelnet.GetONUs(ctx, olt.IPAddress, 23, username, password, board, pon)
			if tErr == nil {
				data = tData
			} else {
				log.Printf("[ZTE OLT] Telnet GetONUs failed for %s (%s) board %d pon %d: %v. Fallback to SNMP...", olt.NamaOlt, olt.IPAddress, board, pon, tErr)
			}
		}
		if data == nil {
			if u.zteClient != nil {
				data, err = u.zteClient.GetONUs(ctx, u.getZTEOltID(olt), board, pon)
			} else {
				return nil, errors.New("ZTE OLT client service is not initialized")
			}
		}
	} else {
		if u.zteClient == nil {
			return nil, errors.New("ZTE OLT client service is not initialized")
		}
		data, err = u.zteClient.GetONUs(ctx, u.getZTEOltID(olt), board, pon)
	}

	if err != nil {
		return nil, err
	}

	return u.enrichONUList(ctx, data), nil
}

func (u *oltUsecase) GetPaginatedONUs(ctx context.Context, id uint64, board int, pon int, page int, limit int) (*domain.ZTEPaginatedONUs, error) {
	olt, err := u.oltRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var paginated *domain.ZTEPaginatedONUs
	if u.isHSGQ(olt) {
		if u.hsgqClient == nil {
			return nil, errors.New("HSGQ OLT client service is not initialized")
		}
		username, password := u.getHSGQCredentials(olt)
		paginated, err = u.hsgqClient.GetPaginatedONUs(ctx, olt.IPAddress, 23, username, password, pon, page, limit)
	} else if u.isZTE(olt) {
		username, password := u.getZTECredentials(olt)
		if u.zteTelnet != nil && username != "" {
			tData, tErr := u.zteTelnet.GetPaginatedONUs(ctx, olt.IPAddress, 23, username, password, board, pon, page, limit)
			if tErr == nil {
				paginated = tData
			} else {
				log.Printf("[ZTE OLT] Telnet GetPaginatedONUs failed for %s (%s): %v. Fallback to SNMP...", olt.NamaOlt, olt.IPAddress, tErr)
			}
		}
		if paginated == nil {
			if u.zteClient != nil {
				paginated, err = u.zteClient.GetPaginatedONUs(ctx, u.getZTEOltID(olt), board, pon, page, limit)
			} else {
				return nil, errors.New("ZTE OLT client service is not initialized")
			}
		}
	} else {
		if u.zteClient == nil {
			return nil, errors.New("ZTE OLT client service is not initialized")
		}
		paginated, err = u.zteClient.GetPaginatedONUs(ctx, u.getZTEOltID(olt), board, pon, page, limit)
	}

	if err != nil {
		return nil, err
	}

	if paginated != nil && len(paginated.Data) > 0 {
		paginated.Data = u.enrichONUList(ctx, paginated.Data)
	}

	return paginated, nil
}

func (u *oltUsecase) GetONUDetail(ctx context.Context, id uint64, board int, pon int, onuID int) (*domain.ZTEONUDetail, error) {
	olt, err := u.oltRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var detail *domain.ZTEONUDetail
	if u.isHSGQ(olt) {
		if u.hsgqClient == nil {
			return nil, errors.New("HSGQ OLT client service is not initialized")
		}
		username, password := u.getHSGQCredentials(olt)
		detail, err = u.hsgqClient.GetONUDetail(ctx, olt.IPAddress, 23, username, password, pon, onuID)
	} else if u.isZTE(olt) {
		username, password := u.getZTECredentials(olt)
		if u.zteTelnet != nil && username != "" {
			tData, tErr := u.zteTelnet.GetONUDetail(ctx, olt.IPAddress, 23, username, password, board, pon, onuID)
			if tErr == nil && tData != nil && (tData.RxPower != "" || tData.Name != "") {
				detail = tData
			} else {
				log.Printf("[ZTE OLT] Telnet GetONUDetail failed for %s (%s) onu %d: %v. Fallback to SNMP...", olt.NamaOlt, olt.IPAddress, onuID, tErr)
			}
		}
		if detail == nil {
			if u.zteClient != nil {
				detail, err = u.zteClient.GetONUDetail(ctx, u.getZTEOltID(olt), board, pon, onuID)
			} else {
				return nil, errors.New("ZTE OLT client service is not initialized")
			}
		}
	} else {
		if u.zteClient == nil {
			return nil, errors.New("ZTE OLT client service is not initialized")
		}
		detail, err = u.zteClient.GetONUDetail(ctx, u.getZTEOltID(olt), board, pon, onuID)
	}

	if err != nil {
		return nil, err
	}

	return u.enrichONUDetail(ctx, detail), nil
}

func (u *oltUsecase) GetEmptyONUIDs(ctx context.Context, id uint64, board int, pon int) ([]int, error) {
	olt, err := u.oltRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if u.isHSGQ(olt) {
		if u.hsgqClient == nil {
			return nil, errors.New("HSGQ OLT client service is not initialized")
		}
		username, password := u.getHSGQCredentials(olt)
		return u.hsgqClient.GetEmptyONUIDs(ctx, olt.IPAddress, 23, username, password, pon)
	}

	if u.isZTE(olt) {
		username, password := u.getZTECredentials(olt)
		if u.zteTelnet != nil && username != "" {
			data, err := u.zteTelnet.GetEmptyONUIDs(ctx, olt.IPAddress, 23, username, password, board, pon)
			if err == nil {
				return data, nil
			}
			log.Printf("[ZTE OLT] Telnet GetEmptyONUIDs failed for %s (%s): %v. Fallback to SNMP...", olt.NamaOlt, olt.IPAddress, err)
		}
		if u.zteClient != nil {
			return u.zteClient.GetEmptyONUIDs(ctx, u.getZTEOltID(olt), board, pon)
		}
		return nil, errors.New("ZTE OLT client service is not initialized")
	}

	if u.zteClient == nil {
		return nil, errors.New("ZTE OLT client service is not initialized")
	}
	return u.zteClient.GetEmptyONUIDs(ctx, u.getZTEOltID(olt), board, pon)
}

func (u *oltUsecase) GetONUSerials(ctx context.Context, id uint64, board int, pon int, noCache bool) ([]domain.ONUSerialInfo, error) {
	olt, err := u.oltRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if u.isHSGQ(olt) {
		if u.hsgqClient == nil {
			return nil, errors.New("HSGQ OLT client service is not initialized")
		}
		username, password := u.getHSGQCredentials(olt)
		onus, err := u.hsgqClient.GetONUs(ctx, olt.IPAddress, 23, username, password, pon)
		if err != nil {
			return nil, err
		}
		var serials []domain.ONUSerialInfo
		for _, o := range onus {
			serials = append(serials, domain.ONUSerialInfo{
				ONUID:        o.ONUID,
				SerialNumber: o.SerialNumber,
			})
		}
		return serials, nil
	}

	if u.isZTE(olt) {
		username, password := u.getZTECredentials(olt)
		if u.zteTelnet != nil && username != "" {
			onus, err := u.zteTelnet.GetONUs(ctx, olt.IPAddress, 23, username, password, board, pon)
			if err == nil {
				var serials []domain.ONUSerialInfo
				for _, o := range onus {
					serials = append(serials, domain.ONUSerialInfo{
						ONUID:        o.ONUID,
						SerialNumber: o.SerialNumber,
					})
				}
				return serials, nil
			}
			log.Printf("[ZTE OLT] Telnet GetONUSerials failed for %s (%s): %v. Fallback to SNMP...", olt.NamaOlt, olt.IPAddress, err)
		}
		if u.zteClient != nil {
			return u.zteClient.GetONUSerials(ctx, u.getZTEOltID(olt), board, pon, noCache)
		}
		return nil, errors.New("ZTE OLT client service is not initialized")
	}

	if u.zteClient == nil {
		return nil, errors.New("ZTE OLT client service is not initialized")
	}
	return u.zteClient.GetONUSerials(ctx, u.getZTEOltID(olt), board, pon, noCache)
}

func (u *oltUsecase) ClearCache(ctx context.Context, id uint64, board int, pon int) error {
	olt, err := u.oltRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if u.isHSGQ(olt) {
		if u.hsgqClient != nil {
			u.hsgqClient.ClearCache(olt.IPAddress)
		}
		return nil
	}

	if u.isZTE(olt) {
		if u.zteTelnet != nil {
			u.zteTelnet.ClearCache(olt.IPAddress)
		}
		if u.zteClient != nil {
			_ = u.zteClient.ClearCache(ctx, u.getZTEOltID(olt), board, pon)
		}
		return nil
	}

	if u.zteClient == nil {
		return errors.New("ZTE OLT client service is not initialized")
	}
	return u.zteClient.ClearCache(ctx, u.getZTEOltID(olt), board, pon)
}

type odcUsecase struct {
	odcRepo domain.ODCRepository
}

func NewODCUsecase(repo domain.ODCRepository) domain.ODCUsecase {
	return &odcUsecase{
		odcRepo: repo,
	}
}

func (u *odcUsecase) FetchAll(ctx context.Context) ([]domain.ODC, error) {
	return u.odcRepo.GetAll(ctx)
}

func (u *odcUsecase) GetByID(ctx context.Context, id uint64) (*domain.ODC, error) {
	return u.odcRepo.GetByID(ctx, id)
}

func (u *odcUsecase) Store(ctx context.Context, odc *domain.ODC) error {
	if odc.KodeOdc == "" || odc.NamaOdc == "" {
		return errors.New("kode_odc and nama_odc are required")
	}
	return u.odcRepo.Create(ctx, odc)
}

func (u *odcUsecase) Update(ctx context.Context, id uint64, req *domain.ODC) error {
	existing, err := u.odcRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	existing.KodeOdc = req.KodeOdc
	existing.NamaOdc = req.NamaOdc
	existing.Alamat = req.Alamat
	existing.KapasitasPort = req.KapasitasPort
	existing.Latitude = req.Latitude
	existing.Longitude = req.Longitude
	existing.OltID = req.OltID
	return u.odcRepo.Update(ctx, existing)
}

func (u *odcUsecase) Delete(ctx context.Context, id uint64) error {
	return u.odcRepo.Delete(ctx, id)
}

type odpUsecase struct {
	odpRepo domain.ODPRepository
}

func NewODPUsecase(repo domain.ODPRepository) domain.ODPUsecase {
	return &odpUsecase{
		odpRepo: repo,
	}
}

func (u *odpUsecase) FetchAll(ctx context.Context) ([]domain.ODP, error) {
	return u.odpRepo.GetAll(ctx)
}

func (u *odpUsecase) GetByID(ctx context.Context, id uint64) (*domain.ODP, error) {
	return u.odpRepo.GetByID(ctx, id)
}

func (u *odpUsecase) Store(ctx context.Context, odp *domain.ODP) error {
	if odp.KodeOdp == "" || odp.OltID == 0 {
		return errors.New("kode_odp and olt_id are required")
	}
	return u.odpRepo.Create(ctx, odp)
}

func (u *odpUsecase) Update(ctx context.Context, id uint64, req *domain.ODP) error {
	existing, err := u.odpRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	existing.KodeOdp = req.KodeOdp
	existing.Alamat = req.Alamat
	existing.KapasitasPort = req.KapasitasPort
	existing.Latitude = req.Latitude
	existing.Longitude = req.Longitude
	existing.ParentOdpID = req.ParentOdpID
	existing.OltID = req.OltID
	existing.OdcID = req.OdcID

	return u.odpRepo.Update(ctx, existing)
}

func (u *odpUsecase) Delete(ctx context.Context, id uint64) error {
	_, err := u.odpRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return u.odpRepo.Delete(ctx, id)
}

func (u *odpUsecase) BatchImport(ctx context.Context, items []domain.BatchODPItem) (int, error) {
	if len(items) == 0 {
		return 0, nil
	}
	var odps []domain.ODP
	for _, item := range items {
		if item.KodeOdp == "" {
			continue
		}
		kapasitas := item.KapasitasPort
		if kapasitas <= 0 {
			kapasitas = 8
		}
		alamat := item.Alamat
		odps = append(odps, domain.ODP{
			KodeOdp:       item.KodeOdp,
			Alamat:        &alamat,
			KapasitasPort: kapasitas,
			Latitude:      item.Latitude,
			Longitude:     item.Longitude,
			OltID:         item.OltID,
			OdcID:         item.OdcID,
		})
	}
	return u.odpRepo.BatchUpsert(ctx, odps)
}

func (u *odpUsecase) UpdateCoordinates(ctx context.Context, id uint64, lat, lng float64) error {
	return u.odpRepo.UpdateCoordinates(ctx, id, lat, lng)
}

func (u *odpUsecase) GetNetworkMapData(ctx context.Context) (*domain.NetworkMapData, error) {
	return u.odpRepo.GetNetworkMapData(ctx)
}

func (u *odpUsecase) UpdateNodeCoordinate(ctx context.Context, nodeType string, id uint64, lat, lng float64) error {
	return u.odpRepo.UpdateNodeCoordinate(ctx, nodeType, id, lat, lng)
}

