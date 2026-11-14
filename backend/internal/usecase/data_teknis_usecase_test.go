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

func (m *mockDataTeknisRepoFull) GetAvailableProfiles(ctx context.Context) ([]string, error) {
	return []string{"10Mbps-a", "10Mbps-b", "10Mbps-c"}, nil
}

func TestDataTeknisStore(t *testing.T) {
	repo := &mockDataTeknisRepoFull{}
	pRepo := &mockPelangganRepo{data: map[uint64]*domain.Pelanggan{
		2: {ID: 2, Nama: "New User"},
	}}
	u := NewDataTeknisUsecase(repo, nil, pRepo, nil, nil, nil)

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
	u := NewDataTeknisUsecase(repo, nil, nil, nil, nil, nil)

	dt := &domain.DataTeknis{IDPelanggan: "UPDATED"}
	err := u.Update(context.Background(), 1, dt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDataTeknisExport(t *testing.T) {
	repo := &mockDataTeknisRepoFull{}
	u := NewDataTeknisUsecase(repo, nil, nil, nil, nil, nil)

	// Test CSV
	_, contentType, err := u.Export(context.Background(), "csv", domain.DataTeknisFilterParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contentType != "text/csv" {
		t.Errorf("expected text/csv, got %s", contentType)
	}

	// Test Excel
	_, contentType, err = u.Export(context.Background(), "excel", domain.DataTeknisFilterParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contentType != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Errorf("expected excel content type, got %s", contentType)
	}
}

func TestAutoSyncProfileForPelanggan(t *testing.T) {
	repo := &mockDataTeknisRepoFull{}
	pRepo := &mockPelangganRepo{data: map[uint64]*domain.Pelanggan{
		1: {ID: 1, Nama: "Pelanggan Test"},
	}}
	u := NewDataTeknisUsecase(repo, nil, pRepo, nil, nil, nil)

	paket := &domain.PaketLayanan{
		ID:        1,
		NamaPaket: "Paket Internet 10 Mbps",
		Kecepatan: 10,
	}

	prof, err := u.AutoSyncProfileForPelanggan(context.Background(), 1, paket)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prof == "" {
		t.Error("expected a non-empty profile name")
	}
}

func TestParseRxPowerToInt(t *testing.T) {
	tests := []struct {
		input    string
		expected *int
	}{
		{"-26.02 dBm", intPtr(-26)},
		{"-24 dBm", intPtr(-24)},
		{"-27.50", intPtr(-28)},
		{"-18.06 dBm (Online)", intPtr(-18)},
		{"", nil},
		{"-", nil},
		{"N/A", nil},
		{"Tidak Terdeteksi", nil},
	}

	for _, tc := range tests {
		actual := parseRxPowerToInt(tc.input)
		if tc.expected == nil {
			if actual != nil {
				t.Errorf("for input %q, expected nil, got %d", tc.input, *actual)
			}
		} else {
			if actual == nil {
				t.Errorf("for input %q, expected %d, got nil", tc.input, *tc.expected)
			} else if *actual != *tc.expected {
				t.Errorf("for input %q, expected %d, got %d", tc.input, *tc.expected, *actual)
			}
		}
	}
}

func intPtr(val int) *int {
	return &val
}

func TestMatchONUToCustomer(t *testing.T) {
	tests := []struct {
		name     string
		onu      domain.ZTEONUInfo
		dt       domain.DataTeknis
		expected bool
	}{
		{
			name: "Exact match IDPelanggan",
			onu: domain.ZTEONUInfo{
				Name: "PRM-B6-2-Mega",
			},
			dt: domain.DataTeknis{
				IDPelanggan: "PRM-B6-2-Mega",
			},
			expected: true,
		},
		{
			name: "Truncated ONU name on HSGQ OLT (16 chars limit)",
			onu: domain.ZTEONUInfo{
				Name: "PRM-B5-19-Anissa",
			},
			dt: domain.DataTeknis{
				IDPelanggan: "PRM-B5-19-Anissah",
				Pelanggan:   &domain.Pelanggan{Nama: "Anissah Rubianti"},
			},
			expected: true,
		},
		{
			name: "Stripped prefix match (without PRM-)",
			onu: domain.ZTEONUInfo{
				Name: "B5-19-Anissah",
			},
			dt: domain.DataTeknis{
				IDPelanggan: "PRM-B5-19-Anissah",
			},
			expected: true,
		},
		{
			name: "Block and unit match",
			onu: domain.ZTEONUInfo{
				Name: "PRM-B5-19",
			},
			dt: domain.DataTeknis{
				IDPelanggan: "PRM-B5-19-Anissah",
			},
			expected: true,
		},
		{
			name: "Customer name word match",
			onu: domain.ZTEONUInfo{
				Name: "PRM-MODEM-ANISSAH",
			},
			dt: domain.DataTeknis{
				IDPelanggan: "CUST-999",
				Pelanggan:   &domain.Pelanggan{Nama: "Anissah Rubianti"},
			},
			expected: true,
		},
		{
			name: "Serial number match with different formatting",
			onu: domain.ZTEONUInfo{
				SerialNumber: "70:2e:22:0a:59:47",
			},
			dt: domain.DataTeknis{
				IDPelanggan: "PRM-A2-26-Ar",
				Sn:          stringPtr("702E220A5947"),
			},
			expected: true,
		},
		{
			name: "Different customer should not match",
			onu: domain.ZTEONUInfo{
				Name: "PRM-A1-43-Ri",
			},
			dt: domain.DataTeknis{
				IDPelanggan: "PRM-B5-19-Anissah",
				Pelanggan:   &domain.Pelanggan{Nama: "Anissah Rubianti"},
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := matchONUToCustomer(&tc.onu, &tc.dt)
			if res != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, res)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
