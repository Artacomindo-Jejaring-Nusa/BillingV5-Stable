package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"billing-backend/internal/domain"

	"gorm.io/gorm"
)

type pelangganRepository struct {
	db *gorm.DB
}

func NewPelangganRepository(db *gorm.DB) domain.PelangganRepository {
	return &pelangganRepository{db: db}
}

func (r *pelangganRepository) GetAll(ctx context.Context, limit, offset int, search, connectionStatus string) ([]domain.Pelanggan, int64, error) {
	var pelanggans []domain.Pelanggan
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Pelanggan{})

	if connectionStatus == "unconfigured" {
		query = query.
			Joins("LEFT JOIN data_teknis ON data_teknis.pelanggan_id = pelanggan.id").
			Where("data_teknis.id IS NULL")
	}

	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where("pelanggan.nama LIKE ? OR pelanggan.no_telp LIKE ? OR pelanggan.email LIKE ?", searchTerm, searchTerm, searchTerm)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch with limit and offset, preload relationships
	fetchQuery := query.
		Select("pelanggan.*").
		Preload("DataTeknis").
		Preload("MikrotikServer").
		Preload("Langganan")

	if limit > 0 {
		fetchQuery = fetchQuery.Limit(limit)
	}
	if offset > 0 {
		fetchQuery = fetchQuery.Offset(offset)
	}

	err := fetchQuery.
		Order("pelanggan.id desc").
		Find(&pelanggans).Error

	if err != nil {
		return nil, 0, err
	}

	return pelanggans, total, nil
}

func (r *pelangganRepository) GetByID(ctx context.Context, id uint64) (*domain.Pelanggan, error) {
	var pelanggan domain.Pelanggan
	err := r.db.WithContext(ctx).
		Preload("DataTeknis").
		Preload("MikrotikServer").
		Preload("Langganan").
		First(&pelanggan, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("pelanggan not found")
		}
		return nil, err
	}
	return &pelanggan, nil
}

func (r *pelangganRepository) Create(ctx context.Context, pelanggan *domain.Pelanggan) error {
	return r.db.WithContext(ctx).Create(pelanggan).Error
}

func (r *pelangganRepository) Update(ctx context.Context, pelanggan *domain.Pelanggan) error {
	// Omit associations to prevent accidental overwriting of nested objects during save
	return r.db.WithContext(ctx).Omit("DataTeknis", "MikrotikServer", "Langganan", "Invoices").Save(pelanggan).Error
}

func (r *pelangganRepository) Delete(ctx context.Context, id uint64) error {
	var pelanggan domain.Pelanggan
	if err := r.db.WithContext(ctx).First(&pelanggan, id).Error; err != nil {
		return err
	}

	// Modify email and NoKtp to release unique constraints
	now := time.Now().Unix()
	pelanggan.Email = fmt.Sprintf("deleted_%d_%s", now, pelanggan.Email)
	if pelanggan.NoKtp != "" {
		pelanggan.NoKtp = fmt.Sprintf("deleted_%d_%s", now, pelanggan.NoKtp)
	}

	// Save modified fields to release constraints
	if err := r.db.WithContext(ctx).Model(&pelanggan).Omit("DataTeknis", "MikrotikServer", "Langganan", "Invoices").Updates(map[string]interface{}{
		"email":  pelanggan.Email,
		"no_ktp": pelanggan.NoKtp,
	}).Error; err != nil {
		return err
	}

	return r.db.WithContext(ctx).Delete(&domain.Pelanggan{}, id).Error
}

func (r *pelangganRepository) GetByEmail(ctx context.Context, email string) (*domain.Pelanggan, error) {
	var pelanggan domain.Pelanggan
	err := r.db.WithContext(ctx).Unscoped().Preload("Langganan").Where("email = ?", email).First(&pelanggan).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &pelanggan, nil
}

func (r *pelangganRepository) GetByEmails(ctx context.Context, emails []string) ([]domain.Pelanggan, error) {
	var list []domain.Pelanggan
	err := r.db.WithContext(ctx).Preload("Langganan").Where("email IN ?", emails).Find(&list).Error
	return list, err
}

func (r *pelangganRepository) GetByNoKtp(ctx context.Context, noKtp string) (*domain.Pelanggan, error) {
	var pelanggan domain.Pelanggan
	err := r.db.WithContext(ctx).Where("no_ktp = ?", noKtp).First(&pelanggan).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &pelanggan, nil
}

func (r *pelangganRepository) GetUniqueLocations(ctx context.Context) ([]string, error) {
	var locations []string
	err := r.db.WithContext(ctx).Model(&domain.Pelanggan{}).
		Where("alamat IS NOT NULL AND alamat != ''").
		Distinct("alamat").
		Order("alamat asc").
		Pluck("alamat", &locations).Error
	return locations, err
}

