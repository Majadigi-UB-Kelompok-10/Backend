package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"math"

	"github.com/farildzaky/bapenda-service/internal/cache"
	"github.com/farildzaky/bapenda-service/internal/db"
	"github.com/farildzaky/bapenda-service/internal/utils"
	"github.com/gofiber/fiber/v3"
)

// =============================================================================
// CEK PAJAK — public endpoint
// =============================================================================

func (h *BapendaHandler) GetInfoPajak(c fiber.Ctx) error {
	rawPlat := c.FormValue("plat_nomor")
	rawRangka := c.FormValue("nomor_rangka")

	// Validate + sanitize
	platNomor, ve := utils.ValidatePlatNomor(rawPlat)
	if ve != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION_FAILED",
			Message: "Input tidak valid",
			Action:  "Harap perbaiki data input",
			Errors:  []FieldError{{Field: ve.Field, Message: ve.Message}},
		})
	}

	nomorRangka, ve := utils.ValidateNomorRangka(rawRangka)
	if ve != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION_FAILED",
			Message: "Input tidak valid",
			Action:  "Harap perbaiki data input",
			Errors:  []FieldError{{Field: ve.Field, Message: ve.Message}},
		})
	}

	slog.Info("cek_pajak.request", slog.String("plat_nomor", platNomor))

	cacheKey := fmt.Sprintf("pajak:info:%s:%s", platNomor, nomorRangka)
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetKendaraanByPlatDanRangka(ctx, db.GetKendaraanByPlatDanRangkaParams{
		PlatNomor:       platNomor,
		LimaDigitRangka: nomorRangka,
	})
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Code:    "ERR_NOT_FOUND",
			Message: "Data kendaraan tidak ditemukan",
			Action:  "Pastikan plat nomor dan rangka sesuai STNK",
		})
	}

	res := SuccessResponse{
		Pesan: "Data ditemukan",
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
			Estimasi5Tahunan: Estimasi5Tahunan{
				CetakStnk: data.CetakStnk,
				CetakTnkb: data.CetakTnkb,
			},
		},
	}

	return cacheJSON(c, cacheKey, CacheTTLDetail, res)
}


// =============================================================================
// CASCADING DROPDOWN — kalkulator NJKB
// =============================================================================

func (h *BapendaHandler) GetDropdownJenis(c fiber.Ctx) error {
	cacheKey := "dropdown:jenis"
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetDistinctJenis(ctx)
	if err != nil {
		slog.Error("dropdown.jenis.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_DB_QUERY",
			Message: "Gagal memuat jenis kendaraan",
		})
	}

	return cacheJSON(c, cacheKey, CacheTTLStatic, SuccessResponse{Pesan: "Sukses", Data: data})
}

func (h *BapendaHandler) GetDropdownMerk(c fiber.Ctx) error {
	jenis := c.Query("jenis")
	if err := requireFields(c, map[string]string{"jenis": jenis}); err != nil {
		return err
	}

	cacheKey := "dropdown:merk:" + normalizeKey(jenis)
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetDistinctMerk(ctx, jenis)
	if err != nil {
		slog.Error("dropdown.merk.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_DB_QUERY",
			Message: "Gagal memuat merk kendaraan",
		})
	}

	return cacheJSON(c, cacheKey, CacheTTLMaps, SuccessResponse{Pesan: "Sukses", Data: data})
}

func (h *BapendaHandler) GetDropdownModel(c fiber.Ctx) error {
	jenis := c.Query("jenis")
	merk := c.Query("merk")
	if err := requireFields(c, map[string]string{"jenis": jenis, "merk": merk}); err != nil {
		return err
	}

	cacheKey := "dropdown:model:" + normalizeKey(jenis, merk)
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetDistinctModel(ctx, db.GetDistinctModelParams{
		JenisKendaraan: jenis,
		Merk:           merk,
	})
	if err != nil {
		slog.Error("dropdown.model.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_DB_QUERY",
			Message: "Gagal memuat model",
		})
	}

	return cacheJSON(c, cacheKey, CacheTTLMaps, SuccessResponse{Pesan: "Sukses", Data: data})
}

func (h *BapendaHandler) GetDropdownTipe(c fiber.Ctx) error {
	jenis := c.Query("jenis")
	merk := c.Query("merk")
	model := c.Query("model")
	if err := requireFields(c, map[string]string{"jenis": jenis, "merk": merk, "model": model}); err != nil {
		return err
	}

	cacheKey := "dropdown:tipe:" + normalizeKey(jenis, merk, model)
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetDistinctTipe(ctx, db.GetDistinctTipeParams{
		JenisKendaraan: jenis,
		Merk:           merk,
		Model:          model,
	})
	if err != nil {
		slog.Error("dropdown.tipe.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_DB_QUERY",
			Message: "Gagal memuat tipe",
		})
	}

	return cacheJSON(c, cacheKey, CacheTTLMaps, SuccessResponse{Pesan: "Sukses", Data: data})
}

func (h *BapendaHandler) GetDropdownTahun(c fiber.Ctx) error {
	jenis := c.Query("jenis")
	merk := c.Query("merk")
	model := c.Query("model")
	tipe := c.Query("tipe")
	if err := requireFields(c, map[string]string{
		"jenis": jenis, "merk": merk, "model": model, "tipe": tipe,
	}); err != nil {
		return err
	}

	cacheKey := "dropdown:tahun:" + normalizeKey(jenis, merk, model, tipe)
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetDistinctTahun(ctx, db.GetDistinctTahunParams{
		JenisKendaraan: jenis,
		Merk:           merk,
		Model:          model,
		Tipe:           tipe,
	})
	if err != nil {
		slog.Error("dropdown.tahun.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_DB_QUERY",
			Message: "Gagal memuat tahun",
		})
	}

	return cacheJSON(c, cacheKey, CacheTTLMaps, SuccessResponse{Pesan: "Sukses", Data: data})
}

// =============================================================================
// HITUNG ESTIMASI PAJAK — kalkulator NJKB
// =============================================================================

func (h *BapendaHandler) HitungKalkulasiNJKB(c fiber.Ctx) error {
	var req HitungNjkbRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_INVALID_JSON",
			Message: "Format permintaan tidak valid",
		})
	}

	if err := requireFields(c, map[string]string{
		"jenis": req.Jenis, "merk": req.Merk, "model": req.Model, "tipe": req.Tipe,
	}); err != nil {
		return err
	}

	cacheKey := "kalkulasi:" + normalizeKey(req.Jenis, req.Merk, req.Model, req.Tipe) +
		fmt.Sprintf(":%d", req.Tahun)
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	njkbData, err := h.Queries.GetNilaiJual(ctx, db.GetNilaiJualParams{
		Jenis: req.Jenis, Merk: req.Merk, Model: req.Model,
		Tipe: req.Tipe, Tahun: req.Tahun,
	})
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Code:    "ERR_NOT_FOUND",
			Message: "Data master NJKB tidak ditemukan",
		})
	}

	tarifList, err := h.Queries.GetAllTarifPKB(ctx)
	if err != nil {
		slog.Error("kalkulasi.tarif.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_DB_QUERY",
			Message: "Gagal memuat tarif PKB",
		})
	}

	estimasi := calculateEstimasi(njkbData.NilaiJual, tarifList)

	res := SuccessResponse{
		Pesan: "Sukses",
		Data: KalkulasiNJKBResponse{
			Njkb:     njkbData.NilaiJual,
			Estimasi: estimasi,
		},
	}
	return cacheJSON(c, cacheKey, CacheTTLDetail, res)
}

func calculateEstimasi(nilaiJual int64, tarifList []db.GetAllTarifPKBRow) []EstimasiTarif {
	estimasi := make([]EstimasiTarif, 0, len(tarifList))
	nilaiFloat := float64(nilaiJual)

	for _, t := range tarifList {
		tpkb, _ := t.TarifPkbPersen.Float64Value()
		tops, _ := t.OpsenPkbPersen.Float64Value()
		pkbPokok := nilaiFloat * tpkb.Float64

		estimasi = append(estimasi, EstimasiTarif{
			JenisPlat: t.JenisPlat,
			Label:     t.Label,
			Pkb:       int64(math.Round(pkbPokok)),
			Opsen:     int64(math.Round(pkbPokok * tops.Float64)),
		})
	}
	return estimasi
}

var _ = cache.GlobalCache