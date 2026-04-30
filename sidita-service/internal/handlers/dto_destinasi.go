package handlers

import "github.com/farildzaky/sidita-service/internal/db"

type CreateDestinasiRequest struct {
	AreaName      string `form:"area"           json:"area"`
	Kategori      string `form:"kategori"       json:"kategori"`
	Nama          string `form:"nama"           json:"nama"`
	Deskripsi     string `form:"deskripsi"      json:"deskripsi"`
	Alamat        string `form:"alamat"         json:"alamat"`
	HighlightText string `form:"highlight_text" json:"highlight_text"`
	Lat           string `form:"lat"            json:"lat"`
	Lng           string `form:"lng"            json:"lng"`
}

type UpdateDestinasiRequest struct {
	AreaName      string `form:"area"           json:"area,omitempty"`
	Kategori      string `form:"kategori"       json:"kategori,omitempty"`
	Nama          string `form:"nama"           json:"nama,omitempty"`
	Deskripsi     string `form:"deskripsi"      json:"deskripsi,omitempty"`
	Alamat        string `form:"alamat"         json:"alamat,omitempty"`
	HighlightText string `form:"highlight_text" json:"highlight_text,omitempty"`
	Lat           string `form:"lat"            json:"lat,omitempty"`
	Lng           string `form:"lng"            json:"lng,omitempty"`
}

type DestinasiActionResponse struct {
	ID                 int32  `json:"id"`
	Nama               string `json:"nama"`
	Slug               string `json:"slug"`
	Kota               string `json:"kota,omitempty"`
	Lat                string `json:"lat,omitempty"`
	Lng                string `json:"lng,omitempty"`
	GambarUrlThumbnail string `json:"gambar_url_thumbnail,omitempty"`
	GambarUrlHero      string `json:"gambar_url_hero,omitempty"`
}


type DestinasiDetailResponse struct {
	Destinasi db.GetDestinasiBySlugRow `json:"destinasi"`
}

type MapsCenter struct {
	Lat  interface{} `json:"lat"`
	Lng  interface{} `json:"lng"`
	Zoom int         `json:"zoom"`
}

type DestinasiMapsResponse struct {
	Center MapsCenter                `json:"center"`
	Points []db.ListDestinasiMapsRow `json:"points"`
}

type DestinasiRecommendationResponse struct {
	Items []db.GetRecommendationDestinasiRow `json:"items"`
}