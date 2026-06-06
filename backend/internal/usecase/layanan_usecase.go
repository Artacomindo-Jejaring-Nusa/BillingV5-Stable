package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"billing-backend/internal/domain"
)

// --- HargaLayananUsecase Implementation ---

type hargaLayananUsecase struct {
	repo domain.HargaLayananRepository
}

func NewHargaLayananUsecase(repo domain.HargaLayananRepository) domain.HargaLayananUsecase {
	return &hargaLayananUsecase{repo: repo}
}

func (u *hargaLayananUsecase) Create(ctx context.Context, brand *domain.HargaLayanan) (*domain.HargaLayanan, error) {
	existing, _ := u.repo.GetByID(ctx, brand.IDBrand)
	if existing != nil {
		return nil, fmt.Errorf("brand dengan ID '%s' sudah terdaftar", brand.IDBrand)
	}

	if err := u.repo.Create(ctx, brand); err != nil {
		return nil, err
	}
	return u.repo.GetByID(ctx, brand.IDBrand)
}

func (u *hargaLayananUsecase) GetAll(ctx context.Context) ([]domain.HargaLayanan, error) {
	return u.repo.GetAll(ctx)
}

func (u *hargaLayananUsecase) GetByID(ctx context.Context, idBrand string) (*domain.HargaLayanan, error) {
	brand, err := u.repo.GetByID(ctx, idBrand)
	if err != nil {
		return nil, err
	}
	if brand == nil {
		return nil, errors.New("brand tidak ditemukan")
	}
	return brand, nil
}

func (u *hargaLayananUsecase) Update(ctx context.Context, idBrand string, updates map[string]interface{}) (*domain.HargaLayanan, error) {
	brand, err := u.repo.GetByID(ctx, idBrand)
	if err != nil || brand == nil {
		return nil, errors.New("brand tidak ditemukan")
	}

	if val, ok := updates["brand"]; ok {
		if s, ok := val.(string); ok {
			brand.Brand = s
		}
	}
	if val, ok := updates["pajak"]; ok {
		if f, ok := val.(float64); ok {
			brand.Pajak = f
		}
	}
	if val, ok := updates["xendit_key_name"]; ok {
		if s, ok := val.(string); ok {
			brand.XenditKeyName = s
		}
	}

	if err := u.repo.Update(ctx, brand); err != nil {
		return nil, err
	}
	return u.repo.GetByID(ctx, idBrand)
}

func (u *hargaLayananUsecase) Delete(ctx context.Context, idBrand string) error {
	brand, err := u.repo.GetByID(ctx, idBrand)
	if err != nil || brand == nil {
		return errors.New("brand tidak ditemukan")
	}
	return u.repo.Delete(ctx, idBrand)
}

// --- PaketLayananUsecase Implementation ---

type paketLayananUsecase struct {
	repo      domain.PaketLayananRepository
	brandRepo domain.HargaLayananRepository
}

func NewPaketLayananUsecase(repo domain.PaketLayananRepository, brandRepo domain.HargaLayananRepository) domain.PaketLayananUsecase {
	return &paketLayananUsecase{
		repo:      repo,
		brandRepo: brandRepo,
	}
}

func (u *paketLayananUsecase) Create(ctx context.Context, paket *domain.PaketLayanan) (*domain.PaketLayanan, error) {
	// Validate if brand ID exists (matches Python logic)
	brand, err := u.brandRepo.GetByID(ctx, paket.IDBrand)
	if err != nil {
		return nil, err
	}
	if brand == nil {
		return nil, fmt.Errorf("brand '%s' tidak ditemukan. Silakan tambahkan brand terlebih dahulu", paket.IDBrand)
	}

	if err := u.repo.Create(ctx, paket); err != nil {
		return nil, err
	}
	return u.repo.GetByID(ctx, paket.ID)
}

func (u *paketLayananUsecase) GetAll(ctx context.Context) ([]domain.PaketLayanan, error) {
	return u.repo.GetAll(ctx)
}

func (u *paketLayananUsecase) GetByID(ctx context.Context, id uint64) (*domain.PaketLayanan, error) {
	paket, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if paket == nil {
		return nil, errors.New("paket layanan tidak ditemukan")
	}
	return paket, nil
}

func (u *paketLayananUsecase) Update(ctx context.Context, id uint64, updates map[string]interface{}) (*domain.PaketLayanan, error) {
	paket, err := u.repo.GetByID(ctx, id)
	if err != nil || paket == nil {
		return nil, errors.New("paket layanan tidak ditemukan")
	}

	if val, ok := updates["id_brand"]; ok {
		if s, ok := val.(string); ok {
			brand, err := u.brandRepo.GetByID(ctx, s)
			if err != nil || brand == nil {
				return nil, fmt.Errorf("brand '%s' tidak ditemukan", s)
			}
			paket.IDBrand = s
		}
	}
	if val, ok := updates["nama_paket"]; ok {
		if s, ok := val.(string); ok {
			paket.NamaPaket = s
		}
	}
	if val, ok := updates["kecepatan"]; ok {
		switch v := val.(type) {
		case float64:
			paket.Kecepatan = int(v)
		case int:
			paket.Kecepatan = v
		}
	}
	if val, ok := updates["harga"]; ok {
		if f, ok := val.(float64); ok {
			paket.Harga = f
		}
	}

	if err := u.repo.Update(ctx, paket); err != nil {
		return nil, err
	}
	return u.repo.GetByID(ctx, id)
}

func (u *paketLayananUsecase) Delete(ctx context.Context, id uint64) error {
	paket, err := u.repo.GetByID(ctx, id)
	if err != nil || paket == nil {
		return errors.New("paket layanan tidak ditemukan")
	}
	return u.repo.Delete(ctx, id)
}

// --- DiskonUsecase Implementation ---

type diskonUsecase struct {
	repo domain.DiskonRepository
}

func NewDiskonUsecase(repo domain.DiskonRepository) domain.DiskonUsecase {
	return &diskonUsecase{repo: repo}
}

func (u *diskonUsecase) Create(ctx context.Context, diskon *domain.Diskon) (*domain.Diskon, error) {
	// Validate dates (matches Python logic)
	if diskon.TglMulai != nil && diskon.TglSelesai != nil {
		if diskon.TglMulai.After(*diskon.TglSelesai) {
			return nil, errors.New("tanggal mulai tidak boleh lebih besar dari tanggal selesai")
		}
	}

	if err := u.repo.Create(ctx, diskon); err != nil {
		return nil, err
	}
	return u.repo.GetByID(ctx, diskon.ID)
}

func (u *diskonUsecase) GetAll(ctx context.Context, page, pageSize int, cluster string, isActive *bool) ([]domain.Diskon, int64, error) {
	offset := (page - 1) * pageSize
	return u.repo.GetAll(ctx, pageSize, offset, cluster, isActive)
}

func (u *diskonUsecase) GetByID(ctx context.Context, id uint64) (*domain.Diskon, error) {
	diskon, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if diskon == nil {
		return nil, errors.New("diskon tidak ditemukan")
	}
	return diskon, nil
}

func (u *diskonUsecase) GetActiveForCluster(ctx context.Context, cluster string, checkDate *time.Time) (*domain.Diskon, error) {
	dateToCheck := time.Now()
	if checkDate != nil {
		dateToCheck = *checkDate
	}
	return u.repo.GetActiveForCluster(ctx, cluster, dateToCheck)
}

func (u *diskonUsecase) Update(ctx context.Context, id uint64, updates map[string]interface{}) (*domain.Diskon, error) {
	diskon, err := u.repo.GetByID(ctx, id)
	if err != nil || diskon == nil {
		return nil, errors.New("diskon tidak ditemukan")
	}

	if val, ok := updates["nama_diskon"]; ok {
		if s, ok := val.(string); ok {
			diskon.NamaDiskon = s
		}
	}
	if val, ok := updates["persentase_diskon"]; ok {
		if f, ok := val.(float64); ok {
			diskon.PersentaseDiskon = f
		}
	}
	if val, ok := updates["cluster"]; ok {
		if s, ok := val.(string); ok {
			diskon.Cluster = s
		}
	}
	if val, ok := updates["is_active"]; ok {
		if b, ok := val.(bool); ok {
			diskon.IsActive = b
		}
	}
	if val, ok := updates["tgl_mulai"]; ok {
		if val == nil {
			diskon.TglMulai = nil
		} else if s, ok := val.(string); ok {
			if parsed, err := time.Parse("2006-01-02", s); err == nil {
				diskon.TglMulai = &parsed
			}
		} else if t, ok := val.(time.Time); ok {
			diskon.TglMulai = &t
		}
	}
	if val, ok := updates["tgl_selesai"]; ok {
		if val == nil {
			diskon.TglSelesai = nil
		} else if s, ok := val.(string); ok {
			if parsed, err := time.Parse("2006-01-02", s); err == nil {
				diskon.TglSelesai = &parsed
			}
		} else if t, ok := val.(time.Time); ok {
			diskon.TglSelesai = &t
		}
	}

	// Validate dates post update
	if diskon.TglMulai != nil && diskon.TglSelesai != nil {
		if diskon.TglMulai.After(*diskon.TglSelesai) {
			return nil, errors.New("tanggal mulai tidak boleh lebih besar dari tanggal selesai")
		}
	}

	if err := u.repo.Update(ctx, diskon); err != nil {
		return nil, err
	}
	return u.repo.GetByID(ctx, id)
}

func (u *diskonUsecase) Delete(ctx context.Context, id uint64) error {
	diskon, err := u.repo.GetByID(ctx, id)
	if err != nil || diskon == nil {
		return errors.New("diskon tidak ditemukan")
	}
	return u.repo.Delete(ctx, id)
}

func (u *diskonUsecase) Activate(ctx context.Context, id uint64) (*domain.Diskon, error) {
	diskon, err := u.repo.GetByID(ctx, id)
	if err != nil || diskon == nil {
		return nil, errors.New("diskon tidak ditemukan")
	}

	diskon.IsActive = true
	if err := u.repo.Update(ctx, diskon); err != nil {
		return nil, err
	}
	return diskon, nil
}

func (u *diskonUsecase) GetClusterList(ctx context.Context) ([]string, error) {
	return u.repo.GetClusterList(ctx)
}
