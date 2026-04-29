-- name: GetKendaraanByPlatDanRangka :one
SELECT
    plat_nomor_display, status_aktif, merk, tipe, warna, tahun_buat, model, masa_pajak,
    pkb_pokok, opsen_pkb, swdkllj, parkir_berlangganan, total_pajak_tahunan,
    cetak_stnk, cetak_tnkb
FROM kendaraan_pajak
WHERE plat_nomor = sqlc.arg('plat_nomor')
  AND RIGHT(nomor_rangka, 5) = sqlc.arg('lima_digit_rangka');



-- name: GetAllTarifPKB :many
SELECT jenis_plat, label, tarif_pkb_persen, opsen_pkb_persen,
       bbn1_persen, opsen_bbn1_persen, bbn2_persen
FROM tarif_pkb
ORDER BY id;

-- name: GetDistinctJenis :many
SELECT DISTINCT jenis_kendaraan FROM master_njkb ORDER BY jenis_kendaraan ASC;

-- name: GetDistinctMerk :many
SELECT DISTINCT merk FROM master_njkb 
WHERE jenis_kendaraan = $1 
ORDER BY merk ASC;

-- name: GetDistinctModel :many
SELECT DISTINCT model FROM master_njkb 
WHERE jenis_kendaraan = $1 AND merk = $2 
ORDER BY model ASC;

-- name: GetDistinctTipe :many
SELECT DISTINCT tipe FROM master_njkb 
WHERE jenis_kendaraan = $1 AND merk = $2 AND model = $3 
ORDER BY tipe ASC;

-- name: GetDistinctTahun :many
SELECT DISTINCT tahun FROM master_njkb 
WHERE jenis_kendaraan = $1 AND merk = $2 AND model = $3 AND tipe = $4 
ORDER BY tahun DESC;

-- name: GetNilaiJual :one
SELECT id, nilai_jual
FROM master_njkb
WHERE jenis_kendaraan = sqlc.arg('jenis')
  AND merk            = sqlc.arg('merk')
  AND model           = sqlc.arg('model')
  AND tipe            = sqlc.arg('tipe')
  AND tahun           = sqlc.arg('tahun')
LIMIT 1;


-- name: GetMasterNjkbTree :many
SELECT jenis_kendaraan, merk, model, tipe, tahun, nilai_jual
FROM master_njkb
ORDER BY jenis_kendaraan, merk, model, tipe, tahun DESC;



-- name: CreateKendaraanPajak :one
INSERT INTO kendaraan_pajak (
    plat_nomor, plat_nomor_display, nomor_rangka, status_aktif,
    merk, tipe, warna, tahun_buat, model, masa_pajak,
    pkb_pokok, opsen_pkb, swdkllj, parkir_berlangganan,
    cetak_stnk, cetak_tnkb
) VALUES (
    sqlc.arg('plat_nomor'), sqlc.arg('plat_nomor_display'), sqlc.arg('nomor_rangka'),
    sqlc.arg('status_aktif'), sqlc.arg('merk'), sqlc.arg('tipe'), sqlc.arg('warna'),
    sqlc.arg('tahun_buat'), sqlc.arg('model'), sqlc.arg('masa_pajak'), sqlc.arg('pkb_pokok'),
    sqlc.arg('opsen_pkb'), sqlc.arg('swdkllj'), sqlc.arg('parkir_berlangganan'),
    sqlc.arg('cetak_stnk'), sqlc.arg('cetak_tnkb')
) RETURNING plat_nomor;

-- name: UpdateKendaraanPajak :exec
UPDATE kendaraan_pajak SET
    nomor_rangka        = sqlc.arg('nomor_rangka'),
    status_aktif        = sqlc.arg('status_aktif'),
    merk                = sqlc.arg('merk'),
    tipe                = sqlc.arg('tipe'),
    warna               = sqlc.arg('warna'),
    tahun_buat          = sqlc.arg('tahun_buat'),
    model               = sqlc.arg('model'),
    masa_pajak          = sqlc.arg('masa_pajak'),
    pkb_pokok           = sqlc.arg('pkb_pokok'),
    opsen_pkb           = sqlc.arg('opsen_pkb'),
    swdkllj             = sqlc.arg('swdkllj'),
    parkir_berlangganan = sqlc.arg('parkir_berlangganan'),
    cetak_stnk          = sqlc.arg('cetak_stnk'),
    cetak_tnkb          = sqlc.arg('cetak_tnkb')
WHERE plat_nomor = sqlc.arg('plat_nomor_key');

-- name: DeleteKendaraanPajak :exec
DELETE FROM kendaraan_pajak WHERE plat_nomor = $1;

-- name: GetKendaraanPajakByPlatAdmin :one
SELECT 
    plat_nomor, plat_nomor_display, nomor_rangka, status_aktif,
    merk, tipe, warna, tahun_buat, model, masa_pajak,
    pkb_pokok, opsen_pkb, swdkllj, parkir_berlangganan, total_pajak_tahunan,
    cetak_stnk, cetak_tnkb,
    created_at, updated_at
FROM kendaraan_pajak
WHERE plat_nomor = sqlc.arg('plat_nomor');

-- name: GetAllKendaraanPajakAdmin :many
SELECT 
    plat_nomor, plat_nomor_display, nomor_rangka, merk, tipe, 
    tahun_buat, status_aktif, masa_pajak
FROM kendaraan_pajak
ORDER BY created_at DESC
LIMIT sqlc.arg('limit_data') OFFSET sqlc.arg('offset_data');

-- name: CountKendaraanPajak :one
SELECT COUNT(*) FROM kendaraan_pajak;

-- name: EstimateKendaraanPajakCount :one
SELECT COALESCE(reltuples::bigint, 0) AS approx_count
FROM pg_class
WHERE relname = 'kendaraan_pajak';

-- =============================================================================
-- ADMIN: MASTER NJKB
-- =============================================================================

-- name: CreateMasterNjkb :one
INSERT INTO master_njkb (
    jenis_kendaraan, merk, model, tipe, tahun, nilai_jual
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;

-- name: UpdateMasterNjkb :exec
UPDATE master_njkb SET
    jenis_kendaraan = $2,
    merk            = $3,
    model           = $4,
    tipe            = $5,
    tahun           = $6,
    nilai_jual      = $7
WHERE id = $1;

-- name: DeleteMasterNjkb :exec
DELETE FROM master_njkb WHERE id = $1;

-- name: GetMasterNjkbById :one
SELECT id, jenis_kendaraan, merk, model, tipe, tahun, nilai_jual, created_at
FROM master_njkb
WHERE id = $1
LIMIT 1;

-- name: GetAllMasterNjkbAdmin :many
SELECT id, jenis_kendaraan, merk, model, tipe, tahun, nilai_jual, created_at
FROM master_njkb
ORDER BY merk, tahun DESC
LIMIT sqlc.arg('limit_data') OFFSET sqlc.arg('offset_data');

-- name: CountMasterNjkb :one
SELECT COUNT(*) FROM master_njkb;

-- name: EstimateMasterNjkbCount :one
SELECT COALESCE(reltuples::bigint, 0) AS approx_count
FROM pg_class
WHERE relname = 'master_njkb';