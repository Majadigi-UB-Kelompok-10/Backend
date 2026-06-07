-- =============================================================================
-- SIDITA — SEED DATA (development/staging only)
-- =============================================================================

-- Master Area
INSERT INTO master_area (nama, slug, lat, lng) VALUES
    ('Kota Malang',       'kota-malang',       -7.9666200, 112.6326400),
    ('Kabupaten Malang',  'kabupaten-malang',  -8.1672500, 112.6635300),
    ('Kota Batu',         'kota-batu',         -7.8671700, 112.5239600);

-- Destinasi (sample 5 records)
INSERT INTO destinasi (
    area_id, nama, slug, kategori, deskripsi, alamat, highlight_text,
    gambar_url_thumbnail, gambar_url_hero, lat, lng
) VALUES
    (3, 'Jatim Park 1', 'jatim-park-1-a3f2b1c4', 'Taman Hiburan',
     'Taman hiburan keluarga dengan wahana edukasi dan rekreasi.',
     'Jl. Kartika No.2, Kota Batu',
     'Wahana edukasi terbaik di Malang',
     'https://upload.wikimedia.org/wikipedia/commons/2/24/Aquarium_Zone_at_Batu_Secret_Zoo_Jatim_Park_2.jpg',
     'https://upload.wikimedia.org/wikipedia/commons/2/24/Aquarium_Zone_at_Batu_Secret_Zoo_Jatim_Park_2.jpg',
     -7.8856700, 112.5283400),

    (3, 'Museum Angkut', 'museum-angkut-7c1d8f2a', 'Museum',
     'Museum transportasi dengan koleksi kendaraan dari berbagai era.',
     'Jl. Terusan Sultan Agung No.2, Kota Batu',
     'Museum transportasi terbesar di Asia Tenggara',
     'https://upload.wikimedia.org/wikipedia/commons/a/a8/Inside_Museum_Angkut_%28Transport_Museum%29%2C_Batu%2C_Malang_Indonesia_01.jpg',
     'https://upload.wikimedia.org/wikipedia/commons/a/a8/Inside_Museum_Angkut_%28Transport_Museum%29%2C_Batu%2C_Malang_Indonesia_01.jpg',
     -7.8853200, 112.5267300),

    (1, 'Kampung Warna-Warni Jodipan', 'kampung-jodipan-5e9a4b2d', 'Wisata Budaya',
     'Kampung dengan rumah-rumah berwarna-warni di tepi sungai Brantas.',
     'Jodipan, Blimbing, Kota Malang',
     'Spot foto Instagramable',
     'https://upload.wikimedia.org/wikipedia/commons/5/51/41_270925_Kampung_Warna-warni_Jodipan.jpg',
     'https://upload.wikimedia.org/wikipedia/commons/5/51/41_270925_Kampung_Warna-warni_Jodipan.jpg',
     -7.9854100, 112.6427800),

    (2, 'Pantai Balekambang', 'pantai-balekambang-9b3c7e1f', 'Pantai',
     'Pantai dengan pura di pulau kecil, sering disebut "Tanah Lot Jawa Timur".',
     'Bantur, Kabupaten Malang',
     'Pemandangan sunset spektakuler',
     'https://upload.wikimedia.org/wikipedia/commons/8/8e/Pura_Balekambang.png',
     'https://upload.wikimedia.org/wikipedia/commons/8/8e/Pura_Balekambang.png',
     -8.4014700, 112.5417200),

    (2, 'Coban Rondo', 'coban-rondo-2f8d6a3e', 'Air Terjun',
     'Air terjun setinggi 84 meter dengan area piknik dan labirin.',
     'Pandesari, Pujon, Kabupaten Malang',
     'Air terjun ikonik dengan labirin',
     'https://upload.wikimedia.org/wikipedia/commons/a/a4/Coban_Rondo_Waterfall.jpg',
     'https://upload.wikimedia.org/wikipedia/commons/a/a4/Coban_Rondo_Waterfall.jpg',
     -7.8846300, 112.4583900);

-- Hotel (sample 4 records)
INSERT INTO hotel (
    area_id, nama, slug, bintang, harga_mulai, deskripsi, alamat,
    highlight_text, gambar_url, lat, lng
) VALUES
    (1, 'Hotel Tugu Malang', 'hotel-tugu-malang-1a2b3c4d', 5, 1500000,
     'Hotel butik mewah dengan koleksi seni dan antik.',
     'Jl. Tugu No.3, Klojen, Kota Malang',
     'Hotel heritage di pusat kota',
     'https://picsum.photos/seed/tugumln/800/600',
     -7.9785400, 112.6304700),

    (3, 'Hotel Kartika Wijaya', 'hotel-kartika-wijaya-5e6f7a8b', 4, 850000,
     'Hotel resort di Batu dengan pemandangan pegunungan.',
     'Jl. Panglima Sudirman No.127, Kota Batu',
     'View pegunungan Panderman',
     'https://picsum.photos/seed/kartikabatu/800/600',
     -7.8742300, 112.5304100),

    (1, 'Aria Gajayana Hotel', 'aria-gajayana-9c1d2e3f', 4, 700000,
     'Hotel modern terintegrasi dengan mall di pusat Malang.',
     'Jl. Letjen S. Parman No.78, Kota Malang',
     'Lokasi strategis dekat mall',
     'https://picsum.photos/seed/ariamlg/800/600',
     -7.9612800, 112.6298500),

    (3, 'Singhasari Resort', 'singhasari-resort-4a5b6c7d', 5, 2000000,
     'Resort mewah dengan tema kerajaan Singhasari.',
     'Jl. Argo Wilis, Karangwidoro, Kota Batu',
     'Resort dengan tema sejarah',
     'https://picsum.photos/seed/singhasari/800/600',
     -7.8612400, 112.5198700);

-- Event (sample 4 records, mix tahun & bulan)
INSERT INTO event (
    area_id, nama, slug, deskripsi, alamat,
    tanggal_mulai, tanggal_selesai, info_tiket, harga_tiket,
    gambar_url_thumbnail, gambar_url_hero, lat, lng
) VALUES
    (1, 'Malang Flower Carnival 2026', 'malang-flower-carnival-2026-7b2c8d1e',
     'Karnaval bunga tahunan dengan kostum bertema flora khas Malang.',
     'Jl. Ijen, Kota Malang',
     '2026-05-15', '2026-05-15',
     'Gratis untuk umum', 0,
     'https://upload.wikimedia.org/wikipedia/commons/d/d3/Malang_flower_carnival.jpg',
     'https://upload.wikimedia.org/wikipedia/commons/d/d3/Malang_flower_carnival.jpg',
     -7.9712300, 112.6234800),

    (3, 'Festival Apel Batu 2026', 'festival-apel-batu-2026-3a4f5b6c',
     'Festival panen apel dengan kuliner khas Batu dan pertunjukan budaya.',
     'Alun-Alun Kota Batu',
     '2026-08-10', '2026-08-12',
     'Tiket masuk Rp25.000, anak-anak gratis', 25000,
     'https://picsum.photos/seed/applebatu/800/600',
     'https://picsum.photos/seed/applebatu/800/600',
     -7.8702100, 112.5238900),

    (2, 'Coban Rondo Run 2026', 'coban-rondo-run-2026-9d1e2f3a',
     'Lari trail 10K menuju air terjun Coban Rondo dengan medali finisher.',
     'Coban Rondo, Pujon',
     '2026-07-20', '2026-07-20',
     'Pendaftaran online Rp150.000', 150000,
     'https://picsum.photos/seed/cobanrun/800/600',
     'https://picsum.photos/seed/cobanrun/800/600',
     -7.8846300, 112.4583900),

    (1, 'Malang Tempo Doeloe 2026', 'malang-tempo-doeloe-2026-6e7a8b9c',
     'Festival nostalgia dengan kostum era 1900-an dan kuliner tempo dulu.',
     'Jl. Ijen, Kota Malang',
     '2026-09-05', '2026-09-07',
     'Gratis, donasi sukarela', 0,
     'https://picsum.photos/seed/tempodloe/800/600',
     'https://picsum.photos/seed/tempodloe/800/600',
     -7.9712300, 112.6234800);