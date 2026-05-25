package utils

import (
	"regexp"
	"strings"
	"testing"
)

var (
	transSuffixFormat = regexp.MustCompile(`-[a-f0-9]{8}$`)
	transSlugFormat   = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
)

func TestGenerateSlug_OutputFormat(t *testing.T) {
	tests := []string{
		"Jadwal Bus Trans Jatim 2026",
		"Rute: Surabaya - Malang!",
		"title with    multiple   spaces",
		"SEMUA HURUF KAPITAL",
		"emoji 🚌 trans jatim",
	}
	for _, title := range tests {
		t.Run(title, func(t *testing.T) {
			slug := GenerateSlug(title)
			if !transSlugFormat.MatchString(slug) {
				t.Errorf("GenerateSlug(%q) = %q does not match slug format", title, slug)
			}
			if len(slug) > 70 {
				t.Errorf("GenerateSlug(%q) = %q too long (%d chars)", title, slug, len(slug))
			}
			if !transSuffixFormat.MatchString(slug) {
				t.Errorf("GenerateSlug(%q) = %q missing 8-char hex suffix", title, slug)
			}
		})
	}
}

func TestGenerateSlug_FallbackForEmpty(t *testing.T) {
	for _, title := range []string{"", "   ", "!!!", "🚌🚌"} {
		t.Run(title, func(t *testing.T) {
			slug := GenerateSlug(title)
			if !strings.HasPrefix(slug, "trans-jatim-") {
				t.Errorf("GenerateSlug(%q) = %q should start with 'trans-jatim-'", title, slug)
			}
		})
	}
}

func TestGenerateSlug_HasRandomSuffix(t *testing.T) {
	prev := GenerateSlug("Same Title")
	for i := 0; i < 10; i++ {
		if cur := GenerateSlug("Same Title"); cur != prev {
			return
		}
	}
	t.Error("GenerateSlug returned same slug 10 times — random suffix not working")
}

func TestGenerateSlug_NoLeadingHyphen(t *testing.T) {
	for _, title := range []string{"---weird---", "-leading", "trailing-", "   "} {
		t.Run(title, func(t *testing.T) {
			if slug := GenerateSlug(title); strings.HasPrefix(slug, "-") {
				t.Errorf("GenerateSlug(%q) = %q starts with hyphen", title, slug)
			}
		})
	}
}

func TestGenerateSlug_TruncatesLongTitle(t *testing.T) {
	longTitle := "Ini adalah judul yang sangat panjang sekali melebihi enam puluh karakter batas maksimum yang diizinkan oleh sistem"
	slug := GenerateSlug(longTitle)
	if len(slug) > 70 {
		t.Errorf("GenerateSlug long title produced slug of length %d (max 70): %q", len(slug), slug)
	}
}

func BenchmarkGenerateSlug(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateSlug("Sample Title for Benchmark")
	}
}
