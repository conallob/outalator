# gRPC Quick Start

## TL;DR

```bash
# 1. Install protoc
brew install protobuf  # macOS
# OR
sudo apt-get install protobuf-compiler  # Ubuntu

# 2. Install Go plugins
make grpc-deps

# 3. Generate code
make proto

# 4. Enable gRPC in config.yaml
# grpc:
#   enabled: true

# 5. Complete implementation
# - internal/grpc/converters.go (implement converter functions)
# - internal/grpc/server.go (implement RPC methods)
# - cmd/outalator/main.go (start gRPC server)

# 6. Test
grpcurl -plaintext localhost:9090 list
```

## Quick Commands

| Task | Command |
|------|---------|
| Install tools | `make grpc-deps` |
| Generate proto | `make proto` |
| Build | `make build` |
| Run | `make run` |
| Clean | `make clean` |

## Configuration

**config.yaml:**
```yaml
grpc:
  enabled: true
  host: 0.0.0.0
  port: 9090
```

**Environment:**
```bash
export GRPC_ENABLED=true
export GRPC_PORT=9090
```

## Testing

**With grpcurl:**
```bash
# Health check
grpcurl -plaintext localhost:9090 outalator.v1.HealthService/Check

# Create outage
grpcurl -plaintext -d '{"title":"Test","severity":"high"}' \
  localhost:9090 outalator.v1.OutageService/CreateOutage
```

**With Go:**
```go
conn, _ := grpc.Dial("localhost:9090", grpc.WithInsecure())
client := pb.NewOutageServiceClient(conn)
resp, _ := client.CreateOutage(ctx, &pb.CreateOutageRequest{...})
```

## Documentation

- **Full Setup**: `docs/GRPC_SETUP.md`
- **API Reference**: `api/proto/README.md`
- **Implementation Summary**: `GRPC_IMPLEMENTATION_SUMMARY.md`
- **Proto Schema**: `api/proto/outalator.proto`

## Services

| Service | Methods | Port |
|---------|---------|------|
| OutageService | Create, Get, List, Update, Delete | 9090 |
| NoteService | Add, Get, List, Update, Delete | 9090 |
| TagService | Add, Get, List, Delete, Search | 9090 |
| AlertService | Import, Get, GetByExternal, List, Update | 9090 |
| HealthService | Check | 9090 |

## Implementation Status

- ✅ Proto definitions
- ✅ Build infrastructure
- ✅ Configuration support
- ✅ Documentation
- ⏳ Code generation (need protoc)
- ⏳ Converter implementation
- ⏳ Server implementation
- ⏳ Main app integration

## Need Help?

1. Check `docs/GRPC_SETUP.md` for detailed setup
2. Read `api/proto/README.md` for API usage
3. See `GRPC_IMPLEMENTATION_SUMMARY.md` for implementation details
