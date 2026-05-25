package handlers

import (
    "context"
    "fmt"
    "log/slog"
    "strconv"
    "strings"
    "time"

    "github.com/farildzaky/bapenda-service/internal/cache"
    "github.com/farildzaky/bapenda-service/internal/db"
    "github.com/farildzaky/bapenda-service/internal/notify"
    "github.com/gofiber/fiber/v3"
    "github.com/jackc/pgx/v5/pgtype"
    "golang.org/x/sync/errgroup"
)

// =============================================================================
// ADMIN: KENDARAAN PAJAK
// =============================================================================

func (h *BapendaHandler) GetAllInfoPajakAdmin(c fiber.Ctx) error {
    page, limit, offset := parsePagination(c)

    ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
    defer cancel()

    g, gCtx := errgroup.WithContext(ctx)
    var data []db.GetAllKendaraanPajakAdminRow
    var total int64

    g.Go(func() error {
        var err error
        data, err = h.Queries.GetAllKendaraanPajakAdmin(gCtx, db.GetAllKendaraanPajakAdminParams{
            LimitData:  int32(limit),
            OffsetData: int32(offset),
        })
        return err
    })
    g.Go(func() error {
        var err error
        total, err = h.Queries.CountKendaraanPajak(gCtx)
        return err
    })

    if err := g.Wait(); err != nil {
        slog.Error("admin.pajak.list.error", slog.String("err", err.Error()))
        return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
            Code: "ERR_DB_QUERY", Message: "Gagal memproses data",
        })
    }

    if data == nil {
        data = []db.GetAllKendaraanPajakAdminRow{}
    }

    return c.JSON(SuccessResponse{
        Pesan: "Sukses",
        Data:  data,
        Pagination: PaginationMeta{
            Page: page, Limit: limit, Total: total,
        },
    })
}

func (h *BapendaHandler) GetDetailPajakAdmin(c fiber.Ctx) error {
    idParam := c.Params("id") 
    var vehicleID pgtype.UUID
    if err := vehicleID.Scan(idParam); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
            Code: "ERR_VALIDATION", Message: "Format ID tidak valid (harus UUID)",
        })
    }

    ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
    defer cancel()

    data, err := h.Queries.GetKendaraanPajakByIdAdmin(ctx, vehicleID) 
    if err != nil {
        return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
            Code: "ERR_NOT_FOUND", Message: "Data tidak ditemukan",
        })
    }
    return c.JSON(SuccessResponse{Pesan: "Sukses", Data: data})
}

func (h *BapendaHandler) CreateInfoPajak(c fiber.Ctx) error {
    var req KendaraanRequest
    if err := c.Bind().JSON(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
            Code: "ERR_INVALID_JSON", Message: "Format data tidak valid",
        })
    }

    if err := requireFields(c, map[string]string{
        "plat_nomor":   req.PlatNomor,
        "nomor_rangka": req.NomorRangka,
        "merk":         req.Merk,
        "masa_pajak":   req.MasaPajak,
    }); err != nil {
        return err
    }

    t, err := time.Parse("2006-01-02", req.MasaPajak)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
            Code: "ERR_VALIDATION", Message: "Format masa_pajak harus YYYY-MM-DD",
        })
    }

    ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
    defer cancel()

    res, err := h.Queries.CreateKendaraanPajak(ctx, db.CreateKendaraanPajakParams{
        PlatNomor: req.PlatNomor, PlatNomorDisplay: req.PlatNomorDisplay, NomorRangka: req.NomorRangka,
        StatusAktif: req.StatusAktif, Merk: req.Merk, Tipe: req.Tipe, Warna: req.Warna,
        TahunBuat: req.TahunBuat, Model: req.Model,
        MasaPajak: pgtype.Date{Time: t, Valid: true},
        PkbPokok:  req.PkbPokok, OpsenPkb: req.OpsenPkb, Swdkllj: req.Swdkllj,
        ParkirBerlangganan: req.ParkirBerlangganan,
        CetakStnk:          req.CetakStnk, CetakTnkb: req.CetakTnkb,
    })
    if err != nil {
        slog.Warn("admin.pajak.create.conflict",
            slog.String("plat", req.PlatNomor),
            slog.String("err", err.Error()),
        )
        return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
            Code: "ERR_CONFLICT", Message: "Gagal simpan data, mungkin plat duplikat",
        })
    }

    invalidatePajakCache(req.PlatNomor)
    notify.ByService("/pajak-kendaraan", "Info Pajak: "+req.Merk+" "+req.Tipe,
        "Data pajak kendaraan "+req.PlatNomor+" ("+req.Merk+" "+req.Tipe+") telah ditambahkan", nil)

    return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
        Data: struct {
            ID string `json:"id"`
        }{
            ID: res.String(),
        },
    })
}

func (h *BapendaHandler) UpdateInfoPajak(c fiber.Ctx) error {
    idParam := c.Params("id") // SEKARANG MENGGUNAKAN ID
    var vehicleID pgtype.UUID
    if err := vehicleID.Scan(idParam); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
            Code: "ERR_VALIDATION", Message: "Format ID tidak valid",
        })
    }

    var req KendaraanRequest
    if err := c.Bind().JSON(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
            Code: "ERR_INVALID_JSON", Message: "Format data tidak valid",
        })
    }

    t, err := time.Parse("2006-01-02", req.MasaPajak)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
            Code: "ERR_VALIDATION", Message: "Format masa_pajak harus YYYY-MM-DD",
        })
    }

    ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
    defer cancel()

    oldData, err := h.Queries.GetKendaraanPajakByIdAdmin(ctx, vehicleID)
    if err != nil {
        return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
            Code: "ERR_NOT_FOUND", Message: "Data kendaraan tidak ditemukan",
        })
    }

    if err := h.Queries.UpdateKendaraanPajak(ctx, db.UpdateKendaraanPajakParams{
        ID:                 vehicleID, 
        PlatNomor:          req.PlatNomor,
        PlatNomorDisplay:   req.PlatNomorDisplay,
        NomorRangka:        req.NomorRangka, 
        StatusAktif:        req.StatusAktif,
        Merk:               req.Merk, 
        Tipe:               req.Tipe, 
        Warna:              req.Warna, 
        TahunBuat:          req.TahunBuat,
        Model:              req.Model,
        MasaPajak:          pgtype.Date{Time: t, Valid: true},
        PkbPokok:           req.PkbPokok, 
        OpsenPkb:           req.OpsenPkb, 
        Swdkllj:            req.Swdkllj,
        ParkirBerlangganan: req.ParkirBerlangganan,
        CetakStnk:          req.CetakStnk, 
        CetakTnkb:          req.CetakTnkb,
    }); err != nil {
        slog.Error("admin.pajak.update.error",
            slog.String("id", idParam),
            slog.String("err", err.Error()),
        )
        return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
            Code: "ERR_UPDATE_FAILED", Message: "Gagal mengupdate data",
        })
    }

    invalidatePajakCache(oldData.PlatNomor)
    if strings.ToUpper(strings.ReplaceAll(oldData.PlatNomor, " ", "")) != strings.ToUpper(strings.ReplaceAll(req.PlatNomor, " ", "")) {
        invalidatePajakCache(req.PlatNomor)
    }
    notify.ByService("/pajak-kendaraan", "Update Pajak: "+req.Merk+" "+req.Tipe,
        "Data pajak kendaraan "+req.PlatNomor+" ("+req.Merk+" "+req.Tipe+") telah diperbarui", nil)

    return c.JSON(SuccessResponse{Pesan: "Data kendaraan terupdate"})
}

func (h *BapendaHandler) DeleteInfoPajak(c fiber.Ctx) error {
    idParam := c.Params("id") 
    var vehicleID pgtype.UUID
    if err := vehicleID.Scan(idParam); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
            Code: "ERR_VALIDATION", Message: "Format ID tidak valid",
        })
    }

    ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
    defer cancel()

    oldData, err := h.Queries.GetKendaraanPajakByIdAdmin(ctx, vehicleID)
    if err != nil {
        return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
            Code: "ERR_NOT_FOUND", Message: "Data tidak ditemukan",
        })
    }

    if err := h.Queries.DeleteKendaraanPajak(ctx, vehicleID); err != nil {
        slog.Error("admin.pajak.delete.error",
            slog.String("id", idParam),
            slog.String("err", err.Error()),
        )
        return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
            Code: "ERR_DELETE_FAILED", Message: "Gagal menghapus data",
        })
    }

    invalidatePajakCache(oldData.PlatNomor)

    return c.JSON(SuccessResponse{Pesan: "Data kendaraan terhapus"})
}

// =============================================================================
// ADMIN: MASTER NJKB
// =============================================================================

func (h *BapendaHandler) GetAllMasterNjkbAdmin(c fiber.Ctx) error {
    page, limit, offset := parsePagination(c)

    ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
    defer cancel()

    g, gCtx := errgroup.WithContext(ctx)
    var data []db.MasterNjkb
    var total int64

    g.Go(func() error {
        var err error
        data, err = h.Queries.GetAllMasterNjkbAdmin(gCtx, db.GetAllMasterNjkbAdminParams{
            LimitData:  int32(limit),
            OffsetData: int32(offset),
        })
        return err
    })
    g.Go(func() error {
        var err error
        total, err = h.Queries.CountMasterNjkb(gCtx)
        return err
    })

    if err := g.Wait(); err != nil {
        slog.Error("admin.njkb.list.error", slog.String("err", err.Error()))
        return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
            Code: "ERR_DB_QUERY", Message: "Error mengambil master NJKB",
        })
    }

    responseData := make([]MasterNjkbSummaryResponse, 0, len(data))
    for _, row := range data {
        responseData = append(responseData, MasterNjkbSummaryResponse{
            ID:             row.ID,
            NamaKendaraan:  fmt.Sprintf("%s %s %s", row.Merk, row.Model, row.Tipe),
            JenisKendaraan: row.JenisKendaraan,
            Merk:           row.Merk,
            Model:          row.Model,
        })
    }

    return c.JSON(SuccessResponse{
        Pesan: "Sukses",
        Data:  responseData,
        Pagination: PaginationMeta{
            Page: page, Limit: limit, Total: total,
        },
    })
}

func (h *BapendaHandler) GetDetailMasterNjkbAdmin(c fiber.Ctx) error {
    id, err := strconv.Atoi(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
            Code: "ERR_VALIDATION", Message: "Format ID salah",
        })
    }

    ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
    defer cancel()

    data, err := h.Queries.GetMasterNjkbById(ctx, int32(id))
    if err != nil {
        return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
            Code: "ERR_NOT_FOUND", Message: "Data master NJKB tidak ditemukan",
        })
    }
    return c.JSON(SuccessResponse{Pesan: "Sukses", Data: data})
}

func (h *BapendaHandler) CreateMasterNjkb(c fiber.Ctx) error {
    var req MasterNjkbRequest
    if err := c.Bind().JSON(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
            Code: "ERR_INVALID_JSON", Message: "Format JSON tidak valid",
        })
    }

    if err := requireFields(c, map[string]string{
        "jenis_kendaraan": req.Jenis, "merk": req.Merk, "model": req.Model, "tipe": req.Tipe,
    }); err != nil {
        return err
    }

    if req.Nilai <= 0 {
        return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
            Code: "ERR_VALIDATION", Message: "nilai_jual harus lebih dari 0",
        })
    }

    ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
    defer cancel()

    res, err := h.Queries.CreateMasterNjkb(ctx, db.CreateMasterNjkbParams{
        JenisKendaraan: req.Jenis, Merk: req.Merk, Model: req.Model,
        Tipe: req.Tipe, Tahun: req.Tahun, NilaiJual: req.Nilai,
    })
    if err != nil {
        return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
            Code: "ERR_CONFLICT", Message: "Master NJKB sudah terdaftar",
        })
    }

    invalidateNjkbCache()

    return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
        Pesan: "Master terbuat",
        Data: struct {
            ID int32 `json:"id"`
        }{
            ID: res,
        },
    })
}

func (h *BapendaHandler) UpdateMasterNjkb(c fiber.Ctx) error {
    id, err := strconv.Atoi(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
            Code: "ERR_VALIDATION", Message: "ID harus angka",
        })
    }

    var req MasterNjkbRequest
    if err := c.Bind().JSON(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
            Code: "ERR_INVALID_JSON", Message: "Format JSON tidak valid",
        })
    }

    ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
    defer cancel()

    if err := h.Queries.UpdateMasterNjkb(ctx, db.UpdateMasterNjkbParams{
        ID: int32(id), JenisKendaraan: req.Jenis, Merk: req.Merk,
        Model: req.Model, Tipe: req.Tipe, Tahun: req.Tahun, NilaiJual: req.Nilai,
    }); err != nil {
        slog.Error("admin.njkb.update.error",
            slog.Int("id", id),
            slog.String("err", err.Error()),
        )
        return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
            Code: "ERR_UPDATE", Message: "Gagal memperbarui master NJKB",
        })
    }

    invalidateNjkbCache()

    return c.JSON(SuccessResponse{Pesan: "Master terupdate"})
}

func (h *BapendaHandler) DeleteMasterNjkb(c fiber.Ctx) error {
    id, err := strconv.Atoi(c.Params("id"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
            Code: "ERR_VALIDATION", Message: "ID harus angka",
        })
    }

    ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
    defer cancel()

    if err := h.Queries.DeleteMasterNjkb(ctx, int32(id)); err != nil {
        slog.Error("admin.njkb.delete.error",
            slog.Int("id", id),
            slog.String("err", err.Error()),
        )
        return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
            Code: "ERR_DELETE", Message: "Gagal menghapus data master",
        })
    }

    invalidateNjkbCache()

    return c.JSON(SuccessResponse{Pesan: "Master terhapus"})
}

func invalidatePajakCache(plat string) {
    plat = strings.ToUpper(strings.ReplaceAll(plat, " ", ""))
    cache.GlobalCache.DeleteByPrefix("pajak:info:" + plat)
}

func invalidateNjkbCache() {
    cache.GlobalCache.DeleteByPrefix("dropdown:")
    cache.GlobalCache.DeleteByPrefix("kalkulasi:")
}