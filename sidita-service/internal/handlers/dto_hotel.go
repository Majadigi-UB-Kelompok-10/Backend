package handlers

import "github.com/farildzaky/sidita-service/internal/db"

type CreateHotelRequest struct {
	AreaName      string `form:"area"           json:"area"`
	Nama          string `form:"nama"           json:"nama"`
	Bintang       string `form:"bintang"        json:"bintang"`
	HargaMulai    string `form:"harga_mulai"    json:"harga_mulai"`
	Deskripsi     string `form:"deskripsi"      json:"deskripsi,omitempty"`
	Alamat        string `form:"alamat"         json:"alamat"`
	HighlightText string `form:"highlight_text" json:"highlight_text,omitempty"`
	Lat           string `form:"lat"            json:"lat"`
	Lng           string `form:"lng"            json:"lng"`
}

type UpdateHotelRequest struct {
	AreaName      string `form:"area"           json:"area,omitempty"`
	Nama          string `form:"nama"           json:"nama,omitempty"`
	Bintang       string `form:"bintang"        json:"bintang,omitempty"`
	HargaMulai    string `form:"harga_mulai"    json:"harga_mulai,omitempty"`
	Deskripsi     string `form:"deskripsi"      json:"deskripsi,omitempty"`
	Alamat        string `form:"alamat"         json:"alamat,omitempty"`
	HighlightText string `form:"highlight_text" json:"highlight_text,omitempty"`
	Lat           string `form:"lat"            json:"lat,omitempty"`
	Lng           string `form:"lng"            json:"lng,omitempty"`
}



type HotelActionResponse struct {
	ID         int32  `json:"id"`
	Nama       string `json:"nama"`
	Slug       string `json:"slug"`
	Kota       string `json:"kota,omitempty"`
	Bintang    int16  `json:"bintang,omitempty"`
	HargaMulai int32  `json:"harga_mulai,omitempty"`
	Lat        string `json:"lat,omitempty"`
	Lng        string `json:"lng,omitempty"`
	GambarUrl  string `json:"gambar_url,omitempty"`
}

type HotelMapsResponse struct {
	Center MapsCenter            `json:"center"`
	Points []db.ListHotelMapsRow `json:"points"`
}

type HotelRecommendationResponse struct {
	Items []db.GetRecommendationHotelRow `json:"items"`
}