package http

import (
	"testing"
)

func TestParseFlexibleDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // YYYY-MM-DD or empty for nil
	}{
		{"Empty string", "", ""},
		{"Dash placeholder", "-", ""},
		{"Double dash placeholder", "--", ""},
		{"N/A placeholder", "N/A", ""},
		{"Null placeholder", "null", ""},
		{"Zero placeholder", "0", ""},
		{"Zero date placeholder", "0000-00-00", ""},
		{"ISO Date YYYY-MM-DD", "2026-09-07", "2026-09-07"},
		{"Slash Date YYYY/MM/DD", "2026/09/07", "2026-09-07"},
		{"Indonesian Slash DD/MM/YYYY", "07/09/2026", "2026-09-07"},
		{"Indonesian Dash DD-MM-YYYY", "07-09-2026", "2026-09-07"},
		{"Short Year DD/MM/YY", "07/09/26", "2026-09-07"},
		{"Single digit Day/Month", "7/9/2026", "2026-09-07"},
		{"Indonesian Month Text", "07 September 2026", "2026-09-07"},
		{"Indonesian Month Short Text", "07-Sep-2026", "2026-09-07"},
		{"Indonesian Month Ags", "17 Ags 2026", "2026-08-17"},
		{"Indonesian Month Okt", "01 Okt 2026", "2026-10-01"},
		{"Excel Serial Date 46272", "46272", "2026-09-07"},
		{"Excel Serial Float 46272.5", "46272.5", "2026-09-07"},
		{"ISO with Timestamp", "2026-09-07 14:30:00", "2026-09-07"},
		{"ISO with T Timestamp", "2026-09-07T00:00:00Z", "2026-09-07"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := parseFlexibleDate(tt.input)
			if tt.expected == "" {
				if res != nil {
					t.Errorf("Expected nil for input '%s', got %v", tt.input, res)
				}
			} else {
				if res == nil {
					t.Errorf("Expected '%s' for input '%s', got nil", tt.expected, tt.input)
				} else {
					formatted := res.Format("2006-01-02")
					if formatted != tt.expected {
						t.Errorf("Expected '%s' for input '%s', got '%s'", tt.expected, tt.input, formatted)
					}
				}
			}
		})
	}
}
