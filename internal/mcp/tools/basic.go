package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/MachineLearning-Nerd/lazydb/internal/db"
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
			Description: "Lists all tables grouped by schema. Use for database overview.",
			Category:    "schema",
			Tags:        []string{"schema", "tables", "discovery", "overview"},
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
			Description: "Get table columns, types, nullability, defaults. Use for structure analysis.",
			Category:    "schema",
			Tags:        []string{"schema", "columns", "structure", "types"},
			InputExamples: []map[string]interface{}{
				{"table_name": "users"},
				{"table_name": "public.orders", "include_constraints": true},
				{"table_name": "sales.customers"},
			},
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"table_name": {
						Type:     "string",
						Format:   "schema.table or table",
						Examples: []string{"users", "public.orders", "sales.customers"},
					},
					"include_constraints": {
						Type:    "boolean",
						Default: true,
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
			Description: "Search tables by pattern (SQL LIKE). Use for discovery.",
			Category:    "discovery",
			Tags:        []string{"search", "discovery", "pattern", "find"},
			InputExamples: []map[string]interface{}{
				{"pattern": "user%"},
				{"pattern": "%order%"},
				{"pattern": "%payment%", "schema": "sales"},
			},
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"pattern": {
						Type:     "string",
						Examples: []string{"user%", "%order%", "%_log"},
					},
					"schema": {
						Type: "string",
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
			Description: "Get 1-10 sample rows. Use to understand data patterns.",
			Category:    "discovery",
			Tags:        []string{"sample", "data", "preview", "rows"},
			InputExamples: []map[string]interface{}{
				{"table_name": "users"},
				{"table_name": "orders", "limit": 10},
			},
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"table_name": {
						Type:     "string",
						Format:   "schema.table or table",
						Examples: []string{"users", "public.orders"},
					},
					"limit": {
						Type:    "integer",
						Default: 5,
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
			Description: "Get row count for a table.",
			Category:    "discovery",
			Tags:        []string{"count", "rows", "size"},
			InputExamples: []map[string]interface{}{
				{"table_name": "users"},
				{"table_name": "public.orders"},
			},
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"table_name": {
						Type:     "string",
						Format:   "schema.table or table",
						Examples: []string{"users", "orders"},
					},
				},
				Required: []string{"table_name"},
			},
		},
		t.getTableCount,
	)

	// Tool 6: execute_query
	registry.Register(
		server.Tool{
			Name:        "execute_query",
			Description: "Execute a read-only SQL query. Only SELECT and WITH (CTE) statements are allowed. INSERT, UPDATE, DELETE, and DDL are rejected. Returns JSON array of row objects. Auto-limits to 50 rows if no LIMIT specified (max 500).",
			Category:    "schema",
			Tags:        []string{"query", "select", "execute", "sql", "read"},
			InputExamples: []map[string]interface{}{
				{"query": "SELECT * FROM users LIMIT 10"},
				{"query": "SELECT id, name FROM products WHERE price > 100"},
				{"query": "WITH recent AS (SELECT * FROM orders WHERE created_at > now() - interval '7 days') SELECT * FROM recent"},
				{"query": "SELECT * FROM users", "limit": 100},
			},
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"query": {
						Type:        "string",
						Description: "SQL SELECT query to execute. Only SELECT and WITH (CTE) statements allowed.",
						Examples:    []string{"SELECT * FROM users LIMIT 10", "WITH cte AS (...) SELECT * FROM cte"},
					},
					"limit": {
						Type:        "integer",
						Description: "Maximum rows to return. Default 50, max 500. Overrides LIMIT in query if lower.",
						Default:     50,
					},
				},
				Required: []string{"query"},
			},
		},
		t.executeQuery,
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

// executeQuery executes a read-only SQL query and returns results as JSON
func (t *BasicTools) executeQuery(ctx context.Context, args map[string]interface{}) (string, error) {
	conn, err := t.connGetter()
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	query, ok := args["query"].(string)
	if !ok || strings.TrimSpace(query) == "" {
		return "", fmt.Errorf("query parameter is required")
	}

	// Validate read-only using AST-based validator
	if err := db.IsReadOnlyQuery(query); err != nil {
		return "", fmt.Errorf("query rejected: %w", err)
	}

	// Parse limit parameter
	maxRows := 50
	if val, ok := args["limit"].(float64); ok {
		maxRows = int(val)
	}
	if maxRows < 1 {
		maxRows = 1
	}
	if maxRows > 500 {
		maxRows = 500
	}

	// Inject LIMIT if outermost SELECT has no LIMIT clause (AST-based check)
	if !db.HasOuterLimit(query) {
		query = fmt.Sprintf("%s LIMIT %d", strings.TrimRight(strings.TrimSpace(query), ";"), maxRows)
	}

	result, err := conn.ExecuteQuery(ctx, query)
	if err != nil {
		return "", fmt.Errorf("query execution failed: %w", err)
	}

	// Convert to JSON array of objects
	rows := make([]map[string]interface{}, 0, len(result.Rows))
	for _, row := range result.Rows {
		rowMap := make(map[string]interface{}, len(result.Columns))
		for j, col := range result.Columns {
			if j < len(row) {
				rowMap[col] = row[j]
			}
		}
		rows = append(rows, rowMap)
	}

	response := map[string]interface{}{
		"columns":      result.Columns,
		"rows":         rows,
		"row_count":    len(rows),
		"execution_ms": result.ExecutionMs,
	}

	output, err := json.MarshalIndent(response, "", "  ")
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
