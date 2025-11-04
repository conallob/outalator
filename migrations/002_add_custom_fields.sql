-- Add custom metadata and fields to outages table
ALTER TABLE outages
    ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS custom_fields JSONB DEFAULT '{}';

-- Add custom metadata and fields to alerts table
ALTER TABLE alerts
    ADD COLUMN IF NOT EXISTS source_metadata JSONB DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS custom_fields JSONB DEFAULT '{}';

-- Add custom metadata and fields to notes table
ALTER TABLE notes
    ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS custom_fields JSONB DEFAULT '{}';

-- Add custom fields to tags table
ALTER TABLE tags
    ADD COLUMN IF NOT EXISTS custom_fields JSONB DEFAULT '{}';

-- Create indexes for JSONB columns for better query performance
-- These GIN indexes allow efficient querying of JSON data

-- Outages metadata indexes
CREATE INDEX IF NOT EXISTS idx_outages_metadata ON outages USING GIN (metadata);
CREATE INDEX IF NOT EXISTS idx_outages_custom_fields ON outages USING GIN (custom_fields);

-- Alerts metadata indexes
CREATE INDEX IF NOT EXISTS idx_alerts_source_metadata ON alerts USING GIN (source_metadata);
CREATE INDEX IF NOT EXISTS idx_alerts_metadata ON alerts USING GIN (metadata);
CREATE INDEX IF NOT EXISTS idx_alerts_custom_fields ON alerts USING GIN (custom_fields);

-- Notes metadata indexes
CREATE INDEX IF NOT EXISTS idx_notes_metadata ON notes USING GIN (metadata);
CREATE INDEX IF NOT EXISTS idx_notes_custom_fields ON notes USING GIN (custom_fields);

-- Tags custom_fields index
CREATE INDEX IF NOT EXISTS idx_tags_custom_fields ON tags USING GIN (custom_fields);

-- Add comments to document the purpose of these fields
COMMENT ON COLUMN outages.metadata IS 'Simple key-value pairs for custom metadata';
COMMENT ON COLUMN outages.custom_fields IS 'Complex structured data for extensibility';

COMMENT ON COLUMN alerts.source_metadata IS 'Source-specific metadata from PagerDuty, OpsGenie, etc.';
COMMENT ON COLUMN alerts.metadata IS 'Simple key-value pairs for custom metadata';
COMMENT ON COLUMN alerts.custom_fields IS 'Complex structured data for extensibility';

COMMENT ON COLUMN notes.metadata IS 'Simple key-value pairs for custom metadata (e.g., version, source_file)';
COMMENT ON COLUMN notes.custom_fields IS 'Complex structured data for extensibility (e.g., attachments, links)';

COMMENT ON COLUMN tags.custom_fields IS 'Additional structured data for the tag (e.g., URLs, related IDs)';
