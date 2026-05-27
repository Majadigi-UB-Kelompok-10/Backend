-- =====================================================================
-- 002_seed.sql
-- DUMMY DATA UNTUK JDIH PROVINSI JAWA TIMUR
-- =====================================================================

-- ---------------------------------------------------------------------
-- 1. SEED PENGUMUMAN (Tampil di Carousel Homepage)
-- ---------------------------------------------------------------------
INSERT INTO pengumuman (judul, isi, aktif) VALUES
('Layanan Digital Baru Kini Tersedia Untuk Publik', 'Kini masyarakat dapat mengakses seluruh dokumen hukum Provinsi Jawa Timur dengan lebih mudah, cepat, dan transparan melalui portal JDIH versi terbaru.', true),
('Sosialisasi Peraturan Daerah Terbaru Tahun 2026', 'Pemerintah Provinsi Jawa Timur akan mengadakan sosialisasi terkait Peraturan Daerah terbaru tentang Rencana Tata Ruang Wilayah pada akhir bulan ini di Gedung Negara Grahadi.', true),
('Pembaruan Database Dokumen Hukum 2020-2024', 'Seluruh dokumen hukum Provinsi Jawa Timur dari tahun 2020 hingga 2024 telah diperbarui dan dapat diakses secara digital melalui portal ini.', true);

-- ---------------------------------------------------------------------
-- 2. SEED DOKUMEN HUKUM — Semua 8 Jenis, Multi-Tahun
-- ---------------------------------------------------------------------
INSERT INTO dokumen_hukum (
    jenis, nomor, tahun, judul, ringkasan, tanggal_penetapan,
    status, pdf_url, pdf_size_kb, urusan_pemerintahan, jumlah_view
) VALUES
-- ==================== PERDA ====================
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
    'perda', '3', 2021,
    'Peraturan Daerah tentang Rencana Tata Ruang Wilayah Provinsi Jawa Timur Tahun 2021-2041',
    'Menetapkan rencana tata ruang wilayah sebagai acuan pembangunan dan pemanfaatan ruang di Jawa Timur selama dua dekade ke depan.',
    '2021-05-20', 'berlaku',
    'https://res.cloudinary.com/demo/raw/upload/v1/jdih/dummy_perda_3_2021.pdf',
    4200, 'Tata Ruang', 2100
),
(
    'perda', '2', 2020,
    'Peraturan Daerah tentang Pengelolaan Lingkungan Hidup dan Konservasi Alam Jawa Timur',
    'Menetapkan standar pengelolaan lingkungan hidup serta perlindungan kawasan konservasi alam di seluruh wilayah Jawa Timur.',
    '2020-07-14', 'berlaku',
    'https://res.cloudinary.com/demo/raw/upload/v1/jdih/dummy_perda_2_2020.pdf',
    3100, 'Lingkungan Hidup', 987
),

-- ==================== PERGUB ====================
(
    'pergub', '32', 2024,
    'Peraturan Gubernur tentang Pengadaan Barang dan Jasa Pemerintah Provinsi Jawa Timur',
    'Menetapkan tata cara dan mekanisme pengadaan barang dan jasa di lingkungan Pemerintah Provinsi Jawa Timur.',
    '2024-02-10', 'berlaku',
    'https://res.cloudinary.com/demo/raw/upload/v1/jdih/dummy_pergub_32_2024.pdf',
    1800, 'Pengadaan Barang dan Jasa', 763
),
(
    'pergub', '15', 2022,
    'Peraturan Gubernur tentang Percepatan Transformasi Digital Pelayanan Publik Jawa Timur',
    'Mendorong seluruh OPD untuk menerapkan sistem pelayanan digital terintegrasi guna meningkatkan kualitas layanan kepada masyarakat.',
    '2022-04-05', 'berlaku',
    'https://res.cloudinary.com/demo/raw/upload/v1/jdih/dummy_pergub_15_2022.pdf',
    2200, 'Komunikasi dan Informatika', 1450
),
(
    'pergub', '7', 2021,
    'Peraturan Gubernur tentang Standar Pelayanan Minimal Bidang Kesehatan Provinsi Jawa Timur',
    'Menetapkan standar minimal pelayanan kesehatan yang wajib dipenuhi oleh seluruh fasilitas kesehatan di Jawa Timur.',
    '2021-03-18', 'berlaku',
    'https://res.cloudinary.com/demo/raw/upload/v1/jdih/dummy_pergub_7_2021.pdf',
    1950, 'Kesehatan', 890
),

-- ==================== PERATURAN ====================
(
    'peraturan', '10', 2025,
    'Peraturan tentang Tata Kelola Data dan Informasi Pemerintah Provinsi Jawa Timur',
    'Mengatur klasifikasi, pengelolaan, dan keamanan data serta informasi milik Pemerintah Provinsi Jawa Timur.',
    '2025-01-20', 'berlaku',
    'https://res.cloudinary.com/demo/raw/upload/v1/jdih/dummy_peraturan_10_2025.pdf',
    1600, 'Komunikasi dan Informatika', 340
),
(
    'peraturan', '4', 2022,
    'Peraturan tentang Pengelolaan dan Pelestarian Arsip Daerah Provinsi Jawa Timur',
    'Menetapkan prosedur pengelolaan arsip daerah dari penciptaan, penyimpanan, hingga pemusnahan dokumen sesuai ketentuan nasional.',
    '2022-09-12', 'berlaku',
    'https://res.cloudinary.com/demo/raw/upload/v1/jdih/dummy_peraturan_4_2022.pdf',
    1300, 'Kearsipan', 215
),

-- ==================== PERDES ====================
(
    'perdes', '2', 2024,
    'Peraturan Desa Wonorejo Kecamatan Rungkut tentang Pengelolaan Dana Desa Tahun Anggaran 2024',
    'Mengatur prioritas penggunaan dan pertanggungjawaban dana desa untuk pembangunan infrastruktur dan pemberdayaan masyarakat.',
    '2024-01-08', 'berlaku',
    'https://res.cloudinary.com/demo/raw/upload/v1/jdih/dummy_perdes_2_2024.pdf',
    980, 'Pemerintahan Desa', 127
),
(
    'perdes', '1', 2023,
    'Peraturan Desa Sidokumpul Kabupaten Gresik tentang Badan Usaha Milik Desa',
    'Menetapkan pembentukan, struktur organisasi, dan mekanisme pengelolaan Badan Usaha Milik Desa Sidokumpul.',
    '2023-02-28', 'berlaku',
    'https://res.cloudinary.com/demo/raw/upload/v1/jdih/dummy_perdes_1_2023.pdf',
    870, 'Pemerintahan Desa', 94
),

-- ==================== SK_GUB ====================
(
    'sk_gub', '188/312/KPTS/013/2025', 2025,
    'Surat Keputusan Gubernur tentang Pembentukan Tim Percepatan Investasi Jawa Timur 2025',
    'Membentuk tim lintas OPD untuk mempercepat realisasi investasi dan kemudahan berusaha di Provinsi Jawa Timur.',
    '2025-03-01', 'berlaku',
    'https://res.cloudinary.com/demo/raw/upload/v1/jdih/dummy_skgub_188_2025.pdf',
    1100, 'Penanaman Modal', 560
),
(
    'sk_gub', '188/145/KPTS/013/2023', 2023,
    'Surat Keputusan Gubernur tentang Penetapan Upah Minimum Provinsi Jawa Timur Tahun 2024',
    'Menetapkan besaran Upah Minimum Provinsi (UMP) Jawa Timur yang berlaku mulai 1 Januari 2024.',
    '2023-11-21', 'berlaku',
    'https://res.cloudinary.com/demo/raw/upload/v1/jdih/dummy_skgub_145_2023.pdf',
    1050, 'Ketenagakerjaan', 3200
),

-- ==================== INSTRUKSI ====================
(
    'instruksi', '1', 2026,
    'Instruksi Gubernur tentang Percepatan Penyaluran Bantuan Sosial Kepada Masyarakat Terdampak',
    'Menginstruksikan seluruh kepala OPD dan bupati/walikota untuk mempercepat penyaluran bantuan sosial kepada masyarakat terdampak.',
    '2026-02-14', 'berlaku',
    'https://res.cloudinary.com/demo/raw/upload/v1/jdih/dummy_instruksi_1_2026.pdf',
    920, 'Sosial', 1840
),
(
    'instruksi', '3', 2024,
    'Instruksi Gubernur tentang Penghematan Energi dan Air di Lingkungan Pemerintah Provinsi Jawa Timur',
    'Menginstruksikan seluruh OPD untuk menerapkan program penghematan energi listrik dan air bersih secara terstruktur.',
    '2024-06-05', 'berlaku',
    'https://res.cloudinary.com/demo/raw/upload/v1/jdih/dummy_instruksi_3_2024.pdf',
    780, 'Energi dan Sumber Daya', 430
),

-- ==================== SE ====================
(
    'se', '800/76/200.1.1/2026', 2026,
    'Surat Edaran tentang Pengibaran Bendera Merah Putih Setengah Tiang',
    'Instruksi kepada seluruh instansi pemerintah untuk mengibarkan bendera setengah tiang sebagai tanda hari berkabung nasional.',
    '2026-01-01', 'berlaku',
    'https://res.cloudinary.com/demo/raw/upload/v1/jdih/dummy_se_bendera_2026.pdf',
    1200, 'Kepegawaian dan Kesekretariatan Daerah', 1500
),
(
    'se', '420/2580/101.1/2025', 2025,
    'Surat Edaran tentang Penerapan Kurikulum Merdeka di Satuan Pendidikan Provinsi Jawa Timur',
    'Menginstruksikan kepala sekolah dan dinas pendidikan kabupaten/kota untuk segera mengimplementasikan Kurikulum Merdeka.',
    '2025-07-10', 'berlaku',
    'https://res.cloudinary.com/demo/raw/upload/v1/jdih/dummy_se_420_2025.pdf',
    1050, 'Pendidikan', 2760
),
(
    'se', '440/3120/102.2/2022', 2022,
    'Surat Edaran Protokol Kesehatan dan Penanganan Pasca Pandemi di Lingkungan Pemerintahan',
    'Menetapkan standar protokol kesehatan yang wajib diterapkan seluruh ASN di lingkungan pemerintahan pasca pandemi.',
    '2022-05-17', 'berlaku',
    'https://res.cloudinary.com/demo/raw/upload/v1/jdih/dummy_se_440_2022.pdf',
    900, 'Kesehatan', 1100
),

-- ==================== KEPUTUSAN ====================
(
    'keputusan', '050/3456/013.2/2025', 2025,
    'Keputusan tentang Penetapan Rencana Aksi Daerah Tujuan Pembangunan Berkelanjutan Jawa Timur 2025-2030',
    'Menetapkan rencana aksi daerah dalam rangka pencapaian Tujuan Pembangunan Berkelanjutan (TPB/SDGs) di Provinsi Jawa Timur.',
    '2025-04-22', 'berlaku',
    'https://res.cloudinary.com/demo/raw/upload/v1/jdih/dummy_keputusan_050_2025.pdf',
    2400, 'Perencanaan Pembangunan', 610
),
(
    'keputusan', '180/234/013/2023', 2023,
    'Keputusan tentang Pengesahan Perubahan Anggaran Pendapatan dan Belanja Daerah Tahun 2023',
    'Mengesahkan perubahan APBD Provinsi Jawa Timur Tahun Anggaran 2023 sebagaimana hasil evaluasi Kementerian Dalam Negeri.',
    '2023-08-30', 'berlaku',
    'https://res.cloudinary.com/demo/raw/upload/v1/jdih/dummy_keputusan_180_2023.pdf',
    3100, 'Keuangan Daerah', 445
)
ON CONFLICT (jenis, nomor, tahun) DO NOTHING;

-- ---------------------------------------------------------------------
-- 3. SEED SUBJEK (Tag Dokumen)
-- ---------------------------------------------------------------------
INSERT INTO subjek (nama) VALUES
('Keuangan Daerah'),
('Ketentraman Umum'),
('Pajak Daerah'),
('Bendera Negara'),
('Hari Berkabung'),
('Tata Ruang'),
('Lingkungan Hidup'),
('Pengadaan Barang'),
('Transformasi Digital'),
('Kesehatan'),
('Dana Desa'),
('Investasi'),
('Upah Minimum'),
('Bantuan Sosial'),
('Kurikulum Merdeka'),
('Pembangunan Berkelanjutan'),
('Anggaran Daerah')
ON CONFLICT (nama) DO NOTHING;

-- ---------------------------------------------------------------------
-- 4. SEED DOKUMEN_SUBJEK (Relasi Many-to-Many)
-- ---------------------------------------------------------------------
INSERT INTO dokumen_subjek (dokumen_id, subjek_id)
SELECT d.id, s.id FROM dokumen_hukum d, subjek s
WHERE (d.nomor = '1'  AND d.tahun = 2023 AND d.jenis = 'perda'     AND s.nama = 'Keuangan Daerah')
   OR (d.nomor = '5'  AND d.tahun = 2023 AND d.jenis = 'perda'     AND s.nama = 'Ketentraman Umum')
   OR (d.nomor = '8'  AND d.tahun = 2023 AND d.jenis = 'perda'     AND s.nama = 'Pajak Daerah')
   OR (d.nomor = '800/76/200.1.1/2026' AND d.tahun = 2026 AND d.jenis = 'se' AND s.nama = 'Bendera Negara')
   OR (d.nomor = '800/76/200.1.1/2026' AND d.tahun = 2026 AND d.jenis = 'se' AND s.nama = 'Hari Berkabung')
   OR (d.nomor = '3'  AND d.tahun = 2021 AND d.jenis = 'perda'     AND s.nama = 'Tata Ruang')
   OR (d.nomor = '2'  AND d.tahun = 2020 AND d.jenis = 'perda'     AND s.nama = 'Lingkungan Hidup')
   OR (d.nomor = '32' AND d.tahun = 2024 AND d.jenis = 'pergub'    AND s.nama = 'Pengadaan Barang')
   OR (d.nomor = '15' AND d.tahun = 2022 AND d.jenis = 'pergub'    AND s.nama = 'Transformasi Digital')
   OR (d.nomor = '7'  AND d.tahun = 2021 AND d.jenis = 'pergub'    AND s.nama = 'Kesehatan')
   OR (d.nomor = '2'  AND d.tahun = 2024 AND d.jenis = 'perdes'    AND s.nama = 'Dana Desa')
   OR (d.nomor = '188/312/KPTS/013/2025' AND d.tahun = 2025 AND d.jenis = 'sk_gub' AND s.nama = 'Investasi')
   OR (d.nomor = '188/145/KPTS/013/2023' AND d.tahun = 2023 AND d.jenis = 'sk_gub' AND s.nama = 'Upah Minimum')
   OR (d.nomor = '1'  AND d.tahun = 2026 AND d.jenis = 'instruksi' AND s.nama = 'Bantuan Sosial')
   OR (d.nomor = '420/2580/101.1/2025' AND d.tahun = 2025 AND d.jenis = 'se' AND s.nama = 'Kurikulum Merdeka')
   OR (d.nomor = '050/3456/013.2/2025' AND d.tahun = 2025 AND d.jenis = 'keputusan' AND s.nama = 'Pembangunan Berkelanjutan')
   OR (d.nomor = '180/234/013/2023' AND d.tahun = 2023 AND d.jenis = 'keputusan' AND s.nama = 'Anggaran Daerah')
ON CONFLICT DO NOTHING;
