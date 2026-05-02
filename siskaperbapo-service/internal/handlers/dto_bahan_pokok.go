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
	ID        int32  `json:"id"`
	Nama      string `json:"nama"`
	Slug      string `json:"slug"`
	Satuan    string `json:"satuan"`
	GambarUrl string `json:"gambar_url"`
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