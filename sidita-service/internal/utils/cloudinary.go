package utils

import (
	"path/filepath"
	"strings"
)

func ExtractPublicID(secureURL string) string {
	if secureURL == "" {
		return ""
	}

	parts := strings.Split(secureURL, "/upload/")
	if len(parts) < 2 {
		return "" 
	}

	path := parts[1]

	pathParts := strings.Split(path, "/")
	if len(pathParts) > 1 && strings.HasPrefix(pathParts[0], "v") {
		path = strings.Join(pathParts[1:], "/") 
	}

	ext := filepath.Ext(path)
	publicID := strings.TrimSuffix(path, ext)

	return publicID
}