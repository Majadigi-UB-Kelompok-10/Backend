package handlers

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/sync/errgroup"

	"github.com/farildzaky/bansos-service/internal/db"
	"github.com/farildzaky/bansos-service/internal/utils"
)

// Helper Invalidate Cache Global Bansos
func (h *BansosHandler) invalidateBansosCache() {
	h.Cache.DeleteByPrefix("bansos:public:")
}

// =============================================================================
// ADMIN: PROGRAM BANSOS
// =============================================================================

func (h *BansosHandler) AdminListProgram(c fiber.Ctx) error {
	page, limit, offset := parsePagination(c)
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)
	var data []db.ListProgramAdminRow
	var total int64

	g.Go(func() error {
		var err error
		data, err = h.Queries.ListProgramAdmin(gCtx, db.ListProgramAdminParams{
			LimitData:  int32(limit),
			OffsetData: int32(offset),
		})
		return err
	})
	g.Go(func() error {
		var err error
		total, err = h.Queries.CountProgram(gCtx)
		return err
	})

	if err := g.Wait(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_DB", Message: "Gagal memuat program bansos"})
	}

	resData := make([]AdminProgramItem, 0, len(data))
	for _, r := range data {
		resData = append(resData, AdminProgramItem{
			ID:        int(r.ID),
			Nama:      r.Nama,
			Kode:      r.Kode,
			Deskripsi: r.Deskripsi.String,
			Aktif:     r.Aktif,
		})
	}

	return c.JSON(PaginatedResponse{Pesan: "Data Program Bansos", Data: resData, Pagination: &PaginationMeta{Page: page, Limit: limit, Total: total}})
}

func (h *BansosHandler) AdminCreateProgram(c fiber.Ctx) error {
	var req struct {
		Nama      string `json:"nama"`
		Kode      string `json:"kode"`
		Deskripsi string `json:"deskripsi"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "Format request body salah"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	_, err := h.Queries.CreateProgram(ctx, db.CreateProgramParams{
		Nama:      req.Nama,
		Kode:      req.Kode,
		Deskripsi: pgtype.Text{String: req.Deskripsi, Valid: req.Deskripsi != ""},
	})
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(ErrorResponse{Code: "ERR_CONFLICT", Message: "Gagal menyimpan. Pastikan Kode atau Nama tidak duplikat."})
	}

	h.invalidateBansosCache()
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{Pesan: "Program bansos berhasil ditambahkan"})
}

// =============================================================================
// ADMIN: PENERIMA (WARGA)
// =============================================================================

func (h *BansosHandler) AdminListPenerima(c fiber.Ctx) error {
	page, limit, offset := parsePagination(c)
	keyword := c.Query("keyword") // Bisa cari NIK atau Nama

	var keywordArg pgtype.Text
	if keyword != "" {
		keywordArg = pgtype.Text{String: keyword, Valid: true}
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)
	var data []db.ListPenerimaAdminRow
	var total int64

	g.Go(func() error {
		var err error
		data, err = h.Queries.ListPenerimaAdmin(gCtx, db.ListPenerimaAdminParams{
			Keyword:    keywordArg,
			LimitData:  int32(limit),
			OffsetData: int32(offset),
		})
		return err
	})
	g.Go(func() error {
		var err error
		total, err = h.Queries.CountPenerima(gCtx, keywordArg)
		return err
	})

	if err := g.Wait(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_DB", Message: "Gagal memuat data penerima"})
	}

	resData := make([]AdminPenerimaItem, 0, len(data))
	for _, r := range data {
		resData = append(resData, AdminPenerimaItem{
			ID:     int(r.ID),
			NIK:    r.Nik,
			Nama:   r.Nama,
			Alamat: r.Alamat,
		})
	}

	return c.JSON(PaginatedResponse{Pesan: "Data Penerima", Data: resData, Pagination: &PaginationMeta{Page: page, Limit: limit, Total: total}})
}

func (h *BansosHandler) AdminCreatePenerima(c fiber.Ctx) error {
	var req struct {
		NIK    string `json:"nik"`
		Nama   string `json:"nama"`
		Alamat string `json:"alamat"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "Format request body salah"})
	}

	validNIK, ve := utils.ValidateNIK(req.NIK)
	if ve != nil {
		return validationErrorResponse(c, ve)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	_, err := h.Queries.CreatePenerima(ctx, db.CreatePenerimaParams{
		Nik:    validNIK,
		Nama:   req.Nama,
		Alamat: req.Alamat,
	})
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(ErrorResponse{Code: "ERR_CONFLICT", Message: "Gagal menyimpan. NIK mungkin sudah terdaftar."})
	}

	h.invalidateBansosCache()
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{Pesan: "Data penerima berhasil ditambahkan"})
}

// =============================================================================
// ADMIN: PENYALURAN BANSOS
// =============================================================================

func (h *BansosHandler) AdminListPenyaluran(c fiber.Ctx) error {
	page, limit, offset := parsePagination(c)

	var progIDArg pgtype.Int4
	if p := c.Query("program_id"); p != "" {
		if pid, err := strconv.Atoi(p); err == nil {
			progIDArg = pgtype.Int4{Int32: int32(pid), Valid: true}
		}
	}

	var statusArg db.NullStatusPenyaluranType
	if s := c.Query("status"); s != "" {
		if validStatus, err := utils.ValidateStatusPenyaluran(s); err == nil {
			statusArg = db.NullStatusPenyaluranType{StatusPenyaluranType: db.StatusPenyaluranType(validStatus), Valid: true}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)
	var data []db.ListPenyaluranAdminRow
	var total int64

	g.Go(func() error {
		var err error
		data, err = h.Queries.ListPenyaluranAdmin(gCtx, db.ListPenyaluranAdminParams{
			ProgramID:  progIDArg,
			Status:     statusArg,
			LimitData:  int32(limit),
			OffsetData: int32(offset),
		})
		return err
	})
	g.Go(func() error {
		var err error
		total, err = h.Queries.CountPenyaluran(gCtx, db.CountPenyaluranParams{
			ProgramID: progIDArg,
			Status:    statusArg,
		})
		return err
	})

	if err := g.Wait(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_DB", Message: "Gagal memuat data penyaluran"})
	}

	resData := make([]AdminPenyaluranItem, 0, len(data))
	for _, r := range data {
		resData = append(resData, AdminPenyaluranItem{
			ID:             int(r.ID),
			PenerimaNIK:    r.PenerimaNik,
			PenerimaNama:   r.PenerimaNama,
			ProgramNama:    r.ProgramNama,
			Nominal:        int(r.Nominal),
			PeriodeMulai:   r.PeriodeMulai.Time.Format("2006-01-02"),
			PeriodeSelesai: r.PeriodeSelesai.Time.Format("2006-01-02"),
			Metode:         string(r.Metode),
			Status:         string(r.Status),
			CreatedAt:      r.CreatedAt.Time.Format("2006-01-02 15:04"),
		})
	}

	return c.JSON(PaginatedResponse{Pesan: "Data Penyaluran", Data: resData, Pagination: &PaginationMeta{Page: page, Limit: limit, Total: total}})
}

func (h *BansosHandler) AdminCreatePenyaluran(c fiber.Ctx) error {
	var req struct {
		PenerimaID     int    `json:"penerima_id"`
		ProgramID      int    `json:"program_id"`
		Nominal        string `json:"nominal"`
		PeriodeMulai   string `json:"periode_mulai"`
		PeriodeSelesai string `json:"periode_selesai"`
		Metode         string `json:"metode"`
		Status         string `json:"status"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "Format request body salah"})
	}

	// Validasi Input (Memanfaatkan utils yang sudah kita buat)
	nominal, vErr := utils.ValidateNominal(req.Nominal)
	tMulai, vErr2 := utils.ValidateDateString(req.PeriodeMulai, "periode_mulai")
	tSelesai, vErr3 := utils.ValidateDateString(req.PeriodeSelesai, "periode_selesai")
	metode, vErr4 := utils.ValidateMetodePenyaluran(req.Metode)
	status, vErr5 := utils.ValidateStatusPenyaluran(req.Status)

	if vErr != nil { return validationErrorResponse(c, vErr) }
	if vErr2 != nil { return validationErrorResponse(c, vErr2) }
	if vErr3 != nil { return validationErrorResponse(c, vErr3) }
	if vErr4 != nil { return validationErrorResponse(c, vErr4) }
	if vErr5 != nil { return validationErrorResponse(c, vErr5) }

	if rentangErr := utils.ValidatePeriode(tMulai, tSelesai); rentangErr != nil {
		return validationErrorResponse(c, rentangErr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	_, err := h.Queries.CreatePenyaluran(ctx, db.CreatePenyaluranParams{
		PenerimaID:     int32(req.PenerimaID),
		ProgramID:      int32(req.ProgramID),
		Nominal:        int32(nominal),
		PeriodeMulai:   pgtype.Date{Time: tMulai, Valid: true},
		PeriodeSelesai: pgtype.Date{Time: tSelesai, Valid: true},
		Metode:         db.MetodePenyaluranType(metode),
		Status:         db.StatusPenyaluranType(status),
	})
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(ErrorResponse{Code: "ERR_CONFLICT", Message: "Data penyaluran untuk periode ini sudah ada"})
	}

	h.invalidateBansosCache()
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{Pesan: "Data penyaluran bansos berhasil ditambahkan"})
}

// Update Status Cepat (Contoh: Mengubah "proses" jadi "diterima")
func (h *BansosHandler) AdminUpdateStatusPenyaluran(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "ID tidak valid"})
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "Format request body salah"})
	}

	status, vErr := utils.ValidateStatusPenyaluran(req.Status)
	if vErr != nil {
		return validationErrorResponse(c, vErr)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.Queries.UpdateStatusPenyaluran(ctx, db.UpdateStatusPenyaluranParams{
		Status: db.StatusPenyaluranType(status),
		ID:     int32(id),
	}); err != nil {
		slog.Error("admin.penyaluran.update_status_failed", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_DB", Message: "Gagal update status"})
	}

	h.invalidateBansosCache()
	return c.JSON(SuccessResponse{Pesan: "Status penyaluran berhasil diperbarui"})
}

func (h *BansosHandler) AdminDeletePenyaluran(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "ID tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.Queries.DeletePenyaluran(ctx, int32(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_DB", Message: "Gagal menghapus data"})
	}

	h.invalidateBansosCache()
	return c.JSON(SuccessResponse{Pesan: "Data penyaluran berhasil dihapus"})
}