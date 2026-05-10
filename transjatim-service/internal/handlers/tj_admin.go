package handlers

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"

	"github.com/farildzaky/transjatim-service/internal/db"
	"github.com/farildzaky/transjatim-service/internal/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/sync/errgroup"
)

// =====================================================================
// ADMIN: TERMINAL
// =====================================================================

func (h *TransJatimHandler) GetAllTerminalsAdmin(c fiber.Ctx) error {
	page, limit, offset := parsePagination(c)
	kotaFilter := c.Query("kota")

	cacheKey := normalizeKey("transjatim:admin:terminals", c.Query("page"), c.Query("limit"), kotaFilter)
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)
	var data []db.ListAllTerminalsAdminRow
	var total int64

	var pgKotaFilter pgtype.Text
	if kotaFilter != "" {
		pgKotaFilter = pgtype.Text{String: kotaFilter, Valid: true}
	}

	g.Go(func() error {
		var err error
		data, err = h.Queries.ListAllTerminalsAdmin(gCtx, db.ListAllTerminalsAdminParams{
			LimitData:  int32(limit),
			OffsetData: int32(offset),
			KotaFilter: pgKotaFilter,
		})
		return err
	})
	g.Go(func() error {
		var err error
		total, err = h.Queries.CountAllTerminals(gCtx, pgKotaFilter)
		return err
	})

	if err := g.Wait(); err != nil {
		slog.Error("admin.terminal.list.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code: "ERR_DB", Message: "Gagal mengambil data terminal",
		})
	}

	res := make([]TerminalResponse, 0, len(data))
	for _, t := range data {
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
			ID: t.ID, Nama: t.Nama, Kota: t.Kota, Slug: t.Slug, Lat: lat, Lng: lng, Aktif: t.Aktif,
		})
	}

	return cacheJSON(c, cacheKey, CacheTTLList, SuccessResponse{
		Pesan: "Sukses", Data: res, Pagination: buildPaginationMeta(page, limit, total),
	})
}

func (h *TransJatimHandler) CreateTerminal(c fiber.Ctx) error {
	var req TerminalRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_JSON", Message: "Format data tidak valid"})
	}

	nama, errVal := utils.ValidateTextContent(req.Nama, 150, "nama")
	if errVal != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: errVal.Message})
	}
	kota, errVal := utils.ValidateTextContent(req.Kota, 100, "kota")
	if errVal != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: errVal.Message})
	}

	var pgLat, pgLng pgtype.Float8
	if req.Lat != nil {
		pgLat = pgtype.Float8{Float64: *req.Lat, Valid: true}
	}
	if req.Lng != nil {
		pgLng = pgtype.Float8{Float64: *req.Lng, Valid: true}
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	slug := utils.GenerateSlug(nama + " " + kota)
	res, err := h.Queries.CreateTerminal(ctx, db.CreateTerminalParams{
		Nama: nama, Kota: kota, Slug: slug, Lat: pgLat, Lng: pgLng,
	})
	
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
				Code: "ERR_CONFLICT", Message: "Terminal dengan nama dan kota ini sudah ada",
			})
		}
		slog.Error("admin.terminal.create.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_INSERT", Message: "Gagal menyimpan terminal"})
	}

	invalidateTerminalCache()
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Pesan: "Terminal berhasil dibuat",
		Data: struct {
			ID   int32  `json:"id"`
			Slug string `json:"slug"`
		}{ID: res.ID, Slug: res.Slug},
	})
}

func (h *TransJatimHandler) UpdateTerminal(c fiber.Ctx) error {
	slugLama := c.Params("slug")
	var req TerminalRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_JSON", Message: "Format tidak valid"})
	}

	nama, errVal := utils.ValidateTextContent(req.Nama, 150, "nama")
	if errVal != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: errVal.Message})
	}
	kota, errVal := utils.ValidateTextContent(req.Kota, 100, "kota")
	if errVal != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: errVal.Message})
	}

	var pgLat, pgLng pgtype.Float8
	if req.Lat != nil {
		pgLat = pgtype.Float8{Float64: *req.Lat, Valid: true}
	}
	if req.Lng != nil {
		pgLng = pgtype.Float8{Float64: *req.Lng, Valid: true}
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	slugBaru := utils.GenerateSlug(nama + " " + kota)
	err := h.Queries.UpdateTerminal(ctx, db.UpdateTerminalParams{
		Nama: nama, Kota: kota, SlugBaru: slugBaru, Lat: pgLat, Lng: pgLng, Aktif: req.Aktif, SlugLama: slugLama,
	})
	
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
				Code: "ERR_CONFLICT", Message: "Nama & Kota terminal sudah dipakai",
			})
		}
		slog.Error("admin.terminal.update.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_UPDATE", Message: "Gagal update terminal"})
	}

	invalidateTerminalCache()
	return c.JSON(SuccessResponse{Pesan: "Terminal terupdate"})
}

func (h *TransJatimHandler) DeactivateTerminal(c fiber.Ctx) error {
	slug := c.Params("slug")
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.Queries.DeactivateTerminal(ctx, slug); err != nil {
		slog.Error("admin.terminal.deactivate.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_DELETE", Message: "Gagal menonaktifkan terminal"})
	}

	invalidateTerminalCache()
	return c.JSON(SuccessResponse{Pesan: "Terminal berhasil dinonaktifkan"})
}

// =====================================================================
// ADMIN: BUS
// =====================================================================

func (h *TransJatimHandler) GetAllBusesAdmin(c fiber.Ctx) error {
	page, limit, offset := parsePagination(c)

	cacheKey := normalizeKey("transjatim:admin:buses", c.Query("page"), c.Query("limit"))
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)
	var data []db.ListBusesAdminRow
	var total int64

	g.Go(func() error {
		var err error
		data, err = h.Queries.ListBusesAdmin(gCtx, db.ListBusesAdminParams{
			LimitData:  int32(limit),
			OffsetData: int32(offset),
		})
		return err
	})
	g.Go(func() error {
		var err error
		total, err = h.Queries.CountBuses(gCtx)
		return err
	})

	if err := g.Wait(); err != nil {
		slog.Error("admin.bus.list.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code: "ERR_DB", Message: "Gagal mengambil data Bus",
		})
	}

	res := make([]BusResponse, 0, len(data))
	for _, b := range data {
		var fas *string
		if b.Fasilitas.Valid {
			fas = &b.Fasilitas.String
		}
		var kap *int32
		if b.Kapasitas.Valid {
			kap = &b.Kapasitas.Int32
		}
		res = append(res, BusResponse{
			ID:        b.ID,
			Kode:      b.Kode,
			Layanan:   string(b.Layanan),
			Fasilitas: fas,
			Kapasitas: kap,
		})
	}

	return cacheJSON(c, cacheKey, CacheTTLList, SuccessResponse{
		Pesan: "Sukses", Data: res, Pagination: buildPaginationMeta(page, limit, total),
	})
}

func (h *TransJatimHandler) CreateBus(c fiber.Ctx) error {
	var req BusRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_JSON", Message: "Format data tidak valid"})
	}

	kode, errVal := utils.ValidateTextContent(req.Kode, 50, "kode")
	if errVal != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: errVal.Message})
	}

	layanan := strings.ToLower(req.Layanan)
	if layanan != "reguler" && layanan != "luxury" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "Layanan harus reguler atau luxury"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	var pgFasilitas pgtype.Text
	if req.Fasilitas != nil && *req.Fasilitas != "" {
		pgFasilitas = pgtype.Text{String: *req.Fasilitas, Valid: true}
	}

	var pgKapasitas pgtype.Int4
	if req.Kapasitas != nil && *req.Kapasitas > 0 {
		pgKapasitas = pgtype.Int4{Int32: *req.Kapasitas, Valid: true}
	}

	res, err := h.Queries.CreateBus(ctx, db.CreateBusParams{
		Kode:      kode,
		Layanan:   db.LayananType(layanan),
		Fasilitas: pgFasilitas,
		Kapasitas: pgKapasitas,
	})
	
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
				Code: "ERR_CONFLICT", Message: "Kode Bus sudah terdaftar",
			})
		}
		slog.Error("admin.bus.create.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_INSERT", Message: "Gagal menyimpan Bus"})
	}

	invalidateBusCache()
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Pesan: "Bus berhasil ditambahkan",
		Data: struct {
			ID int32 `json:"id"`
		}{ID: res},
	})
}

func (h *TransJatimHandler) UpdateBus(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "ID tidak valid"})
	}

	var req BusRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_JSON", Message: "Format tidak valid"})
	}

	kode, errVal := utils.ValidateTextContent(req.Kode, 50, "kode")
	if errVal != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: errVal.Message})
	}
	layanan := strings.ToLower(req.Layanan)
	if layanan != "reguler" && layanan != "luxury" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "Layanan harus reguler atau luxury"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	var pgFasilitas pgtype.Text
	if req.Fasilitas != nil && *req.Fasilitas != "" {
		pgFasilitas = pgtype.Text{String: *req.Fasilitas, Valid: true}
	}

	var pgKapasitas pgtype.Int4
	if req.Kapasitas != nil && *req.Kapasitas > 0 {
		pgKapasitas = pgtype.Int4{Int32: *req.Kapasitas, Valid: true}
	}

	errDb := h.Queries.UpdateBus(ctx, db.UpdateBusParams{
		ID:        int32(id),
		Kode:      kode,
		Layanan:   db.LayananType(layanan),
		Fasilitas: pgFasilitas,
		Kapasitas: pgKapasitas,
	})
	
	if errDb != nil {
		var pgErr *pgconn.PgError
		if errors.As(errDb, &pgErr) && pgErr.Code == "23505" {
			return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
				Code: "ERR_CONFLICT", Message: "Kode Bus sudah dipakai armada lain",
			})
		}
		slog.Error("admin.bus.update.error", slog.String("err", errDb.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_UPDATE", Message: "Gagal update Bus"})
	}

	invalidateBusCache()
	return c.JSON(SuccessResponse{Pesan: "Bus terupdate"})
}

func (h *TransJatimHandler) DeleteBus(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "ID tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.Queries.DeleteBus(ctx, int32(id)); err != nil {
		slog.Error("admin.bus.delete.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_DELETE", Message: "Gagal menghapus Bus. Mungkin sedang dipakai di Jadwal."})
	}

	invalidateBusCache()
	return c.JSON(SuccessResponse{Pesan: "Bus berhasil dihapus"})
}