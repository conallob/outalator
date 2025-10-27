# MCP Server Integration

The Outalator MCP (Model Context Protocol) server provides a standardized interface for AI assistants to interact with the outage database.

## What is MCP?

The Model Context Protocol (MCP) is an open protocol that enables AI assistants to connect to external data sources and tools. The MCP server exposes Outalator's functionality through a standardized JSON-RPC interface over stdio.

## Features

The MCP server provides these tools to AI assistants:

1. **list_outages**: List all outages with pagination
2. **get_outage**: Get details of a specific outage by ID
3. **create_outage**: Create a new outage entry
4. **add_note**: Add a note to an existing outage
5. **update_outage**: Update an existing outage's status, severity, etc.

## Running the MCP Server

### Build the Server

```bash
go build -o mcp-server ./cmd/mcp-server
```

### Run the Server

```bash
./mcp-server -config config.yaml
```

The server communicates via stdin/stdout using the MCP protocol.

## Configuration

The MCP server uses the same configuration file as the main Outalator application:

```yaml
database:
  host: localhost
  port: 5432
  user: outalator
  password: outalator
  dbname: outalator
  sslmode: disable

# Optional: PagerDuty and OpsGenie integrations
pagerduty:
  api_key: your-pagerduty-api-key

opsgenie:
  api_key: your-opsgenie-api-key
```

## Using with Claude Desktop

To use the MCP server with Claude Desktop, add this to your Claude Desktop configuration:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`

**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "outalator": {
      "command": "/path/to/outalator/mcp-server",
      "args": ["-config", "/path/to/config.yaml"]
    }
  }
}
```

After adding this configuration, restart Claude Desktop. The Outalator tools will be available to Claude.

## Available Tools

### list_outages

List all outages with pagination support.

**Parameters:**
- `limit` (number, optional): Maximum number of outages to return (default: 50, max: 100)
- `offset` (number, optional): Offset for pagination (default: 0)

**Example:**
```json
{
  "name": "list_outages",
  "arguments": {
    "limit": 10,
    "offset": 0
  }
}
```

### get_outage

Get details of a specific outage by ID.

**Parameters:**
- `outage_id` (string, required): UUID of the outage

**Example:**
```json
{
  "name": "get_outage",
  "arguments": {
    "outage_id": "123e4567-e89b-12d3-a456-426614174000"
  }
}
```

### create_outage

Create a new outage entry.

**Parameters:**
- `title` (string, required): Title of the outage
- `description` (string, required): Detailed description
- `severity` (string, required): Severity level (critical, high, medium, low)

**Example:**
```json
{
  "name": "create_outage",
  "arguments": {
    "title": "API Gateway Down",
    "description": "Users cannot authenticate",
    "severity": "critical"
  }
}
```

### add_note

Add a note to an existing outage.

**Parameters:**
- `outage_id` (string, required): UUID of the outage
- `content` (string, required): Content of the note
- `author` (string, required): Author of the note
- `format` (string, optional): Format of the note (plaintext or markdown, default: plaintext)

**Example:**
```json
{
  "name": "add_note",
  "arguments": {
    "outage_id": "123e4567-e89b-12d3-a456-426614174000",
    "content": "Restarted API gateway service",
    "author": "alice@example.com",
    "format": "plaintext"
  }
}
```

### update_outage

Update an existing outage.

**Parameters:**
- `outage_id` (string, required): UUID of the outage
- `title` (string, optional): New title
- `description` (string, optional): New description
- `status` (string, optional): New status (open, investigating, resolved, closed)
- `severity` (string, optional): New severity (critical, high, medium, low)

**Example:**
```json
{
  "name": "update_outage",
  "arguments": {
    "outage_id": "123e4567-e89b-12d3-a456-426614174000",
    "status": "resolved"
  }
}
```

## Protocol Details

The MCP server implements the Model Context Protocol version 2024-11-05.

### Request Format

```json
{
  "method": "tools/call",
  "params": {
    "name": "tool_name",
    "arguments": {
      "param1": "value1"
    }
  },
  "id": 1
}
```

### Response Format

```json
{
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Operation completed successfully"
      }
    ],
    "data": {}
  },
  "id": 1
}
```

### Error Format

```json
{
  "error": {
    "code": -32603,
    "message": "Error description"
  },
  "id": 1
}
```

## Architecture

The MCP server consists of:

- **`internal/mcp/server.go`**: MCP protocol implementation
- **`cmd/mcp-server/main.go`**: Server entry point

The server:
1. Reads JSON-RPC requests from stdin
2. Validates requests according to MCP specification
3. Calls the appropriate service methods
4. Returns JSON-RPC responses on stdout
5. Logs errors to stderr

## Use Cases

### AI-Assisted Incident Response

AI assistants can:
- Create outages automatically from alert descriptions
- Add troubleshooting notes as they guide engineers
- Update outage status as incidents progress
- Search historical outages for similar issues

### Natural Language Queries

Ask Claude:
- "What outages have we had in the past week?"
- "Create an outage for the database connection issues we're seeing"
- "Add a note that we restarted the Redis cluster"
- "What was the resolution for that API timeout outage last month?"

### Automated Documentation

AI can:
- Generate post-incident reports from outage data
- Summarize troubleshooting steps across similar outages
- Create runbooks based on resolution patterns

## Security Considerations

- The MCP server has full database access through the service layer
- Runs locally on the same machine as the AI assistant
- No network exposure (stdio-based communication)
- Should use the same database credentials as the main application
- Consider running with restricted database user permissions if needed

## Troubleshooting

### Server won't start

1. Check database connectivity
2. Verify config.yaml path is correct
3. Ensure PostgreSQL is running
4. Check database credentials

### Tools not appearing in Claude Desktop

1. Verify the path to mcp-server binary is correct
2. Check that the binary has execute permissions
3. Restart Claude Desktop after configuration changes
4. Check Claude Desktop logs for errors

### Database errors

1. Ensure database migrations have been run
2. Verify database user has necessary permissions
3. Check database connection string in config.yaml

## Future Enhancements

Potential improvements:

- Support for HTTP transport (in addition to stdio)
- Read-only mode for safer AI access
- Rate limiting for AI requests
- Audit logging of AI operations
- Additional tools for tags, alerts, and analytics
