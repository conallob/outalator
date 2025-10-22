# Outalator Quick Start Guide

Get Outalator up and running in 5 minutes!

## Local Development (No Auth)

```bash
# 1. Set up PostgreSQL
createdb outalator
psql -d outalator -f migrations/001_initial_schema.sql

# 2. Run the application
go run cmd/outalator/main.go

# 3. Test the API
curl http://localhost:8080/health
```

## Kubernetes with Helm (With Auth)

```bash
# 1. Set up Okta (see DEPLOYMENT.md for details)
# You'll need: issuer URL, client ID, and client secret

# 2. Install with Helm
helm install outalator ./helm/outalator \
  --create-namespace \
  --namespace outalator \
  --set ingress.hosts[0].host=outalator.yourcompany.com \
  --set config.auth.enabled=true \
  --set config.auth.issuer=https://yourcompany.okta.com \
  --set secrets.auth.clientId=your-okta-client-id \
  --set secrets.auth.clientSecret=your-okta-client-secret \
  --set secrets.auth.sessionKey=$(openssl rand -base64 32)

# 3. Run database migrations
POSTGRES_POD=$(kubectl get pod -n outalator -l app.kubernetes.io/name=postgresql -o jsonpath='{.items[0].metadata.name}')
kubectl cp migrations/001_initial_schema.sql outalator/$POSTGRES_POD:/tmp/
kubectl exec -n outalator $POSTGRES_POD -- psql -U outalator -d outalator -f /tmp/001_initial_schema.sql

# 4. Access the application
# Set up port-forwarding (for testing)
kubectl port-forward -n outalator svc/outalator 8080:80

# Or configure your DNS to point to the ingress
# Then access: https://outalator.yourcompany.com
```

## Creating Your First Outage

```bash
# 1. Create an outage
curl -X POST http://localhost:8080/api/v1/outages \
  -H "Content-Type: application/json" \
  -d '{
    "title": "API Gateway Outage",
    "description": "Users unable to access API",
    "severity": "high",
    "tags": [
      {"key": "jira", "value": "OPS-1234"},
      {"key": "service", "value": "api-gateway"}
    ]
  }'

# Save the outage ID from the response

# 2. Add a note (will be attributed to your authenticated user if auth is enabled)
curl -X POST http://localhost:8080/api/v1/outages/{OUTAGE_ID}/notes \
  -H "Content-Type: application/json" \
  -d '{
    "content": "## Investigation\n\nChecking database connections...",
    "format": "markdown"
  }'

# 3. Add more tags
curl -X POST http://localhost:8080/api/v1/outages/{OUTAGE_ID}/tags \
  -H "Content-Type: application/json" \
  -d '{
    "key": "region",
    "value": "us-west-2"
  }'

# 4. Get the outage with all details
curl http://localhost:8080/api/v1/outages/{OUTAGE_ID}
```

## Importing Alerts from PagerDuty

```bash
# 1. Configure PagerDuty API key in config.yaml or environment variable
export PAGERDUTY_API_KEY=your-api-key

# 2. Import an alert
curl -X POST http://localhost:8080/api/v1/alerts/import \
  -H "Content-Type: application/json" \
  -d '{
    "source": "pagerduty",
    "external_id": "PXYZ123"
  }'

# 3. Associate with an existing outage
curl -X POST http://localhost:8080/api/v1/alerts/import \
  -H "Content-Type: application/json" \
  -d '{
    "source": "pagerduty",
    "external_id": "PXYZ123",
    "outage_id": "your-outage-uuid"
  }'
```

## Searching Outages by Tag

```bash
# Find all outages tagged with a specific Jira ticket
curl "http://localhost:8080/api/v1/tags/search?key=jira&value=OPS-1234"

# Find all outages for a specific service
curl "http://localhost:8080/api/v1/tags/search?key=service&value=api-gateway"
```

## Updating an Outage

```bash
# Mark an outage as resolved
curl -X PATCH http://localhost:8080/api/v1/outages/{OUTAGE_ID} \
  -H "Content-Type: application/json" \
  -d '{
    "status": "resolved"
  }'

# Update severity
curl -X PATCH http://localhost:8080/api/v1/outages/{OUTAGE_ID} \
  -H "Content-Type: application/json" \
  -d '{
    "severity": "medium"
  }'
```

## Common Configurations

### Development (No Auth)

```yaml
# config.yaml
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

# Auth is disabled by default
```

### Production (With Auth)

```yaml
# config.yaml
server:
  host: 0.0.0.0
  port: 8080

database:
  host: postgres.production.svc.cluster.local
  port: 5432
  user: outalator
  password: ${DB_PASSWORD}  # Use env var
  dbname: outalator
  sslmode: require

auth:
  enabled: true
  issuer: https://yourcompany.okta.com
  client_id: ${AUTH_CLIENT_ID}
  client_secret: ${AUTH_CLIENT_SECRET}
  redirect_url: https://outalator.yourcompany.com/auth/callback
  session_key: ${AUTH_SESSION_KEY}

pagerduty:
  api_key: ${PAGERDUTY_API_KEY}

opsgenie:
  api_key: ${OPSGENIE_API_KEY}
```

## Troubleshooting

### Database Connection Failed
```bash
# Check PostgreSQL is running
psql -U outalator -d outalator -c "SELECT 1"

# Check connection from the application
kubectl logs -n outalator -l app=outalator
```

### Authentication Not Working
```bash
# Verify Okta configuration
# - Check issuer URL is correct
# - Verify client ID and secret
# - Ensure redirect URL matches exactly (including http/https)

# Check application logs
kubectl logs -n outalator -l app=outalator | grep -i auth
```

### Alerts Not Importing
```bash
# Verify API keys are set
kubectl get secret outalator-integrations -n outalator -o yaml

# Test API key manually
curl -H "Authorization: Token token=${PAGERDUTY_API_KEY}" \
  https://api.pagerduty.com/incidents
```

## Next Steps

- Read the [full README](README.md) for detailed API documentation
- See [DEPLOYMENT.md](DEPLOYMENT.md) for production deployment guidance
- Check [docs/KUBERNETES.md](docs/KUBERNETES.md) for comprehensive Kubernetes guide
- Review [migrations/001_initial_schema.sql](migrations/001_initial_schema.sql) to understand the database schema

## Support

- Report issues: https://github.com/conall/outalator/issues
- Read docs: See docs/ directory
- Check examples: See API documentation in README.md
