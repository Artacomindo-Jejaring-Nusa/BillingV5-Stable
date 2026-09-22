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
		"2", "cs", "agent", "admin", "operator", "manusia", "orang",
	}
	for _, match := range exactMatches {
		if cleaned == match {
			return true
		}
	}

	phraseMatches := []string{
		"hubungkan ke cs", "hubungkan cs", "sambungkan ke cs", "sambungkan cs",
		"bicara dengan cs", "bicara dengan admin", "bicara dengan orang", "bicara dengan manusia",
		"mau cs", "minta cs", "panggil cs", "operator manusia", "komplain ke cs",
		"mau orang", "bukan bot", "jangan bot", "customer support",
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
	UserMessage    string  `json:"user_message"`
	PredictedIntent string `json:"predicted_intent"`
	ConfidenceScore float64 `json:"confidence_score"`
	BotReply       string  `json:"bot_reply"`
}

func (s *mlService) GenerateReply(ctx context.Context, room *domain.ChatRoom, history []domain.ChatMessage, userMessage string) (string, error) {
	if s.cfg == nil || s.cfg.MLLocalURL == "" {
		return "", fmt.Errorf("ML local service belum dikonfigurasi")
	}

	brand := "Jakinet"
	if room != nil && room.Brand != "" {
		brand = room.Brand
	}

	reqBody := mlChatRequest{Message: userMessage}
	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("gagal marshal request ML: %w", err)
	}

	endpoint := strings.TrimRight(s.cfg.MLLocalURL, "/") + "/api/chat"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "", fmt.Errorf("gagal membuat request ML: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("gagal menghubungi ML service (%s): %w", endpoint, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("gagal membaca respons ML: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[MLService] Error from FastAPI HTTP %d: %s", resp.StatusCode, string(bodyBytes))
		return "", fmt.Errorf("ML service error HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var mlResp mlChatResponse
	if err := json.Unmarshal(bodyBytes, &mlResp); err != nil {
		return "", fmt.Errorf("gagal parsing JSON ML response: %w", err)
	}

	if strings.TrimSpace(mlResp.BotReply) == "" {
		return "", fmt.Errorf("balasan ML kosong")
	}

	replyContent := strings.TrimSpace(mlResp.BotReply)

	log.Printf("[MLService] Intent: %s | Confidence: %.3f | Room: %d",
		mlResp.PredictedIntent, mlResp.ConfidenceScore, func() uint64 {
			if room != nil {
				return room.ID
			}
			return 0
		}())

	footer := fmt.Sprintf("\n\n---\n🤖 Dibalas otomatis oleh Asisten Virtual AI %s", brand)
	if !strings.Contains(replyContent, "Asisten Virtual AI") {
		replyContent += footer
	}

	return replyContent, nil
}
