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
	// "github.com/farildzaky/sidita-service/internal/utils"
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
// 1. GET ALL AREA (Untuk Dropdown di Frontend)
// =====================================================================
func (h *DestinasiHandler) GetAllArea(c fiber.Ctx) error {
	cacheKey := "areas:all"
	if cachedData, ok := cache.GlobalCache.Get(cacheKey); ok {
		return c.JSON(cachedData)
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
	search := strings.TrimSpace(c.Query("search"))
	kategori := strings.TrimSpace(c.Query("kategori"))
	areaIDStr := c.Query("area_id")

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	cacheKey := fmt.Sprintf("destinasi:list:search_%s:kat_%s:area_%s:page_%d:limit_%d",
		search, kategori, areaIDStr, page, limit)

	// <-- AMAN: Menggunakan fungsi Get() saja
	if cachedData, ok := cache.GlobalCache.Get(cacheKey); ok {
		return c.JSON(cachedData)
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
		Meta: fiber.Map{
			"page":        page,
			"limit":       limit,
			"total_data":  totalData,
			"total_pages": totalPages,
		},
	}

	cache.GlobalCache.Set(cacheKey, response)

	return c.JSON(response)
}


// =====================================================================
// 3. GET DETAIL DESTINASI
// =====================================================================
func (h *DestinasiHandler) GetDetailDestinasi(c fiber.Ctx) error { // <-- AMAN: Tanpa pointer *
	slug := c.Params("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "Slug destinasi tidak valid!"})
	}

	cacheKey := fmt.Sprintf("destinasi:detail:%s", slug)
	if cachedData, ok := cache.GlobalCache.Get(cacheKey); ok { // <-- AMAN: Get()
		return c.JSON(cachedData)
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
		Data: fiber.Map{
			"destinasi": destinasiUtama,
			"galeri":    galeri,
		},
	}

	cache.GlobalCache.Set(cacheKey, response)

	return c.JSON(response)
}

// =====================================================================
// 4. CREATE DESTINASI (Auto Resolve Area Name & Google Maps Extract)
// =====================================================================
func (h *DestinasiHandler) CreateDestinasi(c fiber.Ctx) error {
	// 1. Tangkap Input
	areaName := strings.TrimSpace(c.FormValue("area")) // Menggunakan Nama Kota (Misal: "Surabaya")
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

	// 2. Cari Area ID berdasarkan Nama yang diketik User
	areaData, errArea := h.Queries.GetAreaByName(ctxDb, areaName)
	if errArea != nil {
		return c.Status(fiber.StatusNotFound).JSON(BaseResponse{Error: "Area '" + areaName + "' tidak ditemukan di database!"})
	}

	// 3. Logika Auto-Extract Google Maps Koordinat
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

	// 4. Validasi & Upload Gambar
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

	// 5. Simpan ke Database
	slugBaru := generateSlug(nama)

	arg := db.CreateDestinasiParams{
		AreaID:        areaData.ID, // Memakai ID hasil pencarian areaName
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
		// Rollback gambar di Cloudinary kalau database gagal simpan
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

	// 6. Hapus Cache lama agar data baru muncul
	cache.GlobalCache.DeleteByPrefix("destinasi:list:")

	return c.Status(fiber.StatusCreated).JSON(BaseResponse{
		Pesan: "Destinasi berhasil disimpan!",
		Data: fiber.Map{
			"id":         idBaru.ID,
			"nama":       nama,
			"slug":       slugBaru,
			"kota":       areaData.Nama, 
			"lat":        latVal,
			"lng":        lngVal,
			"gambar_url": resCld.SecureURL,
		},
	})
}