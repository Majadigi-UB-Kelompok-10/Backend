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

func GenerateSlug(title string) string {
	slug := strings.ToLower(strings.TrimSpace(title))
	slug = nonAlphanumericRegex.ReplaceAllString(slug, "-")
	slug = multipleHyphenRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	if utf8.RuneCountInString(slug) > 60 {
		runes := []rune(slug)
		slug = string(runes[:60])
		slug = strings.Trim(slug, "-")
	}

	if slug == "" {
		slug = "berita"
	}

	bytes := make([]byte, 2)
	rand.Read(bytes)
	randomStr := hex.EncodeToString(bytes)

	return slug + "-" + randomStr
}