package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("watzap api returned status code %d", resp.StatusCode)
	}

	var result WatzapResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.Status != "success" {
		return fmt.Errorf("watzap api error: %s", result.Message)
	}

	return nil
}
