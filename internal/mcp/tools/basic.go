package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/MachineLearning-Nerd/lazydb/internal/mcp/server"
)

// BasicTools provides fundamental database schema access tools
type BasicTools struct {
	connGetter server.ConnectionGetter
}

// NewBasicTools creates a new BasicTools instance
func NewBasicTools(connGetter server.ConnectionGetter) *BasicTools {
	return &BasicTools{connGetter: connGetter}
}

// Register registers all basic tools with the tool registry
func (t *BasicTools) Register(registry *server.ToolRegistry) {
	// Tool 1: list_all_tables
	registry.Register(
		server.Tool{
			Name:        "list_all_tables",
			Description: "Get a complete list of all available tables in the database, grouped by schema. Returns a lightweight reference of all tables.",
			InputSchema: server.InputSchema{
				Type:       "object",
				Properties: map[string]server.Property{},
			},
		},
		t.listAllTables,
	)

	// Tool 2: get_table_schema
	registry.Register(
		server.Tool{
			Name:        "get_table_schema",
			Description: "Get detailed schema information for a specific table including columns, data types, nullable constraints, and default values.",
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"table_name": {
						Type:        "string",
						Description: "Name of the table (format: 'schema.table' or just 'table' for public schema)",
					},
					"include_constraints": {
						Type:        "boolean",
						Description: "Include foreign keys and constraints information",
						Default:     true,
					},
				},
				Required: []string{"table_name"},
			},
		},
		t.getTableSchema,
	)

	// Tool 3: search_tables
	registry.Register(
		server.Tool{
			Name:        "search_tables",
			Description: "Search for tables matching a pattern using SQL LIKE syntax. Useful for discovering tables related to a topic.",
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"pattern": {
						Type:        "string",
						Description: "Search pattern using SQL LIKE syntax (e.g., 'user%', '%order%', '%payment%')",
					},
					"schema": {
						Type:        "string",
						Description: "Optional: filter by specific schema name",
					},
				},
				Required: []string{"pattern"},
			},
		},
		t.searchTables,
	)

	// Tool 4: get_sample_data
	registry.Register(
		server.Tool{
			Name:        "get_sample_data",
			Description: "Get sample rows from a table to understand data patterns and actual values. Limited to 10 rows maximum.",
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"table_name": {
						Type:        "string",
						Description: "Table name (format: 'schema.table' or just 'table')",
					},
					"limit": {
						Type:        "integer",
						Description: "Number of sample rows to retrieve (max 10)",
						Default:     5,
					},
				},
				Required: []string{"table_name"},
			},
		},
		t.getSampleData,
	)

	// Tool 5: get_table_count
	registry.Register(
		server.Tool{
			Name:        "get_table_count",
			Description: "Get the total number of rows in a table. Useful for understanding table size.",
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"table_name": {
						Type:        "string",
						Description: "Table name (format: 'schema.table' or just 'table')",
					},
				},
				Required: []string{"table_name"},
			},
		},
		t.getTableCount,
	)
}

// listAllTables returns all tables grouped by schema
func (t *BasicTools) listAllTables(ctx context.Context, args map[string]interface{}) (string, error) {
	// Get current connection
	conn, err := t.connGetter()
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	// Get all schemas
	schemas, err := conn.ListSchemas(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list schemas: %w", err)
	}

	result := make(map[string][]string)

	for _, schema := range schemas {
		// Skip system schemas for cleaner output
		if schema == "pg_catalog" || schema == "information_schema" ||
		   strings.HasPrefix(schema, "pg_toast") || strings.HasPrefix(schema, "pg_temp") {
			continue
		}

		tables, err := conn.ListTables(ctx, schema)
		if err != nil {
			continue
		}

		tableNames := make([]string, len(tables))
		for i, table := range tables {
			tableNames[i] = table.Name
		}

		if len(tableNames) > 0 {
			result[schema] = tableNames
		}
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// getTableSchema returns detailed schema for a specific table
func (t *BasicTools) getTableSchema(ctx context.Context, args map[string]interface{}) (string, error) {
	// Get current connection
	conn, err := t.connGetter()
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	tableName, ok := args["table_name"].(string)
	if !ok {
		return "", fmt.Errorf("table_name parameter is required")
	}

	includeConstraints := true
	if val, ok := args["include_constraints"].(bool); ok {
		includeConstraints = val
	}

	// Parse schema.table
	schema, table := parseTableName(tableName)

	// Get columns
	columns, err := conn.GetTableColumns(ctx, schema, table)
	if err != nil {
		return "", fmt.Errorf("failed to get columns for %s.%s: %w", schema, table, err)
	}

	result := map[string]interface{}{
		"table":  table,
		"schema": schema,
		"columns": func() []map[string]interface{} {
			cols := make([]map[string]interface{}, len(columns))
			for i, col := range columns {
				cols[i] = map[string]interface{}{
					"name":     col.Name,
					"type":     col.Type,
					"nullable": col.Nullable,
					"default":  col.Default,
				}
			}
			return cols
		}(),
		"include_constraints": includeConstraints,
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// searchTables searches for tables matching a pattern
func (t *BasicTools) searchTables(ctx context.Context, args map[string]interface{}) (string, error) {
	// Get current connection
	conn, err := t.connGetter()
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	pattern, ok := args["pattern"].(string)
	if !ok {
		return "", fmt.Errorf("pattern parameter is required")
	}

	filterSchema, _ := args["schema"].(string)

	// Get all schemas
	schemas, err := conn.ListSchemas(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list schemas: %w", err)
	}

	var matches []string

	for _, schema := range schemas {
		// Skip system schemas
		if schema == "pg_catalog" || schema == "information_schema" {
			continue
		}

		// Filter by schema if specified
		if filterSchema != "" && schema != filterSchema {
			continue
		}

		tables, err := conn.ListTables(ctx, schema)
		if err != nil {
			continue
		}

		for _, table := range tables {
			// Simple pattern matching (convert SQL LIKE to Go)
			if matchPattern(table.Name, pattern) {
				matches = append(matches, fmt.Sprintf("%s.%s", schema, table.Name))
			}
		}
	}

	output, err := json.MarshalIndent(matches, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// getSampleData retrieves sample rows from a table
func (t *BasicTools) getSampleData(ctx context.Context, args map[string]interface{}) (string, error) {
	// Get current connection
	conn, err := t.connGetter()
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	tableName, ok := args["table_name"].(string)
	if !ok {
		return "", fmt.Errorf("table_name parameter is required")
	}

	limit := 5
	if val, ok := args["limit"].(float64); ok {
		limit = int(val)
		if limit > 10 {
			limit = 10
		}
		if limit < 1 {
			limit = 1
		}
	}

	schema, table := parseTableName(tableName)

	// Build and execute sample query
	query := fmt.Sprintf("SELECT * FROM %s.%s LIMIT %d", schema, table, limit)

	result, err := conn.ExecuteQuery(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to get sample data: %w", err)
	}

	// Convert rows to JSON
	rows := make([]map[string]interface{}, len(result.Rows))
	for i, row := range result.Rows {
		rowMap := make(map[string]interface{})
		for j, col := range result.Columns {
			if j < len(row) {
				rowMap[col] = row[j]
			}
		}
		rows[i] = rowMap
	}

	output, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// getTableCount returns the number of rows in a table
func (t *BasicTools) getTableCount(ctx context.Context, args map[string]interface{}) (string, error) {
	// Get current connection
	conn, err := t.connGetter()
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	tableName, ok := args["table_name"].(string)
	if !ok {
		return "", fmt.Errorf("table_name parameter is required")
	}

	schema, table := parseTableName(tableName)

	// Execute count query
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s.%s", schema, table)

	result, err := conn.ExecuteQuery(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to get table count: %w", err)
	}

	count := "0"
	if len(result.Rows) > 0 && len(result.Rows[0]) > 0 {
		count = result.Rows[0][0]
	}

	output, err := json.MarshalIndent(map[string]interface{}{
		"table":  table,
		"schema": schema,
		"count":  count,
	}, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
}

// Helper functions

// parseTableName splits "schema.table" or returns "public.table"
func parseTableName(name string) (schema, table string) {
	parts := strings.Split(name, ".")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "public", name
}

// matchPattern performs simple SQL LIKE pattern matching
func matchPattern(text, pattern string) bool {
	// Convert SQL LIKE pattern to simple substring matching
	// % = any characters, _ = single character (simplified)
	pattern = strings.ToLower(pattern)
	text = strings.ToLower(text)

	// Remove % wildcards and check if substring exists
	if strings.HasPrefix(pattern, "%") && strings.HasSuffix(pattern, "%") {
		// %text% = contains
		return strings.Contains(text, strings.Trim(pattern, "%"))
	} else if strings.HasPrefix(pattern, "%") {
		// %text = ends with
		return strings.HasSuffix(text, strings.TrimPrefix(pattern, "%"))
	} else if strings.HasSuffix(pattern, "%") {
		// text% = starts with
		return strings.HasPrefix(text, strings.TrimSuffix(pattern, "%"))
	}

	// Exact match
	return text == pattern
}
