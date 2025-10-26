# Import Historical Data

This guide explains how to use the `import-history` tool to bootstrap your Outalator database with historical incident data from PagerDuty or OpsGenie.

## Overview

The import-history tool allows you to:
- Import historical incidents/alerts from PagerDuty or OpsGenie
- Specify a date range for the import
- Filter by specific teams
- Preview what would be imported with dry-run mode
- Automatically create outages and alerts in your database

## Prerequisites

1. **Configuration File**: Ensure you have a valid `config.yaml` with your database settings and API keys
2. **Database**: Your PostgreSQL database must be running and migrations applied
3. **API Keys**: You need a valid PagerDuty or OpsGenie API key configured

### API Key Configuration

Add your API key to `config.yaml`:

```yaml
# For PagerDuty
pagerduty:
  api_key: "your-pagerduty-api-key"

# For OpsGenie
opsgenie:
  api_key: "your-opsgenie-api-key"
```

Or set environment variables:
```bash
export PAGERDUTY_API_KEY="your-pagerduty-api-key"
export OPSGENIE_API_KEY="your-opsgenie-api-key"
```

## Building the Tool

```bash
make build-import
```

Or build manually:
```bash
go build -o bin/import-history cmd/import-history/main.go
```

## Usage

### Basic Usage

Import all PagerDuty incidents from a specific date:
```bash
./bin/import-history -service pagerduty -since 2024-01-01T00:00:00Z
```

Import OpsGenie alerts from a date range:
```bash
./bin/import-history -service opsgenie -since 2024-01-01T00:00:00Z -until 2024-06-01T00:00:00Z
```

### List Available Teams

Before filtering by team, you can list all available teams:

```bash
# PagerDuty
./bin/import-history -service pagerduty -list-teams

# OpsGenie
./bin/import-history -service opsgenie -list-teams
```

### Filter by Team

Import incidents for specific teams only:

```bash
# Single team
./bin/import-history -service pagerduty -since 2024-01-01T00:00:00Z -teams "TEAM_ID_1"

# Multiple teams (comma-separated)
./bin/import-history -service pagerduty -since 2024-01-01T00:00:00Z -teams "TEAM_ID_1,TEAM_ID_2"
```

### Dry Run Mode

Preview what would be imported without making any changes:

```bash
./bin/import-history -service pagerduty -since 2024-01-01T00:00:00Z -dry-run
```

### Advanced Options

```bash
./bin/import-history \
  -service pagerduty \
  -since 2024-01-01T00:00:00Z \
  -until 2024-12-31T23:59:59Z \
  -teams "TEAM_ID_1,TEAM_ID_2" \
  -config /path/to/config.yaml \
  -batch-size 50 \
  -dry-run
```

## Command-Line Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `-service` | Yes | - | Service to import from (`pagerduty` or `opsgenie`) |
| `-since` | Yes* | - | Start date in RFC3339 format (e.g., `2024-01-01T00:00:00Z`) |
| `-until` | No | Now | End date in RFC3339 format |
| `-teams` | No | All teams | Comma-separated list of team IDs to filter |
| `-list-teams` | No | false | List available teams and exit |
| `-dry-run` | No | false | Preview without making changes |
| `-config` | No | `config.yaml` | Path to configuration file |
| `-batch-size` | No | 100 | Number of incidents to fetch per API call |

*Not required when using `-list-teams`

## How It Works

1. **Fetches Incidents**: The tool queries the PagerDuty or OpsGenie API for incidents/alerts in the specified date range
2. **Pagination**: Automatically handles pagination to fetch all matching incidents
3. **Deduplication**: Checks if each incident already exists in the database (by external ID) and skips duplicates
4. **Creates Records**: For each new incident:
   - Creates an Outage record
   - Creates an Alert record linked to the outage
   - Sets the appropriate status (resolved/open) based on incident state
5. **Progress Reporting**: Provides real-time feedback and final statistics

## Examples

### Import Last 30 Days from PagerDuty

```bash
SINCE=$(date -u -d "30 days ago" +%Y-%m-%dT%H:%M:%SZ)
./bin/import-history -service pagerduty -since "$SINCE"
```

### Import Specific Team's History

```bash
# First, list teams to find the ID
./bin/import-history -service pagerduty -list-teams

# Then import for that team
./bin/import-history -service pagerduty -since 2024-01-01T00:00:00Z -teams "P123ABC"
```

### Test Import with Dry Run

```bash
# Preview what would be imported
./bin/import-history -service opsgenie -since 2024-01-01T00:00:00Z -dry-run

# If it looks good, run for real
./bin/import-history -service opsgenie -since 2024-01-01T00:00:00Z
```

### Import Large Historical Dataset

For large imports, use a smaller batch size to be gentler on the API:

```bash
./bin/import-history \
  -service pagerduty \
  -since 2020-01-01T00:00:00Z \
  -batch-size 50
```

## Output

The tool provides progress updates and final statistics:

```
Starting import from pagerduty
Date range: 2024-01-01T00:00:00Z to 2024-12-31T23:59:59Z
Team filter: [P123ABC]

Fetched 100 incidents/alerts (offset: 0)
  Imported: PXXXXXX - Database Connection Pool Exhausted (Team: Backend)
  Imported: PXXXXXX - API Latency Spike (Team: Backend)
  Skipping PXXXXXX - already exists
...

Import completed!
Total incidents/alerts fetched: 247
New outages created: 235
New alerts created: 235
Skipped (already exists): 12
```

## Troubleshooting

### API Rate Limiting

If you encounter rate limiting errors:
- Use a smaller `-batch-size` (e.g., 25 or 50)
- The tool includes automatic 500ms delays between batches
- Split your import into smaller date ranges

### Database Connection Issues

Ensure:
- PostgreSQL is running
- Database migrations are applied
- Database credentials in `config.yaml` are correct
- Database is reachable from your machine

### Invalid API Key

If you see authentication errors:
- Verify your API key is correct
- Check that the API key has necessary permissions
- Ensure the API key is properly configured in `config.yaml` or environment variables

### Memory Usage

For very large imports, monitor memory usage. The tool processes incidents in batches, so memory usage should remain relatively constant.

## Best Practices

1. **Start with Dry Run**: Always test with `-dry-run` first to verify the import scope
2. **Use Team Filters**: If you only need specific teams, filter to reduce API calls and import time
3. **Import in Chunks**: For multi-year imports, consider breaking into smaller date ranges
4. **Monitor Progress**: Keep an eye on the output to catch any issues early
5. **Backup Database**: Consider backing up your database before large imports

## API Permissions Required

### PagerDuty
- Read access to incidents
- Read access to teams

### OpsGenie
- Read access to alerts
- Read access to teams
