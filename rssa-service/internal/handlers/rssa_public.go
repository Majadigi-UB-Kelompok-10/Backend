package handlers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/farildzaky/rssa-service/internal/db"
	"github.com/farildzaky/rssa-service/internal/utils"
	"github.com/gofiber/fiber/v3"
	"golang.org/x/sync/errgroup"
)

// =============================================================================
// PUBLIC: GET HEADER SUMMARY (Tersedia vs Total)
// =============================================================================

func (h *RSSAHandler) GetHeaderSummary(c fiber.Ctx) error {
	cacheKey := "rssa:summary_ruangan"

	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	summaryRow, err := h.Queries.GetSummaryRuangan(ctx)
	if err != nil {
		slog.Error("public.summary.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal mengambil data summary ruangan",
		})
	}

	response := SummaryRuanganResponse{
		TotalKapasitas: int(summaryRow.TotalKapasitas),
		TotalTersedia:  int(summaryRow.TotalTersedia),
	}

	res := SuccessResponse{Pesan: "Sukses", Data: response}
	return cacheJSON(c, cacheKey, CacheTTLPublic, res)
}

// =============================================================================
// PUBLIC: GET MASTER KELAS (Untuk Filter Chips)
// =============================================================================

func (h *RSSAHandler) GetMasterKelas(c fiber.Ctx) error {
	cacheKey := "rssa:master_kelas"

	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetMasterKelas(ctx)
	if err != nil {
		slog.Error("public.master_kelas.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal mengambil daftar kelas",
		})
	}

	response := make([]MasterKelasResponse, 0, len(data))
	for _, row := range data {
		response = append(response, MasterKelasResponse{
			ID:   row.ID,
			Nama: row.Nama,
			Slug: row.Slug,
		})
	}

	res := SuccessResponse{Pesan: "Daftar Master Kelas", Data: response}
	return cacheJSON(c, cacheKey, CacheTTLKelas, res)
}

// =============================================================================
// PUBLIC: GET LIST RUANGAN (Pencarian + Filter Kelas)
// =============================================================================

func (h *RSSAHandler) GetListRuangan(c fiber.Ctx) error {
	// 1. Ambil Parameter Query
	keyword, _ := utils.SanitizeQueryString(c.Query("search"), 100, "search")
	kelasSlug, _ := utils.ValidateSlug(c.Query("kelas"), "kelas") 

	// 2. Buat Cache Key Spesifik (gabungan keyword dan kelas_slug)
	cacheKey := fmt.Sprintf("rssa:list_ruangan:q_%s:k_%s", keyword, kelasSlug)
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)

	var listData []db.SearchRuanganRow
	var totalData int64

	// Go-routine 1: Ambil daftar data
	g.Go(func() error {
		data, err := h.Queries.SearchRuangan(gCtx, db.SearchRuanganParams{
			Keyword:   keyword,   
			KelasSlug: kelasSlug, 
		})
		if err != nil {
			return err
		}
		listData = data
		return nil
	})

	// Go-routine 2: Hitung total data (Jika tidak ada filter, return semua)
	g.Go(func() error {
		count, err := h.Queries.CountSearchRuangan(gCtx, db.CountSearchRuanganParams{
			Keyword:   keyword,   
			KelasSlug: kelasSlug, 
		})
		if err != nil {
			return err
		}
		totalData = count
		return nil
	})

	// Tunggu DB selesai mencari
	if err := g.Wait(); err != nil {
		slog.Error("public.list_ruangan.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal memuat daftar ruangan",
		})
	}

	// 3. Mapping Data ke DTO
	responseData := make([]RuanganPublicResponse, 0, len(listData))
	for _, row := range listData {
		responseData = append(responseData, RuanganPublicResponse{
			ID:        row.ID,
			Nama:      row.Nama,
			Slug:      row.Slug,
			KelasNama: row.KelasNama,
			KelasSlug: row.KelasSlug,
			Kapasitas: row.Kapasitas,
			Terisi:    row.Terisi,
			Tersedia:  row.Tersedia, // Dihitung otomatis oleh Postgres Generated Column!
		})
	}

	pesan := "Daftar Ruangan RSSA"
	if keyword != "" || kelasSlug != "" {
		pesan = "Hasil Pencarian Ruangan"
	}

	res := SuccessResponse{
		Pesan:      pesan,
		Data:       responseData,
		Pagination: &PaginationMeta{Page: 1, Limit: int(totalData), Total: totalData}, // Karena kita memanggil semua list sekaligus (tanpa offset DB)
	}

	return cacheJSON(c, cacheKey, CacheTTLPublic, res)
}