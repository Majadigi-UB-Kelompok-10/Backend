-- name: GetKendaraanByPlatDanRangka :one
SELECT
    plat_nomor_display, status_aktif, merk, tipe, warna, tahun_buat, model, masa_pajak,
    pkb_pokok, opsen_pkb, swdkllj, parkir_berlangganan, total_pajak_tahunan, cetak_stnk, cetak_tnkb
FROM kendaraan_pajak
WHERE plat_nomor = UPPER(REPLACE($1, ' ', ''))
AND RIGHT(nomor_rangka, 5) = UPPER($2);

-- name: UpsertKendaraanPajak :one
INSERT INTO kendaraan_pajak (
    plat_nomor, plat_nomor_display, nomor_rangka, status_aktif,
    merk, tipe, warna, tahun_buat, model, masa_pajak,
    pkb_pokok, opsen_pkb, swdkllj, parkir_berlangganan,
    cetak_stnk, cetak_tnkb
) VALUES (
    UPPER(REPLACE($1, ' ', '')), $2, $3, $4,
    $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16
)
ON CONFLICT (plat_nomor) DO UPDATE SET
    nomor_rangka        = EXCLUDED.nomor_rangka,
    status_aktif        = EXCLUDED.status_aktif,
    merk                = EXCLUDED.merk,
    tipe                = EXCLUDED.tipe,
    warna               = EXCLUDED.warna,
    tahun_buat          = EXCLUDED.tahun_buat,
    model               = EXCLUDED.model,
    masa_pajak          = EXCLUDED.masa_pajak,
    pkb_pokok           = EXCLUDED.pkb_pokok,
    opsen_pkb           = EXCLUDED.opsen_pkb,
    swdkllj             = EXCLUDED.swdkllj,
    parkir_berlangganan = EXCLUDED.parkir_berlangganan,
    cetak_stnk          = EXCLUDED.cetak_stnk,
    cetak_tnkb          = EXCLUDED.cetak_tnkb,
    updated_at          = CURRENT_TIMESTAMP
RETURNING plat_nomor;

-- name: GetAllTarifPKB :many
SELECT jenis_plat, label, tarif_pkb_persen, opsen_pkb_persen,
       bbn1_persen, opsen_bbn1_persen, bbn2_persen
FROM tarif_pkb
ORDER BY id;

-- name: GetDistinctJenis :many
SELECT DISTINCT jenis_kendaraan FROM master_njkb ORDER BY jenis_kendaraan ASC;

-- name: GetDistinctMerk :many
SELECT DISTINCT merk FROM master_njkb WHERE jenis_kendaraan = $1 ORDER BY merk ASC;

-- name: GetDistinctModel :many
SELECT DISTINCT model FROM master_njkb WHERE jenis_kendaraan = $1 AND merk = $2 ORDER BY model ASC;

-- name: GetDistinctTipe :many
SELECT DISTINCT tipe FROM master_njkb WHERE jenis_kendaraan = $1 AND merk = $2 AND model = $3 ORDER BY tipe ASC;

-- name: GetDistinctTahun :many
SELECT DISTINCT tahun FROM master_njkb WHERE jenis_kendaraan = $1 AND merk = $2 AND model = $3 AND tipe = $4 ORDER BY tahun DESC;

-- name: GetNilaiJual :one
SELECT id, nilai_jual
FROM master_njkb
WHERE jenis_kendaraan = $1 AND merk = $2 AND model = $3 AND tipe = $4 AND tahun = $5
LIMIT 1;

-- name: UpsertMasterNjkb :one
INSERT INTO master_njkb (
    jenis_kendaraan, merk, model, tipe, tahun, nilai_jual
) VALUES (
    $1, $2, $3, $4, $5, $6
)
ON CONFLICT (merk, tipe, tahun) DO UPDATE SET
    jenis_kendaraan = EXCLUDED.jenis_kendaraan,
    model           = EXCLUDED.model,
    nilai_jual      = EXCLUDED.nilai_jual
RETURNING id;