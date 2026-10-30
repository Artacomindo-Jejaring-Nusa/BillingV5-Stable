package fcm

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"billing-backend/pkg/logger"

	"github.com/golang-jwt/jwt/v5"
)

type ServiceAccount struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
}

type PushClient struct {
	serviceAccount *ServiceAccount
	parsedKey      *rsa.PrivateKey
	mu             sync.RWMutex
	cachedToken    string
	tokenExpiry    time.Time
	httpClient     *http.Client
}

var (
	globalClient *PushClient
	clientOnce   sync.Once
)

// GetClient returns a singleton PushClient instance initialized with service account
func GetClient() *PushClient {
	clientOnce.Do(func() {
		candidatePaths := []string{
			"config/firebase-service-account.json",
			"../config/firebase-service-account.json",
			"../../config/firebase-service-account.json",
			"portal-ajnusa-firebase-adminsdk-fbsvc-07b0313b9c.json",
			"c:/laragon/www/portal-jakinet/portal-ajnusa-firebase-adminsdk-fbsvc-07b0313b9c.json",
		}

		var fileBytes []byte
		var err error
		for _, p := range candidatePaths {
			fileBytes, err = os.ReadFile(p)
			if err == nil {
				logger.Info("[FCM] Loaded Firebase Service Account from: %s", p)
				break
			}
		}

		if err != nil || len(fileBytes) == 0 {
			logger.Warn("[FCM] Service account file not found in candidates, FCM direct sending will be disabled: %v", err)
			return
		}

		var sa ServiceAccount
		if err := json.Unmarshal(fileBytes, &sa); err != nil {
			logger.Error("[FCM] Failed to parse service account JSON: %v", err)
			return
		}

		block, _ := pem.Decode([]byte(sa.PrivateKey))
		if block == nil {
			logger.Error("[FCM] Failed to decode PEM block containing private key")
			return
		}

		parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			// Fallback to PKCS1 if needed
			parsedKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				logger.Error("[FCM] Failed to parse RSA private key: %v", err)
				return
			}
		}

		rsaKey, ok := parsedKey.(*rsa.PrivateKey)
		if !ok {
			logger.Error("[FCM] Parsed key is not an RSA private key")
			return
		}

		globalClient = &PushClient{
			serviceAccount: &sa,
			parsedKey:      rsaKey,
			httpClient: &http.Client{
				Timeout: 10 * time.Second,
			},
		}
		logger.Info("[FCM] Client successfully initialized for project: %s", sa.ProjectID)
	})

	return globalClient
}

// getOAuth2Token obtains a fresh OAuth2 Bearer token using Google JWT Assertion
func (c *PushClient) getOAuth2Token() (string, error) {
	c.mu.RLock()
	if c.cachedToken != "" && time.Now().Before(c.tokenExpiry.Add(-2*time.Minute)) {
		token := c.cachedToken
		c.mu.RUnlock()
		return token, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if c.cachedToken != "" && time.Now().Before(c.tokenExpiry.Add(-2*time.Minute)) {
		return c.cachedToken, nil
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iss":   c.serviceAccount.ClientEmail,
		"scope": "https://www.googleapis.com/auth/firebase.messaging",
		"aud":   "https://oauth2.googleapis.com/token",
		"exp":   now.Add(1 * time.Hour).Unix(),
		"iat":   now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = c.serviceAccount.PrivateKeyID

	signedJWT, err := token.SignedString(c.parsedKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT assertion: %w", err)
	}

	data := "grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Ajwt-bearer&assertion=" + signedJWT
	req, err := http.NewRequest("POST", "https://oauth2.googleapis.com/token", bytes.NewBufferString(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("OAuth2 token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OAuth2 token endpoint returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var res struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}

	if err := json.Unmarshal(body, &res); err != nil {
		return "", fmt.Errorf("failed to decode OAuth2 response: %w", err)
	}

	c.cachedToken = res.AccessToken
	c.tokenExpiry = now.Add(time.Duration(res.ExpiresIn) * time.Second)

	return c.cachedToken, nil
}

// SendPushNotification sends an FCM v1 message to a specific device token
func (c *PushClient) SendPushNotification(ctx context.Context, token, title, body string, data map[string]string) error {
	if c == nil || c.serviceAccount == nil {
		return errors.New("FCM client is not initialized")
	}

	accessToken, err := c.getOAuth2Token()
	if err != nil {
		return fmt.Errorf("failed to get FCM access token: %w", err)
	}

	fcmURL := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", c.serviceAccount.ProjectID)

	isChat := false
	if data != nil && (data["type"] == "chat" || data["room_id"] != "") {
		isChat = true
	}

	channelID := "portal_invoice_channel"
	sound := "default"
	if isChat {
		channelID = "portal_chat_channel"
		sound = "chat_notif"
	}

	fcmMessage := map[string]interface{}{
		"message": map[string]interface{}{
			"token": token,
			"notification": map[string]string{
				"title": title,
				"body":  body,
			},
			"data": data,
			"android": map[string]interface{}{
				"priority": "HIGH",
				"notification": map[string]interface{}{
					"channel_id": channelID,
					"sound":      sound,
					"icon":       "@mipmap/launcher_icon",
				},
			},
		},
	}

	jsonBytes, err := json.Marshal(fcmMessage)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fcmURL, bytes.NewReader(jsonBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json; UTF-8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("FCM HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("FCM v1 returned error HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// SendToCustomerTokens sends a notification to all registered tokens belonging to a customer
func SendToCustomerTokens(ctx context.Context, tokens []string, title, body string, data map[string]string) {
	client := GetClient()
	if client == nil {
		logger.Warn("[FCM] Client not initialized, skipping FCM send")
		return
	}

	for _, token := range tokens {
		go func(t string) {
			sendCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := client.SendPushNotification(sendCtx, t, title, body, data); err != nil {
				logger.Warn("[FCM] Failed to send push to token %s...: %v", t[:min(10, len(t))], err)
			} else {
				logger.Info("[FCM] Push successfully sent to token %s...", t[:min(10, len(t))])
			}
		}(token)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

