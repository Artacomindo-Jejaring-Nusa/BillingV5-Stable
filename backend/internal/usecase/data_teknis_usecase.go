package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"math"
	"net"
	"regexp"
	"strconv"
	"strings"

	"billing-backend/internal/domain"
	"billing-backend/internal/websocket"
	"billing-backend/pkg/database"
	"billing-backend/pkg/hsgqclient"
	"billing-backend/pkg/mikrotik"
	"billing-backend/pkg/utils"
	"billing-backend/pkg/zteclient"
	"billing-backend/pkg/ztetelnet"

	"github.com/go-routeros/routeros"
	"github.com/xuri/excelize/v2"
)

type dataTeknisUsecase struct {
	dataTeknisRepo   domain.DataTeknisRepository
	mikrotikRepo     domain.MikrotikRepository
	pelangganRepo    domain.PelangganRepository
	paketLayananRepo domain.PaketLayananRepository
	oltRepo          domain.OLTRepository
	zteClient        zteclient.Client
	hsgqClient       hsgqclient.Client
	zteTelnet        ztetelnet.Client
}

// NewDataTeknisUsecase creates a new instance of DataTeknisUsecase
func NewDataTeknisUsecase(
	dataTeknisRepo domain.DataTeknisRepository,
	mikrotikRepo domain.MikrotikRepository,
	pelangganRepo domain.PelangganRepository,
	paketLayananRepo domain.PaketLayananRepository,
	oltRepo domain.OLTRepository,
	zteClient zteclient.Client,
	clients ...interface{},
) domain.DataTeknisUsecase {
	u := &dataTeknisUsecase{
		dataTeknisRepo:   dataTeknisRepo,
		mikrotikRepo:     mikrotikRepo,
		pelangganRepo:    pelangganRepo,
		paketLayananRepo: paketLayananRepo,
		oltRepo:          oltRepo,
		zteClient:        zteClient,
	}
	for _, c := range clients {
		switch v := c.(type) {
		case hsgqclient.Client:
			u.hsgqClient = v
		case ztetelnet.Client:
			u.zteTelnet = v
		}
	}
	return u
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

func resolveSubscriptionStatus(pelanggan *domain.Pelanggan) string {
	if pelanggan == nil || len(pelanggan.Langganan) == 0 {
		return "Aktif"
	}
	for _, l := range pelanggan.Langganan {
		if l.Status == "Suspended" || l.Status == "Berhenti" {
			return l.Status
		}
		if l.Status == "Aktif" {
			return "Aktif"
		}
	}
	return pelanggan.Langganan[0].Status
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
		if data.Pelanggan != nil && data.Pelanggan.MikrotikServerID != nil {
			data.MikrotikServerID = data.Pelanggan.MikrotikServerID
		} else {
			return errors.New("mikrotik_server_id is nil")
		}
	}
	return u.executeRouterOS(ctx, *data.MikrotikServerID, func(client *routeros.Client) error {
		profile := "default"
		disabled := "no"
		if newStatus == "Aktif" {
			if data.ProfilePppoe != nil && strings.TrimSpace(*data.ProfilePppoe) != "" && !strings.EqualFold(strings.TrimSpace(*data.ProfilePppoe), "SUSPENDED") {
				profile = strings.TrimSpace(*data.ProfilePppoe)
			}
			disabled = "no"
		} else if newStatus == "Suspended" || newStatus == "Berhenti" {
			profile = "SUSPENDED"
			disabled = "yes"
		} else {
			if data.ProfilePppoe != nil && strings.TrimSpace(*data.ProfilePppoe) != "" && !strings.EqualFold(strings.TrimSpace(*data.ProfilePppoe), "SUSPENDED") {
				profile = strings.TrimSpace(*data.ProfilePppoe)
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

		if newStatus == "Suspended" || newStatus == "Berhenti" {
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
		// Check in DB
		existsInDB, err := u.dataTeknisRepo.CheckIPAddress(ctx, *data.IPPelanggan, nil)
		if err != nil {
			return err
		}
		if existsInDB {
			return fmt.Errorf("IP address %s is already in use in database", *data.IPPelanggan)
		}

		// Check in Mikrotik: Only check the RELEVANT Mikrotik server (data.MikrotikServerID or matched by OLT)
		var targetServers []domain.MikrotikServer
		if data.MikrotikServerID != nil {
			if srv, err := u.mikrotikRepo.GetByID(ctx, *data.MikrotikServerID); err == nil && srv != nil && srv.IsActive {
				targetServers = append(targetServers, *srv)
			}
		} else if data.Olt != nil && *data.Olt != "" {
			if srv, err := u.mikrotikRepo.GetByName(ctx, *data.Olt); err == nil && srv != nil && srv.IsActive {
				targetServers = append(targetServers, *srv)
			}
		}

		for _, server := range targetServers {
			decryptedPassword := ""
			if server.Password != "" {
				decryptedPassword = utils.GlobalEncryptionService.Decrypt(server.Password)
			}
			client, err := mikrotik.GlobalPool.GetConnection(server.HostIP, server.Port, server.Username, decryptedPassword)
			if err != nil {
				continue
			}
			owner, err := mikrotik.CheckIPInSecrets(client, *data.IPPelanggan)
			mikrotik.GlobalPool.ReturnConnection(client, server.HostIP, server.Port)
			if err == nil && owner != "" {
				cleanOwner := strings.TrimSpace(strings.ToLower(owner))
				cleanID := strings.TrimSpace(strings.ToLower(data.IDPelanggan))
				if cleanOwner != cleanID {
					return fmt.Errorf("IP address %s is already in use by %s on Mikrotik (%s)", *data.IPPelanggan, owner, server.Name)
				}
			}
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
			existsInDB, err := u.dataTeknisRepo.CheckIPAddress(ctx, *data.IPPelanggan, &id)
			if err != nil {
				return err
			}
			if existsInDB {
				return fmt.Errorf("IP address %s is already in use in database", *data.IPPelanggan)
			}

			var targetServers []domain.MikrotikServer
			if data.MikrotikServerID != nil {
				if srv, err := u.mikrotikRepo.GetByID(ctx, *data.MikrotikServerID); err == nil && srv != nil && srv.IsActive {
					targetServers = append(targetServers, *srv)
				}
			} else if data.Olt != nil && *data.Olt != "" {
				if srv, err := u.mikrotikRepo.GetByName(ctx, *data.Olt); err == nil && srv != nil && srv.IsActive {
					targetServers = append(targetServers, *srv)
				}
			}

			for _, server := range targetServers {
				decryptedPassword := ""
				if server.Password != "" {
					decryptedPassword = utils.GlobalEncryptionService.Decrypt(server.Password)
				}
				client, err := mikrotik.GlobalPool.GetConnection(server.HostIP, server.Port, server.Username, decryptedPassword)
				if err != nil {
					continue
				}
				owner, err := mikrotik.CheckIPInSecrets(client, *data.IPPelanggan)
				mikrotik.GlobalPool.ReturnConnection(client, server.HostIP, server.Port)
				if err == nil && owner != "" {
					cleanOwner := strings.TrimSpace(strings.ToLower(owner))
					cleanID := strings.TrimSpace(strings.ToLower(data.IDPelanggan))
					cleanOldID := strings.TrimSpace(strings.ToLower(oldIDPelanggan))
					if cleanOwner != cleanID && cleanOwner != cleanOldID {
						return fmt.Errorf("IP address %s is already in use by %s on Mikrotik (%s)", *data.IPPelanggan, owner, server.Name)
					}
				}
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
		status := resolveSubscriptionStatus(existing.Pelanggan)
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
	var subnetPrefix string
	source := "mikrotik"
	usedIPsMap := make(map[string]bool)
	var serverName string

	if server, err := u.mikrotikRepo.GetByID(ctx, mikrotikServerID); err == nil && server != nil {
		serverName = server.Name
	}

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

			usedIPsMap[ipStr] = true
			octet := int(ipv4[3])
			if octet > maxOctet {
				maxOctet = octet
				bestIP = ipStr
				subnetPrefix = fmt.Sprintf("%d.%d.%d.", ipv4[0], ipv4[1], ipv4[2])
			}
		}

		if bestIP != "" {
			lastIP = bestIP
			lastOctet = maxOctet
		} else {
			// If no secrets with remote-address, try to check IP pools on router
			pools, err := mikrotik.GetIPPools(client)
			if err == nil && len(pools) > 0 {
				for _, pool := range pools {
					if ranges, ok := pool["ranges"]; ok && ranges != "" {
						parts := strings.Split(ranges, "-")
						if len(parts) > 0 {
							firstIP := strings.TrimSpace(parts[0])
							ip := net.ParseIP(firstIP)
							if ip != nil {
								if ipv4 := ip.To4(); ipv4 != nil {
									subnetPrefix = fmt.Sprintf("%d.%d.%d.", ipv4[0], ipv4[1], ipv4[2])
									break
								}
							}
						}
					}
				}
			}
		}
		return nil
	})

	// Also merge with Database records for this mikrotik server
	list, _, repoErr := u.dataTeknisRepo.GetAll(ctx, 0, 5000, "", "", "", "", nil, nil)
	if repoErr == nil {
		for _, dt := range list {
			if dt.MikrotikServerID != nil && *dt.MikrotikServerID == mikrotikServerID && dt.IPPelanggan != nil && *dt.IPPelanggan != "" {
				usedIPsMap[*dt.IPPelanggan] = true
				if lastIP == "" {
					ip := net.ParseIP(*dt.IPPelanggan)
					if ip != nil {
						if ipv4 := ip.To4(); ipv4 != nil {
							octet := int(ipv4[3])
							if octet > lastOctet {
								lastOctet = octet
								lastIP = *dt.IPPelanggan
								subnetPrefix = fmt.Sprintf("%d.%d.%d.", ipv4[0], ipv4[1], ipv4[2])
							}
						}
					}
				}
			}
		}
	}

	if err != nil {
		source = "database"
	}

	// Calculate suggested_ip and available_ips
	var suggestedIP string
	var availableIPs []string

	if subnetPrefix != "" {
		// First check starting from lastOctet + 1 up to 254
		startOctet := lastOctet + 1
		if startOctet < 2 {
			startOctet = 2
		}

		for octet := startOctet; octet <= 254; octet++ {
			candidateIP := fmt.Sprintf("%s%d", subnetPrefix, octet)
			if !usedIPsMap[candidateIP] {
				if suggestedIP == "" {
					suggestedIP = candidateIP
				}
				availableIPs = append(availableIPs, candidateIP)
				if len(availableIPs) >= 8 {
					break
				}
			}
		}

		// If we still need more available IPs (e.g. lastOctet was near 254 or to fill earlier gaps),
		// scan from octet 2 to startOctet
		if len(availableIPs) < 8 {
			for octet := 2; octet < startOctet; octet++ {
				candidateIP := fmt.Sprintf("%s%d", subnetPrefix, octet)
				if !usedIPsMap[candidateIP] {
					if suggestedIP == "" {
						suggestedIP = candidateIP
					}
					exists := false
					for _, a := range availableIPs {
						if a == candidateIP {
							exists = true
							break
						}
					}
					if !exists {
						availableIPs = append(availableIPs, candidateIP)
						if len(availableIPs) >= 8 {
							break
						}
					}
				}
			}
		}
	}

	message := ""
	if lastIP != "" && suggestedIP != "" {
		message = fmt.Sprintf("IP Terakhir: %s • Rekomendasi IP Berikutnya: %s", lastIP, suggestedIP)
	} else if suggestedIP != "" {
		message = fmt.Sprintf("Rekomendasi IP: %s", suggestedIP)
	} else if lastIP != "" {
		message = fmt.Sprintf("IP Terakhir: %s", lastIP)
	} else {
		message = "Belum ada IP yang terpakai di server ini"
	}

	var lastIPVal interface{} = lastIP
	if lastIP == "" {
		lastIPVal = nil
	}
	var suggestedIPVal interface{} = suggestedIP
	if suggestedIP == "" {
		suggestedIPVal = nil
	}

	return map[string]interface{}{
		"last_ip":       lastIPVal,
		"last_octet":    lastOctet,
		"suggested_ip":  suggestedIPVal,
		"available_ips": availableIPs,
		"total_used":    len(usedIPsMap),
		"subnet_prefix": subnetPrefix,
		"server_name":   serverName,
		"source":        source,
		"message":       message,
	}, nil
}

func (u *dataTeknisUsecase) GetUnconfiguredPelanggan(ctx context.Context, search string) ([]map[string]interface{}, error) {
	pelanggans, err := u.dataTeknisRepo.GetUnconfiguredPelanggan(ctx, search)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(pelanggans))
	for _, p := range pelanggans {
		alamatDisplay := p.Alamat
		if p.AlamatCustom != nil && *p.AlamatCustom != "" {
			if alamatDisplay != "" {
				alamatDisplay += " - " + *p.AlamatCustom
			} else {
				alamatDisplay = *p.AlamatCustom
			}
		}
		layananStr := ""
		if p.Layanan != nil {
			layananStr = *p.Layanan
		}
		result = append(result, map[string]interface{}{
			"id":         p.ID,
			"nama":       p.Nama,
			"alamat":     alamatDisplay,
			"layanan":    layananStr,
			"created_at": p.CreatedAt,
		})
	}
	return result, nil
}

func (u *dataTeknisUsecase) GetPelangganDetail(ctx context.Context, pelangganID uint64) (map[string]interface{}, error) {
	cust, err := u.pelangganRepo.GetByID(ctx, pelangganID)
	if err != nil {
		return nil, err
	}
	if cust == nil {
		return nil, errors.New("pelanggan tidak ditemukan")
	}

	layananStr := ""
	if cust.Layanan != nil {
		layananStr = *cust.Layanan
	}
	brandStr := ""
	if cust.IDBrand != nil {
		brandStr = *cust.IDBrand
	}
	alamatDisplay := cust.Alamat
	if cust.AlamatCustom != nil && *cust.AlamatCustom != "" {
		if alamatDisplay != "" {
			alamatDisplay += " - " + *cust.AlamatCustom
		} else {
			alamatDisplay = *cust.AlamatCustom
		}
	}

	return map[string]interface{}{
		"id":         cust.ID,
		"nama":       cust.Nama,
		"layanan":    layananStr,
		"id_brand":   brandStr,
		"alamat":     alamatDisplay,
		"created_at": cust.CreatedAt,
	}, nil
}

func (u *dataTeknisUsecase) ImportFromCSV(ctx context.Context, csvContent string) (int, error) {
	records, err := utils.ParseCSV(csvContent)
	if err != nil {
		return 0, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) < 2 {
		return 0, errors.New("CSV file is empty or only has headers")
	}

	headers := records[0]
	headerMap := make(map[string]int)
	for i, h := range headers {
		norm := utils.NormalizeCSVHeader(h)
		headerMap[norm] = i
		headerMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	findCol := func(keys ...string) int {
		for _, k := range keys {
			norm := utils.NormalizeCSVHeader(k)
			if idx, ok := headerMap[norm]; ok {
				return idx
			}
			if idx, ok := headerMap[strings.ToLower(strings.TrimSpace(k))]; ok {
				return idx
			}
		}
		return -1
	}

	emailColIdx := findCol("email_pelanggan", "email", "email_customer", "email_user")
	idPelangganColIdx := findCol("id_pelanggan", "id pelanggan", "customer_id", "id_customer", "username_pppoe", "id")
	passwordPppoeColIdx := findCol("password_pppoe", "password pppoe", "password", "pppoe_password", "pass_pppoe", "pass")

	if emailColIdx == -1 {
		return 0, fmt.Errorf("missing required column: email_pelanggan")
	}
	if idPelangganColIdx == -1 {
		return 0, fmt.Errorf("missing required column: id_pelanggan")
	}
	if passwordPppoeColIdx == -1 {
		return 0, fmt.Errorf("missing required column: password_pppoe")
	}

	profileColIdx := findCol("profile_pppoe", "profile pppoe", "profile", "paket", "profile_paket")
	ipColIdx := findCol("ip_pelanggan", "ip pelanggan", "ip_address", "ip", "ip_static")
	vlanColIdx := findCol("id_vlan", "id vlan", "vlan", "vlan_id")
	oltColIdx := findCol("olt", "nama_olt", "olt_name")
	oltCustomColIdx := findCol("olt_custom", "olt_custor", "olt_cust", "olt custom", "custom_olt")
	ponColIdx := findCol("pon", "port_pon", "pon_port")
	otbColIdx := findCol("otb", "port_otb")
	odcColIdx := findCol("odc", "port_odc")
	odpColIdx := findCol("kode_odp", "kode odp", "odp", "nama_odp", "odp_code")
	portOdpColIdx := findCol("port_odp", "port odp", "port", "port_ke")
	snColIdx := findCol("sn", "serial_number", "serialnumber", "sn_modem", "mac_sn")
	onuPowerColIdx := findCol("onu_power", "onu power", "redaman", "power_onu", "rx_power")
	speedtestColIdx := findCol("speedtest_proof", "speedtest proof", "speedtest", "bukti_speedtest")
	serverColIdx := findCol("mikrotik_server", "mikrotik server", "server", "router", "mikrotik")

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
		if len(row) == 0 {
			continue
		}

		getVal := func(colIdx int) string {
			if colIdx >= 0 && colIdx < len(row) {
				return strings.TrimSpace(row[colIdx])
			}
			return ""
		}

		rd := RowData{
			email:          getVal(emailColIdx),
			idPelanggan:    getVal(idPelangganColIdx),
			passwordPppoe:  getVal(passwordPppoeColIdx),
			profilePppoe:   getVal(profileColIdx),
			ipPelanggan:    getVal(ipColIdx),
			idVlan:         getVal(vlanColIdx),
			olt:            getVal(oltColIdx),
			oltCustom:      getVal(oltCustomColIdx),
			pon:            getVal(ponColIdx),
			otb:            getVal(otbColIdx),
			odc:            getVal(odcColIdx),
			kodeOdp:        getVal(odpColIdx),
			portOdp:        getVal(portOdpColIdx),
			sn:             getVal(snColIdx),
			onuPower:       getVal(onuPowerColIdx),
			speedtestProof: getVal(speedtestColIdx),
			serverName:     getVal(serverColIdx),
		}

		if rd.email == "" {
			continue
		}

		emails = append(emails, rd.email)
		if rd.serverName != "" {
			serverNames = append(serverNames, rd.serverName)
		} else if rd.olt != "" {
			serverNames = append(serverNames, rd.olt)
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
		} else if rd.olt != "" {
			if srv, srvExists := serverMap[strings.ToLower(rd.olt)]; srvExists {
				serverID = &srv.ID
			}
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

func (u *dataTeknisUsecase) Export(ctx context.Context, format string, filters domain.DataTeknisFilterParams) ([]byte, string, error) {
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
			data, _, err := u.dataTeknisRepo.GetAll(ctx, offset, limit, filters.Search, filters.Olt, filters.Profile, filters.Vlan, filters.OnuPowerMin, filters.OnuPowerMax)
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
			data, _, err := u.dataTeknisRepo.GetAll(ctx, offset, limit, filters.Search, filters.Olt, filters.Profile, filters.Vlan, filters.OnuPowerMin, filters.OnuPowerMax)
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

// AutoSyncProfileForPelanggan secara otomatis memilih dan mengalokasikan profile PPPoE yang paling renggang (penggunaan paling sedikit)
// pada router Mikrotik berdasarkan kecepatan paket layanan yang baru, lalu memperbarui Data Teknis dan sinkronisasi ke Mikrotik.
func (u *dataTeknisUsecase) AutoSyncProfileForPelanggan(ctx context.Context, pelangganID uint64, paket *domain.PaketLayanan) (string, error) {
	if paket == nil {
		return "", errors.New("paket layanan is nil")
	}

	dt, err := u.dataTeknisRepo.GetByPelangganID(ctx, pelangganID)
	if err != nil || dt == nil {
		// Pelanggan belum memiliki konfigurasi Data Teknis
		return "", nil
	}

	var serverID uint64
	if dt.MikrotikServerID != nil && *dt.MikrotikServerID != 0 {
		serverID = *dt.MikrotikServerID
	} else {
		cust, _ := u.pelangganRepo.GetByID(ctx, pelangganID)
		if cust != nil && cust.MikrotikServerID != nil && *cust.MikrotikServerID != 0 {
			serverID = *cust.MikrotikServerID
			dt.MikrotikServerID = cust.MikrotikServerID
		}
	}

	var bestProfile string
	minUsage := math.MaxInt32

	// 1. Prioritas Utama: Ambil daftar profile dan live usage count dari Mikrotik RouterOS langsung
	if serverID != 0 {
		_ = u.executeRouterOS(ctx, serverID, func(client *routeros.Client) error {
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

			speedStr := fmt.Sprintf("%dmbps", paket.Kecepatan)
			speedShort1 := fmt.Sprintf("%dm-", paket.Kecepatan)
			speedShort2 := fmt.Sprintf("%dm_", paket.Kecepatan)
			speedShort3 := fmt.Sprintf("%dm ", paket.Kecepatan)

			var matchingCandidates []string
			for _, prof := range profiles {
				pLower := strings.ToLower(prof)
				// Abaikan profile default / suspended
				if pLower == "default" || pLower == "default-encryption" || pLower == "suspended" {
					continue
				}
				if strings.Contains(pLower, speedStr) || strings.HasPrefix(pLower, speedShort1) || strings.HasPrefix(pLower, speedShort2) || strings.HasPrefix(pLower, speedShort3) {
					matchingCandidates = append(matchingCandidates, prof)
				}
			}

			// Cari candidate profile dengan pemakaian terkecil (paling renggang)
			for _, prof := range matchingCandidates {
				cnt := usageCount[prof]
				if cnt < minUsage {
					minUsage = cnt
					bestProfile = prof
				}
			}

			return nil
		})
	}

	// 2. Fallback: Jika Mikrotik sedang offline / tidak terjangkau, cari dari profile yang ada di database lokal
	if bestProfile == "" {
		availableProfiles, _ := u.dataTeknisRepo.GetAvailableProfiles(ctx)
		speedStr := fmt.Sprintf("%dmbps", paket.Kecepatan)
		var matchingCandidates []string
		for _, prof := range availableProfiles {
			pLower := strings.ToLower(prof)
			if pLower == "default" || pLower == "default-encryption" || pLower == "suspended" {
				continue
			}
			if strings.Contains(pLower, speedStr) {
				matchingCandidates = append(matchingCandidates, prof)
			}
		}

		if len(matchingCandidates) > 0 {
			db := database.GetDB()
			if db != nil {
				type ProfileCount struct {
					ProfilePppoe string
					Total        int
				}
				var counts []ProfileCount
				db.WithContext(ctx).Table("data_teknis").
					Select("profile_pppoe, count(*) as total").
					Where("profile_pppoe IN ?", matchingCandidates).
					Group("profile_pppoe").
					Scan(&counts)

				countMap := make(map[string]int)
				for _, c := range counts {
					countMap[c.ProfilePppoe] = c.Total
				}

				for _, prof := range matchingCandidates {
					cnt := countMap[prof]
					if cnt < minUsage {
						minUsage = cnt
						bestProfile = prof
					}
				}
			} else {
				bestProfile = matchingCandidates[0]
			}
		}
	}

	if bestProfile == "" {
		return "", fmt.Errorf("tidak ditemukan profile yang sesuai untuk kecepatan %d Mbps", paket.Kecepatan)
	}

	// 3. Update Profile di Data Teknis
	dt.ProfilePppoe = &bestProfile
	err = u.dataTeknisRepo.Update(ctx, dt)
	if err != nil {
		return "", fmt.Errorf("gagal update profile data teknis: %w", err)
	}

	// 4. Sinkronisasi perubahan profile langsung ke Mikrotik PPP Secret
	if serverID != 0 && dt.IDPelanggan != "" {
		status := resolveSubscriptionStatus(dt.Pelanggan)
		syncErr := u.triggerMikrotikUpdate(ctx, dt.IDPelanggan, dt, status)
		if syncErr != nil {
			dt.MikrotikSyncPending = true
			_ = u.dataTeknisRepo.Update(ctx, dt)
		} else if dt.MikrotikSyncPending {
			dt.MikrotikSyncPending = false
			_ = u.dataTeknisRepo.Update(ctx, dt)
		}
	}

	websocket.InvalidateDashboardCache(ctx)
	return bestProfile, nil
}

func (u *dataTeknisUsecase) findOLT(ctx context.Context, oltName string) (*domain.OLT, error) {
	if u.oltRepo == nil {
		return nil, errors.New("OLT repository is not initialized")
	}
	olts, err := u.oltRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	search := strings.TrimSpace(oltName)
	searchLower := strings.ToLower(search)

	// 1. Exact match on NamaOlt
	for i := range olts {
		if strings.EqualFold(olts[i].NamaOlt, search) {
			return &olts[i], nil
		}
	}

	// 2. Exact match on IPAddress
	for i := range olts {
		if olts[i].IPAddress == search {
			return &olts[i], nil
		}
	}

	// 3. Substring match
	for i := range olts {
		nameLower := strings.ToLower(olts[i].NamaOlt)
		if strings.Contains(nameLower, searchLower) || strings.Contains(searchLower, nameLower) {
			return &olts[i], nil
		}
	}

	// 4. Key match (e.g. tipar, pulogebang, pinus, parama)
	searchKey := u.resolveOltKey(search)
	for i := range olts {
		if u.resolveOltKey(olts[i].NamaOlt) == searchKey {
			return &olts[i], nil
		}
	}

	return nil, fmt.Errorf("OLT %s tidak ditemukan di database", oltName)
}

func (u *dataTeknisUsecase) isHSGQ(olt *domain.OLT) bool {
	if olt == nil {
		return false
	}
	tipe := strings.ToUpper(strings.TrimSpace(olt.TipeOlt))
	name := strings.ToUpper(strings.TrimSpace(olt.NamaOlt))
	return strings.Contains(tipe, "HSGQ") || strings.Contains(name, "HSGQ")
}

func (u *dataTeknisUsecase) isZTE(olt *domain.OLT) bool {
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

func (u *dataTeknisUsecase) getHSGQCredentials(olt *domain.OLT) (string, string) {
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

func (u *dataTeknisUsecase) getZTECredentials(olt *domain.OLT) (string, string) {
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

func parseRxPowerToInt(powerStr string) *int {
	clean := strings.TrimSpace(powerStr)
	if clean == "" || clean == "-" || strings.EqualFold(clean, "n/a") || strings.Contains(strings.ToLower(clean), "tidak") {
		return nil
	}
	var sb strings.Builder
	hasDigit := false
	for _, r := range clean {
		if (r >= '0' && r <= '9') || (r == '-' && sb.Len() == 0) || (r == '.' && strings.IndexRune(sb.String(), '.') == -1) {
			sb.WriteRune(r)
			if r >= '0' && r <= '9' {
				hasDigit = true
			}
		} else if hasDigit {
			break
		}
	}
	if !hasDigit {
		return nil
	}
	val, err := strconv.ParseFloat(sb.String(), 64)
	if err != nil {
		return nil
	}
	res := int(math.Round(val))
	return &res
}

func (u *dataTeknisUsecase) resolveOltKey(oltName string) string {
	name := strings.ToLower(strings.TrimSpace(oltName))
	if strings.Contains(name, "tipar") {
		return "tipar"
	}
	if strings.Contains(name, "pulogebang") || strings.Contains(name, "pulo gebang") {
		return "pulogebang"
	}
	if strings.Contains(name, "pinus") {
		return "pinus"
	}
	name = strings.TrimPrefix(name, "olt-")
	name = strings.TrimPrefix(name, "olt ")
	name = strings.TrimPrefix(name, "olt_")
	name = strings.ReplaceAll(name, " ", "-")
	if name == "" {
		return "tipar"
	}
	return name
}

var blockRegex = regexp.MustCompile(`([a-zA-Z]\d+)[-_/]?(\d+)`)

func cleanIdentifier(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, " ", "")
	return s
}

func stripPrefix(s string) string {
	lower := strings.ToLower(strings.TrimSpace(s))
	prefixes := []string{"prm-", "prm_", "tpr-", "tpr_", "ckg-", "ckg_", "pgb-", "pgb_", "rusun-", "rusun_", "blok-", "blok_"}
	for _, pfx := range prefixes {
		if strings.HasPrefix(lower, pfx) {
			lower = strings.TrimPrefix(lower, pfx)
			break
		}
	}
	return lower
}

func matchONUToCustomer(onu *domain.ZTEONUInfo, dt *domain.DataTeknis) bool {
	if onu == nil || dt == nil {
		return false
	}

	// 1. Match by Serial Number / MAC Address
	if onu.SerialNumber != "" && dt.Sn != nil && *dt.Sn != "" {
		normOnuSN := normalizeDeviceIdentifier(onu.SerialNumber)
		normDtSN := normalizeDeviceIdentifier(*dt.Sn)
		if normOnuSN != "" && normDtSN != "" && normOnuSN == normDtSN {
			return true
		}
	}

	rawONUName := strings.TrimSpace(onu.Name)
	rawONUDesc := strings.TrimSpace(onu.CustomerName)
	rawDtID := strings.TrimSpace(dt.IDPelanggan)

	cleanONU := cleanIdentifier(rawONUName)
	cleanDtID := cleanIdentifier(rawDtID)

	// 2. Exact match or prefix match on clean identifiers
	if cleanONU != "" && cleanDtID != "" {
		if cleanONU == cleanDtID {
			return true
		}
		// Handles truncated ONU names on OLT (e.g. "prmb519anissah" vs "prmb519anissa")
		if len(cleanONU) >= 5 && (strings.HasPrefix(cleanDtID, cleanONU) || strings.HasPrefix(cleanONU, cleanDtID)) {
			return true
		}
	}

	// 3. Match with OLT prefix stripped (e.g. "PRM-B5-19-Anissah" vs "B5-19-Anissah")
	strippedDtID := cleanIdentifier(stripPrefix(rawDtID))
	strippedONU := cleanIdentifier(stripPrefix(rawONUName))
	if strippedDtID != "" && strippedONU != "" {
		if strippedDtID == strippedONU {
			return true
		}
		if len(strippedONU) >= 4 && (strings.HasPrefix(strippedDtID, strippedONU) || strings.HasPrefix(strippedONU, strippedDtID)) {
			return true
		}
	}

	// 4. Block & Unit Pattern Match (e.g. "B5-19" from PRM-B5-19-Anissah vs B5-19 on OLT)
	mDt := blockRegex.FindStringSubmatch(rawDtID)
	mOnu := blockRegex.FindStringSubmatch(rawONUName)
	if len(mDt) >= 3 && len(mOnu) >= 3 {
		blockDt := strings.ToLower(mDt[1]) + "-" + mDt[2]
		blockOnu := strings.ToLower(mOnu[1]) + "-" + mOnu[2]
		if blockDt == blockOnu {
			return true
		}
	}

	// 5. Match by Customer Name Words (e.g. "Anissah" from "Anissah Rubianti")
	if dt.Pelanggan != nil && dt.Pelanggan.Nama != "" {
		nameWords := strings.Fields(strings.ToLower(dt.Pelanggan.Nama))
		for _, w := range nameWords {
			if len(w) >= 4 {
				if strings.Contains(strings.ToLower(rawONUName), w) || (rawONUDesc != "" && strings.Contains(strings.ToLower(rawONUDesc), w)) {
					return true
				}
			}
		}
	}

	// 6. Match by Description if technician wrote ID in description
	if rawONUDesc != "" && cleanDtID != "" {
		cleanDesc := cleanIdentifier(rawONUDesc)
		if cleanDesc == cleanDtID || (len(cleanDesc) >= 5 && (strings.HasPrefix(cleanDtID, cleanDesc) || strings.HasPrefix(cleanDesc, cleanDtID))) {
			return true
		}
	}

	return false
}

func (u *dataTeknisUsecase) GetLiveONU(ctx context.Context, id uint64) (*domain.ZTEONUDetail, error) {
	dt, err := u.dataTeknisRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("data teknis not found: %w", err)
	}

	oltName := ""
	if dt.Olt != nil {
		oltName = *dt.Olt
	}

	targetPon := 0
	if dt.Pon != nil {
		targetPon = *dt.Pon
	}

	olt, _ := u.findOLT(ctx, oltName)

	// --- 1. HSGQ Path ---
	if olt != nil && u.isHSGQ(olt) {
		if u.hsgqClient == nil {
			return nil, errors.New("layanan HSGQ OLT client belum diinisialisasi")
		}
		user, pass := u.getHSGQCredentials(olt)
		onus, err := u.hsgqClient.GetONUs(ctx, olt.IPAddress, 23, user, pass, 0)
		if err != nil {
			return nil, fmt.Errorf("gagal mengambil data ONU dari HSGQ %s: %w", olt.NamaOlt, err)
		}

		var matched *domain.ZTEONUInfo
		for i := range onus {
			if matchONUToCustomer(&onus[i], dt) {
				matched = &onus[i]
				break
			}
		}

		if matched != nil {
			detail, err := u.hsgqClient.GetONUDetail(ctx, olt.IPAddress, 23, user, pass, matched.PON, matched.ONUID)
			if err == nil && detail != nil {
				needsUpdate := false
				if detail.SerialNumber != "" && (dt.Sn == nil || *dt.Sn != detail.SerialNumber) {
					dt.Sn = &detail.SerialNumber
					needsUpdate = true
				}
				if dt.Pon == nil || *dt.Pon != matched.PON {
					ponVal := matched.PON
					dt.Pon = &ponVal
					needsUpdate = true
				}
				if detail.RxPower != "" {
					if pwr := parseRxPowerToInt(detail.RxPower); pwr != nil {
						if dt.OnuPower == nil || *dt.OnuPower != *pwr {
							dt.OnuPower = pwr
							needsUpdate = true
						}
					}
				}
				if needsUpdate {
					_ = u.dataTeknisRepo.Update(ctx, dt)
				}
				return detail, nil
			}
		}

		return nil, fmt.Errorf("ONU untuk pelanggan %s tidak ditemukan pada OLT HSGQ %s", dt.IDPelanggan, olt.NamaOlt)
	}

	// --- 2. ZTE Telnet Path ---
	if olt != nil && u.isZTE(olt) && u.zteTelnet != nil {
		user, pass := u.getZTECredentials(olt)
		if user != "" {
			ponsToScan := make([]int, 0)
			if targetPon > 0 {
				ponsToScan = append(ponsToScan, targetPon)
			}
			for p := 1; p <= 16; p++ {
				if targetPon > 0 && p == targetPon {
					continue
				}
				ponsToScan = append(ponsToScan, p)
			}

			var matched *domain.ZTEONUInfo
			for _, p := range ponsToScan {
				onus, err := u.zteTelnet.GetONUs(ctx, olt.IPAddress, 23, user, pass, 1, p)
				if err != nil || len(onus) == 0 {
					continue
				}

				for i := range onus {
					if matchONUToCustomer(&onus[i], dt) {
						matched = &onus[i]
						break
					}
				}

				if matched != nil {
					break
				}
			}

			if matched != nil {
				detail, err := u.zteTelnet.GetONUDetail(ctx, olt.IPAddress, 23, user, pass, 1, matched.PON, matched.ONUID)
				if err == nil && detail != nil {
					needsUpdate := false
					if detail.SerialNumber != "" && (dt.Sn == nil || *dt.Sn != detail.SerialNumber) {
						dt.Sn = &detail.SerialNumber
						needsUpdate = true
					}
					if dt.Pon == nil || *dt.Pon != matched.PON {
						ponVal := matched.PON
						dt.Pon = &ponVal
						needsUpdate = true
					}
					if detail.RxPower != "" {
						if pwr := parseRxPowerToInt(detail.RxPower); pwr != nil {
							if dt.OnuPower == nil || *dt.OnuPower != *pwr {
								dt.OnuPower = pwr
								needsUpdate = true
							}
						}
					}
					if needsUpdate {
						_ = u.dataTeknisRepo.Update(ctx, dt)
					}
					return detail, nil
				}
			}
		}
	}

	// --- 3. Fallback to ZTE SNMP Gateway ---
	if u.zteClient == nil {
		return nil, errors.New("ZTE OLT client service is not initialized")
	}

	oltKey := u.resolveOltKey(oltName)
	boards := []int{1, 2}

	// 1. If SN is provided, match by Serial Number
	if dt.Sn != nil && strings.TrimSpace(*dt.Sn) != "" {
		targetSN := strings.ToUpper(strings.TrimSpace(*dt.Sn))

		// Check specified PON first if available
		if targetPon > 0 {
			for _, b := range boards {
				onus, err := u.zteClient.GetONUs(ctx, oltKey, b, targetPon)
				if err == nil {
					for _, onu := range onus {
						if strings.EqualFold(strings.TrimSpace(onu.SerialNumber), targetSN) {
							return u.zteClient.GetONUDetail(ctx, oltKey, b, targetPon, onu.ONUID)
						}
					}
				}
			}
		}

		// Fallback scan across common PON ports on board 1 and 2
		for _, b := range boards {
			for p := 1; p <= 16; p++ {
				if targetPon > 0 && p == targetPon {
					continue
				}
				onus, err := u.zteClient.GetONUs(ctx, oltKey, b, p)
				if err == nil {
					for _, onu := range onus {
						if strings.EqualFold(strings.TrimSpace(onu.SerialNumber), targetSN) {
							return u.zteClient.GetONUDetail(ctx, oltKey, b, p, onu.ONUID)
						}
					}
				}
			}
		}

		return nil, fmt.Errorf("ONU dengan SN %s tidak ditemukan pada OLT %s", targetSN, oltKey)
	}

	// 2. If SN is empty, attempt smart auto-match by IP Pelanggan and ID/Name Pelanggan
	var targetIP string
	if dt.IPPelanggan != nil {
		targetIP = strings.TrimSpace(*dt.IPPelanggan)
	}
	searchID := strings.TrimSpace(dt.IDPelanggan)

	var searchName string
	cust, _ := u.pelangganRepo.GetByID(ctx, dt.PelangganID)
	if cust != nil {
		searchName = strings.TrimSpace(cust.Nama)
	}

	for _, b := range boards {
		for p := 1; p <= 16; p++ {
			// Prioritize configured PON if available
			if targetPon > 0 && p != targetPon {
				// We will scan configured PON first, but keep general loop
			}

			onus, err := u.zteClient.GetONUs(ctx, oltKey, b, p)
			if err != nil || len(onus) == 0 {
				continue
			}

			for _, onu := range onus {
				onuNameLower := strings.ToLower(onu.Name)

				// Match by ID Pelanggan or Customer Name in ONU Name
				nameMatched := (searchID != "" && strings.Contains(onuNameLower, strings.ToLower(searchID))) ||
					(searchName != "" && (strings.Contains(onuNameLower, strings.ToLower(searchName)) || strings.Contains(strings.ToLower(searchName), onuNameLower)))

				// If name matched or we have targetIP to verify
				if nameMatched {
					detail, err := u.zteClient.GetONUDetail(ctx, oltKey, b, p, onu.ONUID)
					if err == nil && detail != nil {
						snVal := detail.SerialNumber
						ponVal := p
						dt.Sn = &snVal
						if dt.Pon == nil || *dt.Pon == 0 {
							dt.Pon = &ponVal
						}
						_ = u.dataTeknisRepo.Update(ctx, dt)
						return detail, nil
					}
				}

				// Match by IP Address if targetIP is set
				if targetIP != "" {
					detail, err := u.zteClient.GetONUDetail(ctx, oltKey, b, p, onu.ONUID)
					if err == nil && detail != nil {
						if strings.TrimSpace(detail.IPAddress) == targetIP || strings.Contains(detail.Description, targetIP) || strings.Contains(detail.Name, targetIP) {
							snVal := detail.SerialNumber
							ponVal := p
							dt.Sn = &snVal
							if dt.Pon == nil || *dt.Pon == 0 {
								dt.Pon = &ponVal
							}
							_ = u.dataTeknisRepo.Update(ctx, dt)
							return detail, nil
						}
					}
				}
			}
		}
	}

	if targetIP != "" {
		return nil, fmt.Errorf("ONU dengan IP %s belum ditemukan di OLT %s. Pastikan modem online dan terhubung ke PON.", targetIP, oltKey)
	}
	return nil, errors.New("Serial number (SN) belum diisi pada Data Teknis pelanggan ini. Silakan klik Edit > Info ONU > 'Pilih dari OLT'.")
}

func (u *dataTeknisUsecase) SyncLiveOnuPower(ctx context.Context, id uint64, power float64) error {
	dt, err := u.dataTeknisRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("data teknis not found: %w", err)
	}

	rxInt := int(math.Round(power))
	dt.OnuPower = &rxInt

	return u.dataTeknisRepo.Update(ctx, dt)
}

func (u *dataTeknisUsecase) GetDetectedONUs(ctx context.Context, oltName string, pon int, board int) ([]domain.ZTEONUInfo, error) {
	olt, _ := u.findOLT(ctx, oltName)
	if olt != nil && u.isHSGQ(olt) {
		if u.hsgqClient == nil {
			return nil, errors.New("HSGQ OLT client service is not initialized")
		}
		user, pass := u.getHSGQCredentials(olt)
		return u.hsgqClient.GetONUs(ctx, olt.IPAddress, 23, user, pass, pon)
	}

	if olt != nil && u.isZTE(olt) {
		user, pass := u.getZTECredentials(olt)
		if u.zteTelnet != nil && user != "" {
			if board <= 0 {
				board = 1
			}
			data, err := u.zteTelnet.GetONUs(ctx, olt.IPAddress, 23, user, pass, board, pon)
			if err == nil {
				return data, nil
			}
		}
	}

	if u.zteClient == nil {
		return nil, errors.New("ZTE OLT client service is not initialized")
	}

	if board <= 0 {
		board = 1
	}
	oltKey := u.resolveOltKey(oltName)

	return u.zteClient.GetONUs(ctx, oltKey, board, pon)
}

func (u *dataTeknisUsecase) BulkSyncOLT(ctx context.Context, oltName string) (*domain.BulkSyncResult, error) {
	olt, _ := u.findOLT(ctx, oltName)
	targetOLTName := oltName
	if olt != nil && olt.NamaOlt != "" {
		targetOLTName = olt.NamaOlt
	}
	oltKey := u.resolveOltKey(targetOLTName)

	var scannedONUs []domain.ZTEONUInfo

	// 1. Scan ONUs based on OLT type
	if olt != nil && u.isHSGQ(olt) {
		if u.hsgqClient == nil {
			return nil, errors.New("layanan HSGQ OLT client belum diinisialisasi")
		}
		user, pass := u.getHSGQCredentials(olt)
		onus, err := u.hsgqClient.GetONUs(ctx, olt.IPAddress, 23, user, pass, 0)
		if err != nil {
			return nil, fmt.Errorf("gagal scan OLT HSGQ %s (%s): %w", olt.NamaOlt, olt.IPAddress, err)
		}
		scannedONUs = onus
	} else if olt != nil && u.isZTE(olt) && u.zteTelnet != nil {
		user, pass := u.getZTECredentials(olt)
		if user != "" {
			onus, err := u.zteTelnet.GetAllONUs(ctx, olt.IPAddress, 23, user, pass)
			if err == nil && len(onus) > 0 {
				scannedONUs = onus
			}
		}
		if len(scannedONUs) == 0 && u.zteClient != nil {
			for b := 1; b <= 2; b++ {
				for p := 1; p <= 16; p++ {
					onus, err := u.zteClient.GetONUs(ctx, oltKey, b, p)
					if err == nil && len(onus) > 0 {
						scannedONUs = append(scannedONUs, onus...)
					}
				}
			}
		}
	} else {
		if u.zteClient == nil {
			return nil, errors.New("layanan OLT client belum diinisialisasi")
		}
		for b := 1; b <= 2; b++ {
			for p := 1; p <= 16; p++ {
				onus, err := u.zteClient.GetONUs(ctx, oltKey, b, p)
				if err == nil && len(onus) > 0 {
					scannedONUs = append(scannedONUs, onus...)
				}
			}
		}
	}

	result := &domain.BulkSyncResult{
		OLTName:      targetOLTName,
		TotalScanned: len(scannedONUs),
		Details:      make([]domain.BulkSyncItem, 0),
	}

	if len(scannedONUs) == 0 {
		return result, nil
	}

	// 2. Fetch all data teknis records from database
	dtList, _, err := u.dataTeknisRepo.GetAll(ctx, 0, 10000, "", "", "", "", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data teknis: %w", err)
	}

	// 3. Filter candidates
	var candidates []*domain.DataTeknis
	for i := range dtList {
		dt := &dtList[i]
		if dt.Olt != nil && strings.TrimSpace(*dt.Olt) != "" {
			dtOltKey := u.resolveOltKey(*dt.Olt)
			if dtOltKey != oltKey && !strings.EqualFold(*dt.Olt, targetOLTName) && !strings.Contains(strings.ToLower(*dt.Olt), oltKey) {
				continue
			}
		}
		candidates = append(candidates, dt)
	}

	// 4. Pre-match in memory
	type matchPair struct {
		onu *domain.ZTEONUInfo
		dt  *domain.DataTeknis
	}
	var matchedPairs []matchPair
	matchedDTIDs := make(map[uint64]bool)

	for i := range scannedONUs {
		onu := &scannedONUs[i]
		for _, dt := range candidates {
			if matchedDTIDs[dt.ID] {
				continue
			}
			if matchONUToCustomer(onu, dt) {
				matchedDTIDs[dt.ID] = true
				matchedPairs = append(matchedPairs, matchPair{onu: onu, dt: dt})
				break
			}
		}
	}

	// 5. Batch-fetch optical power for HSGQ in ONE single Telnet session
	rxPowerMap := make(map[string]string)
	if olt != nil && u.isHSGQ(olt) && u.hsgqClient != nil && len(matchedPairs) > 0 {
		var targetsToQuery []domain.ZTEONUInfo
		for _, pair := range matchedPairs {
			if (pair.onu.RxPower == "" || pair.onu.RxPower == "-") && (strings.EqualFold(pair.onu.Status, "Online") || strings.EqualFold(pair.onu.Status, "working")) {
				targetsToQuery = append(targetsToQuery, *pair.onu)
			}
		}
		if len(targetsToQuery) > 0 {
			user, pass := u.getHSGQCredentials(olt)
			batchPowers, err := u.hsgqClient.GetBatchOpticalPower(ctx, olt.IPAddress, 23, user, pass, targetsToQuery)
			if err == nil {
				rxPowerMap = batchPowers
			}
		}
	}

	// 6. Update database for all matched records
	for _, pair := range matchedPairs {
		onu := pair.onu
		matchedDT := pair.dt
		result.TotalMatched++

		snToUse := strings.TrimSpace(onu.SerialNumber)
		rxPowerToUse := strings.TrimSpace(onu.RxPower)
		statusToUse := onu.Status

		if rxPowerToUse == "" || rxPowerToUse == "-" {
			key := fmt.Sprintf("%d:%d", onu.PON, onu.ONUID)
			if pwr, ok := rxPowerMap[key]; ok && pwr != "" {
				rxPowerToUse = pwr
			}
		}

		needsUpdate := false

		// Check OLT
		if matchedDT.Olt == nil || *matchedDT.Olt == "" || !strings.EqualFold(*matchedDT.Olt, targetOLTName) {
			matchedDT.Olt = &targetOLTName
			needsUpdate = true
		}

		// Check PON
		if matchedDT.Pon == nil || *matchedDT.Pon != onu.PON {
			ponVal := onu.PON
			matchedDT.Pon = &ponVal
			needsUpdate = true
		}

		// Check SN
		if snToUse != "" && (matchedDT.Sn == nil || *matchedDT.Sn == "" || *matchedDT.Sn != snToUse) {
			matchedDT.Sn = &snToUse
			needsUpdate = true
		}

		// Check Redaman (OnuPower)
		if rxPowerToUse != "" {
			if pwrInt := parseRxPowerToInt(rxPowerToUse); pwrInt != nil {
				if matchedDT.OnuPower == nil || *matchedDT.OnuPower != *pwrInt {
					matchedDT.OnuPower = pwrInt
					needsUpdate = true
				}
			}
		}

		if needsUpdate {
			if err := u.dataTeknisRepo.Update(ctx, matchedDT); err == nil {
				result.UpdatedCount++
			}
		}

		custName := ""
		if matchedDT.Pelanggan != nil {
			custName = matchedDT.Pelanggan.Nama
		}
		itemIP := ""
		if matchedDT.IPPelanggan != nil {
			itemIP = *matchedDT.IPPelanggan
		}

		result.Details = append(result.Details, domain.BulkSyncItem{
			IDPelanggan: matchedDT.IDPelanggan,
			Nama:        custName,
			IP:          itemIP,
			PON:         onu.PON,
			Board:       onu.Board,
			ONUID:       onu.ONUID,
			SN:          snToUse,
			RxPower:     rxPowerToUse,
			Status:      statusToUse,
		})
	}

	return result, nil
}
