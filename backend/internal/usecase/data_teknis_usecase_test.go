package usecase

import (
	"context"
	"testing"

	"billing-backend/internal/domain"
)

type mockDataTeknisRepoFull struct {
	domain.DataTeknisRepository
}

func (m *mockDataTeknisRepoFull) GetAll(ctx context.Context, skip, limit int, search string, olt string, profile string, vlan string, onuPowerMin, onuPowerMax *int) ([]domain.DataTeknis, int64, error) {
	oltName := "OLT-01"
	sn := "SN123"
	return []domain.DataTeknis{
		{ID: 1, IDPelanggan: "CUST001", Olt: &oltName, Sn: &sn},
	}, 1, nil
}

func (m *mockDataTeknisRepoFull) GetByPelangganID(ctx context.Context, pelangganID uint64) (*domain.DataTeknis, error) {
	if pelangganID == 1 {
		return &domain.DataTeknis{ID: 1, PelangganID: 1}, nil
	}
	return nil, nil
}

func (m *mockDataTeknisRepoFull) Create(ctx context.Context, d *domain.DataTeknis) error {
	return nil
}

func (m *mockDataTeknisRepoFull) GetByID(ctx context.Context, id uint64) (*domain.DataTeknis, error) {
	return &domain.DataTeknis{ID: id, PelangganID: 1}, nil
}

func (m *mockDataTeknisRepoFull) Update(ctx context.Context, d *domain.DataTeknis) error {
	return nil
}

func (m *mockDataTeknisRepoFull) CheckIPAddress(ctx context.Context, ip string, excludeID *uint64) (bool, error) {
	return false, nil
}

func TestDataTeknisStore(t *testing.T) {
	repo := &mockDataTeknisRepoFull{}
	pRepo := &mockPelangganRepo{data: map[uint64]*domain.Pelanggan{
		2: {ID: 2, Nama: "New User"},
	}}
	u := NewDataTeknisUsecase(repo, nil, pRepo, nil)

	t.Run("Store successful", func(t *testing.T) {
		dt := &domain.DataTeknis{PelangganID: 2, IDPelanggan: "CUST002"}
		err := u.Store(context.Background(), dt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("Store fail when already exists", func(t *testing.T) {
		dt := &domain.DataTeknis{PelangganID: 1}
		err := u.Store(context.Background(), dt)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestDataTeknisUpdate(t *testing.T) {
	repo := &mockDataTeknisRepoFull{}
	u := NewDataTeknisUsecase(repo, nil, nil, nil)

	dt := &domain.DataTeknis{IDPelanggan: "UPDATED"}
	err := u.Update(context.Background(), 1, dt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDataTeknisExport(t *testing.T) {
	repo := &mockDataTeknisRepoFull{}
	u := NewDataTeknisUsecase(repo, nil, nil, nil)

	// Test CSV
	_, contentType, err := u.Export(context.Background(), "csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contentType != "text/csv" {
		t.Errorf("expected text/csv, got %s", contentType)
	}

	// Test Excel
	_, contentType, err = u.Export(context.Background(), "excel")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contentType != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Errorf("expected excel content type, got %s", contentType)
	}
}
