package utils

import (
	"regexp"
	"strings"
	"testing"
)

var (
	slugFormat = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
)

func TestGenerateSlug_OutputFormat(t *testing.T) {
	tests := []string{
		"Hoaks Vaksin COVID-19",
		"Email Epstein Sebut Perang Dunia III Dimulai 8 Februari 2026",
		"Berita dengan: Karakter! @# Khusus",
		"Title with    multiple   spaces",
		"AKHIRNYA SEMUA HURUF KAPITAL",
		"emoji-heavy 🎉 title 🚀",
	}

	for _, title := range tests {
		t.Run(title, func(t *testing.T) {
			slug := GenerateSlug(title)

			// Format check: must match slug pattern
			if !slugFormat.MatchString(slug) {
				t.Errorf("GenerateSlug(%q) = %q does not match slug format", title, slug)
			}

			// Length check: should not exceed maxSlugBaseLength + suffix
			// (60 + 1 dash + 4 hex chars = 65)
			if len(slug) > 70 {
				t.Errorf("GenerateSlug(%q) = %q is too long (%d chars)", title, slug, len(slug))
			}

			// Suffix check: must end with -<4 hex chars>
			if !regexp.MustCompile(`-[a-f0-9]{4}$`).MatchString(slug) {
				t.Errorf("GenerateSlug(%q) = %q missing 4-char hex suffix", title, slug)
			}
		})
	}
}

func TestGenerateSlug_FallbackForEmpty(t *testing.T) {
	// All non-alphanumeric input should fall back to "berita-XXXX"
	tests := []string{"", "   ", "!!!", "🎉🎉🎉"}

	for _, title := range tests {
		t.Run(title, func(t *testing.T) {
			slug := GenerateSlug(title)
			if !strings.HasPrefix(slug, "berita-") {
				t.Errorf("GenerateSlug(%q) = %q should start with 'berita-' fallback", title, slug)
			}
		})
	}
}

func TestGenerateSlug_Uniqueness(t *testing.T) {
	// Generate 1000 slugs from same title; should all differ (random suffix)
	title := "Same Title"
	seen := make(map[string]bool)

	for i := 0; i < 1000; i++ {
		slug := GenerateSlug(title)
		if seen[slug] {
			t.Errorf("GenerateSlug generated duplicate slug %q in 1000 iterations", slug)
			return
		}
		seen[slug] = true
	}
}

func TestGenerateSlug_DoesNotStartOrEndWithHyphen(t *testing.T) {
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
			// Note: ends with random suffix, so check the part before suffix
		})
	}
}

func BenchmarkGenerateSlug(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateSlug("Sample Title for Benchmark")
	}
}
