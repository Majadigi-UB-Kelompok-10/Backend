package handlers

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	// "net/url"
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
// 2. GET ALL DESTINASI (Single Smart Search Bar)
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

    page, limit := utils.ValidatePaginationParams(c.Query("page", "1"), c.Query("limit", "10"))

    if page < 1 || limit < 1 || limit > 100 {
        fmt.Printf("[VALIDATION ERROR] Invalid pagination: page=%d, limit=%d\n", page, limit)
        return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "Parameter pagination tidak valid (page dan limit harus positif, limit max 100)"})
    }

    cacheKey := fmt.Sprintf("destinasi:list:search_%s:kat_%s:page_%d:limit_%d", search, kategori, page, limit)

    if cachedBytes, ok := cache.GlobalCache.Get(cacheKey); ok {
        c.Set("Content-Type", "application/json")
        return c.Send(cachedBytes)
    }

    ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
    defer cancel()

    offset := (page - 1) * limit
    arg := db.ListDestinasiParams{
        LimitData:  int32(limit),
        OffsetData: int32(offset),
    }
    countArg := db.CountDestinasiParams{}

    if kategori != "" {
        arg.Kategori = pgtype.Text{String: kategori, Valid: true}
        countArg.Kategori = arg.Kategori
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
        fmt.Printf("[DATABASE ERROR] ListDestinasi failed - Search: %s, Kategori: %s, Error: %v\n", search, kategori, err)
        return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal mengambil data destinasi dari database"})
    }

    if data == nil {
        data = []db.ListDestinasiRow{}
    }

    if len(data) == 0 {
        fmt.Printf("[INFO] No destinasi found - Search: %s, Kategori: %s\n", search, kategori)
    }

    totalPages := int(math.Ceil(float64(totalData) / float64(limit)))

    pesan := "Berhasil mengambil data destinasi"
    if totalData == 0 {
        pesan = "Tidak ada data destinasi yang ditemukan sesuai kriteria pencarian"
    }

    response := BaseResponse{
        Pesan: pesan,
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
			fmt.Printf("[INFO] Destinasi not found - Slug: %s\n", slug)
			return c.Status(fiber.StatusNotFound).JSON(BaseResponse{Error: "Destinasi tidak ditemukan!"})
		}
		fmt.Printf("[DATABASE ERROR] GetDetailDestinasi failed - Slug: %s, Error: %v\n", slug, err)
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal mengambil detail destinasi"})
	}

	galeri, _ := h.Queries.ListDestinasiGambar(ctx, destinasiUtama.ID)
	if galeri == nil {
		galeri = []db.ListDestinasiGambarRow{}
	}

	vLat, _ := destinasiUtama.Lat.Value()
	vLng, _ := destinasiUtama.Lng.Value()

	gmapsURL := ""
	if vLat != nil && vLng != nil {
		gmapsURL = fmt.Sprintf("https://www.google.com/maps/dir/?api=1&destination=%v,%v", vLat, vLng)
	}

	response := BaseResponse{
		Pesan: "Berhasil mengambil detail destinasi",
		Data: struct {
			Destinasi db.GetDestinasiBySlugRow    `json:"destinasi"`
			GmapsURL  string                      `json:"gmaps_url"` 
			Galeri    []db.ListDestinasiGambarRow `json:"galeri"`
		}{
			Destinasi: destinasiUtama,
			GmapsURL:  gmapsURL, 
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
// 5. UPDATE DESTINASI (Partial Update / Opsional)
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

    finalAreaID := dataLama.AreaID
    finalKategori := dataLama.Kategori
    finalNama := dataLama.Nama
    finalSlug := dataLama.Slug
    finalDeskripsi := dataLama.Deskripsi
    finalAlamat := dataLama.Alamat
    finalHighlight := dataLama.HighlightText
    finalLat := dataLama.Lat
    finalLng := dataLama.Lng

    if areaName != "" {
        areaData, errArea := h.Queries.GetAreaByName(ctxDb, areaName)
        if errArea != nil {
            return c.Status(fiber.StatusNotFound).JSON(BaseResponse{Error: "Area '" + areaName + "' tidak ditemukan di database!"})
        }
        finalAreaID = areaData.ID
    }

    if kategori != "" {
        finalKategori = kategori
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
        finalHighlight = highlight
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
                Folder: "sidita_destinasi",
            })
            if errUpload == nil {
                gambarUrlFinal = resCld.SecureURL 
                publicIDGambarBaru = resCld.PublicID
            }
        }
    }

    arg := db.UpdateDestinasiParams{
        ID:            dataLama.ID,
        AreaID:        finalAreaID,
        Kategori:      finalKategori,
        Nama:          finalNama,
        Slug:          finalSlug,
        GambarUrl:     gambarUrlFinal,
        Deskripsi:     finalDeskripsi,
        Alamat:        finalAlamat,
        HighlightText: finalHighlight,
        Lat:           finalLat,
        Lng:           finalLng,
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
    if dataLama.Slug != finalSlug {
        cache.GlobalCache.Delete(fmt.Sprintf("destinasi:detail:%s", finalSlug))
    }

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
		fmt.Printf("[WARMUP ERROR] Cache warmup destinasi failed: %v\n", err)
		return
	}

	if data == nil {
		data = []db.ListDestinasiRow{}
	}

	if len(data) == 0 {
		fmt.Printf("[WARMUP INFO] No destinasi data available for cache warmup\n")
	} else {
		fmt.Printf("[WARMUP SUCCESS] Cache warmup destinasi completed with %d records\n", len(data))
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
// 8. GET DATA PETA DESTINASI (Mengikuti UI UX Frontend)
// =====================================================================
func (h *DestinasiHandler) GetDestinasiMaps(c fiber.Ctx) error {
    slugQuery, errSlug := utils.ValidateQueryString(c.Query("slug"), 100, "slug")
    if errSlug != nil {
        return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: errSlug.Message})
    }

    searchQuery, errSearch := utils.ValidateQueryString(c.Query("search"), 100, "search")
    if errSearch != nil {
        return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: errSearch.Message})
    }

    cacheKey := fmt.Sprintf("destinasi:maps:slug_%s:search_%s", slugQuery, searchQuery)
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
        dest, err := h.Queries.GetDestinasiBySlug(ctx, slugQuery)
        if err == nil {
            centerLat = dest.Lat
            centerLng = dest.Lng
            zoomLevel = 15 
            areaIDParam = pgtype.Int4{Int32: dest.AreaID, Valid: true} 
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

    arg := db.ListDestinasiMapsParams{
        AreaID: areaIDParam,
        Search: searchTextParam,
    }

    data, err := h.Queries.ListDestinasiMaps(ctx, arg)
    if err != nil {
        fmt.Printf("[DATABASE ERROR] GetDestinasiMaps failed - Slug: %s, Search: %s, Error: %v\n", slugQuery, searchQuery, err)
        return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal mengambil data peta"})
    }

    if data == nil {
        data = []db.ListDestinasiMapsRow{}
    }

    if searchQuery != "" && slugQuery == "" && searchTextParam.Valid && len(data) > 0 {
        centerLat = data[0].Lat
        centerLng = data[0].Lng
        zoomLevel = 14
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