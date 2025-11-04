# Security Considerations for Custom Fields

## Overview

The custom fields functionality (metadata, custom_fields, source_metadata) allows flexible data storage in Outalator. However, this flexibility introduces security considerations that must be addressed.

## Threat Model

### 1. JSON Injection Attacks

**Risk:** Malicious users could inject crafted JSON to exploit vulnerabilities.

**Mitigation:**
- All JSON is validated before storage using `internal/validation`
- Maximum sizes are enforced (64KB for custom_fields)
- Maximum nesting depth is enforced (10 levels)
- Keys and values have length limits

**Example of blocked attack:**
```json
{
  "custom_fields": {
    "exploit": "<script>alert('xss')</script>"
  }
}
```

While stored as-is, this will be safely escaped when rendered in any HTML context.

### 2. Denial of Service (DoS) via Large Payloads

**Risk:** Attackers could send extremely large or deeply nested JSON to consume resources.

**Mitigation:**
- Maximum custom_fields size: 64KB
- Maximum nesting depth: 10 levels
- Maximum metadata keys: 100
- Maximum metadata key length: 255 bytes
- Maximum metadata value length: 4KB

**Configuration:**
```go
const (
    MaxCustomFieldsSize = 65536    // 64KB
    MaxCustomFieldsDepth = 10
    MaxMetadataKeys = 100
    MaxMetadataKeyLength = 255
    MaxMetadataValueLength = 4096
)
```

### 3. SQL Injection via JSON Fields

**Risk:** JSON fields could be crafted to escape PostgreSQL JSONB queries.

**Mitigation:**
- All database queries use parameterized statements
- PostgreSQL JSONB type handles escaping automatically
- No raw SQL concatenation with user-provided JSON

**Safe query example:**
```go
// Safe - parameterized query
query := `SELECT * FROM outages WHERE metadata @> $1`
db.QueryContext(ctx, query, metadataJSON)

// UNSAFE - DO NOT DO THIS
// query := fmt.Sprintf("SELECT * FROM outages WHERE metadata @> '%s'", userInput)
```

### 4. Information Disclosure

**Risk:** Custom fields could be used to store sensitive information that gets exposed.

**Mitigation:**
- **Never store sensitive data** (passwords, API keys, PII) in custom fields
- Custom fields are NOT encrypted at rest in the database
- Access control should be enforced at the application layer
- Consider field-level encryption if sensitive data must be stored

**Best Practice:**
```go
// GOOD - Store reference to encrypted data
customFields := map[string]any{
    "encrypted_config_id": "cfg_abc123",
}

// BAD - Never do this
customFields := map[string]any{
    "api_key": "sk_live_abc123...",  // NEVER!
    "password": "secret123",         // NEVER!
}
```

### 5. Cross-Site Scripting (XSS)

**Risk:** Custom fields displayed in web UI could execute malicious scripts.

**Mitigation:**
- Always escape/sanitize custom field values when rendering in HTML
- Use JSON-safe encoding when embedding in JavaScript
- Content Security Policy (CSP) headers should be configured

**Safe rendering example:**
```html
<!-- Safe - Template engine escapes HTML -->
<div>{{ .CustomFields.region }}</div>

<!-- UNSAFE - Raw rendering -->
<div>{{ .CustomFields.region | raw }}</div>
```

### 6. Prototype Pollution (JavaScript/JSON)

**Risk:** Malicious JSON could pollute object prototypes in JavaScript clients.

**Mitigation:**
- Use `JSON.parse()` with reviver function to sanitize
- Avoid using `__proto__`, `constructor`, or `prototype` keys
- Validate keys against allowlist when processing client-side

**Safe parsing:**
```javascript
// Safe parsing with validation
const safeJSON = JSON.parse(jsonString, (key, value) => {
    if (key === '__proto__' || key === 'constructor' || key === 'prototype') {
        return undefined;  // Remove dangerous keys
    }
    return value;
});
```

### 7. Resource Exhaustion via JSONB Indexing

**Risk:** Large numbers of updates to JSONB fields could cause index bloat.

**Mitigation:**
- GIN indexes are used for efficient querying
- Monitor index sizes and run `REINDEX` if necessary
- Consider partitioning tables if custom fields grow very large
- Implement rate limiting on update operations

### 8. Regex Denial of Service (ReDoS)

**Risk:** If custom fields contain regex patterns used in queries, malicious patterns could cause DoS.

**Mitigation:**
- Never use custom field values directly as regex patterns
- If regex matching is needed, sanitize and validate patterns first
- Set query timeouts in PostgreSQL

## Validation Rules

All custom fields must pass validation defined in `internal/validation/json_validator.go`:

```go
// Metadata validation
- Max 100 keys
- Key length: 1-255 bytes
- Value length: 1-4096 bytes
- Keys cannot be empty

// Custom fields validation
- Max size: 64KB when JSON-encoded
- Max nesting depth: 10 levels
- Must be valid JSON
```

## Access Control

### Authentication Required

All endpoints that accept custom fields should require authentication:

```go
func (h *Handler) CreateOutage(w http.ResponseWriter, r *http.Request) {
    // Verify authentication
    user := middleware.GetAuthenticatedUser(r)
    if user == nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Proceed with validated user
}
```

### Authorization Checks

Consider implementing role-based access control (RBAC):

```go
// Only allow certain roles to add custom fields
if !user.HasPermission("custom_fields.write") {
    http.Error(w, "Forbidden", http.StatusForbidden)
    return
}
```

### Audit Logging

Log all changes to custom fields for security auditing:

```go
log.Printf("User %s updated outage %s custom fields: %v",
    user.ID, outageID, customFields)
```

## Database Security

### 1. Least Privilege

The database user should have minimum required permissions:

```sql
-- Good: Limited permissions
GRANT SELECT, INSERT, UPDATE, DELETE ON outages TO outalator_app;

-- Bad: Excessive permissions
GRANT ALL PRIVILEGES ON ALL TABLES TO outalator_app;
```

### 2. Connection Security

Always use SSL/TLS for database connections:

```go
connStr := fmt.Sprintf(
    "host=%s port=%d user=%s password=%s dbname=%s sslmode=require",
    cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
)
```

### 3. Backup and Recovery

- Regular backups of the database including JSONB columns
- Test restoration procedures
- Consider point-in-time recovery (PITR) for critical data

## Monitoring and Alerting

### Metrics to Monitor

```
- custom_fields_size_bytes (histogram)
- custom_fields_validation_errors (counter)
- metadata_update_rate (gauge)
- jsonb_query_duration_seconds (histogram)
```

### Alerts to Configure

1. **Large Custom Fields**: Alert if custom_fields exceed 32KB
2. **High Validation Failure Rate**: Alert if >5% of requests fail validation
3. **Slow JSONB Queries**: Alert if queries take >1s
4. **Index Bloat**: Alert if GIN index size grows >2x expected

## Incident Response

### If Custom Fields Are Compromised

1. **Immediately disable** custom fields writes if an attack is detected
2. **Audit logs** to identify affected records
3. **Sanitize** or remove malicious data
4. **Rotate** any exposed credentials
5. **Patch** vulnerabilities and update validation rules
6. **Notify** affected users if required by regulations

### Emergency Rollback

If custom fields cause critical issues:

```bash
# Run rollback migration
psql -d outalator -f migrations/002_add_custom_fields_rollback.sql

# Restart application without custom fields support
export CUSTOM_FIELDS_ENABLED=false
./bin/outalator
```

## Security Checklist

- [ ] Input validation on all custom field endpoints
- [ ] Rate limiting on write operations
- [ ] Authentication required for all custom field operations
- [ ] Authorization checks for sensitive operations
- [ ] Audit logging enabled
- [ ] Database connections use SSL/TLS
- [ ] Content Security Policy (CSP) headers configured
- [ ] XSS protection in web UI
- [ ] Regular security audits of custom field usage
- [ ] Backup and recovery procedures tested
- [ ] Monitoring and alerting configured
- [ ] Incident response plan documented

## References

- [OWASP JSON Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Security_Cheat_Sheet.html)
- [PostgreSQL JSONB Security](https://www.postgresql.org/docs/current/datatype-json.html)
- [CWE-79: Cross-site Scripting (XSS)](https://cwe.mitre.org/data/definitions/79.html)
- [CWE-89: SQL Injection](https://cwe.mitre.org/data/definitions/89.html)

## Contact

For security issues, please email: security@example.com

**DO NOT** open public GitHub issues for security vulnerabilities.
