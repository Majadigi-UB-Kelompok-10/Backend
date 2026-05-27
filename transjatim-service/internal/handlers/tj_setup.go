package handlers

import (
	"context"
	"log/slog"
	"time"

	"github.com/farildzaky/transjatim-service/internal/cache"
	"github.com/farildzaky/transjatim-service/internal/db"
)

// =====================================================================
// BASE STRUCT & CONSTRUCTOR
// =====================================================================

type TransJatimHandler struct {
	Queries   *db.Queries
	OrsApiKey string
}

func NewTransJatimHandler(q *db.Queries, orsApiKey string) *TransJatimHandler {
	return &TransJatimHandler{
		Queries:   q,
		OrsApiKey: orsApiKey,
	}
}

// =====================================================================
// HELPER: CACHE INVALIDATION
// =====================================================================

func invalidateTerminalCache() {
	cache.GlobalCache.DeleteByPrefix("transjatim:public:terminals")
	cache.GlobalCache.DeleteByPrefix("transjatim:admin:terminals")
	cache.GlobalCache.DeleteByPrefix("transjatim:dropdown:kota")
}

func invalidateBusCache() {
	cache.GlobalCache.DeleteByPrefix("transjatim:admin:buses")
	cache.GlobalCache.DeleteByPrefix("transjatim:dropdown:layanan")
	cache.GlobalCache.DeleteByPrefix("transjatim:public:jadwal") 
}

func invalidateRuteCache() {
	cache.GlobalCache.DeleteByPrefix("transjatim:admin:rutes")
	cache.GlobalCache.DeleteByPrefix("transjatim:public:jadwal")
}

func invalidateJadwalCache() {
	cache.GlobalCache.DeleteByPrefix("transjatim:admin:jadwal")
	cache.GlobalCache.DeleteByPrefix("transjatim:public:jadwal")
}

func invalidateHargaCache() {
	cache.GlobalCache.DeleteByPrefix("transjatim:admin:harga")
	cache.GlobalCache.DeleteByPrefix("transjatim:public:jadwal")
}

// =====================================================================
// CACHE WARMUP (Dijalankan saat server baru up)
// =====================================================================

func (h *TransJatimHandler) CacheWarmup() {
	slog.Info("cache.warmup.transjatim.start")
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	terminals, err := h.Queries.ListActiveTerminals(ctx)
	if err != nil {
		slog.Warn("cache.warmup.transjatim.terminals.failed", slog.String("err", err.Error()))
		return
	}

	res := make([]TerminalResponse, 0, len(terminals))
	for _, t := range terminals {
		var lat, lng *float64
		if t.Lat.Valid {
			latVal := t.Lat.Float64
			lat = &latVal
		}
		if t.Lng.Valid {
			lngVal := t.Lng.Float64
			lng = &lngVal
		}

		res = append(res, TerminalResponse{
			ID: t.ID, Nama: t.Nama, Kota: t.Kota, Slug: t.Slug, Lat: lat, Lng: lng, Aktif: true, 
		})
	}

	cacheKey := "transjatim:public:terminals:active"
	cache.GlobalCache.Set(cacheKey, SuccessResponse{
		Pesan: "Sukses", Data: res,
	}, CacheTTLTerminals)

	slog.Info("cache.warmup.transjatim.success", 
        slog.Int("terminals_loaded", len(res)),
        slog.String("duration", time.Since(startTime).String()),
    )
}