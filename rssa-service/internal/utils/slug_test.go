package utils

import (
	"regexp"
	"strings"
	"testing"
)

var (
	rssaSuffixFormat = regexp.MustCompile(`-[a-f0-9]{8}$`)
	rssaSlugFormat   = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
)

func TestGenerateSlug_OutputFormat(t *testing.T) {
	tests := []string{
		"Layanan RSSA Malang",
		"Poli Jantung: Jadwal & Info!",
		"title with    multiple   spaces",
		"SEMUA HURUF KAPITAL",
		"emoji 🏥 rumah sakit",
	}
	for _, title := range tests {
		t.Run(title, func(t *testing.T) {
			slug := GenerateSlug(title)
			if !rssaSlugFormat.MatchString(slug) {
				t.Errorf("GenerateSlug(%q) = %q does not match slug format", title, slug)
			}
			if len(slug) > 70 {
				t.Errorf("GenerateSlug(%q) = %q too long (%d chars)", title, slug, len(slug))
			}
			if !rssaSuffixFormat.MatchString(slug) {
				t.Errorf("GenerateSlug(%q) = %q missing 8-char hex suffix", title, slug)
			}
		})
	}
}

func TestGenerateSlug_FallbackForEmpty(t *testing.T) {
	for _, title := range []string{"", "   ", "!!!", "🏥🏥"} {
		t.Run(title, func(t *testing.T) {
			slug := GenerateSlug(title)
			if !strings.HasPrefix(slug, "item-") {
				t.Errorf("GenerateSlug(%q) = %q should start with 'item-'", title, slug)
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
