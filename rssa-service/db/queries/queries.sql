-- name: GetSummaryRuangan :one
SELECT 
    COALESCE(SUM(kapasitas), 0)::int AS total_kapasitas,
    COALESCE(SUM(tersedia), 0)::int AS total_tersedia
FROM ruangan;

-- name: ListRuanganPublic :many
SELECT 
    r.id, 
    r.nama, 
    r.slug,
    r.kapasitas, 
    r.terisi, 
    r.tersedia,
    k.nama AS kelas_nama,
    k.slug AS kelas_slug
FROM ruangan r
JOIN master_kelas k ON r.kelas_id = k.id
WHERE 
    (sqlc.arg('kelas_slug')::text = '' OR k.slug = sqlc.arg('kelas_slug')::text)
    AND 
    (sqlc.arg('keyword')::text = '' OR r.nama ILIKE '%' || sqlc.arg('keyword')::text || '%')
ORDER BY r.nama ASC
LIMIT $2 OFFSET $1;

-- name: CountRuanganPublic :one
SELECT COUNT(r.id) 
FROM ruangan r
JOIN master_kelas k ON r.kelas_id = k.id
WHERE 
    (sqlc.arg('kelas_slug')::text = '' OR k.slug = sqlc.arg('kelas_slug')::text)
    AND 
    (sqlc.arg('keyword')::text = '' OR r.nama ILIKE '%' || sqlc.arg('keyword')::text || '%');

-- name: GetAllMasterKelas :many
-- Untuk tombol filter chips di UI
SELECT id, nama, slug 
FROM master_kelas 
ORDER BY nama ASC;

-- ==========================================
-- ADMIN: CRUD RUANGAN
-- ==========================================

-- name: CreateRuangan :one
INSERT INTO ruangan (kelas_id, nama, slug, kapasitas, terisi)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, kelas_id, nama, slug, kapasitas, terisi, tersedia, created_at, updated_at;

-- name: UpdateRuangan :one
UPDATE ruangan
SET 
    kelas_id = $2,
    nama = $3,
    slug = $4,
    kapasitas = $5,
    terisi = $6
WHERE id = $1
RETURNING id, kelas_id, nama, slug, kapasitas, terisi, tersedia, created_at, updated_at;

-- name: DeleteRuangan :exec
DELETE FROM ruangan
WHERE id = $1;

-- name: GetRuanganByID :one
SELECT 
    r.id, 
    r.kelas_id,
    k.nama AS kelas_nama,
    r.nama, 
    r.slug, 
    r.kapasitas, 
    r.terisi, 
    r.tersedia
FROM ruangan r
JOIN master_kelas k ON r.kelas_id = k.id
WHERE r.id = $1 LIMIT 1;

-- name: GetMasterKelas :many
SELECT id, nama, slug FROM master_kelas ORDER BY id ASC;

-- name: SearchRuangan :many
SELECT 
    r.id, 
    r.nama, 
    r.slug, 
    k.nama as kelas_nama, 
    k.slug as kelas_slug, 
    r.kapasitas, 
    r.terisi, 
    (r.kapasitas - r.terisi)::int as tersedia
FROM ruangan r
JOIN master_kelas k ON r.kelas_id = k.id
WHERE 
    (@keyword::text = '' OR r.nama ILIKE '%' || @keyword || '%') AND
    (@kelas_slug::text = '' OR k.slug = @kelas_slug)
ORDER BY r.nama ASC;

-- name: CountSearchRuangan :one
SELECT count(*)
FROM ruangan r
JOIN master_kelas k ON r.kelas_id = k.id
WHERE 
    (@keyword::text = '' OR r.nama ILIKE '%' || @keyword || '%') AND
    (@kelas_slug::text = '' OR k.slug = @kelas_slug);