package utils

import (
	"testing"
)

func TestValidateQueryString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		fieldName string
		wantErr   bool
		expected  string
	}{
		{"Valid text", "Pajak Kendaraan", 50, "test", false, "Pajak Kendaraan"},
		{"Empty input", "", 50, "test", false, ""},
		{"Numeric only", "12345", 50, "test", false, "12345"},
		{"With hyphens", "B-1234-XY", 50, "test", false, "B-1234-XY"},
		
		// Invalid inputs
		{"Exceeds max length", "abcdefghijklmnopqrstuvwxyz", 10, "test", true, ""},
		{"Special chars only", "@#$%^&*()", 50, "test", true, ""},
		{"Mixed valid invalid", "Valid@#$Data", 50, "test", false, "ValidData"},
		{"SQL injection attempt", "'; DROP TABLE--", 50, "test", false, "DROP TABLE--"},
		
		// Edge cases - whitespace only returns empty string, not error
		{"Whitespace only", "     ", 50, "test", false, ""},
		{"Single char", "A", 50, "test", false, "A"},
		{"Exact max length", "12345", 5, "test", false, "12345"},
		{"One over max length", "123456", 5, "test", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateQueryString(tt.input, tt.maxLength, tt.fieldName)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateQueryString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If error, check error details
			if err != nil && err.Field != tt.fieldName {
				t.Errorf("ValidateQueryString() field = %v, want %v", err.Field, tt.fieldName)
			}

			// Check result
			if got != tt.expected {
				t.Errorf("ValidateQueryString() = %q, expected %q", got, tt.expected)
			}
		})
	}
}

func TestValidatePaginationParams(t *testing.T) {
	tests := []struct {
		name      string
		pageStr   string
		limitStr  string
		wantPage  int
		wantLimit int
	}{
		{"Default values", "", "", 1, 10},
		{"Valid page and limit", "5", "20", 5, 20},
		{"Zero page", "0", "10", 1, 10},
		{"Negative page", "-5", "10", 1, 10},
		{"Non-numeric page", "abc", "10", 1, 10},
		{"Zero limit", "1", "0", 1, 10},
		{"Negative limit", "1", "-5", 1, 10},
		{"Non-numeric limit", "1", "xyz", 1, 10},
		{"Both invalid", "bad", "bad", 1, 10},
		{"Very large values - capped at 50", "999", "999", 999, 50}, // limit is capped at 50
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, limit := ValidatePaginationParams(tt.pageStr, tt.limitStr)

			if page != tt.wantPage {
				t.Errorf("ValidatePaginationParams() page = %d, want %d", page, tt.wantPage)
			}
			if limit != tt.wantLimit {
				t.Errorf("ValidatePaginationParams() limit = %d, want %d", limit, tt.wantLimit)
			}
		})
	}
}

// Benchmark tests
func BenchmarkValidateQueryString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ValidateQueryString("Test Input Data", 100, "field")
	}
}

func BenchmarkValidatePaginationParams(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ValidatePaginationParams("5", "20")
	}
}
