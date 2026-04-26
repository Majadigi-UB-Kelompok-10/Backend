package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/farildzaky/bapenda-service/internal/cache"
	"github.com/farildzaky/bapenda-service/internal/db"
	"golang.org/x/sync/errgroup"
)

const ContextQueryTimeout = 5 * time.Second

type BapendaHandler struct {
	Queries *db.Queries
}

func NewBapendaHandler(q *db.Queries) *BapendaHandler {
	return &BapendaHandler{Queries: q}
}

func (h *BapendaHandler) RunCacheWarmup() {
	slog.Info("Memulai Cache Warmup...")
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		jenisData, err := h.Queries.GetDistinctJenis(ctx)
		if err != nil {
			return fmt.Errorf("gagal ambil jenis kendaraan: %v", err)
		}
		res := SuccessResponse{Pesan: "Sukses", Data: jenisData}
		cache.GlobalCache.Set("dropdown:jenis", res)
		return nil
	})

	g.Go(func() error {
		tarifData, err := h.Queries.GetAllTarifPKB(ctx)
		if err != nil {
			return fmt.Errorf("gagal ambil tarif PKB: %v", err)
		}
		cache.GlobalCache.Set("master:tarif_pkb", tarifData)
		return nil
	})

	if err := g.Wait(); err != nil {
		slog.Warn("Peringatan: Cache Warmup selesai dengan error (Aplikasi tetap berjalan)", slog.String("error", err.Error()))
		return
	}

	slog.Info("Cache Warmup Berhasil", slog.String("waktu", time.Since(startTime).String()))
}