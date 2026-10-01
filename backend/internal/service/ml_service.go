package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"billing-backend/config"
	"billing-backend/internal/domain"
)

type mlService struct {
	cfg        *config.Config
	httpClient *http.Client
}

func NewMLService(cfg *config.Config) AIService {
	return &mlService{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *mlService) IsHumanHandoverRequested(message string) bool {
	cleaned := strings.ToLower(strings.TrimSpace(message))

	exactMatches := []string{
		"2", "2.", "cs", "agent", "agen", "admin", "operator", "manusia", "orang", "teknisi",
	}
	for _, match := range exactMatches {
		if cleaned == match {
			return true
		}
	}

	phraseMatches := []string{
		"terhubung cs", "hubungkan ke cs", "hubungkan cs", "sambungkan ke cs", "sambungkan cs",
		"terhubung dengan teknisi", "hubungi teknisi", "bantuan teknisi",
		"bicara dengan cs", "bicara dengan admin", "bicara dengan orang", "bicara dengan manusia", "bicara dengan teknisi",
		"mau cs", "minta cs", "panggil cs", "operator manusia", "komplain ke cs",
		"mau orang", "bukan bot", "jangan bot", "customer support", "customer care", "customer service",
		"hubungkan saya", "sambungkan saya", "chat cs", "chat admin", "halo cs", "halo admin",
		"hubungkan dengan cs", "sambungkan dengan cs",
		// Deteksi kata kasar & emosi tinggi untuk dialihkan ke CS
		"anjing", "babi", "bangsat", "goblok", "tolol", "kampret", "kontol", "brengsek",
		"sialan", "bajingan", "jancuk", "pantek", "bego", "asu", "tai",
	}
	for _, phrase := range phraseMatches {
		if strings.Contains(cleaned, phrase) {
			return true
		}
	}

	return false
}

type mlChatRequest struct {
	Message string `json:"message"`
}

type mlChatResponse struct {
	UserMessage        string  `json:"user_message"`
	PredictedIntent    string  `json:"predicted_intent"`
	ConfidenceScore    float64 `json:"confidence_score"`
	BotReply           string  `json:"bot_reply"`
	NeedsHumanHandover bool    `json:"needs_human_handover"`
	IsSolvedByML       bool    `json:"is_solved_by_ml"`
}

func (s *mlService) GenerateReply(ctx context.Context, room *domain.ChatRoom, history []domain.ChatMessage, userMessage string) (string, error) {
	reply, _, err := s.GenerateReplyWithHandover(ctx, room, history, userMessage)
	return reply, err
}

func (s *mlService) GenerateReplyWithHandover(ctx context.Context, room *domain.ChatRoom, history []domain.ChatMessage, userMessage string) (string, bool, error) {
	if s.cfg == nil || s.cfg.MLLocalURL == "" {
		return "", false, fmt.Errorf("ML local service belum dikonfigurasi")
	}

	brand := "Jakinet"
	if room != nil && room.Brand != "" {
		brand = room.Brand
	}

	reqBody := mlChatRequest{Message: userMessage}
	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", false, fmt.Errorf("gagal marshal request ML: %w", err)
	}

	endpoint := strings.TrimRight(s.cfg.MLLocalURL, "/") + "/api/chat"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "", false, fmt.Errorf("gagal membuat request ML: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", false, fmt.Errorf("gagal menghubungi ML service (%s): %w", endpoint, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", false, fmt.Errorf("gagal membaca respons ML: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[MLService] Error from FastAPI HTTP %d: %s", resp.StatusCode, string(bodyBytes))
		return "", false, fmt.Errorf("ML service error HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var mlResp mlChatResponse
	if err := json.Unmarshal(bodyBytes, &mlResp); err != nil {
		return "", false, fmt.Errorf("gagal parsing JSON ML response: %w", err)
	}

	if strings.TrimSpace(mlResp.BotReply) == "" {
		return "", false, fmt.Errorf("balasan ML kosong")
	}

	replyContent := strings.TrimSpace(mlResp.BotReply)

	log.Printf("[MLService] Intent: %s | Confidence: %.3f | Handover: %t | Solved: %t | Room: %d",
		mlResp.PredictedIntent, mlResp.ConfidenceScore, mlResp.NeedsHumanHandover, mlResp.IsSolvedByML, func() uint64 {
			if room != nil {
				return room.ID
			}
			return 0
		}())

	footer := fmt.Sprintf("\n\n---\n🤖 Dibalas otomatis oleh Asisten Virtual AI %s", brand)
	if !strings.Contains(replyContent, "Asisten Virtual AI") {
		replyContent += footer
	}

	return replyContent, mlResp.NeedsHumanHandover, nil
}
