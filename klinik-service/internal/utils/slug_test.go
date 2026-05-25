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

			if !slugFormat.MatchString(slug) {
				t.Errorf("GenerateSlug(%q) = %q does not match slug format", title, slug)
			}

			if len(slug) > 70 {
				t.Errorf("GenerateSlug(%q) = %q is too long (%d chars)", title, slug, len(slug))
			}

			if !suffixFormat.MatchString(slug) {
				t.Errorf("GenerateSlug(%q) = %q missing 8-char hex suffix", title, slug)
			}
		})
	}
}

func TestGenerateSlug_FallbackForEmpty(t *testing.T) {
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

func TestGenerateSlug_HasRandomSuffix(t *testing.T) {
	
	title := "Same Title"

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

func BenchmarkGenerateSlug(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateSlug("Sample Title for Benchmark")
	}
}

func TestGenerateSlug_TruncatesLongTitle(t *testing.T) {
	longTitle := "Ini adalah judul yang sangat panjang sekali melebihi enam puluh karakter batas maksimum yang diizinkan oleh sistem"
	slug := GenerateSlug(longTitle)
	if len(slug) > 70 {
		t.Errorf("GenerateSlug long title produced slug of length %d (max 70): %q", len(slug), slug)
	}
}