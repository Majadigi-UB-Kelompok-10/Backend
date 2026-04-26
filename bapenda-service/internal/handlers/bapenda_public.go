package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strings"

	"github.com/farildzaky/bapenda-service/internal/cache"
	"github.com/farildzaky/bapenda-service/internal/db"
	"github.com/gofiber/fiber/v3"
)

func (h *BapendaHandler) GetInfoPajak(c fiber.Ctx) error {
	platNomor := strings.ToUpper(strings.ReplaceAll(c.FormValue("plat_nomor"), " ", ""))
	nomorRangka := strings.ToUpper(strings.TrimSpace(c.FormValue("nomor_rangka")))

	slog.Info("Menerima request GetInfoPajak", slog.String("plat_nomor", platNomor))

	if platNomor == "" || len(nomorRangka) != 5 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION_FAILED",
			Message: "Input tidak valid",
			Action:  "Harap perbaiki data input",
			Errors: []FieldError{
				{Field: "plat_nomor", Message: "Plat nomor wajib diisi"},
				{Field: "nomor_rangka", Message: "Nomor rangka harus 5 digit terakhir"},
			},
		})
	}

	cacheKey := fmt.Sprintf("pajak:info:%s:%s", platNomor, nomorRangka)
	if cachedBytes, ok := cache.GlobalCache.Get(cacheKey); ok {
		c.Set("Content-Type", "application/json")
		return c.Send(cachedBytes)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	arg := db.GetKendaraanByPlatDanRangkaParams{PlatNomor: platNomor, LimaDigitRangka: nomorRangka}
	data, err := h.Queries.GetKendaraanByPlatDanRangka(ctx, arg)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Code:    "ERR_NOT_FOUND",
			Message: "Data Kendaraan Tidak Ditemukan",
			Action:  "Pastikan plat nomor dan rangka sesuai STNK",
		})
	}

	res := SuccessResponse{
		Pesan: "Data Ditemukan",
		Data: InfoPajakResponse{
			Identitas: IdentitasKendaraan{
				PlatNomor:   data.PlatNomorDisplay,
				Merk:        data.Merk,
				Tipe:        data.Tipe,
				Warna:       data.Warna,
				TahunBuat:   data.TahunBuat,
				MasaPajak:   data.MasaPajak.Time.Format("2006-01-02"),
				StatusAktif: data.StatusAktif,
			},
			RincianBiaya: RincianBiayaPajak{
				PkbPokok:           data.PkbPokok,
				OpsenPkb:           data.OpsenPkb,
				Swdkllj:            data.Swdkllj,
				ParkirBerlangganan: data.ParkirBerlangganan,
				TotalPajak:         int64(data.TotalPajakTahunan.Int32),
			},
			Estimasi5Tahunan: Estimasi5Tahunan{CetakStnk: data.CetakStnk, CetakTnkb: data.CetakTnkb},
		},
	}
	cache.GlobalCache.Set(cacheKey, res)
	return c.JSON(res)
}

func (h *BapendaHandler) GetDropdownJenis(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()
	data, err := h.Queries.GetDistinctJenis(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_DB", Message: "Gagal memuat jenis kendaraan"})
	}
	return c.JSON(SuccessResponse{Pesan: "Sukses", Data: data})
}

func (h *BapendaHandler) GetDropdownMerk(c fiber.Ctx) error {
	jenis := c.Query("jenis")
	cacheKey := "dropdown:merk:" + jenis

	// 1. Cek Cache Dulu
	if cachedBytes, ok := cache.GlobalCache.Get(cacheKey); ok {
		c.Set("Content-Type", "application/json")
		c.Set("X-Cache", "HIT")
		return c.Send(cachedBytes)
	}

	// 2. Kalau miss, baru tembak DB
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetDistinctMerk(ctx, jenis)
	if err != nil {
		slog.Error("Gagal query DropdownMerk", slog.String("error", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_DB_QUERY",
			Message: "Gagal memuat merk kendaraan",
		})
	}

	// 3. Simpan ke Cache
	res := SuccessResponse{Pesan: "Sukses", Data: data}
	cache.GlobalCache.Set(cacheKey, res)
	return c.JSON(res)
}

func (h *BapendaHandler) GetDropdownModel(c fiber.Ctx) error {
	jenis := c.Query("jenis")
	merk := c.Query("merk")
	// Normalisasi key agar aman dari spasi
	cacheKey := fmt.Sprintf("dropdown:model:%s:%s", jenis, merk)

	if cachedBytes, ok := cache.GlobalCache.Get(cacheKey); ok {
		c.Set("Content-Type", "application/json")
		c.Set("X-Cache", "HIT")
		return c.Send(cachedBytes)
	}

	arg := db.GetDistinctModelParams{JenisKendaraan: jenis, Merk: merk}
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetDistinctModel(ctx, arg)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_DB_QUERY", Message: "Gagal memuat model"})
	}

	res := SuccessResponse{Pesan: "Sukses", Data: data}
	cache.GlobalCache.Set(cacheKey, res)
	return c.JSON(res)
}

func (h *BapendaHandler) GetDropdownTipe(c fiber.Ctx) error {
	jenis := c.Query("jenis")
	merk := c.Query("merk")
	model := c.Query("model")
	cacheKey := fmt.Sprintf("dropdown:tipe:%s:%s:%s", jenis, merk, model)

	if cachedBytes, ok := cache.GlobalCache.Get(cacheKey); ok {
		c.Set("Content-Type", "application/json")
		c.Set("X-Cache", "HIT")
		return c.Send(cachedBytes)
	}

	arg := db.GetDistinctTipeParams{JenisKendaraan: jenis, Merk: merk, Model: model}
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetDistinctTipe(ctx, arg)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_DB_QUERY", Message: "Gagal memuat tipe"})
	}

	res := SuccessResponse{Pesan: "Sukses", Data: data}
	cache.GlobalCache.Set(cacheKey, res)
	return c.JSON(res)
}

func (h *BapendaHandler) GetDropdownTahun(c fiber.Ctx) error {
	jenis := c.Query("jenis")
	merk := c.Query("merk")
	model := c.Query("model")
	tipe := c.Query("tipe")
	cacheKey := fmt.Sprintf("dropdown:tahun:%s:%s:%s:%s", strings.ReplaceAll(jenis, " ", ""), strings.ReplaceAll(merk, " ", ""), strings.ReplaceAll(model, " ", ""), strings.ReplaceAll(tipe, " ", ""))

	if cachedBytes, ok := cache.GlobalCache.Get(cacheKey); ok {
		c.Set("Content-Type", "application/json")
		c.Set("X-Cache", "HIT")
		return c.Send(cachedBytes)
	}

	arg := db.GetDistinctTahunParams{JenisKendaraan: jenis, Merk: merk, Model: model, Tipe: tipe}
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetDistinctTahun(ctx, arg)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_DB_QUERY", Message: "Gagal memuat tahun"})
	}

	res := SuccessResponse{Pesan: "Sukses", Data: data}
	cache.GlobalCache.Set(cacheKey, res)
	return c.JSON(res)
}

func (h *BapendaHandler) HitungKalkulasiNJKB(c fiber.Ctx) error {
	var req HitungNjkbRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_INVALID_JSON", Message: "Format permintaan tidak valid"})
	}

	cacheKey := fmt.Sprintf("kalkulasi:%s:%s:%s:%s:%d", req.Jenis, req.Merk, req.Model, req.Tipe, req.Tahun)
	if cached, ok := cache.GlobalCache.Get(cacheKey); ok {
		c.Set("Content-Type", "application/json")
		return c.Send(cached)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	argNjkb := db.GetNilaiJualParams{Jenis: req.Jenis, Merk: req.Merk, Model: req.Model, Tipe: req.Tipe, Tahun: req.Tahun}
	njkbData, err := h.Queries.GetNilaiJual(ctx, argNjkb)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Code: "ERR_NOT_FOUND", Message: "Data Master NJKB tidak ditemukan"})
	}

	tarifList, _ := h.Queries.GetAllTarifPKB(ctx)
	var estimasi []EstimasiTarif
	nilaiJual := float64(njkbData.NilaiJual)

	for _, t := range tarifList {
		tpkb, _ := t.TarifPkbPersen.Float64Value()
		tops, _ := t.OpsenPkbPersen.Float64Value()
		pkbPokok := nilaiJual * tpkb.Float64
		estimasi = append(estimasi, EstimasiTarif{
			JenisPlat: t.JenisPlat, Label: t.Label,
			Pkb: int64(math.Round(pkbPokok)), Opsen: int64(math.Round(pkbPokok * tops.Float64)),
		})
	}

	res := SuccessResponse{Pesan: "Sukses", Data: KalkulasiNJKBResponse{Njkb: njkbData.NilaiJual, Estimasi: estimasi}}
	cache.GlobalCache.Set(cacheKey, res)
	return c.JSON(res)
}