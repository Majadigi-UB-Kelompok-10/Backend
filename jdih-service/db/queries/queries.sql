-- =====================================================================
-- DOKUMEN HUKUM — PUBLIC
-- =====================================================================

-- name: SearchDokumenTerbaru :many
-- Dipakai: halaman Hasil Pencarian, sort Terbaru
SELECT id, jenis, nomor, tahun, judul, ringkasan, tanggal_penetapan, status, jumlah_view
FROM dokumen_hukum
WHERE (sqlc.narg('keyword')::text    IS NULL OR judul ILIKE '%' || sqlc.narg('keyword')::text || '%')
  AND (sqlc.narg('nomor')::text      IS NULL OR nomor ILIKE '%' || sqlc.narg('nomor')::text || '%')
  AND (sqlc.narg('tahun')::smallint  IS NULL OR tahun = sqlc.narg('tahun')::smallint)
  AND (sqlc.narg('jenis')::jenis_dokumen_type IS NULL OR jenis = sqlc.narg('jenis')::jenis_dokumen_type)
ORDER BY tanggal_penetapan DESC
LIMIT sqlc.arg('limit_data') OFFSET sqlc.arg('offset_data');

-- name: SearchDokumenPopuler :many
-- Dipakai: halaman Hasil Pencarian, sort Populer
SELECT id, jenis, nomor, tahun, judul, ringkasan, tanggal_penetapan, status, jumlah_view
FROM dokumen_hukum
WHERE (sqlc.narg('keyword')::text    IS NULL OR judul ILIKE '%' || sqlc.narg('keyword')::text || '%')
  AND (sqlc.narg('nomor')::text      IS NULL OR nomor ILIKE '%' || sqlc.narg('nomor')::text || '%')
  AND (sqlc.narg('tahun')::smallint  IS NULL OR tahun = sqlc.narg('tahun')::smallint)
  AND (sqlc.narg('jenis')::jenis_dokumen_type IS NULL OR jenis = sqlc.narg('jenis')::jenis_dokumen_type)
ORDER BY jumlah_view DESC, tanggal_penetapan DESC
LIMIT sqlc.arg('limit_data') OFFSET sqlc.arg('offset_data');

-- name: CountSearchDokumen :one
SELECT COUNT(id)
FROM dokumen_hukum
WHERE (sqlc.narg('keyword')::text    IS NULL OR judul ILIKE '%' || sqlc.narg('keyword')::text || '%')
  AND (sqlc.narg('nomor')::text      IS NULL OR nomor ILIKE '%' || sqlc.narg('nomor')::text || '%')
  AND (sqlc.narg('tahun')::smallint  IS NULL OR tahun = sqlc.narg('tahun')::smallint)
  AND (sqlc.narg('jenis')::jenis_dokumen_type IS NULL OR jenis = sqlc.narg('jenis')::jenis_dokumen_type);

-- name: ListDokumenByJenis :many
-- Dipakai: halaman Akses Cepat per jenis (search nomor/judul + filter tahun)
SELECT id, nomor, tahun, judul, tanggal_penetapan, status, pdf_url, pdf_size_kb
FROM dokumen_hukum
WHERE jenis = sqlc.arg('jenis')
  AND (sqlc.narg('tahun')::smallint IS NULL OR tahun = sqlc.narg('tahun')::smallint)
  AND (sqlc.narg('keyword')::text   IS NULL
       OR nomor ILIKE '%' || sqlc.narg('keyword')::text || '%'
       OR judul ILIKE '%' || sqlc.narg('keyword')::text || '%')
ORDER BY tanggal_penetapan DESC
LIMIT sqlc.arg('limit_data') OFFSET sqlc.arg('offset_data');

-- name: CountDokumenByJenis :one
SELECT COUNT(id)
FROM dokumen_hukum
WHERE jenis = sqlc.arg('jenis')
  AND (sqlc.narg('tahun')::smallint IS NULL OR tahun = sqlc.narg('tahun')::smallint)
  AND (sqlc.narg('keyword')::text   IS NULL
       OR nomor ILIKE '%' || sqlc.narg('keyword')::text || '%'
       OR judul ILIKE '%' || sqlc.narg('keyword')::text || '%');

-- name: GetDokumenByID :one
SELECT id, jenis, nomor, tahun, judul, ringkasan, tanggal_penetapan, status,
       pdf_url, pdf_size_kb, urusan_pemerintahan, jumlah_view
FROM dokumen_hukum
WHERE id = sqlc.arg('id')
LIMIT 1;

-- name: GetSubjekByDokumen :many
-- Dipakai: bagian SUBJEK di Detail Dokumen
SELECT s.id, s.nama
FROM subjek s
JOIN dokumen_subjek ds ON ds.subjek_id = s.id
WHERE ds.dokumen_id = sqlc.arg('dokumen_id')
ORDER BY s.nama;

-- name: IncrementView :exec
UPDATE dokumen_hukum
SET jumlah_view = jumlah_view + 1
WHERE id = sqlc.arg('id');

-- name: GetDistinctTahunByJenis :many
-- Dipakai: tab filter tahun di halaman Akses Cepat
SELECT DISTINCT tahun
FROM dokumen_hukum
WHERE jenis = sqlc.arg('jenis')
ORDER BY tahun DESC;

-- =====================================================================
-- PENGUMUMAN — PUBLIC
-- =====================================================================

-- name: ListPengumumanTerbaru :many
-- Dipakai: carousel TERBARU di homepage
SELECT id, judul, isi, created_at
FROM pengumuman
WHERE aktif = true
ORDER BY created_at DESC
LIMIT sqlc.arg('limit_data');

-- =====================================================================
-- DOKUMEN HUKUM — ADMIN
-- =====================================================================

-- name: ListAllDokumenAdmin :many
SELECT id, jenis, nomor, tahun, judul, tanggal_penetapan, status, jumlah_view
FROM dokumen_hukum
WHERE (sqlc.narg('jenis')::jenis_dokumen_type IS NULL OR jenis = sqlc.narg('jenis')::jenis_dokumen_type)
  AND (sqlc.narg('tahun')::smallint IS NULL OR tahun = sqlc.narg('tahun')::smallint)
ORDER BY created_at DESC
LIMIT sqlc.arg('limit_data') OFFSET sqlc.arg('offset_data');

-- name: CountAllDokumen :one
SELECT COUNT(id)
FROM dokumen_hukum
WHERE (sqlc.narg('jenis')::jenis_dokumen_type IS NULL OR jenis = sqlc.narg('jenis')::jenis_dokumen_type)
  AND (sqlc.narg('tahun')::smallint IS NULL OR tahun = sqlc.narg('tahun')::smallint);

-- name: CreateDokumen :one
INSERT INTO dokumen_hukum (
    jenis, nomor, tahun, judul, ringkasan, tanggal_penetapan,
    status, pdf_url, pdf_size_kb, urusan_pemerintahan
)
VALUES (
    sqlc.arg('jenis'), sqlc.arg('nomor'), sqlc.arg('tahun'),
    sqlc.arg('judul'), sqlc.narg('ringkasan'), sqlc.arg('tanggal_penetapan'),
    sqlc.arg('status'), sqlc.arg('pdf_url'), sqlc.narg('pdf_size_kb'),
    sqlc.narg('urusan_pemerintahan')
)
RETURNING id;

-- name: UpdateDokumen :exec
UPDATE dokumen_hukum
SET
    jenis               = sqlc.arg('jenis'),
    nomor               = sqlc.arg('nomor'),
    tahun               = sqlc.arg('tahun'),
    judul               = sqlc.arg('judul'),
    ringkasan           = sqlc.narg('ringkasan'),
    tanggal_penetapan   = sqlc.arg('tanggal_penetapan'),
    status              = sqlc.arg('status'),
    pdf_url             = sqlc.arg('pdf_url'),
    pdf_size_kb         = sqlc.narg('pdf_size_kb'),
    urusan_pemerintahan = sqlc.narg('urusan_pemerintahan'),
    updated_at          = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id');

-- name: DeleteDokumen :exec
DELETE FROM dokumen_hukum WHERE id = sqlc.arg('id');

-- =====================================================================
-- SUBJEK — ADMIN
-- =====================================================================

-- name: GetOrCreateSubjek :one
INSERT INTO subjek (nama)
VALUES (sqlc.arg('nama'))
ON CONFLICT (nama) DO UPDATE SET nama = EXCLUDED.nama
RETURNING id, nama;

-- name: AddSubjekToDokumen :exec
INSERT INTO dokumen_subjek (dokumen_id, subjek_id)
VALUES (sqlc.arg('dokumen_id'), sqlc.arg('subjek_id'))
ON CONFLICT DO NOTHING;

-- name: RemoveSubjekFromDokumen :exec
DELETE FROM dokumen_subjek
WHERE dokumen_id = sqlc.arg('dokumen_id')
  AND subjek_id  = sqlc.arg('subjek_id');

-- name: RemoveAllSubjekFromDokumen :exec
DELETE FROM dokumen_subjek WHERE dokumen_id = sqlc.arg('dokumen_id');

-- =====================================================================
-- PENGUMUMAN — ADMIN
-- =====================================================================

-- name: ListPengumumanAdmin :many
SELECT id, judul, isi, aktif, created_at
FROM pengumuman
ORDER BY created_at DESC
LIMIT sqlc.arg('limit_data') OFFSET sqlc.arg('offset_data');

-- name: CountPengumuman :one
SELECT COUNT(id) FROM pengumuman;

-- name: CreatePengumuman :one
INSERT INTO pengumuman (judul, isi)
VALUES (sqlc.arg('judul'), sqlc.narg('isi'))
RETURNING id;

-- name: UpdatePengumuman :exec
UPDATE pengumuman
SET judul = sqlc.arg('judul'), isi = sqlc.narg('isi'),
    aktif = sqlc.arg('aktif'), updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id');

-- name: DeletePengumuman :exec
DELETE FROM pengumuman WHERE id = sqlc.arg('id');
