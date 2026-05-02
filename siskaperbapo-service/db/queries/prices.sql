-- name: GetAllArea :many
SELECT id, nama, slug, created_at, updated_at 
FROM master_area 
ORDER BY nama ASC;

-- name: GetAreaBySlug :one
SELECT id, nama, slug, created_at, updated_at 
FROM master_area 
WHERE slug = sqlc.arg('slug') 
LIMIT 1;

-- name: GetAreaIDByName :one
SELECT id 
FROM master_area 
WHERE nama ILIKE '%' || sqlc.arg('nama')::text || '%' 
LIMIT 1;

-- name: GetAllBahanPokok :many
SELECT id, nama, slug, satuan, gambar_url, created_at, updated_at 
FROM bahan_pokok 
WHERE (sqlc.narg('keyword')::text IS NULL OR nama ILIKE '%' || sqlc.narg('keyword')::text || '%')
ORDER BY nama ASC
LIMIT sqlc.arg('limit_data') OFFSET sqlc.arg('offset_data');

-- name: CountBahanPokok :one
SELECT COUNT(id) 
FROM bahan_pokok
WHERE (sqlc.narg('keyword')::text IS NULL OR nama ILIKE '%' || sqlc.narg('keyword')::text || '%');

-- name: GetBahanPokokBySlug :one
SELECT id, nama, slug, satuan, gambar_url, created_at, updated_at 
FROM bahan_pokok 
WHERE slug = sqlc.arg('slug') 
LIMIT 1;

-- name: GetBahanPokokByID :one
SELECT id, nama, slug, satuan, gambar_url
FROM bahan_pokok 
WHERE id = sqlc.arg('id') 
LIMIT 1;

-- name: GetBahanPokokIDByName :one
SELECT id 
FROM bahan_pokok 
WHERE nama ILIKE '%' || sqlc.arg('nama')::text || '%' 
LIMIT 1;

-- name: CreateBahanPokok :one
INSERT INTO bahan_pokok (nama, slug, satuan, gambar_url)
VALUES (sqlc.arg('nama'), sqlc.arg('slug'), sqlc.arg('satuan'), sqlc.narg('gambar_url'))
RETURNING id, nama, slug;

-- name: UpdateBahanPokok :one
UPDATE bahan_pokok
SET nama       = sqlc.arg('nama'),
    slug       = sqlc.arg('slug'),
    satuan     = sqlc.arg('satuan'),
    gambar_url = sqlc.narg('gambar_url')
WHERE id = sqlc.arg('id')
RETURNING id, nama, slug;

-- name: DeleteBahanPokok :exec
DELETE FROM bahan_pokok WHERE id = sqlc.arg('id');

-- name: UpsertHargaHarian :one
INSERT INTO harga_harian (bahan_pokok_id, area_id, harga, tanggal)
VALUES (sqlc.arg('bahan_pokok_id'), sqlc.arg('area_id'), sqlc.arg('harga'), sqlc.arg('tanggal'))
ON CONFLICT (bahan_pokok_id, area_id, tanggal) 
DO UPDATE SET harga = EXCLUDED.harga
RETURNING id, bahan_pokok_id, area_id, harga, tanggal;

-- name: DeleteHargaHarian :exec
DELETE FROM harga_harian WHERE id = sqlc.arg('id');

-- name: GetTanggalUpdateTerakhir :one
SELECT tanggal 
FROM harga_harian 
WHERE bahan_pokok_id = sqlc.arg('bahan_pokok_id') 
  AND tanggal <= sqlc.arg('tanggal')::date 
ORDER BY tanggal DESC 
LIMIT 1;

-- name: GetDaftarHargaArea :many
SELECT ma.nama AS area, ma.slug AS area_slug, h.harga
FROM harga_harian h
JOIN master_area ma ON h.area_id = ma.id
WHERE h.bahan_pokok_id = sqlc.arg('bahan_pokok_id') 
  AND h.tanggal = sqlc.arg('tanggal')::date
ORDER BY ma.nama ASC;

-- name: GetRiwayatHargaRataRata :many
SELECT tanggal, ROUND(AVG(harga))::int AS rata_rata_harga
FROM harga_harian
WHERE bahan_pokok_id = sqlc.arg('bahan_pokok_id') 
  AND tanggal <= sqlc.arg('tanggal')::date
  AND tanggal > (sqlc.arg('tanggal')::date - INTERVAL '30 days')
GROUP BY tanggal
ORDER BY tanggal ASC;

-- name: GetRataRataArea15Hari :many
SELECT ma.nama AS area, ma.slug AS area_slug, ROUND(AVG(h.harga))::int AS rata_rata_15_hari
FROM harga_harian h
JOIN master_area ma ON h.area_id = ma.id
WHERE h.bahan_pokok_id = sqlc.arg('bahan_pokok_id') 
  AND h.tanggal <= sqlc.arg('tanggal')::date 
  AND h.tanggal > (sqlc.arg('tanggal')::date - INTERVAL '15 days')
GROUP BY ma.id, ma.nama, ma.slug
ORDER BY rata_rata_15_hari DESC;


-- name: GetTrenSemuaBahanPokok :many
SELECT 
    h_outer.bahan_pokok_id,
    h_outer.tanggal,
    ROUND(AVG(h_outer.harga))::int AS rata_rata_harga
FROM harga_harian h_outer
JOIN master_area ma ON h_outer.area_id = ma.id
WHERE h_outer.tanggal <= sqlc.arg('tanggal')::date
  AND (sqlc.narg('area_slug')::text IS NULL OR ma.slug = sqlc.narg('area_slug')::text)
  AND (h_outer.bahan_pokok_id, h_outer.tanggal) IN (
    SELECT ranked.bahan_pokok_id, ranked.tanggal
    FROM (
        SELECT 
            h_inner.bahan_pokok_id, 
            h_inner.tanggal,
            DENSE_RANK() OVER (
                PARTITION BY h_inner.bahan_pokok_id 
                ORDER BY h_inner.tanggal DESC
            ) AS urutan
        FROM harga_harian h_inner
        JOIN master_area ma_inner ON h_inner.area_id = ma_inner.id
        WHERE h_inner.tanggal <= sqlc.arg('tanggal')::date
          AND (sqlc.narg('area_slug')::text IS NULL OR ma_inner.slug = sqlc.narg('area_slug')::text)
    ) ranked
    WHERE ranked.urutan <= 2  
)
GROUP BY h_outer.bahan_pokok_id, h_outer.tanggal
ORDER BY h_outer.bahan_pokok_id ASC, h_outer.tanggal ASC;

-- name: GetTotalBahanPokok :one
SELECT COUNT(*) 
FROM bahan_pokok
WHERE sqlc.arg('keyword')::text IS NULL OR nama ILIKE '%' || sqlc.arg('keyword')::text || '%';

-- name: GetTrenSemuaBahanPokokByArea :many
SELECT 
    bp.id AS bahan_pokok_id,
    hh.tanggal,
    hh.harga AS rata_rata_harga
FROM bahan_pokok bp
JOIN harga_harian hh ON bp.id = hh.bahan_pokok_id
JOIN master_area ma ON hh.area_id = ma.id
WHERE hh.tanggal = $1 AND ma.slug = $2
ORDER BY bp.id;