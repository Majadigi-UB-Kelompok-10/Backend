-- name: GetAllArea :many
SELECT id, nama, slug FROM master_area ORDER BY nama ASC;

-- name: GetAreaBySlug :one
SELECT id, nama, slug FROM master_area WHERE slug = $1 LIMIT 1;

-- name: GetAllBahanPokok :many
SELECT id, nama, slug, satuan, gambar_url 
FROM bahan_pokok 
WHERE (@nama::text = '' OR nama ILIKE '%' || @nama || '%')
ORDER BY id ASC
LIMIT $1 OFFSET $2;

-- name: GetTotalBahanPokok :one
SELECT count(*) 
FROM bahan_pokok
WHERE (@nama::text = '' OR nama ILIKE '%' || @nama || '%');

-- name: GetBahanPokokBySlug :one
SELECT id, nama, slug, satuan, gambar_url 
FROM bahan_pokok 
WHERE slug = $1 LIMIT 1;

-- name: CreateBahanPokok :one
INSERT INTO bahan_pokok (nama, slug, satuan, gambar_url)
VALUES ($1, $2, $3, $4)
RETURNING id, nama, slug, satuan, gambar_url;

-- name: UpdateBahanPokok :one
UPDATE bahan_pokok
SET nama = $2,
    slug = $3,
    satuan = $4,
    gambar_url = $5
WHERE id = $1
RETURNING id, nama, slug, satuan, gambar_url;

-- name: DeleteBahanPokok :exec
DELETE FROM bahan_pokok WHERE id = $1;

-- name: GetTrenSemuaBahanPokok :many
SELECT 
    h_outer.bahan_pokok_id,
    h_outer.tanggal,
    ROUND(AVG(h_outer.harga))::int AS rata_rata_harga
FROM harga_harian h_outer
JOIN master_area ma ON h_outer.area_id = ma.id
WHERE h_outer.tanggal <= @tanggal::date
  AND (@area_slug::text = '' OR ma.slug = @area_slug)
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
        WHERE h_inner.tanggal <= @tanggal::date
          AND (@area_slug::text = '' OR ma_inner.slug = @area_slug)
    ) ranked
    WHERE ranked.urutan <= 2  
)
GROUP BY h_outer.bahan_pokok_id, h_outer.tanggal
ORDER BY h_outer.bahan_pokok_id ASC, h_outer.tanggal ASC;

-- name: GetTrenSemuaBahanPokokByArea :many
SELECT 
    h_outer.bahan_pokok_id,
    h_outer.tanggal,
    h_outer.harga AS rata_rata_harga
FROM harga_harian h_outer
JOIN master_area ma ON h_outer.area_id = ma.id
WHERE h_outer.tanggal <= @tanggal::date
  AND ma.slug = @area_slug
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
        WHERE h_inner.tanggal <= @tanggal::date
          AND ma_inner.slug = @area_slug
    ) ranked
    WHERE ranked.urutan <= 2  
)
ORDER BY h_outer.bahan_pokok_id ASC, h_outer.tanggal ASC;

-- name: GetTanggalUpdateTerakhir :one
SELECT tanggal 
FROM harga_harian 
WHERE bahan_pokok_id = $1 AND tanggal <= $2 
ORDER BY tanggal DESC LIMIT 1;

-- name: GetDaftarHargaArea :many
SELECT ma.nama AS area, ma.slug AS area_slug, h.harga
FROM harga_harian h
JOIN master_area ma ON h.area_id = ma.id
WHERE h.bahan_pokok_id = $1 AND h.tanggal = $2
ORDER BY ma.nama ASC;

-- name: GetRataRataArea15Hari :many
SELECT ma.nama AS area, ma.slug AS area_slug, ROUND(AVG(h.harga))::int AS rata_rata_15_hari
FROM harga_harian h
JOIN master_area ma ON h.area_id = ma.id
WHERE h.bahan_pokok_id = $1 
  AND h.tanggal <= $2 
  AND h.tanggal > ($2::date - INTERVAL '15 days')
GROUP BY ma.id, ma.nama, ma.slug
ORDER BY rata_rata_15_hari DESC;

-- name: GetRiwayatHargaRataRata :many
SELECT tanggal, ROUND(AVG(harga))::int AS rata_rata_harga
FROM harga_harian
WHERE bahan_pokok_id = $1 
  AND tanggal <= $2
  AND tanggal > ($2::date - INTERVAL '30 days')
GROUP BY tanggal
ORDER BY tanggal ASC;

-- name: CreateHargaHarian :one
INSERT INTO harga_harian (bahan_pokok_id, area_id, harga, tanggal)
VALUES ($1, $2, $3, $4)
ON CONFLICT (bahan_pokok_id, area_id, tanggal) 
DO UPDATE SET harga = EXCLUDED.harga
RETURNING id, bahan_pokok_id, area_id, harga, tanggal;

-- name: UpdateHargaHarian :one
UPDATE harga_harian SET harga = $2 WHERE id = $1
RETURNING id, bahan_pokok_id, area_id, harga, tanggal;

-- name: DeleteHargaHarian :exec
DELETE FROM harga_harian WHERE id = $1;

-- name: GetAreaIDByName :one
SELECT id FROM master_area 
WHERE nama ILIKE '%' || sqlc.arg('nama')::text || '%' 
LIMIT 1;

-- name: GetBahanPokokIDByName :one
SELECT id FROM bahan_pokok 
WHERE nama ILIKE '%' || sqlc.arg('nama')::text || '%' 
LIMIT 1;

-- name: GetBahanPokokByID :one
SELECT id, nama, slug, satuan, gambar_url
FROM bahan_pokok 
WHERE id = $1 LIMIT 1;

-- name: GetAllAreas :many
SELECT id, nama, slug 
FROM master_area
ORDER BY nama ASC;