# Update Behavior for Metadata and Custom Fields

## 📌 Quick Summary

**⚠️ IMPORTANT: All update operations use FULL REPLACEMENT for `metadata` and `custom_fields`.**

When you update these fields:
- ✅ The **entire** field is replaced with your new value
- ⚠️ Keys not included in your update will be **permanently deleted**
- ✅ This behavior is consistent across REST API, gRPC API, and all entity types
- ✅ To preserve existing keys, fetch the entity first and merge locally

## Affected Entities

This behavior applies to:
- **Outages**: `metadata`, `custom_fields`
- **Alerts**: `metadata`, `custom_fields`, `source_metadata`
- **Notes**: `metadata`, `custom_fields`
- **Tags**: `custom_fields`

## Affected Operations

- REST: `PUT /outages/:id`, `PUT /notes/:id`
- gRPC: `UpdateOutage`, `UpdateNote`, `UpdateAlert`
- Service Layer: `UpdateOutage()`, `UpdateNote()`

## Overview

This document describes how metadata and custom fields are handled during update operations in the Outalator application.

## Current Behavior: Full Replacement

**All update operations currently use FULL REPLACEMENT for metadata and custom fields.**

This means:
- When you update an entity's metadata or custom_fields, the **entire** field is replaced
- Existing keys that are not in the update request will be **removed**
- This is the simplest and most predictable behavior

### Example: Full Replacement

**Initial State:**
```json
{
  "id": "123",
  "title": "Production API Outage",
  "metadata": {
    "region": "us-east-1",
    "service": "api-gateway"
  }
}
```

**Update Request:**
```json
{
  "id": "123",
  "metadata": {
    "region": "us-west-2"
  }
}
```

**Result (Full Replacement):**
```json
{
  "id": "123",
  "title": "Production API Outage",
  "metadata": {
    "region": "us-west-2"
    // NOTE: "service" key is REMOVED
  }
}
```

## API Examples

### REST API Example

```bash
# Initial state
GET /api/v1/outages/123
{
  "id": "123",
  "title": "API Outage",
  "metadata": {
    "region": "us-east-1",
    "service": "api-gateway",
    "environment": "production"
  },
  "custom_fields": {
    "runbook_url": "https://wiki.example.com/runbook",
    "slack_channel": "#incidents"
  }
}

# Update with partial metadata (THIS WILL DELETE MISSING KEYS!)
PUT /api/v1/outages/123
{
  "metadata": {
    "region": "us-west-2"
  }
}

# Result - "service" and "environment" are DELETED
{
  "id": "123",
  "title": "API Outage",
  "metadata": {
    "region": "us-west-2"
    // ⚠️ "service" and "environment" keys are GONE
  },
  "custom_fields": {
    "runbook_url": "https://wiki.example.com/runbook",
    "slack_channel": "#incidents"
    // ✅ custom_fields unchanged because not in update request
  }
}
```

### gRPC API Example

```protobuf
// Initial state
GetOutage("123") returns:
{
  id: "123"
  metadata: {
    "region": "us-east-1"
    "service": "api-gateway"
  }
}

// Update request
UpdateOutageRequest {
  id: "123"
  metadata: {
    "region": "us-west-2"
  }
}

// Result - "service" key is DELETED
{
  id: "123"
  metadata: {
    "region": "us-west-2"
    // ⚠️ "service" key is GONE
  }
}
```

### When Fields Are Not Touched

If you don't include `metadata` or `custom_fields` in your update request, they are **preserved**:

```bash
# Update only the title
PUT /api/v1/outages/123
{
  "title": "Updated Title"
}

# Result - metadata and custom_fields are unchanged
{
  "id": "123",
  "title": "Updated Title",
  "metadata": {
    "region": "us-east-1",
    "service": "api-gateway"
    // ✅ All keys preserved
  }
}
```

**Rule:** If you **omit** `metadata`/`custom_fields` from the update → they are **preserved**
**Rule:** If you **include** `metadata`/`custom_fields` in the update → they are **fully replaced**

## Workaround: Client-Side Merge

If you need to merge metadata or custom fields, the client must:
1. Fetch the current entity
2. Merge the fields locally
3. Send the complete merged data in the update request

### Example: Client-Side Merge

```go
// 1. Fetch current outage
outage, err := service.GetOutage(ctx, outageID)
if err != nil {
    return err
}

// 2. Merge metadata locally
if outage.Metadata == nil {
    outage.Metadata = make(map[string]string)
}
outage.Metadata["region"] = "us-west-2"  // Update existing key
outage.Metadata["incident_id"] = "INC-456"  // Add new key
// "service" key is preserved

// 3. Update with complete merged data
req := domain.UpdateOutageRequest{
    Metadata: outage.Metadata,
}
updated, err := service.UpdateOutage(ctx, outageID, req)
```

## Future Enhancement: Merge Support

### Planned API for Merge Operations

A future version may support explicit merge operations:

```protobuf
message UpdateOutageRequest {
  string id = 1;
  optional string title = 2;

  // Option 1: Separate merge fields
  map<string, string> metadata_merge = 10;  // Keys to add/update
  repeated string metadata_remove = 11;      // Keys to remove

  // Option 2: Merge behavior flag
  map<string, string> metadata = 6;
  bool metadata_merge_mode = 12;  // If true, merge; if false, replace
}
```

### Service Layer Helper for Merging (Placeholder)

```go
// mergeStringMaps merges updates into base, returning a new map
func mergeStringMaps(base, updates map[string]string) map[string]string {
    if base == nil {
        base = make(map[string]string)
    }
    result := make(map[string]string)

    // Copy base
    for k, v := range base {
        result[k] = v
    }

    // Apply updates
    for k, v := range updates {
        result[k] = v
    }

    return result
}

// mergeAnyMaps merges updates into base for map[string]any
func mergeAnyMaps(base, updates map[string]any) map[string]any {
    if base == nil {
        base = make(map[string]any)
    }
    result := make(map[string]any)

    // Copy base
    for k, v := range base {
        result[k] = v
    }

    // Apply updates
    for k, v := range updates {
        result[k] = v
    }

    return result
}
```

## Best Practices

### 1. Document Expected Behavior

Always document in API responses and client code whether you're using replacement or merge semantics.

### 2. Use Atomic Updates

For concurrent updates, consider using optimistic locking:
```go
type Outage struct {
    ID       uuid.UUID
    Version  int64  // Increment on each update
    // ... other fields
}
```

### 3. Validate Before Update

Always validate metadata and custom fields before updates:
```go
if len(metadata) > 100 {
    return errors.New("metadata cannot have more than 100 keys")
}

for k, v := range metadata {
    if len(k) > 255 || len(v) > 1024 {
        return errors.New("metadata key or value too long")
    }
}
```

### 4. Audit Changes

Log metadata changes for debugging and compliance:
```go
log.Printf("Outage %s metadata changed from %v to %v",
    outageID, oldMetadata, newMetadata)
```

## Related Documentation

- [Custom Fields Refactoring](CUSTOM_FIELDS_REFACTORING.md) - Overview of custom fields architecture
- [Security Considerations](SECURITY.md) - Security guidelines for custom data
- API Documentation - Full API reference (coming soon)
