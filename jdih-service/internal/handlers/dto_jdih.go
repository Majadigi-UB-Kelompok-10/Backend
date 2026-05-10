package handlers

type PengumumanResponse struct {
	ID        int32  `json:"id"`
	Judul     string `json:"judul"`
	Isi       string `json:"isi,omitempty"`
	Tanggal   string `json:"tanggal"` 
}

type DokumenItemResponse struct {
	ID             int32  `json:"id"`
	Jenis          string `json:"jenis"`  
	Judul          string `json:"judul"`  
	Ringkasan      string `json:"ringkasan,omitempty"`
	Tanggal        string `json:"tanggal"` 
	Status         string `json:"status"`  
	JumlahView     int32  `json:"jumlah_view,omitempty"`
	PdfUrl         string `json:"pdf_url,omitempty"`
}

type SubjekResponse struct {
	ID   int32  `json:"id"`
	Nama string `json:"nama"`
}

type DokumenDetailResponse struct {
	ID                  int32            `json:"id"`
	Jenis               string           `json:"jenis"`
	Nomor               string           `json:"nomor"`
	Tahun               int16            `json:"tahun"`
	Judul               string           `json:"judul"`
	Ringkasan           string           `json:"ringkasan,omitempty"`
	TanggalPenetapan    string           `json:"tanggal_penetapan"`
	Status              string           `json:"status"`
	PdfUrl              string           `json:"pdf_url"`
	PdfSizeKb           int32            `json:"pdf_size_kb,omitempty"`
	UrusanPemerintahan  string           `json:"urusan_pemerintahan,omitempty"`
	JumlahView          int32            `json:"jumlah_view"`
	Subjek              []SubjekResponse `json:"subjek"`
}