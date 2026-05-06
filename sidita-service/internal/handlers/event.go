package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/farildzaky/sidita-service/internal/cache"
	"github.com/farildzaky/sidita-service/internal/db"
	"github.com/farildzaky/sidita-service/internal/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/sync/errgroup"
)

type EventHandler struct {
	Queries *db.Queries
	Cld     *cloudinary.Cloudinary
}

func NewEventHandler(q *db.Queries, cld *cloudinary.Cloudinary) *EventHandler {
	return &EventHandler{Queries: q, Cld: cld}
}

func invalidateEventCache(id int32) {
    cache.GlobalCache.DeleteByPrefix("event:list:")
    cache.GlobalCache.DeleteByPrefix("event:maps:")
    cache.GlobalCache.Delete("event:recommendation")
    cache.GlobalCache.Delete("event:tahun-tersedia")
    if id > 0 {
        cache.GlobalCache.Delete(fmt.Sprintf("event:detail:%d", id))
    }
}

func parseBulan(s string) int {
	if s == "" {
		return 0
	}
	if num, err := strconv.Atoi(s); err == nil && num >= 1 && num <= 12 {
		return num
	}
	switch strings.ToLower(s) {
	case "januari", "jan", "january":
		return 1
	case "februari", "feb", "february":
		return 2
	case "maret", "mar", "march":
		return 3
	case "april", "apr":
		return 4
	case "mei", "may":
		return 5
	case "juni", "jun", "june":
		return 6
	case "juli", "jul", "july":
		return 7
	case "agustus", "agu", "august", "aug":
		return 8
	case "september", "sep":
		return 9
	case "oktober", "okt", "october", "oct":
		return 10
	case "november", "nov":
		return 11
	case "desember", "des", "december", "dec":
		return 12
	}
	return 0
}

// =============================================================================
// PUBLIC: LIST EVENT 
// =============================================================================

func (h *EventHandler) ListEvent(c fiber.Ctx) error {
    page, limit, offset := parsePagination(c)

    keyword, errKw := utils.ValidateQueryString(c.Query("search"), 100, "search")
    if errKw != nil {
        return validationErrorResponse(c, errKw)
    }
    areaName, errArea := utils.ValidateQueryString(c.Query("area"), 100, "area")
    if errArea != nil {
        return validationErrorResponse(c, errArea)
    }

    tahunStr := strings.TrimSpace(c.Query("tahun"))
    bulanStr := strings.TrimSpace(c.Query("bulan"))

    cacheKey := fmt.Sprintf("event:list:area_%s:tahun_%s:bulan_%s:keyword_%s:page_%d:limit_%d",
        normalizeKey(areaName), tahunStr, bulanStr, normalizeKey(keyword), page, limit)
    if respondCached(c, cacheKey) {
        return nil
    }

    ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
    defer cancel()

    areaIDArg := pgtype.Int4{Valid: false}
    tahunArg := pgtype.Int4{Valid: false}
    bulanArg := pgtype.Int4{Valid: false}
    keywordArg := pgtype.Text{Valid: false}

    if areaName != "" {
        area, err := h.Queries.GetAreaByName(ctx, areaName) 
        if err == nil {
            areaIDArg = pgtype.Int4{Int32: area.ID, Valid: true}
        }
    }

    if tahunStr != "" {
        if t, err := strconv.Atoi(tahunStr); err == nil && t >= 2000 && t <= 2100 {
            tahunArg = pgtype.Int4{Int32: int32(t), Valid: true}
        }
    }

    if b := parseBulan(bulanStr); b > 0 {
        bulanArg = pgtype.Int4{Int32: int32(b), Valid: true}
    }

    if keyword != "" {
        keywordArg = pgtype.Text{String: keyword, Valid: true}
    }

    listArg := db.ListEventParams{
        AreaID:     areaIDArg,
        Tahun:      tahunArg,
        Bulan:      bulanArg,
        Keyword:    keywordArg,
        LimitData:  int32(limit),
        OffsetData: int32(offset),
    }
    
    countArg := db.CountEventParams{
        AreaID:  areaIDArg,
        Tahun:   tahunArg,
        Bulan:   bulanArg,
        Keyword: keywordArg,
    }

    g, gCtx := errgroup.WithContext(ctx)
    var data []db.ListEventRow
    var total int64

    g.Go(func() error {
        var err error
        data, err = h.Queries.ListEvent(gCtx, listArg)
        return err
    })
    
    g.Go(func() error {
        var err error
        total, err = h.Queries.CountEvent(gCtx, countArg)
        return err
    })

    if err := g.Wait(); err != nil {
        slog.Error("public.event.list_error", slog.String("err", err.Error()))
        return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
            Code:    "ERR_INTERNAL_DB",
            Message: "Gagal mengambil data event",
        })
    }

    if data == nil {
        data = []db.ListEventRow{}
    }

    res := SuccessResponse{
        Pesan:      "Daftar Event",
        Data:       data,
        Pagination: buildPaginationMeta(page, limit, total),
    }
    
    return cacheJSON(c, cacheKey, CacheTTLEventList, res)
}

// =============================================================================
// PUBLIC: GET DETAIL EVENT by SLUG
// =============================================================================

func (h *EventHandler) GetDetailEvent(c fiber.Ctx) error {
    id, err := strconv.Atoi(c.Params("id"))
    if err != nil || id <= 0 {
        return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
            Code: "ERR_VALIDATION", Message: "ID event tidak valid",
        })
    }

    cacheKey := fmt.Sprintf("event:detail:%d", id)
    if respondCached(c, cacheKey) {
        return nil
    }

    ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
    defer cancel()

    data, err := h.Queries.GetEventByIDPublic(ctx, int32(id))
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            return c.Status(fiber.StatusServiceUnavailable).JSON(ErrorResponse{
                Code: "ERR_TIMEOUT", Message: "Server sedang sibuk",
            })
        }
        slog.Warn("public.event.detail_not_found", slog.Int("id", id))
        return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
            Code: "ERR_NOT_FOUND", Message: "Event tidak ditemukan",
        })
    }

    res := SuccessResponse{Pesan: "Detail Event", Data: EventDetailResponse{Event: data}}
    return cacheJSON(c, cacheKey, CacheTTLDetail, res)
}

// =============================================================================
// PUBLIC: GET EVENT MAPS
// =============================================================================

func (h *EventHandler) GetEventMaps(c fiber.Ctx) error {
	areaName, errArea := utils.ValidateQueryString(c.Query("area"), 100, "area")
	if errArea != nil {
		return validationErrorResponse(c, errArea)
	}
	
	keyword, errKw := utils.ValidateQueryString(c.Query("search"), 100, "search")
	if errKw != nil {
		return validationErrorResponse(c, errKw)
	}

	tahunStr := strings.TrimSpace(c.Query("tahun"))

	cacheKey := fmt.Sprintf("event:maps:area_%s:tahun_%s:keyword_%s", normalizeKey(areaName), tahunStr, normalizeKey(keyword))
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	center := MapsCenter{Lat: "-7.697739", Lng: "112.493863", Zoom: 8}

	var areaIDArg pgtype.Int4
	if areaName != "" {
		area, err := h.Queries.GetAreaByName(ctx, areaName)
		if err == nil {
			areaIDArg = pgtype.Int4{Int32: area.ID, Valid: true}
			center.Lat = area.Lat
			center.Lng = area.Lng
			center.Zoom = 11
		}
	}

	var tahunArg pgtype.Int4
	if tahunStr != "" {
		if t, err := strconv.Atoi(tahunStr); err == nil && t >= 2000 && t <= 2100 {
			tahunArg = pgtype.Int4{Int32: int32(t), Valid: true}
		}
	}

	var keywordArg pgtype.Text
	if keyword != "" {
		keywordArg = pgtype.Text{String: keyword, Valid: true}
	}

	data, err := h.Queries.ListEventMaps(ctx, db.ListEventMapsParams{
		AreaID:  areaIDArg,
		Tahun:   tahunArg,
		Keyword: keywordArg, 
	})
	if err != nil {
		slog.Error("public.event.maps_error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code: "ERR_INTERNAL_DB", Message: "Gagal mengambil data peta event",
		})
	}

	if len(data) == 1 && areaName == "" {
		center.Lat = data[0].Lat
		center.Lng = data[0].Lng
		center.Zoom = 13
	}

	if data == nil {
		data = []db.ListEventMapsRow{}
	}

	res := SuccessResponse{
		Pesan: "Peta Event",
		Data:  EventMapsResponse{Center: center, Points: data},
	}
	return cacheJSON(c, cacheKey, CacheTTLMaps, res)
}

// =============================================================================
// PUBLIC: RECOMMENDATION
// =============================================================================

func (h *EventHandler) GetRecommendationEvent(c fiber.Ctx) error {
	cacheKey := "event:recommendation"
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetRecommendationEvent(ctx)
	if err != nil {
		slog.Error("public.event.recommendation_error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code: "ERR_INTERNAL_DB", Message: "Gagal mengambil rekomendasi event",
		})
	}

	if data == nil {
		data = []db.GetRecommendationEventRow{}
	}

	res := SuccessResponse{Pesan: "Rekomendasi Event", Data: EventRecommendationResponse{Items: data}}
	return cacheJSON(c, cacheKey, CacheTTLRecommendation, res)
}

// =============================================================================
// PUBLIC: GET TAHUN TERSEDIA 
// =============================================================================

func (h *EventHandler) GetTahunTersedia(c fiber.Ctx) error {
	cacheKey := "event:tahun-tersedia"
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetAvailableTahunEvent(ctx)
	if err != nil {
		slog.Error("public.event.tahun_error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code: "ERR_INTERNAL_DB", Message: "Gagal mengambil daftar tahun event",
		})
	}

	tahunInt16 := make([]int16, 0, len(data))
	for _, t := range data {
		if t.Valid {
			tahunInt16 = append(tahunInt16, t.Int16)
		}
	}

	res := SuccessResponse{Pesan: "Tahun Event Tersedia", Data: EventTahunTersediaResponse{Tahun: tahunInt16}}
	return cacheJSON(c, cacheKey, CacheTTLRecommendation, res)
}

// =============================================================================
// ADMIN: GET DETAIL by SLUG
// =============================================================================

func (h *EventHandler) GetEventBySlugAdmin(c fiber.Ctx) error {
    slugParam := strings.TrimSpace(c.Params("slug"))
    if slugParam == "" {
        return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
            Code: "ERR_VALIDATION", Message: "Slug event tidak valid",
        })
    }

    ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
    defer cancel()

    data, err := h.Queries.GetEventBySlugAdmin(ctx, slugParam)
    if err != nil {
        slog.Warn("admin.event.not_found", slog.String("slug", slugParam))
        return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
            Code: "ERR_NOT_FOUND", Message: "Event tidak ditemukan",
        })
    }

    return c.JSON(SuccessResponse{Pesan: "Detail Event", Data: data})
}

// =============================================================================
// ADMIN: CREATE EVENT
// =============================================================================

func (h *EventHandler) CreateEvent(c fiber.Ctx) error {
	areaName := strings.TrimSpace(c.FormValue("area"))
	nama := strings.TrimSpace(c.FormValue("nama"))
	deskripsi := strings.TrimSpace(c.FormValue("deskripsi"))
	alamat := strings.TrimSpace(c.FormValue("alamat"))
	tglMulaiStr := strings.TrimSpace(c.FormValue("tanggal_mulai"))
	tglSelesaiStr := strings.TrimSpace(c.FormValue("tanggal_selesai"))
	infoTiket := strings.TrimSpace(c.FormValue("info_tiket"))
	hargaStr := strings.TrimSpace(c.FormValue("harga_tiket"))
	latVal := strings.TrimSpace(c.FormValue("lat"))
	lngVal := strings.TrimSpace(c.FormValue("lng"))

	if err := requireFields(c, map[string]string{
		"area":            areaName,
		"nama":            nama,
		"deskripsi":       deskripsi,
		"alamat":          alamat,
		"tanggal_mulai":   tglMulaiStr,
		"tanggal_selesai": tglSelesaiStr,
		"lat":             latVal,
		"lng":             lngVal,
	}); err != nil {
		return err
	}

	tMulai, errMulai := time.Parse("2006-01-02", tglMulaiStr)
	tSelesai, errSelesai := time.Parse("2006-01-02", tglSelesaiStr)
	if errMulai != nil || errSelesai != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code: "ERR_VALIDATION", Message: "Format tanggal salah (YYYY-MM-DD)",
		})
	}
	if tSelesai.Before(tMulai) {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code: "ERR_VALIDATION", Message: "Tanggal selesai harus >= tanggal mulai",
		})
	}

	if infoTiket == "" {
		infoTiket = "Gratis Umum"
	}

	hargaTiket := 0
	if hargaStr != "" {
		if h, err := strconv.Atoi(hargaStr); err == nil && h >= 0 {
			hargaTiket = h
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextDBTimeout)
	defer cancel()

	area, errArea := h.Queries.GetAreaByName(ctx, areaName)
	if errArea != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Code: "ERR_NOT_FOUND", Message: "Area '" + areaName + "' tidak ditemukan",
		})
	}

	var latPg, lngPg pgtype.Numeric
	if err := latPg.Scan(latVal); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code: "ERR_VALIDATION", Message: "Format latitude tidak valid",
		})
	}
	if err := lngPg.Scan(lngVal); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code: "ERR_VALIDATION", Message: "Format longitude tidak valid",
		})
	}

	exists, errCheck := h.Queries.CheckEventNamaExists(ctx, nama)
	if errCheck != nil {
		slog.Error("admin.event.nama_check_error", slog.String("err", errCheck.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code: "ERR_INTERNAL_DB", Message: "Gagal memvalidasi nama event",
		})
	}
	if exists {
		return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
			Code: "ERR_CONFLICT", Message: "Event dengan nama ini sudah ada",
		})
	}

	slugBaru := utils.GenerateSlug(nama)

	thumbHeader, errT := c.FormFile("gambar_thumbnail")
	if errT != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code: "ERR_VALIDATION", Message: "File gambar_thumbnail wajib diupload",
		})
	}
	heroHeader, errH := c.FormFile("gambar_hero")
	if errH != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code: "ERR_VALIDATION", Message: "File gambar_hero wajib diupload",
		})
	}

	uploadCtx, cancelUpload := context.WithTimeout(context.Background(), ContextUploadTimeout)
	defer cancelUpload()

	thumbURL, thumbPubID, errUpT := uploadImage(uploadCtx, h.Cld, thumbHeader, "sidita_event/thumbnail")
	if errUpT != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code: "ERR_FILE_UPLOAD", Message: "Gagal upload thumbnail: " + errUpT.Error(),
		})
	}
	heroURL, heroPubID, errUpH := uploadImage(uploadCtx, h.Cld, heroHeader, "sidita_event/hero")
	if errUpH != nil {
		destroyImageAsync(h.Cld, thumbPubID)
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code: "ERR_FILE_UPLOAD", Message: "Gagal upload hero: " + errUpH.Error(),
		})
	}

	idBaru, errDb := h.Queries.CreateEvent(ctx, db.CreateEventParams{
		AreaID:               area.ID,
		Nama:                 nama,
		Slug:                 slugBaru,
		Deskripsi:            deskripsi,
		Alamat:               alamat,
		TanggalMulai:         pgtype.Date{Time: tMulai, Valid: true},
		TanggalSelesai:       pgtype.Date{Time: tSelesai, Valid: true},
		InfoTiket:            infoTiket,
		HargaTiket:           int32(hargaTiket),
		GambarUrlThumbnail:   thumbURL,
		GambarUrlHero:        heroURL,
		Lat:                  latPg,
		Lng:                  lngPg,
	})
	if errDb != nil {
		destroyImageAsync(h.Cld, thumbPubID)
		destroyImageAsync(h.Cld, heroPubID)

		var pgErr *pgconn.PgError
		if errors.As(errDb, &pgErr) && pgErr.Code == "23505" {
			return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
				Code: "ERR_CONFLICT", Message: "Event dengan nama atau slug ini sudah ada",
			})
		}
		slog.Error("admin.event.create_error", slog.String("err", errDb.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code: "ERR_INTERNAL_DB", Message: "Gagal menyimpan event",
		})
	}

	invalidateEventCache(idBaru.ID)
	slog.Info("admin.event.created", slog.Int("id", int(idBaru.ID)), slog.String("slug", slugBaru))

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Pesan: "Event berhasil disimpan",
		Data: EventActionResponse{
			ID: idBaru.ID, Nama: nama, Slug: slugBaru,
			GambarUrlThumbnail: thumbURL, GambarUrlHero: heroURL,
		},
	})
}

// =============================================================================
// ADMIN: UPDATE EVENT
// =============================================================================

func (h *EventHandler) UpdateEvent(c fiber.Ctx) error {
    slugParam := strings.TrimSpace(c.Params("slug"))
    if slugParam == "" {
        return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
            Code: "ERR_VALIDATION", Message: "Slug event tidak valid",
        })
    }

    ctx, cancel := context.WithTimeout(context.Background(), ContextDBTimeout)
    defer cancel()

    old, errCari := h.Queries.GetEventBySlugAdmin(ctx, slugParam)
    if errCari != nil {
        return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
            Code: "ERR_NOT_FOUND", Message: "Event tidak ditemukan",
        })
    }

    finalAreaID := old.AreaID
    finalNama := old.Nama
    finalSlug := old.Slug
    finalDeskripsi := old.Deskripsi
    finalAlamat := old.Alamat
    finalTglMulai := old.TanggalMulai
    finalTglSelesai := old.TanggalSelesai
    finalInfoTiket := old.InfoTiket
    finalHarga := old.HargaTiket
    finalLat := old.Lat
    finalLng := old.Lng
    finalThumbURL := old.GambarUrlThumbnail
    finalHeroURL := old.GambarUrlHero

    if v := strings.TrimSpace(c.FormValue("area")); v != "" {
        area, errArea := h.Queries.GetAreaByName(ctx, v)
        if errArea != nil {
            return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
                Code: "ERR_NOT_FOUND", Message: "Area '" + v + "' tidak ditemukan",
            })
        }
        finalAreaID = area.ID
    }
    if v := strings.TrimSpace(c.FormValue("nama")); v != "" {
        finalNama = v
        finalSlug = utils.GenerateSlug(v)
    }
    if v := strings.TrimSpace(c.FormValue("deskripsi")); v != "" {
        finalDeskripsi = v
    }
    if v := strings.TrimSpace(c.FormValue("alamat")); v != "" {
        finalAlamat = v
    }
    if v := strings.TrimSpace(c.FormValue("tanggal_mulai")); v != "" {
        t, err := time.Parse("2006-01-02", v)
        if err != nil {
            return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
                Code: "ERR_VALIDATION", Message: "Format tanggal_mulai salah (YYYY-MM-DD)",
            })
        }
        finalTglMulai = pgtype.Date{Time: t, Valid: true}
    }
    if v := strings.TrimSpace(c.FormValue("tanggal_selesai")); v != "" {
        t, err := time.Parse("2006-01-02", v)
        if err != nil {
            return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
                Code: "ERR_VALIDATION", Message: "Format tanggal_selesai salah (YYYY-MM-DD)",
            })
        }
        finalTglSelesai = pgtype.Date{Time: t, Valid: true}
    }
    if finalTglMulai.Valid && finalTglSelesai.Valid && finalTglSelesai.Time.Before(finalTglMulai.Time) {
        return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
            Code: "ERR_VALIDATION", Message: "Tanggal selesai harus >= tanggal mulai",
        })
    }

    if v := strings.TrimSpace(c.FormValue("info_tiket")); v != "" {
        finalInfoTiket = v
    }
    if v := strings.TrimSpace(c.FormValue("harga_tiket")); v != "" {
        if hg, err := strconv.Atoi(v); err == nil && hg >= 0 {
            finalHarga = int32(hg)
        }
    }
    if v := strings.TrimSpace(c.FormValue("lat")); v != "" {
        if err := finalLat.Scan(v); err != nil {
            return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
                Code: "ERR_VALIDATION", Message: "Format latitude tidak valid",
            })
        }
    }
    if v := strings.TrimSpace(c.FormValue("lng")); v != "" {
        if err := finalLng.Scan(v); err != nil {
            return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
                Code: "ERR_VALIDATION", Message: "Format longitude tidak valid",
            })
        }
    }

    if finalNama != old.Nama {
        exists, errCheck := h.Queries.CheckEventNamaExistsExcluding(ctx, db.CheckEventNamaExistsExcludingParams{
            Nama:    finalNama,
            OldSlug: old.Slug,
        })
        if errCheck != nil {
            slog.Error("admin.event.nama_check_error", slog.String("err", errCheck.Error()))
            return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
                Code: "ERR_INTERNAL_DB", Message: "Gagal memvalidasi nama event",
            })
        }
        if exists {
            return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
                Code: "ERR_CONFLICT", Message: "Event dengan nama ini sudah ada",
            })
        }
    }

    uploadCtx, cancelUpload := context.WithTimeout(context.Background(), ContextUploadTimeout)
    defer cancelUpload()

    var newThumbPubID, newHeroPubID string
    var oldThumbToDelete, oldHeroToDelete string

    if thumbHeader, errT := c.FormFile("gambar_thumbnail"); errT == nil {
        url, pubID, errUp := uploadImage(uploadCtx, h.Cld, thumbHeader, "sidita_event/thumbnail")
        if errUp != nil {
            return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
                Code: "ERR_FILE_UPLOAD", Message: "Gagal upload thumbnail: " + errUp.Error(),
            })
        }
        newThumbPubID = pubID
        oldThumbToDelete = utils.ExtractPublicID(old.GambarUrlThumbnail)
        finalThumbURL = url
    }
    if heroHeader, errH := c.FormFile("gambar_hero"); errH == nil {
        url, pubID, errUp := uploadImage(uploadCtx, h.Cld, heroHeader, "sidita_event/hero")
        if errUp != nil {
            destroyImageAsync(h.Cld, newThumbPubID)
            return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
                Code: "ERR_FILE_UPLOAD", Message: "Gagal upload hero: " + errUp.Error(),
            })
        }
        newHeroPubID = pubID
        oldHeroToDelete = utils.ExtractPublicID(old.GambarUrlHero)
        finalHeroURL = url
    }

    if err := h.Queries.UpdateEvent(ctx, db.UpdateEventParams{
        AreaID:               finalAreaID,
        Nama:                 finalNama,
        SlugBaru:             finalSlug,
        SlugLama:             old.Slug,
        Deskripsi:            finalDeskripsi,
        Alamat:               finalAlamat,
        TanggalMulai:         finalTglMulai,
        TanggalSelesai:       finalTglSelesai,
        InfoTiket:            finalInfoTiket,
        HargaTiket:           finalHarga,
        GambarUrlThumbnail:   finalThumbURL,
        GambarUrlHero:        finalHeroURL,
        Lat:                  finalLat,
        Lng:                  finalLng,
    }); err != nil {
        destroyImageAsync(h.Cld, newThumbPubID)
        destroyImageAsync(h.Cld, newHeroPubID)

        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) && pgErr.Code == "23505" {
            return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
                Code: "ERR_CONFLICT", Message: "Event dengan nama atau slug ini sudah ada",
            })
        }
        slog.Error("admin.event.update_error", slog.String("err", err.Error()))
        return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
            Code: "ERR_INTERNAL_DB", Message: "Gagal mengupdate event",
        })
    }

    destroyImageAsync(h.Cld, oldThumbToDelete)
    destroyImageAsync(h.Cld, oldHeroToDelete)

    invalidateEventCache(old.ID)
    slog.Info("admin.event.updated", slog.Int("id", int(old.ID)), slog.String("slug", finalSlug))

    return c.JSON(SuccessResponse{
        Pesan: "Event berhasil diupdate",
        Data: EventActionResponse{
            ID: old.ID, Nama: finalNama, Slug: finalSlug,
            GambarUrlThumbnail: finalThumbURL, GambarUrlHero: finalHeroURL,
        },
    })
}

// =============================================================================
// ADMIN: DELETE EVENT
// =============================================================================

func (h *EventHandler) DeleteEvent(c fiber.Ctx) error {
    slugParam := strings.TrimSpace(c.Params("slug"))
    if slugParam == "" {
        return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
            Code: "ERR_VALIDATION", Message: "Slug event tidak valid",
        })
    }

    ctx, cancel := context.WithTimeout(context.Background(), ContextDBTimeout)
    defer cancel()

    old, errCari := h.Queries.GetEventBySlugAdmin(ctx, slugParam)
    if errCari != nil {
        return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
            Code: "ERR_NOT_FOUND", Message: "Event tidak ditemukan",
        })
    }

    if err := h.Queries.DeleteEvent(ctx, old.Slug); err != nil {
        slog.Error("admin.event.delete_error", slog.String("err", err.Error()))
        return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
            Code: "ERR_INTERNAL_DB", Message: "Gagal menghapus event",
        })
    }

    destroyImageAsync(h.Cld, utils.ExtractPublicID(old.GambarUrlThumbnail))
    destroyImageAsync(h.Cld, utils.ExtractPublicID(old.GambarUrlHero))

    invalidateEventCache(old.ID)
    slog.Info("admin.event.deleted", slog.Int("id", int(old.ID)), slog.String("slug", old.Slug))

    return c.JSON(SuccessResponse{Pesan: "Event berhasil dihapus"})
}

// =============================================================================
// CACHE WARMUP
// =============================================================================

func (h *EventHandler) CacheWarmup() {
	slog.Info("cache.warmup.event.start")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if data, err := h.Queries.GetRecommendationEvent(ctx); err == nil {
		if data == nil {
			data = []db.GetRecommendationEventRow{}
		}
		cache.GlobalCache.Set("event:recommendation",
			SuccessResponse{Pesan: "Rekomendasi Event", Data: EventRecommendationResponse{Items: data}},
			CacheTTLRecommendation)
	}

	if tahun, err := h.Queries.GetAvailableTahunEvent(ctx); err == nil {
    tahunInt16 := make([]int16, 0, len(tahun))
    for _, t := range tahun {
        if t.Valid {
            tahunInt16 = append(tahunInt16, t.Int16)
        }
    }
    cache.GlobalCache.Set("event:tahun-tersedia",
        SuccessResponse{Pesan: "Tahun Event Tersedia", Data: EventTahunTersediaResponse{Tahun: tahunInt16}},
        CacheTTLRecommendation)
	}

	g, gCtx := errgroup.WithContext(ctx)
	var listData []db.ListEventRow
	var total int64

	g.Go(func() error {
		var err error
		listData, err = h.Queries.ListEvent(gCtx, db.ListEventParams{LimitData: 10, OffsetData: 0})
		return err
	})
	g.Go(func() error {
		var err error
		total, err = h.Queries.CountEvent(gCtx, db.CountEventParams{})
		return err
	})

	if err := g.Wait(); err == nil {
		if listData == nil {
			listData = []db.ListEventRow{}
		}
		warmupKey := "event:list:area_:tahun_:bulan_:keyword_:page_1:limit_10"
		cache.GlobalCache.Set(warmupKey,
			SuccessResponse{
				Pesan:      "Daftar Event",
				Data:       listData,
				Pagination: buildPaginationMeta(1, 10, total),
			},
			CacheTTLEventList)
	}

	slog.Info("cache.warmup.event.done")
}