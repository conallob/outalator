# Kubernetes Manifests

This directory contains Kubernetes manifests for deploying Outalator.

## Structure

```
k8s/
├── base/                    # Base Kubernetes resources
│   ├── deployment.yaml      # Main application deployment
│   ├── service.yaml         # Service for the application
│   ├── configmap.yaml       # Configuration settings
│   ├── secret.yaml          # Secrets (DO NOT commit real secrets!)
│   ├── ingress.yaml         # Ingress configuration
│   ├── postgres.yaml        # PostgreSQL StatefulSet
│   └── kustomization.yaml   # Base kustomization
└── overlays/                # Environment-specific overlays
    ├── dev/                 # Development environment
    │   └── kustomization.yaml
    └── prod/                # Production environment
        └── kustomization.yaml
```

## Quick Start

### Development Environment

```bash
# Review and update secrets in k8s/base/secret.yaml
# IMPORTANT: Use a proper secret management solution in production

# Apply the development overlay
kubectl apply -k overlays/dev

# Check status
kubectl get pods -n outalator-dev
```

### Production Environment

```bash
# Update production secrets using a proper secret management solution
# Example with kubectl create secret:
kubectl create secret generic outalator-db-secret \
  --from-literal=username=outalator \
  --from-literal=password=your-secure-password \
  --namespace=outalator-prod

kubectl create secret generic outalator-auth-secret \
  --from-literal=client_id=your-okta-client-id \
  --from-literal=client_secret=your-okta-client-secret \
  --from-literal=session_key=$(openssl rand -base64 32) \
  --namespace=outalator-prod

# Apply the production overlay
kubectl apply -k overlays/prod

# Check status
kubectl get pods -n outalator-prod
```

## Customization

### Using Kustomize Overlays

1. Create a new overlay directory:
   ```bash
   mkdir -p overlays/staging
   ```

2. Create a `kustomization.yaml`:
   ```yaml
   apiVersion: kustomize.config.k8s.io/v1beta1
   kind: Kustomization

   bases:
     - ../../base

   namespace: outalator-staging

   replicas:
     - name: outalator
       count: 2

   configMapGenerator:
     - name: outalator-config
       behavior: merge
       literals:
         - auth_enabled=true
         - auth_issuer=https://staging.okta.com
   ```

3. Apply:
   ```bash
   kubectl apply -k overlays/staging
   ```

## Security Notes

**IMPORTANT**: The `base/secret.yaml` file contains example secrets for development only.

For production deployments:

1. **Never commit real secrets to git**
2. Use one of these secret management solutions:
   - [Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets)
   - [External Secrets Operator](https://external-secrets.io/)
   - [HashiCorp Vault](https://www.vaultproject.io/)
   - Cloud provider solutions (AWS Secrets Manager, GCP Secret Manager, Azure Key Vault)

3. Remove the `secret.yaml` from the kustomization resources list and create secrets separately

## Configuration

### ConfigMap Values

Edit `base/configmap.yaml` to configure:
- `db_host`: PostgreSQL hostname
- `db_port`: PostgreSQL port
- `db_name`: Database name
- `auth_enabled`: Enable/disable authentication
- `auth_issuer`: OIDC issuer URL
- `auth_redirect_url`: OAuth callback URL

### Secrets

Secrets required:
- `outalator-db-secret`: Database credentials
  - `username`: Database username
  - `password`: Database password
- `outalator-auth-secret`: OIDC credentials
  - `client_id`: OIDC client ID
  - `client_secret`: OIDC client secret
  - `session_key`: Session encryption key (32-byte base64)

## Running Database Migrations

After deploying, run database migrations:

```bash
# Get the PostgreSQL pod name
POSTGRES_POD=$(kubectl get pod -n <namespace> -l app=postgres -o jsonpath='{.items[0].metadata.name}')

# Copy migration file
kubectl cp ../migrations/001_initial_schema.sql <namespace>/$POSTGRES_POD:/tmp/

# Run migration
kubectl exec -n <namespace> $POSTGRES_POD -- \
  psql -U outalator -d outalator -f /tmp/001_initial_schema.sql
```

## Monitoring

Check application health:

```bash
kubectl port-forward -n <namespace> svc/outalator 8080:80
curl http://localhost:8080/health
```

View logs:

```bash
kubectl logs -n <namespace> -l app=outalator -f
```

## Troubleshooting

### Common Issues

1. **Pods not starting**: Check logs
   ```bash
   kubectl describe pod -n <namespace> <pod-name>
   kubectl logs -n <namespace> <pod-name>
   ```

2. **Database connection errors**: Verify database is running
   ```bash
   kubectl get pod -n <namespace> -l app=postgres
   ```

3. **Authentication errors**: Verify OIDC configuration
   - Check that issuer URL is correct
   - Verify client ID and secret
   - Ensure redirect URL matches exactly

## Further Documentation

For complete deployment guide with Helm charts, see [../docs/KUBERNETES.md](../docs/KUBERNETES.md)
