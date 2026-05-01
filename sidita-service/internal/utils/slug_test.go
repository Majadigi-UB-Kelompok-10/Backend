package utils

import (
	"regexp"
	"strings"
	"testing"
)

var (
	slugFormat   = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	suffixFormat = regexp.MustCompile(`-[a-f0-9]{8}$`)
)

// =============================================================================
// Format compliance — hanya huruf kecil, angka, dan dash. --- IGNORE --- 
// =============================================================================

func TestGenerateSlug_OutputFormat(t *testing.T) {
	tests := []string{
		"Pantai Balekambang Indah",
		"Coban Rondo Waterfall",
		"Hotel Bintang Lima Malang",
		"Title with    multiple   spaces",
		"AKHIRNYA SEMUA HURUF KAPITAL",
		"emoji-heavy 🎉 title 🚀",
		"Karakter: spesial! @# disini",
	}

	for _, title := range tests {
		t.Run(title, func(t *testing.T) {
			slug := GenerateSlug(title)

			if !slugFormat.MatchString(slug) {
				t.Errorf("GenerateSlug(%q) = %q does not match slug format", title, slug)
			}

			// Length check: max 60 base + 1 dash + 8 hex = 69 chars
			if len(slug) > 70 {
				t.Errorf("GenerateSlug(%q) = %q is too long (%d chars)", title, slug, len(slug))
			}

			if !suffixFormat.MatchString(slug) {
				t.Errorf("GenerateSlug(%q) = %q missing 8-char hex suffix", title, slug)
			}
		})
	}
}

// =============================================================================
// Fallback for empty/junk input — sidita pakai prefix "item-"
// =============================================================================

func TestGenerateSlug_FallbackForEmpty(t *testing.T) {
	tests := []string{"", "   ", "!!!", "🎉🎉🎉"}

	for _, title := range tests {
		t.Run(title, func(t *testing.T) {
			slug := GenerateSlug(title)
			if !strings.HasPrefix(slug, "item-") {
				t.Errorf("GenerateSlug(%q) = %q should start with 'item-' fallback", title, slug)
			}
		})
	}
}

// =============================================================================
// Random suffix 
// =============================================================================

func TestGenerateSlug_HasRandomSuffix(t *testing.T) {
	title := "Pantai Balekambang"

	differs := false
	prev := GenerateSlug(title)
	for i := 0; i < 10; i++ {
		current := GenerateSlug(title)
		if current != prev {
			differs = true
			break
		}
		prev = current
	}

	if !differs {
		t.Errorf("GenerateSlug returned same slug 10 times in a row — random suffix not working")
	}
}

// =============================================================================
// No leading/trailing hyphens
// =============================================================================

func TestGenerateSlug_DoesNotStartWithHyphen(t *testing.T) {
	tests := []string{
		"---weird---title---",
		"-leading hyphen",
		"trailing hyphen-",
		"   spaces   ",
	}

	for _, title := range tests {
		t.Run(title, func(t *testing.T) {
			slug := GenerateSlug(title)
			if strings.HasPrefix(slug, "-") {
				t.Errorf("GenerateSlug(%q) = %q starts with hyphen", title, slug)
			}
		})
	}
}

// =============================================================================
// Max length truncation
// =============================================================================

func TestGenerateSlug_TruncatesLongTitle(t *testing.T) {
	veryLongTitle := strings.Repeat("a", 200)
	slug := GenerateSlug(veryLongTitle)

	if len(slug) > 70 {
		t.Errorf("GenerateSlug for very long title produced slug > 70 chars: len=%d, slug=%q", len(slug), slug)
	}

	if !suffixFormat.MatchString(slug) {
		t.Errorf("GenerateSlug truncated slug missing hex suffix: %q", slug)
	}
}



func BenchmarkGenerateSlug(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateSlug("Pantai Balekambang Malang Selatan")
	}
}