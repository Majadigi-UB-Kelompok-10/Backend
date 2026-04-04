package handlers

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/sync/errgroup"

	"github.com/farildzaky/sidita-service/internal/cache"
	"github.com/farildzaky/sidita-service/internal/db"
	"github.com/farildzaky/sidita-service/internal/utils"
)

type HotelHandler struct {
	Queries *db.Queries
	Cld     *cloudinary.Cloudinary
}

func NewHotelHandler(q *db.Queries, cld *cloudinary.Cloudinary) *HotelHandler {
	return &HotelHandler{Queries: q, Cld: cld}
}

type HotelListItem struct {
	db.ListHotelRow
	GmapsURL string `json:"gmaps_url"`
}

// =====================================================================
// 1. GET ALL HOTEL (List & Pagination)
// =====================================================================
func (h *HotelHandler) ListHotel(c fiber.Ctx) error {
	search, errSearch := utils.ValidateQueryString(c.Query("search"), 100, "search")
	if errSearch != nil {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: errSearch.Message})
	}

	minBintangStr := c.Query("min_bintang")
	var minBintang int
	if minBintangStr != "" {
		minBintang, _ = strconv.Atoi(minBintangStr)
	}

	page, limit := utils.ValidatePaginationParams(c.Query("page", "1"), c.Query("limit", "10"))

    if page < 1 || limit < 1 || limit > 100 {
        fmt.Printf("[VALIDATION ERROR] Invalid pagination: page=%d, limit=%d\n", page, limit)
        return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "Parameter pagination tidak valid (page dan limit harus positif, limit max 100)"})
    }
	cacheKey := fmt.Sprintf("hotel:list:search_%s:bintang_%d:page_%d:limit_%d", search, minBintang, page, limit)

	if cachedBytes, ok := cache.GlobalCache.Get(cacheKey); ok {
		c.Set("Content-Type", "application/json")
		return c.Send(cachedBytes)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	offset := (page - 1) * limit
	arg := db.ListHotelParams{
		LimitData:  int32(limit),
		OffsetData: int32(offset),
	}
	countArg := db.CountHotelParams{}

	if minBintang > 0 {
		arg.MinBintang = pgtype.Int2{Int16: int16(minBintang), Valid: true}
		countArg.MinBintang = arg.MinBintang
	}

	if search != "" {
		areaData, errArea := h.Queries.GetAreaByName(ctx, search)
		if errArea == nil {
			arg.AreaID = pgtype.Int4{Int32: areaData.ID, Valid: true}
			countArg.AreaID = arg.AreaID
		} else {
			arg.Search = pgtype.Text{String: search, Valid: true}
			countArg.Search = arg.Search
		}
	}

	g, gCtx := errgroup.WithContext(ctx)

	var data []db.ListHotelRow
	var totalData int64

	g.Go(func() error {
		res, err := h.Queries.ListHotel(gCtx, arg)
		data = res
		return err
	})

	g.Go(func() error {
		res, err := h.Queries.CountHotel(gCtx, countArg)
		totalData = res
		return err
	})

	if err := g.Wait(); err != nil {
		fmt.Printf("[DATABASE ERROR] ListHotel failed - Search: %s, MinBintang: %d, Error: %v\n", search, minBintang, err)
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal mengambil data hotel dari database"})
	}

	if len(data) == 0 {
		fmt.Printf("[INFO] No hotel found - Search: %s, MinBintang: %d\n", search, minBintang)
	}

	var resultData []HotelListItem
	for _, item := range data {
		gmapsURL := ""
		if item.Lat != 0 && item.Lng != 0 {
			gmapsURL = fmt.Sprintf("https://www.google.com/maps/dir/?api=1&destination=%v,%v", item.Lat, item.Lng)
		}

		resultData = append(resultData, HotelListItem{
			ListHotelRow: item,
			GmapsURL:     gmapsURL,
		})
	}

	if resultData == nil {
		resultData = []HotelListItem{}
	}

	totalPages := int(math.Ceil(float64(totalData) / float64(limit)))

	pesan := "Berhasil mengambil data hotel"
	if totalData == 0 {
		pesan = "Tidak ada data hotel yang ditemukan sesuai kriteria pencarian"
	}

	response := BaseResponse{
		Pesan: pesan,
		Data:  resultData,
		Meta: struct {
			Page       int   `json:"page"`
			Limit      int   `json:"limit"`
			TotalData  int64 `json:"total_data"`
			TotalPages int   `json:"total_pages"`
		}{
			Page:       page,
			Limit:      limit,
			TotalData:  totalData,
			TotalPages: totalPages,
		},
	}

	cache.GlobalCache.Set(cacheKey, response)

	return c.JSON(response)
}

// =====================================================================
// 2. GET DETAIL HOTEL
// =====================================================================
func (h *HotelHandler) GetDetailHotel(c fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "Slug hotel tidak valid!"})
	}

	cacheKey := fmt.Sprintf("hotel:detail:%s", slug)
	if cachedBytes, ok := cache.GlobalCache.Get(cacheKey); ok {
		c.Set("Content-Type", "application/json")
		return c.Send(cachedBytes)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	hotelUtama, err := h.Queries.GetHotelBySlug(ctx, slug)
	if err != nil {
		if err.Error() == "no rows in result set" {
			fmt.Printf("[INFO] Hotel not found - Slug: %s\n", slug)
			return c.Status(fiber.StatusNotFound).JSON(BaseResponse{Error: "Hotel tidak ditemukan!"})
		}
		fmt.Printf("[DATABASE ERROR] GetDetailHotel failed - Slug: %s, Error: %v\n", slug, err)
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal mengambil detail hotel"})
	}

	gmapsURL := ""
	if hotelUtama.Lat != 0 && hotelUtama.Lng != 0 {
		gmapsURL = fmt.Sprintf("https://www.google.com/maps/dir/?api=1&destination=%v,%v", hotelUtama.Lat, hotelUtama.Lng)
	}

	response := BaseResponse{
		Pesan: "Berhasil mengambil detail hotel",
		Data: struct {
			Hotel    db.GetHotelBySlugRow `json:"hotel"`
			GmapsURL string               `json:"gmaps_url"`
		}{
			Hotel:    hotelUtama,
			GmapsURL: gmapsURL,
		},
	}

	cache.GlobalCache.Set(cacheKey, response)

	return c.JSON(response)
}

// =====================================================================
// 3. CREATE HOTEL
// =====================================================================
func (h *HotelHandler) CreateHotel(c fiber.Ctx) error {
	areaName := strings.TrimSpace(c.FormValue("area"))
	nama := strings.TrimSpace(c.FormValue("nama"))
	hargaMulaiStr := strings.TrimSpace(c.FormValue("harga_mulai"))
	bintangStr := strings.TrimSpace(c.FormValue("bintang"))
	deskripsi := strings.TrimSpace(c.FormValue("deskripsi"))
	alamat := strings.TrimSpace(c.FormValue("alamat"))
	highlight := strings.TrimSpace(c.FormValue("highlight_text"))
	mapsURL := strings.TrimSpace(c.FormValue("maps_url"))

	if areaName == "" || nama == "" {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "Area dan Nama wajib diisi!"})
	}

	hargaMulai, _ := strconv.Atoi(hargaMulaiStr)
	bintang, _ := strconv.Atoi(bintangStr)

	ctxDb, cancelDb := context.WithTimeout(context.Background(), ContextDBTimeout)
	defer cancelDb()

	areaData, errArea := h.Queries.GetAreaByName(ctxDb, areaName)
	if errArea != nil {
		return c.Status(fiber.StatusNotFound).JSON(BaseResponse{Error: "Area '" + areaName + "' tidak ditemukan di database!"})
	}

	latVal := strings.TrimSpace(c.FormValue("lat"))
	lngVal := strings.TrimSpace(c.FormValue("lng"))

	if mapsURL != "" {
		exLat, exLng := extractCoordsFromMapsURL(mapsURL)
		if exLat != "" && exLng != "" {
			latVal = exLat
			lngVal = exLng
		}
	}

	var latPg, lngPg pgtype.Numeric
	_ = latPg.Scan(latVal)
	_ = lngPg.Scan(lngVal)

	fileHeader, err := c.FormFile("gambar")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "File gambar wajib diupload!"})
	}
	if fileHeader.Size > 2*1024*1024 {
		return c.Status(fiber.StatusRequestEntityTooLarge).JSON(BaseResponse{Error: "Ukuran gambar maksimal 2MB!"})
	}

	file, errOpen := fileHeader.Open()
	if errOpen != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal membuka file gambar"})
	}
	defer file.Close()

	buffer := make([]byte, 512)
	file.Read(buffer)
	file.Seek(0, 0)
	mimeType := http.DetectContentType(buffer)
	if mimeType != "image/jpeg" && mimeType != "image/png" && mimeType != "image/webp" {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "File harus berupa gambar asli (JPG/PNG/WEBP)!"})
	}

	ctxCld, cancelCld := context.WithTimeout(context.Background(), ContextUploadTimeout)
	defer cancelCld()
	resCld, errUpload := h.Cld.Upload.Upload(ctxCld, file, uploader.UploadParams{
		Folder: "sidita_hotel",
	})
	if errUpload != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal upload gambar ke Cloudinary!"})
	}

	slugBaru := generateSlug(nama)

	var highlightPg pgtype.Text
	if highlight != "" {
		highlightPg = pgtype.Text{String: highlight, Valid: true}
	}

	var hargaPg pgtype.Int4
	if hargaMulai > 0 {
		hargaPg = pgtype.Int4{Int32: int32(hargaMulai), Valid: true}
	}

	var bintangPg pgtype.Int2
	if bintang > 0 {
		bintangPg = pgtype.Int2{Int16: int16(bintang), Valid: true}
	}

	arg := db.CreateHotelParams{
		AreaID:        areaData.ID,
		Nama:          nama,
		Slug:          slugBaru,
		HargaMulai:    hargaPg,
		Bintang:       bintangPg,
		GambarUrl:     resCld.SecureURL,
		Deskripsi:     deskripsi,
		Alamat:        alamat,
		HighlightText: highlightPg,
		Lat:           latPg,
		Lng:           lngPg,
	}

	idBaru, errDb := h.Queries.CreateHotel(ctxDb, arg)
	if errDb != nil {
		go func(publicID string) {
			ctxHapus, cancelHapus := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancelHapus()
			h.Cld.Upload.Destroy(ctxHapus, uploader.DestroyParams{PublicID: publicID})
		}(resCld.PublicID)

		if strings.Contains(errDb.Error(), "duplicate key value violates unique constraint") {
			return c.Status(fiber.StatusConflict).JSON(BaseResponse{Error: "Hotel dengan nama ini sudah ada!"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal menyimpan ke database."})
	}

	cache.GlobalCache.DeleteByPrefix("hotel:list:")

	return c.Status(fiber.StatusCreated).JSON(BaseResponse{
		Pesan: "Hotel berhasil disimpan!",
		Data: struct {
			ID        int32  `json:"id"`
			Nama      string `json:"nama"`
			Slug      string `json:"slug"`
			Kota      string `json:"kota"`
			Lat       string `json:"lat"`
			Lng       string `json:"lng"`
			GambarURL string `json:"gambar_url"`
		}{
			ID:        idBaru.ID,
			Nama:      nama,
			Slug:      slugBaru,
			Kota:      areaData.Nama,
			Lat:       latVal,
			Lng:       lngVal,
			GambarURL: resCld.SecureURL,
		},
	})
}

// =====================================================================
// 4. UPDATE HOTEL
// =====================================================================
func (h *HotelHandler) UpdateHotel(c fiber.Ctx) error {
	idStr := c.Params("id")
	idHotel, errId := strconv.Atoi(idStr)
	if errId != nil || idHotel <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "ID harus berupa angka yang valid!"})
	}

	ctxDb, cancelDb := context.WithTimeout(context.Background(), ContextDBTimeout)
	defer cancelDb()

	dataLama, errCari := h.Queries.GetHotelByID(ctxDb, int32(idHotel))
	if errCari != nil {
		return c.Status(fiber.StatusNotFound).JSON(BaseResponse{Error: "Hotel tidak ditemukan!"})
	}

	areaName := strings.TrimSpace(c.FormValue("area"))
	nama := strings.TrimSpace(c.FormValue("nama"))
	hargaMulaiStr := strings.TrimSpace(c.FormValue("harga_mulai"))
	bintangStr := strings.TrimSpace(c.FormValue("bintang"))
	deskripsi := strings.TrimSpace(c.FormValue("deskripsi"))
	alamat := strings.TrimSpace(c.FormValue("alamat"))
	highlight := strings.TrimSpace(c.FormValue("highlight_text"))
	mapsURL := strings.TrimSpace(c.FormValue("maps_url"))

	finalAreaID := dataLama.AreaID
	finalNama := dataLama.Nama
	finalSlug := dataLama.Slug
	finalDeskripsi := dataLama.Deskripsi
	finalAlamat := dataLama.Alamat
	finalHighlight := dataLama.HighlightText
	finalLat := dataLama.Lat
	finalLng := dataLama.Lng
	finalHarga := dataLama.HargaMulai
	finalBintang := dataLama.Bintang

	if areaName != "" {
		areaData, errArea := h.Queries.GetAreaByName(ctxDb, areaName)
		if errArea != nil {
			return c.Status(fiber.StatusNotFound).JSON(BaseResponse{Error: "Area '" + areaName + "' tidak ditemukan di database!"})
		}
		finalAreaID = areaData.ID
	}

	if nama != "" {
		finalNama = nama
		finalSlug = generateSlug(nama)
	}

	if deskripsi != "" {
		finalDeskripsi = deskripsi
	}

	if alamat != "" {
		finalAlamat = alamat
	}

	if highlight != "" {
		finalHighlight = pgtype.Text{String: highlight, Valid: true}
	}

	if hargaMulaiStr != "" {
		hargaMulai, _ := strconv.Atoi(hargaMulaiStr)
		finalHarga = pgtype.Int4{Int32: int32(hargaMulai), Valid: true}
	}

	if bintangStr != "" {
		bintang, _ := strconv.Atoi(bintangStr)
		finalBintang = pgtype.Int2{Int16: int16(bintang), Valid: true}
	}

	latVal := strings.TrimSpace(c.FormValue("lat"))
	lngVal := strings.TrimSpace(c.FormValue("lng"))

	if mapsURL != "" {
		exLat, exLng := extractCoordsFromMapsURL(mapsURL)
		if exLat != "" && exLng != "" {
			_ = finalLat.Scan(exLat)
			_ = finalLng.Scan(exLng)
		}
	} else if latVal != "" && lngVal != "" {
		_ = finalLat.Scan(latVal)
		_ = finalLng.Scan(lngVal)
	}

	gambarUrlFinal := dataLama.GambarUrl
	var publicIDGambarBaru string

	fileHeader, errFile := c.FormFile("gambar")
	if errFile == nil {
		if fileHeader.Size > 2*1024*1024 {
			return c.Status(fiber.StatusRequestEntityTooLarge).JSON(BaseResponse{Error: "Ukuran gambar maksimal 2MB!"})
		}
		file, errOpen := fileHeader.Open()
		if errOpen == nil {
			defer file.Close()
			ctxCld, cancelCld := context.WithTimeout(context.Background(), ContextUploadTimeout)
			defer cancelCld()
			resCld, errUpload := h.Cld.Upload.Upload(ctxCld, file, uploader.UploadParams{
				Folder: "sidita_hotel",
			})
			if errUpload == nil {
				gambarUrlFinal = resCld.SecureURL
				publicIDGambarBaru = resCld.PublicID
			}
		}
	}

	arg := db.UpdateHotelParams{
		ID:            dataLama.ID,
		AreaID:        finalAreaID,
		Nama:          finalNama,
		Slug:          finalSlug,
		HargaMulai:    finalHarga,
		Bintang:       finalBintang,
		GambarUrl:     gambarUrlFinal,
		Deskripsi:     finalDeskripsi,
		Alamat:        finalAlamat,
		HighlightText: finalHighlight,
		Lat:           finalLat,
		Lng:           finalLng,
	}

	resDb, errUpdate := h.Queries.UpdateHotel(ctxDb, arg)
	if errUpdate != nil {
		if publicIDGambarBaru != "" {
			go func() {
				h.Cld.Upload.Destroy(context.Background(), uploader.DestroyParams{PublicID: publicIDGambarBaru})
			}()
		}
		if strings.Contains(errUpdate.Error(), "duplicate key value violates unique constraint") {
			return c.Status(fiber.StatusConflict).JSON(BaseResponse{Error: "Hotel dengan nama ini sudah ada!"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal mengupdate database."})
	}

	if publicIDGambarBaru != "" && dataLama.GambarUrl != "" {
		oldPublicID := utils.ExtractPublicID(dataLama.GambarUrl)
		if oldPublicID != "" {
			go func() {
				ctxHapus, cancelHapus := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancelHapus()
				h.Cld.Upload.Destroy(ctxHapus, uploader.DestroyParams{PublicID: oldPublicID})
			}()
		}
	}

	cache.GlobalCache.DeleteByPrefix("hotel:list:")
	cache.GlobalCache.Delete(fmt.Sprintf("hotel:detail:%s", dataLama.Slug))
	if dataLama.Slug != finalSlug {
		cache.GlobalCache.Delete(fmt.Sprintf("hotel:detail:%s", finalSlug))
	}

	return c.JSON(BaseResponse{
		Pesan: "Hotel berhasil diupdate!",
		Data: struct {
			ID   int32  `json:"id"`
			Nama string `json:"nama"`
			Slug string `json:"slug"`
		}{
			ID:   resDb.ID,
			Nama: resDb.Nama,
			Slug: resDb.Slug,
		},
	})
}

// =====================================================================
// 5. DELETE HOTEL
// =====================================================================
func (h *HotelHandler) DeleteHotel(c fiber.Ctx) error {
	idStr := c.Params("id")
	idHotel, errId := strconv.Atoi(idStr)
	if errId != nil || idHotel <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "ID harus berupa angka yang valid!"})
	}

	ctxDb, cancelDb := context.WithTimeout(context.Background(), ContextDBTimeout)
	defer cancelDb()

	dataLama, errCari := h.Queries.GetHotelByID(ctxDb, int32(idHotel))
	if errCari != nil {
		return c.Status(fiber.StatusNotFound).JSON(BaseResponse{Error: "Hotel tidak ditemukan!"})
	}

	errHapus := h.Queries.DeleteHotel(ctxDb, dataLama.ID)
	if errHapus != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal menghapus hotel dari database"})
	}

	if dataLama.GambarUrl != "" {
		publicID := utils.ExtractPublicID(dataLama.GambarUrl)
		if publicID != "" {
			go func() {
				ctxHapus, cancelHapus := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancelHapus()
				h.Cld.Upload.Destroy(ctxHapus, uploader.DestroyParams{PublicID: publicID})
			}()
		}
	}

	cache.GlobalCache.DeleteByPrefix("hotel:list:")
	cache.GlobalCache.Delete(fmt.Sprintf("hotel:detail:%s", dataLama.Slug))

	return c.JSON(BaseResponse{
		Pesan: "Hotel beserta gambarnya berhasil dihapus permanen!",
	})
}

// =====================================================================
// 6. CACHE WARMUP HOTEL
// =====================================================================
func (h *HotelHandler) CacheWarmup() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)

	var data []db.ListHotelRow
	var totalData int64

	arg := db.ListHotelParams{
		LimitData:  10,
		OffsetData: 0,
	}
	countArg := db.CountHotelParams{}

	g.Go(func() error {
		res, err := h.Queries.ListHotel(gCtx, arg)
		data = res
		return err
	})

	g.Go(func() error {
		res, err := h.Queries.CountHotel(gCtx, countArg)
		totalData = res
		return err
	})

	if err := g.Wait(); err != nil {
		fmt.Printf("[WARMUP ERROR] Cache warmup hotel failed: %v\n", err)
		return
	}

	var resultData []HotelListItem
	for _, item := range data {
		gmapsURL := ""
		if item.Lat != 0 && item.Lng != 0 {
			gmapsURL = fmt.Sprintf("https://www.google.com/maps/dir/?api=1&destination=%v,%v", item.Lat, item.Lng)
		}

		resultData = append(resultData, HotelListItem{
			ListHotelRow: item,
			GmapsURL:     gmapsURL,
		})
	}

	if resultData == nil {
		resultData = []HotelListItem{}
	}

	if len(resultData) == 0 {
		fmt.Printf("[WARMUP INFO] No hotel data available for cache warmup\n")
	} else {
		fmt.Printf("[WARMUP SUCCESS] Cache warmup hotel completed with %d records\n", len(resultData))
	}

	totalPages := int(math.Ceil(float64(totalData) / 10.0))

	response := BaseResponse{
		Pesan: "Berhasil mengambil data hotel",
		Data:  resultData,
		Meta: struct {
			Page       int   `json:"page"`
			Limit      int   `json:"limit"`
			TotalData  int64 `json:"total_data"`
			TotalPages int   `json:"total_pages"`
		}{
			Page:       1,
			Limit:      10,
			TotalData:  totalData,
			TotalPages: totalPages,
		},
	}

	cacheKey := "hotel:list:search_:bintang_0:page_1:limit_10"

	cache.GlobalCache.Set(cacheKey, response)
}

// =====================================================================
// 7. GET DATA PETA HOTEL
// =====================================================================
func (h *HotelHandler) GetHotelMaps(c fiber.Ctx) error {
	slugQuery, errSlug := utils.ValidateQueryString(c.Query("slug"), 100, "slug")
	if errSlug != nil {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: errSlug.Message})
	}

	searchQuery, errSearch := utils.ValidateQueryString(c.Query("search"), 100, "search")
	if errSearch != nil {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: errSearch.Message})
	}

	cacheKey := fmt.Sprintf("hotel:maps:slug_%s:search_%s", slugQuery, searchQuery)
	if cachedBytes, ok := cache.GlobalCache.Get(cacheKey); ok {
		c.Set("Content-Type", "application/json")
		return c.Send(cachedBytes)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	var centerLat, centerLng interface{} = "-7.697739", "112.493863"
	zoomLevel := 8
	var areaIDParam pgtype.Int4
	var searchTextParam pgtype.Text

	if slugQuery != "" {
		hotel, err := h.Queries.GetHotelBySlug(ctx, slugQuery)
		if err == nil {
			centerLat = hotel.Lat 
			centerLng = hotel.Lng
			zoomLevel = 15
			areaIDParam = pgtype.Int4{Int32: hotel.AreaID, Valid: true}
		}
	} else if searchQuery != "" {
		areaData, errArea := h.Queries.GetAreaByName(ctx, searchQuery)
		if errArea == nil {
			centerLat = areaData.Lat
			centerLng = areaData.Lng
			zoomLevel = 11
			areaIDParam = pgtype.Int4{Int32: areaData.ID, Valid: true}
		} else {
			searchTextParam = pgtype.Text{String: searchQuery, Valid: true}
		}
	}

	arg := db.ListHotelMapsParams{
		AreaID: areaIDParam,
		Search: searchTextParam,
	}

	data, err := h.Queries.ListHotelMaps(ctx, arg)
	if err != nil {
		fmt.Printf("[DATABASE ERROR] GetHotelMaps failed - Slug: %s, Search: %s, Error: %v\n", slugQuery, searchQuery, err)
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal mengambil data peta hotel"})
	}

	if data == nil {
		data = []db.ListHotelMapsRow{}
	}

	if searchQuery != "" && slugQuery == "" && searchTextParam.Valid && len(data) > 0 {
		vLat, _ := data[0].Lat.Value()
		vLng, _ := data[0].Lng.Value()
		centerLat = vLat
		centerLng = vLng
		zoomLevel = 14
	}

	response := BaseResponse{
		Pesan: "Berhasil mengambil data peta hotel",
		Data: struct {
			Center struct {
				Lat  interface{} `json:"lat"`
				Lng  interface{} `json:"lng"`
				Zoom int         `json:"zoom"`
			} `json:"center"`
			Points []db.ListHotelMapsRow `json:"points"`
		}{
			Center: struct {
				Lat  interface{} `json:"lat"`
				Lng  interface{} `json:"lng"`
				Zoom int         `json:"zoom"`
			}{
				Lat:  centerLat,
				Lng:  centerLng,
				Zoom: zoomLevel,
			},
			Points: data,
		},
	}

	cache.GlobalCache.Set(cacheKey, response)
	return c.JSON(response)
}