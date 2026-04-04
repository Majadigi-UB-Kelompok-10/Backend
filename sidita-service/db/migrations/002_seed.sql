-- =====================================================================
-- 1. SEED MASTER AREA (31 Kota/Kabupaten Jawa Timur Lengkap)
-- =====================================================================
INSERT INTO master_area (nama, slug, lat, lng) VALUES 
('Surabaya', 'surabaya', -7.250445, 112.768845),
('Banyuwangi', 'banyuwangi', -8.219233, 114.369227),
('Malang', 'malang', -7.983908, 112.621391),
('Batu', 'batu', -7.867100, 112.523900),
('Sidoarjo', 'sidoarjo', -7.447800, 112.718300),
('Gresik', 'gresik', -7.159400, 112.651300),
('Mojokerto', 'mojokerto', -7.469700, 112.433800),
('Pasuruan', 'pasuruan', -7.645300, 112.907500),
('Probolinggo', 'probolinggo', -7.744500, 113.215800),
('Jember', 'jember', -8.172500, 113.699500),
('Kediri', 'kediri', -7.822800, 112.011900),
('Blitar', 'blitar', -8.098300, 112.168100),
('Madiun', 'madiun', -7.629800, 111.523900),
('Pacitan', 'pacitan', -8.199400, 111.108500),
('Tulungagung', 'tulungagung', -8.065200, 111.902200),
('Trenggalek', 'trenggalek', -8.050500, 111.712200),
('Ponorogo', 'ponorogo', -7.869600, 111.465200),
('Magetan', 'magetan', -7.652300, 111.328300),
('Ngawi', 'ngawi', -7.406300, 111.446800),
('Bojonegoro', 'bojonegoro', -7.150200, 111.881800),
('Tuban', 'tuban', -6.898800, 112.049900),
('Lamongan', 'lamongan', -7.118600, 112.415100),
('Jombang', 'jombang', -7.546900, 112.226500),
('Nganjuk', 'nganjuk', -7.604400, 111.901500),
('Lumajang', 'lumajang', -8.133300, 113.224800),
('Bondowoso', 'bondowoso', -7.913500, 113.821400),
('Situbondo', 'situbondo', -7.701100, 113.999600),
('Bangkalan', 'bangkalan', -7.025200, 112.735900),
('Sampang', 'sampang', -7.187200, 113.243800),
('Pamekasan', 'pamekasan', -7.161100, 113.480900),
('Sumenep', 'sumenep', -7.013500, 113.865800);


-- =====================================================================
-- 2. SEED DESTINASI
-- =====================================================================
-- Keterangan: area_id 1 = Surabaya, 2 = Banyuwangi, 3 = Malang
INSERT INTO destinasi (area_id, kategori, nama, slug, gambar_url, deskripsi, alamat, highlight_text, lat, lng) VALUES 
(
    1, 
    'Taman', 
    'Taman Bungkul', 
    'taman-bungkul', 
    'https://res.cloudinary.com/dlscbkffi/image/upload/v1775289466/sidita_destinasi/q4lz0zlieispbhjl72di.jpg', 
    'Taman kota yang rindang dengan fasilitas lengkap di pusat Surabaya. Sangat cocok untuk bersantai dan olahraga pagi.', 
    'Jl. Taman Bungkul, Darmo, Kec. Wonokromo, Surabaya', 
    'Taman terbaik se-Asia Tenggara 2013', 
    -7.2913468, 
    112.7372415
),
(
    2, 
    'Pegunungan', 
    'Kawah Ijen', 
    'kawah-ijen', 
    'https://res.cloudinary.com/demo/image/upload/kawah-ijen.jpg', 
    'Gunung berapi aktif yang terkenal dengan fenomena blue fire yang sangat langka di dunia.', 
    'Perbatasan Kabupaten Banyuwangi dan Bondowoso', 
    'Fenomena Blue Fire Langka', 
    -8.0584, 
    114.2420
);


-- =====================================================================
-- 3. SEED HOTEL
-- =====================================================================
INSERT INTO hotel (area_id, nama, slug, harga_mulai, bintang, gambar_url, deskripsi, alamat, highlight_text, lat, lng) VALUES 
(
    1, 
    'Hotel Majapahit', 
    'hotel-majapahit', 
    1500000, 
    5, 
    'https://res.cloudinary.com/demo/image/upload/majapahit.jpg', 
    'Hotel bersejarah dengan arsitektur kolonial yang sangat mewah dan asri.', 
    'Jl. Tunjungan No.65, Genteng, Surabaya', 
    'Hotel bersejarah bintang 5 di pusat kota', 
    -7.2599, 
    112.7388
);


-- =====================================================================
-- 4. SEED EVENT
-- =====================================================================
INSERT INTO event (area_id, nama, slug, gambar_url, deskripsi, tanggal_mulai, tanggal_selesai, info_tiket, harga_tiket, lat, lng) VALUES 
(
    1, 
    'Surabaya Vaganza 2026', 
    'surabaya-vaganza-2026', 
    'https://res.cloudinary.com/dlscbkffi/image/upload/v1775320497/sidita_event/mbspiaqoiwkaq9mih3jy.jpg', 
    'Pawai bunga dan budaya yang sangat meriah untuk menyambut Hari Ulang Tahun (HUT) Kota Surabaya.', 
    '2026-05-26', 
    '2026-05-26', 
    'Gratis untuk Umum', 
    0, 
    -7.2625, 
    112.7425
);