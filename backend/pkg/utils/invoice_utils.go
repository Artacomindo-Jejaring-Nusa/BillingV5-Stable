package utils

import (
	"regexp"
	"strings"
)

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9]`)

// GenerateAlamatSingkat generates a shortened address for the invoice number.
//
// Special handling for Pulogebang location:
// - If blok is "Tower" -> "PGTower{unit}"
// - If blok is anything else -> "PGBlok{blok}{unit}"
//
// Other locations:
// - Sanitize, uppercase, take first max_len characters.
func GenerateAlamatSingkat(alamat string, blok string, unit string, maxLen int) string {
	alamatLower := strings.ToLower(alamat)
	keywords := []string{"pulogebang", "pulo gebang", "pulog", "rusun pulogebang", "rusun pulo gebang"}
	isPulogebang := false
	for _, kw := range keywords {
		if strings.Contains(alamatLower, kw) {
			isPulogebang = true
			break
		}
	}

	if isPulogebang && strings.TrimSpace(blok) != "" {
		blokClean := strings.TrimSpace(blok)
		unitClean := strings.TrimSpace(unit)
		if strings.ToLower(blokClean) == "tower" {
			return "PGTower" + unitClean
		}
		return "PGBlok" + strings.ToUpper(blokClean) + unitClean
	}

	sanitized := nonAlphanumericRegex.ReplaceAllString(alamat, "")
	sanitized = strings.ToUpper(sanitized)
	if len(sanitized) > maxLen {
		return sanitized[:maxLen]
	}
	return sanitized
}

// NormalizePhoneForXendit normalizes a phone number to +62xxx format for Xendit API.
func NormalizePhoneForXendit(phone string) string {
	cleaned := strings.TrimSpace(phone)
	if cleaned == "" {
		return ""
	}

	// Remove non-digit characters
	var digitsBuilder strings.Builder
	for _, r := range cleaned {
		if r >= '0' && r <= '9' {
			digitsBuilder.WriteRune(r)
		}
	}
	digits := digitsBuilder.String()

	if digits == "" {
		return ""
	}

	// Case 1: Starts with 62
	if strings.HasPrefix(digits, "62") && len(digits) > 4 {
		return "+" + digits
	}

	// Case 2: Starts with 0
	if strings.HasPrefix(digits, "0") {
		return "+62" + digits[1:]
	}

	// Case 3: Starts without prefix
	return "+62" + digits
}

// NormalizePhoneDisplay normalizes a phone number to 62xxx (no +) format.
func NormalizePhoneDisplay(phone string) string {
	cleaned := strings.TrimSpace(phone)
	if cleaned == "" {
		return ""
	}

	var digitsBuilder strings.Builder
	for _, r := range cleaned {
		if r >= '0' && r <= '9' {
			digitsBuilder.WriteRune(r)
		}
	}
	digits := digitsBuilder.String()

	if digits == "" {
		return cleaned
	}

	if strings.HasPrefix(digits, "62") && len(digits) > 4 {
		return digits
	}

	if strings.HasPrefix(digits, "0") {
		return "62" + digits[1:]
	}

	return "62" + digits
}
