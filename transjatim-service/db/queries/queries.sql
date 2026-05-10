-- =====================================================================
-- TERMINAL
-- =====================================================================

-- name: ListActiveTerminals :many
SELECT id, nama, kota, slug, lat, lng
FROM terminal
WHERE aktif = true
ORDER BY kota, nama;

-- name: ListAllTerminalsAdmin :many
SELECT id, nama, kota, slug, lat, lng, aktif
FROM terminal
WHERE (sqlc.narg('kota_filter')::text IS NULL OR kota = sqlc.narg('kota_filter')::text)
ORDER BY kota, nama
LIMIT sqlc.arg('limit_data') OFFSET sqlc.arg('offset_data');

-- name: CountAllTerminals :one
SELECT COUNT(id) 
FROM terminal
WHERE (sqlc.narg('kota_filter')::text IS NULL OR kota = sqlc.narg('kota_filter')::text);

-- name: GetTerminalByID :one
SELECT id, nama, kota, slug, lat, lng, aktif
FROM terminal
WHERE id = sqlc.arg('id')
LIMIT 1;

-- name: GetTerminalBySlug :one
SELECT id, nama, kota, slug, lat, lng, aktif
FROM terminal
WHERE slug = sqlc.arg('slug')
LIMIT 1;

-- name: CreateTerminal :one
INSERT INTO terminal (nama, kota, slug, lat, lng)
VALUES (
    sqlc.arg('nama'), sqlc.arg('kota'), sqlc.arg('slug'), 
    sqlc.arg('lat'), sqlc.arg('lng')
)
RETURNING id, slug;

-- name: UpdateTerminal :exec
UPDATE terminal
SET 
    nama       = sqlc.arg('nama'), 
    kota       = sqlc.arg('kota'), 
    slug       = sqlc.arg('slug_baru'), 
    lat        = sqlc.arg('lat'), 
    lng        = sqlc.arg('lng'), 
    aktif      = sqlc.arg('aktif'), 
    updated_at = CURRENT_TIMESTAMP
WHERE slug = sqlc.arg('slug_lama');

-- name: DeactivateTerminal :exec
UPDATE terminal 
SET aktif = false, updated_at = CURRENT_TIMESTAMP 
WHERE slug = sqlc.arg('slug');

-- name: GetDistinctKotaTerminal :many
SELECT DISTINCT kota FROM terminal ORDER BY kota ASC;

-- =====================================================================
-- BUS
-- =====================================================================

-- name: ListBusesAdmin :many
SELECT id, kode, layanan, fasilitas, kapasitas
FROM bus
ORDER BY layanan, kode
LIMIT sqlc.arg('limit_data') OFFSET sqlc.arg('offset_data');

-- name: CountBuses :one
SELECT COUNT(id) FROM bus;

-- name: GetBusByID :one
SELECT id, kode, layanan, fasilitas, kapasitas
FROM bus
WHERE id = sqlc.arg('id')
LIMIT 1;

-- name: CreateBus :one
INSERT INTO bus (kode, layanan, fasilitas, kapasitas)
VALUES (
    sqlc.arg('kode'), sqlc.arg('layanan'), sqlc.arg('fasilitas'), sqlc.narg('kapasitas')
)
RETURNING id;

-- name: UpdateBus :exec
UPDATE bus
SET 
    kode       = sqlc.arg('kode'), 
    layanan    = sqlc.arg('layanan'), 
    fasilitas  = sqlc.arg('fasilitas'), 
    kapasitas  = sqlc.narg('kapasitas'), 
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id');

-- name: DeleteBus :exec
DELETE FROM bus WHERE id = sqlc.arg('id');

-- name: GetDistinctLayananBus :many
SELECT DISTINCT layanan FROM bus ORDER BY layanan ASC;

-- =====================================================================
-- RUTE
-- =====================================================================

-- name: ListRutesAdmin :many
SELECT r.id, r.slug, r.durasi_menit, r.aktif,
       ta.nama AS terminal_asal_nama,   ta.kota AS kota_asal,
       tt.nama AS terminal_tujuan_nama, tt.kota AS kota_tujuan
FROM rute r
JOIN terminal ta ON ta.id = r.terminal_asal_id
JOIN terminal tt ON tt.id = r.terminal_tujuan_id
ORDER BY ta.kota, ta.nama
LIMIT sqlc.arg('limit_data') OFFSET sqlc.arg('offset_data');

-- name: CountRutes :one
SELECT COUNT(id) FROM rute;

-- name: GetRuteBySlug :one
SELECT r.id, r.slug, r.terminal_asal_id, r.terminal_tujuan_id, r.durasi_menit, r.aktif,
       ta.nama AS terminal_asal_nama,  ta.kota AS kota_asal,  ta.slug AS terminal_asal_slug,
       tt.nama AS terminal_tujuan_nama, tt.kota AS kota_tujuan, tt.slug AS terminal_tujuan_slug
FROM rute r
JOIN terminal ta ON ta.id = r.terminal_asal_id
JOIN terminal tt ON tt.id = r.terminal_tujuan_id
WHERE r.slug = sqlc.arg('slug')
LIMIT 1;

-- name: GetRuteByID :one
SELECT r.id, r.slug, r.terminal_asal_id, r.terminal_tujuan_id, r.durasi_menit, r.aktif,
       ta.nama AS terminal_asal_nama,  ta.kota AS kota_asal,
       tt.nama AS terminal_tujuan_nama, tt.kota AS kota_tujuan
FROM rute r
JOIN terminal ta ON ta.id = r.terminal_asal_id
JOIN terminal tt ON tt.id = r.terminal_tujuan_id
WHERE r.id = sqlc.arg('id')
LIMIT 1;

-- name: CreateRute :one
INSERT INTO rute (terminal_asal_id, terminal_tujuan_id, slug, durasi_menit)
VALUES (
    sqlc.arg('terminal_asal_id'), sqlc.arg('terminal_tujuan_id'), 
    sqlc.arg('slug'), sqlc.arg('durasi_menit')
)
RETURNING id, slug;

-- name: UpdateRute :exec
UPDATE rute
SET 
    terminal_asal_id   = sqlc.arg('terminal_asal_id'), 
    terminal_tujuan_id = sqlc.arg('terminal_tujuan_id'), 
    slug               = sqlc.arg('slug_baru'),
    durasi_menit       = sqlc.arg('durasi_menit'), 
    aktif              = sqlc.arg('aktif'), 
    updated_at         = CURRENT_TIMESTAMP
WHERE slug = sqlc.arg('slug_lama');

-- name: DeactivateRute :exec
UPDATE rute 
SET aktif = false, updated_at = CURRENT_TIMESTAMP 
WHERE slug = sqlc.arg('slug');

-- =====================================================================
-- RUTE STOP
-- =====================================================================

-- name: GetRuteStops :many
SELECT rs.urutan, t.id AS terminal_id, t.nama, t.kota
FROM rute_stop rs
JOIN terminal t ON t.id = rs.terminal_id
WHERE rs.rute_id = sqlc.arg('rute_id')
ORDER BY rs.urutan;

-- name: DeleteRuteStops :exec
DELETE FROM rute_stop WHERE rute_id = sqlc.arg('rute_id');

-- name: InsertRuteStop :exec
INSERT INTO rute_stop (rute_id, terminal_id, urutan) 
VALUES (sqlc.arg('rute_id'), sqlc.arg('terminal_id'), sqlc.arg('urutan'));

-- =====================================================================
-- JADWAL
-- =====================================================================

-- name: SearchJadwalPublic :many
SELECT
    j.id,
    j.jam_berangkat,
    j.jam_tiba,
    b.kode      AS bus_kode,
    b.layanan   AS bus_layanan,
    b.fasilitas AS bus_fasilitas,
    r.id        AS rute_id,
    r.durasi_menit,
    ta.nama     AS terminal_asal_nama,
    tt.nama     AS terminal_tujuan_nama,
    COALESCE(h.harga, 0) AS harga_dasar
FROM jadwal j
JOIN bus b       ON b.id  = j.bus_id
JOIN rute r      ON r.id  = j.rute_id
JOIN terminal ta ON ta.id = r.terminal_asal_id
JOIN terminal tt ON tt.id = r.terminal_tujuan_id
LEFT JOIN harga h ON h.rute_id = r.id
    AND h.layanan = b.layanan
    AND (
        (b.layanan = 'luxury'  AND h.tipe_penumpang IS NULL)
     OR (b.layanan = 'reguler' AND h.tipe_penumpang = 'umum')
    )
WHERE r.terminal_asal_id  = sqlc.arg('terminal_asal_id')
  AND r.terminal_tujuan_id = sqlc.arg('terminal_tujuan_id')
  AND r.aktif = true
  AND j.aktif = true
  AND (j.hari_operasi & sqlc.arg('hari_bitmask')::SMALLINT) > 0
  AND NOT EXISTS (
        SELECT 1 FROM jadwal_libur jl
        WHERE jl.tanggal = sqlc.arg('tanggal')
          AND (jl.jadwal_id IS NULL OR jl.jadwal_id = j.id)
      )
ORDER BY j.jam_berangkat;

-- name: GetJadwalByID :one
SELECT
    j.id, j.jam_berangkat, j.jam_tiba, j.hari_operasi, j.aktif,
    b.kode      AS bus_kode,
    b.layanan   AS bus_layanan,
    b.fasilitas AS bus_fasilitas,
    r.id   AS rute_id,
    r.slug AS rute_slug,
    r.durasi_menit,
    ta.nama AS terminal_asal_nama,
    tt.nama AS terminal_tujuan_nama
FROM jadwal j
JOIN bus b       ON b.id  = j.bus_id
JOIN rute r      ON r.id  = j.rute_id
JOIN terminal ta ON ta.id = r.terminal_asal_id
JOIN terminal tt ON tt.id = r.terminal_tujuan_id
WHERE j.id = sqlc.arg('id')
LIMIT 1;

-- name: ListJadwalAdmin :many
SELECT
    j.id, j.jam_berangkat, j.jam_tiba, j.hari_operasi, j.aktif,
    b.kode    AS bus_kode,
    b.layanan AS bus_layanan,
    ta.nama   AS terminal_asal_nama,
    tt.nama   AS terminal_tujuan_nama,
    r.slug    AS rute_slug
FROM jadwal j
JOIN bus b       ON b.id  = j.bus_id
JOIN rute r      ON r.id  = j.rute_id
JOIN terminal ta ON ta.id = r.terminal_asal_id
JOIN terminal tt ON tt.id = r.terminal_tujuan_id
ORDER BY j.id DESC
LIMIT sqlc.arg('limit_data') OFFSET sqlc.arg('offset_data');

-- name: CountJadwal :one
SELECT COUNT(id) FROM jadwal;

-- name: CreateJadwal :one
INSERT INTO jadwal (rute_id, bus_id, jam_berangkat, jam_tiba, hari_operasi)
VALUES (
    sqlc.arg('rute_id'), sqlc.arg('bus_id'), 
    sqlc.arg('jam_berangkat'), sqlc.arg('jam_tiba'), sqlc.arg('hari_operasi')
)
RETURNING id;

-- name: UpdateJadwal :exec
UPDATE jadwal
SET 
    rute_id       = sqlc.arg('rute_id'), 
    bus_id        = sqlc.arg('bus_id'), 
    jam_berangkat = sqlc.arg('jam_berangkat'), 
    jam_tiba      = sqlc.arg('jam_tiba'), 
    hari_operasi  = sqlc.arg('hari_operasi'), 
    aktif         = sqlc.arg('aktif'), 
    updated_at    = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id');

-- name: DeleteJadwal :exec
DELETE FROM jadwal WHERE id = sqlc.arg('id');

-- =====================================================================
-- JADWAL LIBUR
-- =====================================================================

-- name: ListJadwalLibur :many
SELECT id, jadwal_id, tanggal, keterangan
FROM jadwal_libur
ORDER BY tanggal DESC;

-- name: CreateJadwalLibur :one
INSERT INTO jadwal_libur (jadwal_id, tanggal, keterangan)
VALUES (sqlc.narg('jadwal_id'), sqlc.arg('tanggal'), sqlc.narg('keterangan'))
RETURNING id;

-- name: DeleteJadwalLibur :exec
DELETE FROM jadwal_libur WHERE id = sqlc.arg('id');

-- =====================================================================
-- HARGA
-- =====================================================================

-- name: GetHargaByRute :many
SELECT id, layanan, tipe_penumpang, harga
FROM harga
WHERE rute_id = sqlc.arg('rute_id')
ORDER BY layanan, tipe_penumpang;

-- name: GetAllHargaInfo :many
SELECT h.id, h.rute_id, h.layanan, h.tipe_penumpang, h.harga,
       ta.nama AS terminal_asal_nama,  ta.kota AS kota_asal,
       tt.nama AS terminal_tujuan_nama, tt.kota AS kota_tujuan
FROM harga h
JOIN rute r      ON r.id  = h.rute_id
JOIN terminal ta ON ta.id = r.terminal_asal_id
JOIN terminal tt ON tt.id = r.terminal_tujuan_id
WHERE r.aktif = true
ORDER BY h.layanan, r.id, h.tipe_penumpang;

-- name: ListHargaAdmin :many
SELECT h.id, h.rute_id, h.layanan, h.tipe_penumpang, h.harga,
       ta.nama AS terminal_asal_nama,
       tt.nama AS terminal_tujuan_nama
FROM harga h
JOIN rute r      ON r.id  = h.rute_id
JOIN terminal ta ON ta.id = r.terminal_asal_id
JOIN terminal tt ON tt.id = r.terminal_tujuan_id
ORDER BY r.id, h.layanan, h.tipe_penumpang
LIMIT sqlc.arg('limit_data') OFFSET sqlc.arg('offset_data');

-- name: CountHarga :one
SELECT COUNT(id) FROM harga;

-- name: CreateHarga :one
INSERT INTO harga (rute_id, layanan, tipe_penumpang, harga)
VALUES (
    sqlc.arg('rute_id'), sqlc.arg('layanan'), 
    sqlc.narg('tipe_penumpang')::tipe_penumpang_type, sqlc.arg('harga')
)
RETURNING id;

-- name: UpdateHarga :exec
UPDATE harga 
SET harga = sqlc.arg('harga'), updated_at = CURRENT_TIMESTAMP 
WHERE id = sqlc.arg('id');

-- name: DeleteHarga :exec
DELETE FROM harga WHERE id = sqlc.arg('id');