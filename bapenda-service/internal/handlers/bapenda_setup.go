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
	slog.Info("cache.warmup.start")
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		jenisData, err := h.Queries.GetDistinctJenis(gCtx)
		if err != nil {
			return fmt.Errorf("warmup jenis: %w", err)
		}
		cache.GlobalCache.SetWithTTL(
			"dropdown:jenis",
			SuccessResponse{Pesan: "Sukses", Data: jenisData},
			CacheTTLStatic,
		)
		return nil
	})

	g.Go(func() error {
		tarifData, err := h.Queries.GetAllTarifPKB(gCtx)
		if err != nil {
			return fmt.Errorf("warmup tarif: %w", err)
		}
		cache.GlobalCache.SetWithTTL("master:tarif_pkb", tarifData, CacheTTLStatic)
		return nil
	})

	if err := g.Wait(); err != nil {
		slog.Warn("cache.warmup.partial_failure",
			slog.String("err", err.Error()),
			slog.String("note", "service tetap jalan, cache akan diisi on-demand"),
		)
		return
	}

	slog.Info("cache.warmup.success",
		slog.String("duration", time.Since(startTime).String()),
	)
}