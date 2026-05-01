CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Master data tarif PKB per jenis plat
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

-- Master NJKB (Nilai Jual Kendaraan Bermotor)
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

-- Data pajak kendaraan per plat
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

-- kendaraan_pajak: trigram untuk LIKE search by plat (admin search feature)
CREATE INDEX idx_kendaraan_plat_trgm 
    ON kendaraan_pajak USING GIN (plat_nomor gin_trgm_ops);

-- kendaraan_pajak: range query by masa_pajak (reminder jatuh tempo)
CREATE INDEX idx_kendaraan_masa_pajak 
    ON kendaraan_pajak(masa_pajak);

-- kendaraan_pajak: ORDER BY created_at DESC (admin pagination)
CREATE INDEX idx_kendaraan_created_at_desc 
    ON kendaraan_pajak(created_at DESC);

-- kendaraan_pajak: partial index — cuma row aktif (efisien, hemat space)
CREATE INDEX idx_kendaraan_status_aktif 
    ON kendaraan_pajak(status_aktif) 
    WHERE status_aktif = true;

-- master_njkb: composite index untuk cascading dropdown query
CREATE INDEX idx_njkb_full 
    ON master_njkb(jenis_kendaraan, merk, model, tipe, tahun);

-- =============================================================================
-- TRIGGERS — auto-update updated_at column
-- =============================================================================

CREATE OR REPLACE FUNCTION trg_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_kendaraan_pajak_updated_at
    BEFORE UPDATE ON kendaraan_pajak
    FOR EACH ROW
    EXECUTE FUNCTION trg_set_updated_at();