package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/farildzaky/transjatim-service/internal/db"
	"github.com/farildzaky/transjatim-service/internal/notify"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/sync/errgroup"
)

func stringToPgTime(timeStr string) (pgtype.Time, error) {
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return pgtype.Time{}, err
	}
	micros := int64(t.Hour()*3600+t.Minute()*60) * 1_000_000
	return pgtype.Time{Microseconds: micros, Valid: true}, nil
}

// =====================================================================
// ADMIN: JADWAL
// =====================================================================

func (h *TransJatimHandler) GetAllJadwalsAdmin(c fiber.Ctx) error {
	page, limit, offset := parsePagination(c)
	cacheKey := normalizeKey("transjatim:admin:jadwal", c.Query("page"), c.Query("limit"))

	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)
	var data []db.ListJadwalAdminRow
	var total int64

	g.Go(func() error {
		var err error
		data, err = h.Queries.ListJadwalAdmin(gCtx, db.ListJadwalAdminParams{
			LimitData:  int32(limit),
			OffsetData: int32(offset),
		})
		return err
	})
	g.Go(func() error {
		var err error
		total, err = h.Queries.CountJadwal(gCtx)
		return err
	})

	if err := g.Wait(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code: "ERR_DB", Message: "Gagal mengambil data Jadwal",
		})
	}

	res := make([]JadwalAdminItem, 0, len(data))
	for _, j := range data {
		jamBerangkatStr := time.Unix(0, j.JamBerangkat.Microseconds*1000).UTC().Format("15:04")
		jamTibaStr := time.Unix(0, j.JamTiba.Microseconds*1000).UTC().Format("15:04")

		res = append(res, JadwalAdminItem{
			ID:             j.ID,
			BusKode:        j.BusKode,
			BusLayanan:     string(j.BusLayanan),
			TerminalAsal:   j.TerminalAsalNama,
			TerminalTujuan: j.TerminalTujuanNama,
			RuteSlug:       j.RuteSlug,
			JamBerangkat:   jamBerangkatStr,
			JamTiba:        jamTibaStr,
			HariOperasi:    j.HariOperasi,
			Aktif:          j.Aktif,
		})
	}

	return cacheJSON(c, cacheKey, CacheTTLList, SuccessResponse{
		Pesan: "Sukses", Data: res, Pagination: buildPaginationMeta(page, limit, total),
	})
}

func (h *TransJatimHandler) CreateJadwal(c fiber.Ctx) error {
	var req JadwalRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_JSON", Message: "Format data tidak valid"})
	}

	jamBerangkatPg, errJB := stringToPgTime(req.JamBerangkat)
	jamTibaPg, errJT := stringToPgTime(req.JamTiba)
	if errJB != nil || errJT != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "Format jam tidak valid (Gunakan HH:MM)"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	res, err := h.Queries.CreateJadwal(ctx, db.CreateJadwalParams{
		RuteID:       req.RuteID,
		BusID:        req.BusID,
		JamBerangkat: jamBerangkatPg,
		JamTiba:      jamTibaPg,
		HariOperasi:  req.HariOperasi,
	})

	if err != nil {
		slog.Error("admin.jadwal.create.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_INSERT", Message: "Gagal menyimpan jadwal"})
	}

	invalidateJadwalCache()
	body := fmt.Sprintf("Pukul %s - %s | Hari: %s", req.JamBerangkat, req.JamTiba, formatHariOperasi(req.HariOperasi))
	notify.ByService("/tiket-transjatim", "Jadwal Baru TransJatim", body, nil)
	notify.ByService("/rute-transjatim", "Jadwal Baru TransJatim", body, nil)
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Pesan: "Jadwal berhasil ditambahkan",
		Data: struct {
			ID int32 `json:"id"`
		}{ID: res},
	})
}

func (h *TransJatimHandler) UpdateJadwal(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "ID tidak valid"})
	}

	var req JadwalRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_JSON", Message: "Format tidak valid"})
	}

	jamBerangkatPg, errJB := stringToPgTime(req.JamBerangkat)
	jamTibaPg, errJT := stringToPgTime(req.JamTiba)
	if errJB != nil || errJT != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "Format jam tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	errDb := h.Queries.UpdateJadwal(ctx, db.UpdateJadwalParams{
		ID:           int32(id),
		RuteID:       req.RuteID,
		BusID:        req.BusID,
		JamBerangkat: jamBerangkatPg,
		JamTiba:      jamTibaPg,
		HariOperasi:  req.HariOperasi,
		Aktif:        req.Aktif,
	})

	if errDb != nil {
		slog.Error("admin.jadwal.update.error", slog.String("err", errDb.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_UPDATE", Message: "Gagal update jadwal"})
	}

	invalidateJadwalCache()
	bodyUpd := fmt.Sprintf("Jam diperbarui: %s - %s | Hari: %s", req.JamBerangkat, req.JamTiba, formatHariOperasi(req.HariOperasi))
	notify.ByService("/tiket-transjatim", "Perubahan Jadwal TransJatim", bodyUpd, nil)
	notify.ByService("/rute-transjatim", "Perubahan Jadwal TransJatim", bodyUpd, nil)
	return c.JSON(SuccessResponse{Pesan: "Jadwal berhasil diperbarui"})
}

func (h *TransJatimHandler) DeleteJadwal(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "ID tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.Queries.DeleteJadwal(ctx, int32(id)); err != nil {
		slog.Error("admin.jadwal.delete.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_DELETE", Message: "Gagal menghapus jadwal"})
	}

	invalidateJadwalCache()
	return c.JSON(SuccessResponse{Pesan: "Jadwal berhasil dihapus"})
}