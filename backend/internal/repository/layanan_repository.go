package repository

import (
	"context"
	"errors"
	"time"

	"billing-backend/internal/domain"

	"gorm.io/gorm"
)

type layananRepository struct {
	db *gorm.DB
}

// NewLayananRepository creates repositories for Layanan sub-module
func NewLayananRepository(db *gorm.DB) *layananRepository {
	return &layananRepository{db: db}
}

// --- HargaLayananRepository Implementation ---

func (r *layananRepository) CreateHargaLayanan(ctx context.Context, brand *domain.HargaLayanan) error {
	return r.db.WithContext(ctx).Create(brand).Error
}

func (r *layananRepository) GetAllHargaLayanan(ctx context.Context) ([]domain.HargaLayanan, error) {
	var brands []domain.HargaLayanan
	err := r.db.WithContext(ctx).Find(&brands).Error
	return brands, err
}

func (r *layananRepository) GetHargaLayananByID(ctx context.Context, idBrand string) (*domain.HargaLayanan, error) {
	var brand domain.HargaLayanan
	err := r.db.WithContext(ctx).Where("id_brand = ?", idBrand).First(&brand).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &brand, nil
}

func (r *layananRepository) UpdateHargaLayanan(ctx context.Context, brand *domain.HargaLayanan) error {
	return r.db.WithContext(ctx).Save(brand).Error
}

func (r *layananRepository) DeleteHargaLayanan(ctx context.Context, idBrand string) error {
	return r.db.WithContext(ctx).Where("id_brand = ?", idBrand).Delete(&domain.HargaLayanan{}).Error
}

// --- PaketLayananRepository Implementation ---

func (r *layananRepository) CreatePaketLayanan(ctx context.Context, paket *domain.PaketLayanan) error {
	return r.db.WithContext(ctx).Create(paket).Error
}

func (r *layananRepository) GetAllPaketLayanan(ctx context.Context) ([]domain.PaketLayanan, error) {
	var pakets []domain.PaketLayanan
	err := r.db.WithContext(ctx).Preload("HargaLayanan").Find(&pakets).Error
	return pakets, err
}

func (r *layananRepository) GetPaketLayananByID(ctx context.Context, id uint64) (*domain.PaketLayanan, error) {
	var paket domain.PaketLayanan
	err := r.db.WithContext(ctx).Preload("HargaLayanan").First(&paket, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &paket, nil
}

func (r *layananRepository) UpdatePaketLayanan(ctx context.Context, paket *domain.PaketLayanan) error {
	return r.db.WithContext(ctx).Save(paket).Error
}

func (r *layananRepository) DeletePaketLayanan(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.PaketLayanan{}, id).Error
}

// --- DiskonRepository Implementation ---

func (r *layananRepository) CreateDiskon(ctx context.Context, diskon *domain.Diskon) error {
	return r.db.WithContext(ctx).Create(diskon).Error
}

func (r *layananRepository) GetAllDiskon(ctx context.Context, limit, offset int, cluster string, isActive *bool) ([]domain.Diskon, int64, error) {
	var diskons []domain.Diskon
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Diskon{})

	if cluster != "" {
		query = query.Where("cluster LIKE ?", "%"+cluster+"%")
	}

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&diskons).Error
	if err != nil {
		return nil, 0, err
	}

	return diskons, total, nil
}

func (r *layananRepository) GetDiskonByID(ctx context.Context, id uint64) (*domain.Diskon, error) {
	var diskon domain.Diskon
	err := r.db.WithContext(ctx).First(&diskon, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &diskon, nil
}

func (r *layananRepository) GetActiveForCluster(ctx context.Context, cluster string, checkDate time.Time) (*domain.Diskon, error) {
	var diskon domain.Diskon
	// Get diskon with highest percentage active for cluster on given date
	err := r.db.WithContext(ctx).
		Where("cluster = ? AND is_active = ?", cluster, true).
		Where("(tgl_mulai IS NULL OR tgl_mulai <= ?)", checkDate).
		Where("(tgl_selesai IS NULL OR tgl_selesai >= ?)", checkDate).
		Order("persentase_diskon DESC").
		First(&diskon).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &diskon, nil
}

func (r *layananRepository) UpdateDiskon(ctx context.Context, diskon *domain.Diskon) error {
	return r.db.WithContext(ctx).Save(diskon).Error
}

func (r *layananRepository) DeleteDiskon(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.Diskon{}, id).Error
}

func (r *layananRepository) GetClusterList(ctx context.Context) ([]string, error) {
	var clusters []string
	err := r.db.WithContext(ctx).
		Model(&domain.Pelanggan{}).
		Distinct("alamat").
		Where("alamat IS NOT NULL AND alamat != ''").
		Order("alamat").
		Pluck("alamat", &clusters).Error
	return clusters, err
}

// --- Bridge functions to fit the separated interfaces in domain ---

// Wrapper struct for HargaLayanan repository
type hargaLayananRepositoryBridge struct {
	repo *layananRepository
}

func NewHargaLayananRepository(db *gorm.DB) domain.HargaLayananRepository {
	return &hargaLayananRepositoryBridge{repo: NewLayananRepository(db)}
}

func (b *hargaLayananRepositoryBridge) Create(ctx context.Context, brand *domain.HargaLayanan) error {
	return b.repo.CreateHargaLayanan(ctx, brand)
}

func (b *hargaLayananRepositoryBridge) GetAll(ctx context.Context) ([]domain.HargaLayanan, error) {
	return b.repo.GetAllHargaLayanan(ctx)
}

func (b *hargaLayananRepositoryBridge) GetByID(ctx context.Context, idBrand string) (*domain.HargaLayanan, error) {
	return b.repo.GetHargaLayananByID(ctx, idBrand)
}

func (b *hargaLayananRepositoryBridge) Update(ctx context.Context, brand *domain.HargaLayanan) error {
	return b.repo.UpdateHargaLayanan(ctx, brand)
}

func (b *hargaLayananRepositoryBridge) Delete(ctx context.Context, idBrand string) error {
	return b.repo.DeleteHargaLayanan(ctx, idBrand)
}

// Wrapper struct for PaketLayanan repository
type paketLayananRepositoryBridge struct {
	repo *layananRepository
}

func NewPaketLayananRepository(db *gorm.DB) domain.PaketLayananRepository {
	return &paketLayananRepositoryBridge{repo: NewLayananRepository(db)}
}

func (b *paketLayananRepositoryBridge) Create(ctx context.Context, paket *domain.PaketLayanan) error {
	return b.repo.CreatePaketLayanan(ctx, paket)
}

func (b *paketLayananRepositoryBridge) GetAll(ctx context.Context) ([]domain.PaketLayanan, error) {
	return b.repo.GetAllPaketLayanan(ctx)
}

func (b *paketLayananRepositoryBridge) GetByID(ctx context.Context, id uint64) (*domain.PaketLayanan, error) {
	return b.repo.GetPaketLayananByID(ctx, id)
}

func (b *paketLayananRepositoryBridge) Update(ctx context.Context, paket *domain.PaketLayanan) error {
	return b.repo.UpdatePaketLayanan(ctx, paket)
}

func (b *paketLayananRepositoryBridge) Delete(ctx context.Context, id uint64) error {
	return b.repo.DeletePaketLayanan(ctx, id)
}

// Wrapper struct for Diskon repository
type diskonRepositoryBridge struct {
	repo *layananRepository
}

func NewDiskonRepository(db *gorm.DB) domain.DiskonRepository {
	return &diskonRepositoryBridge{repo: NewLayananRepository(db)}
}

func (b *diskonRepositoryBridge) Create(ctx context.Context, diskon *domain.Diskon) error {
	return b.repo.CreateDiskon(ctx, diskon)
}

func (b *diskonRepositoryBridge) GetAll(ctx context.Context, limit, offset int, cluster string, isActive *bool) ([]domain.Diskon, int64, error) {
	return b.repo.GetAllDiskon(ctx, limit, offset, cluster, isActive)
}

func (b *diskonRepositoryBridge) GetByID(ctx context.Context, id uint64) (*domain.Diskon, error) {
	return b.repo.GetDiskonByID(ctx, id)
}

func (b *diskonRepositoryBridge) GetActiveForCluster(ctx context.Context, cluster string, checkDate time.Time) (*domain.Diskon, error) {
	return b.repo.GetActiveForCluster(ctx, cluster, checkDate)
}

func (b *diskonRepositoryBridge) Update(ctx context.Context, diskon *domain.Diskon) error {
	return b.repo.UpdateDiskon(ctx, diskon)
}

func (b *diskonRepositoryBridge) Delete(ctx context.Context, id uint64) error {
	return b.repo.DeleteDiskon(ctx, id)
}

func (b *diskonRepositoryBridge) GetClusterList(ctx context.Context) ([]string, error) {
	return b.repo.GetClusterList(ctx)
}
