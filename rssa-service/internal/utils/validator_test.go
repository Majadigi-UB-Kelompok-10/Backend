package utils

import (
	"testing"
)

func TestSanitizeQueryString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		wantErr   bool
		wantEmpty bool
	}{
		{"valid query", "poli jantung", 100, false, false},
		{"empty input", "", 100, false, true},
		{"over max length", "abc", 2, true, false},
		{"special chars stripped", "drop table;", 100, false, false},
		{"only safe chars stripped leaves empty", "!!!", 100, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := SanitizeQueryString(tt.input, tt.maxLength, "q")
			if tt.wantErr && err == nil {
				t.Errorf("SanitizeQueryString(%q) expected error, got nil", tt.input)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("SanitizeQueryString(%q) unexpected error: %v", tt.input, err)
			}
			if tt.wantEmpty && out != "" {
				t.Errorf("SanitizeQueryString(%q) expected empty output, got %q", tt.input, out)
			}
		})
	}
}

func TestValidateSlug_RSSA(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"layanan-rssa", false},
		{"poli123", false},
		{"a", false},
		{"", true},
		{"With Capital", true},
		{"has space", true},
		{"special!", true},
		{"double--hyphen", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := ValidateSlug(tt.input, "slug")
			if tt.wantErr && err == nil {
				t.Errorf("ValidateSlug(%q) expected error, got nil", tt.input)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateSlug(%q) unexpected error: %v", tt.input, err)
			}
		})
	}
}

func TestParsePagination(t *testing.T) {
	page, limit, offset := ParsePagination("", "")
	if page != 1 || limit != 10 || offset != 0 {
		t.Errorf("defaults: got page=%d limit=%d offset=%d", page, limit, offset)
	}
	page, limit, offset = ParsePagination("3", "15")
	if page != 3 || limit != 15 || offset != 30 {
		t.Errorf("page=3,limit=15: got page=%d limit=%d offset=%d", page, limit, offset)
	}
	_, limit, _ = ParsePagination("1", "500")
	if limit != 100 {
		t.Errorf("limit should be clamped to 100, got %d", limit)
	}
	page, _, _ = ParsePagination("0", "10")
	if page != 1 {
		t.Errorf("page<1 should default to 1, got %d", page)
	}
}
