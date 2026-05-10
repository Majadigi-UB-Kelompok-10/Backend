-- =====================================================================
-- BLK
-- =====================================================================

-- name: ListAktifBLK :many
SELECT id, nama, alamat, kab_kota, kecamatan, slug, lat, lng
FROM blk
WHERE aktif = true
  AND (sqlc.narg('kab_kota_filter')::text IS NULL OR kab_kota = sqlc.narg('kab_kota_filter')::text)
ORDER BY kab_kota, nama;

-- name: CountAktifBLK :one
SELECT COUNT(id)
FROM blk
WHERE aktif = true
  AND (sqlc.narg('kab_kota_filter')::text IS NULL OR kab_kota = sqlc.narg('kab_kota_filter')::text);

-- name: ListAllBLKAdmin :many
SELECT id, nama, alamat, kab_kota, kecamatan, slug, lat, lng, aktif
FROM blk
ORDER BY kab_kota, nama
LIMIT sqlc.arg('limit_data') OFFSET sqlc.arg('offset_data');

-- name: CountAllBLK :one
SELECT COUNT(id) FROM blk;

-- name: GetBLKByID :one
SELECT id, nama, alamat, kab_kota, kecamatan, slug, lat, lng, aktif
FROM blk
WHERE id = sqlc.arg('id')
LIMIT 1;

-- name: GetBLKBySlug :one
SELECT id, nama, alamat, kab_kota, kecamatan, slug, lat, lng, aktif
FROM blk
WHERE slug = sqlc.arg('slug')
LIMIT 1;

-- name: CreateBLK :one
INSERT INTO blk (nama, alamat, kab_kota, kecamatan, slug, lat, lng)
VALUES (
    sqlc.arg('nama'), sqlc.arg('alamat'), sqlc.arg('kab_kota'),
    sqlc.narg('kecamatan'), sqlc.arg('slug'),
    sqlc.narg('lat'), sqlc.narg('lng')
)
RETURNING id, slug;

-- name: UpdateBLK :exec
UPDATE blk
SET
    nama      = sqlc.arg('nama'),
    alamat    = sqlc.arg('alamat'),
    kab_kota  = sqlc.arg('kab_kota'),
    kecamatan = sqlc.narg('kecamatan'),
    slug      = sqlc.arg('slug_baru'),
    lat       = sqlc.narg('lat'),
    lng       = sqlc.narg('lng'),
    aktif     = sqlc.arg('aktif'),
    updated_at = CURRENT_TIMESTAMP
WHERE slug = sqlc.arg('slug_lama');

-- name: DeactivateBLK :exec
UPDATE blk
SET aktif = false, updated_at = CURRENT_TIMESTAMP
WHERE slug = sqlc.arg('slug');

-- name: GetDistinctKabKotaBLK :many
SELECT DISTINCT kab_kota FROM blk WHERE aktif = true ORDER BY kab_kota ASC;

-- =====================================================================
-- KEJURUAN
-- =====================================================================

-- name: ListAktifKejuruanByBLK :many
SELECT id, nama
FROM kejuruan
WHERE blk_id = sqlc.arg('blk_id')
  AND aktif = true
ORDER BY nama;

-- name: ListAllKejuruanByBLKAdmin :many
SELECT id, nama, aktif
FROM kejuruan
WHERE blk_id = sqlc.arg('blk_id')
ORDER BY nama;

-- name: GetKejuruanByID :one
SELECT id, blk_id, nama, aktif
FROM kejuruan
WHERE id = sqlc.arg('id')
LIMIT 1;

-- name: CreateKejuruan :one
INSERT INTO kejuruan (blk_id, nama)
VALUES (sqlc.arg('blk_id'), sqlc.arg('nama'))
RETURNING id;

-- name: UpdateKejuruan :exec
UPDATE kejuruan
SET nama = sqlc.arg('nama'), aktif = sqlc.arg('aktif'), updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id');

-- name: DeactivateKejuruan :exec
UPDATE kejuruan
SET aktif = false, updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id');

-- =====================================================================
-- PENDAFTARAN
-- =====================================================================

-- name: CreatePendaftaran :one
INSERT INTO pendaftaran (
    blk_id, kejuruan_id, nik, nama_lengkap, tempat_lahir, tanggal_lahir,
    email, jenis_kelamin, provinsi, kab_kota, kecamatan, kelurahan,
    rt, rw, alamat_lengkap, penyandang_disabilitas,
    no_wa, no_wa_darurat, pendidikan_terakhir, pendidikan_sekarang,
    asal_sekolah, jurusan, foto_url
)
VALUES (
    sqlc.arg('blk_id'), sqlc.arg('kejuruan_id'), sqlc.arg('nik'),
    sqlc.arg('nama_lengkap'), sqlc.arg('tempat_lahir'), sqlc.arg('tanggal_lahir'),
    sqlc.arg('email'), sqlc.arg('jenis_kelamin'), sqlc.arg('provinsi'),
    sqlc.arg('kab_kota'), sqlc.arg('kecamatan'), sqlc.arg('kelurahan'),
    sqlc.arg('rt'), sqlc.arg('rw'), sqlc.arg('alamat_lengkap'),
    sqlc.arg('penyandang_disabilitas'), sqlc.arg('no_wa'), sqlc.arg('no_wa_darurat'),
    sqlc.arg('pendidikan_terakhir'), sqlc.arg('pendidikan_sekarang'),
    sqlc.narg('asal_sekolah'), sqlc.narg('jurusan'), sqlc.arg('foto_url')
)
RETURNING id, status, created_at;

-- name: CekStatusPendaftaran :many
SELECT
    p.id, p.status, p.created_at,
    b.nama  AS blk_nama,
    b.slug  AS blk_slug,
    k.nama  AS kejuruan_nama
FROM pendaftaran p
JOIN blk      b ON b.id = p.blk_id
JOIN kejuruan k ON k.id = p.kejuruan_id
WHERE p.nik   = sqlc.arg('nik')
  AND p.no_wa = sqlc.arg('no_wa')
ORDER BY p.created_at DESC;

-- name: GetPendaftaranByID :one
SELECT
    p.id, p.nik, p.nama_lengkap, p.tempat_lahir, p.tanggal_lahir,
    p.email, p.jenis_kelamin, p.provinsi, p.kab_kota, p.kecamatan,
    p.kelurahan, p.rt, p.rw, p.alamat_lengkap, p.penyandang_disabilitas,
    p.no_wa, p.no_wa_darurat, p.pendidikan_terakhir, p.pendidikan_sekarang,
    p.asal_sekolah, p.jurusan, p.foto_url, p.status, p.created_at,
    b.nama AS blk_nama,
    k.nama AS kejuruan_nama
FROM pendaftaran p
JOIN blk      b ON b.id = p.blk_id
JOIN kejuruan k ON k.id = p.kejuruan_id
WHERE p.id = sqlc.arg('id')
LIMIT 1;

-- name: ListPendaftaranAdmin :many
SELECT
    p.id, p.nik, p.nama_lengkap, p.no_wa, p.status, p.created_at,
    b.nama AS blk_nama,
    k.nama AS kejuruan_nama
FROM pendaftaran p
JOIN blk      b ON b.id = p.blk_id
JOIN kejuruan k ON k.id = p.kejuruan_id
WHERE (sqlc.narg('blk_id_filter')::int IS NULL OR p.blk_id = sqlc.narg('blk_id_filter')::int)
  AND (sqlc.narg('status_filter')::status_daftar_type IS NULL OR p.status = sqlc.narg('status_filter')::status_daftar_type)
ORDER BY p.created_at DESC
LIMIT sqlc.arg('limit_data') OFFSET sqlc.arg('offset_data');

-- name: CountPendaftaran :one
SELECT COUNT(p.id)
FROM pendaftaran p
WHERE (sqlc.narg('blk_id_filter')::int IS NULL OR p.blk_id = sqlc.narg('blk_id_filter')::int)
  AND (sqlc.narg('status_filter')::status_daftar_type IS NULL OR p.status = sqlc.narg('status_filter')::status_daftar_type);

-- name: UpdateStatusPendaftaran :exec
UPDATE pendaftaran
SET status = sqlc.arg('status'), updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id');
