package handlers

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"log"
	"time"

	"github.com/farildzaky/bapenda-service/internal/cache"
	"github.com/farildzaky/bapenda-service/internal/db"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/sync/errgroup"
)

const ContextQueryTimeout = 5 * time.Second

func (h *BapendaHandler) RunCacheWarmup() {
	fmt.Println("Memulai Cache Warmup...")
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		jenisData, err := h.Queries.GetDistinctJenis(ctx)
		if err != nil {
			return fmt.Errorf("gagal ambil jenis kendaraan: %v", err)
		}
		
		res := SuccessResponse{Pesan: "Sukses", Data: jenisData}
		cache.GlobalCache.Set("dropdown:jenis", res)
		return nil
	})

	g.Go(func() error {
		tarifData, err := h.Queries.GetAllTarifPKB(ctx)
		if err != nil {
			return fmt.Errorf("gagal ambil tarif PKB: %v", err)
		}
		
		cache.GlobalCache.Set("master:tarif_pkb", tarifData)
		return nil
	})

	if err := g.Wait(); err != nil {
		log.Printf("Peringatan: Cache Warmup selesai dengan error (Aplikasi tetap berjalan): %v\n", err)
		return
	}

	fmt.Printf("Cache Warmup Berhasil (Waktu eksekusi: %v)\n", time.Since(startTime))
}

// =====================================================================
// HANDLER SETUP
// =====================================================================

type BapendaHandler struct {
	Queries *db.Queries
}

func NewBapendaHandler(q *db.Queries) *BapendaHandler {
	return &BapendaHandler{Queries: q}
}

// ---------------------------------------------------------------------
// 1. USER API: INFO PAJAK (SECURITY & PRIVACY FOCUS)
// ---------------------------------------------------------------------

func (h *BapendaHandler) GetInfoPajak(c fiber.Ctx) error {
	platNomor := strings.ToUpper(strings.ReplaceAll(c.FormValue("plat_nomor"), " ", ""))
	nomorRangka := strings.ToUpper(strings.TrimSpace(c.FormValue("nomor_rangka")))

	if platNomor == "" || len(nomorRangka) != 5 {
		return c.Status(400).JSON(ErrorResponse{
			Error:  "Input Tidak Valid",
			Detail: "Plat nomor wajib diisi dan nomor rangka harus 5 digit terakhir",
		})
	}

	cacheKey := fmt.Sprintf("pajak:info:%s:%s", platNomor, nomorRangka)
	if cachedBytes, ok := cache.GlobalCache.Get(cacheKey); ok {
		c.Set("Content-Type", "application/json")
		c.Set("X-Cache", "HIT")
		return c.Send(cachedBytes)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	arg := db.GetKendaraanByPlatDanRangkaParams{
		PlatNomor:       platNomor,
		LimaDigitRangka: nomorRangka,
	}

	data, err := h.Queries.GetKendaraanByPlatDanRangka(ctx, arg)
	if err != nil {
		return c.Status(404).JSON(ErrorResponse{
			Error:  "Data Tidak Ditemukan",
			Detail: "Kendaraan tidak terdaftar atau kombinasi rangka salah",
		})
	}

	res := SuccessResponse{
		Pesan: "Data Ditemukan",
		Data: InfoPajakResponse{
			Identitas: IdentitasKendaraan{
				PlatNomor:   data.PlatNomorDisplay,
				Merk:        data.Merk,
				Tipe:        data.Tipe,
				Warna:       data.Warna,
				TahunBuat:   data.TahunBuat,
				MasaPajak:   data.MasaPajak.Time.Format("2006-01-02"),
				StatusAktif: data.StatusAktif,
			},
			RincianBiaya: RincianBiayaPajak{
				PkbPokok:           data.PkbPokok,
				OpsenPkb:           data.OpsenPkb,
				Swdkllj:            data.Swdkllj,
				ParkirBerlangganan: data.ParkirBerlangganan,
				TotalPajak:         int64(data.TotalPajakTahunan.Int32),
			},
			Estimasi5Tahunan: Estimasi5Tahunan{
				CetakStnk: data.CetakStnk,
				CetakTnkb: data.CetakTnkb,
			},
		},
	}

	cache.GlobalCache.Set(cacheKey, res)
	return c.JSON(res)
}

// ---------------------------------------------------------------------
// 2. USER API: DROPDOWN & KALKULATOR NJKB (CASCADE)
// ---------------------------------------------------------------------

func (h *BapendaHandler) GetDropdownJenis(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetDistinctJenis(ctx)
	if err != nil {
		return c.Status(500).JSON(ErrorResponse{Error: "Gagal memuat jenis kendaraan", Detail: err.Error()})
	}
	return c.JSON(SuccessResponse{Pesan: "Sukses", Data: data})
}

func (h *BapendaHandler) GetDropdownMerk(c fiber.Ctx) error {
	jenis := c.Query("jenis")
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetDistinctMerk(ctx, jenis)
	if err != nil {
		return c.Status(500).JSON(ErrorResponse{Error: "Gagal memuat merk", Detail: err.Error()})
	}
	return c.JSON(SuccessResponse{Pesan: "Sukses", Data: data})
}

func (h *BapendaHandler) GetDropdownModel(c fiber.Ctx) error {
	arg := db.GetDistinctModelParams{JenisKendaraan: c.Query("jenis"), Merk: c.Query("merk")}
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetDistinctModel(ctx, arg)
	if err != nil {
		return c.Status(500).JSON(ErrorResponse{Error: "Gagal memuat model", Detail: err.Error()})
	}
	return c.JSON(SuccessResponse{Pesan: "Sukses", Data: data})
}

func (h *BapendaHandler) GetDropdownTipe(c fiber.Ctx) error {
	arg := db.GetDistinctTipeParams{JenisKendaraan: c.Query("jenis"), Merk: c.Query("merk"), Model: c.Query("model")}
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetDistinctTipe(ctx, arg)
	if err != nil {
		return c.Status(500).JSON(ErrorResponse{Error: "Gagal memuat tipe", Detail: err.Error()})
	}
	return c.JSON(SuccessResponse{Pesan: "Sukses", Data: data})
}

func (h *BapendaHandler) GetDropdownTahun(c fiber.Ctx) error {
	arg := db.GetDistinctTahunParams{JenisKendaraan: c.Query("jenis"), Merk: c.Query("merk"), Model: c.Query("model"), Tipe: c.Query("tipe")}
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetDistinctTahun(ctx, arg)
	if err != nil {
		return c.Status(500).JSON(ErrorResponse{Error: "Gagal memuat tahun", Detail: err.Error()})
	}
	return c.JSON(SuccessResponse{Pesan: "Sukses", Data: data})
}

func (h *BapendaHandler) HitungKalkulasiNJKB(c fiber.Ctx) error {
	var req struct {
		Jenis string `json:"jenis"`
		Merk  string `json:"merk"`
		Model string `json:"model"`
		Tipe  string `json:"tipe"`
		Tahun int16  `json:"tahun"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{Error: "JSON tidak valid"})
	}

	cacheKey := fmt.Sprintf("kalkulasi:%s:%s:%s:%s:%d", req.Jenis, req.Merk, req.Model, req.Tipe, req.Tahun)
	if cached, ok := cache.GlobalCache.Get(cacheKey); ok {
		c.Set("Content-Type", "application/json")
		return c.Send(cached)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	argNjkb := db.GetNilaiJualParams{
		Jenis: req.Jenis, Merk: req.Merk, Model: req.Model, Tipe: req.Tipe, Tahun: req.Tahun,
	}
	njkbData, err := h.Queries.GetNilaiJual(ctx, argNjkb)
	if err != nil {
		return c.Status(404).JSON(ErrorResponse{Error: "Master NJKB tidak ditemukan"})
	}

	tarifList, _ := h.Queries.GetAllTarifPKB(ctx)
	var estimasi []EstimasiTarif
	nilaiJual := float64(njkbData.NilaiJual)

	for _, t := range tarifList {
		tpkb, _ := t.TarifPkbPersen.Float64Value()
		tops, _ := t.OpsenPkbPersen.Float64Value()
		pkbPokok := nilaiJual * tpkb.Float64
		estimasi = append(estimasi, EstimasiTarif{
			JenisPlat: t.JenisPlat,
			Label:     t.Label,
			Pkb:       int64(math.Round(pkbPokok)),
			Opsen:     int64(math.Round(pkbPokok * tops.Float64)),
		})
	}

	res := SuccessResponse{
		Pesan: "Sukses",
		Data: KalkulasiNJKBResponse{
			Njkb:     njkbData.NilaiJual,
			Estimasi: estimasi,
		},
	}
	cache.GlobalCache.Set(cacheKey, res)
	return c.JSON(res)
}

// ---------------------------------------------------------------------
// 3. ADMIN API: LIST KENDARAAN (PARALEL)
// ---------------------------------------------------------------------

func (h *BapendaHandler) GetAllInfoPajakAdmin(c fiber.Ctx) error {
	page, errP := strconv.Atoi(c.Query("page", "1"))
	limit, errL := strconv.Atoi(c.Query("limit", "10"))
	if errP != nil || errL != nil || page < 1 || limit < 1 {
		return c.Status(400).JSON(ErrorResponse{Error: "Paginasi tidak valid"})
	}

	offset := (page - 1) * limit
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	var data []db.GetAllKendaraanPajakAdminRow
	var total int64

	g.Go(func() error {
		var err error
		data, err = h.Queries.GetAllKendaraanPajakAdmin(ctx, db.GetAllKendaraanPajakAdminParams{
			LimitData: int32(limit), OffsetData: int32(offset),
		})
		return err
	})

	g.Go(func() error {
		var err error
		total, err = h.Queries.CountKendaraanPajak(ctx)
		return err
    })

	if err := g.Wait(); err != nil {
		return c.Status(500).JSON(ErrorResponse{Error: "Gagal memproses data", Detail: err.Error()})
	}

	return c.JSON(SuccessResponse{
		Pesan: "Sukses",
		Data:  data,
		Pagination: PaginationMeta{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	})
}

func (h *BapendaHandler) GetDetailPajakAdmin(c fiber.Ctx) error {
	plat := strings.ToUpper(strings.ReplaceAll(c.Params("plat"), " ", ""))
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetKendaraanPajakByPlatAdmin(ctx, plat)
	if err != nil {
		return c.Status(404).JSON(ErrorResponse{Error: "Data tidak ditemukan"})
	}
	return c.JSON(SuccessResponse{Pesan: "Sukses", Data: data})
}

// ---------------------------------------------------------------------
// 4. ADMIN API: CRUD (CREATE, UPDATE, DELETE)
// ---------------------------------------------------------------------

func (h *BapendaHandler) CreateInfoPajak(c fiber.Ctx) error {
	var req KendaraanRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{Error: "JSON tidak valid", Detail: err.Error()})
	}

	t, errT := time.Parse("2006-01-02", req.MasaPajak)
	if errT != nil {
		return c.Status(400).JSON(ErrorResponse{Error: "Format tanggal wajib YYYY-MM-DD"})
	}

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
		return c.Status(409).JSON(ErrorResponse{Error: "Gagal simpan (Duplikat plat nomor)"})
	}
	return c.Status(201).JSON(SuccessResponse{Pesan: "Berhasil", Data: res})
}

func (h *BapendaHandler) UpdateInfoPajak(c fiber.Ctx) error {
	platKey := c.Params("plat")
	var req KendaraanRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{Error: "JSON tidak valid"})
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
		return c.Status(500).JSON(ErrorResponse{Error: "Gagal update"})
	}
	cache.GlobalCache.DeleteByPrefix("pajak:info:" + strings.ToUpper(strings.ReplaceAll(platKey, " ", "")))
	return c.JSON(SuccessResponse{Pesan: "Terupdate"})
}

func (h *BapendaHandler) DeleteInfoPajak(c fiber.Ctx) error {
	plat := strings.ToUpper(strings.ReplaceAll(c.Params("plat"), " ", ""))
	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.Queries.DeleteKendaraanPajak(ctx, plat); err != nil {
		return c.Status(500).JSON(ErrorResponse{Error: "Gagal hapus"})
	}
	cache.GlobalCache.DeleteByPrefix("pajak:info:" + plat)
	return c.JSON(SuccessResponse{Pesan: "Terhapus"})
}

// ---------------------------------------------------------------------
// 5. ADMIN API: MASTER NJKB (PARALEL)
// ---------------------------------------------------------------------

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
		data, err = h.Queries.GetAllMasterNjkbAdmin(ctx, db.GetAllMasterNjkbAdminParams{
			LimitData:  int32(limit),
			OffsetData: int32(offset),
		})
		return err
	})

	g.Go(func() error {
		var err error
		total, err = h.Queries.CountMasterNjkb(ctx)
		return err
	})

	if err := g.Wait(); err != nil {
		return c.Status(500).JSON(ErrorResponse{Error: "Error DB NJKB", Detail: err.Error()})
	}

	var responseData []MasterNjkbSummaryResponse
	for _, row := range data {
		responseData = append(responseData, MasterNjkbSummaryResponse{
			ID:             row.ID,
			NamaKendaraan:  fmt.Sprintf("%s %s %s", row.Merk, row.Model, row.Tipe),
			JenisKendaraan: row.JenisKendaraan,
			Merk:           row.Merk,
			Model:          row.Model,
		})
	}

	if responseData == nil {
		responseData = []MasterNjkbSummaryResponse{}
	}

	return c.JSON(SuccessResponse{
		Pesan: "Sukses",
		Data:  responseData, 
		Pagination: PaginationMeta{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	})
}

func (h *BapendaHandler) GetDetailMasterNjkbAdmin(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(400).JSON(ErrorResponse{Error: "ID harus berupa angka"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetMasterNjkbById(ctx, int32(id))
	if err != nil {
		return c.Status(404).JSON(ErrorResponse{Error: "Data Master NJKB tidak ditemukan"})
	}

	return c.JSON(SuccessResponse{
		Pesan: "Sukses",
		Data:  data, 
	})
}

func (h *BapendaHandler) CreateMasterNjkb(c fiber.Ctx) error {
	var req struct {
		Jenis string `json:"jenis_kendaraan"`
		Merk  string `json:"merk"`
		Model string `json:"model"`
		Tipe  string `json:"tipe"`
		Tahun int16  `json:"tahun"`
		Nilai int64  `json:"nilai_jual"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(ErrorResponse{Error: "Invalid JSON NJKB"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	arg := db.CreateMasterNjkbParams{
		JenisKendaraan: req.Jenis, Merk: req.Merk, Model: req.Model, Tipe: req.Tipe, Tahun: req.Tahun, NilaiJual: req.Nilai,
	}
	res, err := h.Queries.CreateMasterNjkb(ctx, arg)
	if err != nil {
		return c.Status(409).JSON(ErrorResponse{Error: "Master sudah ada"})
	}
	return c.Status(201).JSON(fiber.Map{
        "pesan": "Master Terbuat",
        "id":    res,
    })}

func (h *BapendaHandler) UpdateMasterNjkb(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(400).JSON(ErrorResponse{Error: "ID harus angka"})
	}

	var req struct {
		Jenis string `json:"jenis_kendaraan"`
		Merk  string `json:"merk"`
		Model string `json:"model"`
		Tipe  string `json:"tipe"`
		Tahun int16  `json:"tahun"`
		Nilai int64  `json:"nilai_jual"`
	}
	c.Bind().JSON(&req)

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	arg := db.UpdateMasterNjkbParams{
		ID: int32(id), JenisKendaraan: req.Jenis, Merk: req.Merk, Model: req.Model, Tipe: req.Tipe, Tahun: req.Tahun, NilaiJual: req.Nilai,
	}
	if err := h.Queries.UpdateMasterNjkb(ctx, arg); err != nil {
		return c.Status(500).JSON(ErrorResponse{Error: "Gagal update Master"})
	}
	return c.JSON(SuccessResponse{Pesan: "Master Terupdate"})
}

func (h *BapendaHandler) DeleteMasterNjkb(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(400).JSON(ErrorResponse{Error: "ID tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	if err := h.Queries.DeleteMasterNjkb(ctx, int32(id)); err != nil {
		return c.Status(500).JSON(ErrorResponse{Error: "Gagal hapus master"})
	}
	return c.JSON(SuccessResponse{Pesan: "Master Terhapus"})
}