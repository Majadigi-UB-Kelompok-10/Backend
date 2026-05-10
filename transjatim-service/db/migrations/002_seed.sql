-- =====================================================================
-- SEED: Terminal
-- =====================================================================
INSERT INTO terminal (nama, kota, slug, lat, lng) VALUES
    ('Terminal Hamid Rusdi',   'Malang',   'terminal-hamid-rusdi',   -7.9924, 112.6425),
    ('Terminal Arjosari',      'Malang',   'terminal-arjosari',      -7.9328, 112.6501),
    ('Terminal Batu',          'Batu',     'terminal-batu',          -7.8696, 112.5295),
    ('Terminal Bungurasih',    'Surabaya', 'terminal-bungurasih',    -7.3648, 112.7322),
    ('Terminal Osowilangun',   'Surabaya', 'terminal-osowilangun',   -7.2453, 112.6582),
    ('Terminal Taman Dayu',    'Pasuruan', 'terminal-taman-dayu',    -7.7412, 112.7851);

-- =====================================================================
-- SEED: Bus
-- =====================================================================
INSERT INTO bus (kode, layanan, fasilitas, kapasitas) VALUES
    ('BUS-V01', 'reguler', NULL, NULL),
    ('BUS-V02', 'reguler', NULL, NULL),
    ('BUS-V03', 'luxury',  'Kursi premium (tanpa berdiri) dan AC ekstra dingin.', NULL),
    ('BUS-V04', 'luxury',  'Kursi premium (tanpa berdiri) dan AC ekstra dingin.', NULL);

-- =====================================================================
-- SEED: Rute (PP Malang Kota ↔ Batu)
-- =====================================================================
INSERT INTO rute (terminal_asal_id, terminal_tujuan_id, slug, durasi_menit)
SELECT ta.id, tt.id, 'terminal-hamid-rusdi-ke-terminal-batu', 80
FROM terminal ta, terminal tt
WHERE ta.slug = 'terminal-hamid-rusdi' AND tt.slug = 'terminal-batu';

INSERT INTO rute (terminal_asal_id, terminal_tujuan_id, slug, durasi_menit)
SELECT ta.id, tt.id, 'terminal-batu-ke-terminal-hamid-rusdi', 80
FROM terminal ta, terminal tt
WHERE ta.slug = 'terminal-batu' AND tt.slug = 'terminal-hamid-rusdi';

-- Rute PP Surabaya ↔ Malang
INSERT INTO rute (terminal_asal_id, terminal_tujuan_id, slug, durasi_menit)
SELECT ta.id, tt.id, 'terminal-bungurasih-ke-terminal-hamid-rusdi', 120
FROM terminal ta, terminal tt
WHERE ta.slug = 'terminal-bungurasih' AND tt.slug = 'terminal-hamid-rusdi';

INSERT INTO rute (terminal_asal_id, terminal_tujuan_id, slug, durasi_menit)
SELECT ta.id, tt.id, 'terminal-hamid-rusdi-ke-terminal-bungurasih', 120
FROM terminal ta, terminal tt
WHERE ta.slug = 'terminal-hamid-rusdi' AND tt.slug = 'terminal-bungurasih';

-- =====================================================================
-- SEED: Rute Stop (Malang → Batu)
-- =====================================================================
INSERT INTO rute_stop (rute_id, terminal_id, urutan)
SELECT r.id, t.id, 1
FROM rute r, terminal t
WHERE r.slug = 'terminal-hamid-rusdi-ke-terminal-batu' AND t.slug = 'terminal-hamid-rusdi';

INSERT INTO rute_stop (rute_id, terminal_id, urutan)
SELECT r.id, t.id, 2
FROM rute r, terminal t
WHERE r.slug = 'terminal-hamid-rusdi-ke-terminal-batu' AND t.slug = 'terminal-arjosari';

INSERT INTO rute_stop (rute_id, terminal_id, urutan)
SELECT r.id, t.id, 3
FROM rute r, terminal t
WHERE r.slug = 'terminal-hamid-rusdi-ke-terminal-batu' AND t.slug = 'terminal-batu';

-- Rute Stop (Batu → Malang)
INSERT INTO rute_stop (rute_id, terminal_id, urutan)
SELECT r.id, t.id, 1
FROM rute r, terminal t
WHERE r.slug = 'terminal-batu-ke-terminal-hamid-rusdi' AND t.slug = 'terminal-batu';

INSERT INTO rute_stop (rute_id, terminal_id, urutan)
SELECT r.id, t.id, 2
FROM rute r, terminal t
WHERE r.slug = 'terminal-batu-ke-terminal-hamid-rusdi' AND t.slug = 'terminal-arjosari';

INSERT INTO rute_stop (rute_id, terminal_id, urutan)
SELECT r.id, t.id, 3
FROM rute r, terminal t
WHERE r.slug = 'terminal-batu-ke-terminal-hamid-rusdi' AND t.slug = 'terminal-hamid-rusdi';

-- =====================================================================
-- SEED: Jadwal (Malang → Batu, reguler BUS-V01)
-- =====================================================================
INSERT INTO jadwal (rute_id, bus_id, jam_berangkat, jam_tiba, hari_operasi)
SELECT r.id, b.id, '05:00', '06:20', 127
FROM rute r, bus b
WHERE r.slug = 'terminal-hamid-rusdi-ke-terminal-batu' AND b.kode = 'BUS-V01';

INSERT INTO jadwal (rute_id, bus_id, jam_berangkat, jam_tiba, hari_operasi)
SELECT r.id, b.id, '08:00', '09:20', 127
FROM rute r, bus b
WHERE r.slug = 'terminal-hamid-rusdi-ke-terminal-batu' AND b.kode = 'BUS-V02';

-- Jadwal (Malang → Batu, luxury BUS-V03)
INSERT INTO jadwal (rute_id, bus_id, jam_berangkat, jam_tiba, hari_operasi)
SELECT r.id, b.id, '07:00', '08:20', 127
FROM rute r, bus b
WHERE r.slug = 'terminal-hamid-rusdi-ke-terminal-batu' AND b.kode = 'BUS-V03';

-- Jadwal (Surabaya → Malang, luxury BUS-V04)
INSERT INTO jadwal (rute_id, bus_id, jam_berangkat, jam_tiba, hari_operasi)
SELECT r.id, b.id, '06:00', '08:00', 127
FROM rute r, bus b
WHERE r.slug = 'terminal-bungurasih-ke-terminal-hamid-rusdi' AND b.kode = 'BUS-V04';

-- =====================================================================
-- SEED: Harga
-- =====================================================================

-- Rute Malang → Batu (reguler)
INSERT INTO harga (rute_id, layanan, tipe_penumpang, harga)
SELECT r.id, 'reguler', 'umum',           5000 FROM rute r WHERE r.slug = 'terminal-hamid-rusdi-ke-terminal-batu';
INSERT INTO harga (rute_id, layanan, tipe_penumpang, harga)
SELECT r.id, 'reguler', 'pelajar_santri', 2500 FROM rute r WHERE r.slug = 'terminal-hamid-rusdi-ke-terminal-batu';
INSERT INTO harga (rute_id, layanan, tipe_penumpang, harga)
SELECT r.id, 'reguler', 'mahasiswa',      2500 FROM rute r WHERE r.slug = 'terminal-hamid-rusdi-ke-terminal-batu';

-- Rute Malang → Batu (luxury)
INSERT INTO harga (rute_id, layanan, tipe_penumpang, harga)
SELECT r.id, 'luxury', NULL, 20000 FROM rute r WHERE r.slug = 'terminal-hamid-rusdi-ke-terminal-batu';

-- Rute Batu → Malang (reguler)
INSERT INTO harga (rute_id, layanan, tipe_penumpang, harga)
SELECT r.id, 'reguler', 'umum',           5000 FROM rute r WHERE r.slug = 'terminal-batu-ke-terminal-hamid-rusdi';
INSERT INTO harga (rute_id, layanan, tipe_penumpang, harga)
SELECT r.id, 'reguler', 'pelajar_santri', 2500 FROM rute r WHERE r.slug = 'terminal-batu-ke-terminal-hamid-rusdi';
INSERT INTO harga (rute_id, layanan, tipe_penumpang, harga)
SELECT r.id, 'reguler', 'mahasiswa',      2500 FROM rute r WHERE r.slug = 'terminal-batu-ke-terminal-hamid-rusdi';

-- Rute Surabaya → Malang (luxury)
INSERT INTO harga (rute_id, layanan, tipe_penumpang, harga)
SELECT r.id, 'luxury', NULL, 20000 FROM rute r WHERE r.slug = 'terminal-bungurasih-ke-terminal-hamid-rusdi';
