-- name: GetAllArea :many
SELECT id, nama, slug, lat, lng 
FROM master_area 
ORDER BY nama ASC;

-- name: GetAreaBySlug :one
SELECT id, nama, slug, lat, lng 
FROM master_area 
WHERE slug = $1 LIMIT 1;

-- name: GetAreaByName :one
SELECT id, nama, slug, lat, lng 
FROM master_area 
WHERE nama ILIKE $1 LIMIT 1;

-- name: GetDestinasiByID :one
SELECT * FROM destinasi WHERE id = $1 LIMIT 1;

-- name: GetAreaByID :one
SELECT * FROM master_area WHERE id = $1 LIMIT 1;

-- name: ListDestinasiMaps :many
SELECT id, nama, slug, kategori, lat, lng 
FROM destinasi
WHERE (sqlc.narg('area_id')::int IS NULL OR area_id = sqlc.narg('area_id')::int);

-- name: ListDestinasi :many
SELECT 
    d.id, d.kategori, d.nama, d.slug, d.gambar_url, d.alamat, a.nama AS kota
FROM destinasi d
JOIN master_area a ON d.area_id = a.id
WHERE 
    (sqlc.narg('search')::text IS NULL OR 
     d.nama ILIKE '%' || sqlc.narg('search')::text || '%' OR 
     d.alamat ILIKE '%' || sqlc.narg('search')::text || '%') 
    AND (sqlc.narg('kategori')::text IS NULL OR d.kategori = sqlc.narg('kategori')::text)
    AND (sqlc.narg('area_id')::int IS NULL OR d.area_id = sqlc.narg('area_id')::int)
ORDER BY d.created_at DESC
LIMIT @limit_data::int OFFSET @offset_data::int;

-- name: CountDestinasi :one
SELECT COUNT(*) 
FROM destinasi d
WHERE 
    (sqlc.narg('search')::text IS NULL OR 
     d.nama ILIKE '%' || sqlc.narg('search')::text || '%' OR 
     d.alamat ILIKE '%' || sqlc.narg('search')::text || '%') 
    AND (sqlc.narg('kategori')::text IS NULL OR d.kategori = sqlc.narg('kategori')::text)
    AND (sqlc.narg('area_id')::int IS NULL OR d.area_id = sqlc.narg('area_id')::int);

-- name: GetDestinasiBySlug :one
SELECT 
    d.id, d.kategori, d.nama, d.slug, d.gambar_url, d.deskripsi, 
    d.alamat, d.highlight_text, d.lat, d.lng, a.nama AS kota
FROM destinasi d
JOIN master_area a ON d.area_id = a.id
WHERE d.slug = $1 LIMIT 1;

-- name: CreateDestinasi :one
INSERT INTO destinasi (
    area_id, kategori, nama, slug, gambar_url, deskripsi, alamat, highlight_text, lat, lng
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING id, created_at;

-- name: UpdateDestinasi :one
UPDATE destinasi
SET 
    area_id = $2, kategori = $3, nama = $4, slug = $5, gambar_url = $6, 
    deskripsi = $7, alamat = $8, highlight_text = $9, lat = $10, lng = $11
WHERE id = $1
RETURNING id, nama, slug;

-- name: DeleteDestinasi :exec
DELETE FROM destinasi WHERE id = $1;

-- name: ListDestinasiGambar :many
SELECT id, gambar_url, urutan 
FROM destinasi_gambar 
WHERE destinasi_id = $1 
ORDER BY urutan ASC;

-- name: CreateDestinasiGambar :one
INSERT INTO destinasi_gambar (destinasi_id, gambar_url, urutan)
VALUES ($1, $2, $3)
RETURNING id;

-- name: ListHotel :many
SELECT 
    h.id, h.nama, h.slug, h.harga_mulai, h.bintang, h.gambar_url, a.nama AS kota
FROM hotel h
JOIN master_area a ON h.area_id = a.id
WHERE 
    (sqlc.narg('search')::text IS NULL OR h.nama ILIKE '%' || sqlc.narg('search')::text || '%')
    AND (sqlc.narg('area_id')::int IS NULL OR h.area_id = sqlc.narg('area_id')::int)
    AND (sqlc.narg('min_bintang')::smallint IS NULL OR h.bintang >= sqlc.narg('min_bintang')::smallint)
ORDER BY h.bintang DESC, h.harga_mulai ASC
LIMIT @limit_data::int OFFSET @offset_data::int;

-- name: CountHotel :one
SELECT COUNT(*) 
FROM hotel h
WHERE 
    (sqlc.narg('search')::text IS NULL OR h.nama ILIKE '%' || sqlc.narg('search')::text || '%')
    AND (sqlc.narg('area_id')::int IS NULL OR h.area_id = sqlc.narg('area_id')::int)
    AND (sqlc.narg('min_bintang')::smallint IS NULL OR h.bintang >= sqlc.narg('min_bintang')::smallint);

-- name: GetHotelBySlug :one
SELECT 
    h.id, h.nama, h.slug, h.harga_mulai, h.bintang, h.gambar_url, 
    h.deskripsi, h.alamat, h.highlight_text, h.lat, h.lng, a.nama AS kota
FROM hotel h
JOIN master_area a ON h.area_id = a.id
WHERE h.slug = $1 LIMIT 1;

-- name: CreateHotel :one
INSERT INTO hotel (
    area_id, nama, slug, harga_mulai, bintang, gambar_url, deskripsi, alamat, highlight_text, lat, lng
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING id, created_at;

-- name: ListEvent :many
SELECT 
    e.id, e.nama, e.slug, e.gambar_url, e.tanggal_mulai, e.tanggal_selesai, 
    e.tahun, e.info_tiket, a.nama AS kota
FROM event e
JOIN master_area a ON e.area_id = a.id
WHERE 
    (sqlc.narg('search')::text IS NULL OR e.nama ILIKE '%' || sqlc.narg('search')::text || '%')
    AND (sqlc.narg('area_id')::int IS NULL OR e.area_id = sqlc.narg('area_id')::int)
    AND (sqlc.narg('tahun')::smallint IS NULL OR e.tahun = sqlc.narg('tahun')::smallint)
ORDER BY e.tanggal_mulai ASC
LIMIT @limit_data::int OFFSET @offset_data::int;

-- name: CountEvent :one
SELECT COUNT(*) 
FROM event e
WHERE 
    (sqlc.narg('search')::text IS NULL OR e.nama ILIKE '%' || sqlc.narg('search')::text || '%')
    AND (sqlc.narg('area_id')::int IS NULL OR e.area_id = sqlc.narg('area_id')::int)
    AND (sqlc.narg('tahun')::smallint IS NULL OR e.tahun = sqlc.narg('tahun')::smallint);

-- name: GetEventBySlug :one
SELECT 
    e.id, e.nama, e.slug, e.gambar_url, e.deskripsi, e.tanggal_mulai, 
    e.tanggal_selesai, e.tahun, e.info_tiket, e.lat, e.lng, a.nama AS kota
FROM event e
JOIN master_area a ON e.area_id = a.id
WHERE e.slug = $1 LIMIT 1;

-- name: CreateEvent :one
INSERT INTO event (
    area_id, nama, slug, gambar_url, deskripsi, tanggal_mulai, tanggal_selesai, info_tiket, lat, lng
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING id, created_at;