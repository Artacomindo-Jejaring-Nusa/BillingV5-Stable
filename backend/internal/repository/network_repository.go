package repository

import (
	"context"
	"errors"
	"strings"

	"billing-backend/internal/domain"

	"gorm.io/gorm"
)

type mikrotikRepository struct {
	db *gorm.DB
}

func NewMikrotikRepository(db *gorm.DB) domain.MikrotikRepository {
	return &mikrotikRepository{db: db}
}

func (r *mikrotikRepository) GetAll(ctx context.Context) ([]domain.MikrotikServer, error) {
	var servers []domain.MikrotikServer
	err := r.db.WithContext(ctx).Order("id desc").Find(&servers).Error
	return servers, err
}

func (r *mikrotikRepository) GetByID(ctx context.Context, id uint64) (*domain.MikrotikServer, error) {
	var server domain.MikrotikServer
	err := r.db.WithContext(ctx).First(&server, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("mikrotik server not found")
		}
		return nil, err
	}
	return &server, nil
}

func (r *mikrotikRepository) Create(ctx context.Context, server *domain.MikrotikServer) error {
	return r.db.WithContext(ctx).Create(server).Error
}

func (r *mikrotikRepository) Update(ctx context.Context, server *domain.MikrotikServer) error {
	return r.db.WithContext(ctx).Save(server).Error
}

func (r *mikrotikRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.MikrotikServer{}, id).Error
}

func (r *mikrotikRepository) GetByName(ctx context.Context, name string) (*domain.MikrotikServer, error) {
	var server domain.MikrotikServer
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&server).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &server, nil
}

func (r *mikrotikRepository) GetByNames(ctx context.Context, names []string) ([]domain.MikrotikServer, error) {
	var list []domain.MikrotikServer
	err := r.db.WithContext(ctx).Where("name IN ?", names).Find(&list).Error
	return list, err
}

type oltRepository struct {
	db *gorm.DB
}

func NewOLTRepository(db *gorm.DB) domain.OLTRepository {
	return &oltRepository{db: db}
}

func (r *oltRepository) GetAll(ctx context.Context) ([]domain.OLT, error) {
	var olts []domain.OLT
	err := r.db.WithContext(ctx).Preload("MikrotikServer").Order("id desc").Find(&olts).Error
	return olts, err
}

func (r *oltRepository) GetByID(ctx context.Context, id uint64) (*domain.OLT, error) {
	var olt domain.OLT
	err := r.db.WithContext(ctx).Preload("MikrotikServer").First(&olt, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("OLT not found")
		}
		return nil, err
	}
	return &olt, nil
}

func (r *oltRepository) Create(ctx context.Context, olt *domain.OLT) error {
	return r.db.WithContext(ctx).Create(olt).Error
}

func (r *oltRepository) Update(ctx context.Context, olt *domain.OLT) error {
	return r.db.WithContext(ctx).Save(olt).Error
}

func (r *oltRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.OLT{}, id).Error
}

func normalizeDeviceIdentifier(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, " ", "")
	return s
}

func (r *oltRepository) GetCustomerDeviceInfoMap(ctx context.Context) (map[string]domain.CustomerDeviceInfo, []domain.CustomerDeviceInfo, error) {
	type rowResult struct {
		IDPelanggan *string `gorm:"column:id_pelanggan"`
		SN          *string `gorm:"column:sn"`
		Nama        *string `gorm:"column:nama"`
	}
	var rows []rowResult
	err := r.db.WithContext(ctx).Table("data_teknis dt").
		Select("dt.id_pelanggan, dt.sn, p.nama").
		Joins("LEFT JOIN pelanggan p ON p.id = dt.pelanggan_id AND p.deleted_at IS NULL").
		Where("dt.id_pelanggan IS NOT NULL OR dt.sn IS NOT NULL").
		Find(&rows).Error
	if err != nil {
		return nil, nil, err
	}

	snMap := make(map[string]domain.CustomerDeviceInfo, len(rows))
	var allList []domain.CustomerDeviceInfo

	for _, row := range rows {
		idPelanggan := ""
		if row.IDPelanggan != nil {
			idPelanggan = strings.TrimSpace(*row.IDPelanggan)
		}
		sn := ""
		if row.SN != nil {
			sn = strings.TrimSpace(*row.SN)
		}
		nama := ""
		if row.Nama != nil {
			nama = strings.TrimSpace(*row.Nama)
		}

		info := domain.CustomerDeviceInfo{
			IDPelanggan: idPelanggan,
			Nama:        nama,
			SN:          sn,
		}
		allList = append(allList, info)

		if sn != "" {
			normSN := normalizeDeviceIdentifier(sn)
			if normSN != "" {
				snMap[normSN] = info
			}
		}
	}

	return snMap, allList, nil
}

// ODC Repository implementation
type odcRepository struct {
	db *gorm.DB
}

func NewODCRepository(db *gorm.DB) domain.ODCRepository {
	return &odcRepository{db: db}
}

func (r *odcRepository) GetAll(ctx context.Context) ([]domain.ODC, error) {
	var odcs []domain.ODC
	err := r.db.WithContext(ctx).Preload("Olt").Order("id desc").Find(&odcs).Error
	return odcs, err
}

func (r *odcRepository) GetByID(ctx context.Context, id uint64) (*domain.ODC, error) {
	var odc domain.ODC
	err := r.db.WithContext(ctx).Preload("Olt").First(&odc, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("ODC not found")
		}
		return nil, err
	}
	return &odc, nil
}

func (r *odcRepository) Create(ctx context.Context, odc *domain.ODC) error {
	return r.db.WithContext(ctx).Create(odc).Error
}

func (r *odcRepository) Update(ctx context.Context, odc *domain.ODC) error {
	return r.db.WithContext(ctx).Save(odc).Error
}

func (r *odcRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.ODC{}, id).Error
}

// ODP Repository implementation
type odpRepository struct {
	db *gorm.DB
}

func NewODPRepository(db *gorm.DB) domain.ODPRepository {
	return &odpRepository{db: db}
}

func (r *odpRepository) GetAll(ctx context.Context) ([]domain.ODP, error) {
	var odps []domain.ODP
	err := r.db.WithContext(ctx).
		Preload("Olt").
		Preload("Odc").
		Preload("ParentOdp").
		Order("id desc").
		Find(&odps).Error
	if err != nil {
		return nil, err
	}

	// Calculate PortTerpakai for each ODP with single batch GROUP BY query (prevents N+1)
	if len(odps) > 0 {
		type ODPUsage struct {
			ODPID uint64 `gorm:"column:odp_id"`
			Count int64  `gorm:"column:count"`
		}
		var usages []ODPUsage
		var odpIDs []uint64
		for _, odp := range odps {
			odpIDs = append(odpIDs, odp.ID)
		}
		_ = r.db.WithContext(ctx).Model(&domain.DataTeknis{}).
			Select("odp_id, COUNT(id) as count").
			Where("odp_id IN (?)", odpIDs).
			Group("odp_id").
			Scan(&usages).Error

		usageMap := make(map[uint64]int)
		for _, u := range usages {
			usageMap[u.ODPID] = int(u.Count)
		}
		for i := range odps {
			odps[i].PortTerpakai = usageMap[odps[i].ID]
		}
	}

	return odps, nil
}

func (r *odpRepository) GetByID(ctx context.Context, id uint64) (*domain.ODP, error) {
	var odp domain.ODP
	err := r.db.WithContext(ctx).Preload("Olt").Preload("Odc").Preload("ParentOdp").First(&odp, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("ODP not found")
		}
		return nil, err
	}

	var count int64
	r.db.WithContext(ctx).Model(&domain.DataTeknis{}).Where("odp_id = ?", odp.ID).Count(&count)
	odp.PortTerpakai = int(count)

	return &odp, nil
}

func (r *odpRepository) Create(ctx context.Context, odp *domain.ODP) error {
	return r.db.WithContext(ctx).Create(odp).Error
}

func (r *odpRepository) Update(ctx context.Context, odp *domain.ODP) error {
	return r.db.WithContext(ctx).Save(odp).Error
}

func (r *odpRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.ODP{}, id).Error
}

func (r *odpRepository) BatchUpsert(ctx context.Context, odps []domain.ODP) (int, error) {
	saved := 0
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range odps {
			var existing domain.ODP
			err := tx.Where("kode_odp = ?", item.KodeOdp).First(&existing).Error
			if err == nil {
				// Update existing coordinates & info
				if item.Alamat != nil && *item.Alamat != "" {
					existing.Alamat = item.Alamat
				}
				if item.Latitude != nil {
					existing.Latitude = item.Latitude
				}
				if item.Longitude != nil {
					existing.Longitude = item.Longitude
				}
				if item.KapasitasPort > 0 {
					existing.KapasitasPort = item.KapasitasPort
				}
				if item.OltID > 0 {
					existing.OltID = item.OltID
				}
				if item.OdcID != nil {
					existing.OdcID = item.OdcID
				}
				if err := tx.Save(&existing).Error; err != nil {
					return err
				}
				saved++
			} else if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create new ODP
				newItem := item
				if err := tx.Create(&newItem).Error; err != nil {
					return err
				}
				saved++
			} else {
				return err
			}
		}
		return nil
	})
	return saved, err
}

func (r *odpRepository) UpdateCoordinates(ctx context.Context, id uint64, lat, lng float64) error {
	return r.db.WithContext(ctx).Model(&domain.ODP{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"latitude":  lat,
			"longitude": lng,
		}).Error
}

func (r *odpRepository) UpdateNodeCoordinate(ctx context.Context, nodeType string, id uint64, lat, lng float64) error {
	switch nodeType {
	case "server":
		return r.db.WithContext(ctx).Model(&domain.MikrotikServer{}).Where("id = ?", id).
			Updates(map[string]interface{}{"latitude": lat, "longitude": lng}).Error
	case "olt":
		return r.db.WithContext(ctx).Model(&domain.OLT{}).Where("id = ?", id).
			Updates(map[string]interface{}{"latitude": lat, "longitude": lng}).Error
	case "odc":
		return r.db.WithContext(ctx).Model(&domain.ODC{}).Where("id = ?", id).
			Updates(map[string]interface{}{"latitude": lat, "longitude": lng}).Error
	case "odp":
		return r.db.WithContext(ctx).Model(&domain.ODP{}).Where("id = ?", id).
			Updates(map[string]interface{}{"latitude": lat, "longitude": lng}).Error
	case "pelanggan":
		return r.db.WithContext(ctx).Model(&domain.Pelanggan{}).Where("id = ?", id).
			Updates(map[string]interface{}{"latitude": lat, "longitude": lng}).Error
	default:
		return errors.New("invalid node type for coordinate update")
	}
}

func (r *odpRepository) GetNetworkMapData(ctx context.Context) (*domain.NetworkMapData, error) {
	var servers []domain.MikrotikServer
	_ = r.db.WithContext(ctx).Order("name asc").Find(&servers).Error
	for i := range servers {
		servers[i].Password = "" // Hide sensitive password
	}

	var olts []domain.OLT
	_ = r.db.WithContext(ctx).Order("nama_olt asc").Find(&olts).Error
	for i := range olts {
		olts[i].Password = nil
	}

	var odcs []domain.ODC
	_ = r.db.WithContext(ctx).Order("nama_odc asc").Find(&odcs).Error

	odps, err := r.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Load customers with data teknis & status
	type PelangganMapRow struct {
		ID            uint64   `gorm:"column:id"`
		Nama          string   `gorm:"column:nama"`
		Alamat        string   `gorm:"column:alamat"`
		NoTelp        string   `gorm:"column:no_telp"`
		Latitude      *float64 `gorm:"column:latitude"`
		Longitude     *float64 `gorm:"column:longitude"`
		OdpID         *uint64  `gorm:"column:odp_id"`
		KodeOdp       *string  `gorm:"column:kode_odp"`
		PortOdp       *int     `gorm:"column:port_odp"`
		UsernamePppoe *string  `gorm:"column:id_pelanggan"`
		Status        *string  `gorm:"column:status"`
	}
	var rows []PelangganMapRow
	_ = r.db.WithContext(ctx).Table("pelanggan p").
		Select(`p.id, p.nama, p.alamat, p.no_telp, p.latitude, p.longitude,
				dt.odp_id, odp.kode_odp, dt.port_odp, dt.id_pelanggan,
				COALESCE(l.status, 'Unknown') as status`).
		Joins("LEFT JOIN data_teknis dt ON dt.pelanggan_id = p.id").
		Joins("LEFT JOIN odp ON odp.id = dt.odp_id").
		Joins("LEFT JOIN langganan l ON l.pelanggan_id = p.id AND l.deleted_at IS NULL").
		Where("p.deleted_at IS NULL").
		Find(&rows).Error

	var pelanggans []domain.NetworkMapPelanggan
	for _, row := range rows {
		kodeOdp := ""
		if row.KodeOdp != nil {
			kodeOdp = *row.KodeOdp
		}
		pppoe := ""
		if row.UsernamePppoe != nil {
			pppoe = *row.UsernamePppoe
		}
		status := "Aktif"
		if row.Status != nil && *row.Status != "" {
			status = *row.Status
		}

		pelanggans = append(pelanggans, domain.NetworkMapPelanggan{
			ID:            row.ID,
			Nama:          row.Nama,
			Alamat:        row.Alamat,
			NoTelp:        row.NoTelp,
			Latitude:      row.Latitude,
			Longitude:     row.Longitude,
			OdpID:         row.OdpID,
			KodeOdp:       kodeOdp,
			PortOdp:       row.PortOdp,
			UsernamePppoe: pppoe,
			Status:        status,
		})
	}

	return &domain.NetworkMapData{
		Servers:    servers,
		Olts:       olts,
		Odcs:       odcs,
		Odps:       odps,
		Pelanggans: pelanggans,
	}, nil
}
