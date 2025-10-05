# LazyDB MCP Server Setup Guide

## Overview

The LazyDB MCP (Model Context Protocol) server provides AI assistants with smart access to your database schema and data. It integrates with Claude Code, Gemini CLI, Qwen Code, and GitHub Copilot.

## Prerequisites

- LazyDB installed and configured
- At least one database connection saved in LazyDB
- Claude Code (or another MCP-compatible AI CLI tool)
- Go 1.21+ (for building from source)

## Installation

### 1. Build the MCP Server

```bash
cd /Users/dineshjinjala/Documents/AllCode/LazyDB
go build -o bin/lazydb-mcp ./cmd/lazydb-mcp
```

### 2. Verify the Binary

```bash
./bin/lazydb-mcp --help
```

## Configuration

### For Claude Code

1. **Locate your Claude Code config file:**
   ```bash
   ~/.claude/settings.json
   ```

2. **Edit the MCP configuration:**

   Add the `mcpServers` section to your `~/.claude/settings.json`:
   ```json
   {
     "mcpServers": {
       "lazydb": {
         "command": "/Users/dineshjinjala/Documents/AllCode/LazyDB/bin/lazydb-mcp",
         "args": [
           "--verbose"
         ]
       }
     }
   }
   ```

   **Note:** If your settings.json already has other settings, just add the `mcpServers` section to the existing JSON object.

3. **Restart Claude Code** to load the new MCP server configuration.

### Command-Line Options

```bash
lazydb-mcp [flags]

Flags:
  --connection <name>   Specify which LazyDB connection to use
  --verbose             Enable detailed logging to stderr
  --help                Show help message
```

**Examples:**
```bash
# Use active connection (default)
lazydb-mcp

# Use specific connection
lazydb-mcp --connection production_db

# Enable verbose logging
lazydb-mcp --verbose
```

## Available Tools

The MCP server provides 5 basic tools:

### 1. list_all_tables
Get a complete list of all tables grouped by schema.

**Parameters:** None

**Example Use:**
> "Show me all tables in the database"

### 2. get_table_schema
Get detailed schema information for a specific table.

**Parameters:**
- `table_name` (string, required): Table name (format: 'schema.table' or 'table')
- `include_constraints` (boolean, optional): Include foreign keys and constraints (default: true)

**Example Use:**
> "Show me the schema for the users table"

### 3. search_tables
Search for tables matching a pattern.

**Parameters:**
- `pattern` (string, required): Search pattern using SQL LIKE syntax (e.g., 'user%', '%order%')
- `schema` (string, optional): Filter by specific schema

**Example Use:**
> "Find all tables related to orders"

### 4. get_sample_data
Get sample rows from a table to understand data patterns.

**Parameters:**
- `table_name` (string, required): Table name
- `limit` (integer, optional): Number of rows (max 10, default 5)

**Example Use:**
> "Show me some sample data from the users table"

### 5. get_table_count
Get the total number of rows in a table.

**Parameters:**
- `table_name` (string, required): Table name

**Example Use:**
> "How many rows are in the orders table?"

## Testing

### 1. Test with Claude Code

Start a conversation with Claude Code and try these queries:

**Basic Discovery:**
```
"What tables are available in my database?"
```

**Schema Exploration:**
```
"Show me the schema for the users table"
```

**Pattern Search:**
```
"Find all tables related to authentication"
```

**Data Sampling:**
```
"Show me sample data from the orders table"
```

**Complex Query:**
```
"I want to write a query to find all orders for a specific user.
What tables do I need and how are they related?"
```

### 2. Check MCP Server Logs

The MCP server logs to stderr (with `--verbose` flag):

```bash
# Watch logs in a separate terminal
tail -f ~/.lazydb/mcp-server.log
```

### 3. Verify Tools Are Loaded

When Claude Code starts, it should show:
```
Loading MCP servers...
✓ lazydb: 5 tools available
```

## Troubleshooting

### Issue: MCP server not loading

**Solution:**
1. Check Claude Code config file syntax (must be valid JSON)
2. Verify binary path is correct
3. Ensure binary has execute permissions: `chmod +x bin/lazydb-mcp`
4. Check logs with `--verbose` flag

### Issue: "No active connection found"

**Solution:**
1. Ensure you have connections saved in LazyDB
2. Run LazyDB TUI and set an active connection
3. Or specify connection explicitly: `--connection <name>`

### Issue: Permission denied

**Solution:**
```bash
chmod +x bin/lazydb-mcp
```

### Issue: Database connection fails

**Solution:**
1. Test connection in LazyDB TUI first
2. Check connection credentials
3. Verify database is accessible
4. Check firewall/network settings

## Advanced Configuration

### Using Specific Connection

To always use a specific connection, modify the config:

```json
{
  "mcpServers": {
    "lazydb": {
      "command": "/path/to/lazydb-mcp",
      "args": [
        "--connection", "production_db",
        "--verbose"
      ]
    }
  }
}
```

### Multiple Database Servers

You can configure multiple MCP servers for different databases:

```json
{
  "mcpServers": {
    "lazydb-production": {
      "command": "/path/to/lazydb-mcp",
      "args": ["--connection", "production_db"]
    },
    "lazydb-staging": {
      "command": "/path/to/lazydb-mcp",
      "args": ["--connection", "staging_db"]
    }
  }
}
```

## Next Steps

This is Phase 1 (Basic Tools). Coming in Phase 3:

- **Smart Context Tool**: AI-powered query analysis for intelligent schema selection
- **Relationship Mapping**: Automatic foreign key traversal
- **Index Information**: Performance optimization suggestions
- **Sample Data Analysis**: Data pattern recognition

## Support

For issues or questions:
1. Check the troubleshooting section above
2. Review MCP server logs (with `--verbose`)
3. Test connection in LazyDB TUI first
4. Refer to the implementation plan: `docs/mcp-server-implementation-plan.md`
