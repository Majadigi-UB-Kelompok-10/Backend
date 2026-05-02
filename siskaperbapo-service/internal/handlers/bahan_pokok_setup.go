package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/farildzaky/siskaperbapo-service/internal/cache"
	"github.com/farildzaky/siskaperbapo-service/internal/db"
	"github.com/farildzaky/siskaperbapo-service/internal/worker"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
)

var slugRegex = regexp.MustCompile(`[^a-z0-9]+`)

type SiskaperbapoHandler struct {
	Queries    *db.Queries
	DB         *pgxpool.Pool
	Cld        *cloudinary.Cloudinary
	WorkerPool *worker.HargaWorkerPool
}

func NewSiskaperbapoHandler(q *db.Queries, dbPool *pgxpool.Pool, cld *cloudinary.Cloudinary, wp *worker.HargaWorkerPool) *SiskaperbapoHandler {
	return &SiskaperbapoHandler{
		Queries:    q,
		DB:         dbPool,
		Cld:        cld,
		WorkerPool: wp,
	}
}

// =============================================================================
// HELPER: SLUG & CACHE INVALIDATION
// =============================================================================

func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = slugRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "unnamed"
	}
	return slug
}

func invalidateAllBahanCache() {
	cache.GlobalCache.DeleteByPrefix("all_bahan:")
	cache.GlobalCache.DeleteByPrefix("detail_bahan:")
	cache.GlobalCache.DeleteByPrefix("sdui_main_interpolated")
}

// =============================================================================
// LOGIKA BISNIS: TREN HARGA
// =============================================================================

func calculateTrend(prices []int32) (int32, string) {
	if len(prices) == 0 {
		return 0, "TETAP"
	}

	hargaTerkini := prices[len(prices)-1]
	trend := "TETAP"

	for i := len(prices) - 2; i >= 0; i-- {
		hargaMasaLalu := prices[i]
		if hargaMasaLalu != hargaTerkini {
			if hargaTerkini > hargaMasaLalu {
				trend = "NAIK"
			} else {
				trend = "TURUN"
			}
			break
		}
	}

	return hargaTerkini, trend
}

func buildHistoryMap(riwayatData []db.GetTrenSemuaBahanPokokRow) map[int32][]int32 {
	historyMap := make(map[int32][]int32)
	for _, row := range riwayatData {
		historyMap[row.BahanPokokID] = append(historyMap[row.BahanPokokID], row.RataRataHarga)
	}
	return historyMap
}

// =============================================================================
// CACHE WARMUP
// =============================================================================

func (h *SiskaperbapoHandler) CacheWarmup() {
	slog.Info("cache.warmup.siskaperbapo.start")
	startTime := time.Now()

	hariIni := time.Now().Format("2006-01-02")
	parsedTime, _ := time.Parse("2006-01-02", hariIni)
	tanggalPg := pgtype.Date{Time: parsedTime, Valid: true}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)

	var bahanData []db.BahanPokok
	var totalCount int64
	var trenData []db.GetTrenSemuaBahanPokokRow

	g.Go(func() error {
		data, err := h.Queries.GetAllBahanPokok(gCtx, db.GetAllBahanPokokParams{LimitData: 10, OffsetData: 0})
		bahanData = data
		return err
	})
	g.Go(func() error {
		count, err := h.Queries.GetTotalBahanPokok(gCtx, "")
		totalCount = count
		return err
	})
	g.Go(func() error {
		data, err := h.Queries.GetTrenSemuaBahanPokok(gCtx, db.GetTrenSemuaBahanPokokParams{Tanggal: tanggalPg})
		trenData = data
		return err
	})

	if err := g.Wait(); err != nil {
		slog.Warn("cache.warmup.partial_failure", slog.String("err", err.Error()))
		return
	}

	historyMap := buildHistoryMap(trenData)
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

	cacheKey := fmt.Sprintf("all_bahan:%s:page_1:limit_10:bahan_pokok_:area_", hariIni)
	res := SuccessResponse{
		Pesan:      "Daftar Komoditas (Warmup)",
		Data:       finalData,
		Pagination: buildPaginationMeta(1, 10, totalCount),
	}
	cache.GlobalCache.Set(cacheKey, res, CacheTTLBahan)

	slog.Info("cache.warmup.siskaperbapo.success", slog.String("duration", time.Since(startTime).String()))
}