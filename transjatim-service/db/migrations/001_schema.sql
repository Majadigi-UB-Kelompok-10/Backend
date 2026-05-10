CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TYPE layanan_type AS ENUM ('reguler', 'luxury');
CREATE TYPE tipe_penumpang_type AS ENUM ('umum', 'pelajar_santri', 'mahasiswa');

CREATE TABLE terminal (
    id         INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    nama       VARCHAR(150) NOT NULL,
    kota       VARCHAR(100) NOT NULL,
    slug       VARCHAR(200) NOT NULL,
    lat        DOUBLE PRECISION,
    lng        DOUBLE PRECISION,
    aktif      BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_terminal_slug      UNIQUE (slug),
    CONSTRAINT uq_terminal_nama_kota UNIQUE (nama, kota),
    CONSTRAINT chk_terminal_slug     CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$')
);

CREATE TABLE bus (
    id         INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    kode       VARCHAR(50) NOT NULL,
    layanan    layanan_type NOT NULL,
    fasilitas  TEXT,
    kapasitas  INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_bus_kode       UNIQUE (kode),
    CONSTRAINT chk_bus_kapasitas CHECK (kapasitas IS NULL OR kapasitas > 0)
);

CREATE TABLE rute (
    id                 INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    terminal_asal_id   INT NOT NULL REFERENCES terminal(id),
    terminal_tujuan_id INT NOT NULL REFERENCES terminal(id),
    slug               VARCHAR(300) NOT NULL,
    durasi_menit       INT NOT NULL,
    aktif              BOOLEAN NOT NULL DEFAULT true,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_rute_slug        UNIQUE (slug),
    CONSTRAINT uq_rute_asal_tujuan UNIQUE (terminal_asal_id, terminal_tujuan_id),
    CONSTRAINT chk_rute_beda       CHECK (terminal_asal_id != terminal_tujuan_id),
    CONSTRAINT chk_rute_durasi     CHECK (durasi_menit > 0)
);

CREATE TABLE rute_stop (
    id          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    rute_id     INT NOT NULL REFERENCES rute(id) ON DELETE CASCADE,
    terminal_id INT NOT NULL REFERENCES terminal(id),
    urutan      INT NOT NULL,
    CONSTRAINT uq_rute_stop_urutan   UNIQUE (rute_id, urutan),
    CONSTRAINT uq_rute_stop_terminal UNIQUE (rute_id, terminal_id)
);

-- hari_operasi bitmask: 1=Sen,2=Sel,4=Rab,8=Kam,16=Jum,32=Sab,64=Min, 127=setiap hari
CREATE TABLE jadwal (
    id            INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    rute_id       INT NOT NULL REFERENCES rute(id) ON DELETE CASCADE,
    bus_id        INT NOT NULL REFERENCES bus(id),
    jam_berangkat TIME NOT NULL,
    jam_tiba      TIME NOT NULL,
    hari_operasi  SMALLINT NOT NULL DEFAULT 127,
    aktif         BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_jadwal_hari CHECK (hari_operasi > 0 AND hari_operasi <= 127)
);

CREATE TABLE jadwal_libur (
    id          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    jadwal_id   INT REFERENCES jadwal(id) ON DELETE CASCADE,
    tanggal     DATE NOT NULL,
    keterangan  VARCHAR(255),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_jadwal_libur UNIQUE (jadwal_id, tanggal)
);

CREATE TABLE harga (
    id             INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    rute_id        INT NOT NULL REFERENCES rute(id) ON DELETE CASCADE,
    layanan        layanan_type NOT NULL,
    tipe_penumpang tipe_penumpang_type,
    harga          INT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_harga_rute_layanan_tipe UNIQUE (rute_id, layanan, tipe_penumpang),
    CONSTRAINT chk_harga_positif          CHECK (harga > 0),
    CONSTRAINT chk_harga_tipe             CHECK (
        (layanan = 'luxury'  AND tipe_penumpang IS NULL) OR
        (layanan = 'reguler' AND tipe_penumpang IS NOT NULL)
    )
);

CREATE INDEX idx_terminal_kota      ON terminal(kota);
CREATE INDEX idx_terminal_nama_trgm ON terminal USING GIN (nama gin_trgm_ops);
CREATE INDEX idx_rute_asal_tujuan   ON rute(terminal_asal_id, terminal_tujuan_id) WHERE aktif = true;
CREATE INDEX idx_rute_stop_rute     ON rute_stop(rute_id, urutan);
CREATE INDEX idx_jadwal_rute        ON jadwal(rute_id) WHERE aktif = true;
CREATE INDEX idx_jadwal_libur_tgl   ON jadwal_libur(tanggal);
CREATE INDEX idx_harga_rute         ON harga(rute_id, layanan);

CREATE TRIGGER set_ts_terminal BEFORE UPDATE ON terminal FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
CREATE TRIGGER set_ts_bus      BEFORE UPDATE ON bus      FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
CREATE TRIGGER set_ts_rute     BEFORE UPDATE ON rute     FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
CREATE TRIGGER set_ts_jadwal   BEFORE UPDATE ON jadwal   FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
CREATE TRIGGER set_ts_harga    BEFORE UPDATE ON harga    FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();
