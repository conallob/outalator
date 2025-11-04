# Custom Fields Refactoring

## Overview

This document describes the refactoring of the Outalator protobuf data structure to support flexible data structure customization. The changes enable users to add custom metadata and structured fields to all core entities without modifying the schema.

## Changes Made

### 1. Protocol Buffer Definitions (`api/proto/outalator.proto`)

#### Added Import
```protobuf
import "google/protobuf/struct.proto";
```

#### New Messages for Source-Specific Alert Metadata

- **PagerDutyMetadata**: Contains PagerDuty-specific alert information (incident_key, service_id, escalation_policy, etc.)
- **OpsGenieMetadata**: Contains OpsGenie-specific alert information (alias, responders, priority, owner, etc.)
- **GenericMetadata**: For other alert sources with flexible properties

#### Updated Core Messages

**Outage**:
- Added `map<string, string> metadata` - Simple key-value pairs
- Added `google.protobuf.Struct custom_fields` - Complex structured data

**Alert**:
- Added `oneof source_metadata` - Source-specific data (PagerDuty, OpsGenie, Generic)
- Added `map<string, string> metadata` - Simple key-value pairs
- Added `google.protobuf.Struct custom_fields` - Complex structured data

**Note**:
- Added `map<string, string> metadata` - Simple key-value pairs (e.g., version, source_file)
- Added `google.protobuf.Struct custom_fields` - Complex structured data (e.g., attachments, links)

**Tag**:
- Added `google.protobuf.Struct custom_fields` - Additional structured data (e.g., URLs, related IDs)

#### Updated Request Messages

All create and update request messages now support metadata and custom_fields:
- `CreateOutageRequest`
- `UpdateOutageRequest`
- `AddNoteRequest`
- `UpdateNoteRequest`
- `AddTagRequest`
- `ImportAlertRequest`
- `UpdateAlertRequest`
- `TagInput`

### 2. Domain Models (`internal/domain/models.go`)

Updated all core domain models to include custom fields:

**Outage**:
```go
Metadata     map[string]string `json:"metadata,omitempty"`
CustomFields map[string]any    `json:"custom_fields,omitempty"`
```

**Alert**:
```go
SourceMetadata   map[string]any    `json:"source_metadata,omitempty"`
Metadata         map[string]string `json:"metadata,omitempty"`
CustomFields     map[string]any    `json:"custom_fields,omitempty"`
```

**Note**:
```go
Metadata     map[string]string `json:"metadata,omitempty"`
CustomFields map[string]any    `json:"custom_fields,omitempty"`
```

**Tag**:
```go
CustomFields map[string]any `json:"custom_fields,omitempty"`
```

### 3. Database Schema (`migrations/002_add_custom_fields.sql`)

Added JSONB columns to all tables to store flexible structured data:

**outages table**:
- `metadata JSONB` - Simple key-value pairs
- `custom_fields JSONB` - Complex structured data

**alerts table**:
- `source_metadata JSONB` - Source-specific data
- `metadata JSONB` - Simple key-value pairs
- `custom_fields JSONB` - Complex structured data

**notes table**:
- `metadata JSONB` - Simple key-value pairs
- `custom_fields JSONB` - Complex structured data

**tags table**:
- `custom_fields JSONB` - Additional structured data

GIN indexes were added on all JSONB columns for efficient querying:
```sql
CREATE INDEX idx_outages_metadata ON outages USING GIN (metadata);
CREATE INDEX idx_alerts_source_metadata ON alerts USING GIN (source_metadata);
-- etc.
```

### 4. Storage Layer (`internal/storage/postgres/`)

Updated all CRUD operations in:
- `outage.go` - Now handles metadata and custom_fields JSON marshaling/unmarshaling
- `alert.go` - Now handles source_metadata, metadata, and custom_fields
- Note: Similar updates needed for `note.go` and `tag.go`

All INSERT, SELECT, and UPDATE queries now include the new JSONB columns.

## Benefits

1. **Extensibility**: Add custom data without schema changes
2. **Flexibility**: Support different alert sources with source-specific metadata
3. **Backward Compatibility**: All new fields are optional (JSONB defaults to '{}')
4. **Performance**: GIN indexes enable efficient querying of JSON data
5. **Type Safety**: Protocol Buffers provide strong typing for gRPC APIs
6. **Queryability**: JSONB in PostgreSQL allows SQL queries on custom data

## Usage Examples

### Example 1: Adding Custom Metadata to an Outage

```go
outage := &domain.Outage{
    Title:    "Production API Outage",
    Severity: "critical",
    Metadata: map[string]string{
        "region":       "us-east-1",
        "service":      "api-gateway",
        "incident_id":  "INC-12345",
    },
    CustomFields: map[string]any{
        "affected_endpoints": []string{"/api/v1/users", "/api/v1/orders"},
        "estimated_impact":   map[string]any{
            "users":    10000,
            "revenue":  25000.50,
        },
    },
}
```

### Example 2: PagerDuty-Specific Alert Data

```protobuf
message ImportAlertRequest {
    source = "pagerduty"
    external_id = "PD12345"
    pagerduty = {
        incident_key: "my-service/error-rate"
        service_id: "PABCDEF"
        service_name: "API Service"
        escalation_policy: "Engineering Team"
        urgency: "high"
    }
}
```

### Example 3: Querying Custom Fields in PostgreSQL

```sql
-- Find outages with a specific metadata key
SELECT * FROM outages WHERE metadata ? 'incident_id';

-- Find outages in a specific region
SELECT * FROM outages WHERE metadata->>'region' = 'us-east-1';

-- Find alerts with high PagerDuty urgency
SELECT * FROM alerts WHERE source_metadata->>'urgency' = 'high';
```

## Migration Path

1. Run the new migration: `migrations/002_add_custom_fields.sql`
2. Regenerate protobuf code: `make proto` (after installing protoc and dependencies)
3. Update service layer to handle new fields
4. Update gRPC converters to map between protobuf and domain models
5. Test with existing data (backward compatible - new fields default to empty)

## Next Steps

1. Complete updates to `note.go` and `tag.go` storage implementations
2. Update gRPC converters in `internal/grpc/converters.go`
3. Update service layer to propagate custom fields
4. Add validation for custom field structure (optional)
5. Add API examples and documentation
6. Consider adding custom field schemas/templates

## Notes

- All custom fields are optional and backward compatible
- Empty metadata/custom_fields are stored as '{}' in JSONB
- GIN indexes enable efficient queries on custom fields
- Use `google.protobuf.Struct` for complex nested JSON in protobufs
- Use `map<string, any>` for flexible JSON in Go domain models
