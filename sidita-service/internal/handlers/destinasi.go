package handlers

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"regexp"
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

const (
	ContextQueryTimeout  = 5 * time.Second
	ContextUploadTimeout = 15 * time.Second
	ContextDBTimeout     = 5 * time.Second
)

var slugRegex = regexp.MustCompile(`[^a-z0-9]+`)

func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = slugRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "unnamed"
	}
	return slug
}

type BaseResponse struct {
	Pesan string      `json:"pesan,omitempty"`
	Error string      `json:"error,omitempty"`
	Data  interface{} `json:"data,omitempty"`
	Meta  interface{} `json:"meta,omitempty"`
}

type DestinasiHandler struct {
	Queries *db.Queries
	Cld     *cloudinary.Cloudinary
}

func NewDestinasiHandler(q *db.Queries, cld *cloudinary.Cloudinary) *DestinasiHandler {
	return &DestinasiHandler{Queries: q, Cld: cld}
}

func extractCoordsFromMapsURL(mapsURL string) (string, string) {
	if strings.Contains(mapsURL, "maps.app.goo.gl") {
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		resp, err := client.Get(mapsURL)
		if err == nil {
			mapsURL = resp.Header.Get("Location")
		}
	}

	re := regexp.MustCompile(`@(-?\d+\.\d+),(-?\d+\.\d+)`)
	matches := re.FindStringSubmatch(mapsURL)

	if len(matches) >= 3 {
		return matches[1], matches[2]
	}
	return "", ""
}

// =====================================================================
// 1. GET ALL AREA
// =====================================================================
func (h *DestinasiHandler) GetAllArea(c fiber.Ctx) error {
	cacheKey := "areas:all"
    if cachedBytes, ok := cache.GlobalCache.Get(cacheKey); ok {
        c.Set("Content-Type", "application/json") 
        return c.Send(cachedBytes) 
    }

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	areas, err := h.Queries.GetAllArea(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal mengambil daftar area"})
	}

	response := BaseResponse{
		Pesan: "Berhasil mengambil daftar area",
		Data:  areas,
	}

	cache.GlobalCache.Set(cacheKey, response)
	return c.JSON(response)
}

// =====================================================================
// 2. GET ALL DESTINASI
// =====================================================================
func (h *DestinasiHandler) ListDestinasi(c fiber.Ctx) error {
	search, errSearch := utils.ValidateQueryString(c.Query("search"), 100, "search")
	if errSearch != nil {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: errSearch.Message})
	}

	kategori, errKat := utils.ValidateQueryString(c.Query("kategori"), 50, "kategori")
	if errKat != nil {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: errKat.Message})
	}

	areaIDStr := strings.TrimSpace(c.Query("area_id"))

	page, limit := utils.ValidatePaginationParams(c.Query("page", "1"), c.Query("limit", "10"))

	cacheKey := fmt.Sprintf("destinasi:list:search_%s:kat_%s:area_%s:page_%d:limit_%d",
		search, kategori, areaIDStr, page, limit)

	if cachedBytes, ok := cache.GlobalCache.Get(cacheKey); ok {
		c.Set("Content-Type", "application/json")
		return c.Send(cachedBytes)
	}

	offset := (page - 1) * limit
	arg := db.ListDestinasiParams{
		LimitData:  int32(limit),
		OffsetData: int32(offset),
	}
	countArg := db.CountDestinasiParams{}

	if search != "" {
		arg.Search = pgtype.Text{String: search, Valid: true}
		countArg.Search = arg.Search
	}
	if kategori != "" {
		arg.Kategori = pgtype.Text{String: kategori, Valid: true}
		countArg.Kategori = arg.Kategori
	}
	if areaIDStr != "" {
		if areaID, err := strconv.Atoi(areaIDStr); err == nil {
			arg.AreaID = pgtype.Int4{Int32: int32(areaID), Valid: true}
			countArg.AreaID = arg.AreaID
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)

	var data []db.ListDestinasiRow
	var totalData int64

	g.Go(func() error {
		res, err := h.Queries.ListDestinasi(gCtx, arg)
		data = res
		return err
	})

	g.Go(func() error {
		res, err := h.Queries.CountDestinasi(gCtx, countArg)
		totalData = res
		return err
	})

	if err := g.Wait(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal mengambil data destinasi"})
	}

	if data == nil {
		data = []db.ListDestinasiRow{}
	}

	totalPages := int(math.Ceil(float64(totalData) / float64(limit)))

	response := BaseResponse{
		Pesan: "Berhasil mengambil data destinasi",
		Data:  data,
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
// 3. GET DETAIL DESTINASI
// =====================================================================
func (h *DestinasiHandler) GetDetailDestinasi(c fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "Slug destinasi tidak valid!"})
	}

	cacheKey := fmt.Sprintf("destinasi:detail:%s", slug)
	if cachedBytes, ok := cache.GlobalCache.Get(cacheKey); ok {
		c.Set("Content-Type", "application/json")
		return c.Send(cachedBytes)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	destinasiUtama, err := h.Queries.GetDestinasiBySlug(ctx, slug)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return c.Status(fiber.StatusNotFound).JSON(BaseResponse{Error: "Destinasi tidak ditemukan!"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal mengambil detail destinasi"})
	}

	galeri, _ := h.Queries.ListDestinasiGambar(ctx, destinasiUtama.ID)
	if galeri == nil {
		galeri = []db.ListDestinasiGambarRow{}
	}

	response := BaseResponse{
		Pesan: "Berhasil mengambil detail destinasi",
		Data: struct {
			Destinasi db.GetDestinasiBySlugRow   `json:"destinasi"`
			Galeri    []db.ListDestinasiGambarRow `json:"galeri"`
		}{
			Destinasi: destinasiUtama,
			Galeri:    galeri,
		},
	}

	cache.GlobalCache.Set(cacheKey, response)

	return c.JSON(response)
}

// =====================================================================
// 4. CREATE DESTINASI
// =====================================================================
func (h *DestinasiHandler) CreateDestinasi(c fiber.Ctx) error {
	areaName := strings.TrimSpace(c.FormValue("area"))
	kategori := strings.TrimSpace(c.FormValue("kategori"))
	nama := strings.TrimSpace(c.FormValue("nama"))
	deskripsi := strings.TrimSpace(c.FormValue("deskripsi"))
	alamat := strings.TrimSpace(c.FormValue("alamat"))
	highlight := strings.TrimSpace(c.FormValue("highlight_text"))
	mapsURL := strings.TrimSpace(c.FormValue("maps_url"))

	if areaName == "" || nama == "" || kategori == "" {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "Area, Kategori, dan Nama wajib diisi!"})
	}

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
		Folder: "sidita_destinasi",
	})
	if errUpload != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal upload gambar ke Cloudinary!"})
	}

	slugBaru := generateSlug(nama)

	arg := db.CreateDestinasiParams{
		AreaID:        areaData.ID,
		Kategori:      kategori,
		Nama:          nama,
		Slug:          slugBaru,
		GambarUrl:     resCld.SecureURL,
		Deskripsi:     deskripsi,
		Alamat:        alamat,
		HighlightText: highlight,
		Lat:           latPg,
		Lng:           lngPg,
	}

	idBaru, errDb := h.Queries.CreateDestinasi(ctxDb, arg)
	if errDb != nil {
		go func(publicID string) {
			ctxHapus, cancelHapus := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancelHapus()
			h.Cld.Upload.Destroy(ctxHapus, uploader.DestroyParams{PublicID: publicID})
		}(resCld.PublicID)

		if strings.Contains(errDb.Error(), "duplicate key value violates unique constraint") {
			return c.Status(fiber.StatusConflict).JSON(BaseResponse{Error: "Destinasi dengan nama ini sudah ada!"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal menyimpan ke database."})
	}

	cache.GlobalCache.DeleteByPrefix("destinasi:list:")

	return c.Status(fiber.StatusCreated).JSON(BaseResponse{
		Pesan: "Destinasi berhasil disimpan!",
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
// 5. UPDATE DESTINASI
// =====================================================================
func (h *DestinasiHandler) UpdateDestinasi(c fiber.Ctx) error {
	idStr := c.Params("id")
	idDestinasi, errId := strconv.Atoi(idStr)
	if errId != nil || idDestinasi <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "ID harus berupa angka yang valid!"})
	}

	ctxDb, cancelDb := context.WithTimeout(context.Background(), ContextDBTimeout)
	defer cancelDb()

	dataLama, errCari := h.Queries.GetDestinasiByID(ctxDb, int32(idDestinasi))
	if errCari != nil {
		return c.Status(fiber.StatusNotFound).JSON(BaseResponse{Error: "Destinasi tidak ditemukan!"})
	}

	areaName := strings.TrimSpace(c.FormValue("area"))
	kategori := strings.TrimSpace(c.FormValue("kategori"))
	nama := strings.TrimSpace(c.FormValue("nama"))
	deskripsi := strings.TrimSpace(c.FormValue("deskripsi"))
	alamat := strings.TrimSpace(c.FormValue("alamat"))
	highlight := strings.TrimSpace(c.FormValue("highlight_text"))
	mapsURL := strings.TrimSpace(c.FormValue("maps_url"))

	if areaName == "" || nama == "" || kategori == "" {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "Area, Kategori, dan Nama wajib diisi!"})
	}

	areaData, errArea := h.Queries.GetAreaByName(ctxDb, areaName)
	if errArea != nil {
		return c.Status(fiber.StatusNotFound).JSON(BaseResponse{Error: "Area '" + areaName + "' tidak ditemukan di database!"})
	}

	latPg := dataLama.Lat
	lngPg := dataLama.Lng

	latVal := strings.TrimSpace(c.FormValue("lat"))
	lngVal := strings.TrimSpace(c.FormValue("lng"))

	if mapsURL != "" {
		exLat, exLng := extractCoordsFromMapsURL(mapsURL)
		if exLat != "" && exLng != "" {
			_ = latPg.Scan(exLat)
			_ = lngPg.Scan(exLng)
		}
	} else if latVal != "" && lngVal != "" {
		_ = latPg.Scan(latVal)
		_ = lngPg.Scan(lngVal)
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
				Folder: "sidita_destinasi",
			})
			if errUpload == nil {
				gambarUrlFinal = resCld.SecureURL
				publicIDGambarBaru = resCld.PublicID
			}
		}
	}

	slugBaru := generateSlug(nama)
	arg := db.UpdateDestinasiParams{
		ID:            dataLama.ID,
		AreaID:        areaData.ID,
		Kategori:      kategori,
		Nama:          nama,
		Slug:          slugBaru,
		GambarUrl:     gambarUrlFinal,
		Deskripsi:     deskripsi,
		Alamat:        alamat,
		HighlightText: highlight,
		Lat:           latPg,
		Lng:           lngPg,
	}

	resDb, errUpdate := h.Queries.UpdateDestinasi(ctxDb, arg)
	if errUpdate != nil {
		if publicIDGambarBaru != "" {
			go func() {
				h.Cld.Upload.Destroy(context.Background(), uploader.DestroyParams{PublicID: publicIDGambarBaru})
			}()
		}
		if strings.Contains(errUpdate.Error(), "duplicate key value violates unique constraint") {
			return c.Status(fiber.StatusConflict).JSON(BaseResponse{Error: "Destinasi dengan nama ini sudah ada!"})
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

	cache.GlobalCache.DeleteByPrefix("destinasi:list:")
	cache.GlobalCache.Delete(fmt.Sprintf("destinasi:detail:%s", dataLama.Slug))
	cache.GlobalCache.Delete(fmt.Sprintf("destinasi:detail:%s", slugBaru))

	return c.JSON(BaseResponse{
		Pesan: "Destinasi berhasil diupdate!",
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
// 6. DELETE DESTINASI
// =====================================================================
func (h *DestinasiHandler) DeleteDestinasi(c fiber.Ctx) error {
	idStr := c.Params("id")
	idDestinasi, errId := strconv.Atoi(idStr)
	if errId != nil || idDestinasi <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "ID harus berupa angka yang valid!"})
	}

	ctxDb, cancelDb := context.WithTimeout(context.Background(), ContextDBTimeout)
	defer cancelDb()

	dataLama, errCari := h.Queries.GetDestinasiByID(ctxDb, int32(idDestinasi))
	if errCari != nil {
		return c.Status(fiber.StatusNotFound).JSON(BaseResponse{Error: "Destinasi tidak ditemukan!"})
	}

	errHapus := h.Queries.DeleteDestinasi(ctxDb, dataLama.ID)
	if errHapus != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal menghapus destinasi dari database"})
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

	cache.GlobalCache.DeleteByPrefix("destinasi:list:")
	cache.GlobalCache.Delete(fmt.Sprintf("destinasi:detail:%s", dataLama.Slug))

	return c.JSON(BaseResponse{
		Pesan: "Destinasi beserta gambarnya berhasil dihapus permanen!",
	})
}

// =====================================================================
// 7. CACHE WARMUP DESTINASI
// =====================================================================
func (h *DestinasiHandler) CacheWarmup() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)

	var data []db.ListDestinasiRow
	var totalData int64

	arg := db.ListDestinasiParams{
		LimitData:  10,
		OffsetData: 0,
	}
	countArg := db.CountDestinasiParams{}

	g.Go(func() error {
		res, err := h.Queries.ListDestinasi(gCtx, arg)
		data = res
		return err
	})

	g.Go(func() error {
		res, err := h.Queries.CountDestinasi(gCtx, countArg)
		totalData = res
		return err
	})

	if err := g.Wait(); err != nil {
		fmt.Printf("⚠️ Gagal melakukan Cache Warmup Destinasi: %v\n", err)
		return
	}

	if data == nil {
		data = []db.ListDestinasiRow{}
	}

	totalPages := int(math.Ceil(float64(totalData) / 10.0))

	response := BaseResponse{
		Pesan: "Berhasil mengambil data destinasi",
		Data:  data,
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

	cacheKey := "destinasi:list:search_:kat_:area_:page_1:limit_10"

	cache.GlobalCache.Set(cacheKey, response)
	fmt.Println("Cache Warmup Destinasi Selesai!")
}

// =====================================================================
// 8. GET DATA PETA DESTINASI (Interaktif & Dinamis)
// =====================================================================
func (h *DestinasiHandler) GetDestinasiMaps(c fiber.Ctx) error {
	areaIDStr := strings.TrimSpace(c.Query("area_id"))

	cacheKey := fmt.Sprintf("destinasi:maps:area_%s", areaIDStr)
	if cachedBytes, ok := cache.GlobalCache.Get(cacheKey); ok {
		c.Set("Content-Type", "application/json")
		return c.Send(cachedBytes)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	var centerLat, centerLng interface{}
	var zoomLevel int = 8

	var areaIDParam pgtype.Int4

	if areaIDStr != "" {
		if areaID, err := strconv.Atoi(areaIDStr); err == nil {
			areaIDParam = pgtype.Int4{Int32: int32(areaID), Valid: true}

			if area, errArea := h.Queries.GetAreaByID(ctx, int32(areaID)); errArea == nil {
				centerLat = area.Lat
				centerLng = area.Lng
				zoomLevel = 11
			}
		}
	} else {
		centerLat = "-7.697739"
		centerLng = "112.493863"
	}

	data, err := h.Queries.ListDestinasiMaps(ctx, areaIDParam)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal mengambil data peta"})
	}

	if data == nil {
		data = []db.ListDestinasiMapsRow{}
	}

	response := BaseResponse{
		Pesan: "Berhasil mengambil data peta",
		Data: struct {
			Center struct {
				Lat  interface{} `json:"lat"`
				Lng  interface{} `json:"lng"`
				Zoom int         `json:"zoom"`
			} `json:"center"`
			Points []db.ListDestinasiMapsRow `json:"points"`
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