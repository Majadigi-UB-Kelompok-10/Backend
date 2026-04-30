package utils

import (
	"testing"
)

// =============================================================================
// SanitizeQueryString
// =============================================================================

func TestSanitizeQueryString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		wantOk    bool
		wantClean string 
	}{
		{"empty input", "", 100, true, ""},
		{"normal text", "Toyota Avanza", 100, true, "Toyota Avanza"},
		{"with allowed chars", "AB-123_test.foo", 100, true, "AB-123_test.foo"},
		{"with slash", "model/year", 100, true, "model/year"},
		{"with dangerous chars", "test<script>", 100, true, "testscript"},
		{"with quotes", `test"value'`, 100, true, "testvalue"},
		{"with semicolon", "test;DROP TABLE", 100, true, "testDROP TABLE"},
		{"trim whitespace", "  spaced  ", 100, true, "spaced"},

		{"too long", "abcdefghij", 5, false, ""},
		{"only dangerous chars", "<>&'\"", 100, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SanitizeQueryString(tt.input, tt.maxLength, "test_field")

			if tt.wantOk {
				if err != nil {
					t.Errorf("SanitizeQueryString(%q) unexpected error: %v", tt.input, err)
					return
				}
				if result != tt.wantClean {
					t.Errorf("SanitizeQueryString(%q) = %q, want %q", tt.input, result, tt.wantClean)
				}
				return
			}

			if err == nil {
				t.Errorf("SanitizeQueryString(%q) expected error, got nil", tt.input)
			}
		})
	}
}

// =============================================================================
// ParsePagination
// =============================================================================

func TestParsePagination(t *testing.T) {
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
		{"limit clamped to MaxLimit", "1", "200", 1, 100, 0},
		{"page below 1 falls back", "0", "10", 1, 10, 0},
		{"negative page falls back", "-1", "10", 1, 10, 0},
		{"non-numeric falls back", "abc", "xyz", 1, 10, 0},
		{"too many digits in page", "9999999", "10", 1, 10, 0}, // > maxPageDigits (6)
		{"too many digits in limit", "1", "9999", 1, 10, 0},    // > maxLimitDigits (3)
		{"limit exactly at max", "1", "100", 1, 100, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, limit, offset := ParsePagination(tt.pageStr, tt.limitStr)
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

// =============================================================================
// ValidatePlatNomor — Indonesia plate format
// =============================================================================

func TestValidatePlatNomor(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOk  bool
		wantOut string 
	}{
		{"normal plat", "AB1234CD", true, "AB1234CD"},
		{"with spaces", "AB 1234 CD", true, "AB1234CD"},
		{"lowercase", "ab1234cd", true, "AB1234CD"},
		{"mixed case", "Ab1234Cd", true, "AB1234CD"},
		{"min length (3)", "B12", true, "B12"},
		{"max length (15)", "ABCDEFGHIJ12345", true, "ABCDEFGHIJ12345"},

		{"empty", "", false, ""},
		{"only spaces", "   ", false, ""},
		{"too short (2)", "B1", false, ""},
		{"too long (16)", "ABCDEFGHIJK123456", false, ""},
		{"with special chars", "B1234-CD", false, ""},
		{"with symbols", "B1234@CD", false, ""},
		{"with unicode", "B1234日本", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidatePlatNomor(tt.input)

			if tt.wantOk {
				if err != nil {
					t.Errorf("ValidatePlatNomor(%q) unexpected error: %v", tt.input, err)
					return
				}
				if result != tt.wantOut {
					t.Errorf("ValidatePlatNomor(%q) = %q, want %q", tt.input, result, tt.wantOut)
				}
				return
			}

			if err == nil {
				t.Errorf("ValidatePlatNomor(%q) expected error, got nil (result: %q)", tt.input, result)
			}
		})
	}
}

// =============================================================================
// ValidateNomorRangka — must be exactly 5 alphanumeric chars
// =============================================================================

func TestValidateNomorRangka(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOk  bool
		wantOut string
	}{
		{"5 digits", "12345", true, "12345"},
		{"5 letters", "ABCDE", true, "ABCDE"},
		{"mixed letters/digits", "AB123", true, "AB123"},
		{"lowercase becomes uppercase", "ab123", true, "AB123"},
		{"with leading/trailing spaces", "  AB123  ", true, "AB123"},

		{"empty", "", false, ""},
		{"4 chars (too short)", "AB12", false, ""},
		{"6 chars (too long)", "AB1234", false, ""},
		{"with special char", "AB-12", false, ""},
		{"with space inside", "A 123", false, ""},
		{"with dot", "AB.12", false, ""},
		{"unicode", "AB12日", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateNomorRangka(tt.input)

			if tt.wantOk {
				if err != nil {
					t.Errorf("ValidateNomorRangka(%q) unexpected error: %v", tt.input, err)
					return
				}
				if result != tt.wantOut {
					t.Errorf("ValidateNomorRangka(%q) = %q, want %q", tt.input, result, tt.wantOut)
				}
				return
			}

			if err == nil {
				t.Errorf("ValidateNomorRangka(%q) expected error, got nil", tt.input)
			}
		})
	}
}

// =============================================================================
// SanitizeFilename — defensive filename cleaning
// =============================================================================

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"normal filename", "photo.jpg", "photo.jpg"},
		{"with path traversal '..'", "../../etc/passwd", "etcpasswd"},
		{"with forward slash", "subdir/file.png", "subdirfile.png"},
		{"with backslash", "subdir\\file.png", "subdirfile.png"},
		{"with special chars", "my photo!@#.jpg", "myphoto.jpg"},
		{"with spaces", "my file.png", "myfile.png"},
		{"empty input", "", "image"},
		{"only special chars", "!@#$%^&*", "image"},
		{"unicode chars", "café.jpg", "caf.jpg"},
		{"multiple dots", "..hidden..file..", "hiddenfile"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeFilename(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// =============================================================================
// ValidationError implements error interface
// =============================================================================

func TestValidationError_Error(t *testing.T) {
	ve := &ValidationError{
		Field:   "plat_nomor",
		Message: "format invalid",
	}

	expected := "plat_nomor: format invalid"
	if got := ve.Error(); got != expected {
		t.Errorf("Error() = %q, want %q", got, expected)
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkValidatePlatNomor(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ValidatePlatNomor("AB1234CD")
	}
}

func BenchmarkSanitizeQueryString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = SanitizeQueryString("Toyota Avanza 2020", 100, "search")
	}
}

func BenchmarkParsePagination(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _ = ParsePagination("5", "20")
	}
}