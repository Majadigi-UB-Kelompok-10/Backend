package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/farildzaky/bansos-service/internal/utils"
)



func (h *BansosHandler) PublicCekBansos(c fiber.Ctx) error {
	nik := c.Query("nik")
	validNIK, ve := utils.ValidateNIK(nik)
	if ve != nil {
		return validationErrorResponse(c, ve)
	}

	cacheKey := normalizeKey("bansos:public:cek_nik", validNIK)
	if respondCached(c, h.Cache, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	penerima, err := h.Queries.GetPenerimaByNIK(ctx, validNIK)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Code:    "ERR_NOT_FOUND",
			Message: "Data tidak ditemukan. Pastikan NIK Anda terdaftar sebagai penerima bansos.",
		})
	}

	penyaluranList, err := h.Queries.GetPenyaluranByPenerima(ctx, penerima.ID)
	if err != nil {
		slog.Error("public.cek_bansos.error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal memuat riwayat bansos",
		})
	}

	riwayatData := make([]RiwayatBansosItem, 0, len(penyaluranList))
	for _, row := range penyaluranList {
		riwayatData = append(riwayatData, RiwayatBansosItem{
			PenyaluranID: int(row.ID),
			ProgramNama:  row.ProgramNama,
			Periode:      FormatPeriodeTunggal(row.PeriodeMulai.Time), 
			Nominal:      FormatRupiah(int(row.Nominal)),             
			Status:       strings.ToUpper(string(row.Status)),         
		})
	}

	response := CekBansosResponse{
		Profil: ProfilPenerima{
			Nama:   penerima.Nama,
			Alamat: penerima.Alamat,
			NIK:    penerima.Nik,
		},
		Riwayat: riwayatData,
	}

	res := SuccessResponse{Pesan: "Data ditemukan", Data: response}
	return cacheJSON(c, h.Cache, cacheKey, CacheTTLInfoBansos, res)
}

// =============================================================================
// PUBLIC: DETAIL PENYALURAN BANSOS
// =============================================================================

func (h *BansosHandler) PublicGetDetailBansos(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "ID Penyaluran tidak valid",
		})
	}

	cacheKey := fmt.Sprintf("bansos:public:detail:%d", id)
	if respondCached(c, h.Cache, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	detail, err := h.Queries.GetDetailPenyaluran(ctx, int32(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Code:    "ERR_NOT_FOUND",
			Message: "Detail penyaluran tidak ditemukan",
		})
	}

	var deskripsi string
	if detail.ProgramDeskripsi.Valid {
		deskripsi = detail.ProgramDeskripsi.String
	}

	// 5. Mapping ke DTO Detail
	response := DetailBansosResponse{
		ProgramNama:      detail.ProgramNama,
		Nominal:          FormatRupiah(int(detail.Nominal)),
		Periode:          FormatPeriodeRentang(detail.PeriodeMulai.Time, detail.PeriodeSelesai.Time), 
		MetodePenyaluran: FormatMetode(string(detail.Metode)),                                        
		Status:           strings.ToUpper(string(detail.Status)),                                     
		DeskripsiProgram: deskripsi,
	}

	res := SuccessResponse{Pesan: "Detail Bansos", Data: response}
	return cacheJSON(c, h.Cache, cacheKey, CacheTTLInfoBansos, res)
}