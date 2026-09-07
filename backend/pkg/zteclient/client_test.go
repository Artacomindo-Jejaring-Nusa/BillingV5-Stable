package zteclient_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"billing-backend/internal/domain"
	"billing-backend/pkg/zteclient"
)

func TestZTEClient(t *testing.T) {
	mux := http.NewServeMux()

	// Mock /health
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	// Mock /readyz
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":     "ready",
			"snmp_pinus": "up",
		})
	})

	// Mock /api/v1/olt/pinus/uplinks
	mux.HandleFunc("/api/v1/olt/pinus/uplinks", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code":       401,
				"error_code": "UNAUTHORIZED",
				"message":    "Invalid API key",
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(zteclient.APIResponse[domain.ZTEUplinksData]{
			Code:   200,
			Status: "success",
			Data: domain.ZTEUplinksData{
				Cards: []domain.ZTECardInfo{
					{Slot: 1, Type: "GTGO", Role: "gpon"},
					{Slot: 3, Type: "HUVQ", Role: "uplink"},
				},
				Ports: []domain.ZTEPortInfo{
					{Name: "gei_1/3/1", Slot: 3, Port: 1, Kind: "1G", OperStatus: "up", SpeedMbps: 1000},
				},
			},
		})
	})

	// Mock /api/v1/olt/pinus/board/1/pon/1
	mux.HandleFunc("/api/v1/olt/pinus/board/1/pon/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(zteclient.APIResponse[[]domain.ZTEONUInfo]{
			Code:   200,
			Status: "success",
			Data: []domain.ZTEONUInfo{
				{
					Board:        1,
					PON:          1,
					ONUID:        1,
					Name:         "Pelanggan-01",
					ONUType:      "F670L",
					SerialNumber: "ZTEGC1234567",
					RxPower:      "-21.50",
					Status:       "Online",
				},
			},
		})
	})

	// Mock /api/v1/olt/pinus/board/1/pon/1/onu/1
	mux.HandleFunc("/api/v1/olt/pinus/board/1/pon/1/onu/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(zteclient.APIResponse[domain.ZTEONUDetail]{
			Code:   200,
			Status: "success",
			Data: domain.ZTEONUDetail{
				Board:                1,
				PON:                  1,
				ONUID:                1,
				Name:                 "Pelanggan-01",
				RxPower:              "-21.50",
				TxPower:              "2.10",
				Status:               "Online",
				GponOpticalDistance:  "1500",
				Uptime:               "2 days 4 hours",
				OfflineReason:        "None",
			},
		})
	})

	// Mock /api/v1/olt/pinus/board/1/pon/1/onu_id/empty
	mux.HandleFunc("/api/v1/olt/pinus/board/1/pon/1/onu_id/empty", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(zteclient.APIResponse[[]int]{
			Code:   200,
			Status: "success",
			Data:   []int{2, 3, 4, 5},
		})
	})

	// Mock /api/v1/olt/pinus/board/1/pon/1/cache/clear
	mux.HandleFunc("/api/v1/olt/pinus/board/1/pon/1/cache/clear", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "cleared"})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := zteclient.NewClient(server.URL, "test-api-key")
	ctx := context.Background()

	// 1. Health
	health, err := client.CheckHealth(ctx)
	if err != nil {
		t.Fatalf("CheckHealth failed: %v", err)
	}
	if health["status"] != "healthy" {
		t.Errorf("expected healthy, got %v", health["status"])
	}

	// 2. Uplinks
	uplinks, err := client.GetUplinks(ctx, "pinus")
	if err != nil {
		t.Fatalf("GetUplinks failed: %v", err)
	}
	if len(uplinks.Cards) != 2 {
		t.Errorf("expected 2 cards, got %d", len(uplinks.Cards))
	}

	// 3. ONUs
	onus, err := client.GetONUs(ctx, "pinus", 1, 1)
	if err != nil {
		t.Fatalf("GetONUs failed: %v", err)
	}
	if len(onus) != 1 || onus[0].SerialNumber != "ZTEGC1234567" {
		t.Errorf("unexpected ONUs data: %+v", onus)
	}

	// 4. ONU Detail
	detail, err := client.GetONUDetail(ctx, "pinus", 1, 1, 1)
	if err != nil {
		t.Fatalf("GetONUDetail failed: %v", err)
	}
	if detail.RxPower != "-21.50" || detail.GponOpticalDistance != "1500" {
		t.Errorf("unexpected ONU Detail data: %+v", detail)
	}

	// 5. Empty ONU IDs
	emptyIDs, err := client.GetEmptyONUIDs(ctx, "pinus", 1, 1)
	if err != nil {
		t.Fatalf("GetEmptyONUIDs failed: %v", err)
	}
	if len(emptyIDs) != 4 || emptyIDs[0] != 2 {
		t.Errorf("unexpected empty IDs: %+v", emptyIDs)
	}

	// 6. Clear Cache
	if err := client.ClearCache(ctx, "pinus", 1, 1); err != nil {
		t.Fatalf("ClearCache failed: %v", err)
	}
}
