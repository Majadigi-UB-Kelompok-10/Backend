package utils

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9]+`)
	multipleHyphenRegex  = regexp.MustCompile(`-+`)
)

const (
	maxSlugBaseLength = 60
	defaultSlugBase   = "dokumen" 
)

func GenerateSlug(title string) string {
	slug := strings.ToLower(strings.TrimSpace(title))
	slug = nonAlphanumericRegex.ReplaceAllString(slug, "-")
	slug = multipleHyphenRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	if utf8.RuneCountInString(slug) > maxSlugBaseLength {
		runes := []rune(slug)
		slug = string(runes[:maxSlugBaseLength])
		slug = strings.Trim(slug, "-")
	}

	if slug == "" {
		slug = defaultSlugBase
	}

	bytes := make([]byte, 4)
	_, _ = rand.Read(bytes)
	return slug + "-" + hex.EncodeToString(bytes)
}