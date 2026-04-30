CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE master_area (
    id          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    nama        VARCHAR(100) NOT NULL,
    slug        VARCHAR(100) NOT NULL,
    lat         DECIMAL(10, 7),
    lng         DECIMAL(10, 7),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uq_master_area_slug UNIQUE (slug),
    CONSTRAINT uq_master_area_nama UNIQUE (nama),
    CONSTRAINT chk_master_area_slug CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$')
);


CREATE TABLE destinasi (
    id                   INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    area_id              INT NOT NULL REFERENCES master_area(id) ON DELETE RESTRICT,
    nama                 VARCHAR(255) NOT NULL,
    slug                 VARCHAR(255) NOT NULL,
    kategori             VARCHAR(50)  NOT NULL,
    deskripsi            TEXT NOT NULL,
    alamat               TEXT NOT NULL,
    highlight_text       VARCHAR(255),
    gambar_url_thumbnail TEXT NOT NULL,
    gambar_url_hero      TEXT NOT NULL,
    lat                  DECIMAL(10, 7) NOT NULL,
    lng                  DECIMAL(10, 7) NOT NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uq_destinasi_slug UNIQUE (slug),
    CONSTRAINT chk_destinasi_slug CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    CONSTRAINT chk_destinasi_thumbnail CHECK (gambar_url_thumbnail ~* '^https?://[^\s]+$'),
    CONSTRAINT chk_destinasi_hero CHECK (gambar_url_hero ~* '^https?://[^\s]+$')
);


CREATE TABLE hotel (
    id              INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    area_id         INT NOT NULL REFERENCES master_area(id) ON DELETE RESTRICT,
    nama            VARCHAR(255) NOT NULL,
    slug            VARCHAR(255) NOT NULL,
    bintang         SMALLINT NOT NULL,
    harga_mulai     INTEGER NOT NULL DEFAULT 0,
    deskripsi       TEXT,
    alamat          TEXT NOT NULL,
    highlight_text  VARCHAR(255),
    gambar_url      TEXT NOT NULL,
    lat             DECIMAL(10, 7) NOT NULL,
    lng             DECIMAL(10, 7) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uq_hotel_slug UNIQUE (slug),
    CONSTRAINT chk_hotel_slug CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    CONSTRAINT chk_hotel_bintang CHECK (bintang BETWEEN 1 AND 5),
    CONSTRAINT chk_hotel_harga CHECK (harga_mulai >= 0),
    CONSTRAINT chk_hotel_gambar CHECK (gambar_url ~* '^https?://[^\s]+$')
);


CREATE TABLE event (
    id                   INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    area_id              INT NOT NULL REFERENCES master_area(id) ON DELETE RESTRICT,
    nama                 VARCHAR(255) NOT NULL,
    slug                 VARCHAR(255) NOT NULL,
    deskripsi            TEXT NOT NULL,
    alamat               TEXT NOT NULL,
    tanggal_mulai        DATE NOT NULL,
    tanggal_selesai      DATE NOT NULL,
    info_tiket           TEXT NOT NULL DEFAULT 'Gratis Umum',
    harga_tiket          INTEGER NOT NULL DEFAULT 0,
    gambar_url_thumbnail TEXT NOT NULL,
    gambar_url_hero      TEXT NOT NULL,
    lat                  DECIMAL(10, 7) NOT NULL,
    lng                  DECIMAL(10, 7) NOT NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    tahun  SMALLINT GENERATED ALWAYS AS (EXTRACT(YEAR  FROM tanggal_mulai)::SMALLINT) STORED,
    bulan  SMALLINT GENERATED ALWAYS AS (EXTRACT(MONTH FROM tanggal_mulai)::SMALLINT) STORED,

    CONSTRAINT uq_event_slug UNIQUE (slug),
    CONSTRAINT chk_event_slug CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    CONSTRAINT chk_event_tanggal CHECK (tanggal_selesai >= tanggal_mulai),
    CONSTRAINT chk_event_harga CHECK (harga_tiket >= 0),
    CONSTRAINT chk_event_thumbnail CHECK (gambar_url_thumbnail ~* '^https?://[^\s]+$'),
    CONSTRAINT chk_event_hero CHECK (gambar_url_hero ~* '^https?://[^\s]+$')
);


CREATE INDEX idx_destinasi_area         ON destinasi(area_id);
CREATE INDEX idx_destinasi_kategori     ON destinasi(kategori);
CREATE INDEX idx_destinasi_created_at   ON destinasi(created_at DESC);
CREATE INDEX idx_destinasi_nama_trgm    ON destinasi USING GIN (nama gin_trgm_ops);

CREATE INDEX idx_hotel_area         ON hotel(area_id);
CREATE INDEX idx_hotel_bintang      ON hotel(bintang DESC);
CREATE INDEX idx_hotel_created_at   ON hotel(created_at DESC);
CREATE INDEX idx_hotel_nama_trgm    ON hotel USING GIN (nama gin_trgm_ops);

CREATE INDEX idx_event_area             ON event(area_id);
CREATE INDEX idx_event_tahun            ON event(tahun);
CREATE INDEX idx_event_bulan            ON event(bulan);
CREATE INDEX idx_event_tanggal_mulai    ON event(tanggal_mulai DESC);
CREATE INDEX idx_event_created_at       ON event(created_at DESC);
CREATE INDEX idx_event_nama_trgm        ON event USING GIN (nama gin_trgm_ops);



CREATE TRIGGER set_timestamp_destinasi
    BEFORE UPDATE ON destinasi
    FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_hotel
    BEFORE UPDATE ON hotel
    FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_event
    BEFORE UPDATE ON event
    FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();