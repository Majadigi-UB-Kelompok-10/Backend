package handlers

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/sync/errgroup"

	"github.com/farildzaky/sinaker-service/internal/db"
	"github.com/farildzaky/sinaker-service/internal/notify"
	"github.com/farildzaky/sinaker-service/internal/utils"
)



// =====================================================================
// ADMIN: KELOLA BLK
// =====================================================================

func (h *SinakerHandler) AdminGetListBLK(c fiber.Ctx) error {
	page, limit, offset := utils.ValidatePaginationParams(c.Query("page"), c.Query("limit"))

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)
	var blks []db.ListAllBLKAdminRow 
	var total int64

	g.Go(func() error {
		var err error
		blks, err = h.Queries.ListAllBLKAdmin(gCtx, db.ListAllBLKAdminParams{
			LimitData:  int32(limit),
			OffsetData: int32(offset),
		})
		return err
	})
	g.Go(func() error {
		var err error
		total, err = h.Queries.CountAllBLK(gCtx)
		return err
	})

	if err := g.Wait(); err != nil {
		slog.Error("admin.blk.list_error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal memuat data BLK admin",
		})
	}

	return c.Status(fiber.StatusOK).JSON(PaginatedResponse{
		Pesan: "Berhasil mengambil data BLK",
		Data:  blks,
		Pagination: &PaginationMeta{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	})
}

func (h *SinakerHandler) AdminCreateBLK(c fiber.Ctx) error {
	var req BLKRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_JSON",
			Message: "Format data tidak valid",
		})
	}

	nama, errVal := utils.ValidateTextContent(req.Nama, 200, "Nama BLK")
	if errVal != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: errVal.Message})
	}

	slug := utils.GenerateSlug(nama)

	var lat, lng pgtype.Float8
	if req.Lat != nil {
		lat = pgtype.Float8{Float64: *req.Lat, Valid: true}
	}
	if req.Lng != nil {
		lng = pgtype.Float8{Float64: *req.Lng, Valid: true}
	}

	var kec pgtype.Text
	if req.Kecamatan != nil {
		kec = pgtype.Text{String: *req.Kecamatan, Valid: true}
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	res, err := h.Queries.CreateBLK(ctx, db.CreateBLKParams{
		Nama:      nama,
		Alamat:    req.Alamat,
		KabKota:   req.KabKota,
		Kecamatan: kec,
		Slug:      slug,
		Lat:       lat,
		Lng:       lng,
	})

	if err != nil {
		slog.Error("admin.blk.create_error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal membuat BLK baru (mungkin nama sudah ada)",
		})
	}

	h.invalidatePublicCache()
	notify.ByService("/balai-latihan-kerja", "BLK Baru Tersedia", "BLK "+nama+" di "+req.KabKota+" telah dibuka. Daftar sekarang!", nil)

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Pesan: "Berhasil membuat BLK baru",
		Data:  res,
	})
}

// =====================================================================
// ADMIN: KELOLA PENDAFTARAN & STATUS
// =====================================================================

func (h *SinakerHandler) AdminGetListPendaftaran(c fiber.Ctx) error {
	page, limit, offset := utils.ValidatePaginationParams(c.Query("page"), c.Query("limit"))

	blkIDFilter := c.Query("blk_id")
	statusFilterStr := c.Query("status")

	var pgBlkID pgtype.Int4
	if blkIDFilter != "" {
		if id, err := strconv.Atoi(blkIDFilter); err == nil {
			pgBlkID = pgtype.Int4{Int32: int32(id), Valid: true}
		}
	}

	var pgStatus db.NullStatusDaftarType
	if statusFilterStr != "" {
		if val, err := utils.ValidateStatusPendaftaran(statusFilterStr); err == nil {
			pgStatus = db.NullStatusDaftarType{StatusDaftarType: db.StatusDaftarType(val), Valid: true}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)
	var pendaftarans []db.ListPendaftaranAdminRow 
	var total int64

	g.Go(func() error {
		var err error
		pendaftarans, err = h.Queries.ListPendaftaranAdmin(gCtx, db.ListPendaftaranAdminParams{
			BlkIDFilter:  pgBlkID, 
			StatusFilter: pgStatus,
			LimitData:    int32(limit),
			OffsetData:   int32(offset),
		})
		return err
	})
	g.Go(func() error {
		var err error
		total, err = h.Queries.CountPendaftaran(gCtx, db.CountPendaftaranParams{
			BlkIDFilter:  pgBlkID, // SUDAH DIUBAH JADI ID KAPITAL
			StatusFilter: pgStatus,
		})
		return err
	})

	if err := g.Wait(); err != nil {
		slog.Error("admin.pendaftaran.list_error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal memuat data pendaftaran",
		})
	}

	resData := make([]PendaftaranAdminResponse, 0, len(pendaftarans))
	for _, p := range pendaftarans {
		resData = append(resData, PendaftaranAdminResponse{
			ID:            p.ID,
			NIK:           p.Nik,
			NamaLengkap:   p.NamaLengkap,
			NoWA:          p.NoWa,
			Status:        string(p.Status),
			BlkNama:       p.BlkNama,
			KejuruanNama:  p.KejuruanNama,
			TanggalDaftar: p.CreatedAt.Time.Format("2006-01-02 15:04 WIB"),
		})
	}

	return c.Status(fiber.StatusOK).JSON(PaginatedResponse{
		Pesan: "Data Pendaftaran",
		Data:  resData,
		Pagination: &PaginationMeta{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	})
}

func (h *SinakerHandler) AdminUpdateStatusPendaftaran(c fiber.Ctx) error {
	pendaftaranIDStr := c.Params("id")
	pendaftaranID, err := strconv.Atoi(pendaftaranIDStr)
	if err != nil || pendaftaranID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "ID Pendaftaran tidak valid",
		})
	}

	var req UpdateStatusRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_JSON",
			Message: "Format data tidak valid",
		})
	}

	statusValid, errVal := utils.ValidateStatusPendaftaran(req.Status)
	if errVal != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: errVal.Message})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	errUpdate := h.Queries.UpdateStatusPendaftaran(ctx, db.UpdateStatusPendaftaranParams{
		Status: db.StatusDaftarType(statusValid),
		ID:     int32(pendaftaranID),
	})

	if errUpdate != nil {
		slog.Error("admin.pendaftaran.update_status_error", slog.String("err", errUpdate.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal memperbarui status pendaftaran",
		})
	}

	detail, errGet := h.Queries.GetPendaftaranByID(ctx, int32(pendaftaranID))
	if errGet == nil {
		sendEmailStatusAsync(
			detail.Email,
			detail.NamaLengkap,
			detail.BlkNama,
			detail.KejuruanNama,
			statusValid,
		)
	} else {
		slog.Warn("admin.pendaftaran.email_skip", slog.String("note", "Gagal mengambil detail untuk email"))
	}

	slog.Info("admin.pendaftaran.status_updated",
		slog.String("pendaftaran_id", pendaftaranIDStr),
		slog.String("status", statusValid),
	)

	return c.Status(fiber.StatusOK).JSON(SuccessResponse{
		Pesan: "Status pendaftaran berhasil diperbarui. Email notifikasi sedang dikirim ke peserta.",
	})
}

// =============================================================================
// CACHE INVALIDATION — Centralize biar konsisten
// =============================================================================

func (h *SinakerHandler) invalidatePublicCache() {
	h.Cache.Delete("sinaker:public:blk:kota")
	h.Cache.DeleteByPrefix("sinaker:public:blk:list")
}