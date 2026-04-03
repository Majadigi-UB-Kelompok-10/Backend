CREATE TABLE master_area (
    id SERIAL PRIMARY KEY,
    nama VARCHAR(100) UNIQUE NOT NULL, 
    slug VARCHAR(100) UNIQUE NOT NULL  
);

CREATE TABLE destinasi(
    id SERIAL PRIMARY KEY,
    area_id INT NOT NULL REFERENCES master_area(id),
    kategori VARCHAR(100) NOT NULL,
    nama VARCHAR(255) UNIQUE NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    gambar_url TEXT,
    deskripsi TEXT NOT NULL,
    alamat TEXT NOT NULL,
    highlight_text TEXT NOT NULL
    lat DECIMAL(10, 8), 
    lng DECIMAL(11, 8),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)

CREATE TABLE hotel(
    id SERIAL PRIMARY KEY,
    area_id INT NOT NULL REFERENCES master_area(id),
    nama VARCHAR(255) UNIQUE NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    gambar_url TEXT,
    deskripsi TEXT NOT NULL,
    alamat TEXT NOT NULL,
    highlight_text TEXT,
    lat DECIMAL(10, 8), 
    lng DECIMAL(11, 8),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)

CREATE TABLE event (
    id SERIAL PRIMARY KEY,
    area_id INT NOT NULL REFERENCES master_area(id),
    nama VARCHAR(255) UNIQUE NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    gambar_url TEXT NOT NULL,
    deskripsi TEXT NOT NULL,
    tanggal_mulai DATE NOT NULL,
    tanggal_selesai DATE NOT NULL,
    info_tiket VARCHAR(100) DEFAULT 'Gratis Umum',
    lat DECIMAL(10, 8),
    lng DECIMAL(11, 8),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);