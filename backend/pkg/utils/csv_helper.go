package utils

import (
	"bytes"
	"encoding/csv"
	"errors"
	"strings"
	"unicode/utf8"
)

// CleanCSVContent removes UTF-8 BOM, UTF-16 BOM, and normalizes line endings
func CleanCSVContent(content string) string {
	// Strip UTF-8 BOM if present
	content = strings.TrimPrefix(content, "\ufeff")
	content = strings.TrimPrefix(content, "\xef\xbb\xbf")

	// Strip null bytes if any UTF-16 LE/BE conversion artifacts
	if strings.Contains(content, "\x00") {
		content = strings.ReplaceAll(content, "\x00", "")
	}

	return strings.TrimSpace(content)
}

// DetectCSVDelimiter detects the delimiter by analyzing the first line
func DetectCSVDelimiter(content string) rune {
	clean := CleanCSVContent(content)
	firstLine := clean
	if idx := strings.Index(clean, "\n"); idx != -1 {
		firstLine = clean[:idx]
	}
	firstLine = strings.TrimRight(firstLine, "\r")

	semicolons := strings.Count(firstLine, ";")
	commas := strings.Count(firstLine, ",")
	tabs := strings.Count(firstLine, "\t")
	pipes := strings.Count(firstLine, "|")

	if semicolons > commas && semicolons >= tabs && semicolons >= pipes {
		return ';'
	}
	if tabs > commas && tabs > semicolons && tabs >= pipes {
		return '\t'
	}
	if pipes > commas && pipes > semicolons && pipes > tabs {
		return '|'
	}
	return ','
}

// ParseCSV cleans BOM, auto-detects delimiter, and parses CSV content into records
func ParseCSV(content string) ([][]string, error) {
	clean := CleanCSVContent(content)
	if len(clean) == 0 {
		return nil, errors.New("CSV content is empty")
	}

	delimiter := DetectCSVDelimiter(clean)

	reader := csv.NewReader(strings.NewReader(clean))
	reader.Comma = delimiter
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1 // Allow variable number of fields per line (tolerate trailing empty separators)

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// Clean any remaining BOM or whitespace in cells
	for i := range records {
		for j := range records[i] {
			records[i][j] = strings.TrimPrefix(records[i][j], "\ufeff")
			records[i][j] = strings.TrimPrefix(records[i][j], "\xef\xbb\xbf")
			records[i][j] = strings.TrimSpace(records[i][j])
		}
	}

	return records, nil
}

// NormalizeCSVHeader normalizes a CSV column header name (lowercased, spaces to underscores, BOM stripped)
func NormalizeCSVHeader(h string) string {
	h = strings.TrimPrefix(h, "\ufeff")
	h = strings.TrimPrefix(h, "\xef\xbb\xbf")
	h = strings.Trim(h, "\"'` \t\r\n")
	h = strings.ToLower(h)
	h = strings.ReplaceAll(h, " ", "_")
	h = strings.ReplaceAll(h, "-", "_")

	// Ensure valid UTF-8
	if !utf8.ValidString(h) {
		var buf bytes.Buffer
		for _, r := range h {
			if r != utf8.RuneError {
				buf.WriteRune(r)
			}
		}
		h = buf.String()
	}

	return h
}
