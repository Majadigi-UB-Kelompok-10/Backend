package utils

import (
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "user@example.com", false},
		{"valid with dot", "first.last@example.com", false},
		{"valid with plus", "user+tag@example.com", false},
		{"valid with subdomain", "user@sub.example.com", false},
		{"valid uppercase becomes lowercase", "User@Example.COM", false},

		{"empty", "", true},
		{"missing @", "userexample.com", true},
		{"missing domain", "user@", true},
		{"missing TLD", "user@example", true},
		{"single char TLD", "user@example.c", true},
		{"spaces", "user @example.com", true},
		{"only @", "@", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateEmail(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateEmail(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("ValidateEmail(%q) unexpected error: %v", tt.input, err)
			}
			if result == "" {
				t.Errorf("ValidateEmail(%q) returned empty string", tt.input)
			}
		})
	}
}

func TestValidatePhone(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid 08 prefix", "081234567890", false},
		{"valid +62 prefix", "+6281234567890", false},
		{"valid 62 prefix", "6281234567890", false},
		{"valid 11 digits", "08123456789", false},
		{"valid 13 digits", "0812345678901", false},

		{"empty", "", true},
		{"too short", "0812", true},
		{"starts with 09", "091234567890", true},
		{"starts with 07", "071234567890", true},
		{"with letters", "08abc1234567", true},
		{"with spaces preserved as input", "081 234 567 890", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidatePhone(tt.input)
			if tt.wantErr && err == nil {
				t.Errorf("ValidatePhone(%q) expected error, got nil", tt.input)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidatePhone(%q) unexpected error: %v", tt.input, err)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid https", "https://example.com", false},
		{"valid http", "http://example.com", false},
		{"valid with path", "https://example.com/path/to/page", false},
		{"valid with query", "https://example.com?q=test", false},
		{"empty (allowed - optional)", "", false},

		{"missing scheme", "example.com", true},
		{"javascript scheme", "javascript:alert(1)", true},
		{"ftp scheme", "ftp://example.com", true},
		{"just text", "not a url", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateURL(tt.input, "test_url")
			if tt.wantErr && err == nil {
				t.Errorf("ValidateURL(%q) expected error, got nil", tt.input)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateURL(%q) unexpected error: %v", tt.input, err)
			}
		})
	}
}

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid lowercase", "550e8400-e29b-41d4-a716-446655440000", false},
		{"valid uppercase", "550E8400-E29B-41D4-A716-446655440000", false},
		{"valid mixed case", "550E8400-e29b-41D4-a716-446655440000", false},

		{"empty", "", true},
		{"missing dashes", "550e8400e29b41d4a716446655440000", true},
		{"too short", "550e8400-e29b-41d4-a716", true},
		{"non-hex chars", "550e8400-e29b-41d4-a716-zzzzzzzzzzzz", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateUUID(tt.input, "test_id")
			if tt.wantErr && err == nil {
				t.Errorf("ValidateUUID(%q) expected error, got nil", tt.input)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateUUID(%q) unexpected error: %v", tt.input, err)
			}
		})
	}
}

func TestValidatePaginationParams(t *testing.T) {
	tests := []struct {
		name       string
		pageStr    string
		limitStr   string
		wantPage   int
		wantLimit  int
		wantOffset int
	}{
		{"defaults when empty", "", "", 1, 10, 0},
		{"valid page 2", "2", "10", 2, 10, 10},
		{"valid page 5 limit 20", "5", "20", 5, 20, 80},
		{"limit clamped to max", "1", "200", 1, 100, 0},
		{"page below 1 falls back", "0", "10", 1, 10, 0},
		{"negative page falls back", "-1", "10", 1, 10, 0},
		{"non-numeric falls back to default", "abc", "xyz", 1, 10, 0},
		{"too large page falls back", "9999999", "10", 1, 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, limit, offset := ValidatePaginationParams(tt.pageStr, tt.limitStr)
			if page != tt.wantPage {
				t.Errorf("page = %d, want %d", page, tt.wantPage)
			}
			if limit != tt.wantLimit {
				t.Errorf("limit = %d, want %d", limit, tt.wantLimit)
			}
			if offset != tt.wantOffset {
				t.Errorf("offset = %d, want %d", offset, tt.wantOffset)
			}
		})
	}
}

func TestValidateImageFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple filename", "photo.jpg", "photo.jpg"},
		{"with path traversal", "../../etc/passwd", "passwd"},
		{"with special chars", "my photo!@#.jpg", "myphoto.jpg"},
		{"empty", "", "image_file"},
		{"just dot", ".", "image_file"},
		{"unicode chars stripped", "café.jpg", "caf.jpg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateImageFilename(tt.input)
			if got != tt.want {
				t.Errorf("ValidateImageFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func BenchmarkValidateEmail(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ValidateEmail("user@example.com")
	}
}

func BenchmarkValidatePagination(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _ = ValidatePaginationParams("5", "20")
	}
}
