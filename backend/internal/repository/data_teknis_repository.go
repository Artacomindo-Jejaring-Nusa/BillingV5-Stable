package repository

import (
	"context"
	"errors"

	"billing-backend/internal/domain"

	"gorm.io/gorm"
)

type dataTeknisRepository struct {
	db *gorm.DB
}

func NewDataTeknisRepository(db *gorm.DB) domain.DataTeknisRepository {
	return &dataTeknisRepository{db: db}
}

func (r *dataTeknisRepository) GetAll(ctx context.Context, skip, limit int, search string, olt string, profile string, vlan string, onuPowerMin, onuPowerMax *int) ([]domain.DataTeknis, int64, error) {
	var list []domain.DataTeknis
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.DataTeknis{})

	if search != "" {
		searchTerm := "%" + search + "%"
		db = db.Joins("LEFT JOIN pelanggan ON pelanggan.id = data_teknis.pelanggan_id").
			Where("pelanggan.nama LIKE ? OR data_teknis.id_pelanggan LIKE ? OR data_teknis.ip_pelanggan LIKE ? OR data_teknis.sn LIKE ?",
				searchTerm, searchTerm, searchTerm, searchTerm)
	}

	if olt != "" && olt != "Semua" {
		db = db.Where("data_teknis.olt = ?", olt)
	}

	if profile != "" && profile != "Semua" {
		db = db.Where("data_teknis.profile_pppoe = ?", profile)
	}

	if vlan != "" && vlan != "Semua" {
		db = db.Where("data_teknis.id_vlan = ?", vlan)
	}

	if onuPowerMin != nil {
		db = db.Where("data_teknis.onu_power >= ?", *onuPowerMin)
	}

	if onuPowerMax != nil {
		db = db.Where("data_teknis.onu_power <= ?", *onuPowerMax)
	}

	// Count total
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch page with preload relations
	err := db.Preload("Pelanggan").
		Preload("MikrotikServer").
		Preload("Odp").
		Order("data_teknis.id desc").
		Offset(skip).
		Limit(limit).
		Find(&list).Error

	return list, total, err
}

func (r *dataTeknisRepository) GetByID(ctx context.Context, id uint64) (*domain.DataTeknis, error) {
	var data domain.DataTeknis
	err := r.db.WithContext(ctx).
		Preload("Pelanggan").
		Preload("MikrotikServer").
		Preload("Odp").
		First(&data, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("data teknis not found")
		}
		return nil, err
	}
	return &data, nil
}

func (r *dataTeknisRepository) GetByPelangganID(ctx context.Context, pelangganID uint64) (*domain.DataTeknis, error) {
	var data domain.DataTeknis
	err := r.db.WithContext(ctx).
		Preload("Pelanggan").
		Preload("MikrotikServer").
		Preload("Odp").
		Where("pelanggan_id = ?", pelangganID).
		First(&data).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("data teknis not found for this customer")
		}
		return nil, err
	}
	return &data, nil
}

func (r *dataTeknisRepository) Create(ctx context.Context, data *domain.DataTeknis) error {
	return r.db.WithContext(ctx).Create(data).Error
}

func (r *dataTeknisRepository) Update(ctx context.Context, data *domain.DataTeknis) error {
	return r.db.WithContext(ctx).Save(data).Error
}

func (r *dataTeknisRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.DataTeknis{}, id).Error
}

func (r *dataTeknisRepository) GetAvailableOLT(ctx context.Context) ([]string, error) {
	var olts []string
	err := r.db.WithContext(ctx).Model(&domain.DataTeknis{}).
		Where("olt IS NOT NULL AND olt != ''").
		Order("olt").
		Pluck("DISTINCT olt", &olts).Error
	return olts, err
}

func (r *dataTeknisRepository) GetAvailableProfiles(ctx context.Context) ([]string, error) {
	var profiles []string
	err := r.db.WithContext(ctx).Model(&domain.DataTeknis{}).
		Where("profile_pppoe IS NOT NULL AND profile_pppoe != ''").
		Order("profile_pppoe").
		Pluck("DISTINCT profile_pppoe", &profiles).Error
	return profiles, err
}

func (r *dataTeknisRepository) GetAvailableVlans(ctx context.Context) ([]string, error) {
	var vlans []string
	err := r.db.WithContext(ctx).Model(&domain.DataTeknis{}).
		Where("id_vlan IS NOT NULL AND id_vlan != ''").
		Order("id_vlan").
		Pluck("DISTINCT id_vlan", &vlans).Error
	return vlans, err
}

func (r *dataTeknisRepository) GetOnuPowerRanges(ctx context.Context) (*int, *int, error) {
	type Result struct {
		Min *int
		Max *int
	}
	var res Result
	err := r.db.WithContext(ctx).Model(&domain.DataTeknis{}).
		Select("MIN(onu_power) as min, MAX(onu_power) as max").
		Scan(&res).Error
	if err != nil {
		return nil, nil, err
	}
	return res.Min, res.Max, nil
}

func (r *dataTeknisRepository) CheckIPAddress(ctx context.Context, ip string, excludeID *uint64) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&domain.DataTeknis{}).Where("ip_pelanggan = ?", ip)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

func (r *dataTeknisRepository) GetOdpByCode(ctx context.Context, code string) (*domain.ODP, error) {
	var odp domain.ODP
	err := r.db.WithContext(ctx).Where("kode_odp = ?", code).First(&odp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &odp, nil
}

func (r *dataTeknisRepository) GetOdpByCodes(ctx context.Context, codes []string) ([]domain.ODP, error) {
	var list []domain.ODP
	err := r.db.WithContext(ctx).Where("kode_odp IN ?", codes).Find(&list).Error
	return list, err
}

func (r *dataTeknisRepository) GetPendingSync(ctx context.Context) ([]domain.DataTeknis, error) {
	var list []domain.DataTeknis
	err := r.db.WithContext(ctx).
		Preload("Pelanggan").
		Preload("Pelanggan.Langganan").
		Preload("MikrotikServer").
		Where("mikrotik_sync_pending = ?", true).
		Find(&list).Error
	return list, err
}
