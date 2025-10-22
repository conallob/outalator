# Database Migrations

This directory contains SQL migration files for the outalator database schema.

## Running Migrations

### Option 1: Manual Migration (psql)

```bash
psql -U outalator -d outalator -f migrations/001_initial_schema.sql
```

### Option 2: Using Docker

If you're running PostgreSQL in Docker:

```bash
docker exec -i postgres_container psql -U outalator -d outalator < migrations/001_initial_schema.sql
```

### Option 3: Using a Migration Tool

For production deployments, consider using a migration tool like:

- [golang-migrate](https://github.com/golang-migrate/migrate)
- [goose](https://github.com/pressly/goose)
- [dbmate](https://github.com/amacneil/dbmate)

Example with golang-migrate:

```bash
migrate -path migrations -database "postgresql://outalator:outalator@localhost:5432/outalator?sslmode=disable" up
```

## Migration Files

- `001_initial_schema.sql` - Initial database schema including tables for outages, alerts, notes, and tags

## Schema Overview

### Tables

1. **outages** - Main table for tracking outages/incidents
2. **alerts** - Paging alerts from notification services (PagerDuty, OpsGenie)
3. **notes** - Free-form plaintext or markdown notes attached to outages
4. **tags** - Key-value metadata tags for outages (e.g., Jira tickets)

All tables use UUIDs for primary keys and include appropriate indexes for query performance.
