package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"billing-backend/config"
	"billing-backend/internal/domain"
)

func TestCalculatePrice(t *testing.T) {
	brandID := "ajn-01"
	brand := &domain.HargaLayanan{
		IDBrand: brandID,
		Brand:   "Jakinet",
		Pajak:   11.0,
	}
	paket := &domain.PaketLayanan{
		ID:        1,
		NamaPaket: "10Mbps",
		Harga:     300000.0,
	}
	pelanggan := &domain.Pelanggan{
		ID:      1,
		Nama:    "John Doe",
		IDBrand: &brandID,
	}

	pRepo := &mockPelangganRepo{data: map[uint64]*domain.Pelanggan{1: pelanggan}}
	pkRepo := &mockPaketRepo{data: map[uint64]*domain.PaketLayanan{1: paket}}
	bRepo := &mockBrandRepo{data: map[string]*domain.HargaLayanan{brandID: brand}}
	u := NewBillingUsecase(nil, nil, pRepo, pkRepo, bRepo, nil, nil, nil, nil, nil)

	t.Run("Otomatis payment method", func(t *testing.T) {
		tglMulai := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
		req := &domain.LanggananCalculateRequest{
			PelangganID:      1,
			PaketLayananID:   1,
			MetodePembayaran: "Otomatis",
			TglMulai:         &tglMulai,
		}

		res, err := u.CalculatePrice(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// 300,000 * 1.11 = 333,000
		if res.HargaAwal != 333000 {
			t.Errorf("expected HargaAwal to be 333000, got %f", res.HargaAwal)
		}

		// Tanggal 1 bulan depan sesuai logika legacy Python
		expectedTempo := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
		if !res.TglJatuhTempo.Equal(expectedTempo) {
			t.Errorf("expected TglJatuhTempo to be %v, got %v", expectedTempo, res.TglJatuhTempo)
		}
	})

	t.Run("Prorate payment method (15th of June, 30 days)", func(t *testing.T) {
		tglMulai := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
		req := &domain.LanggananCalculateRequest{
			PelangganID:      1,
			PaketLayananID:   1,
			MetodePembayaran: "Prorate",
			TglMulai:         &tglMulai,
		}

		res, err := u.CalculatePrice(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// (300,000 / 30 * 16) * 1.11 = 160,000 * 1.11 = 177,600
		if res.HargaAwal != 177600 {
			t.Errorf("expected HargaAwal to be 177600, got %f", res.HargaAwal)
		}

		// End of month: 2026-06-30
		expectedTempo := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)
		if !res.TglJatuhTempo.Equal(expectedTempo) {
			t.Errorf("expected TglJatuhTempo to be %v, got %v", expectedTempo, res.TglJatuhTempo)
		}
	})
}

func TestCalculateProratePlusFull(t *testing.T) {
	brandID := "ajn-01"
	brand := &domain.HargaLayanan{IDBrand: brandID, Brand: "Jakinet", Pajak: 11.0}
	paket := &domain.PaketLayanan{ID: 1, NamaPaket: "10Mbps", Harga: 300000.0}
	pelanggan := &domain.Pelanggan{ID: 1, Nama: "John Doe", IDBrand: &brandID}

	pRepo := &mockPelangganRepo{data: map[uint64]*domain.Pelanggan{1: pelanggan}}
	pkRepo := &mockPaketRepo{data: map[uint64]*domain.PaketLayanan{1: paket}}
	bRepo := &mockBrandRepo{data: map[string]*domain.HargaLayanan{brandID: brand}}
	u := NewBillingUsecase(nil, nil, pRepo, pkRepo, bRepo, nil, nil, nil, nil, nil)

	tglMulai := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	req := &domain.LanggananCalculateRequest{
		PelangganID:    1,
		PaketLayananID: 1,
		TglMulai:       &tglMulai,
	}

	res, err := u.CalculateProratePlusFull(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Prorate: 177,600
	// Full: 333,000
	// Total: 510,600
	if res.HargaTotalAwal != 510600 {
		t.Errorf("expected HargaTotalAwal 510600, got %f", res.HargaTotalAwal)
	}
}

func TestCalculateProrate(t *testing.T) {
	brandID := "ajn-01"
	brand := &domain.HargaLayanan{IDBrand: brandID, Brand: "Jakinet", Pajak: 11.0}
	paket := &domain.PaketLayanan{ID: 1, NamaPaket: "10Mbps", Harga: 300000.0}

	pkRepo := &mockPaketRepo{data: map[uint64]*domain.PaketLayanan{1: paket}}
	bRepo := &mockBrandRepo{data: map[string]*domain.HargaLayanan{brandID: brand}}
	u := NewBillingUsecase(nil, nil, nil, pkRepo, bRepo, nil, nil, nil, nil, nil)

	t.Run("without next month PPN", func(t *testing.T) {
		tglMulai := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
		req := &domain.ProrateCalculationRequest{
			PaketLayananID:      1,
			IDBrand:             brandID,
			TglMulai:            &tglMulai,
			IncludePpnNextMonth: false,
		}

		res, err := u.CalculateProrate(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if res.HargaDasarProrate != 160000 {
			t.Errorf("expected dasar prorate 160000, got %f", res.HargaDasarProrate)
		}
		if res.Pajak != 17600 {
			t.Errorf("expected pajak 17600, got %f", res.Pajak)
		}
		if res.TotalHargaProrate != 177600 {
			t.Errorf("expected total prorate 177600, got %f", res.TotalHargaProrate)
		}
		if res.PeriodeHari != 16 {
			t.Errorf("expected 16 days, got %d", res.PeriodeHari)
		}
	})

	t.Run("with next month PPN", func(t *testing.T) {
		tglMulai := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
		req := &domain.ProrateCalculationRequest{
			PaketLayananID:      1,
			IDBrand:             brandID,
			TglMulai:            &tglMulai,
			IncludePpnNextMonth: true,
		}

		res, err := u.CalculateProrate(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if res.TotalHargaProrate != 177600 {
			t.Errorf("expected total prorate 177600, got %f", res.TotalHargaProrate)
		}
		if *res.HargaBulanDepan != 300000 {
			t.Errorf("expected HargaBulanDepan 300000")
		}
		if *res.PpnBulanDepan != 33000 {
			t.Errorf("expected PpnBulanDepan 33000")
		}
		if *res.TotalBulanDepanDenganPpn != 333000 {
			t.Errorf("expected TotalBulanDepanDenganPpn 333000")
		}
		if *res.TotalKeseluruhan != 510600 {
			t.Errorf("expected TotalKeseluruhan 510600, got %f", *res.TotalKeseluruhan)
		}
	})
}

func TestCalculateDiskon(t *testing.T) {
	brandID := "ajn-01"
	brand := &domain.HargaLayanan{IDBrand: brandID, Brand: "Jakinet", Pajak: 11.0}
	paket := &domain.PaketLayanan{ID: 1, NamaPaket: "10Mbps", Harga: 300000.0}

	pkRepo := &mockPaketRepo{data: map[uint64]*domain.PaketLayanan{1: paket}}
	bRepo := &mockBrandRepo{data: map[string]*domain.HargaLayanan{brandID: brand}}
	u := NewBillingUsecase(nil, nil, nil, pkRepo, bRepo, nil, nil, nil, nil, nil)

	req := &domain.DiskonCalculationRequest{
		PaketLayananID:   1,
		IDBrand:          brandID,
		PersentaseDiskon: 10.0,
	}

	res, err := u.CalculateDiskon(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 300,000 + 33,000 = 333,000
	// 333,000 - 10% (33,300) = 299,700
	if res.HargaFinal != 299700 {
		t.Errorf("expected HargaFinal 299700, got %f", res.HargaFinal)
	}
}

func TestCreateLanggananValidation(t *testing.T) {
	brandID := "ajn-01"
	brand := &domain.HargaLayanan{IDBrand: brandID, Brand: "Jakinet", Pajak: 11.0}
	paket := &domain.PaketLayanan{ID: 1, NamaPaket: "10Mbps", Harga: 300000.0}

	t.Run("CreateLangganan throws error when DataTeknis is missing", func(t *testing.T) {
		pelanggan := &domain.Pelanggan{
			ID:      1,
			Nama:    "Ahmad Rizki",
			IDBrand: &brandID,
		}

		pRepo := &mockPelangganRepo{data: map[uint64]*domain.Pelanggan{1: pelanggan}}
		pkRepo := &mockPaketRepo{data: map[uint64]*domain.PaketLayanan{1: paket}}
		bRepo := &mockBrandRepo{data: map[string]*domain.HargaLayanan{brandID: brand}}
		u := NewBillingUsecase(nil, nil, pRepo, pkRepo, bRepo, nil, nil, nil, nil, nil)

		langganan := &domain.Langganan{
			PelangganID:    1,
			PaketLayananID: 1,
		}

		err := u.CreateLangganan(context.Background(), langganan)
		expectedErr := "Langganan tidak dapat dibuat. Pelanggan 'Ahmad Rizki' belum memiliki data teknis. Tim NOC harus menambahkan data teknis terlebih dahulu sebelum membuat langganan."
		if err == nil || err.Error() != expectedErr {
			t.Errorf("expected error '%s', got '%v'", expectedErr, err)
		}
	})

	t.Run("CreateLangganan succeeds when DataTeknis is present", func(t *testing.T) {
		pelanggan := &domain.Pelanggan{
			ID:         1,
			Nama:       "Ahmad Rizki",
			IDBrand:    &brandID,
			DataTeknis: &domain.DataTeknis{ID: 1},
		}

		pRepo := &mockPelangganRepo{data: map[uint64]*domain.Pelanggan{1: pelanggan}}
		pkRepo := &mockPaketRepo{data: map[uint64]*domain.PaketLayanan{1: paket}}
		bRepo := &mockBrandRepo{data: map[string]*domain.HargaLayanan{brandID: brand}}
		lRepo := &mockLanggananRepo{data: make(map[uint64]*domain.Langganan)}
		u := NewBillingUsecase(nil, lRepo, pRepo, pkRepo, bRepo, nil, nil, nil, nil, nil)

		langganan := &domain.Langganan{
			PelangganID:    1,
			PaketLayananID: 1,
		}

		err := u.CreateLangganan(context.Background(), langganan)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestProcessXenditCallback(t *testing.T) {
	// Setup
	cfg := &config.Config{
		XenditCallbackTokenArtacomindo: "artacom_token_123",
		XenditCallbackTokenJelantik:    "jelantik_token_123",
	}

	invoice := &domain.Invoice{
		ID:            1,
		InvoiceNumber: "JELANTIK/INV/001",
		StatusInvoice: "Belum Dibayar",
		TotalHarga:    300000.0,
		Pelanggan: &domain.Pelanggan{
			ID:   1,
			Nama: "John Doe",
			Langganan: []domain.Langganan{
				{
					ID:               1,
					Status:           "Suspended",
					MetodePembayaran: "Otomatis",
				},
			},
		},
	}

	invRepo := &mockInvoiceRepoCallback{
		invoices: map[string]*domain.Invoice{
			"JELANTIK/INV/001": invoice,
		},
	}

	langgRepo := &mockLanggananRepo{
		data: map[uint64]*domain.Langganan{
			1: &invoice.Pelanggan.Langganan[0],
		},
	}

	u := NewBillingUsecase(invRepo, langgRepo, nil, nil, nil, nil, nil, nil, nil, cfg)

	payload := map[string]interface{}{
		"id":          "xendit_123",
		"external_id": "JELANTIK/INV/001",
		"status":      "PAID",
		"paid_amount": 300000.0,
		"paid_at":     "2026-06-03T15:00:00Z",
	}

	err := u.ProcessXenditCallback(context.Background(), "jelantik_token_123", payload, "idempotency_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if invoice.StatusInvoice != "Lunas" {
		t.Errorf("expected invoice status to be Lunas, got %s", invoice.StatusInvoice)
	}

	if invoice.Pelanggan.Langganan[0].Status != "Aktif" {
		t.Errorf("expected langganan status to be Aktif, got %s", invoice.Pelanggan.Langganan[0].Status)
	}
}

func TestExportLangganan(t *testing.T) {
	langRepo := &mockLanggananRepo{
		data: map[uint64]*domain.Langganan{
			1: {ID: 1, Status: "Aktif", Pelanggan: &domain.Pelanggan{Nama: "Jajang"}},
		},
	}
	u := NewBillingUsecase(nil, langRepo, nil, nil, nil, nil, nil, nil, nil, nil)

	// Test CSV
	data, contentType, err := u.ExportLangganan(context.Background(), "csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contentType != "text/csv" {
		t.Errorf("expected text/csv, got %s", contentType)
	}
	if len(data) == 0 { t.Error("empty data") }

	// Test Excel
	data, contentType, err = u.ExportLangganan(context.Background(), "excel")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contentType != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Errorf("expected excel content type, got %s", contentType)
	}
	if len(data) == 0 { t.Error("empty data") }
}

func TestExportLanggananMultiSheet(t *testing.T) {
	langRepo := &mockLanggananRepo{
		data: map[uint64]*domain.Langganan{
			1: {ID: 1, Status: "Aktif", Pelanggan: &domain.Pelanggan{Nama: "Jajang"}},
		},
	}
	invRepo := &mockInvoiceRepoCallback{
		invoices: make(map[string]*domain.Invoice),
	}
	dtRepo := &mockDataTeknisRepo{}

	u := NewBillingUsecase(invRepo, langRepo, nil, nil, nil, dtRepo, nil, nil, nil, nil)

	data, contentType, err := u.ExportLanggananMultiSheet(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contentType != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Errorf("expected excel content type, got %s", contentType)
	}
	if len(data) == 0 { t.Error("empty data") }
}

func TestExportInvoices(t *testing.T) {
	invRepo := &mockInvoiceRepoCallback{
		invoices: map[string]*domain.Invoice{
			"INV001": {ID: 1, InvoiceNumber: "INV001", TotalHarga: 1000, StatusInvoice: "Lunas", TglInvoice: time.Now()},
		},
	}
	u := NewBillingUsecase(invRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	// Test CSV
	_, _, err := u.ExportInvoices(context.Background(), "csv")
	if err != nil { t.Fatalf("unexpected error: %v", err) }

	// Test Excel
	_, _, err = u.ExportInvoices(context.Background(), "excel")
	if err != nil { t.Fatalf("unexpected error: %v", err) }
}

func TestExportPaymentLinksExcel(t *testing.T) {
	invRepo := &mockInvoiceRepoCallback{
		invoices: map[string]*domain.Invoice{
			"INV001": {ID: 1, InvoiceNumber: "INV001", TotalHarga: 1000},
		},
	}
	u := NewBillingUsecase(invRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	_, err := u.ExportPaymentLinksExcel(context.Background(), nil)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
}

func (m *mockInvoiceRepoCallback) GetRevenueReport(ctx context.Context, params *domain.RevenueReportParams) (*domain.RevenueReportResponse, error) {
	return &domain.RevenueReportResponse{TotalPendapatan: 1000}, nil
}

func (m *mockInvoiceRepoCallback) GetRevenueReportDetails(ctx context.Context, params *domain.RevenueReportParams) ([]domain.InvoiceReportItem, error) {
	return []domain.InvoiceReportItem{{ID: 1, InvoiceNumber: "INV001"}}, nil
}

func (m *mockInvoiceRepoCallback) ExportPaymentLinksExcel(ctx context.Context, filters map[string]string) ([]byte, error) {
	return []byte("dummy excel data"), nil
}

func TestAutoSuspend(t *testing.T) {
	invRepo := &mockInvoiceRepoCallback{
		invoices: map[string]*domain.Invoice{
			"INV001": {ID: 1, StatusInvoice: "Belum Dibayar", TglJatuhTempo: time.Now().AddDate(0, 0, -1)},
		},
	}
	langRepo := &mockLanggananRepo{data: make(map[uint64]*domain.Langganan)}
	sysRepo := &mockSystemRepo{}
	u := NewBillingUsecase(invRepo, langRepo, nil, nil, nil, nil, nil, nil, sysRepo, nil).(*billingUsecase)

	err := u.AutoSuspend(context.Background())
	if err != nil { t.Fatalf("unexpected error: %v", err) }
}

func TestAutoVerifyPayments(t *testing.T) {
	xenID := "xen_123"
	invRepo := &mockInvoiceRepoCallback{
		invoices: map[string]*domain.Invoice{
			"INV001": {ID: 1, StatusInvoice: "Belum Dibayar", XenditID: &xenID},
		},
	}
	sysRepo := &mockSystemRepo{}
	u := NewBillingUsecase(invRepo, nil, nil, nil, nil, nil, nil, nil, sysRepo, nil).(*billingUsecase)

	err := u.VerifyPayments(context.Background())
	if err != nil { t.Fatalf("unexpected error: %v", err) }
}

func TestRetryFailedMikrotikSync(t *testing.T) {
	// 1. Test empty pending list
	t.Run("empty pending list", func(t *testing.T) {
		dtRepo := &mockDataTeknisRepoWithPending{}
		sysRepo := &mockSystemRepo{}
		u := NewBillingUsecase(nil, nil, nil, nil, nil, dtRepo, nil, nil, sysRepo, nil).(*billingUsecase)

		err := u.RetryFailedMikrotikSync(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(dtRepo.updatedList) != 0 {
			t.Errorf("expected no updates, got %d", len(dtRepo.updatedList))
		}
	})

	// 2. Test pending list with failed sync (due to router offline/error)
	t.Run("failed sync updates pending state", func(t *testing.T) {
		serverID := uint64(1)
		dt := domain.DataTeknis{
			ID:                  1,
			IDPelanggan:         "john_pppoe",
			PasswordPppoe:       "secret",
			MikrotikServerID:    &serverID,
			MikrotikSyncPending: true,
			Pelanggan: &domain.Pelanggan{
				ID:   10,
				Nama: "John Doe",
				Langganan: []domain.Langganan{
					{
						ID:     20,
						Status: "Aktif",
					},
				},
			},
		}

		dtRepo := &mockDataTeknisRepoWithPending{
			pendingList: []domain.DataTeknis{dt},
		}
		mikRepo := &mockMikrotikRepo{
			GetByIDFunc: func(ctx context.Context, id uint64) (*domain.MikrotikServer, error) {
				return nil, errors.New("router offline")
			},
		}
		sysRepo := &mockSystemRepo{}
		u := NewBillingUsecase(nil, nil, nil, nil, nil, dtRepo, mikRepo, nil, sysRepo, nil).(*billingUsecase)

		err := u.RetryFailedMikrotikSync(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Since dt.MikrotikSyncPending was already true, and it failed again,
		// it should remain true and not call Update again to set it to true.
		if len(dtRepo.updatedList) != 0 {
			t.Errorf("expected no database update since status is already pending, got %d", len(dtRepo.updatedList))
		}
	})

	// 3. Test pending list with transition from pending=false to pending=true
	t.Run("transition to pending true on failure", func(t *testing.T) {
		serverID := uint64(1)
		dt := domain.DataTeknis{
			ID:                  1,
			IDPelanggan:         "john_pppoe",
			PasswordPppoe:       "secret",
			MikrotikServerID:    &serverID,
			MikrotikSyncPending: false, // Start as false
		}

		dtRepo := &mockDataTeknisRepoWithPending{}
		mikRepo := &mockMikrotikRepo{
			GetByIDFunc: func(ctx context.Context, id uint64) (*domain.MikrotikServer, error) {
				return nil, errors.New("router offline")
			},
		}
		sysRepo := &mockSystemRepo{}
		u := NewBillingUsecase(nil, nil, nil, nil, nil, dtRepo, mikRepo, nil, sysRepo, nil).(*billingUsecase)

		err := u.triggerMikrotikUpdate(context.Background(), "john_pppoe", &dt, "Aktif")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !dt.MikrotikSyncPending {
			t.Error("expected MikrotikSyncPending to be set to true")
		}
		if len(dtRepo.updatedList) != 1 {
			t.Fatalf("expected 1 database update, got %d", len(dtRepo.updatedList))
		}
		if !dtRepo.updatedList[0].MikrotikSyncPending {
			t.Error("expected updated DataTeknis to have MikrotikSyncPending = true")
		}
	})
}

type mockMikrotikRepo struct {
	domain.MikrotikRepository
	GetByIDFunc func(ctx context.Context, id uint64) (*domain.MikrotikServer, error)
}

func (m *mockMikrotikRepo) GetByID(ctx context.Context, id uint64) (*domain.MikrotikServer, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

type mockDataTeknisRepoWithPending struct {
	domain.DataTeknisRepository
	pendingList []domain.DataTeknis
	updatedList []*domain.DataTeknis
	GetPendingSyncFunc func(ctx context.Context) ([]domain.DataTeknis, error)
	UpdateFunc func(ctx context.Context, data *domain.DataTeknis) error
}

func (m *mockDataTeknisRepoWithPending) GetPendingSync(ctx context.Context) ([]domain.DataTeknis, error) {
	if m.GetPendingSyncFunc != nil {
		return m.GetPendingSyncFunc(ctx)
	}
	return m.pendingList, nil
}

func (m *mockDataTeknisRepoWithPending) Update(ctx context.Context, data *domain.DataTeknis) error {
	m.updatedList = append(m.updatedList, data)
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, data)
	}
	return nil
}

type mockInvoiceRepoForGenerate struct {
	domain.InvoiceRepository
	created []*domain.Invoice
}

func (m *mockInvoiceRepoForGenerate) GetInvoiceByPelangganAndDueDateRange(ctx context.Context, pelangganID uint64, start, end time.Time) (*domain.Invoice, error) {
	return nil, nil
}

func (m *mockInvoiceRepoForGenerate) GetByInvoiceNumber(ctx context.Context, num string) (*domain.Invoice, error) {
	return nil, errors.New("invoice not found")
}

func (m *mockInvoiceRepoForGenerate) Create(ctx context.Context, inv *domain.Invoice) error {
	m.created = append(m.created, inv)
	return nil
}

func (m *mockInvoiceRepoForGenerate) Update(ctx context.Context, inv *domain.Invoice) error {
	return nil
}

type mockLanggananRepoForGenerate struct {
	domain.LanggananRepository
	data []domain.Langganan
}

func (m *mockLanggananRepoForGenerate) GetAll(ctx context.Context, limit, offset int, search, status string, forInvoiceSelection bool, sortBy, sortOrder string) ([]domain.Langganan, int64, error) {
	return m.data, int64(len(m.data)), nil
}

func (m *mockLanggananRepoForGenerate) Update(ctx context.Context, l *domain.Langganan) error {
	return nil
}

type mockDataTeknisRepoForGenerate struct {
	domain.DataTeknisRepository
	data *domain.DataTeknis
}

func (m *mockDataTeknisRepoForGenerate) GetByPelangganID(ctx context.Context, pelangganID uint64) (*domain.DataTeknis, error) {
	return m.data, nil
}

type mockDiskonRepoForGenerate struct {
	domain.DiskonRepository
}

func (m *mockDiskonRepoForGenerate) GetActiveForCluster(ctx context.Context, cluster string, t time.Time) (*domain.Diskon, error) {
	return nil, nil
}

func TestGenerateInvoices(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}

	today := time.Now().In(loc)
	dueDate := today.AddDate(0, 0, 35) // subscription period end (e.g. 35 days from today)
	dueDatePembayaran := today.AddDate(0, 0, 5) // payment due date (exactly 5 days from today)

	brandID := "ajn-01"
	brand := &domain.HargaLayanan{
		IDBrand: brandID,
		Brand:   "Jakinet",
		Pajak:   11.0,
	}

	paket := &domain.PaketLayanan{
		ID:        1,
		NamaPaket: "10Mbps",
		Harga:     300000.0,
		Kecepatan: 10,
	}

	pelanggan := &domain.Pelanggan{
		ID:      1,
		Nama:    "Ahmad Rizki",
		IDBrand: &brandID,
		Alamat:  "Jl. Kebon Jeruk",
		NoTelp:  "081234567890",
		Email:   "ahmad@gmail.com",
	}

	langganan := domain.Langganan{
		ID:                        1,
		PelangganID:               1,
		PaketLayananID:            1,
		Status:                    "Aktif",
		TglJatuhTempo:             &dueDate,
		TglJatuhTempoPembayaran:   &dueDatePembayaran,
		HargaAwal:                 &paket.Harga,
		MetodePembayaran:          "Otomatis",
	}

	langRepo := &mockLanggananRepoForGenerate{data: []domain.Langganan{langganan}}
	invRepo := &mockInvoiceRepoForGenerate{}
	pelRepo := &mockPelangganRepo{data: map[uint64]*domain.Pelanggan{1: pelanggan}}
	bRepo := &mockBrandRepo{data: map[string]*domain.HargaLayanan{brandID: brand}}
	dtRepo := &mockDataTeknisRepoForGenerate{data: &domain.DataTeknis{ID: 1, IDPelanggan: "JKT-001"}}
	pkRepo := &mockPaketRepo{data: map[uint64]*domain.PaketLayanan{1: paket}}
	diskRepo := &mockDiskonRepoForGenerate{}
	sysRepo := &mockSystemRepo{}

	cfg := &config.Config{
		XenditApiUrl: "https://api.xendit.co/v2/invoices",
	}

	u := NewBillingUsecase(invRepo, langRepo, pelRepo, pkRepo, bRepo, dtRepo, nil, diskRepo, sysRepo, cfg).(*billingUsecase)

	err = u.GenerateInvoices(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify that exactly 1 invoice was created
	if len(invRepo.created) != 1 {
		t.Fatalf("expected 1 invoice to be created, got %d", len(invRepo.created))
	}

	inv := invRepo.created[0]
	if inv.PelangganID != 1 {
		t.Errorf("expected PelangganID 1, got %d", inv.PelangganID)
	}

	expectedNumberPrefix := "JAKINET/ftth/AHMADRIZKI/"
	if !strings.HasPrefix(inv.InvoiceNumber, expectedNumberPrefix) {
		t.Errorf("expected invoice number to start with %q, got %q", expectedNumberPrefix, inv.InvoiceNumber)
	}

	if inv.Brand != "Jakinet" {
		t.Errorf("expected Brand 'Jakinet', got %q", inv.Brand)
	}

	if inv.IDPelanggan != "JKT-001" {
		t.Errorf("expected IDPelanggan 'JKT-001', got %q", inv.IDPelanggan)
	}

	if inv.TotalHarga != 300000.0 {
		t.Errorf("expected TotalHarga 300000.0, got %f", inv.TotalHarga)
	}

	if inv.StatusInvoice != "Belum Bayar" {
		t.Errorf("expected StatusInvoice 'Belum Bayar', got %q", inv.StatusInvoice)
	}

	if inv.InvoiceType != "automatic" {
		t.Errorf("expected InvoiceType 'automatic', got %q", inv.InvoiceType)
	}
}

