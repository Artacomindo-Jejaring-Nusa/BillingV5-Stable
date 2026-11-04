package service

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"billing-backend/config"
	"billing-backend/internal/domain"
)

type mockMLService struct {
	reply         string
	needsHandover bool
	err           error
}

func (m *mockMLService) GenerateReply(ctx context.Context, room *domain.ChatRoom, history []domain.ChatMessage, userMessage string) (string, error) {
	return m.reply, m.err
}

func (m *mockMLService) GenerateReplyWithHandover(ctx context.Context, room *domain.ChatRoom, history []domain.ChatMessage, userMessage string) (string, bool, error) {
	return m.reply, m.needsHandover, m.err
}

func (m *mockMLService) IsHumanHandoverRequested(message string) bool {
	return strings.Contains(strings.ToLower(message), "cs")
}

func TestAIService_HumanHandoverDetection(t *testing.T) {
	cfg := &config.Config{AIRouterEnabled: false}
	svc := NewAIService(cfg, nil)

	handoverInputs := []string{
		"cs", "2", "bicara dengan cs", "hubungkan ke cs", "minta admin", "operator manusia",
		"anjing", "babi", "goblok",
	}
	for _, input := range handoverInputs {
		if !svc.IsHumanHandoverRequested(input) {
			t.Errorf("Expected IsHumanHandoverRequested to return true for input '%s'", input)
		}
	}

	normalInputs := []string{
		"halo", "berapa tagihan saya?", "cara bayar gimana", "wifi saya los merah",
	}
	for _, input := range normalInputs {
		if svc.IsHumanHandoverRequested(input) {
			t.Errorf("Expected IsHumanHandoverRequested to return false for input '%s'", input)
		}
	}
}

func TestAIService_FallbackToMLWhen9RouterFails(t *testing.T) {
	cfg := &config.Config{
		AIRouterEnabled: true,
		AIRouterURL:     "https://invalid-host-for-test.example.com",
		AIRouterKey:     "fake_key",
		AIRouterModel:   "fake_model",
	}

	mockML := &mockMLService{
		reply:         "Ini balasan dari ML Local",
		needsHandover: false,
		err:           nil,
	}

	svc := NewAIServiceWithFallback(cfg, nil, mockML)

	room := &domain.ChatRoom{
		ID:    1,
		Brand: "Jakinet",
		Pelanggan: &domain.Pelanggan{
			Nama: "Budi Santoso",
		},
	}

	reply, needsHandover, err := svc.GenerateReplyWithHandover(context.Background(), room, nil, "internet saya lemot")
	if err != nil {
		t.Fatalf("Expected no error on ML fallback, got: %v", err)
	}

	if reply != "Ini balasan dari ML Local" {
		t.Errorf("Expected reply from ML Local, got: %s", reply)
	}

	if needsHandover {
		t.Errorf("Expected needsHandover=false, got true")
	}
}

func TestAIService_EmergencyFallbackWhenBothFail(t *testing.T) {
	cfg := &config.Config{
		AIRouterEnabled: true,
		AIRouterURL:     "https://invalid-host-for-test.example.com",
		AIRouterKey:     "fake_key",
		AIRouterModel:   "fake_model",
	}

	mockML := &mockMLService{
		reply: "",
		err:   fmt.Errorf("ML service unreachable"),
	}

	svc := NewAIServiceWithFallback(cfg, nil, mockML)

	room := &domain.ChatRoom{
		ID:    2,
		Brand: "Jakinet",
		Pelanggan: &domain.Pelanggan{
			Nama: "Siti Rahma",
		},
	}

	reply, needsHandover, err := svc.GenerateReplyWithHandover(context.Background(), room, nil, "tolong bantu cek tagihan")
	if err != nil {
		t.Fatalf("Expected no error on emergency fallback, got: %v", err)
	}

	if !strings.Contains(reply, "Siti Rahma") || !strings.Contains(reply, "Jakinet") {
		t.Errorf("Expected reply to contain customer name and brand, got: %s", reply)
	}

	if !needsHandover {
		t.Errorf("Expected needsHandover=true on emergency fallback, got false")
	}
}

func TestAIService_DirectHandoverOnCSRequest(t *testing.T) {
	cfg := &config.Config{AIRouterEnabled: true}
	svc := NewAIService(cfg, nil)

	room := &domain.ChatRoom{
		ID:    3,
		Brand: "Jelantik",
		Pelanggan: &domain.Pelanggan{
			Nama: "Ahmad",
		},
	}

	reply, needsHandover, err := svc.GenerateReplyWithHandover(context.Background(), room, nil, "bicara dengan cs")
	if err != nil {
		t.Fatalf("Expected no error on CS request, got: %v", err)
	}

	if !needsHandover {
		t.Errorf("Expected needsHandover=true on CS request, got false")
	}

	if !strings.Contains(reply, "Customer Support Jelantik") {
		t.Errorf("Expected reply to mention Customer Support Jelantik, got: %s", reply)
	}
}
