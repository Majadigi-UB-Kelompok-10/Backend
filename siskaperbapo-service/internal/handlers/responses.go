package handlers

type BaseResponse struct {
	Pesan string      `json:"pesan,omitempty"`
	Error string      `json:"error,omitempty"`
	Data  interface{} `json:"data,omitempty"`
}

type ItemBahanPokok struct {
	ID            int32  `json:"id"`
	Komoditas     string `json:"komoditas"`
	Slug          string `json:"slug"`
	Satuan        string `json:"satuan"`
	GambarURL     string `json:"gambar_url"`
	HargaSekarang int32  `json:"harga_sekarang"`
	Tren          string `json:"tren"`
}

type PaginatedResponse struct {
	Tanggal      string           `json:"tanggal"`
	AreaPilihan  string           `json:"area_pilihan"`
	HalamanIni   int              `json:"halaman_ini"`
	DataPerHal   int              `json:"data_per_hal"`
	TotalData    int64            `json:"total_data"`
	TotalHalaman int              `json:"total_halaman"`
	Data         []ItemBahanPokok `json:"data"`
}

type StatistikHarga struct {
	Area     string `json:"area"`
	AreaSlug string `json:"area_slug"`
	Harga    int32  `json:"harga"`
}

type ItemGrafik struct {
	Tanggal       string `json:"tanggal"`
	RataRataHarga int32  `json:"rata_rata_harga"`
}

type ItemKabKota struct {
	Area     string `json:"area"`
	AreaSlug string `json:"area_slug"`
	Harga    int32  `json:"harga"`
}

type DetailBahanResponse struct {
	IDKomoditas       int32          `json:"id_komoditas"` 
	Komoditas         string         `json:"komoditas"`
	Slug              string         `json:"slug"`
	Satuan            string         `json:"satuan"`
	GambarURL         string         `json:"gambar_url"`
	Tanggal		      string         `json:"tanggal"`
	TanggalDataAktual string         `json:"tanggal_data_aktual"`
	AreaPilihan       string         `json:"area_pilihan"`
	HargaUtama        int32          `json:"harga_utama"`
	Tren              string         `json:"tren"`
	Statistik15Hari   struct {
		Tertinggi *StatistikHarga `json:"tertinggi"`
		Terendah  *StatistikHarga `json:"terendah"`
	} `json:"statistik_15_hari"`
	GrafikRiwayat []ItemGrafik  `json:"grafik_riwayat"` 
	ListKabKota   []ItemKabKota `json:"list_kabkota"`   
}

type DataHargaHarian struct {
	ID        int32  `json:"id"`
	Komoditas string `json:"komoditas,omitempty"`
	Area      string `json:"area,omitempty"`
	Harga     int32  `json:"harga"`
	Tanggal   string `json:"tanggal"`
}

type ItemArea struct {
	ID   int32  `json:"id"`
	Nama string `json:"nama"`
	Slug string `json:"slug"` 
}

type AreaListResponse struct {
	Pesan string     `json:"pesan"`
	Data  []ItemArea `json:"data"`
}