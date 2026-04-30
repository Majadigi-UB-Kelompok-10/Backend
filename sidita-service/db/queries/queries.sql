-- name: GetAllArea :many
SELECT id, nama, slug, lat, lng
FROM master_area
ORDER BY nama ASC;

-- name: GetAreaBySlug :one
SELECT id, nama, slug, lat, lng
FROM master_area
WHERE slug = $1
LIMIT 1;

-- name: GetAreaByID :one
SELECT id, nama, slug, lat, lng
FROM master_area
WHERE id = $1
LIMIT 1;

-- =============================================================================
-- PUBLIC: DESTINASI — list, detail, maps, recommendation
-- =============================================================================

-- name: ListDestinasi :many
SELECT
    d.id, d.nama, d.slug, d.kategori, d.alamat, d.highlight_text,
    d.gambar_url_thumbnail, d.lat, d.lng,
    a.nama AS area_nama, a.slug AS area_slug
FROM destinasi d
JOIN master_area a ON d.area_id = a.id
WHERE
    (sqlc.narg('area_id')::int IS NULL OR d.area_id = sqlc.narg('area_id')::int)
    AND (sqlc.narg('keyword')::text IS NULL OR d.nama ILIKE '%' || sqlc.narg('keyword')::text || '%')
ORDER BY d.created_at DESC
LIMIT sqlc.arg('limit_data') OFFSET sqlc.arg('offset_data');

-- name: CountDestinasi :one
SELECT COUNT(id)
FROM destinasi d
WHERE
    (sqlc.narg('area_id')::int IS NULL OR d.area_id = sqlc.narg('area_id')::int)
    AND (sqlc.narg('keyword')::text IS NULL OR d.nama ILIKE '%' || sqlc.narg('keyword')::text || '%');

-- name: GetDestinasiBySlug :one
SELECT
    d.id, d.nama, d.slug, d.kategori, d.deskripsi, d.alamat, d.highlight_text,
    d.gambar_url_hero, d.lat, d.lng, d.created_at,
    a.id AS area_id, a.nama AS area_nama, a.slug AS area_slug
FROM destinasi d
JOIN master_area a ON d.area_id = a.id
WHERE d.slug = $1
LIMIT 1;

-- name: GetDestinasiByID :one
SELECT
    d.id, d.area_id, d.nama, d.slug, d.kategori, d.deskripsi, d.alamat,
    d.highlight_text, d.gambar_url_thumbnail, d.gambar_url_hero,
    d.lat, d.lng, d.created_at, d.updated_at
FROM destinasi d
WHERE d.id = $1
LIMIT 1;

-- name: ListDestinasiMaps :many
SELECT id, nama, slug, kategori, gambar_url_thumbnail, lat, lng
FROM destinasi
WHERE
    (sqlc.narg('area_id')::int IS NULL OR area_id = sqlc.narg('area_id')::int);

-- name: GetRecommendationDestinasi :many
SELECT
    d.id, d.nama, d.slug, d.kategori, d.highlight_text,
    d.gambar_url_thumbnail, d.lat, d.lng,
    a.nama AS area_nama
FROM destinasi d
JOIN master_area a ON d.area_id = a.id
ORDER BY d.created_at DESC
LIMIT 6;

-- name: GetDestinasiByIDPublic :one
-- Mobile app pake ini (lebih efisien dari slug lookup).
SELECT
    d.id, d.nama, d.slug, d.kategori, d.deskripsi, d.alamat, d.highlight_text,
    d.gambar_url_hero, d.lat, d.lng, d.created_at,
    a.id AS area_id, a.nama AS area_nama, a.slug AS area_slug
FROM destinasi d
JOIN master_area a ON d.area_id = a.id
WHERE d.id = $1
LIMIT 1;

-- =============================================================================
-- ADMIN: DESTINASI — CRUD
-- =============================================================================

-- name: CreateDestinasi :one
INSERT INTO destinasi (
    area_id, nama, slug, kategori, deskripsi, alamat, highlight_text,
    gambar_url_thumbnail, gambar_url_hero, lat, lng
) VALUES (
    sqlc.arg('area_id'), sqlc.arg('nama'), sqlc.arg('slug'), sqlc.arg('kategori'),
    sqlc.arg('deskripsi'), sqlc.arg('alamat'), sqlc.narg('highlight_text'),
    sqlc.arg('gambar_url_thumbnail'), sqlc.arg('gambar_url_hero'),
    sqlc.arg('lat'), sqlc.arg('lng')
) RETURNING id, slug;

-- name: UpdateDestinasi :exec
UPDATE destinasi SET
    area_id              = sqlc.arg('area_id'),
    nama                 = sqlc.arg('nama'),
    slug                 = sqlc.arg('slug'),
    kategori             = sqlc.arg('kategori'),
    deskripsi            = sqlc.arg('deskripsi'),
    alamat               = sqlc.arg('alamat'),
    highlight_text       = sqlc.narg('highlight_text'),
    gambar_url_thumbnail = sqlc.arg('gambar_url_thumbnail'),
    gambar_url_hero      = sqlc.arg('gambar_url_hero'),
    lat                  = sqlc.arg('lat'),
    lng                  = sqlc.arg('lng')
WHERE id = sqlc.arg('id');

-- name: DeleteDestinasi :exec
DELETE FROM destinasi WHERE id = $1;

-- =============================================================================
-- PUBLIC: HOTEL — list, maps, recommendation (NO detail by slug)
-- =============================================================================

-- name: ListHotel :many
SELECT
    h.id, h.nama, h.slug, h.bintang, h.harga_mulai, h.alamat, h.highlight_text,
    h.gambar_url, h.lat, h.lng,
    a.nama AS area_nama, a.slug AS area_slug
FROM hotel h
JOIN master_area a ON h.area_id = a.id
WHERE
    (sqlc.narg('area_id')::int IS NULL OR h.area_id = sqlc.narg('area_id')::int)
    AND (sqlc.narg('keyword')::text IS NULL OR h.nama ILIKE '%' || sqlc.narg('keyword')::text || '%')
ORDER BY h.bintang DESC, h.created_at DESC
LIMIT sqlc.arg('limit_data') OFFSET sqlc.arg('offset_data');

-- name: CountHotel :one
SELECT COUNT(id)
FROM hotel h
WHERE
    (sqlc.narg('area_id')::int IS NULL OR h.area_id = sqlc.narg('area_id')::int)
    AND (sqlc.narg('keyword')::text IS NULL OR h.nama ILIKE '%' || sqlc.narg('keyword')::text || '%');

-- name: GetHotelByID :one
SELECT
    h.id, h.area_id, h.nama, h.slug, h.bintang, h.harga_mulai,
    h.deskripsi, h.alamat, h.highlight_text, h.gambar_url,
    h.lat, h.lng, h.created_at, h.updated_at
FROM hotel h
WHERE h.id = $1
LIMIT 1;

-- name: ListHotelMaps :many
SELECT id, nama, slug, bintang, gambar_url, lat, lng
FROM hotel
WHERE
    (sqlc.narg('area_id')::int IS NULL OR area_id = sqlc.narg('area_id')::int);

-- name: GetRecommendationHotel :many
SELECT
    h.id, h.nama, h.slug, h.bintang, h.harga_mulai, h.highlight_text,
    h.gambar_url, h.lat, h.lng,
    a.nama AS area_nama
FROM hotel h
JOIN master_area a ON h.area_id = a.id
ORDER BY h.created_at DESC
LIMIT 6;

-- =============================================================================
-- ADMIN: HOTEL — CRUD
-- =============================================================================

-- name: CreateHotel :one
INSERT INTO hotel (
    area_id, nama, slug, bintang, harga_mulai, deskripsi, alamat,
    highlight_text, gambar_url, lat, lng
) VALUES (
    sqlc.arg('area_id'), sqlc.arg('nama'), sqlc.arg('slug'), sqlc.arg('bintang'),
    sqlc.arg('harga_mulai'), sqlc.narg('deskripsi'), sqlc.arg('alamat'),
    sqlc.narg('highlight_text'), sqlc.arg('gambar_url'),
    sqlc.arg('lat'), sqlc.arg('lng')
) RETURNING id, slug;

-- name: UpdateHotel :exec
UPDATE hotel SET
    area_id        = sqlc.arg('area_id'),
    nama           = sqlc.arg('nama'),
    slug           = sqlc.arg('slug'),
    bintang        = sqlc.arg('bintang'),
    harga_mulai    = sqlc.arg('harga_mulai'),
    deskripsi      = sqlc.narg('deskripsi'),
    alamat         = sqlc.arg('alamat'),
    highlight_text = sqlc.narg('highlight_text'),
    gambar_url     = sqlc.arg('gambar_url'),
    lat            = sqlc.arg('lat'),
    lng            = sqlc.arg('lng')
WHERE id = sqlc.arg('id');

-- name: DeleteHotel :exec
DELETE FROM hotel WHERE id = $1;

-- =============================================================================
-- PUBLIC: EVENT — list, detail, maps, recommendation
-- Multi-filter: area, tahun, bulan, search keyword.
-- =============================================================================

-- name: ListEvent :many
SELECT
    e.id, e.nama, e.slug, e.alamat, e.tanggal_mulai, e.tanggal_selesai,
    e.harga_tiket, e.gambar_url_thumbnail, e.tahun, e.bulan,
    a.nama AS area_nama, a.slug AS area_slug
FROM event e
JOIN master_area a ON e.area_id = a.id
WHERE
    (sqlc.narg('area_id')::int IS NULL OR e.area_id = sqlc.narg('area_id')::int)
    AND (sqlc.narg('tahun')::int IS NULL OR e.tahun = sqlc.narg('tahun')::int)
    AND (sqlc.narg('bulan')::int IS NULL OR e.bulan = sqlc.narg('bulan')::int)
    AND (sqlc.narg('keyword')::text IS NULL OR e.nama ILIKE '%' || sqlc.narg('keyword')::text || '%')
ORDER BY e.tanggal_mulai DESC
LIMIT sqlc.arg('limit_data') OFFSET sqlc.arg('offset_data');

-- name: CountEvent :one
SELECT COUNT(id)
FROM event e
WHERE
    (sqlc.narg('area_id')::int IS NULL OR e.area_id = sqlc.narg('area_id')::int)
    AND (sqlc.narg('tahun')::int IS NULL OR e.tahun = sqlc.narg('tahun')::int)
    AND (sqlc.narg('bulan')::int IS NULL OR e.bulan = sqlc.narg('bulan')::int)
    AND (sqlc.narg('keyword')::text IS NULL OR e.nama ILIKE '%' || sqlc.narg('keyword')::text || '%');

-- name: GetEventBySlug :one
SELECT
    e.id, e.nama, e.slug, e.deskripsi, e.alamat,
    e.tanggal_mulai, e.tanggal_selesai, e.info_tiket, e.harga_tiket,
    e.gambar_url_hero, e.lat, e.lng, e.created_at,
    a.id AS area_id, a.nama AS area_nama, a.slug AS area_slug
FROM event e
JOIN master_area a ON e.area_id = a.id
WHERE e.slug = $1
LIMIT 1;

-- name: GetEventByID :one
SELECT
    e.id, e.area_id, e.nama, e.slug, e.deskripsi, e.alamat,
    e.tanggal_mulai, e.tanggal_selesai, e.info_tiket, e.harga_tiket,
    e.gambar_url_thumbnail, e.gambar_url_hero,
    e.lat, e.lng, e.tahun, e.bulan, e.created_at, e.updated_at
FROM event e
WHERE e.id = $1
LIMIT 1;

-- name: ListEventMaps :many
SELECT id, nama, slug, gambar_url_thumbnail, tanggal_mulai, lat, lng
FROM event
WHERE
    (sqlc.narg('area_id')::int IS NULL OR area_id = sqlc.narg('area_id')::int)
    AND (sqlc.narg('tahun')::int IS NULL OR tahun = sqlc.narg('tahun')::int);

-- name: GetRecommendationEvent :many
SELECT
    e.id, e.nama, e.slug, e.tanggal_mulai, e.tanggal_selesai,
    e.harga_tiket, e.gambar_url_thumbnail, e.tahun, e.bulan,
    a.nama AS area_nama
FROM event e
JOIN master_area a ON e.area_id = a.id
ORDER BY e.created_at DESC
LIMIT 6;

-- name: GetEventByIDPublic :one
SELECT
    e.id, e.nama, e.slug, e.deskripsi, e.alamat,
    e.tanggal_mulai, e.tanggal_selesai, e.info_tiket, e.harga_tiket,
    e.gambar_url_hero, e.lat, e.lng, e.created_at,
    a.id AS area_id, a.nama AS area_nama, a.slug AS area_slug
FROM event e
JOIN master_area a ON e.area_id = a.id
WHERE e.id = $1
LIMIT 1;

-- name: GetAvailableTahunEvent :many
SELECT DISTINCT tahun FROM event ORDER BY tahun DESC;

-- =============================================================================
-- ADMIN: EVENT — CRUD
-- =============================================================================

-- name: CreateEvent :one
INSERT INTO event (
    area_id, nama, slug, deskripsi, alamat,
    tanggal_mulai, tanggal_selesai, info_tiket, harga_tiket,
    gambar_url_thumbnail, gambar_url_hero, lat, lng
) VALUES (
    sqlc.arg('area_id'), sqlc.arg('nama'), sqlc.arg('slug'), sqlc.arg('deskripsi'),
    sqlc.arg('alamat'), sqlc.arg('tanggal_mulai'), sqlc.arg('tanggal_selesai'),
    sqlc.narg('info_tiket'), sqlc.arg('harga_tiket'),
    sqlc.arg('gambar_url_thumbnail'), sqlc.arg('gambar_url_hero'),
    sqlc.arg('lat'), sqlc.arg('lng')
) RETURNING id, slug;

-- name: UpdateEvent :exec
UPDATE event SET
    area_id              = sqlc.arg('area_id'),
    nama                 = sqlc.arg('nama'),
    slug                 = sqlc.arg('slug'),
    deskripsi            = sqlc.arg('deskripsi'),
    alamat               = sqlc.arg('alamat'),
    tanggal_mulai        = sqlc.arg('tanggal_mulai'),
    tanggal_selesai      = sqlc.arg('tanggal_selesai'),
    info_tiket           = sqlc.narg('info_tiket'),
    harga_tiket          = sqlc.arg('harga_tiket'),
    gambar_url_thumbnail = sqlc.arg('gambar_url_thumbnail'),
    gambar_url_hero      = sqlc.arg('gambar_url_hero'),
    lat                  = sqlc.arg('lat'),
    lng                  = sqlc.arg('lng')
WHERE id = sqlc.arg('id');

-- name: DeleteEvent :exec
DELETE FROM event WHERE id = $1;