# Secret Management Guide

This guide covers various approaches to managing secrets for Outalator in Kubernetes.

## Table of Contents

- [Overview](#overview)
- [Development vs Production](#development-vs-production)
- [Helm Chart Secret Management](#helm-chart-secret-management)
- [Kustomize Secret Management](#kustomize-secret-management)
- [Production-Ready Solutions](#production-ready-solutions)
- [Best Practices](#best-practices)

## Overview

Outalator requires several secrets:

| Secret Type | Keys Required | Purpose |
|------------|---------------|---------|
| Database | `username`, `password` | PostgreSQL credentials |
| Authentication | `client_id`, `client_secret`, `session_key` | OIDC provider credentials |
| Integrations (optional) | `pagerduty_api_key`, `opsgenie_api_key` | External service API keys |

## Development vs Production

### Development

For local development and testing, it's acceptable to:
- Create secrets from literal values
- Use secretGenerator in Kustomize
- Include secrets in Helm values (values-dev.yaml)

**Never commit these to git!**

### Production

For production environments, always use:
- External Secrets Operator
- Sealed Secrets
- HashiCorp Vault
- Cloud provider secret managers
- Or other enterprise secret management solutions

## Helm Chart Secret Management

### Option 1: External Secrets (Recommended for Production)

The Helm chart supports referencing existing Kubernetes secrets:

```yaml
# values-production.yaml
secretManagement:
  useExistingSecrets: true
  existingSecrets:
    database: "outalator-db-secret"
    auth: "outalator-auth-secret"
    integrations: "outalator-integrations-secret"
```

Create the secrets before installing:

```bash
# Option A: kubectl create
kubectl create secret generic outalator-db-secret \
  --namespace=outalator \
  --from-literal=username=outalator \
  --from-literal=password=$(openssl rand -base64 32)

kubectl create secret generic outalator-auth-secret \
  --namespace=outalator \
  --from-literal=client_id=your-okta-client-id \
  --from-literal=client_secret=your-okta-client-secret \
  --from-literal=session_key=$(openssl rand -base64 32)

# Option B: External Secrets Operator (see examples/)
kubectl apply -f helm/outalator/examples/external-secrets-operator.yaml
```

Then install Helm chart:

```bash
helm install outalator ./helm/outalator \
  --namespace outalator \
  -f values-production.yaml
```

### Option 2: Create Secrets from Values (Development Only)

```yaml
# values-dev.yaml - DO NOT USE IN PRODUCTION
secretManagement:
  useExistingSecrets: false

secrets:
  createFromValues: true
  database:
    username: "outalator"
    password: "development-only-password"
  auth:
    clientId: "dev-client-id"
    clientSecret: "dev-client-secret"
    sessionKey: "dev-session-key"
```

**Warning:** Add values-dev.yaml to .gitignore!

## Kustomize Secret Management

### Option 1: SecretGenerator with Files (Development)

```yaml
# k8s/examples/kustomize-secretgenerator/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

bases:
  - ../../base

secretGenerator:
  - name: outalator-db-secret
    files:
      - username=secrets/db-username.txt
      - password=secrets/db-password.txt
```

Create secret files:

```bash
mkdir -p secrets
echo -n "outalator" > secrets/db-username.txt
openssl rand -base64 32 > secrets/db-password.txt
echo "secrets/" >> .gitignore
```

Apply:

```bash
kubectl apply -k k8s/examples/kustomize-secretgenerator
```

### Option 2: External Secrets Operator (Production)

```yaml
# k8s/examples/external-secrets/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

bases:
  - ../../base

resources:
  - external-secrets.yaml
```

See `k8s/examples/external-secrets/` for full example.

## Production-Ready Solutions

### 1. External Secrets Operator (ESO)

**Best for:** Multi-cloud, AWS, GCP, Azure, Vault

Install ESO:

```bash
helm repo add external-secrets https://charts.external-secrets.io
helm install external-secrets external-secrets/external-secrets \
  --namespace external-secrets-system \
  --create-namespace
```

Create ExternalSecret resources:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: outalator-db-external
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secretsmanager
    kind: SecretStore
  target:
    name: outalator-db-secret
  data:
    - secretKey: username
      remoteRef:
        key: outalator/database
        property: username
```

See `helm/outalator/examples/external-secrets-operator.yaml` for complete example.

**Supported Backends:**
- AWS Secrets Manager
- AWS Parameter Store
- Azure Key Vault
- GCP Secret Manager
- HashiCorp Vault
- And many more

### 2. Sealed Secrets

**Best for:** GitOps workflows, encrypted secrets in git

Install controller:

```bash
kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.24.0/controller.yaml
```

Install CLI:

```bash
brew install kubeseal  # macOS
# or download from GitHub releases
```

Create sealed secret:

```bash
kubectl create secret generic outalator-db-secret \
  --dry-run=client \
  --from-literal=username=outalator \
  --from-literal=password=$(openssl rand -base64 32) \
  -o yaml | \
kubeseal -o yaml > sealed-db-secret.yaml
```

The sealed secret CAN be committed to git safely:

```bash
git add sealed-db-secret.yaml
git commit -m "Add sealed database secret"
```

See `helm/outalator/examples/sealed-secrets.yaml` for complete example.

### 3. HashiCorp Vault with Vault Secrets Operator

**Best for:** Enterprise environments, existing Vault deployment

Install Vault Secrets Operator:

```bash
helm repo add hashicorp https://helm.releases.hashicorp.com
helm install vault-secrets-operator hashicorp/vault-secrets-operator \
  --namespace vault-secrets-operator-system \
  --create-namespace
```

Store secrets in Vault:

```bash
vault kv put secret/outalator/database \
  username=outalator \
  password=$(openssl rand -base64 32)
```

Create VaultStaticSecret:

```yaml
apiVersion: secrets.hashicorp.com/v1beta1
kind: VaultStaticSecret
metadata:
  name: outalator-db-vault
spec:
  vaultAuthRef: vault-auth
  mount: secret
  path: outalator/database
  destination:
    name: outalator-db-secret
```

See `helm/outalator/examples/vault-integration.yaml` for complete example.

### 4. Cloud Provider Solutions

#### AWS Secrets Manager + IRSA

```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: aws-secretsmanager
spec:
  provider:
    aws:
      service: SecretsManager
      region: us-west-2
      auth:
        jwt:
          serviceAccountRef:
            name: outalator  # With IAM role annotation
```

Annotate ServiceAccount:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: outalator
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::ACCOUNT:role/outalator
```

#### GCP Secret Manager + Workload Identity

```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: gcp-secretmanager
spec:
  provider:
    gcpsm:
      projectID: "your-project-id"
      auth:
        workloadIdentity:
          clusterLocation: us-central1
          clusterName: my-cluster
          serviceAccountRef:
            name: outalator
```

#### Azure Key Vault + Managed Identity

```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: azure-keyvault
spec:
  provider:
    azurekv:
      vaultUrl: "https://your-vault.vault.azure.net"
      authType: ManagedIdentity
```

## Best Practices

### 1. Never Commit Secrets to Git

```bash
# Add to .gitignore
echo "config.yaml" >> .gitignore
echo "values-*.yaml" >> .gitignore
echo "k8s/base/secret.yaml" >> .gitignore
echo "k8s/examples/*/secrets/" >> .gitignore
```

### 2. Use Different Secrets for Each Environment

```
Development: Simple secrets, rotated infrequently
Staging: Similar to production, separate secrets
Production: Strong secrets, automatic rotation, auditing
```

### 3. Rotate Secrets Regularly

- Database passwords: Every 90 days
- API keys: When employees leave or on schedule
- Session keys: Every 30-90 days

```bash
# Example rotation with External Secrets
# Update secret in AWS Secrets Manager
aws secretsmanager update-secret \
  --secret-id outalator/database \
  --secret-string '{"username":"outalator","password":"new-password"}'

# ESO will automatically sync and restart pods
```

### 4. Use Strong Random Secrets

```bash
# Database password (32 chars)
openssl rand -base64 32 | tr -d "=+/" | cut -c1-32

# Session key (32 bytes, base64)
openssl rand -base64 32

# API token (64 chars)
openssl rand -hex 32
```

### 5. Implement Secret Scanning

Use tools like:
- [git-secrets](https://github.com/awslabs/git-secrets)
- [truffleHog](https://github.com/trufflesecurity/trufflehog)
- [detect-secrets](https://github.com/Yelp/detect-secrets)

```bash
# Install git-secrets
brew install git-secrets

# Set up in repository
git secrets --install
git secrets --register-aws
```

### 6. Use Helm's Resource Policy

Prevent accidental deletion of secrets:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: outalator-db-secret
  annotations:
    helm.sh/resource-policy: keep
```

### 7. Audit Secret Access

Enable audit logging in Kubernetes:

```yaml
# audit-policy.yaml
apiVersion: audit.k8s.io/v1
kind: Policy
rules:
  - level: RequestResponse
    resources:
      - group: ""
        resources: ["secrets"]
```

### 8. Limit Secret Scope

Use namespace-scoped secrets, not cluster-wide:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: outalator-db-secret
  namespace: outalator  # Namespace-scoped
```

### 9. Use RBAC to Restrict Access

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: secret-reader
  namespace: outalator
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    resourceNames: ["outalator-db-secret"]
    verbs: ["get"]
```

### 10. Monitor Secret Usage

Set up alerts for:
- Failed secret retrievals
- Unauthorized secret access attempts
- Secret rotation failures
- Expired secrets

## Troubleshooting

### Secret Not Found

```bash
# Check if secret exists
kubectl get secrets -n outalator

# Describe secret
kubectl describe secret outalator-db-secret -n outalator

# Check secret data (base64 encoded)
kubectl get secret outalator-db-secret -n outalator -o yaml
```

### External Secrets Not Syncing

```bash
# Check ExternalSecret status
kubectl get externalsecrets -n outalator
kubectl describe externalsecret outalator-db-external -n outalator

# Check SecretStore
kubectl get secretstore -n outalator
kubectl describe secretstore aws-secretsmanager -n outalator

# Check ESO logs
kubectl logs -n external-secrets-system -l app.kubernetes.io/name=external-secrets
```

### Sealed Secrets Not Decrypting

```bash
# Check controller logs
kubectl logs -n kube-system -l name=sealed-secrets-controller

# Verify sealed secret
kubectl get sealedsecrets -n outalator
kubectl describe sealedsecret outalator-db-sealed -n outalator
```

## Quick Reference

### Generate All Secrets Script

```bash
#!/bin/bash
# generate-secrets.sh

NAMESPACE=outalator

# Database
kubectl create secret generic outalator-db-secret \
  --namespace=$NAMESPACE \
  --from-literal=username=outalator \
  --from-literal=password=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-32) \
  --dry-run=client -o yaml | kubectl apply -f -

# Auth
kubectl create secret generic outalator-auth-secret \
  --namespace=$NAMESPACE \
  --from-literal=client_id=$OKTA_CLIENT_ID \
  --from-literal=client_secret=$OKTA_CLIENT_SECRET \
  --from-literal=session_key=$(openssl rand -base64 32) \
  --dry-run=client -o yaml | kubectl apply -f -

# Integrations (optional)
if [ -n "$PAGERDUTY_API_KEY" ] || [ -n "$OPSGENIE_API_KEY" ]; then
  kubectl create secret generic outalator-integrations-secret \
    --namespace=$NAMESPACE \
    --from-literal=pagerduty_api_key=${PAGERDUTY_API_KEY:-""} \
    --from-literal=opsgenie_api_key=${OPSGENIE_API_KEY:-""} \
    --dry-run=client -o yaml | kubectl apply -f -
fi
```

Usage:

```bash
export OKTA_CLIENT_ID=your-client-id
export OKTA_CLIENT_SECRET=your-client-secret
export PAGERDUTY_API_KEY=your-pagerduty-key
./generate-secrets.sh
```

## Further Reading

- [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)
- [External Secrets Operator](https://external-secrets.io/)
- [Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets)
- [HashiCorp Vault](https://www.vaultproject.io/)
- [AWS Secrets Manager](https://aws.amazon.com/secrets-manager/)
- [GCP Secret Manager](https://cloud.google.com/secret-manager)
- [Azure Key Vault](https://azure.microsoft.com/en-us/products/key-vault)
