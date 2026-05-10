-- =====================================================================
-- PENERIMA — PUBLIC
-- =====================================================================

-- name: GetPenerimaByNIK :one
-- Dipakai: validasi NIK dan tampilkan kartu identitas di Info Bansos
SELECT id, nik, nama, alamat
FROM penerima
WHERE nik = sqlc.arg('nik')
LIMIT 1;

-- =====================================================================
-- PENYALURAN — PUBLIC
-- =====================================================================

-- name: GetPenyaluranByPenerima :many
-- Dipakai: daftar bansos di halaman Info Bansos (list setelah lookup NIK)
SELECT
    py.id,
    py.nominal,
    py.periode_mulai,
    py.periode_selesai,
    py.metode,
    py.status,
    pb.nama AS program_nama,
    pb.kode AS program_kode
FROM penyaluran py
JOIN program_bansos pb ON pb.id = py.program_id
WHERE py.penerima_id = sqlc.arg('penerima_id')
ORDER BY py.periode_selesai DESC;

-- name: GetDetailPenyaluran :one
-- Dipakai: halaman Detail Bansos (nominal, periode, metode, status, deskripsi program)
SELECT
    py.id,
    py.nominal,
    py.periode_mulai,
    py.periode_selesai,
    py.metode,
    py.status,
    pb.nama      AS program_nama,
    pb.kode      AS program_kode,
    pb.deskripsi AS program_deskripsi
FROM penyaluran py
JOIN program_bansos pb ON pb.id = py.program_id
WHERE py.id = sqlc.arg('id')
LIMIT 1;

-- =====================================================================
-- PROGRAM BANSOS — ADMIN
-- =====================================================================

-- name: ListProgramAdmin :many
SELECT id, nama, kode, deskripsi, aktif
FROM program_bansos
ORDER BY nama
LIMIT sqlc.arg('limit_data') OFFSET sqlc.arg('offset_data');

-- name: CountProgram :one
SELECT COUNT(id) FROM program_bansos;

-- name: GetProgramByID :one
SELECT id, nama, kode, deskripsi, aktif
FROM program_bansos
WHERE id = sqlc.arg('id')
LIMIT 1;

-- name: CreateProgram :one
INSERT INTO program_bansos (nama, kode, deskripsi)
VALUES (sqlc.arg('nama'), sqlc.arg('kode'), sqlc.narg('deskripsi'))
RETURNING id;

-- name: UpdateProgram :exec
UPDATE program_bansos
SET
    nama      = sqlc.arg('nama'),
    kode      = sqlc.arg('kode'),
    deskripsi = sqlc.narg('deskripsi'),
    aktif     = sqlc.arg('aktif'),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id');

-- name: DeleteProgram :exec
DELETE FROM program_bansos WHERE id = sqlc.arg('id');

-- =====================================================================
-- PENERIMA — ADMIN
-- =====================================================================

-- name: ListPenerimaAdmin :many
SELECT id, nik, nama, alamat
FROM penerima
WHERE (sqlc.narg('keyword')::text IS NULL
       OR nama ILIKE '%' || sqlc.narg('keyword')::text || '%'
       OR nik  ILIKE '%' || sqlc.narg('keyword')::text || '%')
ORDER BY nama
LIMIT sqlc.arg('limit_data') OFFSET sqlc.arg('offset_data');

-- name: CountPenerima :one
SELECT COUNT(id)
FROM penerima
WHERE (sqlc.narg('keyword')::text IS NULL
       OR nama ILIKE '%' || sqlc.narg('keyword')::text || '%'
       OR nik  ILIKE '%' || sqlc.narg('keyword')::text || '%');

-- name: GetPenerimaByID :one
SELECT id, nik, nama, alamat
FROM penerima
WHERE id = sqlc.arg('id')
LIMIT 1;

-- name: CreatePenerima :one
INSERT INTO penerima (nik, nama, alamat)
VALUES (sqlc.arg('nik'), sqlc.arg('nama'), sqlc.arg('alamat'))
RETURNING id;

-- name: UpdatePenerima :exec
UPDATE penerima
SET nik = sqlc.arg('nik'), nama = sqlc.arg('nama'),
    alamat = sqlc.arg('alamat'), updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id');

-- name: DeletePenerima :exec
DELETE FROM penerima WHERE id = sqlc.arg('id');

-- =====================================================================
-- PENYALURAN — ADMIN
-- =====================================================================

-- name: ListPenyaluranAdmin :many
SELECT
    py.id, py.nominal, py.periode_mulai, py.periode_selesai,
    py.metode, py.status, py.created_at,
    pn.nik   AS penerima_nik,
    pn.nama  AS penerima_nama,
    pb.nama  AS program_nama,
    pb.kode  AS program_kode
FROM penyaluran py
JOIN penerima      pn ON pn.id = py.penerima_id
JOIN program_bansos pb ON pb.id = py.program_id
WHERE (sqlc.narg('program_id')::int IS NULL OR py.program_id = sqlc.narg('program_id')::int)
  AND (sqlc.narg('status')::status_penyaluran_type IS NULL OR py.status = sqlc.narg('status')::status_penyaluran_type)
ORDER BY py.created_at DESC
LIMIT sqlc.arg('limit_data') OFFSET sqlc.arg('offset_data');

-- name: CountPenyaluran :one
SELECT COUNT(py.id)
FROM penyaluran py
WHERE (sqlc.narg('program_id')::int IS NULL OR py.program_id = sqlc.narg('program_id')::int)
  AND (sqlc.narg('status')::status_penyaluran_type IS NULL OR py.status = sqlc.narg('status')::status_penyaluran_type);

-- name: CreatePenyaluran :one
INSERT INTO penyaluran (penerima_id, program_id, nominal, periode_mulai, periode_selesai, metode, status)
VALUES (
    sqlc.arg('penerima_id'), sqlc.arg('program_id'), sqlc.arg('nominal'),
    sqlc.arg('periode_mulai'), sqlc.arg('periode_selesai'),
    sqlc.arg('metode'), sqlc.arg('status')
)
RETURNING id;

-- name: UpdateStatusPenyaluran :exec
UPDATE penyaluran
SET status = sqlc.arg('status'), updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id');

-- name: UpdatePenyaluran :exec
UPDATE penyaluran
SET
    program_id      = sqlc.arg('program_id'),
    nominal         = sqlc.arg('nominal'),
    periode_mulai   = sqlc.arg('periode_mulai'),
    periode_selesai = sqlc.arg('periode_selesai'),
    metode          = sqlc.arg('metode'),
    status          = sqlc.arg('status'),
    updated_at      = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id');

-- name: DeletePenyaluran :exec
DELETE FROM penyaluran WHERE id = sqlc.arg('id');
