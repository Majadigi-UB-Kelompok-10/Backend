package handlers

type ProfilPenerima struct {
	Nama   string `json:"nama"`
	Alamat string `json:"alamat"`
	NIK    string `json:"nik"`
}

type RiwayatBansosItem struct {
	PenyaluranID int    `json:"penyaluran_id"` 
	ProgramNama  string `json:"program_nama"`  
	Periode      string `json:"periode"`       
	Nominal      string `json:"nominal"`       
	Status       string `json:"status"`        
}

type CekBansosResponse struct {
	Profil  ProfilPenerima      `json:"profil"`
	Riwayat []RiwayatBansosItem `json:"riwayat"`
}

type DetailBansosResponse struct {
	ProgramNama      string `json:"program_nama"`       
	Nominal          string `json:"nominal"`            
	Periode          string `json:"periode"`            
	MetodePenyaluran string `json:"metode_penyaluran"`  
	Status           string `json:"status"`             
	DeskripsiProgram string `json:"deskripsi_program"`  
}


type AdminProgramItem struct {
	ID        int    `json:"id"`
	Nama      string `json:"nama"`
	Kode      string `json:"kode"`
	Deskripsi string `json:"deskripsi"`
	Aktif     bool   `json:"aktif"`
}

type AdminPenerimaItem struct {
	ID     int    `json:"id"`
	NIK    string `json:"nik"`
	Nama   string `json:"nama"`
	Alamat string `json:"alamat"`
}

type AdminPenyaluranItem struct {
	ID             int    `json:"id"`
	PenerimaNIK    string `json:"penerima_nik"`
	PenerimaNama   string `json:"penerima_nama"`
	ProgramNama    string `json:"program_nama"`
	Nominal        int    `json:"nominal"` 
	PeriodeMulai   string `json:"periode_mulai"`
	PeriodeSelesai string `json:"periode_selesai"`
	Metode         string `json:"metode"`
	Status         string `json:"status"`
	CreatedAt      string `json:"created_at"`
}