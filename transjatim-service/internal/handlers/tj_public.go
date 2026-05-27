package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/farildzaky/transjatim-service/internal/db"
	"github.com/farildzaky/transjatim-service/internal/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/sync/errgroup"
)

// =====================================================================
// PUBLIC: GET ALL ACTIVE TERMINALS (Untuk Dropdown Form)
// =====================================================================
func (h *TransJatimHandler) GetActiveTerminals(c fiber.Ctx) error {
	cacheKey := "transjatim:public:terminals:active"

	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	terminals, err := h.Queries.ListActiveTerminals(ctx)
	if err != nil {
		slog.Error("public.terminals.list_error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal mengambil daftar terminal",
		})
	}

	resData := make([]TerminalResponse, 0, len(terminals))
	for _, t := range terminals {
		var lat, lng *float64
		if t.Lat.Valid {
			latVal := t.Lat.Float64
			lat = &latVal
		}
		if t.Lng.Valid {
			lngVal := t.Lng.Float64
			lng = &lngVal
		}

		resData = append(resData, TerminalResponse{
			ID:    t.ID,
			Nama:  t.Nama,
			Kota:  t.Kota,
			Slug:  t.Slug,
			Lat:   lat,
			Lng:   lng,
			Aktif: true,
		})
	}

	res := SuccessResponse{
		Pesan: "Daftar Terminal Aktif",
		Data:  resData,
	}
	return cacheJSON(c, cacheKey, CacheTTLTerminals, res)
}

// =====================================================================
// PUBLIC: SEARCH JADWAL BUS (Sesuai UI Mobile)
// =====================================================================
func (h *TransJatimHandler) SearchJadwal(c fiber.Ctx) error {
	asalStr := c.Query("asal_id")
	tujuanStr := c.Query("tujuan_id")
	tanggalStr := c.Query("tanggal")

	if asalStr == "" || tujuanStr == "" || tanggalStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Parameter 'asal_id', 'tujuan_id', dan 'tanggal' wajib diisi",
		})
	}

	asalID, errA := strconv.Atoi(asalStr)
	tujuanID, errT := strconv.Atoi(tujuanStr)
	if errA != nil || errT != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "ID terminal harus berupa angka",
		})
	}

	if asalID == tujuanID {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "Pilih terminal tujuan yang berbeda dari terminal asal.",
		})
	}

	tanggalValid, errVal := utils.ValidateDate(tanggalStr, "tanggal")
	if errVal != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: errVal.Message,
		})
	}

	cacheKey := fmt.Sprintf("transjatim:public:jadwal:asal_%d:tujuan_%d:tgl_%s", asalID, tujuanID, tanggalValid)
	if respondCached(c, cacheKey) {
		return nil
	}

	parsedTime, _ := time.Parse("2006-01-02", tanggalValid)
	hariBitmask := dayBitmask(parsedTime)
	tanggalPg := pgtype.Date{Time: parsedTime, Valid: true}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	jadwals, err := h.Queries.SearchJadwalPublic(ctx, db.SearchJadwalPublicParams{
		TerminalAsalID:   int32(asalID),
		TerminalTujuanID: int32(tujuanID),
		HariBitmask:      int16(hariBitmask),
		Tanggal:          tanggalPg,
	})

	if err != nil {
		slog.Error("public.jadwal.search_error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal mencari jadwal bus",
		})
	}

	resData := make([]JadwalSearchItem, 0, len(jadwals))
	for _, j := range jadwals {
		jamBerangkatStr := time.Unix(0, j.JamBerangkat.Microseconds*1000).UTC().Format("15:04")
		jamTibaStr := time.Unix(0, j.JamTiba.Microseconds*1000).UTC().Format("15:04")

		hargaDasar := int64(j.HargaDasar)

		resData = append(resData, JadwalSearchItem{
			ID:             j.ID,
			BusKode:        j.BusKode,
			BusLayanan:     string(j.BusLayanan),
			JamBerangkat:   jamBerangkatStr,
			JamTiba:        jamTibaStr,
			DurasiMenit:    j.DurasiMenit,
			TerminalAsal:   j.TerminalAsalNama,
			TerminalTujuan: j.TerminalTujuanNama,
			Harga:          hargaDasar,
		})
	}

	if resData == nil {
		resData = []JadwalSearchItem{}
	}

	if len(resData) == 0 {
        res := SuccessResponse{
            Pesan: "Mohon maaf, jadwal bus belum tersedia untuk rute dan tanggal tersebut.",
            Data:  []JadwalSearchItem{},
        }
        return cacheJSON(c, cacheKey, CacheTTLSearch, res)
    }

	res := SuccessResponse{
		Pesan: "Pencarian Jadwal Berhasil",
		Data:  resData,
	}
	return cacheJSON(c, cacheKey, CacheTTLSearch, res)
}

// =====================================================================
// PUBLIC: DETAIL JADWAL (Titik Pemberhentian & List Harga)
// =====================================================================
func (h *TransJatimHandler) GetJadwalDetail(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Code:    "ERR_VALIDATION",
			Message: "ID Jadwal tidak valid",
		})
	}

	cacheKey := fmt.Sprintf("transjatim:public:jadwal:detail:%d", id)
	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	jadwal, errJ := h.Queries.GetJadwalByID(ctx, int32(id))
	if errJ != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Code:    "ERR_NOT_FOUND",
			Message: "Jadwal tidak ditemukan",
		})
	}

	g, gCtx := errgroup.WithContext(ctx)
	var stops []db.GetRuteStopsRow
	var hargas []db.GetHargaByRuteRow
	var routeGeometry []byte

	g.Go(func() error {
		var err error
		stops, err = h.Queries.GetRuteStops(gCtx, jadwal.RuteID)
		return err
	})

	g.Go(func() error {
		var err error
		hargas, err = h.Queries.GetHargaByRute(gCtx, jadwal.RuteID)
		return err
	})

	g.Go(func() error {
		var err error
		routeGeometry, err = h.Queries.GetRuteGeometry(gCtx, jadwal.RuteID)
		if err != nil {
			slog.Debug("public.jadwal.detail.geometry_empty", slog.Int("rute_id", int(jadwal.RuteID)))
		}
		return nil // geometry is optional, don't fail the request
	})

	if err := g.Wait(); err != nil {
		slog.Error("public.jadwal.detail_error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal memuat detail jadwal",
		})
	}

	stopItems := make([]StopItem, 0, len(stops))
	for _, s := range stops {
		item := StopItem{
			Urutan: s.Urutan,
			Nama:   s.Nama,
			Kota:   s.Kota,
		}
		if s.Lat.Valid {
			v := s.Lat.Float64
			item.Lat = &v
		}
		if s.Lng.Valid {
			v := s.Lng.Float64
			item.Lng = &v
		}
		stopItems = append(stopItems, item)
	}

	hargaItems := make([]HargaItem, 0)
	for _, h := range hargas {
		if string(h.Layanan) == string(jadwal.BusLayanan) {
			var tipe *string
			if h.TipePenumpang.Valid {
				str := string(h.TipePenumpang.TipePenumpangType)
				tipe = &str
			}
			hargaItems = append(hargaItems, HargaItem{
				TipePenumpang: tipe,
				Harga:         h.Harga,
			})
		}
	}

	var fas *string
	if jadwal.BusFasilitas.Valid {
		fas = &jadwal.BusFasilitas.String
	}

	jamBerangkatStr := time.Unix(0, jadwal.JamBerangkat.Microseconds*1000).UTC().Format("15:04")
	jamTibaStr := time.Unix(0, jadwal.JamTiba.Microseconds*1000).UTC().Format("15:04")

	resData := JadwalDetailResponse{
		ID:             jadwal.ID,
		BusKode:        jadwal.BusKode,
		BusLayanan:     string(jadwal.BusLayanan),
		BusFasilitas:   fas,
		JamBerangkat:   jamBerangkatStr,
		JamTiba:        jamTibaStr,
		DurasiMenit:    jadwal.DurasiMenit,
		TerminalAsal:   jadwal.TerminalAsalNama,
		TerminalTujuan: jadwal.TerminalTujuanNama,
		RuteID:         jadwal.RuteID,
		Stops:          stopItems,
		SemuaHarga:     hargaItems,
		KoordinatRute:  routeGeometry,
	}

	res := SuccessResponse{
		Pesan: "Detail Jadwal",
		Data:  resData,
	}

	return cacheJSON(c, cacheKey, CacheTTLDetail, res)
}

// =====================================================================
// PUBLIC: INFORMASI HARGA TIKET (Untuk Homepage UI)
// =====================================================================
func (h *TransJatimHandler) GetInformasiHarga(c fiber.Ctx) error {
	cacheKey := "transjatim:public:harga:all"

	if respondCached(c, cacheKey) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContextQueryTimeout)
	defer cancel()

	hargas, err := h.Queries.GetAllHargaInfo(ctx)
	if err != nil {
		slog.Error("public.harga.list_error", slog.String("err", err.Error()))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Code:    "ERR_INTERNAL_DB",
			Message: "Gagal memuat informasi harga tiket",
		})
	}

	var regulerList []HargaRegulerInfo
	var luxuryList []HargaLuxuryInfo
	regulerTracker := make(map[string]bool)

	for _, h := range hargas {
		if h.Layanan == "reguler" && h.TipePenumpang.Valid {
			tipe := string(h.TipePenumpang.TipePenumpangType)
			if !regulerTracker[tipe] {

				keterangan := ""
				switch tipe {
				case "umum":
					keterangan = "Penumpang dewasa / umum."
				case "pelajar_santri":
					keterangan = "Menunjukkan kartu pelajar atau berseragam."
				case "mahasiswa":
					keterangan = "Menunjukkan Kartu Tanda Mahasiswa (KTM)."
				}

				regulerList = append(regulerList, HargaRegulerInfo{
					TipePenumpang: tipe,
					Harga:         h.Harga,
					Keterangan:    keterangan,
				})
				regulerTracker[tipe] = true
			}
		} else if h.Layanan == "luxury" {
			ruteNamaPendek := fmt.Sprintf("%s - %s", h.KotaAsal, h.KotaTujuan)
			fasilitas := "Kursi premium (tanpa berdiri) dan AC ekstra dingin."

			luxuryList = append(luxuryList, HargaLuxuryInfo{
				RuteNama:  ruteNamaPendek,
				Harga:     h.Harga,
				Fasilitas: &fasilitas,
			})
		}
	}

	if regulerList == nil {
		regulerList = []HargaRegulerInfo{}
	}
	if luxuryList == nil {
		luxuryList = []HargaLuxuryInfo{}
	}

	res := SuccessResponse{
		Pesan: "Informasi Harga Tiket",
		Data: HargaInfoResponse{
			Reguler: regulerList,
			Luxury:  luxuryList,
		},
	}

	return cacheJSON(c, cacheKey, CacheTTLHargaInfo, res)
}