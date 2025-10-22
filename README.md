# Outalator

A modern web application for tracking outage troubleshooting notes for Site Reliability Engineers (SREs), inspired by [Outalator](https://sre.google/sre-book/tracking-outages/) from the Google SRE book.

## Features

- **Outage Tracking**: Create and manage outages with comprehensive details
- **Multi-Service Alerts**: Import alerts from multiple oncall notification services
  - PagerDuty
  - OpsGenie
  - Extensible architecture for additional services
- **Note-Taking**: Add plaintext or markdown notes to outages
- **Tagging System**: Organize outages with flexible key-value tags (e.g., Jira tickets, services, regions)
- **Modular Storage**: Interface-based storage layer with PostgreSQL implementation
- **RESTful API**: Clean HTTP API for all operations

## Architecture

The application is built with Go and follows clean architecture principles:

- **Domain Layer**: Core business entities (Outage, Alert, Note, Tag)
- **Storage Layer**: Pluggable storage interface with PostgreSQL implementation
- **Notification Layer**: Extensible notification service integrations
- **Service Layer**: Business logic and orchestration
- **API Layer**: HTTP handlers and routing

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- (Optional) PagerDuty API key
- (Optional) OpsGenie API key

## Installation

1. Clone the repository:
```bash
git clone https://github.com/conall/outalator.git
cd outalator
```

2. Install Go dependencies:
```bash
go mod download
```

3. Set up PostgreSQL:
```bash
# Create database and user
createdb outalator
createuser outalator

# Run migrations
psql -U outalator -d outalator -f migrations/001_initial_schema.sql
```

4. Configure the application:
```bash
cp config.example.yaml config.yaml
# Edit config.yaml with your settings
```

5. Build and run:
```bash
go build -o outalator cmd/outalator/main.go
./outalator -config config.yaml
```

Or run directly:
```bash
go run cmd/outalator/main.go -config config.yaml
```

## Configuration

Configuration can be provided via YAML file and/or environment variables.

### Configuration File (config.yaml)

```yaml
server:
  host: 0.0.0.0
  port: 8080

database:
  host: localhost
  port: 5432
  user: outalator
  password: outalator
  dbname: outalator
  sslmode: disable

# Optional: PagerDuty integration
pagerduty:
  api_key: your-pagerduty-api-key

# Optional: OpsGenie integration
opsgenie:
  api_key: your-opsgenie-api-key
```

### Environment Variables

Environment variables override config file values:

- `SERVER_HOST` - Server host
- `SERVER_PORT` - Server port
- `DB_HOST` - Database host
- `DB_PORT` - Database port
- `DB_USER` - Database user
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name
- `PAGERDUTY_API_KEY` - PagerDuty API key
- `OPSGENIE_API_KEY` - OpsGenie API key

## API Documentation

### Outages

#### Create Outage
```bash
POST /api/v1/outages
Content-Type: application/json

{
  "title": "API Gateway Outage",
  "description": "Users unable to access API endpoints",
  "severity": "high",
  "alert_ids": ["PXYZ123"],
  "tags": [
    {"key": "jira", "value": "OPS-1234"},
    {"key": "service", "value": "api-gateway"}
  ]
}
```

#### List Outages
```bash
GET /api/v1/outages?limit=50&offset=0
```

#### Get Outage
```bash
GET /api/v1/outages/{id}
```

#### Update Outage
```bash
PATCH /api/v1/outages/{id}
Content-Type: application/json

{
  "status": "resolved",
  "severity": "medium"
}
```

### Notes

#### Add Note to Outage
```bash
POST /api/v1/outages/{id}/notes
Content-Type: application/json

{
  "content": "Identified root cause: database connection pool exhaustion",
  "format": "markdown",
  "author": "alice@example.com"
}
```

### Tags

#### Add Tag to Outage
```bash
POST /api/v1/outages/{id}/tags
Content-Type: application/json

{
  "key": "jira",
  "value": "OPS-5678"
}
```

#### Search Outages by Tag
```bash
GET /api/v1/tags/search?key=jira&value=OPS-1234
```

### Alerts

#### Import Alert
```bash
POST /api/v1/alerts/import
Content-Type: application/json

{
  "source": "pagerduty",
  "external_id": "PXYZ123",
  "outage_id": "optional-uuid-to-associate"
}
```

### Health Check

```bash
GET /health
```

## Authentication

Outalator supports OIDC authentication with providers like Okta, Auth0, Google, etc. When authentication is enabled, all notes are automatically tagged with the authenticated user's email address.

### Configuring Authentication

Add to your `config.yaml`:

```yaml
auth:
  enabled: true
  issuer: https://your-company.okta.com
  client_id: your-okta-client-id
  client_secret: your-okta-client-secret
  redirect_url: http://localhost:8080/auth/callback
  session_key: generate-a-random-32-byte-base64-key
```

Generate a session key:
```bash
openssl rand -base64 32
```

When authentication is disabled, the application runs without authentication (useful for development).

## Kubernetes Deployment

Outalator can be deployed to Kubernetes using either Helm or Kustomize.

### Quick Start with Helm

```bash
# Create namespace
kubectl create namespace outalator

# Install with Helm
helm install outalator ./helm/outalator \
  --namespace outalator \
  --set ingress.hosts[0].host=outalator.example.com \
  --set config.auth.issuer=https://your-company.okta.com \
  --set secrets.auth.clientId=your-client-id \
  --set secrets.auth.clientSecret=your-client-secret
```

### Quick Start with Kustomize

```bash
# Update configuration in k8s/base/configmap.yaml and k8s/base/secret.yaml
# Then apply
kubectl apply -k k8s/overlays/prod
```

For complete Kubernetes deployment documentation including:
- Detailed Helm configuration
- Kustomize overlays
- Okta/OIDC setup
- Security best practices
- Monitoring and troubleshooting

See [docs/KUBERNETES.md](docs/KUBERNETES.md)

### Building the Container Image

```bash
docker build -t outalator:latest .
docker tag outalator:latest your-registry/outalator:v1.0.0
docker push your-registry/outalator:v1.0.0
```

## Database Schema

The application uses PostgreSQL with the following tables:

- **outages**: Main outage tracking table
- **alerts**: Imported alerts from notification services
- **notes**: Troubleshooting notes attached to outages (with author attribution)
- **tags**: Key-value metadata tags

See `migrations/001_initial_schema.sql` for the complete schema.

## Extending the Application

### Adding a New Notification Service

1. Create a new package under `internal/notification/`
2. Implement the `notification.Service` interface:
   - `Name() string`
   - `FetchAlert(ctx, alertID) (*Alert, error)`
   - `FetchRecentAlerts(ctx, since) ([]*Alert, error)`
   - `WebhookHandler() interface{}`

3. Register the service in `cmd/outalator/main.go`

### Adding a New Storage Backend

1. Create a new package under `internal/storage/`
2. Implement the `storage.Storage` interface
3. Update `cmd/outalator/main.go` to use the new storage

## Development

### Running Tests

```bash
go test ./...
```

### Project Structure

```
outalator/
├── cmd/
│   └── outalator/          # Main application entry point
├── internal/
│   ├── api/                # HTTP handlers and routes
│   ├── config/             # Configuration management
│   ├── domain/             # Domain models and DTOs
│   ├── notification/       # Notification service integrations
│   │   ├── opsgenie/
│   │   └── pagerduty/
│   ├── service/            # Business logic
│   └── storage/            # Storage layer
│       └── postgres/       # PostgreSQL implementation
├── migrations/             # Database migration scripts
├── config.example.yaml     # Example configuration file
├── go.mod                  # Go module definition
└── README.md              # This file
```

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

MIT License - See LICENSE file for details

## References

- [Google SRE Book - Tracking Outages](https://sre.google/sre-book/tracking-outages/)
- [PagerDuty API Documentation](https://developer.pagerduty.com/api-reference/)
- [OpsGenie API Documentation](https://docs.opsgenie.com/docs/api-overview)
