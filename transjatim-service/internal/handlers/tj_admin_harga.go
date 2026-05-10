package handlers

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"

	"github.com/farildzaky/transjatim-service/internal/db"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/sync/errgroup"
)

// =====================================================================
// ADMIN: HARGA TIKET
// =====================================================================

func (h *TransJatimHandler) GetAllHargaAdmin(c fiber.Ctx) error {
	page, limit, offset := parsePagination(c)
	cacheKey := normalizeKey("transjatim:admin:harga", c.Query("page"), c.Query("limit"))
	
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)
	var data []db.ListHargaAdminRow
	var total int64

	g.Go(func() error {
		var err error
		data, err = h.Queries.ListHargaAdmin(gCtx, db.ListHargaAdminParams{
			LimitData:  int32(limit),
			OffsetData: int32(offset),
		})
		return err
	})
	g.Go(func() error {
		var err error
		total, err = h.Queries.CountHarga(gCtx)
		return err
	})

	if err := g.Wait(); err != nil {
		slog.Error("admin.harga.list.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code: "ERR_DB", Message: "Gagal mengambil data Harga",
		})
	}

	res := make([]HargaAdminItem, 0, len(data))
	for _, h := range data {
		var tipe *string
		if h.TipePenumpang.Valid {
			str := string(h.TipePenumpang.TipePenumpangType)
			tipe = &str
		}
		res = append(res, HargaAdminItem{
			ID:             h.ID,
			RuteID:         h.RuteID,
			TerminalAsal:   h.TerminalAsalNama,
			TerminalTujuan: h.TerminalTujuanNama,
			Layanan:        string(h.Layanan),
			TipePenumpang:  tipe,
			Harga:          h.Harga,
		})
	}

	return cacheJSON(c, cacheKey, CacheTTLList, SuccessResponse{
		Pesan: "Sukses", Data: res, Pagination: buildPaginationMeta(page, limit, total),
	})
}

func (h *TransJatimHandler) CreateHarga(c fiber.Ctx) error {
	var req HargaRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_JSON", Message: "Format data tidak valid"})
	}

	if req.Harga <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "Harga harus lebih dari Rp0"})
	}

	layanan := strings.ToLower(req.Layanan)
	if layanan != "reguler" && layanan != "luxury" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "Layanan harus reguler atau luxury"})
	}

	var pgTipePenumpang db.NullTipePenumpangType
	if layanan == "luxury" {
		pgTipePenumpang = db.NullTipePenumpangType{Valid: false}
	} else {
		if req.TipePenumpang == nil || strings.TrimSpace(*req.TipePenumpang) == "" {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Code: "ERR_VALIDATION", Message: "Bus reguler wajib memilih tipe penumpang (umum/pelajar_santri/mahasiswa)",
			})
		}
		tipe := strings.ToLower(*req.TipePenumpang)
		if tipe != "umum" && tipe != "pelajar_santri" && tipe != "mahasiswa" {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Code: "ERR_VALIDATION", Message: "Tipe penumpang tidak valid",
			})
		}
		pgTipePenumpang = db.NullTipePenumpangType{
			TipePenumpangType: db.TipePenumpangType(tipe),
			Valid:             true,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	res, err := h.Queries.CreateHarga(ctx, db.CreateHargaParams{
		RuteID:        req.RuteID,
		Layanan:       db.LayananType(layanan),
		TipePenumpang: pgTipePenumpang,
		Harga:         req.Harga,
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
				Code: "ERR_CONFLICT", Message: "Harga untuk kombinasi rute, layanan, dan tipe ini sudah ada",
			})
		}
		slog.Error("admin.harga.create.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_INSERT", Message: "Gagal menyimpan Harga"})
	}

	invalidateHargaCache()
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Pesan: "Harga tiket berhasil ditambahkan",
		Data: struct {
			ID int32 `json:"id"`
		}{ID: res},
	})
}

func (h *TransJatimHandler) UpdateHarga(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "ID tidak valid"})
	}

	var req HargaRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_JSON", Message: "Format tidak valid"})
	}

	if req.Harga <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "Harga harus lebih dari Rp0"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.Queries.UpdateHarga(ctx, db.UpdateHargaParams{
		Harga: req.Harga,
		ID:    int32(id),
	}); err != nil {
		slog.Error("admin.harga.update.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_UPDATE", Message: "Gagal update Harga"})
	}

	invalidateHargaCache()
	return c.JSON(SuccessResponse{Pesan: "Harga berhasil diperbarui"})
}

func (h *TransJatimHandler) DeleteHarga(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "ID tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.Queries.DeleteHarga(ctx, int32(id)); err != nil {
		slog.Error("admin.harga.delete.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_DELETE", Message: "Gagal menghapus Harga."})
	}

	invalidateHargaCache()
	return c.JSON(SuccessResponse{Pesan: "Harga berhasil dihapus"})
}