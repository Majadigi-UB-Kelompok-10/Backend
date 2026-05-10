package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/farildzaky/transjatim-service/internal/db"
	"github.com/farildzaky/transjatim-service/internal/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/sync/errgroup"
)

// =====================================================================
// ADMIN: RUTE
// =====================================================================

func (h *TransJatimHandler) GetAllRutesAdmin(c fiber.Ctx) error {
	page, limit, offset := parsePagination(c)
	cacheKey := normalizeKey("transjatim:admin:rutes", c.Query("page"), c.Query("limit"))

	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)
	var data []db.ListRutesAdminRow
	var total int64

	g.Go(func() error {
		var err error
		data, err = h.Queries.ListRutesAdmin(gCtx, db.ListRutesAdminParams{
			LimitData:  int32(limit),
			OffsetData: int32(offset),
		})
		return err
	})
	g.Go(func() error {
		var err error
		total, err = h.Queries.CountRutes(gCtx)
		return err
	})

	if err := g.Wait(); err != nil {
		slog.Error("admin.rute.list.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code: "ERR_DB", Message: "Gagal mengambil data Rute",
		})
	}

	res := make([]RuteResponse, 0, len(data))
	for _, r := range data {
		res = append(res, RuteResponse{
			ID:             r.ID,
			Slug:           r.Slug,
			TerminalAsal:   r.TerminalAsalNama,
			KotaAsal:       r.KotaAsal,
			TerminalTujuan: r.TerminalTujuanNama,
			KotaTujuan:     r.KotaTujuan,
			DurasiMenit:    r.DurasiMenit,
			Aktif:          r.Aktif,
		})
	}

	return cacheJSON(c, cacheKey, CacheTTLList, SuccessResponse{
		Pesan: "Sukses", Data: res, Pagination: buildPaginationMeta(page, limit, total),
	})
}

func (h *TransJatimHandler) CreateRute(c fiber.Ctx) error {
	var req RuteRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_JSON", Message: "Format data tidak valid"})
	}

	if req.TerminalAsalID == req.TerminalTujuanID {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "Terminal asal dan tujuan tidak boleh sama"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	asal, err := h.Queries.GetTerminalByID(ctx, req.TerminalAsalID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Code: "ERR_NOT_FOUND", Message: "Terminal asal tidak ditemukan"})
	}
	tujuan, err := h.Queries.GetTerminalByID(ctx, req.TerminalTujuanID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Code: "ERR_NOT_FOUND", Message: "Terminal tujuan tidak ditemukan"})
	}
	slug := utils.GenerateSlug(fmt.Sprintf("%s %s ke %s %s", asal.Kota, asal.Nama, tujuan.Kota, tujuan.Nama))

	res, err := h.Queries.CreateRute(ctx, db.CreateRuteParams{
		TerminalAsalID:   req.TerminalAsalID,
		TerminalTujuanID: req.TerminalTujuanID,
		Slug:             slug,
		DurasiMenit:      req.DurasiMenit,
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
				Code: "ERR_CONFLICT", Message: "Rute untuk kombinasi terminal ini sudah ada",
			})
		}
		slog.Error("admin.rute.create.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_INSERT", Message: "Gagal menyimpan rute"})
	}

	for i, terminalID := range req.Stops {
		if err := h.Queries.InsertRuteStop(ctx, db.InsertRuteStopParams{
			RuteID:     res.ID,
			TerminalID: terminalID,
			Urutan:     int32(i + 1),
		}); err != nil {
			slog.Error("admin.rute.create.stop_error",
				slog.String("rute_slug", res.Slug),
				slog.Int("urutan", i+1),
				slog.String("err", err.Error()),
			)
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
				Code: "ERR_STOP_INSERT", Message: "Gagal menyimpan pemberhentian rute",
			})
		}
	}

	invalidateRuteCache()
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Pesan: "Rute berhasil dibuat",
		Data: struct {
			ID   int32  `json:"id"`
			Slug string `json:"slug"`
		}{ID: res.ID, Slug: res.Slug},
	})
}

func (h *TransJatimHandler) UpdateRute(c fiber.Ctx) error {
	slugLama := c.Params("slug")
	var req RuteRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_JSON", Message: "Format tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	asal, err := h.Queries.GetTerminalByID(ctx, req.TerminalAsalID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Code: "ERR_NOT_FOUND", Message: "Terminal asal tidak ditemukan"})
	}
	tujuan, err := h.Queries.GetTerminalByID(ctx, req.TerminalTujuanID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Code: "ERR_NOT_FOUND", Message: "Terminal tujuan tidak ditemukan"})
	}
	slugBaru := utils.GenerateSlug(fmt.Sprintf("%s %s ke %s %s", asal.Kota, asal.Nama, tujuan.Kota, tujuan.Nama))

	errUpdate := h.Queries.UpdateRute(ctx, db.UpdateRuteParams{
		TerminalAsalID:   req.TerminalAsalID,
		TerminalTujuanID: req.TerminalTujuanID,
		SlugBaru:         slugBaru,
		DurasiMenit:      req.DurasiMenit,
		Aktif:            req.Aktif,
		SlugLama:         slugLama,
	})

	if errUpdate != nil {
		var pgErr *pgconn.PgError
		if errors.As(errUpdate, &pgErr) && pgErr.Code == "23505" {
			return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
				Code: "ERR_CONFLICT", Message: "Rute untuk kombinasi terminal ini sudah dipakai",
			})
		}
		slog.Error("admin.rute.update.error", slog.String("err", errUpdate.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_UPDATE", Message: "Gagal update rute"})
	}


	ruteDb, err := h.Queries.GetRuteBySlug(ctx, slugBaru)
	if err != nil {
		slog.Error("admin.rute.update.get_rute_error",
			slog.String("slug", slugBaru),
			slog.String("err", err.Error()),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code: "ERR_INTERNAL", Message: "Gagal mengambil data rute setelah update",
		})
	}

	_ = h.Queries.DeleteRuteStops(ctx, ruteDb.ID)
	for i, terminalID := range req.Stops {
		if err := h.Queries.InsertRuteStop(ctx, db.InsertRuteStopParams{
			RuteID:     ruteDb.ID,
			TerminalID: terminalID,
			Urutan:     int32(i + 1),
		}); err != nil {
			slog.Error("admin.rute.update.stop_error",
				slog.String("rute_slug", slugBaru),
				slog.Int("urutan", i+1),
				slog.String("err", err.Error()),
			)
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
				Code: "ERR_STOP_INSERT", Message: "Gagal menyimpan pemberhentian rute",
			})
		}
	}

	invalidateRuteCache()
	return c.JSON(SuccessResponse{Pesan: "Rute berhasil diperbarui"})
}

func (h *TransJatimHandler) DeactivateRute(c fiber.Ctx) error {
	slug := c.Params("slug")
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.Queries.DeactivateRute(ctx, slug); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_DELETE", Message: "Gagal menonaktifkan rute"})
	}

	invalidateRuteCache()
	return c.JSON(SuccessResponse{Pesan: "Rute berhasil dinonaktifkan"})
}