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

type EventHandler struct {
	Queries *db.Queries
	Cld     *cloudinary.Cloudinary
}

func NewEventHandler(q *db.Queries, cld *cloudinary.Cloudinary) *EventHandler {
	return &EventHandler{Queries: q, Cld: cld}
}

// =====================================================================
// 1. GET ALL EVENT (List & Pagination) - Disesuaikan dengan Filter UI
// =====================================================================
func (h *EventHandler) ListEvent(c fiber.Ctx) error {
	areaStr := c.Query("area")
	bulanStr := c.Query("bulan")
	tahunStr := c.Query("tahun")

	page, limit := utils.ValidatePaginationParams(c.Query("page", "1"), c.Query("limit", "10"))

    if page < 1 || limit < 1 || limit > 100 {
        fmt.Printf("[VALIDATION ERROR] Invalid pagination: page=%d, limit=%d\n", page, limit)
        return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "Parameter pagination tidak valid (page dan limit harus positif, limit max 100)"})
    }
	cacheKey := fmt.Sprintf("event:list:area_%s:bulan_%s:tahun_%s:page_%d:limit_%d", areaStr, bulanStr, tahunStr, page, limit)

	if cachedBytes, ok := cache.GlobalCache.Get(cacheKey); ok {
		c.Set("Content-Type", "application/json")
		return c.Send(cachedBytes)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	offset := (page - 1) * limit
	arg := db.ListEventParams{
		LimitData:  int32(limit),
		OffsetData: int32(offset),
	}
	countArg := db.CountEventParams{}

	tahunValInt := time.Now().Year() 
	if tahunStr != "" {
		if val, err := strconv.Atoi(tahunStr); err == nil {
			arg.Tahun = pgtype.Int2{Int16: int16(val), Valid: true}
			countArg.Tahun = arg.Tahun
			tahunValInt = val
		}
	}

	if bulanStr != "" {
		var bulanVal int
		if num, err := strconv.Atoi(bulanStr); err == nil {
			bulanVal = num
		} else {
			switch strings.ToLower(bulanStr) {
			case "januari", "jan", "january": bulanVal = 1
			case "februari", "feb", "february": bulanVal = 2
			case "maret", "mar", "march": bulanVal = 3
			case "april", "apr": bulanVal = 4
			case "mei", "may": bulanVal = 5
			case "juni", "jun", "june": bulanVal = 6
			case "juli", "jul", "july": bulanVal = 7
			case "agustus", "agu", "august", "aug": bulanVal = 8
			case "september", "sep": bulanVal = 9
			case "oktober", "okt", "october", "oct": bulanVal = 10
			case "november", "nov": bulanVal = 11
			case "desember", "des", "december", "dec": bulanVal = 12
			}
		}

		if bulanVal >= 1 && bulanVal <= 12 {
			startDate := time.Date(tahunValInt, time.Month(bulanVal), 1, 0, 0, 0, 0, time.UTC)
			endDate := startDate.AddDate(0, 1, 0) 

			arg.StartDate = pgtype.Date{Time: startDate, Valid: true}
			arg.EndDate = pgtype.Date{Time: endDate, Valid: true}
			countArg.StartDate = arg.StartDate
			countArg.EndDate = arg.EndDate
		}
	}

	if areaStr != "" {
		areaData, errArea := h.Queries.GetAreaByName(ctx, areaStr)
		if errArea == nil {
			arg.AreaID = pgtype.Int4{Int32: areaData.ID, Valid: true}
			countArg.AreaID = arg.AreaID
		} else {
			arg.AreaID = pgtype.Int4{Int32: -1, Valid: true}
			countArg.AreaID = arg.AreaID
		}
	}

	g, gCtx := errgroup.WithContext(ctx)
	var data []db.ListEventRow
	var totalData int64

	g.Go(func() error {
		res, err := h.Queries.ListEvent(gCtx, arg)
		data = res
		return err
	})

	g.Go(func() error {
		res, err := h.Queries.CountEvent(gCtx, countArg)
		totalData = res
		return err
	})

	if err := g.Wait(); err != nil {
        fmt.Printf("[DATABASE ERROR] ListEvent failed - Area: %s, Year: %s, Error: %v\n", areaStr, tahunStr, err)
        return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal mengambil data event dari database"})
    }

    if data == nil {
        data = []db.ListEventRow{}
    }

    if len(data) == 0 {
        fmt.Printf("[INFO] No events found - Area: %s, Year: %s\n", areaStr, tahunStr)
    } 

    totalPages := int(math.Ceil(float64(totalData) / float64(limit)))

	pesan := "Berhasil mengambil data event"
	if totalData == 0 {
		pesan = "Tidak ada data event yang ditemukan sesuai kriteria pencarian"
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
// 2. GET DETAIL EVENT
// =====================================================================
func (h *EventHandler) GetDetailEvent(c fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "Slug event tidak valid!"})
	}

	cacheKey := fmt.Sprintf("event:detail:%s", slug)
	if cachedBytes, ok := cache.GlobalCache.Get(cacheKey); ok {
		c.Set("Content-Type", "application/json")
		return c.Send(cachedBytes)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	eventUtama, err := h.Queries.GetEventBySlug(ctx, slug)
	if err != nil {
		if err.Error() == "no rows in result set" {
			fmt.Printf("[INFO] Event not found - Slug: %s\n", slug)
			return c.Status(fiber.StatusNotFound).JSON(BaseResponse{Error: "Event tidak ditemukan!"})
		}
		fmt.Printf("[DATABASE ERROR] GetDetailEvent failed - Slug: %s, Error: %v\n", slug, err)
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal mengambil detail event"})
	}

	gmapsURL := ""
	if eventUtama.Lat != 0 && eventUtama.Lng != 0 {
		gmapsURL = fmt.Sprintf("https://www.google.com/maps?q=%f,%f", eventUtama.Lat, eventUtama.Lng)
	}

	response := BaseResponse{
		Pesan: "Berhasil mengambil detail event",
		Data: struct {
			Event    db.GetEventBySlugRow `json:"event"`
			GmapsURL string               `json:"gmaps_url"`
		}{
			Event:    eventUtama,
			GmapsURL: gmapsURL,
		},
	}

	cache.GlobalCache.Set(cacheKey, response)

	return c.JSON(response)
}

// =====================================================================
// 3. GET EVENT MAPS
// =====================================================================
func (h *EventHandler) GetEventMaps(c fiber.Ctx) error {
    slugQuery, errSlug := utils.ValidateQueryString(c.Query("slug"), 100, "slug")
    if errSlug != nil {
        return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: errSlug.Message})
    }

    // Ubah search jadi area agar konsisten dengan filter
    areaQuery, errArea := utils.ValidateQueryString(c.Query("area"), 100, "area")
    if errArea != nil {
        return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: errArea.Message})
    }

    cacheKey := fmt.Sprintf("event:maps:slug_%s:area_%s", slugQuery, areaQuery)
    if cachedBytes, ok := cache.GlobalCache.Get(cacheKey); ok {
        c.Set("Content-Type", "application/json")
        return c.Send(cachedBytes)
    }

    ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
    defer cancel()

    var centerLat, centerLng interface{} = "-7.697739", "112.493863"
    zoomLevel := 8
    var areaIDParam pgtype.Int4

    if slugQuery != "" {
        eventData, err := h.Queries.GetEventBySlug(ctx, slugQuery)
        if err == nil {
            centerLat = eventData.Lat
            centerLng = eventData.Lng
            zoomLevel = 15
            areaIDParam = pgtype.Int4{Int32: eventData.AreaID, Valid: true}
        }
    } else if areaQuery != "" {
        areaData, errArea := h.Queries.GetAreaByName(ctx, areaQuery)
        if errArea == nil {
            centerLat = areaData.Lat
            centerLng = areaData.Lng
            zoomLevel = 11
            areaIDParam = pgtype.Int4{Int32: areaData.ID, Valid: true}
        }
    }

    arg := db.ListEventMapsParams{
        AreaID: areaIDParam,
    }

    data, err := h.Queries.ListEventMaps(ctx, arg)
    if err != nil {
        fmt.Printf("[DATABASE ERROR] GetEventMaps failed - Slug: %s, Area: %s, Error: %v\n", slugQuery, areaQuery, err)
        return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal mengambil data peta event"})
    }

    if data == nil {
        data = []db.ListEventMapsRow{}
    }

    response := BaseResponse{
        Pesan: "Berhasil mengambil data peta",
        Data: struct {
            Center struct {
                Lat  interface{} `json:"lat"`
                Lng  interface{} `json:"lng"`
                Zoom int         `json:"zoom"`
            } `json:"center"`
            Points []db.ListEventMapsRow `json:"points"`
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

// =====================================================================
// 4. CREATE EVENT
// =====================================================================
func (h *EventHandler) CreateEvent(c fiber.Ctx) error {
	areaName := strings.TrimSpace(c.FormValue("area"))
	nama := strings.TrimSpace(c.FormValue("nama"))
	deskripsi := strings.TrimSpace(c.FormValue("deskripsi"))
	tglMulaiStr := strings.TrimSpace(c.FormValue("tanggal_mulai"))
	tglSelesaiStr := strings.TrimSpace(c.FormValue("tanggal_selesai"))
	infoTiket := strings.TrimSpace(c.FormValue("info_tiket"))
	mapsURL := strings.TrimSpace(c.FormValue("maps_url"))
	hargaTiketStr := strings.TrimSpace(c.FormValue("harga_tiket"))

	if areaName == "" || nama == "" || tglMulaiStr == "" || tglSelesaiStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "Area, Nama, Tanggal Mulai, dan Tanggal Selesai wajib diisi!"})
	}

	if infoTiket == "" {
		infoTiket = "Gratis Umum"
	}

	var hargaTiket pgtype.Int4
	if hargaTiketStr != "" {
		if val, err := strconv.Atoi(hargaTiketStr); err == nil {
			hargaTiket = pgtype.Int4{Int32: int32(val), Valid: true}
		}
	} else {
		hargaTiket = pgtype.Int4{Int32: 0, Valid: true} 
	}

	tMulai, errMulai := time.Parse("2006-01-02", tglMulaiStr)
	tSelesai, errSelesai := time.Parse("2006-01-02", tglSelesaiStr)
	if errMulai != nil || errSelesai != nil {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "Format tanggal salah. Gunakan YYYY-MM-DD"})
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
		Folder: "sidita_event",
	})
	if errUpload != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal upload gambar ke Cloudinary!"})
	}

	slugBaru := generateSlug(nama)

	arg := db.CreateEventParams{
		AreaID:         areaData.ID,
		Nama:           nama,
		Slug:           slugBaru,
		GambarUrl:      resCld.SecureURL,
		Deskripsi:      deskripsi,
		TanggalMulai:   pgtype.Date{Time: tMulai, Valid: true},
		TanggalSelesai: pgtype.Date{Time: tSelesai, Valid: true},
		InfoTiket:      pgtype.Text{String: infoTiket, Valid: true},
		HargaTiket:     hargaTiket,
		Lat:            latPg,
		Lng:            lngPg,
	}

	idBaru, errDb := h.Queries.CreateEvent(ctxDb, arg)
	if errDb != nil {
		go func(publicID string) {
			ctxHapus, cancelHapus := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancelHapus()
			h.Cld.Upload.Destroy(ctxHapus, uploader.DestroyParams{PublicID: publicID})
		}(resCld.PublicID)

		if strings.Contains(errDb.Error(), "duplicate key value violates unique constraint") {
			return c.Status(fiber.StatusConflict).JSON(BaseResponse{Error: "Event dengan nama ini sudah ada!"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal menyimpan ke database."})
	}

	cache.GlobalCache.DeleteByPrefix("event:list:")

	return c.Status(fiber.StatusCreated).JSON(BaseResponse{
		Pesan: "Event berhasil disimpan!",
		Data: struct {
			ID        int32  `json:"id"`
			Nama      string `json:"nama"`
			Slug      string `json:"slug"`
			GambarURL string `json:"gambar_url"`
		}{
			ID:        idBaru.ID,
			Nama:      nama,
			Slug:      slugBaru,
			GambarURL: resCld.SecureURL,
		},
	})
}

// =====================================================================
// 5. UPDATE EVENT
// =====================================================================
func (h *EventHandler) UpdateEvent(c fiber.Ctx) error {
	idStr := c.Params("id")
	idEvent, errId := strconv.Atoi(idStr)
	if errId != nil || idEvent <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "ID harus berupa angka yang valid!"})
	}

	ctxDb, cancelDb := context.WithTimeout(context.Background(), ContextDBTimeout)
	defer cancelDb()

	dataLama, errCari := h.Queries.GetEventByID(ctxDb, int32(idEvent))
	if errCari != nil {
		return c.Status(fiber.StatusNotFound).JSON(BaseResponse{Error: "Event tidak ditemukan!"})
	}

	areaName := strings.TrimSpace(c.FormValue("area"))
	nama := strings.TrimSpace(c.FormValue("nama"))
	deskripsi := strings.TrimSpace(c.FormValue("deskripsi"))
	tglMulaiStr := strings.TrimSpace(c.FormValue("tanggal_mulai"))
	tglSelesaiStr := strings.TrimSpace(c.FormValue("tanggal_selesai"))
	infoTiket := strings.TrimSpace(c.FormValue("info_tiket"))
	hargaTiketStr := strings.TrimSpace(c.FormValue("harga_tiket"))
	mapsURL := strings.TrimSpace(c.FormValue("maps_url"))

	finalAreaID := dataLama.AreaID
	finalNama := dataLama.Nama
	finalSlug := dataLama.Slug
	finalDeskripsi := dataLama.Deskripsi
	finalTglMulai := dataLama.TanggalMulai
	finalTglSelesai := dataLama.TanggalSelesai
	finalInfoTiket := dataLama.InfoTiket
	finalHargaTiket := dataLama.HargaTiket
	finalLat := dataLama.Lat
	finalLng := dataLama.Lng

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

	if tglMulaiStr != "" {
		if t, err := time.Parse("2006-01-02", tglMulaiStr); err == nil {
			finalTglMulai = pgtype.Date{Time: t, Valid: true}
		}
	}

	if tglSelesaiStr != "" {
		if t, err := time.Parse("2006-01-02", tglSelesaiStr); err == nil {
			finalTglSelesai = pgtype.Date{Time: t, Valid: true}
		}
	}

	if infoTiket != "" {
		finalInfoTiket = pgtype.Text{String: infoTiket, Valid: true}
	}

	if hargaTiketStr != "" {
		if val, err := strconv.Atoi(hargaTiketStr); err == nil {
			finalHargaTiket = pgtype.Int4{Int32: int32(val), Valid: true}
		}
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
				Folder: "sidita_event",
			})
			if errUpload == nil {
				gambarUrlFinal = resCld.SecureURL
				publicIDGambarBaru = resCld.PublicID
			}
		}
	}

	arg := db.UpdateEventParams{
		ID:             dataLama.ID,
		AreaID:         finalAreaID,
		Nama:           finalNama,
		Slug:           finalSlug,
		GambarUrl:      gambarUrlFinal,
		Deskripsi:      finalDeskripsi,
		TanggalMulai:   finalTglMulai,
		TanggalSelesai: finalTglSelesai,
		InfoTiket:      finalInfoTiket,
		HargaTiket:     finalHargaTiket,
		Lat:            finalLat,
		Lng:            finalLng,
	}

	resDb, errUpdate := h.Queries.UpdateEvent(ctxDb, arg)
	if errUpdate != nil {
		if publicIDGambarBaru != "" {
			go func() {
				h.Cld.Upload.Destroy(context.Background(), uploader.DestroyParams{PublicID: publicIDGambarBaru})
			}()
		}
		if strings.Contains(errUpdate.Error(), "duplicate key value violates unique constraint") {
			return c.Status(fiber.StatusConflict).JSON(BaseResponse{Error: "Event dengan nama ini sudah ada!"})
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

	cache.GlobalCache.DeleteByPrefix("event:list:")
	cache.GlobalCache.DeleteByPrefix("event:maps:")
	cache.GlobalCache.Delete(fmt.Sprintf("event:detail:%s", dataLama.Slug))
	if dataLama.Slug != finalSlug {
		cache.GlobalCache.Delete(fmt.Sprintf("event:detail:%s", finalSlug))
	}

	return c.JSON(BaseResponse{
		Pesan: "Event berhasil diupdate!",
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
// 6. DELETE EVENT
// =====================================================================
func (h *EventHandler) DeleteEvent(c fiber.Ctx) error {
	idStr := c.Params("id")
	idEvent, errId := strconv.Atoi(idStr)
	if errId != nil || idEvent <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{Error: "ID harus berupa angka yang valid!"})
	}

	ctxDb, cancelDb := context.WithTimeout(context.Background(), ContextDBTimeout)
	defer cancelDb()

	dataLama, errCari := h.Queries.GetEventByID(ctxDb, int32(idEvent))
	if errCari != nil {
		return c.Status(fiber.StatusNotFound).JSON(BaseResponse{Error: "Event tidak ditemukan!"})
	}

	errHapus := h.Queries.DeleteEvent(ctxDb, dataLama.ID)
	if errHapus != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{Error: "Gagal menghapus event dari database"})
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

	cache.GlobalCache.DeleteByPrefix("event:list:")
	cache.GlobalCache.DeleteByPrefix("event:maps:")
	cache.GlobalCache.Delete(fmt.Sprintf("event:detail:%s", dataLama.Slug))

	return c.JSON(BaseResponse{
		Pesan: "Event beserta gambarnya berhasil dihapus permanen!",
	})
}

// =====================================================================
// 7. CACHE WARMUP EVENT
// =====================================================================
func (h *EventHandler) CacheWarmup() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	arg := db.ListEventParams{
		LimitData:  10,
		OffsetData: 0,
	}
	countArg := db.CountEventParams{}

	data, err := h.Queries.ListEvent(ctx, arg)
	totalData, _ := h.Queries.CountEvent(ctx, countArg)

	if err != nil {
		fmt.Printf("[WARMUP ERROR] Cache warmup event failed: %v\n", err)
		return
	}

	if len(data) == 0 {
		fmt.Printf("[WARMUP INFO] No event data available for cache warmup\n")
	} else {
		fmt.Printf("[WARMUP SUCCESS] Cache warmup event completed with %d records\n", len(data))
	}

	totalPages := int(math.Ceil(float64(totalData) / 10.0))
	response := BaseResponse{
		Pesan: "Berhasil mengambil data event",
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

	cacheKey := "event:list:area_:bulan_:tahun_:page_1:limit_10"
	cache.GlobalCache.Set(cacheKey, response)
}