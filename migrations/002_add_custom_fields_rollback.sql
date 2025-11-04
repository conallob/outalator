-- Rollback migration for custom fields
-- This script reverses the changes made in 002_add_custom_fields.sql

-- Remove GIN indexes on JSONB columns
DROP INDEX IF EXISTS idx_tags_custom_fields;
DROP INDEX IF EXISTS idx_notes_custom_fields;
DROP INDEX IF EXISTS idx_notes_metadata;
DROP INDEX IF EXISTS idx_alerts_custom_fields;
DROP INDEX IF EXISTS idx_alerts_metadata;
DROP INDEX IF EXISTS idx_alerts_source_metadata;
DROP INDEX IF EXISTS idx_outages_custom_fields;
DROP INDEX IF EXISTS idx_outages_metadata;

-- Remove JSONB columns from tags table
ALTER TABLE tags DROP COLUMN IF EXISTS custom_fields;

-- Remove JSONB columns from notes table
ALTER TABLE notes
    DROP COLUMN IF EXISTS custom_fields,
    DROP COLUMN IF EXISTS metadata;

-- Remove JSONB columns from alerts table
ALTER TABLE alerts
    DROP COLUMN IF EXISTS custom_fields,
    DROP COLUMN IF EXISTS metadata,
    DROP COLUMN IF EXISTS source_metadata;

-- Remove JSONB columns from outages table
ALTER TABLE outages
    DROP COLUMN IF EXISTS custom_fields,
    DROP COLUMN IF EXISTS metadata;

-- Note: This rollback will permanently delete all custom metadata that was stored.
-- Make sure to backup your data before running this rollback script.
