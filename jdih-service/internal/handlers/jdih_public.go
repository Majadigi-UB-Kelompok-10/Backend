package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/sync/errgroup"

	"github.com/farildzaky/jdih-service/internal/db"
	"github.com/farildzaky/jdih-service/internal/utils"
)

// =============================================================================
// PUBLIC: PENGUMUMAN (Carousel)
// =============================================================================

func (h *JdihHandler) PublicGetPengumuman(c fiber.Ctx) error {
	cacheKey := "jdih:public:pengumuman"
	if respondCached(c, h.Cache, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.ListPengumumanTerbaru(ctx, 5) // Ambil 5 terbaru
	if err != nil {
		slog.Error("public.pengumuman.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal mengambil daftar pengumuman",
		})
	}

	resData := make([]PengumumanResponse, 0, len(data))
	for _, row := range data {
		var isiStr string
		if row.Isi.Valid {
			isiStr = row.Isi.String
		}
		resData = append(resData, PengumumanResponse{
			ID:      row.ID,
			Judul:   row.Judul,
			Isi:     isiStr,
			Tanggal: row.CreatedAt.Time.Format("02 Jan 2006"),
		})
	}

	res := SuccessResponse{Pesan: "Pengumuman Terbaru", Data: resData}
	return cacheJSON(c, h.Cache, cacheKey, CacheTTLPengumuman, res)
}

// =============================================================================
// PUBLIC: LIST JENIS DOKUMEN
// =============================================================================

func (h *JdihHandler) PublicGetJenisList(c fiber.Ctx) error {
	type JenisItem struct {
		Value string `json:"value"`
		Label string `json:"label"`
	}
	data := []JenisItem{
		{Value: "perda", Label: "Peraturan Daerah"},
		{Value: "pergub", Label: "Peraturan Gubernur"},
		{Value: "peraturan", Label: "Peraturan"},
		{Value: "perdes", Label: "Peraturan Desa"},
		{Value: "sk_gub", Label: "Surat Keputusan Gubernur"},
		{Value: "instruksi", Label: "Instruksi"},
		{Value: "se", Label: "Surat Edaran"},
		{Value: "keputusan", Label: "Keputusan"},
	}
	return c.JSON(SuccessResponse{Pesan: "Daftar Jenis Dokumen", Data: data})
}

// =============================================================================
// PUBLIC: PENCARIAN GLOBAL (Terbaru & Populer)
// =============================================================================

func (h *JdihHandler) PublicSearchDokumen(c fiber.Ctx) error {
	page, limit, offset := parsePagination(c)
	keyword, _ := utils.ValidateQueryString(c.Query("keyword"), 100, "keyword")
	nomor, _ := utils.ValidateQueryString(c.Query("nomor"), 50, "nomor")
	sort := c.Query("sort", "terbaru") // "terbaru" atau "populer"

	var tahunArg pgtype.Int2
	if thn := c.Query("tahun"); thn != "" {
		if t, err := utils.ValidateTahun(thn); err == nil {
			tahunArg = pgtype.Int2{Int16: t, Valid: true}
		}
	}

	var jenisArg db.NullJenisDokumenType
	if jns := c.Query("jenis"); jns != "" {
		if j, err := utils.ValidateJenisDokumen(jns); err == nil {
			jenisArg = db.NullJenisDokumenType{JenisDokumenType: db.JenisDokumenType(j), Valid: true}
		}
	}

	var keywordArg, nomorArg pgtype.Text
	if keyword != "" {
		keywordArg = pgtype.Text{String: keyword, Valid: true}
	}
	if nomor != "" {
		nomorArg = pgtype.Text{String: nomor, Valid: true}
	}

	cacheKey := normalizeKey("jdih:public:search", keyword, nomor, sort, c.Query("tahun"), c.Query("jenis"), fmt.Sprint(page), fmt.Sprint(limit))
	if respondCached(c, h.Cache, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)
	var total int64
	var resData []DokumenItemResponse

	g.Go(func() error {
		var err error
		total, err = h.Queries.CountSearchDokumen(gCtx, db.CountSearchDokumenParams{
			Keyword: keywordArg,
			Nomor:   nomorArg,
			Tahun:   tahunArg,
			Jenis:   jenisArg,
		})
		return err
	})

	g.Go(func() error {
		if sort == "populer" {
			data, err := h.Queries.SearchDokumenPopuler(gCtx, db.SearchDokumenPopulerParams{
				Keyword:    keywordArg,
				Nomor:      nomorArg,
				Tahun:      tahunArg,
				Jenis:      jenisArg,
				LimitData:  int32(limit),
				OffsetData: int32(offset),
			})
			if err != nil {
				return err
			}
			resData = mapSearchRowToItemResponse(data)
		} else {
			data, err := h.Queries.SearchDokumenTerbaru(gCtx, db.SearchDokumenTerbaruParams{
				Keyword:    keywordArg,
				Nomor:      nomorArg,
				Tahun:      tahunArg,
				Jenis:      jenisArg,
				LimitData:  int32(limit),
				OffsetData: int32(offset),
			})
			if err != nil {
				return err
			}
			resData = mapSearchRowToItemResponse(data) // Fungsi helper ada di bawah
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		slog.Error("public.search.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal melakukan pencarian dokumen",
		})
	}

	if resData == nil {
		resData = []DokumenItemResponse{}
	}

	res := PaginatedResponse{
		Pesan:      "Hasil Pencarian",
		Data:       resData,
		Pagination: &PaginationMeta{Page: page, Limit: limit, Total: total},
	}
	return cacheJSON(c, h.Cache, cacheKey, CacheTTLDokumen, res)
}

// =============================================================================
// PUBLIC: AKSES CEPAT (Perda, Pergub, dll)
// =============================================================================

func (h *JdihHandler) PublicListDokumenByJenis(c fiber.Ctx) error {
	jenis, ve := utils.ValidateJenisDokumen(c.Params("jenis"))
	if ve != nil {
		return validationErrorResponse(c, ve)
	}

	page, limit, offset := parsePagination(c)
	keyword, _ := utils.ValidateQueryString(c.Query("keyword"), 100, "keyword")
	
	var tahunArg pgtype.Int2
	if thn := c.Query("tahun"); thn != "" {
		if t, err := utils.ValidateTahun(thn); err == nil {
			tahunArg = pgtype.Int2{Int16: t, Valid: true}
		}
	}

	var keywordArg pgtype.Text
	if keyword != "" {
		keywordArg = pgtype.Text{String: keyword, Valid: true}
	}

	cacheKey := normalizeKey("jdih:public:jenis", jenis, keyword, c.Query("tahun"), fmt.Sprint(page), fmt.Sprint(limit))
	if respondCached(c, h.Cache, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)
	var total int64
	var resData []DokumenItemResponse

	g.Go(func() error {
		var err error
		total, err = h.Queries.CountDokumenByJenis(gCtx, db.CountDokumenByJenisParams{
			Jenis:   db.JenisDokumenType(jenis),
			Tahun:   tahunArg,
			Keyword: keywordArg,
		})
		return err
	})

	g.Go(func() error {
		data, err := h.Queries.ListDokumenByJenis(gCtx, db.ListDokumenByJenisParams{
			Jenis:      db.JenisDokumenType(jenis),
			Tahun:      tahunArg,
			Keyword:    keywordArg,
			LimitData:  int32(limit),
			OffsetData: int32(offset),
		})
		if err != nil {
			return err
		}
		for _, row := range data {
			var sizeKb int32
			if row.PdfSizeKb.Valid {
				sizeKb = row.PdfSizeKb.Int32
			}
			resData = append(resData, DokumenItemResponse{
				ID:        row.ID,
				Jenis:     jenis,
				Nomor:     row.Nomor,
				Tahun:     row.Tahun,
				Judul:     row.Judul,
				Tanggal:   row.TanggalPenetapan.Time.Format("02 Jan 2006"),
				Status:    string(row.Status),
				PdfUrl:    row.PdfUrl,
				PdfSizeKb: sizeKb,
			})
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		slog.Error("public.jenis.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal memuat dokumen",
		})
	}

	if resData == nil {
		resData = []DokumenItemResponse{}
	}

	res := PaginatedResponse{
		Pesan:      fmt.Sprintf("Daftar %s", strings.ToUpper(jenis)),
		Data:       resData,
		Pagination: &PaginationMeta{Page: page, Limit: limit, Total: total},
	}
	return cacheJSON(c, h.Cache, cacheKey, CacheTTLDokumen, res)
}

// =============================================================================
// PUBLIC: FILTER TAHUN DROPDOWN
// =============================================================================

func (h *JdihHandler) PublicGetFilterTahun(c fiber.Ctx) error {
	jenis, ve := utils.ValidateJenisDokumen(c.Params("jenis"))
	if ve != nil {
		return validationErrorResponse(c, ve)
	}

	cacheKey := normalizeKey("jdih:public:filter_tahun", jenis)
	if respondCached(c, h.Cache, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	tahunList, err := h.Queries.GetDistinctTahunByJenis(ctx, db.JenisDokumenType(jenis))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_DB", Message: "Gagal ambil filter tahun"})
	}

	res := SuccessResponse{Pesan: "Filter Tahun", Data: tahunList}
	return cacheJSON(c, h.Cache, cacheKey, CacheTTLPengumuman, res) // Cache agak lama krn jarang tambah tahun baru
}

// =============================================================================
// PUBLIC: FILTER TAHUN GLOBAL (untuk halaman Pencarian)
// =============================================================================

func (h *JdihHandler) PublicGetFilterTahunAll(c fiber.Ctx) error {
	cacheKey := "jdih:public:filter_tahun_all"
	if respondCached(c, h.Cache, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	tahunList, err := h.Queries.GetDistinctTahunAll(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Code: "ERR_DB", Message: "Gagal ambil filter tahun"})
	}

	res := SuccessResponse{Pesan: "Filter Tahun", Data: tahunList}
	return cacheJSON(c, h.Cache, cacheKey, CacheTTLPengumuman, res)
}

// =============================================================================
// PUBLIC: DETAIL DOKUMEN & INCREMENT VIEW
// =============================================================================

func (h *JdihHandler) PublicGetDetailDokumen(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: "ID tidak valid"})
	}

	// 1. Trigger Increment View secara Async di Background
	go func(dokID int32) {
		bgCtx, bgCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer bgCancel()
		_ = h.Queries.IncrementView(bgCtx, dokID)
	}(int32(id))

	cacheKey := fmt.Sprintf("jdih:public:detail:%d", id)
	if respondCached(c, h.Cache, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	dok, err := h.Queries.GetDokumenByID(ctx, int32(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Code: "ERR_NOT_FOUND", Message: "Dokumen tidak ditemukan"})
	}

	subjekDB, _ := h.Queries.GetSubjekByDokumen(ctx, int32(id))
	subjekList := make([]SubjekResponse, 0, len(subjekDB))
	for _, s := range subjekDB {
		subjekList = append(subjekList, SubjekResponse{ID: s.ID, Nama: s.Nama})
	}

	var ringkasan, urusan string
	if dok.Ringkasan.Valid {
		ringkasan = dok.Ringkasan.String
	}
	if dok.UrusanPemerintahan.Valid {
		urusan = dok.UrusanPemerintahan.String
	}
	var sizeKb int32
	if dok.PdfSizeKb.Valid {
		sizeKb = dok.PdfSizeKb.Int32
	}

	response := DokumenDetailResponse{
		ID:                 dok.ID,
		Jenis:              string(dok.Jenis),
		Nomor:              dok.Nomor,
		Tahun:              dok.Tahun,
		Judul:              dok.Judul,
		Ringkasan:          ringkasan,
		TanggalPenetapan:   dok.TanggalPenetapan.Time.Format("02 Jan 2006"),
		Status:             string(dok.Status),
		PdfUrl:             dok.PdfUrl,
		PdfSizeKb:          sizeKb,
		UrusanPemerintahan: urusan,
		JumlahView:         dok.JumlahView + 1, // Tambah 1 untuk UX yg bagus krn bg query jalan belakangan
		Subjek:             subjekList,
	}

	res := SuccessResponse{Pesan: "Detail Dokumen", Data: response}
	return cacheJSON(c, h.Cache, cacheKey, CacheTTLDetail, res)
}

// Helper Mapping agar kodingan rapi
func mapSearchRowToItemResponse(data interface{}) []DokumenItemResponse {
	var resData []DokumenItemResponse
	switch v := data.(type) {
	case []db.SearchDokumenTerbaruRow:
		for _, row := range v {
			var ringkasan string
			if row.Ringkasan.Valid {
				ringkasan = row.Ringkasan.String
			}
			resData = append(resData, DokumenItemResponse{
				ID:         row.ID,
				Jenis:      string(row.Jenis),
				Nomor:      row.Nomor,
				Tahun:      row.Tahun,
				Judul:      row.Judul,
				Ringkasan:  ringkasan,
				Tanggal:    row.TanggalPenetapan.Time.Format("02 Jan 2006"),
				Status:     string(row.Status),
				JumlahView: row.JumlahView,
			})
		}
	case []db.SearchDokumenPopulerRow:
		for _, row := range v {
			var ringkasan string
			if row.Ringkasan.Valid {
				ringkasan = row.Ringkasan.String
			}
			resData = append(resData, DokumenItemResponse{
				ID:         row.ID,
				Jenis:      string(row.Jenis),
				Nomor:      row.Nomor,
				Tahun:      row.Tahun,
				Judul:      row.Judul,
				Ringkasan:  ringkasan,
				Tanggal:    row.TanggalPenetapan.Time.Format("02 Jan 2006"),
				Status:     string(row.Status),
				JumlahView: row.JumlahView,
			})
		}
	}
	return resData
}