package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// MekariQontakClient manages communication with Mekari Qontak WhatsApp Broadcast API via HMAC Auth
type MekariQontakClient struct {
	ClientID             string
	ClientSecret         string
	ChannelIntegrationID string
	HTTPClient           *http.Client
}

// QontakParamBody represents template variable parameter in body (e.g. {{1}}, {{2}})
type QontakParamBody struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	ValueText string `json:"value_text"`
}

// QontakParamButton represents dynamic URL button parameter if template has URL button
type QontakParamButton struct {
	Index string `json:"index"`
	Type  string `json:"type"`  // e.g. "url"
	Value string `json:"value"` // e.g. suffix path or parameter for dynamic URL button
}

// QontakParameters holds parameters array for body and buttons
type QontakParameters struct {
	Body    []QontakParamBody   `json:"body,omitempty"`
	Buttons []QontakParamButton `json:"buttons,omitempty"`
}

// QontakLanguage language setting for template
type QontakLanguage struct {
	Code string `json:"code"`
}

// DirectBroadcastPayload payload structure for POST /qontak/chat/v1/broadcasts/whatsapp/direct
type DirectBroadcastPayload struct {
	ToName               string           `json:"to_name"`
	ToNumber             string           `json:"to_number"`
	MessageTemplateID    string           `json:"message_template_id"`
	ChannelIntegrationID string           `json:"channel_integration_id"`
	Language             QontakLanguage   `json:"language"`
	Parameters           QontakParameters `json:"parameters"`
}

// DirectBroadcastResponse API response structure from Mekari Qontak
type DirectBroadcastResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
	Error  interface{} `json:"error"`
}

// NewMekariQontakClient creates client instance
func NewMekariQontakClient(clientID, clientSecret, channelIntegrationID string) *MekariQontakClient {
	return &MekariQontakClient{
		ClientID:             clientID,
		ClientSecret:         clientSecret,
		ChannelIntegrationID: channelIntegrationID,
		HTTPClient:           &http.Client{Timeout: 15 * time.Second},
	}
}

// GenerateHMACAuth builds the RFC7231 Date header and HMAC-SHA256 Authorization header for Mekari API
func (c *MekariQontakClient) GenerateHMACAuth(method, path string, now time.Time) (authHeader string, dateStr string) {
	// 1. Format date according to RFC 7231 (HTTP-date)
	dateStr = now.UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")

	// 2. Request line string format: "METHOD PATH HTTP/1.1"
	requestLine := fmt.Sprintf("%s %s HTTP/1.1", strings.ToUpper(method), path)

	// 3. Payload to sign: "date: <DATE>\n<REQUEST_LINE>"
	stringToSign := fmt.Sprintf("date: %s\n%s", dateStr, requestLine)

	// 4. Calculate HMAC-SHA256 hash using ClientSecret
	h := hmac.New(sha256.New, []byte(c.ClientSecret))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// 5. Construct Authorization Header value
	authHeader = fmt.Sprintf(`hmac username="%s", algorithm="hmac-sha256", headers="date request-line", signature="%s"`, c.ClientID, signature)

	return authHeader, dateStr
}

// SendDirectBroadcast sends a WhatsApp message template directly via Mekari Qontak API using HMAC Auth
func (c *MekariQontakClient) SendDirectBroadcast(payload DirectBroadcastPayload) (*DirectBroadcastResponse, error) {
	if c.ClientID == "" || c.ClientSecret == "" {
		return nil, fmt.Errorf("mekari qontak client_id and client_secret must be configured")
	}

	apiPath := "/qontak/chat/v1/broadcasts/whatsapp/direct"
	fullURL := "https://api.mekari.com" + apiPath

	if payload.ChannelIntegrationID == "" {
		payload.ChannelIntegrationID = c.ChannelIntegrationID
	}
	if payload.Language.Code == "" {
		payload.Language.Code = "id"
	}

	// Sanitize phone number (convert 08xx to 628xx)
	payload.ToNumber = FormatPhoneNumber62(payload.ToNumber)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal broadcast payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, fullURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	now := time.Now()
	authHeader, dateStr := c.GenerateHMACAuth(http.MethodPost, apiPath, now)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Date", dateStr)
	req.Header.Set("Authorization", authHeader)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute http request to mekari: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("mekari qontak api error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result DirectBroadcastResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response json: %w (body: %s)", err, string(bodyBytes))
	}

	return &result, nil
}

// FormatPhoneNumber62 ensures phone numbers start with 62 instead of 0
func FormatPhoneNumber62(phone string) string {
	phone = strings.TrimSpace(phone)
	if strings.HasPrefix(phone, "0") {
		return "62" + phone[1:]
	}
	if strings.HasPrefix(phone, "+") {
		return phone[1:]
	}
	return phone
}
