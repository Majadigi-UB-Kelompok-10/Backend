
INSERT INTO program_bansos (nama, kode, deskripsi, aktif) VALUES
(
    'Bantuan PKH', 
    'PKH-2026', 
    'program jaminan sosial (Jaminan Kehilangan Pekerjaan/JKP) dari BPJS Ketenagakerjaan yang diberikan kepada pekerja korban PHK berupa uang tunai (selama 6 bulan), pelatihan kerja, dan akses informasi pasar kerja.', 
    true
),
(
    'BPNT', 
    'BPNT-2026', 
    'Bantuan Pangan Non Tunai (Sembako) untuk membantu memenuhi kebutuhan pokok harian warga yang terdaftar dalam DTKS.', 
    true
)
ON CONFLICT (kode) DO NOTHING;

-- =====================================================================
-- 3. INSERT PENERIMA (Profil Warga)
-- =====================================================================
INSERT INTO penerima (nik, nama, alamat) VALUES
(
    '1234567890123456', 
    'Budi Santoso', 
    'Kota Malang'
)
ON CONFLICT (nik) DO NOTHING;

-- =====================================================================
-- 4. INSERT PENYALURAN (Riwayat Bantuan Warga)
-- =====================================================================
-- Kita gunakan subquery untuk mencari ID secara otomatis agar aman.

-- Riwayat 1: Bantuan PKH (DITERIMA, Jan - Mar 2026)
INSERT INTO penyaluran (penerima_id, program_id, nominal, periode_mulai, periode_selesai, metode, status)
VALUES (
    (SELECT id FROM penerima WHERE nik = '1234567890123456'),
    (SELECT id FROM program_bansos WHERE kode = 'PKH-2026'),
    750000,
    '2026-01-01',
    '2026-03-31',
    'transfer_bri',
    'diterima'
);

-- Riwayat 2: BPNT (PROSES, April 2026)
INSERT INTO penyaluran (penerima_id, program_id, nominal, periode_mulai, periode_selesai, metode, status)
VALUES (
    (SELECT id FROM penerima WHERE nik = '1234567890123456'),
    (SELECT id FROM program_bansos WHERE kode = 'BPNT-2026'),
    500000,
    '2026-04-01',
    '2026-04-30',
    'transfer_pos', 
    'proses'
);