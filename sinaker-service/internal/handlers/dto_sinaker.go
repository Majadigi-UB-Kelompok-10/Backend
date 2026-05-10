package handlers

// =====================================================================
// BLK (Balai Latihan Kerja)
// =====================================================================

type BLKResponse struct {
	ID        int32    `json:"id"`
	Nama      string   `json:"nama"`
	Alamat    string   `json:"alamat"`
	KabKota   string   `json:"kab_kota"`
	Kecamatan *string  `json:"kecamatan,omitempty"`
	Slug      string   `json:"slug"`
	Lat       *float64 `json:"lat,omitempty"`
	Lng       *float64 `json:"lng,omitempty"`
	Aktif     bool     `json:"aktif,omitempty"`
}

type BLKRequest struct {
	Nama      string   `json:"nama"`
	Alamat    string   `json:"alamat"`
	KabKota   string   `json:"kab_kota"`
	Kecamatan *string  `json:"kecamatan,omitempty"`
	Lat       *float64 `json:"lat,omitempty"`
	Lng       *float64 `json:"lng,omitempty"`
	Aktif     bool     `json:"aktif"`
}

// =====================================================================
// KEJURUAN
// =====================================================================

type KejuruanResponse struct {
	ID    int32  `json:"id"`
	Nama  string `json:"nama"`
	Aktif bool   `json:"aktif,omitempty"`
}

type KejuruanRequest struct {
	BlkID int32  `json:"blk_id"`
	Nama  string `json:"nama"`
	Aktif bool   `json:"aktif"`
}

// =====================================================================
// PENDAFTARAN
// =====================================================================

// Request dari Form UI Mobile/Web
type PendaftaranRequest struct {
	BlkID                 int32   `json:"blk_id"`
	KejuruanID            int32   `json:"kejuruan_id"`
	NIK                   string  `json:"nik"`
	NamaLengkap           string  `json:"nama_lengkap"`
	TempatLahir           string  `json:"tempat_lahir"`
	TanggalLahir          string  `json:"tanggal_lahir"` // Format YYYY-MM-DD
	Email                 string  `json:"email"`
	JenisKelamin          string  `json:"jenis_kelamin"` // 'laki_laki' atau 'perempuan'
	Provinsi              string  `json:"provinsi"`
	KabKota               string  `json:"kab_kota"`
	Kecamatan             string  `json:"kecamatan"`
	Kelurahan             string  `json:"kelurahan"`
	RT                    string  `json:"rt"`
	RW                    string  `json:"rw"`
	AlamatLengkap         string  `json:"alamat_lengkap"`
	PenyandangDisabilitas bool    `json:"penyandang_disabilitas"`
	NoWA                  string  `json:"no_wa"`
	NoWADarurat           string  `json:"no_wa_darurat"`
	PendidikanTerakhir    string  `json:"pendidikan_terakhir"`
	PendidikanSekarang    string  `json:"pendidikan_sekarang"`
	AsalSekolah           *string `json:"asal_sekolah,omitempty"`
	Jurusan               *string `json:"jurusan,omitempty"`
	FotoUrl               string  `json:"foto_url"`
}

// Request untuk Form Cek Status
type CekStatusRequest struct {
	NIK  string `json:"nik"`
	NoWA string `json:"no_wa"`
}

// Response untuk Cek Status (List Riwayat)
type StatusPendaftaranItem struct {
	ID           int32  `json:"id"`
	BlkNama      string `json:"blk_nama"`
	BlkSlug      string `json:"blk_slug"`
	KejuruanNama string `json:"kejuruan_nama"`
	Status       string `json:"status"` // pending, diterima, ditolak
	TanggalDaftar string `json:"tanggal_daftar"`
}

// Response Detail untuk Admin
type PendaftaranAdminResponse struct {
	ID                    int32   `json:"id"`
	NIK                   string  `json:"nik"`
	NamaLengkap           string  `json:"nama_lengkap"`
	NoWA                  string  `json:"no_wa"`
	Status                string  `json:"status"`
	BlkNama               string  `json:"blk_nama"`
	KejuruanNama          string  `json:"kejuruan_nama"`
	TanggalDaftar         string  `json:"tanggal_daftar"`
	// Field detail lainnya bisa ditambahkan nanti sesuai kebutuhan admin
}

type UpdateStatusRequest struct {
	Status string `json:"status"` // pending, diterima, ditolak
}