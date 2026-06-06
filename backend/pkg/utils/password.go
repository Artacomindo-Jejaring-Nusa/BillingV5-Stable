package utils

import (
	"strings"
	"unicode"
)

// PasswordRequirements defines the password policy
type PasswordRequirements struct {
	MinLength        int  `json:"min_length"`
	RequireUppercase bool `json:"require_uppercase"`
	RequireLowercase bool `json:"require_lowercase"`
	RequireDigit     bool `json:"require_digit"`
	RequireSpecial   bool `json:"require_special"`
	NoWhitespace     bool `json:"no_whitespace"`
}

// DefaultPasswordRequirements returns the default password policy
func DefaultPasswordRequirements() PasswordRequirements {
	return PasswordRequirements{
		MinLength:        8,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigit:     true,
		RequireSpecial:   true,
		NoWhitespace:     true,
	}
}

// ValidatePasswordStrength validates a password against the security requirements.
// Returns (isValid bool, errors []string)
func ValidatePasswordStrength(password string) (bool, []string) {
	var errs []string
	req := DefaultPasswordRequirements()

	if len(password) < req.MinLength {
		errs = append(errs, "Password minimal 8 karakter")
	}

	if req.RequireUppercase {
		hasUpper := false
		for _, c := range password {
			if unicode.IsUpper(c) {
				hasUpper = true
				break
			}
		}
		if !hasUpper {
			errs = append(errs, "Password harus mengandung huruf besar")
		}
	}

	if req.RequireLowercase {
		hasLower := false
		for _, c := range password {
			if unicode.IsLower(c) {
				hasLower = true
				break
			}
		}
		if !hasLower {
			errs = append(errs, "Password harus mengandung huruf kecil")
		}
	}

	if req.RequireDigit {
		hasDigit := false
		for _, c := range password {
			if unicode.IsDigit(c) {
				hasDigit = true
				break
			}
		}
		if !hasDigit {
			errs = append(errs, "Password harus mengandung angka")
		}
	}

	if req.RequireSpecial {
		hasSpecial := false
		for _, c := range password {
			if unicode.IsPunct(c) || unicode.IsSymbol(c) {
				hasSpecial = true
				break
			}
		}
		if !hasSpecial {
			errs = append(errs, "Password harus mengandung karakter spesial")
		}
	}

	if req.NoWhitespace && strings.ContainsAny(password, " \t\n\r") {
		errs = append(errs, "Password tidak boleh mengandung spasi")
	}

	return len(errs) == 0, errs
}
