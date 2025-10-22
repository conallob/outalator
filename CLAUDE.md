# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**outalator** is a tool for tracking troubleshooting notes for SREs (Site Reliability Engineers).

## Current Status

The project has been implemented with a complete backend infrastructure:
- RESTful API server built with Go
- gRPC API server (protobuf definitions ready, implementation pending proto generation)
- PostgreSQL database with migration scripts
- Support for PagerDuty and OpsGenie integrations
- Modular architecture for easy extension

## Technology Stack

- **Language**: Go 1.19+ (currently 1.19 for compatibility)
- **Web Framework**: Gorilla Mux (REST API)
- **RPC Framework**: gRPC with Protocol Buffers
- **Database**: PostgreSQL with lib/pq driver
- **Configuration**: YAML-based with environment variable overrides
- **External Integrations**: PagerDuty, OpsGenie (extensible)

## Development Workflow

1. Copy `config.example.yaml` to `config.yaml` and configure your settings
2. Set up PostgreSQL database
3. Run database migrations from `migrations/001_initial_schema.sql`
4. (Optional) Generate gRPC code: `make proto` (requires protoc installation)
5. Build and run the application: `go run cmd/outalator/main.go`
6. Access REST API at `http://localhost:8080`
7. (Optional) Access gRPC API at `localhost:9090` if enabled

## Architecture

The application follows a layered architecture:

```
cmd/outalator/          - Application entry point
internal/
  ├── domain/           - Core domain models (Outage, Alert, Note, Tag)
  ├── storage/          - Storage interface and implementations
  │   └── postgres/     - PostgreSQL implementation
  ├── notification/     - Notification service interface
  │   ├── pagerduty/    - PagerDuty integration
  │   └── opsgenie/     - OpsGenie integration
  ├── service/          - Business logic layer
  ├── api/              - HTTP handlers and routes (REST)
  ├── grpc/             - gRPC handlers and converters
  └── config/           - Configuration management
api/proto/              - Protocol Buffer definitions
migrations/             - Database migration scripts
scripts/                - Build and generation scripts
```

### Key Design Principles

1. **Modularity**: Storage and notification services use interface-based design
2. **Extensibility**: Easy to add new notification services or storage backends
3. **Clean Architecture**: Clear separation between layers (domain, service, api, storage)
4. **Configuration**: Flexible YAML + environment variable configuration
5. **API Flexibility**: Both REST and gRPC APIs supported, functionally equivalent

## gRPC Support

The project includes comprehensive gRPC support:
- **Proto Definitions**: `api/proto/outalator.proto` defines all messages and services
- **Generated Code**: Run `make proto` to generate Go code (requires protoc)
- **Setup Guide**: See `docs/GRPC_SETUP.md` for detailed instructions
- **Configuration**: Enable via `grpc.enabled: true` in config.yaml or `GRPC_ENABLED=true` env var

### gRPC Services
- **OutageService**: Manage outages (create, get, list, update, delete)
- **NoteService**: Manage troubleshooting notes
- **TagService**: Manage metadata tags and search
- **AlertService**: Import and manage alerts from external services
- **HealthService**: Health check endpoint

### Proto Generation Requirements
1. Install protoc: `brew install protobuf` (macOS) or `apt-get install protobuf-compiler` (Ubuntu)
2. Install Go plugins: `make grpc-deps`
3. Generate code: `make proto`

See `docs/GRPC_SETUP.md` for complete setup instructions.
