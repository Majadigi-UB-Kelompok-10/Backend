-- =====================================================================
-- 002_seed.sql
-- DUMMY DATA UNTUK JDIH PROVINSI JAWA TIMUR
-- =====================================================================

-- ---------------------------------------------------------------------
-- 1. SEED PENGUMUMAN (Tampil di Carousel Homepage)
-- ---------------------------------------------------------------------
INSERT INTO pengumuman (judul, isi, aktif) VALUES
('Layanan Digital Baru Kini Tersedia Untuk Publik', 'Kini masyarakat dapat mengakses seluruh dokumen hukum Provinsi Jawa Timur dengan lebih mudah, cepat, dan transparan melalui portal JDIH versi terbaru.', true),
('Sosialisasi Peraturan Daerah Terbaru Tahun 2026', 'Pemerintah Provinsi Jawa Timur akan mengadakan sosialisasi terkait Peraturan Daerah terbaru tentang Rencana Tata Ruang Wilayah pada akhir bulan ini di Gedung Negara Grahadi.', true);

-- ---------------------------------------------------------------------
-- 2. SEED DOKUMEN HUKUM (Diambil dari desain UI)
-- ---------------------------------------------------------------------
INSERT INTO dokumen_hukum (
    jenis, nomor, tahun, judul, ringkasan, tanggal_penetapan, 
    status, pdf_url, pdf_size_kb, urusan_pemerintahan, jumlah_view
) VALUES
(
    'perda', '1', 2023, 
    'Peraturan Daerah tentang Pengelolaan Keuangan Daerah Provinsi Jawa Timur Tahun 2023', 
    'Mengatur pedoman pengelolaan keuangan daerah yang transparan dan akuntabel sesuai dengan standar akuntansi pemerintahan.', 
    '2023-01-15', 'berlaku', 
    'https://res.cloudinary.com/demo/raw/upload/v1/jdih/dummy_perda_1_2023.pdf', 
    1250, 'Keuangan Daerah', 1200
),
(
    'perda', '5', 2023, 
    'Penyelenggaraan Ketentraman, Ketertiban Umum, dan Pelindungan Masyarakat di Provinsi Jawa Timur', 
    'Pedoman untuk Satuan Polisi Pamong Praja dan instansi terkait dalam menjaga ketertiban umum di wilayah Jawa Timur.', 
    '2023-03-22', 'berlaku', 
    'https://res.cloudinary.com/demo/raw/upload/v1/jdih/dummy_perda_5_2023.pdf', 
    2100, 'Ketentraman dan Ketertiban', 856
),
(
    'perda', '8', 2023, 
    'Pajak Daerah dan Retribusi Daerah Sebagai Upaya Optimalisasi Pendapatan Asli Daerah', 
    'Penyesuaian tarif pajak dan retribusi daerah di seluruh wilayah administratif Jawa Timur.', 
    '2023-06-10', 'diubah', 
    'https://res.cloudinary.com/demo/raw/upload/v1/jdih/dummy_perda_8_2023.pdf', 
    3400, 'Pendapatan Daerah', 432
),
(
    'se', '800/76/200.1.1/2026', 2026, 
    'Pengibaran Bendera Merah Putih Setengah Tiang', 
    'Instruksi kepada seluruh instansi pemerintah untuk mengibarkan bendera setengah tiang sebagai tanda hari berkabung nasional.', 
    '2026-01-01', 'berlaku', 
    'https://res.cloudinary.com/demo/raw/upload/v1/jdih/dummy_se_bendera_2026.pdf', 
    1200, 'Kepegawaian dan Kesekretariatan Daerah', 1500
);

-- ---------------------------------------------------------------------
-- 3. SEED SUBJEK (Tag Dokumen)
-- ---------------------------------------------------------------------
INSERT INTO subjek (nama) VALUES
('Keuangan Daerah'),
('Ketentraman Umum'),
('Pajak Daerah'),
('Bendera Negara'),
('Hari Berkabung');

-- ---------------------------------------------------------------------
-- 4. SEED DOKUMEN_SUBJEK (Relasi Many-to-Many pakai Subquery Aman)
-- ---------------------------------------------------------------------
INSERT INTO dokumen_subjek (dokumen_id, subjek_id) VALUES
-- Relasi untuk Perda No 1 / 2023
((SELECT id FROM dokumen_hukum WHERE nomor = '1' AND tahun = 2023), (SELECT id FROM subjek WHERE nama = 'Keuangan Daerah')),

-- Relasi untuk Perda No 5 / 2023
((SELECT id FROM dokumen_hukum WHERE nomor = '5' AND tahun = 2023), (SELECT id FROM subjek WHERE nama = 'Ketentraman Umum')),

-- Relasi untuk Perda No 8 / 2023
((SELECT id FROM dokumen_hukum WHERE nomor = '8' AND tahun = 2023), (SELECT id FROM subjek WHERE nama = 'Pajak Daerah')),

-- Relasi untuk SE Bendera (2 subjek sekaligus)
((SELECT id FROM dokumen_hukum WHERE nomor = '800/76/200.1.1/2026' AND tahun = 2026), (SELECT id FROM subjek WHERE nama = 'Bendera Negara')),
((SELECT id FROM dokumen_hukum WHERE nomor = '800/76/200.1.1/2026' AND tahun = 2026), (SELECT id FROM subjek WHERE nama = 'Hari Berkabung'));