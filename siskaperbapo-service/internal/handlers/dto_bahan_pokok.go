package handlers

type BahanPokokFormRequest struct {
	Nama   string `form:"nama"   json:"nama"`
	Satuan string `form:"satuan" json:"satuan"`
}

type UpsertHargaRequest struct {
	BahanPokokID int32  `json:"bahan_pokok_id"`
	AreaID       int32  `json:"area_id"`
	Harga        int32  `json:"harga"`
	Tanggal      string `json:"tanggal"` 
}


type AreaItemResponse struct {
	ID   int32  `json:"id"`
	Nama string `json:"nama"`
	Slug string `json:"slug"`
}

type BahanPokokItemResponse struct {
    ID            int32  `json:"id"`
    Nama          string `json:"komoditas"`     
    Slug          string `json:"slug"`
    Satuan        string `json:"satuan"`
    GambarUrl     string `json:"gambar_url"`
    HargaSekarang int32  `json:"harga_sekarang"` 
    Tren          string `json:"tren"`          
}

type RiwayatHargaResponse struct {
	Tanggal      string `json:"tanggal"`
	RataRataHarga int32  `json:"rata_rata_harga"`
}

type TrenPerbandinganResponse struct {
	BahanPokokID int32  `json:"bahan_pokok_id"`
	Tanggal      string `json:"tanggal"`
	Harga        int32  `json:"harga"`
	StatusTren   string `json:"status_tren"` 
	Selisih      int32  `json:"selisih"`     
}

type AreaHargaItem struct {
    Area     string `json:"area"`
    AreaSlug string `json:"area_slug"`
    Harga    int32  `json:"harga"`
}

type Statistik15Hari struct {
    Tertinggi *AreaHargaItem `json:"tertinggi"` 
    Terendah  *AreaHargaItem `json:"terendah"`
}

type DetailBahanPokokResponse struct {
    ID                int32                  `json:"id"`
    Komoditas         string                 `json:"komoditas"`
    Slug              string                 `json:"slug"`
    Satuan            string                 `json:"satuan"`
    GambarUrl         string                 `json:"gambar_url"`
    Tanggal           string                 `json:"tanggal"`
    TanggalDataAktual string                 `json:"tanggal_data_aktual"`
    AreaPilihan       string                 `json:"area_pilihan"`
    HargaUtama        int32                  `json:"harga_utama"`
    Tren              string                 `json:"tren"`
    GrafikRiwayat     []RiwayatHargaResponse `json:"grafik_riwayat"` 
    ListKabKota       []AreaHargaItem        `json:"list_kab_kota"`
    Statistik15Hari   Statistik15Hari        `json:"statistik_15_hari"`
}