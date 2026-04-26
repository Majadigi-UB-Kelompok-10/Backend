package handlers

type HitungNjkbRequest struct {
	Jenis string `json:"jenis"`
	Merk  string `json:"merk"`
	Model string `json:"model"`
	Tipe  string `json:"tipe"`
	Tahun int16  `json:"tahun"`
}

type MasterNjkbRequest struct {
	Jenis string `json:"jenis_kendaraan"`
	Merk  string `json:"merk"`
	Model string `json:"model"`
	Tipe  string `json:"tipe"`
	Tahun int16  `json:"tahun"`
	Nilai int64  `json:"nilai_jual"`
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