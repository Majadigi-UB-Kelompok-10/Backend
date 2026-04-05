package handlers

type SuccessResponse struct {
	Pesan      string      `json:"pesan"`
	Data       interface{} `json:"data,omitempty"`
	Pagination interface{} `json:"pagination,omitempty"`
}

type ErrorResponse struct {
	Error  string `json:"error"`
	Detail string `json:"detail,omitempty"`
}

type PaginationMeta struct {
	Page  int   `json:"page,omitempty"`
	Limit int   `json:"limit,omitempty"`
	Total int64 `json:"total"`
}

type InfoPajakResponse struct {
	Identitas        IdentitasKendaraan `json:"identitas"`
	RincianBiaya     RincianBiayaPajak  `json:"rincian_biaya"`
	Estimasi5Tahunan Estimasi5Tahunan   `json:"estimasi_5_tahunan"`
}

type IdentitasKendaraan struct {
	PlatNomor   string `json:"plat_nomor"`
	Merk        string `json:"merk"`
	Tipe        string `json:"tipe"`
	Warna       string `json:"warna"`
	TahunBuat   int16  `json:"tahun_buat"`
	MasaPajak   string `json:"masa_pajak"`
	StatusAktif bool   `json:"status_aktif"`
}

type RincianBiayaPajak struct {
	PkbPokok           int32 `json:"pkb_pokok"`
	OpsenPkb           int32 `json:"opsen_pkb"`
	Swdkllj            int32 `json:"swdkllj"`
	ParkirBerlangganan int32 `json:"parkir_berlangganan"`
	TotalPajak         int64 `json:"total_pajak"`
}

type Estimasi5Tahunan struct {
	CetakStnk int32 `json:"cetak_stnk"`
	CetakTnkb int32 `json:"cetak_tnkb"`
}

type KalkulasiNJKBResponse struct {
	Njkb     int64           `json:"njkb"`
	Estimasi []EstimasiTarif `json:"estimasi"`
}

type EstimasiTarif struct {
	JenisPlat string `json:"jenis_plat"`
	Label     string `json:"label"`
	Pkb       int64  `json:"pkb"`
	Opsen     int64  `json:"opsen"`
}

type MasterNjkbSummaryResponse struct {
	ID             int32  `json:"id"`
	NamaKendaraan  string `json:"nama_kendaraan"`
	JenisKendaraan string `json:"jenis_kendaraan"`
	Merk           string `json:"merk"`
	Model          string `json:"model"`
}


type KendaraanRequest struct {
	PlatNomor          string `json:"plat_nomor"`
	PlatNomorDisplay   string `json:"plat_nomor_display"`
	NomorRangka        string `json:"nomor_rangka"`
	StatusAktif        bool   `json:"status_aktif"`
	Merk               string `json:"merk"`
	Tipe               string `json:"tipe"`
	Warna              string `json:"warna"`
	TahunBuat          int16  `json:"tahun_buat"`
	Model              string `json:"model"`
	MasaPajak          string `json:"masa_pajak"`
	PkbPokok           int32  `json:"pkb_pokok"`
	OpsenPkb           int32  `json:"opsen_pkb"`
	Swdkllj            int32  `json:"swdkllj"`
	ParkirBerlangganan int32  `json:"parkir_berlangganan"`
	CetakStnk          int32  `json:"cetak_stnk"`
	CetakTnkb          int32  `json:"cetak_tnkb"`
}