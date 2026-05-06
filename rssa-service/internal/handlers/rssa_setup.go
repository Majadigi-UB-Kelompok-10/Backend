package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/farildzaky/rssa-service/internal/cache"
	"github.com/farildzaky/rssa-service/internal/db"
	"golang.org/x/sync/errgroup"
)

// =============================================================================
// HANDLER INJECTION & CONSTRUCTOR
// =============================================================================

type RSSAHandler struct {
	Queries *db.Queries
}

func NewRSSAHandler(q *db.Queries) *RSSAHandler {
	return &RSSAHandler{
		Queries: q,
	}
}

// =============================================================================
// CACHE WARMUP (Dijalankan di background saat server start)
// =============================================================================

func (h *RSSAHandler) RunCacheWarmup() {
	slog.Info("cache.warmup.rssa.start")
	startTime := time.Now()

	// Gunakan timeout agar jika DB nyangkut, server tetap bisa start
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)

	// 1. Warmup Master Kelas (Digunakan untuk Filter Chips Frontend)
	g.Go(func() error {
		kelasData, err := h.Queries.GetMasterKelas(gCtx)
		if err != nil {
			return fmt.Errorf("warmup master_kelas: %w", err)
		}

		response := make([]MasterKelasResponse, 0, len(kelasData))
		for _, row := range kelasData {
			response = append(response, MasterKelasResponse{
				ID:   row.ID,
				Nama: row.Nama,
				Slug: row.Slug,
			})
		}

		// Simpan ke cache (karena master data jarang berubah, pakai TTL panjang)
		cache.GlobalCache.Set("rssa:master_kelas", SuccessResponse{
			Pesan: "Sukses",
			Data:  response,
		}, CacheTTLKelas)

		return nil
	})

	// 2. Warmup Summary Total Ruangan
	g.Go(func() error {
		summaryRow, err := h.Queries.GetSummaryRuangan(gCtx)
		if err != nil {
			return fmt.Errorf("warmup summary_ruangan: %w", err)
		}

		// Konversi pgtype.Int8 ke int (Aman karena database sudah pakai coalesce 0)
		totalKapasitas := int(summaryRow.TotalKapasitas)
		totalTersedia := int(summaryRow.TotalTersedia)

		response := SummaryRuanganResponse{
			TotalKapasitas: totalKapasitas,
			TotalTersedia:  totalTersedia,
		}

		cache.GlobalCache.Set("rssa:summary_ruangan", SuccessResponse{
			Pesan: "Sukses",
			Data:  response,
		}, CacheTTLPublic)

		return nil
	})

	// Tunggu semua proses background selesai
	if err := g.Wait(); err != nil {
		slog.Warn("cache.warmup.partial_failure",
			slog.String("err", err.Error()),
			slog.String("note", "service tetap jalan, cache akan diisi on-demand"),
		)
		return
	}

	slog.Info("cache.warmup.rssa.success",
		slog.String("duration", time.Since(startTime).String()),
	)
}

// =============================================================================
// CACHE INVALIDATION HELPER
// =============================================================================


func invalidateRuanganCache() {
	cache.GlobalCache.Delete("rssa:summary_ruangan")

	cache.GlobalCache.DeleteByPrefix("rssa:list_ruangan:")
}