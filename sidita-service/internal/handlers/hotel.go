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

type HotelHandler struct {
	Queries *db.Queries
	Cld     *cloudinary.Cloudinary
}

func NewHotelHandler(q *db.Queries, cld *cloudinary.Cloudinary) *HotelHandler {
	return &HotelHandler{Queries: q, Cld: cld}
}

func invalidateHotelCache(id int32) {
	cache.GlobalCache.DeleteByPrefix("hotel:list:")
	cache.GlobalCache.DeleteByPrefix("hotel:maps:")
	cache.GlobalCache.Delete("hotel:recommendation")
	if id > 0 {
		cache.GlobalCache.Delete(fmt.Sprintf("hotel:detail:id_%d", id))
	}
}

// =============================================================================
// PUBLIC: LIST HOTEL
// =============================================================================

func (h *HotelHandler) ListHotel(c fiber.Ctx) error {
	page, limit, offset := parsePagination(c)

	keyword, errKw := utils.ValidateQueryString(c.Query("search"), 100, "search")
	if errKw != nil {
		return validationErrorResponse(c, errKw)
	}
	areaName, errArea := utils.ValidateQueryString(c.Query("area"), 100, "area")
	if errArea != nil {
		return validationErrorResponse(c, errArea)
	}

	cacheKey := fmt.Sprintf("hotel:list:area_%s:keyword_%s:page_%d:limit_%d",
		normalizeKey(areaName), normalizeKey(keyword), page, limit)
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	var areaIDArg pgtype.Int4
	if areaName != "" {
		area, err := h.Queries.GetAreaByName(ctx, areaName)
		if err == nil {
			areaIDArg = pgtype.Int4{Int32: area.ID, Valid: true}
		} else {
			areaIDArg = pgtype.Int4{Int32: -1, Valid: true}
		}
	}

	var keywordArg pgtype.Text
	if keyword != "" {
		keywordArg = pgtype.Text{String: keyword, Valid: true}
	}

	listArg := db.ListHotelParams{
		AreaID:     areaIDArg,
		Keyword:    keywordArg,
		LimitData:  int32(limit),
		OffsetData: int32(offset),
	}
	countArg := db.CountHotelParams{
		AreaID:  areaIDArg,
		Keyword: keywordArg,
	}

	g, gCtx := errgroup.WithContext(ctx)
	var data []db.ListHotelRow
	var total int64

	g.Go(func() error {
		var err error
		data, err = h.Queries.ListHotel(gCtx, listArg)
		return err
	})
	g.Go(func() error {
		var err error
		total, err = h.Queries.CountHotel(gCtx, countArg)
		return err
	})

	if err := g.Wait(); err != nil {
		slog.Error("public.hotel.list_error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal mengambil data hotel",
		})
	}

	if data == nil {
		data = []db.ListHotelRow{}
	}

	res := SuccessResponse{
		Pesan:      "Daftar Hotel",
		Data:       data,
		Pagination: buildPaginationMeta(page, limit, total),
	}
	return cacheJSON(c, cacheKey, CacheTTLList, res)
}

// =============================================================================
// PUBLIC: GET DETAIL by ID
// =============================================================================

func (h *HotelHandler) GetHotelByIDPublic(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "ID hotel tidak valid, harus berupa angka",
		})
	}

	cacheKey := fmt.Sprintf("hotel:detail:id_%d", id)
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetHotelByIDPublic(ctx, int32(id))
	if err != nil {
		slog.Warn("public.hotel.not_found", slog.Int("id", id))
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Code:    "ERR_NOT_FOUND",
			Message: "Hotel tidak ditemukan",
		})
	}

	res := SuccessResponse{
		Pesan: "Detail Hotel",
		Data:  data,
	}
	
	return cacheJSON(c, cacheKey, CacheTTLDetail, res)
}

// =============================================================================
// PUBLIC: GET HOTEL MAPS
// =============================================================================

func (h *HotelHandler) GetHotelMaps(c fiber.Ctx) error {
	areaName, errArea := utils.ValidateQueryString(c.Query("area"), 100, "area")
	if errArea != nil {
		return validationErrorResponse(c, errArea)
	}
	keyword, errKw := utils.ValidateQueryString(c.Query("search"), 100, "search")
	if errKw != nil {
		return validationErrorResponse(c, errKw)
	}

	cacheKey := fmt.Sprintf("hotel:maps:area_%s:keyword_%s", normalizeKey(areaName), normalizeKey(keyword))
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

	var keywordArg pgtype.Text
	if keyword != "" {
		keywordArg = pgtype.Text{String: keyword, Valid: true}
	}

	data, err := h.Queries.ListHotelMaps(ctx, db.ListHotelMapsParams{
		AreaID:  areaIDArg,
		Keyword: keywordArg,
	})
	if err != nil {
		slog.Error("public.hotel.maps_error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal mengambil data peta hotel",
		})
	}

	if len(data) == 1 && areaName == "" {
		center.Lat = data[0].Lat
		center.Lng = data[0].Lng
		center.Zoom = 13
	}

	if data == nil {
		data = []db.ListHotelMapsRow{}
	}

	res := SuccessResponse{
		Pesan: "Peta Hotel",
		Data:  HotelMapsResponse{Center: center, Points: data},
	}
	return cacheJSON(c, cacheKey, CacheTTLMaps, res)
}

// =============================================================================
// PUBLIC: RECOMMENDATION
// =============================================================================

func (h *HotelHandler) GetRecommendationHotel(c fiber.Ctx) error {
	cacheKey := "hotel:recommendation"
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetRecommendationHotel(ctx)
	if err != nil {
		slog.Error("public.hotel.recommendation_error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal mengambil rekomendasi hotel",
		})
	}

	if data == nil {
		data = []db.GetRecommendationHotelRow{}
	}

	res := SuccessResponse{Pesan: "Rekomendasi Hotel", Data: HotelRecommendationResponse{Items: data}}
	return cacheJSON(c, cacheKey, CacheTTLRecommendation, res)
}

// =============================================================================
// ADMIN: GET DETAIL 
// =============================================================================

func (h *HotelHandler) GetHotelBySlugAdmin(c fiber.Ctx) error {
    slugParam := c.Params("slug")
    if slugParam == "" {
        return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
            Code:    "ERR_VALIDATION",
            Message: "Slug hotel tidak valid",
        })
    }

    ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
    defer cancel()

    data, err := h.Queries.GetHotelBySlugAdmin(ctx, slugParam)
    if err != nil {
        slog.Warn("admin.hotel.not_found", slog.String("slug", slugParam))
        return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
            Code:    "ERR_NOT_FOUND",
            Message: "Hotel tidak ditemukan",
        })
    }

    return c.JSON(SuccessResponse{Pesan: "Detail Hotel", Data: data})
}

// =============================================================================
// ADMIN: CREATE HOTEL
// =============================================================================

func (h *HotelHandler) CreateHotel(c fiber.Ctx) error {
	areaName := strings.TrimSpace(c.FormValue("area"))
	nama := strings.TrimSpace(c.FormValue("nama"))
	bintangStr := strings.TrimSpace(c.FormValue("bintang"))
	hargaStr := strings.TrimSpace(c.FormValue("harga_mulai"))
	deskripsi := strings.TrimSpace(c.FormValue("deskripsi"))
	alamat := strings.TrimSpace(c.FormValue("alamat"))
	highlight := strings.TrimSpace(c.FormValue("highlight_text"))
	latVal := strings.TrimSpace(c.FormValue("lat"))
	lngVal := strings.TrimSpace(c.FormValue("lng"))

	if err := requireFields(c, map[string]string{
		"area":        areaName,
		"nama":        nama,
		"bintang":     bintangStr,
		"harga_mulai": hargaStr,
		"alamat":      alamat,
		"lat":         latVal,
		"lng":         lngVal,
	}); err != nil {
		return err
	}

	bintang, errB := strconv.Atoi(bintangStr)
	if errB != nil || bintang < 1 || bintang > 5 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Bintang harus berupa angka 1-5",
		})
	}
	harga, errH := strconv.Atoi(hargaStr)
	if errH != nil || harga < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Harga mulai harus berupa angka non-negatif",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextDBTimeout)
	defer cancel()

	area, errArea := h.Queries.GetAreaByName(ctx, areaName)
	if errArea != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Code:    "ERR_NOT_FOUND",
			Message: "Area '" + areaName + "' tidak ditemukan",
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

	exists, errCheck := h.Queries.CheckHotelNamaExists(ctx, nama)
	if errCheck != nil {
		slog.Error("admin.hotel.nama_check_error", slog.String("err", errCheck.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal memvalidasi nama hotel",
		})
	}
	if exists {
		return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
			Code:    "ERR_CONFLICT",
			Message: "Hotel dengan nama ini sudah ada",
		})
	}

	slugBaru := utils.GenerateSlug(nama)

	fileHeader, errFile := c.FormFile("gambar")
	if errFile != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "File gambar wajib diupload",
		})
	}

	uploadCtx, cancelUpload := context.WithTimeout(context.Background(), ContextUploadTimeout)
	defer cancelUpload()

	gambarURL, gambarPubID, errUp := uploadImage(uploadCtx, h.Cld, fileHeader, "sidita_hotel")
	if errUp != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_FILE_UPLOAD",
			Message: "Gagal upload gambar: " + errUp.Error(),
		})
	}

	idBaru, errDb := h.Queries.CreateHotel(ctx, db.CreateHotelParams{
		AreaID:        area.ID,
		Nama:          nama,
		Slug:          slugBaru,
		Bintang:       int16(bintang),
		HargaMulai:    int32(harga),
		Deskripsi:     pgtype.Text{String: deskripsi, Valid: deskripsi != ""},
		Alamat:        alamat,
		HighlightText: pgtype.Text{String: highlight, Valid: highlight != ""},
		GambarUrl:     gambarURL,
		Lat:           latPg,
		Lng:           lngPg,
	})
	if errDb != nil {
		destroyImageAsync(h.Cld, gambarPubID)

		var pgErr *pgconn.PgError
		if errors.As(errDb, &pgErr) && pgErr.Code == "23505" {
			return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
				Code:    "ERR_CONFLICT",
				Message: "Hotel dengan nama atau slug ini sudah ada",
			})
		}
		slog.Error("admin.hotel.create_error", slog.String("err", errDb.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal menyimpan hotel",
		})
	}

	invalidateHotelCache(idBaru.ID)
	slog.Info("admin.hotel.created", slog.Int("id", int(idBaru.ID)), slog.String("slug", slugBaru))

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Pesan: "Hotel berhasil disimpan",
		Data: HotelActionResponse{
			ID:         idBaru.ID,
			Nama:       nama,
			Slug:       slugBaru,
			Kota:       area.Nama,
			Bintang:    int16(bintang),
			HargaMulai: int32(harga),
			Lat:        latVal,
			Lng:        lngVal,
			GambarUrl:  gambarURL,
		},
	})
}

// =============================================================================
// ADMIN: UPDATE HOTEL
// =============================================================================

func (h *HotelHandler) UpdateHotel(c fiber.Ctx) error {
    slugParam := c.Params("slug")
    if slugParam == "" {
        return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
            Code: "ERR_VALIDATION", Message: "Slug hotel tidak valid",
        })
    }

    ctx, cancel := context.WithTimeout(context.Background(), ContextDBTimeout)
    defer cancel()

    old, errCari := h.Queries.GetHotelBySlugAdmin(ctx, slugParam)
    if errCari != nil {
        return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
            Code: "ERR_NOT_FOUND", Message: "Hotel tidak ditemukan",
        })
    }

    finalAreaID := old.AreaID
    finalNama := old.Nama
    finalSlug := old.Slug
    finalBintang := old.Bintang
    finalHarga := old.HargaMulai
    finalDeskripsi := old.Deskripsi
    finalAlamat := old.Alamat
    finalHighlight := old.HighlightText
    finalLat := old.Lat
    finalLng := old.Lng
    finalGambar := old.GambarUrl

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
    if v := strings.TrimSpace(c.FormValue("bintang")); v != "" {
        b, err := strconv.Atoi(v)
        if err != nil || b < 1 || b > 5 {
            return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
                Code: "ERR_VALIDATION", Message: "Bintang harus berupa angka 1-5",
            })
        }
        finalBintang = int16(b)
    }
    if v := strings.TrimSpace(c.FormValue("harga_mulai")); v != "" {
        hg, err := strconv.Atoi(v)
        if err != nil || hg < 0 {
            return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
                Code: "ERR_VALIDATION", Message: "Harga mulai harus berupa angka non-negatif",
            })
        }
        finalHarga = int32(hg)
    }
    if v := strings.TrimSpace(c.FormValue("deskripsi")); v != "" {
        finalDeskripsi = pgtype.Text{String: v, Valid: true}
    }
    if v := strings.TrimSpace(c.FormValue("alamat")); v != "" {
        finalAlamat = v
    }
    if v := strings.TrimSpace(c.FormValue("highlight_text")); v != "" {
        finalHighlight = pgtype.Text{String: v, Valid: true}
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
        exists, errCheck := h.Queries.CheckHotelNamaExistsExcluding(ctx, db.CheckHotelNamaExistsExcludingParams{
            Nama:    finalNama,
            OldSlug: old.Slug,
        })
        if errCheck != nil {
            slog.Error("admin.hotel.nama_check_error", slog.String("err", errCheck.Error()))
            return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
                Code:    "ERR_INTERNAL_DB",
                Message: "Gagal memvalidasi nama hotel",
            })
        }
        if exists {
            return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
                Code:    "ERR_CONFLICT",
                Message: "Hotel dengan nama ini sudah ada",
            })
        }
    }

    uploadCtx, cancelUpload := context.WithTimeout(context.Background(), ContextUploadTimeout)
    defer cancelUpload()

    var newPubID, oldToDelete string
    if fileHeader, errFile := c.FormFile("gambar"); errFile == nil {
        url, pubID, errUp := uploadImage(uploadCtx, h.Cld, fileHeader, "sidita_hotel")
        if errUp != nil {
            return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
                Code:    "ERR_FILE_UPLOAD",
                Message: "Gagal upload gambar: " + errUp.Error(),
            })
        }
        newPubID = pubID
        oldToDelete = utils.ExtractPublicID(old.GambarUrl)
        finalGambar = url
    }

    if err := h.Queries.UpdateHotel(ctx, db.UpdateHotelParams{
        AreaID:        finalAreaID,
        Nama:          finalNama,
        SlugBaru:      finalSlug,
        SlugLama:      old.Slug,
        Bintang:       finalBintang,
        HargaMulai:    finalHarga,
        Deskripsi:     finalDeskripsi,
        Alamat:        finalAlamat,
        HighlightText: finalHighlight,
        GambarUrl:     finalGambar,
        Lat:           finalLat,
        Lng:           finalLng,
    }); err != nil {
        destroyImageAsync(h.Cld, newPubID)

        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) && pgErr.Code == "23505" {
            return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
                Code:    "ERR_CONFLICT",
                Message: "Hotel dengan nama atau slug ini sudah ada",
            })
        }
        slog.Error("admin.hotel.update_error", slog.String("err", err.Error()))
        return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
            Code:    "ERR_INTERNAL_DB",
            Message: "Gagal mengupdate hotel",
        })
    }

    destroyImageAsync(h.Cld, oldToDelete)

    invalidateHotelCache(old.ID)
    slog.Info("admin.hotel.updated", slog.Int("id", int(old.ID)), slog.String("slug", finalSlug))

    return c.JSON(SuccessResponse{
        Pesan: "Hotel berhasil diupdate",
        Data:  HotelActionResponse{ID: old.ID, Nama: finalNama, Slug: finalSlug, GambarUrl: finalGambar},
    })
}

// =============================================================================
// ADMIN: DELETE HOTEL
// =============================================================================

func (h *HotelHandler) DeleteHotel(c fiber.Ctx) error {
    slugParam := c.Params("slug")
    if slugParam == "" {
        return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
            Code: "ERR_VALIDATION", Message: "Slug hotel tidak valid",
        })
    }

    ctx, cancel := context.WithTimeout(context.Background(), ContextDBTimeout)
    defer cancel()

    old, errCari := h.Queries.GetHotelBySlugAdmin(ctx, slugParam)
    if errCari != nil {
        return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
            Code: "ERR_NOT_FOUND", Message: "Hotel tidak ditemukan",
        })
    }

    if err := h.Queries.DeleteHotel(ctx, old.Slug); err != nil {
        slog.Error("admin.hotel.delete_error", slog.String("err", err.Error()))
        return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
            Code: "ERR_INTERNAL_DB", Message: "Gagal menghapus hotel",
        })
    }

    destroyImageAsync(h.Cld, utils.ExtractPublicID(old.GambarUrl))

    invalidateHotelCache(old.ID)
    slog.Info("admin.hotel.deleted", slog.Int("id", int(old.ID)), slog.String("slug", old.Slug))

    return c.JSON(SuccessResponse{Pesan: "Hotel berhasil dihapus"})
}

// =============================================================================
// CACHE WARMUP
// =============================================================================

func (h *HotelHandler) CacheWarmup() {
	slog.Info("cache.warmup.hotel.start")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if data, err := h.Queries.GetRecommendationHotel(ctx); err == nil {
		if data == nil {
			data = []db.GetRecommendationHotelRow{}
		}
		cache.GlobalCache.Set("hotel:recommendation",
			SuccessResponse{Pesan: "Rekomendasi Hotel", Data: HotelRecommendationResponse{Items: data}},
			CacheTTLRecommendation)
	}

	g, gCtx := errgroup.WithContext(ctx)
	var listData []db.ListHotelRow
	var total int64

	g.Go(func() error {
		var err error
		listData, err = h.Queries.ListHotel(gCtx, db.ListHotelParams{LimitData: 10, OffsetData: 0})
		return err
	})
	g.Go(func() error {
		var err error
		total, err = h.Queries.CountHotel(gCtx, db.CountHotelParams{})
		return err
	})

	if err := g.Wait(); err == nil {
		if listData == nil {
			listData = []db.ListHotelRow{}
		}
		warmupKey := "hotel:list:area_:keyword_:page_1:limit_10"
		cache.GlobalCache.Set(warmupKey,
			SuccessResponse{
				Pesan:      "Daftar Hotel",
				Data:       listData,
				Pagination: buildPaginationMeta(1, 10, total),
			},
			CacheTTLList)
	}

	slog.Info("cache.warmup.hotel.done")
}