package handlers

// =============================================================================
// PUBLIC DTO (Untuk Konsumsi Frontend Mobile/Web RSSA)
// =============================================================================

// SummaryRuanganResponse melayani Header: "228 Kamar Masih Tersedia dari 909"
type SummaryRuanganResponse struct {
	TotalKapasitas int `json:"total_kapasitas"`
	TotalTersedia  int `json:"total_tersedia"`
}

// MasterKelasResponse melayani Filter Chips: "Kelas I", "Kelas II", "Kelas VIP"
type MasterKelasResponse struct {
	ID   int32  `json:"id"`
	Nama string `json:"nama"`
	Slug string `json:"slug"`
}

// RuanganPublicResponse melayani Card List Pencarian
type RuanganPublicResponse struct {
	ID        int32  `json:"id"`
	Nama      string `json:"nama"`
	Slug      string `json:"slug"`
	KelasNama string `json:"kelas_nama"`
	KelasSlug string `json:"kelas_slug"`
	Kapasitas int32  `json:"kapasitas"`
	Terisi    int32  `json:"terisi"`
	Tersedia  int32  `json:"tersedia"`
}

// =============================================================================
// ADMIN DTO (Untuk Operasi CRUD oleh Petugas RS)
// =============================================================================

type CreateRuanganRequest struct {
	KelasID   int32  `json:"kelas_id"`
	Nama      string `json:"nama"`
	Kapasitas int32  `json:"kapasitas"`
	Terisi    int32  `json:"terisi"`
}

type UpdateRuanganRequest struct {
	KelasID   int32  `json:"kelas_id"`
	Nama      string `json:"nama"`
	Kapasitas int32  `json:"kapasitas"`
	Terisi    int32  `json:"terisi"`
}