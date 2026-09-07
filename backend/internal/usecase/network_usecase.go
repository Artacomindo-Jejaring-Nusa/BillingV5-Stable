package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"billing-backend/internal/domain"
	"billing-backend/internal/websocket"
	"billing-backend/pkg/mikrotik"
	"billing-backend/pkg/utils"
	"billing-backend/pkg/zteclient"
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
	oltRepo   domain.OLTRepository
	zteClient zteclient.Client
}

func NewOLTUsecase(repo domain.OLTRepository, zteClient zteclient.Client) domain.OLTUsecase {
	return &oltUsecase{
		oltRepo:   repo,
		zteClient: zteClient,
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

	if u.zteClient != nil && u.isZTE(olt) {
		oltID := u.getZTEOltID(olt)
		uplinks, err := u.zteClient.GetUplinks(ctx, oltID)
		if err != nil {
			return "", fmt.Errorf("gagal terhubung ke SNMP ZTE OLT %s (%s): %w", olt.NamaOlt, olt.IPAddress, err)
		}
		return fmt.Sprintf("Berhasil terhubung ke ZTE OLT %s (%s). Terdeteksi %d kartu dan %d port uplink", olt.NamaOlt, olt.IPAddress, len(uplinks.Cards), len(uplinks.Ports)), nil
	}

	return fmt.Sprintf("Successfully connected to OLT %s at %s", olt.NamaOlt, olt.IPAddress), nil
}

func (u *oltUsecase) GetUplinks(ctx context.Context, id uint64) (*domain.ZTEUplinksData, error) {
	olt, err := u.oltRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u.zteClient == nil {
		return nil, errors.New("ZTE OLT client service is not initialized")
	}
	return u.zteClient.GetUplinks(ctx, u.getZTEOltID(olt))
}

func (u *oltUsecase) GetONUs(ctx context.Context, id uint64, board int, pon int) ([]domain.ZTEONUInfo, error) {
	olt, err := u.oltRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u.zteClient == nil {
		return nil, errors.New("ZTE OLT client service is not initialized")
	}
	return u.zteClient.GetONUs(ctx, u.getZTEOltID(olt), board, pon)
}

func (u *oltUsecase) GetPaginatedONUs(ctx context.Context, id uint64, board int, pon int, page int, limit int) (*domain.ZTEPaginatedONUs, error) {
	olt, err := u.oltRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u.zteClient == nil {
		return nil, errors.New("ZTE OLT client service is not initialized")
	}
	return u.zteClient.GetPaginatedONUs(ctx, u.getZTEOltID(olt), board, pon, page, limit)
}

func (u *oltUsecase) GetONUDetail(ctx context.Context, id uint64, board int, pon int, onuID int) (*domain.ZTEONUDetail, error) {
	olt, err := u.oltRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u.zteClient == nil {
		return nil, errors.New("ZTE OLT client service is not initialized")
	}
	return u.zteClient.GetONUDetail(ctx, u.getZTEOltID(olt), board, pon, onuID)
}

func (u *oltUsecase) GetEmptyONUIDs(ctx context.Context, id uint64, board int, pon int) ([]int, error) {
	olt, err := u.oltRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
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
	if u.zteClient == nil {
		return errors.New("ZTE OLT client service is not initialized")
	}
	return u.zteClient.ClearCache(ctx, u.getZTEOltID(olt), board, pon)
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

	return u.odpRepo.Update(ctx, existing)
}

func (u *odpUsecase) Delete(ctx context.Context, id uint64) error {
	_, err := u.odpRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return u.odpRepo.Delete(ctx, id)
}

