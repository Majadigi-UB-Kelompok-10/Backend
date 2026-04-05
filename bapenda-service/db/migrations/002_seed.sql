INSERT INTO tarif_pkb (jenis_plat, label, tarif_pkb_persen, opsen_pkb_persen, bbn1_persen, opsen_bbn1_persen, bbn2_persen)
VALUES
    ('hitam', 'Umum/Pribadi',  0.0150, 0.6600, 0.1200, 0.6600, 0.0100),
    ('kuning','Angkutan Umum', 0.0075, 0.6600, 0.0600, 0.6600, 0.0075),
    ('merah', 'Pemerintah',    0.0075, 0.6600, 0.0000, 0.0000, 0.0000);


INSERT INTO master_njkb (jenis_kendaraan, merk, model, tipe, tahun, nilai_jual) VALUES
('Motor', 'HONDA', 'VARIO', '125 CBS', 2022, 15000000),
('Motor', 'YAMAHA', 'NMAX', '155 ABS', 2023, 24000000),
('Motor', 'HONDA', 'BEAT', 'CBS ISS', 2021, 12000000),
('Mobil', 'TOYOTA', 'AVANZA', '1.5 G MT', 2023, 155000000),
('Mobil', 'HONDA', 'BRIO', 'SATYA E CVT', 2022, 110000000),
('Mobil', 'MITSUBISHI', 'PAJERO SPORT', 'DAKAR 4X2', 2023, 450000000),
('Truk', 'MITSUBISHI', 'FUSO', 'FE 74 HD 125 PS', 2020, 250000000),
('Truk', 'HINO', 'DUTRO', '130 HD', 2021, 275000000);


INSERT INTO kendaraan_pajak (
    plat_nomor, plat_nomor_display, nomor_rangka, status_aktif, 
    merk, model, tipe, warna, tahun_buat, masa_pajak, 
    pkb_pokok, opsen_pkb, swdkllj, parkir_berlangganan, cetak_stnk, cetak_tnkb
) VALUES

('N1234AB', 'N 1234 AB', 'MHK12345', true, 
 'HONDA', 'VARIO', '125 CBS', 'HITAM', 2022, '2026-05-10', 
 225000, 150000, 35000, 20000, 0, 0),

('N5678CD', 'N 5678 CD', 'MHY56789', true, 
 'YAMAHA', 'NMAX', '155 ABS', 'BIRU', 2023, '2028-08-17', 
 360000, 240000, 35000, 20000, 100000, 60000),

('N9999ZZ', 'N 9999 ZZ', 'MHT99999', false, 
 'TOYOTA', 'AVANZA', '1.5 G MT', 'PUTIH', 2023, '2024-01-01', 
 2325000, 1550000, 143000, 40000, 0, 0),

('L1111AA', 'L 1111 AA', 'MHB11111', true, 
 'HONDA', 'BRIO', 'SATYA E CVT', 'MERAH', 2022, '2026-11-20', 
 1650000, 1100000, 143000, 40000, 0, 0);