package usecase

import (
	"context"
	"testing"
	"time"

	"billing-backend/config"
	"billing-backend/internal/domain"
)

// Mock repos
type mockInvoiceRepo struct {
	domain.InvoiceRepository
}

type mockLanggananRepo struct {
	domain.LanggananRepository
	data      map[uint64]*domain.Langganan
	created   []*domain.Langganan
	updated   []*domain.Langganan
	nextID    uint64
}

func (m *mockLanggananRepo) Create(ctx context.Context, langganan *domain.Langganan) error {
	m.nextID++
	langganan.ID = m.nextID
	m.created = append(m.created, langganan)
	m.data[langganan.ID] = langganan
	return nil
}

func (m *mockLanggananRepo) GetByID(ctx context.Context, id uint64) (*domain.Langganan, error) {
	if l, exists := m.data[id]; exists {
		return l, nil
	}
	return nil, nil
}

func (m *mockLanggananRepo) Update(ctx context.Context, langganan *domain.Langganan) error {
	m.updated = append(m.updated, langganan)
	m.data[langganan.ID] = langganan
	return nil
}

type mockPelangganRepo struct {
	domain.PelangganRepository
	data map[uint64]*domain.Pelanggan
}

func (m *mockPelangganRepo) GetByID(ctx context.Context, id uint64) (*domain.Pelanggan, error) {
	if p, exists := m.data[id]; exists {
		return p, nil
	}
	return nil, nil
}

func (m *mockPelangganRepo) GetByNoKtp(ctx context.Context, noKtp string) (*domain.Pelanggan, error) {
	for _, p := range m.data {
		if p.NoKtp == noKtp {
			return p, nil
		}
	}
	return nil, nil
}


type mockPaketRepo struct {
	domain.PaketLayananRepository
	data map[uint64]*domain.PaketLayanan
}

func (m *mockPaketRepo) GetByID(ctx context.Context, id uint64) (*domain.PaketLayanan, error) {
	if p, exists := m.data[id]; exists {
		return p, nil
	}
	return nil, nil
}

type mockBrandRepo struct {
	domain.HargaLayananRepository
	data map[string]*domain.HargaLayanan
}

func (m *mockBrandRepo) GetByID(ctx context.Context, idBrand string) (*domain.HargaLayanan, error) {
	if b, exists := m.data[idBrand]; exists {
		return b, nil
	}
	return nil, nil
}

func TestCalculatePrice(t *testing.T) {
	brandID := "indihome"
	brand := &domain.HargaLayanan{
		IDBrand: brandID,
		Brand:   "Indihome",
		Pajak:   11.0, // 11% tax
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

		// Next month 1st: 2026-07-01
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

		// June has 30 days. Starting 15th means 30 - 15 + 1 = 16 remaining days.
		// Price per day = 300,000 / 30 = 10,000
		// Prorated = 10,000 * 16 = 160,000
		// With 11% tax = 160,000 * 1.11 = 177,600
		if res.HargaAwal != 177600 {
			t.Errorf("expected HargaAwal to be 177600, got %f", res.HargaAwal)
		}

		// Last day of month: 2026-06-30
		expectedTempo := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)
		if !res.TglJatuhTempo.Equal(expectedTempo) {
			t.Errorf("expected TglJatuhTempo to be %v, got %v", expectedTempo, res.TglJatuhTempo)
		}
	})
}

func TestCalculateProratePlusFull(t *testing.T) {
	brandID := "indihome"
	brand := &domain.HargaLayanan{
		IDBrand: brandID,
		Brand:   "Indihome",
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

	// Normal full month = 300,000 * 1.11 = 333,000
	// Prorated 16 days = 177,600
	// Total = 333,000 + 177,600 = 510,600
	if res.HargaProrate != 177600 {
		t.Errorf("expected HargaProrate to be 177600, got %f", res.HargaProrate)
	}
	if res.HargaNormal != 333000 {
		t.Errorf("expected HargaNormal to be 333000, got %f", res.HargaNormal)
	}
	if res.HargaTotalAwal != 510600 {
		t.Errorf("expected HargaTotalAwal to be 510600, got %f", res.HargaTotalAwal)
	}
}

func TestCalculateProrate(t *testing.T) {
	brandID := "brand-a"
	brand := &domain.HargaLayanan{
		IDBrand: brandID,
		Brand:   "Brand A",
		Pajak:   11.0,
	}
	paket := &domain.PaketLayanan{
		ID:        1,
		NamaPaket: "10Mbps",
		Harga:     300000.0,
	}

	pkRepo := &mockPaketRepo{data: map[uint64]*domain.PaketLayanan{1: paket}}
	bRepo := &mockBrandRepo{data: map[string]*domain.HargaLayanan{brandID: brand}}
	u := NewBillingUsecase(nil, nil, nil, pkRepo, bRepo, nil, nil, nil, nil, nil)

	tglMulai := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)

	t.Run("without next month PPN", func(t *testing.T) {
		req := &domain.ProrateCalculationRequest{
			PaketLayananID:       1,
			IDBrand:              brandID,
			TglMulai:             &tglMulai,
			IncludePpnNextMonth: false,
		}
		res, err := u.CalculateProrate(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Basic prorated daily logic:
		// 16 days remaining
		// harga_dasar_prorate = 300,000 / 30 * 16 = 160,000
		// pajak_mentah = 160,000 * 0.11 = 17,600
		// total = 177,600
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
		if res.TotalKeseluruhan != nil {
			t.Errorf("expected TotalKeseluruhan to be nil")
		}
	})

	t.Run("with next month PPN", func(t *testing.T) {
		req := &domain.ProrateCalculationRequest{
			PaketLayananID:       1,
			IDBrand:              brandID,
			TglMulai:             &tglMulai,
			IncludePpnNextMonth: true,
		}
		res, err := u.CalculateProrate(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if res.TotalHargaProrate != 177600 {
			t.Errorf("expected total prorate 177600, got %f", res.TotalHargaProrate)
		}
		if res.HargaBulanDepan == nil || *res.HargaBulanDepan != 300000 {
			t.Errorf("expected HargaBulanDepan 300000")
		}
		if res.PpnBulanDepan == nil || *res.PpnBulanDepan != 33000 {
			t.Errorf("expected PpnBulanDepan 33000")
		}
		if res.TotalBulanDepanDenganPpn == nil || *res.TotalBulanDepanDenganPpn != 333000 {
			t.Errorf("expected TotalBulanDepanDenganPpn 333000")
		}
		if res.TotalKeseluruhan == nil || *res.TotalKeseluruhan != 510600 {
			t.Errorf("expected TotalKeseluruhan 510600, got %f", *res.TotalKeseluruhan)
		}
	})
}

func TestCalculateDiskon(t *testing.T) {
	brandID := "brand-a"
	brand := &domain.HargaLayanan{
		IDBrand: brandID,
		Brand:   "Brand A",
		Pajak:   11.0,
	}
	paket := &domain.PaketLayanan{
		ID:        1,
		NamaPaket: "10Mbps",
		Harga:     300000.0,
	}

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

	// 300,000 + 33,000 pajak = 333,000 subtotal
	// 333,000 * 10% diskon = 33,300 diskon amount
	// 333,000 - 33,300 = 299,700 harga final
	if res.SubtotalSebelumDiskon != 333000 {
		t.Errorf("expected subtotal 333000, got %f", res.SubtotalSebelumDiskon)
	}
	if res.DiskonAmount != 33300 {
		t.Errorf("expected diskon 33300, got %f", res.DiskonAmount)
	}
	if res.HargaFinal != 299700 {
		t.Errorf("expected final price 299700, got %f", res.HargaFinal)
	}
}

func TestCreateLanggananValidation(t *testing.T) {
	brandID := "brand-a"
	brand := &domain.HargaLayanan{
		IDBrand: brandID,
		Brand:   "Brand A",
		Pajak:   11.0,
	}
	paket := &domain.PaketLayanan{
		ID:        1,
		NamaPaket: "10Mbps",
		Harga:     300000.0,
	}

	t.Run("CreateLangganan throws error when DataTeknis is missing", func(t *testing.T) {
		pelangganNoDataTeknis := &domain.Pelanggan{
			ID:         1,
			Nama:       "Ahmad Rizki",
			IDBrand:    &brandID,
			DataTeknis: nil, // Missing data teknis
		}

		pRepo := &mockPelangganRepo{data: map[uint64]*domain.Pelanggan{1: pelangganNoDataTeknis}}
		pkRepo := &mockPaketRepo{data: map[uint64]*domain.PaketLayanan{1: paket}}
		bRepo := &mockBrandRepo{data: map[string]*domain.HargaLayanan{brandID: brand}}
		lRepo := &mockLanggananRepo{data: make(map[uint64]*domain.Langganan)}
		u := NewBillingUsecase(nil, lRepo, pRepo, pkRepo, bRepo, nil, nil, nil, nil, nil)

		langganan := &domain.Langganan{
			PelangganID:    1,
			PaketLayananID: 1,
		}

		err := u.CreateLangganan(context.Background(), langganan)
		if err == nil {
			t.Fatalf("expected error due to missing data teknis, got nil")
		}

		expectedErrStr := "Langganan tidak dapat dibuat. Pelanggan 'Ahmad Rizki' belum memiliki data teknis. Tim NOC harus menambahkan data teknis terlebih dahulu sebelum membuat langganan."
		if err.Error() != expectedErrStr {
			t.Errorf("expected error '%s', got '%s'", expectedErrStr, err.Error())
		}
	})

	t.Run("CreateLangganan succeeds when DataTeknis is present", func(t *testing.T) {
		pelangganWithDataTeknis := &domain.Pelanggan{
			ID:      1,
			Nama:    "Ahmad Rizki",
			IDBrand: &brandID,
			DataTeknis: &domain.DataTeknis{
				ID:          1,
				PelangganID: 1,
				IDPelanggan: "ahmad-2026",
			},
		}

		pRepo := &mockPelangganRepo{data: map[uint64]*domain.Pelanggan{1: pelangganWithDataTeknis}}
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

		if len(lRepo.created) != 1 {
			t.Errorf("expected 1 langganan to be created, got %d", len(lRepo.created))
		}
	})
}

type mockInvoiceRepoCallback struct {
	domain.InvoiceRepository
	logs            []*domain.PaymentCallbackLog
	invoices        map[string]*domain.Invoice
	updatedInvoices []*domain.Invoice
}

func (m *mockInvoiceRepoCallback) GetCallbackLog(ctx context.Context, xenditID, externalID, idempotencyKey string) (*domain.PaymentCallbackLog, error) {
	for _, l := range m.logs {
		if (xenditID != "" && l.XenditID == xenditID) || (externalID != "" && l.ExternalID == externalID) {
			return l, nil
		}
	}
	return nil, nil
}

func (m *mockInvoiceRepoCallback) CreateCallbackLog(ctx context.Context, log *domain.PaymentCallbackLog) error {
	m.logs = append(m.logs, log)
	return nil
}

func (m *mockInvoiceRepoCallback) GetInvoiceWithRelations(ctx context.Context, externalID string) (*domain.Invoice, error) {
	if inv, exists := m.invoices[externalID]; exists {
		return inv, nil
	}
	return nil, nil
}

func (m *mockInvoiceRepoCallback) Update(ctx context.Context, invoice *domain.Invoice) error {
	m.updatedInvoices = append(m.updatedInvoices, invoice)
	return nil
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

