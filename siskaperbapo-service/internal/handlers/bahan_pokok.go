package handlers

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/farildzaky/siskaperbapo-service/internal/cache"
	"github.com/farildzaky/siskaperbapo-service/internal/db"
	"github.com/farildzaky/siskaperbapo-service/internal/worker"
	"github.com/farildzaky/siskaperbapo-service/internal/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/sync/errgroup"
)

const (
	ContextQueryTimeout   = 5 * time.Second
	ContextUploadTimeout  = 15 * time.Second
	ContextDBTimeout      = 5 * time.Second
	CacheWarmupTimeout    = 10 * time.Second
	CacheShortTTL         = 1 * time.Minute
	CacheLongTTL          = 1 * time.Hour
	CacheAreaTTL          = 1 * time.Hour
)

var slugRegex = regexp.MustCompile(`[^a-z0-9]+`)


func calculateTrend(prices []int32) (int32, string) {
	if len(prices) == 0 {
		return 0, "TETAP"
	}

	hargaTerkini := prices[len(prices)-1]
	trend := "TETAP"

	for i := len(prices) - 2; i >= 0; i-- {
		hargaMasaLalu := prices[i]
		if hargaMasaLalu != hargaTerkini {
			if hargaTerkini > hargaMasaLalu {
				trend = "NAIK"
			} else {
				trend = "TURUN"
			}
			break
		}
	}

	return hargaTerkini, trend
}

func buildBahanPokokItems(bahanData []db.BahanPokok, historyMap map[int32][]int32) []ItemBahanPokok {
	var result []ItemBahanPokok

	for _, bp := range bahanData {
		prices := historyMap[bp.ID]
		harga, tren := calculateTrend(prices)

		result = append(result, ItemBahanPokok{
			ID:            bp.ID,
			Komoditas:     bp.Nama,
			Slug:          bp.Slug,
			Satuan:        bp.Satuan,
			GambarURL:     bp.GambarUrl.String,
			HargaSekarang: harga,
			Tren:          tren,
		})
	}

	return result
}

func buildHistoryMap(riwayatData []db.GetTrenSemuaBahanPokokRow) map[int32][]int32 {
	historyMap := make(map[int32][]int32)
	for _, row := range riwayatData {
		historyMap[row.BahanPokokID] = append(historyMap[row.BahanPokokID], row.RataRataHarga)
	}
	return historyMap
}

type BahanPokokHandler struct {
	Queries    *db.Queries
	Cld        *cloudinary.Cloudinary
	WorkerPool *worker.HargaWorkerPool
	Logger     *Logger
}

func NewBahanPokokHandler(q *db.Queries, cld *cloudinary.Cloudinary, wp *worker.HargaWorkerPool) *BahanPokokHandler {
	return &BahanPokokHandler{
		Queries:    q,
		Cld:        cld,
		WorkerPool: wp,
		Logger:     NewLogger(),
	}
}


func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = slugRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "unnamed"
	}
	return slug
}

// =====================================================================
// 1. GET ALL BAHAN POKOK
// =====================================================================
func (h *BahanPokokHandler) GetAllBahanPokok(c fiber.Ctx) error {
	hariIni := time.Now().Format("2006-01-02")
	tanggalStr := c.Query("tanggal", hariIni)
	bahanPokokRaw := c.Query("bahan_pokok", "")
	areaInputRaw := c.Query("area", "")
	pageStr := c.Query("page", "1")
	limitStr := c.Query("limit", "10")

	bahanPokok, errBahan := utils.ValidateQueryString(bahanPokokRaw, 100, "bahan_pokok")
	if errBahan != nil {
		return c.Status(400).JSON(BaseResponse{Error: "Input bahan_pokok tidak valid"})
	}

	areaInput, errArea := utils.ValidateQueryString(areaInputRaw, 100, "area")
	if errArea != nil {
		return c.Status(400).JSON(BaseResponse{Error: "Input area tidak valid"})
	}

	page, limit := utils.ValidatePaginationParams(pageStr, limitStr)

	areaSlugPilihan := ""
	if areaInput != "" {
		areaSlugPilihan = generateSlug(areaInput)
	}

	parsedTime, err := time.Parse("2006-01-02", tanggalStr)
	if err != nil {
		return c.Status(400).JSON(BaseResponse{Error: "Format tanggal salah (gunakan YYYY-MM-DD)"})
	}

	cacheKey := fmt.Sprintf("all_bahan:%s:page_%d:limit_%d:bahan_pokok_%s:area_%s",
		tanggalStr, page, limit, bahanPokok, areaSlugPilihan)

	if cachedData, ok := cache.GlobalCache.GetImmutable(cacheKey); ok {



	tanggalPg := pgtype.Date{Time: parsedTime, Valid: true}
	offset := (page - 1) * limit

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)

	var bahanData []db.BahanPokok
	var totalCount int64
	var riwayatData []db.GetTrenSemuaBahanPokokRow

	g.Go(func() error {
		data, err := h.Queries.GetAllBahanPokok(gCtx, db.GetAllBahanPokokParams{
			Limit:  int32(limit),
			Offset: int32(offset),
			Nama:   bahanPokok,
		})
		bahanData = data
		return err
	})

	g.Go(func() error {
		count, err := h.Queries.GetTotalBahanPokok(gCtx, bahanPokok)
		totalCount = count
		return err
	})

	g.Go(func() error {
		if areaSlugPilihan == "" {
			data, err := h.Queries.GetTrenSemuaBahanPokok(gCtx, db.GetTrenSemuaBahanPokokParams{
				Tanggal:  tanggalPg,
				AreaSlug: areaSlugPilihan,
			})
			riwayatData = data
			return err
		} else {
			data, err := h.Queries.GetTrenSemuaBahanPokokByArea(gCtx, db.GetTrenSemuaBahanPokokByAreaParams{
				Tanggal:  tanggalPg,
				AreaSlug: areaSlugPilihan,
			})
			if err != nil {
				return err
			}
			for _, row := range data {
				riwayatData = append(riwayatData, db.GetTrenSemuaBahanPokokRow{
					BahanPokokID:  row.BahanPokokID,
					Tanggal:       row.Tanggal,
					RataRataHarga: row.RataRataHarga,
				})
			}
			return nil
		}
	})

	if err := g.Wait(); err != nil {
		return c.Status(500).JSON(BaseResponse{Error: "Gagal mengambil data komoditas"})
	}

	historyMap := buildHistoryMap(riwayatData)
	hasilAkhir := buildBahanPokokItems(bahanData, historyMap)

	totalHalaman := (int(totalCount) + limit - 1) / limit

	response := PaginatedResponse{
		Tanggal:      tanggalStr,
		AreaPilihan:  areaSlugPilihan,
		HalamanIni:   page,
		DataPerHal:   limit,
		TotalData:    totalCount,
		TotalHalaman: totalHalaman,
		Data:         hasilAkhir,
	}

	if response.Data == nil {
		response.Data = []ItemBahanPokok{}
	}

	cache.GlobalCache.SetImmutable(cacheKey, response)

	return c.JSON(response)
}

// =====================================================================
// CACHE WARMUP
// =====================================================================
func (h *BahanPokokHandler) CacheWarmup() {
	hariIni := time.Now().Format("2006-01-02")
	parsedTime, _ := time.Parse("2006-01-02", hariIni)
	tanggalPg := pgtype.Date{Time: parsedTime, Valid: true}

	ctx, cancel := context.WithTimeout(context.Background(), CacheWarmupTimeout)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)

	var bahanData []db.BahanPokok
	var totalCount int64
	var trenData []db.GetTrenSemuaBahanPokokRow

	g.Go(func() error {
		data, err := h.Queries.GetAllBahanPokok(gCtx, db.GetAllBahanPokokParams{Limit: 10, Offset: 0, Nama: ""})
		bahanData = data
		return err
	})
	g.Go(func() error {
		count, err := h.Queries.GetTotalBahanPokok(gCtx, "")
		totalCount = count
		return err
	})
	g.Go(func() error {
		data, err := h.Queries.GetTrenSemuaBahanPokok(gCtx, db.GetTrenSemuaBahanPokokParams{Tanggal: tanggalPg, AreaSlug: ""})
		trenData = data
		return err
	})

	if err := g.Wait(); err != nil {
		h.Logger.Error("Cache warmup failed", err, WithContext("step", "database_query"))
		return
	}

	historyMap := buildHistoryMap(trenData)
	hasilAkhir := buildBahanPokokItems(bahanData, historyMap)

	totalHalaman := (int(totalCount) + 10 - 1) / 10
	response := PaginatedResponse{
		Tanggal:      hariIni,
		AreaPilihan:  "",
		HalamanIni:   1,
		DataPerHal:   10,
		TotalData:    totalCount,
		TotalHalaman: totalHalaman,
		Data:         hasilAkhir,
	}

	cacheKey := fmt.Sprintf("all_bahan:%s:page_1:limit_10:bahan_pokok_:area_", hariIni)
	cache.GlobalCache.Set(cacheKey, response, CacheShortTTL)
}

// =====================================================================
// 2. GET DETAIL BAHAN POKOK
// =====================================================================
func (h *BahanPokokHandler) GetDetailBahanPokok(c fiber.Ctx) error {
	slugParam := c.Params("slug")
	if slugParam == "" {
		return c.Status(400).JSON(BaseResponse{Error: "Slug komoditas tidak valid!"})
	}

	slugParam, errSlug := utils.ValidateQueryString(slugParam, 255, "slug")
	if errSlug != nil {
		return c.Status(400).JSON(BaseResponse{Error: "Slug komoditas tidak valid"})
	}

	hariIni := time.Now().Format("2006-01-02")
	tanggalReqStr := c.Query("tanggal", hariIni)
	areaInputRaw := c.Query("area", "")

	areaInput, errArea := utils.ValidateQueryString(areaInputRaw, 100, "area")
	if errArea != nil {
		return c.Status(400).JSON(BaseResponse{Error: "Input area tidak valid"})
	}

	areaSlugPilihan := ""
	if areaInput != "" {
		areaSlugPilihan = generateSlug(areaInput)
	}

	parsedTime, err := time.Parse("2006-01-02", tanggalReqStr)
	if err != nil {
		return c.Status(400).JSON(BaseResponse{Error: "Format tanggal salah (gunakan YYYY-MM-DD)"})
	}

	cacheKey := fmt.Sprintf("detail_bahan:%s:%s:%s", slugParam, tanggalReqStr, areaSlugPilihan)
	if cachedData, ok := cache.GlobalCache.GetImmutable(cacheKey); ok {
		return c.JSON(cachedData)
	}


	tanggalReqPg := pgtype.Date{Time: parsedTime, Valid: true}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	bahanPokok, errBp := h.Queries.GetBahanPokokBySlug(ctx, slugParam)
	if errBp != nil {
		return c.Status(404).JSON(BaseResponse{Error: "Komoditas tidak ditemukan!"})
	}

	tanggalTerakhir, errTanggal := h.Queries.GetTanggalUpdateTerakhir(ctx, db.GetTanggalUpdateTerakhirParams{
		BahanPokokID: bahanPokok.ID,
		Tanggal:      tanggalReqPg,
	})

	if errTanggal != nil {
		if errTanggal.Error() == "no rows in result set" {
			return c.Status(fiber.StatusNotFound).JSON(BaseResponse{
				Pesan: "Data harga belum tersedia sama sekali untuk komoditas ini hingga tanggal " + tanggalReqStr,
			})
		}
		return c.Status(500).JSON(BaseResponse{Error: "Gagal memeriksa riwayat harga"})
	}

	g, gCtx := errgroup.WithContext(ctx)

	var daftarRes []db.GetDaftarHargaAreaRow
	var rataRes []db.GetRataRataArea15HariRow
	var grafikRes []db.GetRiwayatHargaRataRataRow

	g.Go(func() error {
		data, err := h.Queries.GetDaftarHargaArea(gCtx, db.GetDaftarHargaAreaParams{
			BahanPokokID: bahanPokok.ID,
			Tanggal:      tanggalTerakhir,
		})
		daftarRes = data
		return err
	})

	g.Go(func() error {
		data, err := h.Queries.GetRataRataArea15Hari(gCtx, db.GetRataRataArea15HariParams{
			BahanPokokID: bahanPokok.ID,
			Tanggal:      tanggalTerakhir,
		})
		rataRes = data
		return err
	})

	g.Go(func() error {
		data, err := h.Queries.GetRiwayatHargaRataRata(gCtx, db.GetRiwayatHargaRataRataParams{
			BahanPokokID: bahanPokok.ID,
			Tanggal:      tanggalTerakhir,
		})
		grafikRes = data
		return err
	})

	if err := g.Wait(); err != nil {
		return c.Status(500).JSON(BaseResponse{Error: "Gagal mengambil detail area/grafik"})
	}

	var hargaUtama int32 = 0
	var totalHarga int32 = 0

	for _, item := range daftarRes {
		totalHarga += item.Harga
		if item.AreaSlug == areaSlugPilihan {
			hargaUtama = item.Harga
		}
	}

	if hargaUtama == 0 && len(daftarRes) > 0 {
		hargaUtama = totalHarga / int32(len(daftarRes))
	}

	var finalKabKota []ItemKabKota
	for _, item := range daftarRes {
		finalKabKota = append(finalKabKota, ItemKabKota{
			Area:     item.Area,
			AreaSlug: item.AreaSlug,
			Harga:    item.Harga,
		})
	}

	var finalGrafik []ItemGrafik
	for _, item := range grafikRes {
		finalGrafik = append(finalGrafik, ItemGrafik{
			Tanggal:       item.Tanggal.Time.Format("2006-01-02"), 
			RataRataHarga: item.RataRataHarga,
		})
	}

	response := DetailBahanResponse{
		IDKomoditas:       bahanPokok.ID, 
		Komoditas:         bahanPokok.Nama,
		Slug:              bahanPokok.Slug,
		Satuan:            bahanPokok.Satuan,
		GambarURL:         bahanPokok.GambarUrl.String,
		Tanggal:           tanggalReqStr,
		TanggalDataAktual: tanggalTerakhir.Time.Format("2006-01-02"),
		AreaPilihan:       areaSlugPilihan,
		HargaUtama:        hargaUtama,
		GrafikRiwayat:     finalGrafik,  
		ListKabKota:       finalKabKota, 
	}

	if len(rataRes) > 0 {
		dataTertinggi := rataRes[0]
		dataTerendah := rataRes[len(rataRes)-1]
		response.Statistik15Hari.Tertinggi = &StatistikHarga{Area: dataTertinggi.Area, AreaSlug: dataTertinggi.AreaSlug, Harga: dataTertinggi.RataRata15Hari}
		response.Statistik15Hari.Terendah = &StatistikHarga{Area: dataTerendah.Area, AreaSlug: dataTerendah.AreaSlug, Harga: dataTerendah.RataRata15Hari}
	}

	response.Tren = "TETAP"
	if len(grafikRes) > 0 {
		hargaTerkini := grafikRes[len(grafikRes)-1].RataRataHarga
		for i := len(grafikRes) - 2; i >= 0; i-- {
			hargaMasaLalu := grafikRes[i].RataRataHarga
			if hargaMasaLalu != hargaTerkini {
				if hargaTerkini > hargaMasaLalu { response.Tren = "NAIK" } else { response.Tren = "TURUN" }
				break
			}
		}
	}

	cache.GlobalCache.SetImmutable(cacheKey, response)
	if response.GrafikRiwayat == nil { 
		response.GrafikRiwayat = []ItemGrafik{} 
	}
	if response.ListKabKota == nil { 
		response.ListKabKota = []ItemKabKota{} 
	}
	return c.JSON(response)
}

// =====================================================================
// 3. LIST AREA (IMMUTABLE CACHE - No DB query after first request)
// =====================================================================
func (h *BahanPokokHandler) GetAllAreas(c fiber.Ctx) error {
	cacheKey := "all_areas_list"
	
	if cachedData, ok := cache.GlobalCache.GetImmutable(cacheKey); ok {
		return c.JSON(cachedData)
	}


	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	areas, err := h.Queries.GetAllAreas(ctx) 
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{
			Error: "Gagal mengambil daftar area dari database.",
		})
	}

	var finalAreas []ItemArea
	for _, a := range areas {
		finalAreas = append(finalAreas, ItemArea{
			ID:   a.ID,
			Nama: a.Nama,   
			Slug: a.Slug,   
		})
	}

	if finalAreas == nil {
		finalAreas = []ItemArea{}
	}

	response := AreaListResponse{
		Pesan: "Berhasil mengambil daftar area",
		Data:  finalAreas,
	}

	cache.GlobalCache.SetImmutable(cacheKey, response)

	return c.JSON(response)
}

// =====================================================================
// 4. CREATE BAHAN POKOK
// =====================================================================
func (h *BahanPokokHandler) CreateBahanPokok(c fiber.Ctx) error {
	nama := strings.TrimSpace(c.FormValue("nama"))
	satuan := strings.TrimSpace(c.FormValue("satuan"))

	if nama == "" {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "Nama komoditas wajib diisi!"})
	}
	if satuan == "" {
		satuan = "kg"
	}

	fileHeader, err := c.FormFile("gambar")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "File gambar wajib diupload!"})
	}
	if fileHeader.Size > 2*1024*1024 {
		return c.Status(fiber.StatusRequestEntityTooLarge).JSON(BaseResponse{Error: "Ukuran gambar maksimal 2MB!"})
	}

	file, errOpen := fileHeader.Open()
	if errOpen != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal memproses file gambar"})
	}
	defer file.Close()

	buffer := make([]byte, 512)
    _, errRead := file.Read(buffer)
    if errRead != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal membaca konten file"})
    }

	if _, errSeek := file.Seek(0, 0); errSeek != nil {
         return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal me-reset pembacaan file"})
    }

	mimeType := http.DetectContentType(buffer)

	if mimeType != "image/jpeg" && 
       mimeType != "image/png" && 
       mimeType != "image/webp" {
        return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "Virus terdeteksi! File harus berupa gambar asli (JPG, PNG, atau WEBP)!"})
    }

	ctxCld, cancelCld := context.WithTimeout(context.Background(), ContextUploadTimeout)
	defer cancelCld()
	resCld, errUpload := h.Cld.Upload.Upload(ctxCld, file, uploader.UploadParams{
		Folder: "siskaperbapo_images",
	})
	if errUpload != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal upload gambar ke Cloudinary!"})
	}

	slugBaru := generateSlug(nama)

	ctxDb, cancelDb := context.WithTimeout(context.Background(), ContextDBTimeout)
	defer cancelDb()

	idBaru, errDb := h.Queries.CreateBahanPokok(ctxDb, db.CreateBahanPokokParams{
		Nama:      nama,
		Slug:      slugBaru,
		Satuan:    satuan,
		GambarUrl: pgtype.Text{String: resCld.SecureURL, Valid: true},
	})

	if errDb != nil {
		go func(publicID string) {
            ctxHapus, cancelHapus := context.WithTimeout(context.Background(), 10*time.Second)
            defer cancelHapus()
            
            _, err := h.Cld.Upload.Destroy(ctxHapus, uploader.DestroyParams{PublicID: publicID})
            if err != nil {
                fmt.Printf("Failed to delete garbage image from Cloudinary (ID: %s)\n", publicID)
            }
        }(resCld.PublicID)		
		if strings.Contains(errDb.Error(), "duplicate key value violates unique constraint") {
			return c.Status(fiber.StatusConflict).JSON(BaseResponse{Error: "Bahan pokok ini sudah ada di database!"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal menyimpan ke database."})
	}

	cache.GlobalCache.DeleteByPrefix("all_bahan:")

	return c.Status(fiber.StatusCreated).JSON(BaseResponse{
		Pesan: "Mantap! Komoditas berhasil disimpan.",
		Data: ItemBahanPokok{
			ID:        idBaru.ID,
			Komoditas: idBaru.Nama,
			Slug:      idBaru.Slug,
			Satuan:    idBaru.Satuan,
			GambarURL: idBaru.GambarUrl.String,
		},
	})
}

// =====================================================================
// 5. CREATE HARGA HARIAN
// =====================================================================
func (h *BahanPokokHandler) CreateHargaHarian(c fiber.Ctx) error {
	namaKomoditas := strings.TrimSpace(c.FormValue("komoditas"))
	namaArea := strings.TrimSpace(c.FormValue("area"))
	hargaStr := c.FormValue("harga")
	tanggalStr := strings.TrimSpace(c.FormValue("tanggal"))

	harga, errH := strconv.Atoi(hargaStr)
	if errH != nil || harga <= 0 {
		return c.Status(400).JSON(BaseResponse{Error: "Harga harus berupa angka valid!"})
	}

	parsedTime, errT := time.Parse("2006-01-02", tanggalStr)
	if errT != nil {
		return c.Status(400).JSON(BaseResponse{Error: "Format tanggal salah (YYYY-MM-DD)"})
	}
	tanggalPg := pgtype.Date{Time: parsedTime, Valid: true}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	idBP, errBp := h.Queries.GetBahanPokokIDByName(ctx, namaKomoditas)
	if errBp != nil {
		return c.Status(404).JSON(BaseResponse{Error: "Komoditas '" + namaKomoditas + "' tidak ditemukan!"})
	}

	idArea, errAr := h.Queries.GetAreaIDByName(ctx, namaArea)
	if errAr != nil {
		return c.Status(404).JSON(BaseResponse{Error: "Area '" + namaArea + "' tidak ditemukan!"})
	}

	dataHargaBaru, errDb := h.Queries.CreateHargaHarian(ctx, db.CreateHargaHarianParams{
		BahanPokokID: idBP,
		AreaID:       idArea,
		Harga:        int32(harga),
		Tanggal:      tanggalPg,
	})

	if errDb != nil {
		return c.Status(500).JSON(BaseResponse{Error: "Gagal menyimpan harga"})
	}

	cache.GlobalCache.InvalidatePattern("all_bahan:")
	cache.GlobalCache.InvalidatePattern("detail_bahan:")

	return c.Status(201).JSON(BaseResponse{
		Pesan: "Mantap! Harga " + namaKomoditas + " di " + namaArea + " berhasil dicatat.",
		Data: DataHargaHarian{
			ID:        dataHargaBaru.ID,
			Komoditas: namaKomoditas,
			Area:      namaArea,
			Harga:     dataHargaBaru.Harga,
			Tanggal:   dataHargaBaru.Tanggal.Time.Format("2006-01-02"),
		},
	})
}

// =====================================================================
// 6. UPDATE HARGA HARIAN
// =====================================================================
func (h *BahanPokokHandler) UpdateHargaHarian(c fiber.Ctx) error {
	idStr := c.Params("id")
	idHarga, err := strconv.Atoi(idStr)
    if err != nil || idHarga <= 0 {
        return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "ID harus berupa angka yang valid!"})
    }

    hargaBaru, err := strconv.Atoi(c.FormValue("harga"))
    if err != nil || hargaBaru <= 0 {
        return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "Harga harus berupa angka yang valid!"})
    }

	ctxDb, cancelDb := context.WithTimeout(context.Background(), ContextDBTimeout)
	defer cancelDb()

	dataUpdate, errDb := h.Queries.UpdateHargaHarian(ctxDb, db.UpdateHargaHarianParams{
		ID:    int32(idHarga),
		Harga: int32(hargaBaru),
	})

	if errDb != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal mengupdate harga."})
	}

	cache.GlobalCache.InvalidatePattern("all_bahan:")
	cache.GlobalCache.InvalidatePattern("detail_bahan:")

	return c.Status(fiber.StatusOK).JSON(BaseResponse{
		Pesan: "Harga berhasil direvisi.",
		Data: DataHargaHarian{
			ID:      dataUpdate.ID,
			Harga:   dataUpdate.Harga,
			Tanggal: dataUpdate.Tanggal.Time.Format("2006-01-02"),
		},
	})
}

// =====================================================================
// 7. DELETE HARGA HARIAN
// =====================================================================
func (h *BahanPokokHandler) DeleteHargaHarian(c fiber.Ctx) error {
	idStr := c.Params("id")
	idHarga, err := strconv.Atoi(idStr)
    if err != nil || idHarga <= 0 {
        return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "ID harus berupa angka yang valid!"})
    }

	ctxDb, cancelDb := context.WithTimeout(context.Background(), ContextDBTimeout)
	defer cancelDb()

	errDb := h.Queries.DeleteHargaHarian(ctxDb, int32(idHarga))
	if errDb != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal menghapus harga harian."})
	}

	cache.GlobalCache.InvalidatePattern("all_bahan:")
	cache.GlobalCache.InvalidatePattern("detail_bahan:")

	return c.Status(fiber.StatusOK).JSON(BaseResponse{
		Pesan: "Data harga berhasil dihapus.",
	})
}

// =====================================================================
// 8. UPDATE BAHAN POKOK
// =====================================================================
func (h *BahanPokokHandler) UpdateBahanPokok(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "ID harus berupa angka yang valid!"})
	}

	ctxDb, cancelDb := context.WithTimeout(context.Background(), ContextDBTimeout)
	defer cancelDb()
	oldData, errDbRead := h.Queries.GetBahanPokokByID(ctxDb, int32(id))
	if errDbRead != nil {
		return c.Status(fiber.StatusNotFound).JSON(BaseResponse{Error: "Data bahan pokok tidak ditemukan!"})
	}

	finalImageUrl := oldData.GambarUrl 

	nama := strings.TrimSpace(c.FormValue("nama"))
	satuan := strings.TrimSpace(c.FormValue("satuan"))
	if nama == "" {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "Nama komoditas wajib diisi!"})
	}
	if satuan == "" {
		satuan = "kg"
	}

	gambarBaru, errFile := c.FormFile("gambar")
	if errFile != nil && errFile != fiber.ErrUnprocessableEntity {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "Gagal memproses file gambar!"})
	}

	if gambarBaru != nil {
		if gambarBaru.Size > 2*1024*1024 {
			return c.Status(fiber.StatusRequestEntityTooLarge).JSON(BaseResponse{Error: "Ukuran gambar maksimal 2MB!"})
		}

		file, errOpen := gambarBaru.Open()
		if errOpen != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal membuka file gambar"})
		}
		defer file.Close()

		buffer := make([]byte, 512)
		if _, errRead := file.Read(buffer); errRead != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal membaca file"})
		}
		if _, errSeek := file.Seek(0, 0); errSeek != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal me-reset file"})
		}

		mimeType := http.DetectContentType(buffer)
		if mimeType != "image/jpeg" && mimeType != "image/png" && mimeType != "image/webp" {
			return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "File harus berupa gambar asli (JPG/PNG/WEBP)!"})
		}

		ctxCld, cancelCld := context.WithTimeout(context.Background(), ContextUploadTimeout)
		defer cancelCld()
		resCld, errUpload := h.Cld.Upload.Upload(ctxCld, file, uploader.UploadParams{
			Folder: "siskaperbapo_images",
		})
		if errUpload != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal upload ke Cloudinary!"})
		}

		finalImageUrl = pgtype.Text{String: resCld.SecureURL, Valid: true}

		if oldData.GambarUrl.Valid && oldData.GambarUrl.String != "" {
			go func(oldUrl string) {
				publicID := utils.ExtractPublicID(oldUrl)
				if publicID != "" {
					h.Cld.Upload.Destroy(context.Background(), uploader.DestroyParams{PublicID: publicID})
				}
			}(oldData.GambarUrl.String)
		}
	}

	slugBaru := generateSlug(nama) 
	dataUpdate, errDb := h.Queries.UpdateBahanPokok(ctxDb, db.UpdateBahanPokokParams{
		ID:        int32(id),
		Nama:      nama,
		Slug:      slugBaru,
		Satuan:    satuan,
		GambarUrl: finalImageUrl, 
	})

	if errDb != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal mengupdate bahan pokok di database."})
	}

	cache.GlobalCache.InvalidatePattern("all_bahan:")
	cache.GlobalCache.InvalidatePattern("detail_bahan:")

	return c.Status(fiber.StatusOK).JSON(BaseResponse{
		Pesan: "Komoditas berhasil diperbarui.",
		Data: ItemBahanPokok{
			ID:        dataUpdate.ID,
			Komoditas: dataUpdate.Nama,
			Slug:      dataUpdate.Slug,
			Satuan:    dataUpdate.Satuan,
			GambarURL: dataUpdate.GambarUrl.String,
		},
	})
}

// =====================================================================
// 10. DELETE BAHAN POKOK
// =====================================================================
func (h *BahanPokokHandler) DeleteBahanPokok(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "ID harus berupa angka yang valid!"})
	}

	ctxDb, cancelDb := context.WithTimeout(context.Background(), ContextDBTimeout)
	defer cancelDb()

	oldData, errDbRead := h.Queries.GetBahanPokokByID(ctxDb, int32(id))
	if errDbRead != nil {
		return c.Status(fiber.StatusNotFound).JSON(BaseResponse{Error: "Data bahan pokok tidak ditemukan!"})
	}

	errDbDelete := h.Queries.DeleteBahanPokok(ctxDb, int32(id))
	if errDbDelete != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal menghapus bahan pokok di database."})
	}

	if oldData.GambarUrl.Valid && oldData.GambarUrl.String != "" {
		go func(imageUrl string) {
			publicID := utils.ExtractPublicID(imageUrl) 
			
			if publicID != "" {
				ctxCld, cancelCld := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancelCld()
				
				_, errCld := h.Cld.Upload.Destroy(ctxCld, uploader.DestroyParams{PublicID: publicID})
				if errCld != nil {
                        h.Logger.Error("Gagal menghapus gambar lama di Cloudinary", errCld, WithContext("public_id", publicID))
                    }
			}
		}(oldData.GambarUrl.String) 
	}

	cache.GlobalCache.DeleteByPrefix("all_bahan:")
	cache.GlobalCache.DeleteByPrefix("detail_bahan:")

	return c.Status(fiber.StatusOK).JSON(BaseResponse{
		Pesan: "Bahan pokok dan fotonya berhasil dihapus secara permanen!",
	})
}