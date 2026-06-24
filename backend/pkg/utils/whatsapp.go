package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type WatzapPayload struct {
	ApiKey    string `json:"api_key"`
	NumberKey string `json:"number_key"`
	PhoneNo   string `json:"phone_no"`
	Message   string `json:"message"`
}

type WatzapResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func SendWhatsAppMessage(apiKey, numberKey, toPhone, message string) error {
	if apiKey == "" || numberKey == "" {
		return errors.New("watzap credentials are not configured")
	}

	url := "https://api.watzap.id/v1/send_message"
	payload := WatzapPayload{
		ApiKey:    apiKey,
		NumberKey: numberKey,
		PhoneNo:   toPhone,
		Message:   message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("watzap api returned status code %d (body: %s)", resp.StatusCode, string(bodyBytes))
	}

	var result WatzapResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return fmt.Errorf("failed to decode json response: %w (body: %s)", err, string(bodyBytes))
	}

	// Watzap API can return Status as "success" or "200", and Message as "The message is being delivered"
	// We treat these as successful cases.
	if result.Status == "success" || result.Status == "200" || result.Message == "The message is being delivered" || result.Message == "Message sent successfully." {
		return nil
	}

	return fmt.Errorf("watzap api error: %s (status: %s, body: %s)", result.Message, result.Status, string(bodyBytes))
}

