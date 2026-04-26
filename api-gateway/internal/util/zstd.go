package util

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v3"
	"github.com/klauspost/compress/zstd"
)

// zstdEncoderPool maintains a pool of zstd encoders for efficient compression
var zstdEncoderPool = sync.Pool{
	New: func() interface{} {
		encoder, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedDefault))
		if err != nil {
			panic(err)
		}
		return encoder
	},
}

// isCompressible determines if a content type should be compressed
func isCompressible(contentType string) bool {
	compressibleTypes := []string{
		"application/json",
		"application/xml",
		"text/",
		"application/javascript",
		"application/x-www-form-urlencoded",
	}

	for _, ct := range compressibleTypes {
		if bytes.Contains([]byte(contentType), []byte(ct)) {
			return true
		}
	}

	return false
}

// compressData compresses data using zstd compression
func compressData(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	encoder := zstdEncoderPool.Get().(*zstd.Encoder)
	defer zstdEncoderPool.Put(encoder)

	var buf bytes.Buffer
	encoder.Reset(&buf)

	if _, err := encoder.Write(data); err != nil {
		return nil, err
	}

	if err := encoder.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ZstdCompressionMiddleware compresses response bodies using zstd
func ZstdCompressionMiddleware(c fiber.Ctx) error {
	// Check if client explicitly accepts zstd encoding
	acceptEncoding := string(c.Request().Header.Peek("Accept-Encoding"))
	if !bytes.Contains([]byte(acceptEncoding), []byte("zstd")) {
		return c.Next()
	}

	// Skip compression for certain content types
	contentType := string(c.Response().Header.ContentType())

	// Skip if already encoded
	if c.Response().Header.Peek("Content-Encoding") != nil {
		return c.Next()
	}

	// Process the request
	err := c.Next()
	if err != nil {
		return err
	}

	// Get response body
	body := c.Response().Body()

	// Only compress if:
	// 1. Body has content
	// 2. Response status is successful (< 400)
	// 3. Content is compressible (not binary streams, images, or already compressed)
	if len(body) > 0 && c.Response().StatusCode() < 400 && isCompressible(contentType) {
		compressed, compErr := compressData(body)
		if compErr != nil {
			// If compression fails, return uncompressed response
			return nil
		}

		// Only use compressed if it's actually smaller
		if len(compressed) < len(body) {
			c.Response().Header.Set("Content-Encoding", "zstd")
			c.Response().Header.Set("Content-Length", fmt.Sprintf("%d", len(compressed)))
			c.Response().SetBody(compressed)
		}
	}

	return nil
}
