package repository

import (
	"context"
	"errors"
	"strings"

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

	dbCount := r.db.WithContext(ctx).Model(&domain.DataTeknis{}).
		Joins("LEFT JOIN pelanggan ON pelanggan.id = data_teknis.pelanggan_id AND pelanggan.deleted_at IS NULL")

	dbFind := r.db.WithContext(ctx).
		Select("data_teknis.*").
		Preload("Pelanggan").
		Preload("MikrotikServer").
		Preload("Odp").
		Joins("LEFT JOIN pelanggan ON pelanggan.id = data_teknis.pelanggan_id AND pelanggan.deleted_at IS NULL")

	if search != "" {
		words := strings.Fields(search)
		if len(words) == 1 {
			searchTerm := "%" + search + "%"
			condition := "pelanggan.nama LIKE ? OR pelanggan.customer_id LIKE ? OR data_teknis.id_pelanggan LIKE ? OR data_teknis.ip_pelanggan LIKE ? OR data_teknis.sn LIKE ? OR data_teknis.olt LIKE ? OR data_teknis.profile_pppoe LIKE ? OR pelanggan.no_telp LIKE ? OR pelanggan.alamat LIKE ?"
			dbCount = dbCount.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
			dbFind = dbFind.Where(condition, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
		} else {
			for _, word := range words {
				lowerWord := strings.ToLower(word)
				if lowerWord == "blok" || lowerWord == "unit" || lowerWord == "no" || lowerWord == "nomor" || lowerWord == "rt" || lowerWord == "rw" {
					continue
				}
				wordTerm := "%" + word + "%"
				condition := "(pelanggan.nama LIKE ? OR pelanggan.customer_id LIKE ? OR data_teknis.id_pelanggan LIKE ? OR data_teknis.ip_pelanggan LIKE ? OR data_teknis.sn LIKE ? OR data_teknis.olt LIKE ? OR data_teknis.profile_pppoe LIKE ? OR pelanggan.no_telp LIKE ? OR pelanggan.alamat LIKE ?)"
				dbCount = dbCount.Where(condition, wordTerm, wordTerm, wordTerm, wordTerm, wordTerm, wordTerm, wordTerm, wordTerm, wordTerm)
				dbFind = dbFind.Where(condition, wordTerm, wordTerm, wordTerm, wordTerm, wordTerm, wordTerm, wordTerm, wordTerm, wordTerm)
			}
		}
	}

	if olt != "" && olt != "Semua" && olt != "Semua OLT" {
		dbCount = dbCount.Where("data_teknis.olt = ?", olt)
		dbFind = dbFind.Where("data_teknis.olt = ?", olt)
	}

	if profile != "" && profile != "Semua" && profile != "Semua Profile" {
		dbCount = dbCount.Where("data_teknis.profile_pppoe = ?", profile)
		dbFind = dbFind.Where("data_teknis.profile_pppoe = ?", profile)
	}

	if vlan != "" && vlan != "Semua" && vlan != "Semua VLAN" {
		dbCount = dbCount.Where("data_teknis.id_vlan = ?", vlan)
		dbFind = dbFind.Where("data_teknis.id_vlan = ?", vlan)
	}

	if onuPowerMin != nil {
		dbCount = dbCount.Where("data_teknis.onu_power >= ?", *onuPowerMin)
		dbFind = dbFind.Where("data_teknis.onu_power >= ?", *onuPowerMin)
	}

	if onuPowerMax != nil {
		dbCount = dbCount.Where("data_teknis.onu_power <= ?", *onuPowerMax)
		dbFind = dbFind.Where("data_teknis.onu_power <= ?", *onuPowerMax)
	}

	// Count total
	if err := dbCount.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch page with preload relations
	err := dbFind.
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

func (r *dataTeknisRepository) GetUnconfiguredPelanggan(ctx context.Context, search string) ([]domain.Pelanggan, error) {
	var list []domain.Pelanggan
	query := r.db.WithContext(ctx).Table("pelanggan").
		Select("pelanggan.id, pelanggan.nama, pelanggan.alamat, pelanggan.alamat_2, pelanggan.created_at").
		Joins("LEFT JOIN data_teknis ON data_teknis.pelanggan_id = pelanggan.id AND data_teknis.deleted_at IS NULL").
		Where("pelanggan.deleted_at IS NULL AND data_teknis.id IS NULL")

	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where("(pelanggan.nama LIKE ? OR pelanggan.customer_id LIKE ? OR pelanggan.alamat LIKE ? OR pelanggan.no_telp LIKE ?)", searchTerm, searchTerm, searchTerm, searchTerm)
	}

	err := query.Order("pelanggan.id desc").Limit(500).Find(&list).Error
	return list, err
}
