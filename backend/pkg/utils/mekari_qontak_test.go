package utils

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateHMACAuth(t *testing.T) {
	clientID := "XmjyqN8Lsmken6Rt"
	clientSecret := "7i6nmavTR2wX0EwPcvifrzMFw4KCGT8W"
	client := NewMekariQontakClient(clientID, clientSecret, "channel_id_123")

	// Fixed time: Wed, 10 Nov 2021 07:24:29 GMT
	fixedTime := time.Date(2021, time.November, 10, 7, 24, 29, 0, time.UTC)
	authHeader, dateStr := client.GenerateHMACAuth("POST", "/qontak/chat/v1/broadcasts/whatsapp/direct", fixedTime)

	expectedDate := "Wed, 10 Nov 2021 07:24:29 GMT"
	if dateStr != expectedDate {
		t.Errorf("Expected date %s, got %s", expectedDate, dateStr)
	}

	if !strings.HasPrefix(authHeader, `hmac username="XmjyqN8Lsmken6Rt", algorithm="hmac-sha256", headers="date request-line", signature=`) {
		t.Errorf("Unexpected Authorization header format: %s", authHeader)
	}
}

func TestFormatPhoneNumber62(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"081234567890", "6281234567890"},
		{"6281234567890", "6281234567890"},
		{"+6281234567890", "6281234567890"},
		{" 0855123456 ", "62855123456"},
	}

	for _, tt := range tests {
		got := FormatPhoneNumber62(tt.input)
		if got != tt.expected {
			t.Errorf("FormatPhoneNumber62(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}
