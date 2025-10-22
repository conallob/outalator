# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**outalator** is a tool for tracking troubleshooting notes for SREs (Site Reliability Engineers).

## Current Status

The project has been implemented with a complete backend infrastructure:
- RESTful API server built with Go
- PostgreSQL database with migration scripts
- Support for PagerDuty and OpsGenie integrations
- Modular architecture for easy extension

## Technology Stack

- **Language**: Go 1.21+
- **Web Framework**: Gorilla Mux
- **Database**: PostgreSQL with lib/pq driver
- **Configuration**: YAML-based with environment variable overrides
- **External Integrations**: PagerDuty, OpsGenie (extensible)

## Development Workflow

1. Copy `config.example.yaml` to `config.yaml` and configure your settings
2. Set up PostgreSQL database
3. Run database migrations from `migrations/001_initial_schema.sql`
4. Build and run the application: `go run cmd/outalator/main.go`
5. Access API at `http://localhost:8080`

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
  ├── api/              - HTTP handlers and routes
  └── config/           - Configuration management
migrations/             - Database migration scripts
```

### Key Design Principles

1. **Modularity**: Storage and notification services use interface-based design
2. **Extensibility**: Easy to add new notification services or storage backends
3. **Clean Architecture**: Clear separation between layers (domain, service, api, storage)
4. **Configuration**: Flexible YAML + environment variable configuration
