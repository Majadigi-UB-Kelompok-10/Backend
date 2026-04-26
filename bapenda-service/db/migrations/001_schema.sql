CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE kendaraan_pajak (
    plat_nomor           VARCHAR(15) PRIMARY KEY,
    plat_nomor_display   VARCHAR(15) NOT NULL,
    nomor_rangka         VARCHAR(50) NOT NULL,
    status_aktif         BOOLEAN NOT NULL DEFAULT TRUE,
    merk                 VARCHAR(50) NOT NULL,
    tipe                 VARCHAR(100) NOT NULL,
    warna                VARCHAR(50) NOT NULL,
    tahun_buat           SMALLINT NOT NULL,
    model                VARCHAR(50) NOT NULL,
    masa_pajak           DATE NOT NULL,
    pkb_pokok            INTEGER NOT NULL DEFAULT 0,
    opsen_pkb            INTEGER NOT NULL DEFAULT 0,
    swdkllj              INTEGER NOT NULL DEFAULT 0,
    parkir_berlangganan  INTEGER NOT NULL DEFAULT 0,
    cetak_stnk           INTEGER NOT NULL DEFAULT 0,
    cetak_tnkb           INTEGER NOT NULL DEFAULT 0,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    total_pajak_tahunan  INTEGER GENERATED ALWAYS AS (
        pkb_pokok + opsen_pkb + swdkllj + parkir_berlangganan
    ) STORED
);

CREATE TABLE tarif_pkb (
    id                  SERIAL PRIMARY KEY,
    jenis_plat          VARCHAR(10) NOT NULL,
    label               VARCHAR(50) NOT NULL,
    tarif_pkb_persen    DECIMAL(5,4) NOT NULL,
    opsen_pkb_persen    DECIMAL(5,4) NOT NULL,
    bbn1_persen         DECIMAL(5,4) NOT NULL DEFAULT 0,
    opsen_bbn1_persen   DECIMAL(5,4) NOT NULL DEFAULT 0,
    bbn2_persen         DECIMAL(5,4) NOT NULL DEFAULT 0,
    CONSTRAINT uq_tarif_plat UNIQUE (jenis_plat)
);

CREATE TABLE master_njkb (
    id              INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    jenis_kendaraan VARCHAR(50)  NOT NULL,
    merk            VARCHAR(50)  NOT NULL,
    model           VARCHAR(50)  NOT NULL,
    tipe            VARCHAR(100) NOT NULL,
    tahun           SMALLINT     NOT NULL,
    nilai_jual      BIGINT       NOT NULL CHECK (nilai_jual > 0),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_njkb UNIQUE (merk, tipe, tahun)
);


CREATE UNIQUE INDEX idx_kendaraan_plat_rangka ON kendaraan_pajak(plat_nomor, RIGHT(nomor_rangka, 5));
CREATE INDEX idx_kendaraan_plat_trgm ON kendaraan_pajak USING GIN (plat_nomor gin_trgm_ops);
CREATE INDEX idx_kendaraan_masa_pajak ON kendaraan_pajak(masa_pajak);
CREATE INDEX idx_njkb_full ON master_njkb(jenis_kendaraan, merk, model, tipe, tahun);

CREATE INDEX idx_kendaraan_status_aktif ON kendaraan_pajak(status_aktif) WHERE status_aktif = true;