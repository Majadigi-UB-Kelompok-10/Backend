package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/farildzaky/sinaker-service/internal/db"
	"github.com/farildzaky/sinaker-service/internal/utils"
)



// =====================================================================
// PUBLIC: DAFTAR KOTA BLK (Untuk Dropdown Filter)
// =====================================================================
func (h *SinakerHandler) GetKotaBLK(c fiber.Ctx) error {
	cacheKey := "sinaker:public:blk:kota"
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	kotaList, err := h.Queries.GetDistinctKabKotaBLK(ctx)
	if err != nil {
		slog.Error("public.blk.kota.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code: "ERR_DB", Message: "Gagal mengambil data kota",
		})
	}

	res := SuccessResponse{
		Pesan: "Berhasil mengambil data kota",
		Data:  kotaList,
	}
	return cacheJSON(c, cacheKey, 24*time.Hour, res)
}

// =====================================================================
// PUBLIC: DAFTAR BLK AKTIF (Untuk Halaman Utama)
// =====================================================================
func (h *SinakerHandler) GetListBLK(c fiber.Ctx) error {
	kotaFilter := c.Query("kota")

	cacheKey := normalizeKey("sinaker:public:blk:list", kotaFilter)
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	var pgKota pgtype.Text
	if kotaFilter != "" {
		pgKota = pgtype.Text{String: kotaFilter, Valid: true}
	}

	blks, err := h.Queries.ListAktifBLK(ctx, pgKota)
	if err != nil {
		slog.Error("public.blk.list.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code: "ERR_DB", Message: "Gagal mengambil data BLK",
		})
	}

	resData := make([]BLKResponse, 0, len(blks))
	for _, b := range blks {
		var kec *string
		if b.Kecamatan.Valid {
			kec = &b.Kecamatan.String
		}
		var lat, lng *float64
		if b.Lat.Valid {
			latVal := b.Lat.Float64
			lat = &latVal
		}
		if b.Lng.Valid {
			lngVal := b.Lng.Float64
			lng = &lngVal
		}

		resData = append(resData, BLKResponse{
			ID:        b.ID,
			Nama:      b.Nama,
			Alamat:    b.Alamat,
			KabKota:   b.KabKota,
			Kecamatan: kec,
			Slug:      b.Slug,
			Lat:       lat,
			Lng:       lng,
		})
	}

	res := SuccessResponse{
		Pesan: "Berhasil mengambil data BLK",
		Data:  resData,
	}
	return cacheJSON(c, cacheKey, 2*time.Hour, res)
}

// =====================================================================
// PUBLIC: DAFTAR KEJURUAN PER BLK (Untuk Dropdown Form Pendaftaran)
// =====================================================================
func (h *SinakerHandler) GetKejuruanByBLK(c fiber.Ctx) error {
	blkIDStr := c.Params("id")
	blkID, err := strconv.Atoi(blkIDStr)
	if err != nil || blkID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code: "ERR_VALIDATION", Message: "ID BLK tidak valid",
		})
	}

	cacheKey := fmt.Sprintf("sinaker:public:kejuruan:blk_%d", blkID)
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	kejuruanList, err := h.Queries.ListAktifKejuruanByBLK(ctx, int32(blkID))
	if err != nil {
		slog.Error("public.kejuruan.list.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code: "ERR_DB", Message: "Gagal mengambil data kejuruan",
		})
	}

	resData := make([]KejuruanResponse, 0, len(kejuruanList))
	for _, k := range kejuruanList {
		resData = append(resData, KejuruanResponse{
			ID:   k.ID,
			Nama: k.Nama,
		})
	}

	res := SuccessResponse{
		Pesan: "Berhasil mengambil data kejuruan",
		Data:  resData,
	}
	return cacheJSON(c, cacheKey, 2*time.Hour, res)
}

// =====================================================================
// PUBLIC: SUBMIT PENDAFTARAN PELATIHAN
// =====================================================================
func (h *SinakerHandler) SubmitPendaftaran(c fiber.Ctx) error {
	fileHeader, err := c.FormFile("foto")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code: "ERR_VALIDATION", Message: "Pas foto wajib diunggah",
		})
	}

	blkID, errConv := strconv.Atoi(c.FormValue("blk_id"))
	if errConv != nil || blkID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code: "ERR_VALIDATION", Message: "blk_id harus berupa angka positif",
		})
	}
	kejuruanID, errConv := strconv.Atoi(c.FormValue("kejuruan_id"))
	if errConv != nil || kejuruanID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code: "ERR_VALIDATION", Message: "kejuruan_id harus berupa angka positif",
		})
	}
	asalSekolahVal := c.FormValue("asal_sekolah")
	jurusanVal := c.FormValue("jurusan")

	nik, errVal := utils.ValidateNIK(c.FormValue("nik"), "NIK")
	if errVal != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: errVal.Message})
	}
	email, errVal := utils.ValidateEmail(c.FormValue("email"), "Email")
	if errVal != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: errVal.Message})
	}
	noWA, errVal := utils.ValidatePhone(c.FormValue("no_wa"), "No WA")
	if errVal != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: errVal.Message})
	}
	jk, errVal := utils.ValidateJenisKelamin(c.FormValue("jenis_kelamin"))
	if errVal != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: errVal.Message})
	}
	pendTerakhir, errVal := utils.ValidatePendidikan(c.FormValue("pendidikan_terakhir"), "Pendidikan Terakhir")
	if errVal != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: errVal.Message})
	}
	pendSekarang, errVal := utils.ValidatePendidikan(c.FormValue("pendidikan_sekarang"), "Pendidikan Sekarang")
	if errVal != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: errVal.Message})
	}
	tglLahirStr, errVal := utils.ValidateDate(c.FormValue("tanggal_lahir"), "Tanggal Lahir")
	if errVal != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: errVal.Message})
	}

	parsedDate, _ := time.Parse("2006-01-02", tglLahirStr)
	pgTglLahir := pgtype.Date{Time: parsedDate, Valid: true}

	var asalSekolah, jurusan pgtype.Text
	if asalSekolahVal != "" {
		asalSekolah = pgtype.Text{String: asalSekolahVal, Valid: true}
	}
	if jurusanVal != "" {
		jurusan = pgtype.Text{String: jurusanVal, Valid: true}
	}
	penyandangDisabilitas := c.FormValue("penyandang_disabilitas") == "true"

	ctxUpload, cancelUpload := context.WithTimeout(context.Background(), ContextUploadTimeout)
	defer cancelUpload()

	fotoUrl, err := h.uploadImageReal(ctxUpload, fileHeader, "sinaker_pendaftaran")
	if err != nil {
		slog.Error("public.pendaftaran.upload.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code: "ERR_UPLOAD", Message: err.Error(), 
		})
	}

	ctxDB, cancelDB := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancelDB()

	res, err := h.Queries.CreatePendaftaran(ctxDB, db.CreatePendaftaranParams{
		BlkID:                 int32(blkID),
		KejuruanID:            int32(kejuruanID),
		Nik:                   nik,
		NamaLengkap:           c.FormValue("nama_lengkap"),
		TempatLahir:           c.FormValue("tempat_lahir"),
		TanggalLahir:          pgTglLahir,
		Email:                 email,
		JenisKelamin:          db.JenisKelaminType(jk),
		Provinsi:              c.FormValue("provinsi"),
		KabKota:               c.FormValue("kab_kota"),
		Kecamatan:             c.FormValue("kecamatan"),
		Kelurahan:             c.FormValue("kelurahan"),
		Rt:                    c.FormValue("rt"),
		Rw:                    c.FormValue("rw"),
		AlamatLengkap:         c.FormValue("alamat_lengkap"),
		PenyandangDisabilitas: penyandangDisabilitas,
		NoWa:                  noWA,
		NoWaDarurat:           c.FormValue("no_wa_darurat"),
		PendidikanTerakhir:    db.PendidikanType(pendTerakhir),
		PendidikanSekarang:    db.PendidikanType(pendSekarang),
		AsalSekolah:           asalSekolah,
		Jurusan:               jurusan,
		FotoUrl:               fotoUrl, 
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
				Code: "ERR_CONFLICT", Message: "NIK ini sudah terdaftar di BLK tersebut.",
			})
		}
		slog.Error("public.pendaftaran.create.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code: "ERR_INSERT", Message: "Gagal menyimpan formulir pendaftaran",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{
		Pesan: "Pendaftaran berhasil dikirim. Silakan cek status secara berkala.",
		Data: map[string]interface{}{
			"id":         res.ID,
			"status":     res.Status,
			"created_at": res.CreatedAt,
		},
	})
}

// =====================================================================
// PUBLIC: CEK STATUS PENDAFTARAN
// =====================================================================
func (h *SinakerHandler) CekStatus(c fiber.Ctx) error {
	var req CekStatusRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code: "ERR_JSON", Message: "Format data tidak valid",
		})
	}

	nik, errVal := utils.ValidateNIK(req.NIK, "NIK")
	if errVal != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: errVal.Message})
	}
	noWA, errVal := utils.ValidatePhone(req.NoWA, "No WA")
	if errVal != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Code: "ERR_VALIDATION", Message: errVal.Message})
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	riwayat, err := h.Queries.CekStatusPendaftaran(ctx, db.CekStatusPendaftaranParams{
		Nik:  nik,
		NoWa: noWA,
	})

	if err != nil {
		slog.Error("public.pendaftaran.status.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code: "ERR_DB", Message: "Gagal mengecek status pendaftaran",
		})
	}

	if len(riwayat) == 0 {
		return c.Status(fiber.StatusOK).JSON(SuccessResponse{
			Pesan: "Tidak ditemukan riwayat pendaftaran untuk NIK dan Nomor WA tersebut.",
			Data:  []StatusPendaftaranItem{},
		})
	}

	resData := make([]StatusPendaftaranItem, 0, len(riwayat))
	for _, r := range riwayat {
		resData = append(resData, StatusPendaftaranItem{
			ID:            r.ID,
			BlkNama:       r.BlkNama,
			BlkSlug:       r.BlkSlug,
			KejuruanNama:  r.KejuruanNama,
			Status:        string(r.Status),
			TanggalDaftar: r.CreatedAt.Time.Format("2006-01-02 15:04"),
		})
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse{
		Pesan: "Berhasil mengambil riwayat pendaftaran",
		Data:  resData,
	})
}

