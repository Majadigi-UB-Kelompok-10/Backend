package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/farildzaky/sinaker-service/internal/cache"
	"github.com/gofiber/fiber/v3"
)

const (
	wilayahBase = "https://ibnux.github.io/data-indonesia"
	wilayahTTL  = 30 * 24 * time.Hour
)

var (
	provIDPattern = regexp.MustCompile(`^\d{2}$`)
	kabIDPattern  = regexp.MustCompile(`^\d{4}$`)
	kecIDPattern  = regexp.MustCompile(`^\d{6,7}$`)
	wilayahClient = &http.Client{Timeout: 10 * time.Second}
)

func (h *SinakerHandler) fetchAndCacheWilayah(c fiber.Ctx, url, cacheKey string) error {
	if respondCached(c, cacheKey) {
		return nil
	}

	resp, err := wilayahClient.Get(url)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(ErrorResponse{
			Code: "ERR_WILAYAH", Message: "Gagal mengambil data wilayah dari sumber",
		})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Code: "ERR_WILAYAH_NOT_FOUND", Message: "Data wilayah tidak ditemukan",
		})
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code: "ERR_WILAYAH", Message: "Gagal membaca data wilayah",
		})
	}

	var raw json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code: "ERR_WILAYAH", Message: "Format data wilayah tidak valid",
		})
	}

	wrapped, _ := json.Marshal(SuccessResponse{Pesan: "Data Wilayah", Data: raw})
	cache.GlobalCache.Set(cacheKey, wrapped, wilayahTTL)

	c.Set("Content-Type", "application/json")
	return c.Send(wrapped)
}

func (h *SinakerHandler) GetWilayahProvinsi(c fiber.Ctx) error {
	url := fmt.Sprintf("%s/provinsi.json", wilayahBase)
	return h.fetchAndCacheWilayah(c, url, "sinaker:wilayah:provinsi")
}

func (h *SinakerHandler) GetWilayahKabKota(c fiber.Ctx) error {
	provID := c.Params("prov_id")
	if !provIDPattern.MatchString(provID) {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code: "ERR_VALIDATION", Message: "prov_id tidak valid, harus 2 digit angka (contoh: 35)",
		})
	}
	url := fmt.Sprintf("%s/kabupaten/%s.json", wilayahBase, provID)
	return h.fetchAndCacheWilayah(c, url, "sinaker:wilayah:kab:"+provID)
}

func (h *SinakerHandler) GetWilayahKecamatan(c fiber.Ctx) error {
	kabID := c.Params("kab_id")
	if !kabIDPattern.MatchString(kabID) {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code: "ERR_VALIDATION", Message: "kab_id tidak valid, harus 4 digit angka (contoh: 3578)",
		})
	}
	url := fmt.Sprintf("%s/kecamatan/%s.json", wilayahBase, kabID)
	return h.fetchAndCacheWilayah(c, url, "sinaker:wilayah:kec:"+kabID)
}

func (h *SinakerHandler) GetWilayahKelurahan(c fiber.Ctx) error {
	kecID := c.Params("kec_id")
	if !kecIDPattern.MatchString(kecID) {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code: "ERR_VALIDATION", Message: "kec_id tidak valid, harus 6-7 digit angka (contoh: 357801)",
		})
	}
	url := fmt.Sprintf("%s/kelurahan/%s.json", wilayahBase, kecID)
	return h.fetchAndCacheWilayah(c, url, "sinaker:wilayah:kel:"+kecID)
}
