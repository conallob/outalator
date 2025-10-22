# Outalator gRPC API

This directory contains the Protocol Buffer definitions for the Outalator gRPC API.

## Overview

The Outalator API is available in two formats:
- **REST API**: HTTP/JSON API (port 8080 by default)
- **gRPC API**: High-performance gRPC API (port 9090 by default)

Both APIs provide the same functionality and can be used interchangeably.

## Prerequisites

To generate Go code from the proto files, you need:

1. **Protocol Buffers Compiler (protoc)**
   ```bash
   # macOS
   brew install protobuf

   # Ubuntu/Debian
   sudo apt-get install -y protobuf-compiler

   # Or download from: https://grpc.io/docs/protoc-installation/
   ```

2. **Go plugins for protoc**
   ```bash
   make grpc-deps
   ```

## Generating Code

After installing prerequisites, generate the Go code:

```bash
make proto
```

This will create two files:
- `api/proto/outalator.pb.go` - Protocol Buffer message definitions
- `api/proto/outalator_grpc.pb.go` - gRPC service definitions

## API Structure

The gRPC API is organized into five services:

### OutageService
Manages outages (incidents):
- `CreateOutage` - Create a new outage
- `GetOutage` - Get an outage by ID
- `ListOutages` - List outages with pagination
- `UpdateOutage` - Update an outage (partial updates supported)
- `DeleteOutage` - Delete an outage

### NoteService
Manages troubleshooting notes:
- `AddNote` - Add a note to an outage
- `GetNote` - Get a note by ID
- `ListNotesByOutage` - List all notes for an outage
- `UpdateNote` - Update a note
- `DeleteNote` - Delete a note

### TagService
Manages metadata tags:
- `AddTag` - Add a tag to an outage
- `GetTag` - Get a tag by ID
- `ListTagsByOutage` - List all tags for an outage
- `DeleteTag` - Delete a tag
- `SearchOutagesByTag` - Find outages by tag key/value

### AlertService
Manages alerts from external paging systems (PagerDuty, OpsGenie):
- `ImportAlert` - Import an alert from external service
- `GetAlert` - Get an alert by ID
- `GetAlertByExternalID` - Get alert by external ID
- `ListAlertsByOutage` - List all alerts for an outage
- `UpdateAlert` - Update an alert

### HealthService
Health check:
- `Check` - Health check endpoint

## Data Models

### Outage
Represents an ongoing or historical incident:
- `id` (string/UUID) - Unique identifier
- `title` (string) - Outage title
- `description` (string) - Detailed description
- `status` (string) - Status: "open", "investigating", "resolved", "closed"
- `severity` (string) - Severity: "critical", "high", "medium", "low"
- `created_at` (timestamp) - Creation time
- `updated_at` (timestamp) - Last update time
- `resolved_at` (timestamp, optional) - Resolution time
- `alerts` (repeated Alert) - Associated alerts
- `notes` (repeated Note) - Associated notes
- `tags` (repeated Tag) - Associated tags

### Alert
Paging alert from external services:
- `id` (string/UUID) - Internal identifier
- `outage_id` (string/UUID) - Parent outage
- `external_id` (string) - Alert ID from external service
- `source` (string) - Service: "pagerduty", "opsgenie"
- `team_name` (string) - Team receiving the alert
- `title` (string) - Alert title
- `description` (string) - Alert description
- `severity` (string) - Alert severity
- `triggered_at` (timestamp) - When alert triggered
- `acknowledged_at` (timestamp, optional) - Acknowledgment time
- `resolved_at` (timestamp, optional) - Resolution time
- `created_at` (timestamp) - Creation time in system

### Note
Troubleshooting note:
- `id` (string/UUID) - Unique identifier
- `outage_id` (string/UUID) - Parent outage
- `content` (string) - Note content
- `format` (string) - Format: "plaintext", "markdown"
- `author` (string) - Author email
- `created_at` (timestamp) - Creation time
- `updated_at` (timestamp) - Last update time

### Tag
Metadata tag:
- `id` (string/UUID) - Unique identifier
- `outage_id` (string/UUID) - Parent outage
- `key` (string) - Tag key (e.g., "jira", "service", "region")
- `value` (string) - Tag value (e.g., "PROJ-123", "api")
- `created_at` (timestamp) - Creation time

## Example Usage

### Using grpcurl (CLI tool)

```bash
# List services
grpcurl -plaintext localhost:9090 list

# List methods for OutageService
grpcurl -plaintext localhost:9090 list outalator.v1.OutageService

# Create an outage
grpcurl -plaintext -d '{
  "title": "API Service Down",
  "description": "Users unable to access API endpoints",
  "severity": "critical",
  "tags": [
    {"key": "service", "value": "api"},
    {"key": "region", "value": "us-west-2"}
  ]
}' localhost:9090 outalator.v1.OutageService/CreateOutage

# Get an outage
grpcurl -plaintext -d '{
  "id": "550e8400-e29b-41d4-a716-446655440000"
}' localhost:9090 outalator.v1.OutageService/GetOutage

# List outages
grpcurl -plaintext -d '{
  "limit": 10,
  "offset": 0
}' localhost:9090 outalator.v1.OutageService/ListOutages

# Add a note
grpcurl -plaintext -d '{
  "outage_id": "550e8400-e29b-41d4-a716-446655440000",
  "content": "Restarted service, investigating root cause",
  "format": "plaintext",
  "author": "sre@example.com"
}' localhost:9090 outalator.v1.NoteService/AddNote

# Search by tag
grpcurl -plaintext -d '{
  "key": "service",
  "value": "api"
}' localhost:9090 outalator.v1.TagService/SearchOutagesByTag

# Health check
grpcurl -plaintext localhost:9090 outalator.v1.HealthService/Check
```

### Using Go client

```go
import (
    pb "github.com/conall/outalator/api/proto/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

// Connect to server
conn, err := grpc.Dial("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

// Create client
client := pb.NewOutageServiceClient(conn)

// Create an outage
resp, err := client.CreateOutage(context.Background(), &pb.CreateOutageRequest{
    Title:       "API Service Down",
    Description: "Users unable to access API endpoints",
    Severity:    "critical",
    Tags: []*pb.TagInput{
        {Key: "service", Value: "api"},
        {Key: "region", Value: "us-west-2"},
    },
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Created outage: %s\n", resp.Outage.Id)
```

## Protocol Buffer Schema

The complete schema is defined in `outalator.proto`. Key features:

- Uses `proto3` syntax
- All timestamps use `google.protobuf.Timestamp`
- Optional fields use the `optional` keyword
- UUIDs are represented as strings
- Supports pagination with limit/offset
- All operations return structured responses

## Testing

You can test the gRPC API using:

1. **grpcurl** - Command-line tool for gRPC
   ```bash
   brew install grpcurl  # macOS
   ```

2. **BloomRPC** - GUI client for gRPC (like Postman)
   Download from: https://github.com/bloomrpc/bloomrpc

3. **Postman** - Now supports gRPC
   Import the proto file directly

## Development

When modifying the API:

1. Edit `outalator.proto`
2. Regenerate code: `make proto`
3. Update the gRPC server implementation in `internal/grpc/`
4. Update converter functions in `internal/grpc/converters.go`
5. Test with `grpcurl` or a client

## Migration from REST

The gRPC API is designed to be functionally equivalent to the REST API:

| REST Endpoint | gRPC Method |
|--------------|-------------|
| POST /api/v1/outages | OutageService.CreateOutage |
| GET /api/v1/outages | OutageService.ListOutages |
| GET /api/v1/outages/{id} | OutageService.GetOutage |
| PATCH /api/v1/outages/{id} | OutageService.UpdateOutage |
| POST /api/v1/outages/{id}/notes | NoteService.AddNote |
| POST /api/v1/outages/{id}/tags | TagService.AddTag |
| GET /api/v1/tags/search | TagService.SearchOutagesByTag |
| POST /api/v1/alerts/import | AlertService.ImportAlert |
| GET /health | HealthService.Check |

## Benefits of gRPC

- **Performance**: Binary protocol, more efficient than JSON
- **Type Safety**: Strong typing with code generation
- **Streaming**: Support for bi-directional streaming (future enhancement)
- **HTTP/2**: Multiplexing, server push
- **Code Generation**: Automatic client libraries for many languages
- **Contract-First**: API definition separate from implementation
