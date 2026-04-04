-- Aktifkan ekstensi Trigram untuk fitur pencarian super pintar (Typo-Tolerant)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE master_area (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    nama VARCHAR(100) UNIQUE NOT NULL, 
    slug VARCHAR(100) UNIQUE NOT NULL,
    lat DECIMAL(10, 8), 
    lng DECIMAL(11, 8)
);

CREATE TABLE destinasi (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    area_id INT NOT NULL REFERENCES master_area(id),
    kategori VARCHAR(100) NOT NULL,
    nama VARCHAR(255) UNIQUE NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    gambar_url TEXT NOT NULL,
    deskripsi TEXT NOT NULL,
    alamat TEXT NOT NULL,
    highlight_text TEXT NOT NULL,
    lat DECIMAL(10, 8), 
    lng DECIMAL(11, 8),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE destinasi_gambar (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    destinasi_id INT NOT NULL REFERENCES destinasi(id) ON DELETE CASCADE,
    gambar_url TEXT NOT NULL,
    urutan INT DEFAULT 0
);

CREATE TABLE hotel (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    area_id INT NOT NULL REFERENCES master_area(id),
    nama VARCHAR(255) UNIQUE NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    harga_mulai INTEGER DEFAULT 0,
    bintang SMALLINT DEFAULT 0 CHECK (bintang BETWEEN 0 AND 5),
    gambar_url TEXT NOT NULL,
    deskripsi TEXT NOT NULL,
    alamat TEXT NOT NULL,
    highlight_text TEXT,
    lat DECIMAL(10, 8), 
    lng DECIMAL(11, 8),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE event (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    area_id INT NOT NULL REFERENCES master_area(id),
    nama VARCHAR(255) UNIQUE NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    gambar_url TEXT NOT NULL,
    deskripsi TEXT NOT NULL,
    tanggal_mulai DATE NOT NULL,
    tanggal_selesai DATE NOT NULL,
    tahun SMALLINT GENERATED ALWAYS AS (EXTRACT(YEAR FROM tanggal_mulai)::SMALLINT) STORED,
    info_tiket VARCHAR(100) DEFAULT 'Gratis Umum',
    lat DECIMAL(10, 8),
    lng DECIMAL(11, 8),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- INDEXES
CREATE INDEX idx_master_area_slug ON master_area(slug);

CREATE INDEX idx_destinasi_slug     ON destinasi(slug);
CREATE INDEX idx_destinasi_area     ON destinasi(area_id);
CREATE INDEX idx_destinasi_kategori ON destinasi(kategori);
CREATE INDEX idx_destinasi_nama_search ON destinasi USING GIN (nama gin_trgm_ops);

CREATE INDEX idx_hotel_slug  ON hotel(slug);
CREATE INDEX idx_hotel_area  ON hotel(area_id);
CREATE INDEX idx_hotel_nama_search ON hotel USING GIN (nama gin_trgm_ops);

CREATE INDEX idx_event_slug         ON event(slug);
CREATE INDEX idx_event_area         ON event(area_id);
CREATE INDEX idx_event_tanggal      ON event(tanggal_mulai);
CREATE INDEX idx_event_tahun        ON event(tahun);