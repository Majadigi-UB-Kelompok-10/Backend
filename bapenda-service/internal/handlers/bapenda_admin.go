package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/farildzaky/bapenda-service/internal/cache"
	"github.com/farildzaky/bapenda-service/internal/db"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/sync/errgroup"
)

// --- ADMIN PAJAK KENDARAAN ---

func (h *BapendaHandler) GetAllInfoPajakAdmin(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	var data []db.GetAllKendaraanPajakAdminRow
	var total int64

	g.Go(func() error {
		var err error
		data, err = h.Queries.GetAllKendaraanPajakAdmin(ctx, db.GetAllKendaraanPajakAdminParams{LimitData: int32(limit), OffsetData: int32(offset)})
		return err
	})
	g.Go(func() error {
		var err error
		total, err = h.Queries.CountKendaraanPajak(ctx)
		return err
	})

	if err := g.Wait(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_DB_QUERY", Message: "Gagal memproses data"})
	}

	return c.JSON(SuccessResponse{Pesan: "Sukses", Data: data, Pagination: PaginationMeta{Page: page, Limit: limit, Total: total}})
}

func (h *BapendaHandler) GetDetailPajakAdmin(c fiber.Ctx) error {
	plat := strings.ToUpper(strings.ReplaceAll(c.Params("plat"), " ", ""))
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetKendaraanPajakByPlatAdmin(ctx, plat)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Code: "ERR_NOT_FOUND", Message: "Data tidak ditemukan"})
	}
	return c.JSON(SuccessResponse{Pesan: "Sukses", Data: data})
}

func (h *BapendaHandler) CreateInfoPajak(c fiber.Ctx) error {
	var req KendaraanRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_INVALID_JSON", Message: "Format data tidak valid"})
	}

	t, _ := time.Parse("2006-01-02", req.MasaPajak)
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	arg := db.CreateKendaraanPajakParams{
		PlatNomor: req.PlatNomor, PlatNomorDisplay: req.PlatNomorDisplay, NomorRangka: req.NomorRangka,
		StatusAktif: req.StatusAktif, Merk: req.Merk, Tipe: req.Tipe, Warna: req.Warna, TahunBuat: req.TahunBuat,
		Model: req.Model, MasaPajak: pgtype.Date{Time: t, Valid: true}, PkbPokok: req.PkbPokok,
		OpsenPkb: req.OpsenPkb, Swdkllj: req.Swdkllj, ParkirBerlangganan: req.ParkirBerlangganan,
		CetakStnk: req.CetakStnk, CetakTnkb: req.CetakTnkb,
	}

	res, err := h.Queries.CreateKendaraanPajak(ctx, arg)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(ErrorResponse{Code: "ERR_CONFLICT", Message: "Gagal simpan data, mungkin plat duplikat"})
	}
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{Pesan: "Berhasil menambahkan kendaraan", Data: res})
}

func (h *BapendaHandler) UpdateInfoPajak(c fiber.Ctx) error {
	platKey := c.Params("plat")
	var req KendaraanRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_INVALID_JSON", Message: "Format data tidak valid"})
	}

	t, _ := time.Parse("2006-01-02", req.MasaPajak)
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	arg := db.UpdateKendaraanPajakParams{
		PlatNomorKey: platKey, NomorRangka: req.NomorRangka, StatusAktif: req.StatusAktif,
		Merk: req.Merk, Tipe: req.Tipe, Warna: req.Warna, TahunBuat: req.TahunBuat,
		Model: req.Model, MasaPajak: pgtype.Date{Time: t, Valid: true}, PkbPokok: req.PkbPokok,
		OpsenPkb: req.OpsenPkb, Swdkllj: req.Swdkllj, ParkirBerlangganan: req.ParkirBerlangganan,
		CetakStnk: req.CetakStnk, CetakTnkb: req.CetakTnkb,
	}

	if err := h.Queries.UpdateKendaraanPajak(ctx, arg); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_UPDATE_FAILED", Message: "Gagal mengupdate data"})
	}
	cache.GlobalCache.DeleteByPrefix("pajak:info:" + strings.ToUpper(strings.ReplaceAll(platKey, " ", "")))
	return c.JSON(SuccessResponse{Pesan: "Data kendaraan terupdate"})
}

func (h *BapendaHandler) DeleteInfoPajak(c fiber.Ctx) error {
	plat := strings.ToUpper(strings.ReplaceAll(c.Params("plat"), " ", ""))
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.Queries.DeleteKendaraanPajak(ctx, plat); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_DELETE_FAILED", Message: "Gagal menghapus data"})
	}
	cache.GlobalCache.DeleteByPrefix("pajak:info:" + plat)
	return c.JSON(SuccessResponse{Pesan: "Data kendaraan terhapus"})
}

// --- ADMIN MASTER NJKB ---

func (h *BapendaHandler) GetAllMasterNjkbAdmin(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	var data []db.MasterNjkb
	var total int64

	g.Go(func() error {
		var err error
		data, err = h.Queries.GetAllMasterNjkbAdmin(ctx, db.GetAllMasterNjkbAdminParams{LimitData: int32(limit), OffsetData: int32(offset)})
		return err
	})
	g.Go(func() error {
		var err error
		total, err = h.Queries.CountMasterNjkb(ctx)
		return err
	})

	if err := g.Wait(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_DB_QUERY", Message: "Error mengambil master NJKB"})
	}

	var responseData []MasterNjkbSummaryResponse
	for _, row := range data {
		responseData = append(responseData, MasterNjkbSummaryResponse{
			ID: row.ID, NamaKendaraan: fmt.Sprintf("%s %s %s", row.Merk, row.Model, row.Tipe),
			JenisKendaraan: row.JenisKendaraan, Merk: row.Merk, Model: row.Model,
		})
	}
	if responseData == nil {
		responseData = []MasterNjkbSummaryResponse{}
	}

	return c.JSON(SuccessResponse{Pesan: "Sukses", Data: responseData, Pagination: PaginationMeta{Page: page, Limit: limit, Total: total}})
}

func (h *BapendaHandler) GetDetailMasterNjkbAdmin(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "Format ID salah"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetMasterNjkbById(ctx, int32(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Code: "ERR_NOT_FOUND", Message: "Data Master NJKB tidak ditemukan"})
	}
	return c.JSON(SuccessResponse{Pesan: "Sukses", Data: data})
}

func (h *BapendaHandler) CreateMasterNjkb(c fiber.Ctx) error {
	var req MasterNjkbRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_INVALID_JSON", Message: "Format JSON tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	arg := db.CreateMasterNjkbParams{JenisKendaraan: req.Jenis, Merk: req.Merk, Model: req.Model, Tipe: req.Tipe, Tahun: req.Tahun, NilaiJual: req.Nilai}
	res, err := h.Queries.CreateMasterNjkb(ctx, arg)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(ErrorResponse{Code: "ERR_CONFLICT", Message: "Master NJKB sudah terdaftar"})
	}
	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{Pesan: "Master Terbuat", Data: map[string]interface{}{"id": res}})
}

func (h *BapendaHandler) UpdateMasterNjkb(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "ID harus angka"})
	}

	var req MasterNjkbRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_INVALID_JSON", Message: "Format JSON tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	arg := db.UpdateMasterNjkbParams{ID: int32(id), JenisKendaraan: req.Jenis, Merk: req.Merk, Model: req.Model, Tipe: req.Tipe, Tahun: req.Tahun, NilaiJual: req.Nilai}
	if err := h.Queries.UpdateMasterNjkb(ctx, arg); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_UPDATE", Message: "Gagal memperbarui master NJKB"})
	}
	return c.JSON(SuccessResponse{Pesan: "Master Terupdate"})
}

func (h *BapendaHandler) DeleteMasterNjkb(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "ID harus angka"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.Queries.DeleteMasterNjkb(ctx, int32(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_DELETE", Message: "Gagal menghapus data master"})
	}
	return c.JSON(SuccessResponse{Pesan: "Master Terhapus"})
}