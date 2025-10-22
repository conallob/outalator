# Kubernetes Deployment Guide

This guide covers deploying Outalator on Kubernetes using either Helm or Kustomize.

## Prerequisites

- Kubernetes cluster (1.19+)
- kubectl configured to access your cluster
- Helm 3+ (for Helm deployment)
- PostgreSQL database (can be deployed with the chart)
- OIDC provider configured (e.g., Okta, Auth0, Google, etc.)
- **Secret management solution** (see [Secret Management Guide](SECRET-MANAGEMENT.md))

## Important: Secret Management

**Before deploying**, review the [Secret Management Guide](SECRET-MANAGEMENT.md) to choose the right approach for your environment:

- **Development**: Use kubectl create secrets or secretGenerator
- **Production**: Use External Secrets Operator, Sealed Secrets, or Vault

Never commit plaintext secrets to git!

## Deployment Options

### Option 1: Helm Chart (Recommended)

#### 1. Configure Values

Create a `values-override.yaml` file with your configuration:

```yaml
image:
  repository: your-registry/outalator
  tag: "v1.0.0"

ingress:
  hosts:
    - host: outalator.yourcompany.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: outalator-tls
      hosts:
        - outalator.yourcompany.com

config:
  database:
    host: "postgres"
    port: 5432
    name: "outalator"

  auth:
    enabled: true
    issuer: "https://yourcompany.okta.com"
    redirectUrl: "https://outalator.yourcompany.com/auth/callback"

# Secret Management Options:

# Option 1: Use existing Kubernetes secrets (RECOMMENDED for production)
secretManagement:
  useExistingSecrets: true
  existingSecrets:
    database: "outalator-db-secret"
    auth: "outalator-auth-secret"
    integrations: "outalator-integrations-secret"

# Option 2: Create secrets from values (ONLY for development/testing)
# secrets:
#   createFromValues: true
#   database:
#     username: "outalator"
#     password: "super-secret-password"
#   auth:
#     clientId: "your-okta-client-id"
#     clientSecret: "your-okta-client-secret"
#     sessionKey: "generate-a-random-32-byte-base64-key"

postgresql:
  enabled: true
  auth:
    username: outalator
    password: super-secret-password
    database: outalator
  primary:
    persistence:
      enabled: true
      size: 20Gi
```

#### 2. Install with Helm

```bash
# Create namespace
kubectl create namespace outalator

# Install the chart
helm install outalator ./helm/outalator \
  --namespace outalator \
  --values values-override.yaml

# Or upgrade existing installation
helm upgrade --install outalator ./helm/outalator \
  --namespace outalator \
  --values values-override.yaml
```

#### 3. Verify Deployment

```bash
kubectl get pods -n outalator
kubectl get svc -n outalator
kubectl get ingress -n outalator
```

#### 4. Run Database Migrations

```bash
# Get the PostgreSQL pod name
POSTGRES_POD=$(kubectl get pod -n outalator -l app.kubernetes.io/name=postgresql -o jsonpath='{.items[0].metadata.name}')

# Copy migration file to pod
kubectl cp migrations/001_initial_schema.sql outalator/$POSTGRES_POD:/tmp/

# Run migration
kubectl exec -n outalator $POSTGRES_POD -- psql -U outalator -d outalator -f /tmp/001_initial_schema.sql
```

### Option 2: Kustomize

#### 1. Update Base Configuration

Edit `k8s/base/configmap.yaml` and `k8s/base/secret.yaml` with your values.

**Important:** Never commit real secrets to git. Use a secret management solution like:
- Sealed Secrets
- External Secrets Operator
- HashiCorp Vault
- Cloud provider secret managers (AWS Secrets Manager, GCP Secret Manager, Azure Key Vault)

#### 2. Deploy with Kustomize

For development:
```bash
kubectl apply -k k8s/overlays/dev
```

For production:
```bash
kubectl apply -k k8s/overlays/prod
```

#### 3. Run Database Migrations

```bash
# Get the PostgreSQL pod name
POSTGRES_POD=$(kubectl get pod -n outalator -l app=postgres -o jsonpath='{.items[0].metadata.name}')

# Copy migration file to pod
kubectl cp migrations/001_initial_schema.sql outalator/$POSTGRES_POD:/tmp/

# Run migration
kubectl exec -n outalator $POSTGRES_POD -- psql -U outalator -d outalator -f /tmp/001_initial_schema.sql
```

## Configuring Okta

### 1. Create an Okta Application

1. Log in to your Okta admin console
2. Go to **Applications** > **Applications**
3. Click **Create App Integration**
4. Select **OIDC - OpenID Connect**
5. Select **Web Application**
6. Configure the application:
   - **App integration name**: Outalator
   - **Sign-in redirect URIs**: `https://outalator.yourcompany.com/auth/callback`
   - **Sign-out redirect URIs**: `https://outalator.yourcompany.com/`
   - **Controlled access**: Choose appropriate access control

### 2. Get Credentials

After creating the application:
- Note the **Client ID**
- Note the **Client Secret**
- Note your Okta domain (e.g., `https://yourcompany.okta.com`)

### 3. Assign Users

1. Go to the **Assignments** tab
2. Assign users or groups who should have access to Outalator

## Configuration Options

### Environment Variables

All configuration can be overridden with environment variables:

- `AUTH_ENABLED` - Enable/disable authentication (true/false)
- `AUTH_ISSUER` - OIDC issuer URL
- `AUTH_CLIENT_ID` - OIDC client ID
- `AUTH_CLIENT_SECRET` - OIDC client secret
- `AUTH_REDIRECT_URL` - OAuth callback URL
- `AUTH_SESSION_KEY` - Session encryption key (32-byte base64 string)
- `DB_HOST` - Database host
- `DB_PORT` - Database port
- `DB_NAME` - Database name
- `DB_USER` - Database username
- `DB_PASSWORD` - Database password
- `PAGERDUTY_API_KEY` - PagerDuty API key (optional)
- `OPSGENIE_API_KEY` - OpsGenie API key (optional)

### Generating a Session Key

```bash
openssl rand -base64 32
```

## Security Best Practices

### 1. Use External Secrets Management

Instead of storing secrets in values files or ConfigMaps:

```yaml
# Example with External Secrets Operator
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: outalator-auth
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: vault-backend
    kind: SecretStore
  target:
    name: outalator-auth-secret
  data:
    - secretKey: client_id
      remoteRef:
        key: outalator/okta
        property: client_id
    - secretKey: client_secret
      remoteRef:
        key: outalator/okta
        property: client_secret
```

### 2. Enable Network Policies

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: outalator-network-policy
spec:
  podSelector:
    matchLabels:
      app: outalator
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: ingress-nginx
  egress:
    - to:
        - podSelector:
            matchLabels:
              app: postgres
      ports:
        - protocol: TCP
          port: 5432
    - to:
        - namespaceSelector: {}
      ports:
        - protocol: TCP
          port: 443  # OIDC and external APIs
```

### 3. Use Pod Security Standards

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: outalator
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

## Monitoring and Observability

### Health Checks

The application exposes a health endpoint at `/health` which is used for:
- Kubernetes liveness probes
- Kubernetes readiness probes
- External monitoring

### Metrics (Future Enhancement)

Consider adding Prometheus metrics:

```yaml
# ServiceMonitor for Prometheus Operator
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: outalator
spec:
  selector:
    matchLabels:
      app: outalator
  endpoints:
    - port: http
      path: /metrics
```

## Troubleshooting

### Check Pod Logs

```bash
kubectl logs -n outalator -l app=outalator -f
```

### Check Events

```bash
kubectl get events -n outalator --sort-by='.lastTimestamp'
```

### Debug Authentication Issues

1. Verify Okta configuration:
   - Client ID and Secret are correct
   - Redirect URL matches exactly
   - Users are assigned to the application

2. Check session key:
   ```bash
   kubectl get secret outalator-auth-secret -n outalator -o jsonpath='{.data.session_key}' | base64 -d
   ```

3. Review auth middleware logs for specific errors

### Database Connection Issues

```bash
# Test database connectivity from a pod
kubectl run -it --rm debug --image=postgres:15 --restart=Never -n outalator -- \
  psql -h postgres -U outalator -d outalator
```

## Scaling

### Manual Scaling

```bash
kubectl scale deployment outalator -n outalator --replicas=5
```

### Horizontal Pod Autoscaling

```bash
kubectl autoscale deployment outalator -n outalator \
  --cpu-percent=80 \
  --min=2 \
  --max=10
```

Or in Helm values:

```yaml
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80
```

## Backup and Recovery

### Database Backups

```bash
# Create backup
kubectl exec -n outalator postgres-0 -- \
  pg_dump -U outalator outalator | gzip > outalator-backup-$(date +%Y%m%d).sql.gz

# Restore backup
gunzip -c outalator-backup-20240101.sql.gz | \
  kubectl exec -i -n outalator postgres-0 -- \
  psql -U outalator -d outalator
```

Consider using automated backup solutions like:
- Velero for Kubernetes resources and persistent volumes
- PostgreSQL backup operators
- Cloud provider backup services

## Upgrading

### Helm Upgrade

```bash
helm upgrade outalator ./helm/outalator \
  --namespace outalator \
  --values values-override.yaml
```

### Kustomize Upgrade

```bash
kubectl apply -k k8s/overlays/prod
```

### Database Migrations

Always run database migrations after upgrading:

```bash
kubectl cp migrations/002_new_migration.sql outalator/$POSTGRES_POD:/tmp/
kubectl exec -n outalator $POSTGRES_POD -- psql -U outalator -d outalator -f /tmp/002_new_migration.sql
```
