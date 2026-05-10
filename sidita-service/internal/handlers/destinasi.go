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

type DestinasiHandler struct {
	Queries *db.Queries
	Cld     *cloudinary.Cloudinary
}

func NewDestinasiHandler(q *db.Queries, cld *cloudinary.Cloudinary) *DestinasiHandler {
	return &DestinasiHandler{Queries: q, Cld: cld}
}

// =============================================================================
// CACHE INVALIDATION HELPERS
// =============================================================================

func invalidateDestinasiCache(id int32, slugs ...string) {
	cache.GlobalCache.DeleteByPrefix("destinasi:list:")
	cache.GlobalCache.DeleteByPrefix("destinasi:maps:")
	cache.GlobalCache.Delete("destinasi:recommendation")
	
	if id > 0 {
		cache.GlobalCache.Delete(fmt.Sprintf("destinasi:detail:id:%d", id))
	}
	for _, s := range slugs {
		if s != "" {
			cache.GlobalCache.Delete("destinasi:detail:" + s)
		}
	}
}

// =============================================================================
// PUBLIC: GET ALL AREA 
// =============================================================================

func (h *DestinasiHandler) GetAllArea(c fiber.Ctx) error {
	cacheKey := "areas:all"
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	areas, err := h.Queries.GetAllArea(ctx)
	if err != nil {
		slog.Error("public.areas.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal mengambil daftar area",
		})
	}

	res := SuccessResponse{Pesan: "Daftar Area", Data: areas}
	return cacheJSON(c, cacheKey, CacheTTLArea, res)
}

// =============================================================================
// PUBLIC: LIST DESTINASI 
// =============================================================================

func (h *DestinasiHandler) ListDestinasi(c fiber.Ctx) error {
	page, limit, offset := parsePagination(c)

	keyword, errKw := utils.ValidateQueryString(c.Query("search"), 100, "search")
	if errKw != nil {
		return validationErrorResponse(c, errKw)
	}

	areaName, errArea := utils.ValidateQueryString(c.Query("area"), 100, "area")
	if errArea != nil {
		return validationErrorResponse(c, errArea)
	}

	cacheKey := fmt.Sprintf("destinasi:list:area_%s:keyword_%s:page_%d:limit_%d",
		normalizeKey(areaName), normalizeKey(keyword), page, limit)
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	var areaIDArg pgtype.Int4
	if areaName != "" {
		area, err := h.Queries.GetAreaByName(ctx, areaName)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Code:    "ERR_AREA_NOT_FOUND",
				Message: "Area tidak ditemukan",
			})
		}
		areaIDArg = pgtype.Int4{Int32: area.ID, Valid: true}
	}

	var keywordArg pgtype.Text
	if keyword != "" {
		keywordArg = pgtype.Text{String: keyword, Valid: true}
	}

	listArg := db.ListDestinasiParams{
		AreaID:     areaIDArg,
		Keyword:    keywordArg,
		LimitData:  int32(limit),
		OffsetData: int32(offset),
	}
	countArg := db.CountDestinasiParams{
		AreaID:  areaIDArg,
		Keyword: keywordArg,
	}

	g, gCtx := errgroup.WithContext(ctx)
	var data []db.ListDestinasiRow
	var total int64

	g.Go(func() error {
		var err error
		data, err = h.Queries.ListDestinasi(gCtx, listArg)
		return err
	})
	g.Go(func() error {
		var err error
		total, err = h.Queries.CountDestinasi(gCtx, countArg)
		return err
	})

	if err := g.Wait(); err != nil {
		slog.Error("public.destinasi.list_error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal mengambil data destinasi",
		})
	}

	if data == nil {
		data = []db.ListDestinasiRow{}
	}

	res := SuccessResponse{
		Pesan:      "Daftar Destinasi",
		Data:       data,
		Pagination: buildPaginationMeta(page, limit, total),
	}
	return cacheJSON(c, cacheKey, CacheTTLList, res)
}

// =============================================================================
// PUBLIC / MOBILE: GET DETAIL DESTINASI by ID
// =============================================================================

func (h *DestinasiHandler) GetDetailDestinasi(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "ID destinasi tidak valid",
		})
	}

	cacheKey := fmt.Sprintf("destinasi:detail:id:%d", id)
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetDestinasiByIDPublic(ctx, int32(id))
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return c.Status(fiber.StatusServiceUnavailable).JSON(ErrorResponse{
				Code:    "ERR_TIMEOUT",
				Message: "Server sedang sibuk",
			})
		}
		slog.Warn("public.destinasi.detail_not_found", slog.Int("id", id))
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Code:    "ERR_NOT_FOUND",
			Message: "Destinasi tidak ditemukan",
		})
	}

	res := SuccessResponse{Pesan: "Detail Destinasi", Data: data}
	return cacheJSON(c, cacheKey, CacheTTLDetail, res)
}

// =============================================================================
// PUBLIC: GET DESTINASI MAPS 
// =============================================================================

func (h *DestinasiHandler) GetDestinasiMaps(c fiber.Ctx) error {
    areaName, _ := utils.ValidateQueryString(c.Query("area"), 100, "area")
    keyword, _ := utils.ValidateQueryString(c.Query("search"), 100, "search")

    cacheKey := fmt.Sprintf("destinasi:maps:area_%s:keyword_%s", normalizeKey(areaName), normalizeKey(keyword))
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

    data, err := h.Queries.ListDestinasiMaps(ctx, db.ListDestinasiMapsParams{
        AreaID:  areaIDArg,
        Keyword: keywordArg,
    })
	if err != nil {
		slog.Error("public.destinasi.maps_error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal mengambil data peta destinasi",
		})
	}

	if len(data) == 1 && areaName == "" {
        center.Lat = data[0].Lat
        center.Lng = data[0].Lng
        center.Zoom = 13 
    }

	if data == nil {
		data = []db.ListDestinasiMapsRow{}
	}

	res := SuccessResponse{
		Pesan: "Peta Destinasi",
		Data: DestinasiMapsResponse{
			Center: center,
			Points: data,
		},
	}
	return cacheJSON(c, cacheKey, CacheTTLMaps, res)
}

// =============================================================================
// PUBLIC: RECOMMENDATION 
// =============================================================================

func (h *DestinasiHandler) GetRecommendationDestinasi(c fiber.Ctx) error {
	cacheKey := "destinasi:recommendation"
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetRecommendationDestinasi(ctx)
	if err != nil {
		slog.Error("public.destinasi.recommendation_error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal mengambil rekomendasi destinasi",
		})
	}

	if data == nil {
		data = []db.GetRecommendationDestinasiRow{}
	}

	res := SuccessResponse{Pesan: "Rekomendasi Destinasi", Data: DestinasiRecommendationResponse{Items: data}}
	return cacheJSON(c, cacheKey, CacheTTLRecommendation, res)
}

// =============================================================================
// ADMIN: GET DETAIL by SLUG 
// =============================================================================

func (h *DestinasiHandler) GetDestinasiBySlugAdmin(c fiber.Ctx) error {
	slugParam := strings.TrimSpace(c.Params("slug"))
	if slugParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Slug destinasi tidak valid",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	data, err := h.Queries.GetDestinasiBySlugAdmin(ctx, slugParam)
	if err != nil {
		slog.Warn("admin.destinasi.not_found", slog.String("slug", slugParam))
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Code:    "ERR_NOT_FOUND",
			Message: "Destinasi tidak ditemukan",
		})
	}

	return c.JSON(SuccessResponse{Pesan: "Detail Destinasi", Data: data})
}

// =============================================================================
// ADMIN: CREATE DESTINASI
// =============================================================================

func (h *DestinasiHandler) CreateDestinasi(c fiber.Ctx) error {
	areaName := strings.TrimSpace(c.FormValue("area"))
	kategori := strings.TrimSpace(c.FormValue("kategori"))
	nama := strings.TrimSpace(c.FormValue("nama"))
	deskripsi := strings.TrimSpace(c.FormValue("deskripsi"))
	alamat := strings.TrimSpace(c.FormValue("alamat"))
	highlight := strings.TrimSpace(c.FormValue("highlight_text"))
	latVal := strings.TrimSpace(c.FormValue("lat"))
	lngVal := strings.TrimSpace(c.FormValue("lng"))

	if err := requireFields(c, map[string]string{
		"area":      areaName,
		"kategori":  kategori,
		"nama":      nama,
		"deskripsi": deskripsi,
		"alamat":    alamat,
		"lat":       latVal,
		"lng":       lngVal,
	}); err != nil {
		return err
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
			Code:    "ERR_VALIDATION",
			Message: "Format latitude tidak valid",
		})
	}
	if err := lngPg.Scan(lngVal); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Format longitude tidak valid",
		})
	}

	exists, errCheck := h.Queries.CheckDestinasiNamaExists(ctx, nama)
	if errCheck != nil {
		slog.Error("admin.destinasi.nama_check_error", slog.String("err", errCheck.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal memvalidasi nama destinasi",
		})
	}
	if exists {
		return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
			Code:    "ERR_CONFLICT",
			Message: "Destinasi dengan nama ini sudah ada",
		})
	}

	slugBaru := utils.GenerateSlug(nama)

	thumbHeader, errT := c.FormFile("gambar_thumbnail")
	if errT != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "File gambar_thumbnail wajib diupload",
		})
	}
	heroHeader, errH := c.FormFile("gambar_hero")
	if errH != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "File gambar_hero wajib diupload",
		})
	}

	uploadCtx, cancelUpload := context.WithTimeout(context.Background(), ContextUploadTimeout)
	defer cancelUpload()

	thumbURL, thumbPubID, errUpT := uploadImage(uploadCtx, h.Cld, thumbHeader, "sidita_destinasi/thumbnail")
	if errUpT != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_FILE_UPLOAD",
			Message: "Gagal upload thumbnail: " + errUpT.Error(),
		})
	}
	heroURL, heroPubID, errUpH := uploadImage(uploadCtx, h.Cld, heroHeader, "sidita_destinasi/hero")
	if errUpH != nil {
		destroyImageAsync(h.Cld, thumbPubID)
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_FILE_UPLOAD",
			Message: "Gagal upload hero: " + errUpH.Error(),
		})
	}

	idBaru, errDb := h.Queries.CreateDestinasi(ctx, db.CreateDestinasiParams{
		AreaID:             area.ID,
		Nama:               nama,
		Slug:               slugBaru,
		Kategori:           kategori,
		Deskripsi:          deskripsi,
		Alamat:             alamat,
		HighlightText:      pgtype.Text{String: highlight, Valid: highlight != ""},
		GambarUrlThumbnail: thumbURL,
		GambarUrlHero:      heroURL,
		Lat:                latPg,
		Lng:                lngPg,
	})
	if errDb != nil {
		destroyImageAsync(h.Cld, thumbPubID)
		destroyImageAsync(h.Cld, heroPubID)

		var pgErr *pgconn.PgError
		if errors.As(errDb, &pgErr) && pgErr.Code == "23505" {
			return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
				Code:    "ERR_CONFLICT",
				Message: "Destinasi dengan nama atau slug ini sudah ada",
			})
		}
		slog.Error("admin.destinasi.create_error", slog.String("err", errDb.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal menyimpan destinasi",
		})
	}

	invalidateDestinasiCache(idBaru.ID, slugBaru)
	slog.Info("admin.destinasi.created", slog.Int("id", int(idBaru.ID)), slog.String("slug", slugBaru))

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Pesan: "Destinasi berhasil disimpan",
		Data: DestinasiActionResponse{
			ID:                 idBaru.ID,
			Nama:               nama,
			Slug:               slugBaru,
			Kota:               area.Nama,
			Lat:                latVal,
			Lng:                lngVal,
			GambarUrlThumbnail: thumbURL,
			GambarUrlHero:      heroURL,
		},
	})
}

// =============================================================================
// ADMIN: UPDATE DESTINASI 
// =============================================================================

func (h *DestinasiHandler) UpdateDestinasi(c fiber.Ctx) error {
	slugParam := strings.TrimSpace(c.Params("slug"))
	if slugParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Slug destinasi tidak valid",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextDBTimeout)
	defer cancel()

	old, errCari := h.Queries.GetDestinasiBySlugAdmin(ctx, slugParam)
	if errCari != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Code:    "ERR_NOT_FOUND",
			Message: "Destinasi tidak ditemukan",
		})
	}

	finalAreaID := old.AreaID
	finalKategori := old.Kategori
	finalNama := old.Nama
	finalSlug := old.Slug
	finalDeskripsi := old.Deskripsi
	finalAlamat := old.Alamat
	finalHighlight := old.HighlightText
	finalLat := old.Lat
	finalLng := old.Lng
	finalThumbURL := old.GambarUrlThumbnail
	finalHeroURL := old.GambarUrlHero

	if v := strings.TrimSpace(c.FormValue("area")); v != "" {
		area, errArea := h.Queries.GetAreaByName(ctx, v)
		if errArea != nil {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Code:    "ERR_NOT_FOUND",
				Message: "Area '" + v + "' tidak ditemukan",
			})
		}
		finalAreaID = area.ID
	}

	if v := strings.TrimSpace(c.FormValue("kategori")); v != "" {
		finalKategori = v
	}
	if v := strings.TrimSpace(c.FormValue("nama")); v != "" {
		finalNama = v
		// Generate slug baru jika nama diubah
		finalSlug = utils.GenerateSlug(v) 
	}
	if v := strings.TrimSpace(c.FormValue("deskripsi")); v != "" {
		finalDeskripsi = v
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
		exists, errCheck := h.Queries.CheckDestinasiNamaExistsExcluding(ctx, db.CheckDestinasiNamaExistsExcludingParams{
			Nama:    finalNama,
			OldSlug: old.Slug,
		})
		if errCheck != nil {
			slog.Error("admin.destinasi.nama_check_error", slog.String("err", errCheck.Error()))
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
				Code:    "ERR_INTERNAL_DB",
				Message: "Gagal memvalidasi nama destinasi",
			})
		}
		if exists {
			return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
				Code:    "ERR_CONFLICT",
				Message: "Destinasi dengan nama ini sudah ada",
			})
		}
	}

	uploadCtx, cancelUpload := context.WithTimeout(context.Background(), ContextUploadTimeout)
	defer cancelUpload()

	var newThumbPubID, newHeroPubID string
	var oldThumbToDelete, oldHeroToDelete string

	if thumbHeader, errT := c.FormFile("gambar_thumbnail"); errT == nil {
		url, pubID, errUp := uploadImage(uploadCtx, h.Cld, thumbHeader, "sidita_destinasi/thumbnail")
		if errUp != nil {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Code:    "ERR_FILE_UPLOAD",
				Message: "Gagal upload thumbnail: " + errUp.Error(),
			})
		}
		newThumbPubID = pubID
		oldThumbToDelete = utils.ExtractPublicID(old.GambarUrlThumbnail)
		finalThumbURL = url
	}
	if heroHeader, errH := c.FormFile("gambar_hero"); errH == nil {
		url, pubID, errUp := uploadImage(uploadCtx, h.Cld, heroHeader, "sidita_destinasi/hero")
		if errUp != nil {
			destroyImageAsync(h.Cld, newThumbPubID) // rollback thumbnail kalau hero gagal
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Code:    "ERR_FILE_UPLOAD",
				Message: "Gagal upload hero: " + errUp.Error(),
			})
		}
		newHeroPubID = pubID
		oldHeroToDelete = utils.ExtractPublicID(old.GambarUrlHero)
		finalHeroURL = url
	}

		if err := h.Queries.UpdateDestinasi(ctx, db.UpdateDestinasiParams{
        AreaID:               finalAreaID,
        Nama:                 finalNama,
        SlugBaru:             finalSlug, 
        SlugLama:             old.Slug,  
        Kategori:             finalKategori,
        Deskripsi:            finalDeskripsi,
        Alamat:               finalAlamat,
        HighlightText:        finalHighlight,
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
				Code:    "ERR_CONFLICT",
				Message: "Destinasi dengan nama atau slug ini sudah ada",
			})
		}
		slog.Error("admin.destinasi.update_error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal mengupdate destinasi",
		})
	}

	destroyImageAsync(h.Cld, oldThumbToDelete)
	destroyImageAsync(h.Cld, oldHeroToDelete)

	invalidateDestinasiCache(old.ID, old.Slug, finalSlug)
	slog.Info("admin.destinasi.updated", slog.Int("id", int(old.ID)), slog.String("slug", finalSlug))

	return c.JSON(SuccessResponse{
		Pesan: "Destinasi berhasil diupdate",
		Data: DestinasiActionResponse{
			ID:                 old.ID,
			Nama:               finalNama,
			Slug:               finalSlug,
			GambarUrlThumbnail: finalThumbURL,
			GambarUrlHero:      finalHeroURL,
		},
	})
}

// =============================================================================
// ADMIN: DELETE DESTINASI
// =============================================================================

func (h *DestinasiHandler) DeleteDestinasi(c fiber.Ctx) error {
	slugParam := strings.TrimSpace(c.Params("slug"))
	if slugParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Slug destinasi tidak valid",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextDBTimeout)
	defer cancel()

	old, errCari := h.Queries.GetDestinasiBySlugAdmin(ctx, slugParam)
	if errCari != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Code:    "ERR_NOT_FOUND",
			Message: "Destinasi tidak ditemukan",
		})
	}

	if err := h.Queries.DeleteDestinasi(ctx, old.Slug); err != nil {
		slog.Error("admin.destinasi.delete_error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal menghapus destinasi",
		})
	}

	// Hapus image dari Cloudinary (best-effort)
	destroyImageAsync(h.Cld, utils.ExtractPublicID(old.GambarUrlThumbnail))
	destroyImageAsync(h.Cld, utils.ExtractPublicID(old.GambarUrlHero))

	// Hapus cache
	invalidateDestinasiCache(old.ID, old.Slug)
	slog.Info("admin.destinasi.deleted", slog.Int("id", int(old.ID)), slog.String("slug", old.Slug))

	return c.JSON(SuccessResponse{Pesan: "Destinasi berhasil dihapus"})
}

// =============================================================================
// CACHE WARMUP
// =============================================================================

func (h *DestinasiHandler) CacheWarmup() {
	slog.Info("cache.warmup.destinasi.start")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if areas, err := h.Queries.GetAllArea(ctx); err == nil {
		cache.GlobalCache.Set("areas:all",
			SuccessResponse{Pesan: "Daftar Area", Data: areas},
			CacheTTLArea)
	}

	if data, err := h.Queries.GetRecommendationDestinasi(ctx); err == nil {
		if data == nil {
			data = []db.GetRecommendationDestinasiRow{}
		}
		cache.GlobalCache.Set("destinasi:recommendation",
			SuccessResponse{Pesan: "Rekomendasi Destinasi", Data: DestinasiRecommendationResponse{Items: data}},
			CacheTTLRecommendation)
	}

	g, gCtx := errgroup.WithContext(ctx)
	var listData []db.ListDestinasiRow
	var total int64

	g.Go(func() error {
		var err error
		listData, err = h.Queries.ListDestinasi(gCtx, db.ListDestinasiParams{
			LimitData: 10, OffsetData: 0,
		})
		return err
	})
	g.Go(func() error {
		var err error
		total, err = h.Queries.CountDestinasi(gCtx, db.CountDestinasiParams{})
		return err
	})

	if err := g.Wait(); err == nil {
		if listData == nil {
			listData = []db.ListDestinasiRow{}
		}
		warmupKey := "destinasi:list:area_:keyword_:page_1:limit_10"
		cache.GlobalCache.Set(warmupKey,
			SuccessResponse{
				Pesan:      "Daftar Destinasi",
				Data:       listData,
				Pagination: buildPaginationMeta(1, 10, total),
			},
			CacheTTLList)
	}

	slog.Info("cache.warmup.destinasi.done")
}