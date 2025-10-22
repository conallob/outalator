# Secret Management Summary

## Overview

Outalator Helm chart and Kustomize configurations now support flexible secret management suitable for both development and production environments.

## Key Changes

### ✅ Helm Chart Enhancements

1. **External Secrets Support**: Can reference existing Kubernetes secrets instead of creating from values
2. **Conditional Secret Creation**: Secrets only created if explicitly enabled
3. **Resource Policy**: Secrets annotated with `helm.sh/resource-policy: keep` to prevent deletion
4. **Optional Integration Secrets**: PagerDuty/OpsGenie secrets are optional

### ✅ Kustomize Enhancements

1. **SecretGenerator Support**: Multiple options for generating secrets
2. **Example Overlays**: Pre-configured examples for different approaches
3. **No Committed Secrets**: Base secret.yaml excluded by default

## Usage Patterns

### Development

```yaml
# Helm: values-dev.yaml
secretManagement:
  useExistingSecrets: false
secrets:
  createFromValues: true
  database:
    username: "outalator"
    password: "dev-password"
```

```yaml
# Kustomize: secretGenerator
secretGenerator:
  - name: outalator-db-secret
    literals:
      - username=outalator
      - password=dev-password
```

### Production

```yaml
# Helm: values-prod.yaml
secretManagement:
  useExistingSecrets: true
  existingSecrets:
    database: "outalator-db-secret"
    auth: "outalator-auth-secret"
```

Create secrets using External Secrets Operator, Sealed Secrets, or Vault.

## Secret Management Solutions

| Solution | Files | Use Case |
|----------|-------|----------|
| **kubectl** | `helm/outalator/examples/kubectl-create-secrets.yaml` | Quick setup, development |
| **External Secrets Operator** | `helm/outalator/examples/external-secrets-operator.yaml` | Production, AWS/GCP/Azure |
| **Sealed Secrets** | `helm/outalator/examples/sealed-secrets.yaml` | GitOps workflows |
| **Vault** | `helm/outalator/examples/vault-integration.yaml` | Enterprise environments |
| **Kustomize secretGenerator** | `k8s/examples/kustomize-secretgenerator/` | Development |
| **Kustomize + External Secrets** | `k8s/examples/external-secrets/` | Production |

## Quick Start

### 1. Create Secrets with kubectl (Development)

```bash
kubectl create secret generic outalator-db-secret \
  --namespace=outalator \
  --from-literal=username=outalator \
  --from-literal=password=$(openssl rand -base64 32)

kubectl create secret generic outalator-auth-secret \
  --namespace=outalator \
  --from-literal=client_id=your-okta-client-id \
  --from-literal=client_secret=your-okta-client-secret \
  --from-literal=session_key=$(openssl rand -base64 32)
```

### 2. Install Helm Chart

```bash
helm install outalator ./helm/outalator \
  --namespace outalator \
  --set secretManagement.useExistingSecrets=true
```

## Production Deployment

### Option 1: External Secrets Operator (Recommended)

1. Install ESO:
```bash
helm install external-secrets external-secrets/external-secrets \
  --namespace external-secrets-system --create-namespace
```

2. Apply ExternalSecret resources:
```bash
kubectl apply -f helm/outalator/examples/external-secrets-operator.yaml
```

3. Install Outalator:
```bash
helm install outalator ./helm/outalator \
  --set secretManagement.useExistingSecrets=true
```

### Option 2: Sealed Secrets

1. Install controller:
```bash
kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.24.0/controller.yaml
```

2. Create and seal secrets:
```bash
kubectl create secret generic outalator-db-secret \
  --dry-run=client --from-literal=username=outalator \
  --from-literal=password=$(openssl rand -base64 32) \
  -o yaml | kubeseal -o yaml > sealed-db-secret.yaml

kubectl apply -f sealed-db-secret.yaml
```

3. Install Outalator:
```bash
helm install outalator ./helm/outalator \
  --set secretManagement.useExistingSecrets=true
```

## Security Best Practices

✅ **DO:**
- Use external secret management in production
- Rotate secrets regularly
- Use strong random passwords
- Add secret files to .gitignore
- Use RBAC to limit secret access
- Enable audit logging

❌ **DON'T:**
- Commit plaintext secrets to git
- Use weak passwords
- Share secrets between environments
- Store secrets in values files for production
- Disable authentication in production

## Documentation

- **[SECRET-MANAGEMENT.md](docs/SECRET-MANAGEMENT.md)** - Complete guide with all options
- **[KUBERNETES.md](docs/KUBERNETES.md)** - Kubernetes deployment guide
- **[helm/outalator/examples/](helm/outalator/examples/)** - Example configurations
- **[k8s/examples/](k8s/examples/)** - Kustomize examples

## Required Secrets

### Database (outalator-db-secret)
- `username` - PostgreSQL username
- `password` - PostgreSQL password

### Authentication (outalator-auth-secret)
- `client_id` - OIDC client ID (Okta, Auth0, etc.)
- `client_secret` - OIDC client secret
- `session_key` - 32-byte base64 session encryption key

### Integrations (outalator-integrations-secret) - Optional
- `pagerduty_api_key` - PagerDuty API key
- `opsgenie_api_key` - OpsGenie API key

## Generating Secrets

```bash
# Database password
openssl rand -base64 32 | tr -d "=+/" | cut -c1-32

# Session key
openssl rand -base64 32

# General secret
openssl rand -hex 32
```

## Troubleshooting

### Helm Install Fails with "secret not found"

```bash
# Check if secrets exist
kubectl get secrets -n outalator

# Create missing secrets
kubectl create secret generic outalator-db-secret \
  --namespace=outalator \
  --from-literal=username=outalator \
  --from-literal=password=changeme
```

### External Secret Not Syncing

```bash
# Check ExternalSecret status
kubectl get externalsecrets -n outalator
kubectl describe externalsecret outalator-db-external -n outalator

# Check ESO logs
kubectl logs -n external-secrets-system \
  -l app.kubernetes.io/name=external-secrets
```

## Migration Guide

### From Inline Secrets to External Secrets

1. Create secrets externally:
```bash
kubectl create secret generic outalator-db-secret --from-literal=...
```

2. Update values:
```yaml
# Before
secrets:
  database:
    username: "outalator"
    password: "password"

# After
secretManagement:
  useExistingSecrets: true
  existingSecrets:
    database: "outalator-db-secret"
```

3. Upgrade:
```bash
helm upgrade outalator ./helm/outalator -f values-updated.yaml
```

## Support

For issues or questions:
- Review [SECRET-MANAGEMENT.md](docs/SECRET-MANAGEMENT.md)
- Check [examples/](helm/outalator/examples/)
- Open an issue on GitHub
