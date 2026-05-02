package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/farildzaky/siskaperbapo-service/internal/db"
	"github.com/farildzaky/siskaperbapo-service/internal/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/sync/errgroup"
)

// =====================================================================
// PUBLIC: GET ALL BAHAN POKOK
// =====================================================================
func (h *SiskaperbapoHandler) GetAllBahanPokok(c fiber.Ctx) error {
	page, limit, offset := parsePagination(c)
	hariIni := time.Now().Format("2006-01-02")
	tanggalStr := c.Query("tanggal", hariIni)

	bahanPokok, _ := utils.ValidateQueryString(c.Query("bahan_pokok"), 100, "bahan_pokok")
	areaInput, _ := utils.ValidateQueryString(c.Query("area"), 100, "area")
	areaSlugPilihan := ""
	if areaInput != "" {
		areaSlugPilihan = generateSlug(areaInput)
	}

	cacheKey := fmt.Sprintf("all_bahan:%s:page_%d:limit_%d:bahan_%s:area_%s",
		tanggalStr, page, limit, normalizeKey(bahanPokok), areaSlugPilihan)

	if respondCached(c, cacheKey) {
		return nil
	}

	parsedTime, err := time.Parse("2006-01-02", tanggalStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Format tanggal salah (gunakan YYYY-MM-DD)",
		})
	}
	tanggalPg := pgtype.Date{Time: parsedTime, Valid: true}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)
	var bahanData []db.BahanPokok
	var totalCount int64
	var riwayatData []db.GetTrenSemuaBahanPokokRow

	g.Go(func() error {
		var err error
		bahanData, err = h.Queries.GetAllBahanPokok(gCtx, db.GetAllBahanPokokParams{
			LimitData:  int32(limit),
			OffsetData: int32(offset),
			Keyword:    pgtype.Text{String: bahanPokok, Valid: bahanPokok != ""},
		})
		return err
	})

	g.Go(func() error {
		var err error
		totalCount, err = h.Queries.GetTotalBahanPokok(gCtx, bahanPokok)
		return err
	})

	g.Go(func() error {
		if areaSlugPilihan == "" {
			data, err := h.Queries.GetTrenSemuaBahanPokok(gCtx, db.GetTrenSemuaBahanPokokParams{
				Tanggal: tanggalPg,
			})
			riwayatData = data
			return err
		} else {
			data, err := h.Queries.GetTrenSemuaBahanPokokByArea(gCtx, db.GetTrenSemuaBahanPokokByAreaParams{
    		Tanggal: tanggalPg,
    		Slug:    areaSlugPilihan, 
			})
			for _, row := range data {
				riwayatData = append(riwayatData, db.GetTrenSemuaBahanPokokRow{
					BahanPokokID:  row.BahanPokokID,
					RataRataHarga: row.RataRataHarga,
				})
			}
			return err
		}
	})

	if err := g.Wait(); err != nil {
		slog.Error("public.bahan_pokok.list_error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal mengambil data komoditas",
		})
	}

	historyMap := buildHistoryMap(riwayatData)
	var finalData []map[string]interface{}
	for _, bp := range bahanData {
		prices := historyMap[bp.ID]
		harga, tren := calculateTrend(prices)
		finalData = append(finalData, map[string]interface{}{
			"id":             bp.ID,
			"komoditas":      bp.Nama,
			"slug":           bp.Slug,
			"satuan":         bp.Satuan,
			"gambar_url":     bp.GambarUrl.String,
			"harga_sekarang": harga,
			"tren":           tren,
		})
	}

	if finalData == nil {
		finalData = []map[string]interface{}{}
	}

	res := SuccessResponse{
		Pesan:      "Daftar Komoditas",
		Data:       finalData,
		Pagination: buildPaginationMeta(page, limit, totalCount),
	}
	return cacheJSON(c, cacheKey, CacheTTLBahan, res)
}

// =====================================================================
// PUBLIC: GET DETAIL BAHAN POKOK
// =====================================================================
func (h *SiskaperbapoHandler) GetDetailBahanPokok(c fiber.Ctx) error {
	slugParam := strings.TrimSpace(c.Params("slug"))
	if slugParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Slug komoditas tidak valid",
		})
	}

	hariIni := time.Now().Format("2006-01-02")
	tanggalReqStr := c.Query("tanggal", hariIni)
	areaInput, _ := utils.ValidateQueryString(c.Query("area"), 100, "area")

	areaSlugPilihan := ""
	if areaInput != "" {
		areaSlugPilihan = generateSlug(areaInput)
	}

	cacheKey := fmt.Sprintf("detail_bahan:%s:%s:%s", slugParam, tanggalReqStr, areaSlugPilihan)
	
	if respondCached(c, cacheKey) {
		return nil
	}

	parsedTime, err := time.Parse("2006-01-02", tanggalReqStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Format tanggal salah (gunakan YYYY-MM-DD)",
		})
	}
	tanggalReqPg := pgtype.Date{Time: parsedTime, Valid: true}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	bahanPokok, errBp := h.Queries.GetBahanPokokBySlug(ctx, slugParam)
	if errBp != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Code:    "ERR_NOT_FOUND",
			Message: "Komoditas tidak ditemukan",
		})
	}

	tanggalTerakhir, errTanggal := h.Queries.GetTanggalUpdateTerakhir(ctx, db.GetTanggalUpdateTerakhirParams{
		BahanPokokID: bahanPokok.ID,
		Tanggal:      tanggalReqPg,
	})

	if errTanggal != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Code:    "ERR_NO_DATA",
			Message: "Data harga belum tersedia hingga tanggal " + tanggalReqStr,
		})
	}

	g, gCtx := errgroup.WithContext(ctx)
	var daftarRes []db.GetDaftarHargaAreaRow
	var rataRes []db.GetRataRataArea15HariRow
	var grafikRes []db.GetRiwayatHargaRataRataRow

	g.Go(func() error {
		var err error
		daftarRes, err = h.Queries.GetDaftarHargaArea(gCtx, db.GetDaftarHargaAreaParams{
			BahanPokokID: bahanPokok.ID,
			Tanggal:      tanggalTerakhir,
		})
		return err
	})

	g.Go(func() error {
		var err error
		rataRes, err = h.Queries.GetRataRataArea15Hari(gCtx, db.GetRataRataArea15HariParams{
			BahanPokokID: bahanPokok.ID,
			Tanggal:      tanggalTerakhir,
		})
		return err
	})

	g.Go(func() error {
		var err error
		grafikRes, err = h.Queries.GetRiwayatHargaRataRata(gCtx, db.GetRiwayatHargaRataRataParams{
			BahanPokokID: bahanPokok.ID,
			Tanggal:      tanggalTerakhir,
		})
		return err
	})

	if err := g.Wait(); err != nil {
		slog.Error("public.detail_bahan.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal mengambil detail grafik komoditas",
		})
	}

	var hargaUtama int32 = 0
	var totalHarga int32 = 0
	var finalKabKota []map[string]interface{}

	for _, item := range daftarRes {
		totalHarga += item.Harga
		if item.AreaSlug == areaSlugPilihan {
			hargaUtama = item.Harga
		}
		finalKabKota = append(finalKabKota, map[string]interface{}{
			"area":      item.Area,
			"area_slug": item.AreaSlug,
			"harga":     item.Harga,
		})
	}

	if hargaUtama == 0 && len(daftarRes) > 0 {
		hargaUtama = totalHarga / int32(len(daftarRes))
	}

	var finalGrafik []map[string]interface{}
	prices := make([]int32, 0, len(grafikRes))
	
	for _, item := range grafikRes {
		finalGrafik = append(finalGrafik, map[string]interface{}{
			"tanggal":         item.Tanggal.Time.Format("2006-01-02"),
			"rata_rata_harga": item.RataRataHarga,
		})
		prices = append(prices, item.RataRataHarga)
	}

	_, tren := calculateTrend(prices)

	var statTertinggi, statTerendah interface{}
	if len(rataRes) > 0 {
		dataTertinggi := rataRes[0]
		dataTerendah := rataRes[len(rataRes)-1]
		statTertinggi = map[string]interface{}{
			"area": dataTertinggi.Area, "area_slug": dataTertinggi.AreaSlug, "harga": dataTertinggi.RataRata15Hari,
		}
		statTerendah = map[string]interface{}{
			"area": dataTerendah.Area, "area_slug": dataTerendah.AreaSlug, "harga": dataTerendah.RataRata15Hari,
		}
	}

	if finalGrafik == nil {
		finalGrafik = []map[string]interface{}{}
	}
	if finalKabKota == nil {
		finalKabKota = []map[string]interface{}{}
	}

	response := map[string]interface{}{
		"id":                  bahanPokok.ID,
		"komoditas":           bahanPokok.Nama,
		"slug":                bahanPokok.Slug,
		"satuan":              bahanPokok.Satuan,
		"gambar_url":          bahanPokok.GambarUrl.String,
		"tanggal":             tanggalReqStr,
		"tanggal_data_aktual": tanggalTerakhir.Time.Format("2006-01-02"),
		"area_pilihan":        areaSlugPilihan,
		"harga_utama":         hargaUtama,
		"tren":                tren,
		"grafik_riwayat":      finalGrafik,
		"list_kab_kota":       finalKabKota,
		"statistik_15_hari": map[string]interface{}{
			"tertinggi": statTertinggi,
			"terendah":  statTerendah,
		},
	}

	res := SuccessResponse{
		Pesan: "Detail Komoditas",
		Data:  response,
	}
	
	return cacheJSON(c, cacheKey, CacheTTLBahan, res)
}

// =====================================================================
// PUBLIC: GET SEMUA AREA (KOTA/KAB)
// =====================================================================
func (h *SiskaperbapoHandler) GetAllAreas(c fiber.Ctx) error {
	cacheKey := "areas:all"
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	areas, err := h.Queries.GetAllArea(ctx)
	if err != nil {
		slog.Error("public.areas.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal mengambil daftar area",
		})
	}

	var finalAreas []AreaItemResponse
	for _, a := range areas {
		finalAreas = append(finalAreas, AreaItemResponse{
			ID:   a.ID,
			Nama: a.Nama,
			Slug: a.Slug,
		})
	}

	res := SuccessResponse{Pesan: "Daftar Area", Data: finalAreas}
	return cacheJSON(c, cacheKey, CacheTTLArea, res)
}

// =====================================================================
// PUBLIC: SDUI MAIN DATA
// =====================================================================
func (h *SiskaperbapoHandler) GetSduiMainData(c fiber.Ctx) error {
	cacheKey := "sdui_main_interpolated"
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)
	var bahanData []db.BahanPokok
	var areaData []db.MasterArea
	var templateBytes []byte

	g.Go(func() error {
		data, err := h.Queries.GetAllBahanPokok(gCtx, db.GetAllBahanPokokParams{LimitData: 1000, OffsetData: 0})
		bahanData = data
		return err
	})

	g.Go(func() error {
		data, err := h.Queries.GetAllArea(gCtx)
		areaData = data
		return err
	})

	g.Go(func() error {
		templateURL := os.Getenv("SDUI_TEMPLATE_URL")
		if templateURL == "" {
			return errors.New("SDUI_TEMPLATE_URL tidak dikonfigurasi")
		}
		req, err := http.NewRequestWithContext(gCtx, http.MethodGet, templateURL, nil)
		if err != nil {
			return err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("gagal fetch template, status: %d", resp.StatusCode)
		}
		body, err := io.ReadAll(resp.Body)
		templateBytes = body
		return err
	})

	if err := g.Wait(); err != nil {
		slog.Error("public.sdui.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL",
			Message: "Gagal memuat layout SDUI",
		})
	}

	finalBahan := make([]string, 0, len(bahanData))
	for _, b := range bahanData {
		finalBahan = append(finalBahan, b.Nama)
	}

	finalAreas := make([]string, 0, len(areaData))
	for _, a := range areaData {
		finalAreas = append(finalAreas, a.Nama)
	}

	bahanJSONBytes, _ := json.Marshal(finalBahan)
	bahanStr := string(bahanJSONBytes)
	bahanReplacement := ""
	if len(bahanStr) >= 2 {
		bahanReplacement = bahanStr[1 : len(bahanStr)-1]
	}

	areaJSONBytes, _ := json.Marshal(finalAreas)
	areaStr := string(areaJSONBytes)
	areaReplacement := ""
	if len(areaStr) >= 2 {
		areaReplacement = areaStr[1 : len(areaStr)-1]
	}

	templateStr := string(templateBytes)
	templateStr = strings.ReplaceAll(templateStr, `"$ListBapok"`, bahanReplacement)
	templateStr = strings.ReplaceAll(templateStr, `"$ListArea"`, areaReplacement)
	templateStr = strings.ReplaceAll(templateStr, `$cardDataUrl`, os.Getenv("CARD_DATA_URL"))

	finalResponseBytes := []byte(templateStr)

	_ = cacheJSON(c, cacheKey, CacheTTLBahan, json.RawMessage(finalResponseBytes))
	c.Set("Content-Type", "application/json")
	return c.Send(finalResponseBytes)
}