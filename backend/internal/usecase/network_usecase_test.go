package usecase

import (
	"context"
	"testing"

	"billing-backend/internal/domain"
)

type mockOLTRepo struct {
	domain.OLTRepository
	snMap   map[string]domain.CustomerDeviceInfo
	allList []domain.CustomerDeviceInfo
}

func (m *mockOLTRepo) GetCustomerDeviceInfoMap(ctx context.Context) (map[string]domain.CustomerDeviceInfo, []domain.CustomerDeviceInfo, error) {
	return m.snMap, m.allList, nil
}

func (m *mockOLTRepo) GetByID(ctx context.Context, id uint64) (*domain.OLT, error) {
	return &domain.OLT{
		ID:      1,
		NamaOlt: "Parama",
		TipeOlt: "HSGQ",
	}, nil
}

func TestEnrichONUList_TruncatedNameAndMACMatch(t *testing.T) {
	mockRepo := &mockOLTRepo{
		snMap: map[string]domain.CustomerDeviceInfo{
			"702e220a5947": {
				IDPelanggan: "PRM-A2-26-Aris",
				Nama:        "Aris Munandar",
				SN:          "70:2e:22:0a:59:47",
			},
		},
		allList: []domain.CustomerDeviceInfo{
			{
				IDPelanggan: "PRM-A2-26-Aris",
				Nama:        "Aris Munandar",
				SN:          "70:2e:22:0a:59:47",
			},
		},
	}

	u := &oltUsecase{
		oltRepo: mockRepo,
	}

	rawONUs := []domain.ZTEONUInfo{
		{
			PON:          1,
			ONUID:        1,
			Name:         "PRM-A2-26-Ar", // truncated by HSGQ CLI
			SerialNumber: "70:2e:22:0a:59:47",
			Status:       "Online",
		},
	}

	enriched := u.enrichONUList(context.Background(), rawONUs)

	if len(enriched) != 1 {
		t.Fatalf("expected 1 ONU, got %d", len(enriched))
	}

	if enriched[0].Name != "PRM-A2-26-Aris" {
		t.Errorf("expected full name PRM-A2-26-Aris, got %s", enriched[0].Name)
	}

	if enriched[0].CustomerName != "Aris Munandar" {
		t.Errorf("expected customer name Aris Munandar, got %s", enriched[0].CustomerName)
	}
}

func TestEnrichONUList_PrefixFallback(t *testing.T) {
	mockRepo := &mockOLTRepo{
		snMap: map[string]domain.CustomerDeviceInfo{},
		allList: []domain.CustomerDeviceInfo{
			{
				IDPelanggan: "PRM-A4-28-Muhammad",
				Nama:        "Muhammad Fadhil",
				SN:          "",
			},
		},
	}

	u := &oltUsecase{
		oltRepo: mockRepo,
	}

	rawONUs := []domain.ZTEONUInfo{
		{
			PON:          1,
			ONUID:        4,
			Name:         "PRM-A4-28-Mu", // 12-char truncated name without SN match
			SerialNumber: "10:22:33:44:55:66",
			Status:       "Online",
		},
	}

	enriched := u.enrichONUList(context.Background(), rawONUs)

	if enriched[0].Name != "PRM-A4-28-Muhammad" {
		t.Errorf("expected full name PRM-A4-28-Muhammad, got %s", enriched[0].Name)
	}
	if enriched[0].CustomerName != "Muhammad Fadhil" {
		t.Errorf("expected customer name Muhammad Fadhil, got %s", enriched[0].CustomerName)
	}
}

