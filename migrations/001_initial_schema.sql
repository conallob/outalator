-- Create outages table
CREATE TABLE IF NOT EXISTS outages (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL,
    severity VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    resolved_at TIMESTAMP
);

-- Create alerts table
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY,
    outage_id UUID NOT NULL REFERENCES outages(id) ON DELETE CASCADE,
    external_id VARCHAR(255) NOT NULL,
    source VARCHAR(50) NOT NULL,
    team_name VARCHAR(255),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    severity VARCHAR(50),
    triggered_at TIMESTAMP NOT NULL,
    acknowledged_at TIMESTAMP,
    resolved_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    UNIQUE(external_id, source)
);

-- Create notes table
CREATE TABLE IF NOT EXISTS notes (
    id UUID PRIMARY KEY,
    outage_id UUID NOT NULL REFERENCES outages(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    format VARCHAR(20) NOT NULL DEFAULT 'plaintext',
    author VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Create tags table
CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY,
    outage_id UUID NOT NULL REFERENCES outages(id) ON DELETE CASCADE,
    key VARCHAR(100) NOT NULL,
    value VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_outages_created_at ON outages(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_outages_status ON outages(status);
CREATE INDEX IF NOT EXISTS idx_outages_severity ON outages(severity);

CREATE INDEX IF NOT EXISTS idx_alerts_outage_id ON alerts(outage_id);
CREATE INDEX IF NOT EXISTS idx_alerts_external_id ON alerts(external_id, source);
CREATE INDEX IF NOT EXISTS idx_alerts_triggered_at ON alerts(triggered_at DESC);

CREATE INDEX IF NOT EXISTS idx_notes_outage_id ON notes(outage_id);
CREATE INDEX IF NOT EXISTS idx_notes_created_at ON notes(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_tags_outage_id ON tags(outage_id);
CREATE INDEX IF NOT EXISTS idx_tags_key_value ON tags(key, value);
