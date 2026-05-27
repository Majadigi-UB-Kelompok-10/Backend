package handlers

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

type InfoPajakResponse struct {
	Identitas        IdentitasKendaraan `json:"identitas"`
	RincianBiaya     RincianBiayaPajak  `json:"rincian_biaya"`
	Estimasi5Tahunan Estimasi5Tahunan   `json:"estimasi_5_tahunan"`
}

type IdentitasKendaraan struct {
	PlatNomor   string `json:"plat_nomor"`
	Merk        string `json:"merk"`
	Tipe        string `json:"tipe"`
	Model       string `json:"model"`
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