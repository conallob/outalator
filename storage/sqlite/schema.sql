-- SQLite schema for outalator (local testing).
-- Applied automatically on startup when using the sqlite driver.
-- For production use PostgreSQL instead.
--
-- This schema mirrors:
--   migrations/001_initial_schema.sql
--   migrations/002_add_custom_fields.sql
-- Keep this file in sync when adding new PostgreSQL migration files.
--
-- Note: SQLite DATETIME stores timestamps with second precision. PostgreSQL
-- stores microseconds via timestamptz. Tests truncate to the second
-- (time.Truncate(time.Second)) to avoid spurious precision-related failures.

CREATE TABLE IF NOT EXISTS outages (
    id            TEXT PRIMARY KEY,
    title         TEXT NOT NULL,
    description   TEXT,
    status        TEXT NOT NULL,
    severity      TEXT NOT NULL,
    created_at    DATETIME NOT NULL,
    updated_at    DATETIME NOT NULL,
    resolved_at   DATETIME,
    metadata      TEXT NOT NULL DEFAULT '{}',
    custom_fields TEXT NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS alerts (
    id              TEXT PRIMARY KEY,
    outage_id       TEXT NOT NULL REFERENCES outages(id) ON DELETE CASCADE,
    external_id     TEXT NOT NULL,
    source          TEXT NOT NULL,
    team_name       TEXT,
    title           TEXT NOT NULL,
    description     TEXT,
    severity        TEXT,
    triggered_at    DATETIME NOT NULL,
    acknowledged_at DATETIME,
    resolved_at     DATETIME,
    created_at      DATETIME NOT NULL,
    source_metadata TEXT NOT NULL DEFAULT '{}',
    metadata        TEXT NOT NULL DEFAULT '{}',
    custom_fields   TEXT NOT NULL DEFAULT '{}',
    UNIQUE(external_id, source)
);

CREATE TABLE IF NOT EXISTS notes (
    id            TEXT PRIMARY KEY,
    outage_id     TEXT NOT NULL REFERENCES outages(id) ON DELETE CASCADE,
    content       TEXT NOT NULL,
    format        TEXT NOT NULL DEFAULT 'plaintext',
    author        TEXT NOT NULL,
    created_at    DATETIME NOT NULL,
    updated_at    DATETIME NOT NULL,
    metadata      TEXT NOT NULL DEFAULT '{}',
    custom_fields TEXT NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS tags (
    id            TEXT PRIMARY KEY,
    outage_id     TEXT NOT NULL REFERENCES outages(id) ON DELETE CASCADE,
    key           TEXT NOT NULL,
    value         TEXT NOT NULL,
    created_at    DATETIME NOT NULL,
    custom_fields TEXT NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_outages_created_at ON outages(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_outages_status     ON outages(status);
CREATE INDEX IF NOT EXISTS idx_outages_severity   ON outages(severity);

CREATE INDEX IF NOT EXISTS idx_alerts_outage_id    ON alerts(outage_id);
CREATE INDEX IF NOT EXISTS idx_alerts_external_id  ON alerts(external_id, source);
CREATE INDEX IF NOT EXISTS idx_alerts_triggered_at ON alerts(triggered_at DESC);

CREATE INDEX IF NOT EXISTS idx_notes_outage_id  ON notes(outage_id);
CREATE INDEX IF NOT EXISTS idx_notes_created_at ON notes(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_tags_outage_id ON tags(outage_id);
CREATE INDEX IF NOT EXISTS idx_tags_key_value ON tags(key, value);
