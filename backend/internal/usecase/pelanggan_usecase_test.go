package usecase

import (
	"context"
	"testing"

	"billing-backend/internal/domain"
)

func TestPelangganExport(t *testing.T) {
	repo := &mockPelangganRepo{
		data: map[uint64]*domain.Pelanggan{
			1: {ID: 1, Nama: "Jajang", Alamat: "Tambun", Email: "jajang@mail.com"},
		},
	}
	u := NewPelangganUsecase(repo)

	// Test CSV Export
	data, contentType, err := u.Export(context.Background(), "csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contentType != "text/csv" {
		t.Errorf("expected text/csv, got %s", contentType)
	}
	if len(data) == 0 {
		t.Error("exported data is empty")
	}

	// Test Excel Export
	data, contentType, err = u.Export(context.Background(), "excel")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contentType != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Errorf("expected excel content type, got %s", contentType)
	}
}

func TestPelangganImport(t *testing.T) {
	repo := &mockPelangganRepo{
		data: make(map[uint64]*domain.Pelanggan),
	}
	u := NewPelangganUsecase(repo)

	csvContent := "No KTP;Nama;Alamat;Blok;Unit;No Telp;Email;Layanan;ID Brand;Tgl Instalasi\n" +
		"12345;Ahmad;Tambun;A;1;0812;ahmad@mail.com;Internet;ajn-01;2024-01-01"

	count, err := u.ImportFromCSV(context.Background(), csvContent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 imported record, got %d", count)
	}

	if len(repo.data) != 1 {
		t.Errorf("expected 1 record in repo, got %d", len(repo.data))
	}
}

func TestPelangganCRUD(t *testing.T) {
	repo := &mockPelangganRepo{data: make(map[uint64]*domain.Pelanggan)}
	u := NewPelangganUsecase(repo)

	t.Run("Store - Email Required", func(t *testing.T) {
		err := u.Store(context.Background(), &domain.Pelanggan{Nama: "Test"})
		if err == nil || err.Error() != "email is required" {
			t.Errorf("expected email required error, got %v", err)
		}
	})

	t.Run("Store - Success", func(t *testing.T) {
		p := &domain.Pelanggan{Nama: "Budi", Email: "budi@mail.com"}
		err := u.Store(context.Background(), p)
		if err != nil { t.Fatalf("unexpected error: %v", err) }
		if p.ID == 0 { t.Error("expected non-zero ID") }
	})

	t.Run("Update - Not Found", func(t *testing.T) {
		err := u.Update(context.Background(), 99, &domain.Pelanggan{Nama: "X"})
		if err == nil { t.Error("expected error for non-existent ID") }
	})

	t.Run("Update - Success", func(t *testing.T) {
		err := u.Update(context.Background(), 1, &domain.Pelanggan{Nama: "Budi Updated", Email: "budi@mail.com"})
		if err != nil { t.Fatalf("unexpected error: %v", err) }
		
		p, _ := repo.GetByID(context.Background(), 1)
		if p.Nama != "Budi Updated" { t.Errorf("expected name update, got %s", p.Nama) }
	})

	t.Run("Delete - Success", func(t *testing.T) {
		err := u.Delete(context.Background(), 1)
		if err != nil { t.Fatalf("unexpected error: %v", err) }
		
		p, _ := repo.GetByID(context.Background(), 1)
		if p != nil { t.Error("expected nil after delete") }
	})
}

