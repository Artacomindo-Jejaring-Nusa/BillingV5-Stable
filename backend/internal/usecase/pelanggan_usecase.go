package usecase

import (
	"context"
	"errors"

	"billing-backend/internal/domain"
	"billing-backend/internal/websocket"
)

type pelangganUsecase struct {
	pelangganRepo domain.PelangganRepository
}

func NewPelangganUsecase(p domain.PelangganRepository) domain.PelangganUsecase {
	return &pelangganUsecase{
		pelangganRepo: p,
	}
}

func (u *pelangganUsecase) FetchAll(ctx context.Context, page, pageSize int) ([]domain.Pelanggan, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	return u.pelangganRepo.GetAll(ctx, pageSize, offset)
}

func (u *pelangganUsecase) GetByID(ctx context.Context, id uint64) (*domain.Pelanggan, error) {
	return u.pelangganRepo.GetByID(ctx, id)
}

func isDummyKtp(ktp string) bool {
	if ktp == "" {
		return true
	}
	for _, char := range ktp {
		if char != '0' {
			return false
		}
	}
	return true
}

func (u *pelangganUsecase) Store(ctx context.Context, pelanggan *domain.Pelanggan) error {
	// Add business validation here
	if pelanggan.Email == "" {
		return errors.New("email is required")
	}

	// Validate email uniqueness
	existingEmail, err := u.pelangganRepo.GetByEmail(ctx, pelanggan.Email)
	if err == nil && existingEmail != nil {
		return errors.New("Email sudah terdaftar")
	}

	// Validate NIK/NoKtp uniqueness, except for dummy values (empty or all zeros)
	if !isDummyKtp(pelanggan.NoKtp) {
		existing, err := u.pelangganRepo.GetByNoKtp(ctx, pelanggan.NoKtp)
		if err == nil && existing != nil {
			return errors.New("NIK/No KTP sudah terdaftar")
		}
	}

	if err := u.pelangganRepo.Create(ctx, pelanggan); err != nil {
		return err
	}

	if websocket.GlobalHub != nil {
		websocket.GlobalHub.BroadcastNotification("new_customer", map[string]interface{}{
			"pelanggan_nama": pelanggan.Nama,
		})
	}

	return nil
}

func (u *pelangganUsecase) Update(ctx context.Context, id uint64, req *domain.Pelanggan) error {
	existing, err := u.pelangganRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Validate email uniqueness
	if req.Email != "" {
		dupEmail, err := u.pelangganRepo.GetByEmail(ctx, req.Email)
		if err == nil && dupEmail != nil && dupEmail.ID != id {
			return errors.New("Email sudah terdaftar oleh pelanggan lain")
		}
	}

	// Validate NIK/NoKtp uniqueness, except for dummy values (empty or all zeros)
	if !isDummyKtp(req.NoKtp) {
		dup, err := u.pelangganRepo.GetByNoKtp(ctx, req.NoKtp)
		if err == nil && dup != nil && dup.ID != id {
			return errors.New("NIK/No KTP sudah terdaftar oleh pelanggan lain")
		}
	}

	// Update fields safely
	existing.Nama = req.Nama
	existing.NoKtp = req.NoKtp
	existing.Alamat = req.Alamat
	existing.AlamatCustom = req.AlamatCustom
	existing.TglInstalasi = req.TglInstalasi
	existing.Blok = req.Blok
	existing.Unit = req.Unit
	existing.NoTelp = req.NoTelp
	existing.Email = req.Email
	existing.IDBrand = req.IDBrand
	existing.Layanan = req.Layanan
	existing.BrandDefault = req.BrandDefault
	existing.MikrotikServerID = req.MikrotikServerID

	return u.pelangganRepo.Update(ctx, existing)
}

func (u *pelangganUsecase) Delete(ctx context.Context, id uint64) error {
	_, err := u.pelangganRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return u.pelangganRepo.Delete(ctx, id)
}
