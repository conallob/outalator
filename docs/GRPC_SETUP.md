# gRPC Setup Guide

This guide explains how to set up and use the gRPC API for Outalator.

## Overview

Outalator provides both REST and gRPC APIs that offer the same functionality. The gRPC API provides:
- Better performance with binary protocol
- Type safety through code generation
- Support for streaming (future enhancement)
- Automatic client generation for multiple languages

## Prerequisites

### 1. Install Protocol Buffers Compiler (protoc)

**macOS:**
```bash
brew install protobuf
```

**Ubuntu/Debian:**
```bash
sudo apt-get install -y protobuf-compiler
```

**Other platforms:**
Download from https://grpc.io/docs/protoc-installation/

Verify installation:
```bash
protoc --version
# Should show: libprotoc 3.x.x or higher
```

### 2. Install Go Plugins

```bash
make grpc-deps
```

This installs:
- `protoc-gen-go` - Generates Protocol Buffer messages
- `protoc-gen-go-grpc` - Generates gRPC service code

## Generating gRPC Code

After installing prerequisites, generate the Go code from proto files:

```bash
make proto
```

This creates:
- `api/proto/outalator.pb.go` - Protocol Buffer message definitions
- `api/proto/outalator_grpc.pb.go` - gRPC service stubs

## Configuration

### Enable gRPC in config.yaml

```yaml
grpc:
  enabled: true
  host: 0.0.0.0
  port: 9090
```

### Or use environment variables

```bash
export GRPC_ENABLED=true
export GRPC_HOST=0.0.0.0
export GRPC_PORT=9090
```

## Running the Server

Once configured and code is generated, the gRPC server will start automatically alongside the REST API:

```bash
go run cmd/outalator/main.go
```

You should see:
```
Starting HTTP server on 0.0.0.0:8080
Starting gRPC server on 0.0.0.0:9090
```

## Testing the gRPC API

### Using grpcurl

Install grpcurl:
```bash
# macOS
brew install grpcurl

# Or with Go
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

List available services:
```bash
grpcurl -plaintext localhost:9090 list
```

Describe a service:
```bash
grpcurl -plaintext localhost:9090 describe outalator.v1.OutageService
```

Create an outage:
```bash
grpcurl -plaintext -d '{
  "title": "API Service Down",
  "description": "Users unable to access API endpoints",
  "severity": "critical",
  "tags": [
    {"key": "service", "value": "api"},
    {"key": "region", "value": "us-west-2"}
  ]
}' localhost:9090 outalator.v1.OutageService/CreateOutage
```

List outages:
```bash
grpcurl -plaintext -d '{
  "limit": 10,
  "offset": 0
}' localhost:9090 outalator.v1.OutageService/ListOutages
```

### Using a Go Client

```go
package main

import (
    "context"
    "fmt"
    "log"

    pb "github.com/conall/outalator/api/proto/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    // Connect to server
    conn, err := grpc.Dial("localhost:9090",
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    // Create outage service client
    client := pb.NewOutageServiceClient(conn)

    // Create an outage
    resp, err := client.CreateOutage(context.Background(), &pb.CreateOutageRequest{
        Title:       "Database Performance Degradation",
        Description: "Query response times increased by 300%",
        Severity:    "high",
        Tags: []*pb.TagInput{
            {Key: "service", Value: "database"},
            {Key: "team", Value: "platform"},
        },
    })
    if err != nil {
        log.Fatalf("CreateOutage failed: %v", err)
    }

    fmt.Printf("Created outage: %s\n", resp.Outage.Id)

    // Add a note
    noteClient := pb.NewNoteServiceClient(conn)
    noteResp, err := noteClient.AddNote(context.Background(), &pb.AddNoteRequest{
        OutageId: resp.Outage.Id,
        Content:  "Restarted primary database instance",
        Format:   "plaintext",
        Author:   "sre@example.com",
    })
    if err != nil {
        log.Fatalf("AddNote failed: %v", err)
    }

    fmt.Printf("Added note: %s\n", noteResp.Note.Id)
}
```

## Implementation Steps

To complete the gRPC implementation, follow these steps:

### 1. Generate Proto Files

```bash
make proto
```

### 2. Complete the Converter Functions

Edit `internal/grpc/converters.go` and uncomment/complete the converter functions. These convert between domain models and protobuf messages.

Key functions to implement:
- `OutageDomainToProto(*domain.Outage) *pb.Outage`
- `AlertDomainToProto(*domain.Alert) *pb.Alert`
- `NoteDomainToProto(*domain.Note) *pb.Note`
- `TagDomainToProto(*domain.Tag) *pb.Tag`
- `CreateOutageRequestProtoToDomain(*pb.CreateOutageRequest) domain.CreateOutageRequest`

### 3. Complete the gRPC Server Implementation

Edit `internal/grpc/server.go`:

1. Uncomment the embedded interface fields
2. Add the proper proto import
3. Uncomment and complete all RPC method implementations
4. Implement proper error handling with gRPC status codes

### 4. Update Main Application

Edit `cmd/outalator/main.go` to:

1. Check if gRPC is enabled in config
2. Create a gRPC server
3. Register the gRPC services
4. Start the gRPC server in a goroutine alongside the HTTP server

Example:
```go
if cfg.GRPC.Enabled {
    lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port))
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    grpcServer := grpc.NewServer()
    grpcHandler := grpcpkg.NewServer(svc)
    grpcHandler.RegisterServices(grpcServer)

    go func() {
        log.Printf("Starting gRPC server on %s:%d", cfg.GRPC.Host, cfg.GRPC.Port)
        if err := grpcServer.Serve(lis); err != nil {
            log.Fatalf("Failed to serve gRPC: %v", err)
        }
    }()
}
```

### 5. Add Tests

Create `internal/grpc/server_test.go` with unit tests for each RPC method.

## API Services

The gRPC API provides five services:

### OutageService (port 9090)
- `CreateOutage` - Create new outage
- `GetOutage` - Get outage by ID
- `ListOutages` - List with pagination
- `UpdateOutage` - Partial updates
- `DeleteOutage` - Delete outage

### NoteService
- `AddNote` - Add troubleshooting note
- `GetNote` - Get note by ID
- `ListNotesByOutage` - List notes for outage
- `UpdateNote` - Update note content
- `DeleteNote` - Delete note

### TagService
- `AddTag` - Add metadata tag
- `GetTag` - Get tag by ID
- `ListTagsByOutage` - List tags for outage
- `DeleteTag` - Delete tag
- `SearchOutagesByTag` - Search outages by tag

### AlertService
- `ImportAlert` - Import from PagerDuty/OpsGenie
- `GetAlert` - Get alert by ID
- `GetAlertByExternalID` - Get by external ID
- `ListAlertsByOutage` - List alerts for outage
- `UpdateAlert` - Update alert

### HealthService
- `Check` - Health check

## Error Handling

gRPC uses status codes for errors:

| Code | Usage |
|------|-------|
| OK | Success |
| INVALID_ARGUMENT | Invalid UUID, missing required fields |
| NOT_FOUND | Resource not found |
| ALREADY_EXISTS | Duplicate resource |
| INTERNAL | Database errors, service failures |
| UNIMPLEMENTED | Not yet implemented |

Example:
```go
return nil, status.Errorf(codes.NotFound, "outage not found: %s", id)
```

## Client Libraries

gRPC supports code generation for multiple languages:

- **Go**: Already implemented
- **Python**: `python -m grpc_tools.protoc`
- **Java**: `protoc --java_out=.`
- **JavaScript/TypeScript**: `protoc --js_out=.`
- **C++, C#, Ruby, PHP**: See https://grpc.io/docs/

## Performance Considerations

gRPC offers several performance advantages:
- Binary protocol (protobuf) vs JSON
- HTTP/2 multiplexing
- Connection reuse
- Smaller payload size

Benchmarks typically show:
- 2-5x faster serialization
- 30-50% smaller payload
- Lower CPU usage

## Troubleshooting

### protoc not found
```bash
# Install protoc first
brew install protobuf  # macOS
sudo apt-get install protobuf-compiler  # Ubuntu
```

### Failed to generate code
```bash
# Install Go plugins
make grpc-deps

# Ensure $GOPATH/bin is in PATH
export PATH=$PATH:$(go env GOPATH)/bin
```

### Connection refused
```bash
# Verify gRPC is enabled
grep -A3 "^grpc:" config.yaml

# Check if server is listening
lsof -i :9090
```

### Import errors after generation
```bash
# Run go mod tidy
go mod tidy

# Rebuild
go build ./...
```

## Next Steps

1. Generate proto files: `make proto`
2. Complete converter implementations
3. Complete gRPC server implementations
4. Update main.go to start gRPC server
5. Test with grpcurl
6. Write integration tests
7. Add gRPC health checking
8. Consider adding TLS/authentication
9. Add metrics and monitoring
10. Document for API consumers

## Resources

- Protocol Buffers: https://protobuf.dev/
- gRPC Go: https://grpc.io/docs/languages/go/
- gRPC Basics: https://grpc.io/docs/what-is-grpc/introduction/
- Proto3 Spec: https://protobuf.dev/programming-guides/proto3/
