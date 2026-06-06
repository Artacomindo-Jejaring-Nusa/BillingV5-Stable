package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"billing-backend/internal/domain"
	"billing-backend/pkg/mikrotik"
	"billing-backend/pkg/utils"
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
	oltRepo domain.OLTRepository
}

func NewOLTUsecase(repo domain.OLTRepository) domain.OLTUsecase {
	return &oltUsecase{
		oltRepo: repo,
	}
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
	return fmt.Sprintf("Successfully connected to OLT %s at %s", olt.NamaOlt, olt.IPAddress), nil
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

