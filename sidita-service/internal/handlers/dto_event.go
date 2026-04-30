package handlers

import "github.com/farildzaky/sidita-service/internal/db"

type CreateEventRequest struct {
	AreaName       string `form:"area"             json:"area"`
	Nama           string `form:"nama"             json:"nama"`
	Deskripsi      string `form:"deskripsi"        json:"deskripsi"`
	Alamat         string `form:"alamat"           json:"alamat"`
	TanggalMulai   string `form:"tanggal_mulai"    json:"tanggal_mulai"` 
	TanggalSelesai string `form:"tanggal_selesai"  json:"tanggal_selesai"` 
	InfoTiket      string `form:"info_tiket"       json:"info_tiket,omitempty"`
	HargaTiket     string `form:"harga_tiket"      json:"harga_tiket,omitempty"`
	Lat            string `form:"lat"              json:"lat"`
	Lng            string `form:"lng"              json:"lng"`
}

type UpdateEventRequest struct {
	AreaName       string `form:"area"             json:"area,omitempty"`
	Nama           string `form:"nama"             json:"nama,omitempty"`
	Deskripsi      string `form:"deskripsi"        json:"deskripsi,omitempty"`
	Alamat         string `form:"alamat"           json:"alamat,omitempty"`
	TanggalMulai   string `form:"tanggal_mulai"    json:"tanggal_mulai,omitempty"`
	TanggalSelesai string `form:"tanggal_selesai"  json:"tanggal_selesai,omitempty"`
	InfoTiket      string `form:"info_tiket"       json:"info_tiket,omitempty"`
	HargaTiket     string `form:"harga_tiket"      json:"harga_tiket,omitempty"`
	Lat            string `form:"lat"              json:"lat,omitempty"`
	Lng            string `form:"lng"              json:"lng,omitempty"`
}



type EventActionResponse struct {
	ID                 int32  `json:"id"`
	Nama               string `json:"nama"`
	Slug               string `json:"slug"`
	GambarUrlThumbnail string `json:"gambar_url_thumbnail,omitempty"`
	GambarUrlHero      string `json:"gambar_url_hero,omitempty"`
}

type EventDetailResponse struct {
	Event db.GetEventBySlugRow `json:"event"`
}

type EventMapsResponse struct {
	Center MapsCenter            `json:"center"`
	Points []db.ListEventMapsRow `json:"points"`
}

type EventRecommendationResponse struct {
	Items []db.GetRecommendationEventRow `json:"items"`
}

type EventTahunTersediaResponse struct {
	Tahun []int16 `json:"tahun"`
}