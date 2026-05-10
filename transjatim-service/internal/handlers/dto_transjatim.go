package handlers

// =====================================================================
// TERMINAL
// =====================================================================

type TerminalResponse struct {
	ID    int32    `json:"id"`
	Nama  string   `json:"nama"`
	Kota  string   `json:"kota"`
	Slug  string   `json:"slug"`
	Lat   *float64 `json:"lat,omitempty"`
	Lng   *float64 `json:"lng,omitempty"`
	Aktif bool     `json:"aktif,omitempty"`
}

// =====================================================================
// BUS
// =====================================================================

type BusResponse struct {
	ID        int32   `json:"id"`
	Kode      string  `json:"kode"`
	Layanan   string  `json:"layanan"`
	Fasilitas *string `json:"fasilitas,omitempty"`
	Kapasitas *int32  `json:"kapasitas,omitempty"`
}

// =====================================================================
// RUTE
// =====================================================================

type RuteResponse struct {
	ID             int32  `json:"id"`
	Slug           string `json:"slug"`
	TerminalAsal   string `json:"terminal_asal"`
	KotaAsal       string `json:"kota_asal"`
	TerminalTujuan string `json:"terminal_tujuan"`
	KotaTujuan     string `json:"kota_tujuan"`
	DurasiMenit    int32  `json:"durasi_menit"`
	Aktif          bool   `json:"aktif"`
}

type RuteStopItem struct {
	Urutan     int32  `json:"urutan"`
	TerminalID int32  `json:"terminal_id"`
	Nama       string `json:"nama"`
	Kota       string `json:"kota"`
}

// =====================================================================
// JADWAL — PUBLIC
// =====================================================================

type JadwalSearchItem struct {
	ID             int32  `json:"id"`
	BusKode        string `json:"bus_kode"`
	BusLayanan     string `json:"bus_layanan"`
	JamBerangkat   string `json:"jam_berangkat"`
	JamTiba        string `json:"jam_tiba"`
	DurasiMenit    int32  `json:"durasi_menit"`
	TerminalAsal   string `json:"terminal_asal"`
	TerminalTujuan string `json:"terminal_tujuan"`
	Harga          int64  `json:"harga"`
}

type JadwalDetailResponse struct {
	ID             int32       `json:"id"`
	BusKode        string      `json:"bus_kode"`
	BusLayanan     string      `json:"bus_layanan"`
	BusFasilitas   *string     `json:"bus_fasilitas,omitempty"`
	JamBerangkat   string      `json:"jam_berangkat"`
	JamTiba        string      `json:"jam_tiba"`
	DurasiMenit    int32       `json:"durasi_menit"`
	TerminalAsal   string      `json:"terminal_asal"`
	TerminalTujuan string      `json:"terminal_tujuan"`
	RuteID         int32       `json:"rute_id"`
	Stops          []string    `json:"stops"`
	SemuaHarga     []HargaItem `json:"semua_harga"`
}

type HargaItem struct {
	TipePenumpang *string `json:"tipe_penumpang,omitempty"`
	Harga         int32   `json:"harga"`
}

// =====================================================================
// JADWAL — ADMIN
// =====================================================================

type JadwalAdminItem struct {
	ID             int32  `json:"id"`
	BusKode        string `json:"bus_kode"`
	BusLayanan     string `json:"bus_layanan"`
	TerminalAsal   string `json:"terminal_asal"`
	TerminalTujuan string `json:"terminal_tujuan"`
	RuteSlug       string `json:"rute_slug"`
	JamBerangkat   string `json:"jam_berangkat"`
	JamTiba        string `json:"jam_tiba"`
	HariOperasi    int16  `json:"hari_operasi"`
	Aktif          bool   `json:"aktif"`
}

// =====================================================================
// JADWAL LIBUR
// =====================================================================

type JadwalLiburResponse struct {
	ID         int32   `json:"id"`
	JadwalID   *int32  `json:"jadwal_id,omitempty"`
	Tanggal    string  `json:"tanggal"`
	Keterangan *string `json:"keterangan,omitempty"`
}

// =====================================================================
// HARGA
// =====================================================================

type HargaAdminItem struct {
	ID             int32   `json:"id"`
	RuteID         int32   `json:"rute_id"`
	TerminalAsal   string  `json:"terminal_asal"`
	TerminalTujuan string  `json:"terminal_tujuan"`
	Layanan        string  `json:"layanan"`
	TipePenumpang  *string `json:"tipe_penumpang,omitempty"`
	Harga          int32   `json:"harga"`
}

type HargaInfoResponse struct {
	Reguler []HargaRegulerInfo `json:"reguler"`
	Luxury  []HargaLuxuryInfo  `json:"luxury"`
}

type HargaRegulerInfo struct {
	TipePenumpang string `json:"tipe_penumpang"`
	Harga         int32  `json:"harga"`
	Keterangan    string `json:"keterangan"`
}

type HargaLuxuryInfo struct {
	RuteNama  string  `json:"rute_nama"`
	Harga     int32   `json:"harga"`
	Fasilitas *string `json:"fasilitas,omitempty"`
}

// =====================================================================
// REQUEST DTOs (INPUT DARI ADMIN)
// =====================================================================

type TerminalRequest struct {
	Nama  string   `json:"nama"`
	Kota  string   `json:"kota"`
	Lat   *float64 `json:"lat,omitempty"`
	Lng   *float64 `json:"lng,omitempty"`
	Aktif bool     `json:"aktif"`
}

type BusRequest struct {
	Kode      string  `json:"kode"`
	Layanan   string  `json:"layanan"`
	Fasilitas *string `json:"fasilitas,omitempty"`
	Kapasitas *int32  `json:"kapasitas,omitempty"`
}

type HargaRequest struct {
	RuteID        int32   `json:"rute_id"`
	Layanan       string  `json:"layanan"`        
	TipePenumpang *string `json:"tipe_penumpang"` 
	Harga         int32   `json:"harga"`
}

// =====================================================================
// TAMBAHAN REQUEST DTOs (Di dto_transjatim.go)
// =====================================================================

type RuteRequest struct {
	TerminalAsalID   int32   `json:"terminal_asal_id"`
	TerminalTujuanID int32   `json:"terminal_tujuan_id"`
	DurasiMenit      int32   `json:"durasi_menit"`
	Aktif            bool    `json:"aktif"`
	Stops            []int32 `json:"stops"` 
}

type JadwalRequest struct {
	RuteID       int32  `json:"rute_id"`
	BusID        int32  `json:"bus_id"`
	JamBerangkat string `json:"jam_berangkat"` 
	JamTiba      string `json:"jam_tiba"`      
	HariOperasi  int16  `json:"hari_operasi"`  
	Aktif        bool   `json:"aktif"`
}