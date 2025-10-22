# Outalator Deployment Summary

This document provides a quick reference for deploying Outalator in various environments.

## Authentication & User Attribution

Outalator now includes **OIDC authentication** with automatic user attribution for notes:

- ✅ OIDC/OAuth2 support (Okta, Auth0, Google, Azure AD, etc.)
- ✅ All notes automatically tagged with authenticated user's email
- ✅ Session-based authentication with secure cookies
- ✅ Optional authentication (can be disabled for development)

### Key Authentication Features

When a user adds a note to an outage:
1. The system extracts the user's email from their authenticated session
2. The note's `author` field is automatically set to this email
3. Users cannot impersonate others - attribution is enforced by the backend

## Deployment Options

### 1. Local Development

```bash
# Copy and edit configuration
cp config.example.yaml config.yaml

# Set up PostgreSQL
createdb outalator
psql -U outalator -d outalator -f migrations/001_initial_schema.sql

# Run application
go run cmd/outalator/main.go
```

### 2. Docker

```bash
# Build image
docker build -t outalator:latest .

# Run with docker-compose (example)
docker-compose up -d
```

### 3. Kubernetes with Helm (Recommended)

```bash
# Install with authentication enabled
helm install outalator ./helm/outalator \
  --namespace outalator \
  --create-namespace \
  --set config.auth.enabled=true \
  --set config.auth.issuer=https://yourcompany.okta.com \
  --set secrets.auth.clientId=<your-client-id> \
  --set secrets.auth.clientSecret=<your-client-secret> \
  --set secrets.auth.sessionKey=$(openssl rand -base64 32)
```

### 4. Kubernetes with Kustomize

```bash
# Production
kubectl apply -k k8s/overlays/prod

# Development
kubectl apply -k k8s/overlays/dev
```

## Configuration Overview

### Minimum Required Configuration

```yaml
server:
  host: 0.0.0.0
  port: 8080

database:
  host: postgres
  port: 5432
  user: outalator
  password: <secure-password>
  dbname: outalator

auth:
  enabled: true
  issuer: https://yourcompany.okta.com
  client_id: <your-client-id>
  client_secret: <your-client-secret>
  redirect_url: https://outalator.yourcompany.com/auth/callback
  session_key: <32-byte-base64-key>
```

### Optional Integrations

```yaml
# PagerDuty
pagerduty:
  api_key: <your-pagerduty-key>

# OpsGenie
opsgenie:
  api_key: <your-opsgenie-key>
```

## Setting Up Okta

1. **Create OIDC Application in Okta**
   - Go to Applications → Create App Integration
   - Choose OIDC - Web Application
   - Set Sign-in redirect URI: `https://outalator.yourcompany.com/auth/callback`
   - Set Sign-out redirect URI: `https://outalator.yourcompany.com/`

2. **Get Credentials**
   - Copy Client ID
   - Copy Client Secret
   - Note your Okta domain: `https://yourcompany.okta.com`

3. **Assign Users**
   - Go to Assignments tab
   - Assign users or groups who should have access

## Environment Variables

All configuration can be set via environment variables (useful for Kubernetes):

| Variable | Description | Example |
|----------|-------------|---------|
| `AUTH_ENABLED` | Enable authentication | `true` |
| `AUTH_ISSUER` | OIDC issuer URL | `https://yourcompany.okta.com` |
| `AUTH_CLIENT_ID` | OIDC client ID | `0oa1234abcd` |
| `AUTH_CLIENT_SECRET` | OIDC client secret | `secret123` |
| `AUTH_REDIRECT_URL` | OAuth callback URL | `https://outalator.com/auth/callback` |
| `AUTH_SESSION_KEY` | Session encryption key | `base64-encoded-32-bytes` |
| `DB_HOST` | Database host | `postgres` |
| `DB_PORT` | Database port | `5432` |
| `DB_USER` | Database username | `outalator` |
| `DB_PASSWORD` | Database password | `secret` |
| `DB_NAME` | Database name | `outalator` |
| `PAGERDUTY_API_KEY` | PagerDuty API key | Optional |
| `OPSGENIE_API_KEY` | OpsGenie API key | Optional |

## Security Best Practices

### Production Deployment Checklist

- [ ] Use HTTPS/TLS for all connections (set up ingress with cert-manager)
- [ ] Enable authentication (`auth.enabled: true`)
- [ ] Use strong database passwords
- [ ] Generate secure session key: `openssl rand -base64 32`
- [ ] Never commit secrets to git
- [ ] Use external secret management (Vault, AWS Secrets Manager, etc.)
- [ ] Enable network policies in Kubernetes
- [ ] Set resource limits on containers
- [ ] Enable database backups
- [ ] Configure log aggregation and monitoring
- [ ] Set up pod security policies/standards
- [ ] Use non-root container users

### Generating Secure Keys

```bash
# Session key (32 bytes, base64 encoded)
openssl rand -base64 32

# Database password (32 characters)
openssl rand -base64 32 | tr -d "=+/" | cut -c1-32
```

## Architecture Components

```
┌─────────────────────────────────────────────────────────────┐
│                         Ingress (HTTPS)                      │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│                  Outalator Application                       │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  OIDC Middleware (Auth)                              │  │
│  └───────────────────────┬───────────────────────────────┘  │
│  ┌───────────────────────▼───────────────────────────────┐  │
│  │  API Handlers (REST)                                  │  │
│  └───────────────────────┬───────────────────────────────┘  │
│  ┌───────────────────────▼───────────────────────────────┐  │
│  │  Service Layer (Business Logic)                       │  │
│  └─────────┬──────────────────────────────┬──────────────┘  │
│            │                              │                  │
│  ┌─────────▼─────────┐        ┌──────────▼──────────┐      │
│  │  Storage Layer    │        │  Notification Layer │      │
│  │   (PostgreSQL)    │        │  (PagerDuty/Opsgenie)│      │
│  └───────────────────┘        └─────────────────────┘      │
└─────────────────────────────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│                    PostgreSQL Database                       │
└─────────────────────────────────────────────────────────────┘
```

## Key Features

1. **Automatic User Attribution**: Notes are automatically tagged with the authenticated user's email
2. **Multi-Service Alerts**: Import alerts from PagerDuty, OpsGenie, or add custom integrations
3. **Flexible Tagging**: Tag outages with Jira tickets, service names, regions, etc.
4. **Markdown Support**: Notes can be formatted with markdown
5. **RESTful API**: Clean HTTP API for all operations
6. **Cloud Native**: Designed for Kubernetes with Helm charts and Kustomize manifests

## API Quick Reference

```bash
# Create outage
POST /api/v1/outages

# Add note (user automatically attributed)
POST /api/v1/outages/{id}/notes
{
  "content": "Root cause identified",
  "format": "markdown"
}

# Add tag
POST /api/v1/outages/{id}/tags
{
  "key": "jira",
  "value": "OPS-1234"
}

# Search by tag
GET /api/v1/tags/search?key=jira&value=OPS-1234

# Import alert
POST /api/v1/alerts/import
{
  "source": "pagerduty",
  "external_id": "PXYZ123"
}
```

## Documentation

- [README.md](README.md) - Main documentation
- [docs/KUBERNETES.md](docs/KUBERNETES.md) - Detailed Kubernetes deployment guide
- [k8s/README.md](k8s/README.md) - Kustomize manifests documentation
- [helm/outalator/](helm/outalator/) - Helm chart

## Support

For issues, questions, or contributions:
- GitHub Issues: https://github.com/conall/outalator/issues
- Documentation: See docs/ directory

## License

MIT License - See LICENSE file for details
