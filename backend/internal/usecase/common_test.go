package usecase

import (
	"context"
	"billing-backend/internal/domain"
)

// Shared Mocks for Usecase Tests

type mockPelangganRepo struct {
	domain.PelangganRepository
	data map[uint64]*domain.Pelanggan
}

func (m *mockPelangganRepo) GetAll(ctx context.Context, limit, offset int, connectionStatus string) ([]domain.Pelanggan, int64, error) {
	var result []domain.Pelanggan
	for _, p := range m.data {
		result = append(result, *p)
	}
	return result, int64(len(result)), nil
}

func (m *mockPelangganRepo) GetByID(ctx context.Context, id uint64) (*domain.Pelanggan, error) {
	return m.data[id], nil
}

func (m *mockPelangganRepo) Create(ctx context.Context, p *domain.Pelanggan) error {
	if m.data == nil { m.data = make(map[uint64]*domain.Pelanggan) }
	p.ID = uint64(len(m.data) + 1)
	m.data[p.ID] = p
	return nil
}

func (m *mockPelangganRepo) GetByEmail(ctx context.Context, email string) (*domain.Pelanggan, error) {
	for _, p := range m.data {
		if p.Email == email {
			return p, nil
		}
	}
	return nil, nil
}

func (m *mockPelangganRepo) GetUniqueLocations(ctx context.Context) ([]string, error) {
	return []string{"Tambun", "Bekasi"}, nil
}

func (m *mockPelangganRepo) Update(ctx context.Context, p *domain.Pelanggan) error {
	m.data[p.ID] = p
	return nil
}

func (m *mockPelangganRepo) Delete(ctx context.Context, id uint64) error {
	delete(m.data, id)
	return nil
}

func (m *mockPelangganRepo) GetByEmails(ctx context.Context, emails []string) ([]domain.Pelanggan, error) {
	var res []domain.Pelanggan
	for _, e := range emails {
		for _, p := range m.data {
			if p.Email == e { res = append(res, *p) }
		}
	}
	return res, nil
}

type mockLanggananRepo struct {
	domain.LanggananRepository
	data map[uint64]*domain.Langganan
}

func (m *mockLanggananRepo) GetAll(ctx context.Context, limit, offset int, search, status string, forInvoiceSelection bool) ([]domain.Langganan, int64, error) {
	var res []domain.Langganan
	for _, l := range m.data { res = append(res, *l) }
	return res, int64(len(res)), nil
}

func (m *mockLanggananRepo) Create(ctx context.Context, l *domain.Langganan) error {
	if m.data == nil { m.data = make(map[uint64]*domain.Langganan) }
	l.ID = uint64(len(m.data) + 1)
	m.data[l.ID] = l
	return nil
}

func (m *mockLanggananRepo) GetByID(ctx context.Context, id uint64) (*domain.Langganan, error) {
	return m.data[id], nil
}

func (m *mockLanggananRepo) Update(ctx context.Context, l *domain.Langganan) error {
	m.data[l.ID] = l
	return nil
}

type mockInvoiceRepoCallback struct {
	domain.InvoiceRepository
	invoices        map[string]*domain.Invoice
	updatedInvoices []*domain.Invoice
}

func (m *mockInvoiceRepoCallback) GetInvoiceWithRelations(ctx context.Context, externalID string) (*domain.Invoice, error) {
	return m.invoices[externalID], nil
}

func (m *mockInvoiceRepoCallback) Update(ctx context.Context, invoice *domain.Invoice) error {
	m.updatedInvoices = append(m.updatedInvoices, invoice)
	return nil
}

func (m *mockInvoiceRepoCallback) GetAll(ctx context.Context, limit, offset int) ([]domain.Invoice, int64, error) {
	var res []domain.Invoice
	for _, inv := range m.invoices { res = append(res, *inv) }
	return res, int64(len(res)), nil
}

type mockPaketRepo struct {
	domain.PaketLayananRepository
	data map[uint64]*domain.PaketLayanan
}

func (m *mockPaketRepo) GetByID(ctx context.Context, id uint64) (*domain.PaketLayanan, error) {
	return m.data[id], nil
}

type mockBrandRepo struct {
	domain.HargaLayananRepository
	data map[string]*domain.HargaLayanan
}

func (m *mockBrandRepo) GetByID(ctx context.Context, id string) (*domain.HargaLayanan, error) {
	return m.data[id], nil
}

type mockDataTeknisRepo struct {
	domain.DataTeknisRepository
}

func (m *mockDataTeknisRepo) GetAll(ctx context.Context, skip, limit int, search string, olt string, profile string, vlan string, onuPowerMin, onuPowerMax *int) ([]domain.DataTeknis, int64, error) {
	return []domain.DataTeknis{}, 0, nil
}
