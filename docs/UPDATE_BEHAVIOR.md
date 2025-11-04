# Update Behavior for Metadata and Custom Fields

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
