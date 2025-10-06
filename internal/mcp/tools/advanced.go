package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/MachineLearning-Nerd/lazydb/internal/mcp/server"
)

// AdvancedTools provides advanced database inspection and analysis tools
type AdvancedTools struct {
	connGetter server.ConnectionGetter
}

// NewAdvancedTools creates a new AdvancedTools instance
func NewAdvancedTools(connGetter server.ConnectionGetter) *AdvancedTools {
	return &AdvancedTools{connGetter: connGetter}
}

// Register registers all advanced tools with the tool registry
func (t *AdvancedTools) Register(registry *server.ToolRegistry) {
	// Category 1: Schema Definition (DDL) Tools
	t.registerDDLTools(registry)

	// Category 2: Index & Performance Tools
	t.registerIndexTools(registry)

	// Category 3: Relationship & Constraint Tools
	t.registerRelationshipTools(registry)

	// Category 4: Trigger & Event Tools
	t.registerTriggerTools(registry)

	// Category 5: Statistics & Analysis Tools
	t.registerStatsTools(registry)

	// Category 6: Advanced Discovery Tools
	t.registerDiscoveryTools(registry)
}

// ============================================================================
// Category 1: Schema Definition (DDL) Tools
// ============================================================================

func (t *AdvancedTools) registerDDLTools(registry *server.ToolRegistry) {
	// Tool 1: get_table_ddl
	registry.Register(
		server.Tool{
			Name:        "get_table_ddl",
			Description: "Generate CREATE TABLE DDL statement including columns, constraints, and optionally indexes. Reconstructs the table definition from system catalogs.",
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"table_name": {
						Type:        "string",
						Description: "Table name (format: 'schema.table' or just 'table' for public schema)",
					},
					"include_indexes": {
						Type:        "boolean",
						Description: "Include CREATE INDEX statements in the DDL",
						Default:     true,
					},
				},
				Required: []string{"table_name"},
			},
		},
		t.getTableDDL,
	)

	// Tool 2: get_view_definition
	registry.Register(
		server.Tool{
			Name:        "get_view_definition",
			Description: "Get the SQL definition (SELECT statement) for a view.",
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"view_name": {
						Type:        "string",
						Description: "View name (format: 'schema.view' or just 'view')",
					},
				},
				Required: []string{"view_name"},
			},
		},
		t.getViewDefinition,
	)

	// Tool 3: get_function_definition
	registry.Register(
		server.Tool{
			Name:        "get_function_definition",
			Description: "Get the complete source code and definition for a function or stored procedure.",
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"function_name": {
						Type:        "string",
						Description: "Function name",
					},
					"schema": {
						Type:        "string",
						Description: "Schema name (defaults to 'public')",
						Default:     "public",
					},
				},
				Required: []string{"function_name"},
			},
		},
		t.getFunctionDefinition,
	)
}

// ============================================================================
// Category 2: Index & Performance Tools
// ============================================================================

func (t *AdvancedTools) registerIndexTools(registry *server.ToolRegistry) {
	// Tool 4: get_table_indexes
	registry.Register(
		server.Tool{
			Name:        "get_table_indexes",
			Description: "List all indexes on a table including index type, columns, uniqueness, and definition.",
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"table_name": {
						Type:        "string",
						Description: "Table name (format: 'schema.table' or just 'table')",
					},
					"include_definition": {
						Type:        "boolean",
						Description: "Include full CREATE INDEX statement",
						Default:     true,
					},
				},
				Required: []string{"table_name"},
			},
		},
		t.getTableIndexes,
	)

	// Tool 5: get_table_size
	registry.Register(
		server.Tool{
			Name:        "get_table_size",
			Description: "Get physical size of a table including row count, disk usage, index sizes, and bloat estimation.",
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"table_name": {
						Type:        "string",
						Description: "Table name (format: 'schema.table' or just 'table')",
					},
					"include_indexes": {
						Type:        "boolean",
						Description: "Include individual index sizes",
						Default:     true,
					},
				},
				Required: []string{"table_name"},
			},
		},
		t.getTableSize,
	)

	// Tool 6: explain_query
	registry.Register(
		server.Tool{
			Name:        "explain_query",
			Description: "Run EXPLAIN or EXPLAIN ANALYZE on a SELECT query to understand query execution plan and performance. Only SELECT queries are allowed for safety.",
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"query": {
						Type:        "string",
						Description: "SELECT query to explain",
					},
					"analyze": {
						Type:        "boolean",
						Description: "Run EXPLAIN ANALYZE (actually executes the query)",
						Default:     false,
					},
					"format": {
						Type:        "string",
						Description: "Output format: 'text' or 'json'",
						Default:     "text",
					},
				},
				Required: []string{"query"},
			},
		},
		t.explainQuery,
	)
}

// ============================================================================
// Category 3: Relationship & Constraint Tools
// ============================================================================

func (t *AdvancedTools) registerRelationshipTools(registry *server.ToolRegistry) {
	// Tool 7: get_foreign_keys
	registry.Register(
		server.Tool{
			Name:        "get_foreign_keys",
			Description: "Get all foreign key relationships for a table, including both incoming (tables referencing this table) and outgoing (tables this table references).",
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"table_name": {
						Type:        "string",
						Description: "Table name (format: 'schema.table' or just 'table')",
					},
					"direction": {
						Type:        "string",
						Description: "Direction: 'incoming', 'outgoing', or 'both'",
						Default:     "both",
					},
				},
				Required: []string{"table_name"},
			},
		},
		t.getForeignKeys,
	)

	// Tool 8: get_table_constraints
	registry.Register(
		server.Tool{
			Name:        "get_table_constraints",
			Description: "Get all constraints on a table including PRIMARY KEY, FOREIGN KEY, UNIQUE, CHECK, and NOT NULL constraints.",
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"table_name": {
						Type:        "string",
						Description: "Table name (format: 'schema.table' or just 'table')",
					},
					"constraint_type": {
						Type:        "string",
						Description: "Filter by type: 'PRIMARY KEY', 'FOREIGN KEY', 'UNIQUE', 'CHECK', or 'all'",
						Default:     "all",
					},
				},
				Required: []string{"table_name"},
			},
		},
		t.getTableConstraints,
	)

	// Tool 9: get_table_dependencies
	registry.Register(
		server.Tool{
			Name:        "get_table_dependencies",
			Description: "Find all database objects (views, materialized views, functions) that depend on this table.",
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
		t.getTableDependencies,
	)
}

// ============================================================================
// Category 4: Trigger & Event Tools
// ============================================================================

func (t *AdvancedTools) registerTriggerTools(registry *server.ToolRegistry) {
	// Tool 10: get_table_triggers
	registry.Register(
		server.Tool{
			Name:        "get_table_triggers",
			Description: "List all triggers on a table including event type (INSERT/UPDATE/DELETE), timing (BEFORE/AFTER), and execution (FOR EACH ROW/STATEMENT).",
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
		t.getTableTriggers,
	)

	// Tool 11: get_trigger_definition
	registry.Register(
		server.Tool{
			Name:        "get_trigger_definition",
			Description: "Get the complete definition and function source code for a specific trigger.",
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"trigger_name": {
						Type:        "string",
						Description: "Trigger name",
					},
					"table_name": {
						Type:        "string",
						Description: "Table name where the trigger is defined",
					},
				},
				Required: []string{"trigger_name", "table_name"},
			},
		},
		t.getTriggerDefinition,
	)
}

// ============================================================================
// Category 5: Statistics & Analysis Tools
// ============================================================================

func (t *AdvancedTools) registerStatsTools(registry *server.ToolRegistry) {
	// Tool 12: get_column_stats
	registry.Register(
		server.Tool{
			Name:        "get_column_stats",
			Description: "Get statistical information about table columns including null percentage, distinct values, most common values, and data distribution.",
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"table_name": {
						Type:        "string",
						Description: "Table name (format: 'schema.table' or just 'table')",
					},
					"column_name": {
						Type:        "string",
						Description: "Specific column name (optional - if not provided, returns stats for all columns)",
					},
				},
				Required: []string{"table_name"},
			},
		},
		t.getColumnStats,
	)

	// Tool 13: get_table_stats
	registry.Register(
		server.Tool{
			Name:        "get_table_stats",
			Description: "Get table-level statistics including row count estimates, dead tuples, last vacuum/analyze times, and index usage.",
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
		t.getTableStats,
	)
}

// ============================================================================
// Category 6: Advanced Discovery Tools
// ============================================================================

func (t *AdvancedTools) registerDiscoveryTools(registry *server.ToolRegistry) {
	// Tool 14: list_sequences
	registry.Register(
		server.Tool{
			Name:        "list_sequences",
			Description: "List all sequences in the database with current value, increment, min/max values, and associated tables.",
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"schema": {
						Type:        "string",
						Description: "Filter by schema name (optional)",
					},
				},
			},
		},
		t.listSequences,
	)

	// Tool 15: list_materialized_views
	registry.Register(
		server.Tool{
			Name:        "list_materialized_views",
			Description: "List all materialized views with size, definition, and last refresh time.",
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"schema": {
						Type:        "string",
						Description: "Filter by schema name (optional)",
					},
				},
			},
		},
		t.listMaterializedViews,
	)

	// Tool 16: get_table_references
	registry.Register(
		server.Tool{
			Name:        "get_table_references",
			Description: "Find all tables that this table references (through foreign keys) and all tables that reference this table. Provides a complete relationship map.",
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
		t.getTableReferences,
	)
}

// ============================================================================
// Tool Implementations - Category 1: DDL Tools
// ============================================================================

func (t *AdvancedTools) getTableDDL(ctx context.Context, args map[string]interface{}) (string, error) {
	// Get current connection
	conn, err := t.connGetter()
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	tableName, ok := args["table_name"].(string)
	if !ok {
		return "", fmt.Errorf("table_name parameter is required")
	}

	includeIndexes := true
	if val, ok := args["include_indexes"].(bool); ok {
		includeIndexes = val
	}

	schema, table := parseTableName(tableName)

	// Get columns
	columns, err := conn.GetTableColumns(ctx, schema, table)
	if err != nil {
		return "", fmt.Errorf("failed to get columns: %w", err)
	}

	// Get constraints
	constraintsQuery := `
		SELECT
			conname,
			contype,
			pg_catalog.pg_get_constraintdef(oid, true) as definition
		FROM pg_catalog.pg_constraint
		WHERE conrelid = $1::regclass
		ORDER BY contype, conname
	`

	constraintsResult, err := conn.ExecuteQuery(ctx,
		strings.ReplaceAll(constraintsQuery, "$1", fmt.Sprintf("'%s.%s'", schema, table)))
	if err != nil {
		return "", fmt.Errorf("failed to get constraints: %w", err)
	}

	// Build DDL
	var ddl strings.Builder
	ddl.WriteString(fmt.Sprintf("CREATE TABLE %s.%s (\n", schema, table))

	// Add columns
	for i, col := range columns {
		nullability := "NULL"
		if !col.Nullable {
			nullability = "NOT NULL"
		}

		defaultVal := ""
		if col.Default != "" {
			defaultVal = fmt.Sprintf(" DEFAULT %s", col.Default)
		}

		ddl.WriteString(fmt.Sprintf("    %s %s%s %s", col.Name, col.Type, defaultVal, nullability))

		if i < len(columns)-1 || len(constraintsResult.Rows) > 0 {
			ddl.WriteString(",")
		}
		ddl.WriteString("\n")
	}

	// Add constraints
	for i, row := range constraintsResult.Rows {
		if len(row) >= 3 {
			ddl.WriteString(fmt.Sprintf("    CONSTRAINT %s %s", row[0], row[2]))
			if i < len(constraintsResult.Rows)-1 {
				ddl.WriteString(",")
			}
			ddl.WriteString("\n")
		}
	}

	ddl.WriteString(");\n")

	// Add indexes if requested
	if includeIndexes {
		indexesQuery := `
			SELECT indexdef
			FROM pg_indexes
			WHERE schemaname = $1 AND tablename = $2
			  AND indexname NOT IN (
				  SELECT conname FROM pg_constraint
				  WHERE conrelid = $3::regclass AND contype IN ('p', 'u')
			  )
		`
		indexesQuery = strings.ReplaceAll(indexesQuery, "$1", fmt.Sprintf("'%s'", schema))
		indexesQuery = strings.ReplaceAll(indexesQuery, "$2", fmt.Sprintf("'%s'", table))
		indexesQuery = strings.ReplaceAll(indexesQuery, "$3", fmt.Sprintf("'%s.%s'", schema, table))

		indexesResult, err := conn.ExecuteQuery(ctx, indexesQuery)
		if err == nil && len(indexesResult.Rows) > 0 {
			ddl.WriteString("\n-- Indexes\n")
			for _, row := range indexesResult.Rows {
				if len(row) > 0 {
					ddl.WriteString(row[0])
					ddl.WriteString(";\n")
				}
			}
		}
	}

	return ddl.String(), nil
}

func (t *AdvancedTools) getViewDefinition(ctx context.Context, args map[string]interface{}) (string, error) {
	// Get current connection
	conn, err := t.connGetter()
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	viewName, ok := args["view_name"].(string)
	if !ok {
		return "", fmt.Errorf("view_name parameter is required")
	}

	schema, view := parseTableName(viewName)

	query := fmt.Sprintf(`
		SELECT pg_catalog.pg_get_viewdef('%s.%s'::regclass, true) as definition
	`, schema, view)

	result, err := conn.ExecuteQuery(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to get view definition: %w", err)
	}

	if len(result.Rows) == 0 || len(result.Rows[0]) == 0 {
		return "", fmt.Errorf("view not found: %s.%s", schema, view)
	}

	output, _ := json.MarshalIndent(map[string]interface{}{
		"view":       view,
		"schema":     schema,
		"definition": result.Rows[0][0],
	}, "", "  ")

	return string(output), nil
}

func (t *AdvancedTools) getFunctionDefinition(ctx context.Context, args map[string]interface{}) (string, error) {
	// Get current connection
	conn, err := t.connGetter()
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	functionName, ok := args["function_name"].(string)
	if !ok {
		return "", fmt.Errorf("function_name parameter is required")
	}

	schema := "public"
	if val, ok := args["schema"].(string); ok && val != "" {
		schema = val
	}

	query := fmt.Sprintf(`
		SELECT
			p.proname as name,
			pg_catalog.pg_get_functiondef(p.oid) as definition,
			pg_catalog.pg_get_function_identity_arguments(p.oid) as arguments,
			t.typname as return_type
		FROM pg_catalog.pg_proc p
		JOIN pg_catalog.pg_namespace n ON n.oid = p.pronamespace
		LEFT JOIN pg_catalog.pg_type t ON t.oid = p.prorettype
		WHERE n.nspname = '%s' AND p.proname = '%s'
	`, schema, functionName)

	result, err := conn.ExecuteQuery(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to get function definition: %w", err)
	}

	if len(result.Rows) == 0 {
		return "", fmt.Errorf("function not found: %s.%s", schema, functionName)
	}

	row := result.Rows[0]
	output, _ := json.MarshalIndent(map[string]interface{}{
		"function":    row[0],
		"schema":      schema,
		"arguments":   row[2],
		"return_type": row[3],
		"definition":  row[1],
	}, "", "  ")

	return string(output), nil
}

// ============================================================================
// Tool Implementations - Category 2: Index & Performance Tools
// ============================================================================

func (t *AdvancedTools) getTableIndexes(ctx context.Context, args map[string]interface{}) (string, error) {
	// Get current connection
	conn, err := t.connGetter()
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	tableName, ok := args["table_name"].(string)
	if !ok {
		return "", fmt.Errorf("table_name parameter is required")
	}

	includeDefinition := true
	if val, ok := args["include_definition"].(bool); ok {
		includeDefinition = val
	}

	schema, table := parseTableName(tableName)

	query := fmt.Sprintf(`
		SELECT
			i.indexname,
			i.indexdef,
			am.amname as index_type,
			ix.indisunique as is_unique,
			ix.indisprimary as is_primary,
			pg_size_pretty(pg_relation_size(i.indexname::regclass)) as size
		FROM pg_indexes i
		JOIN pg_class c ON c.relname = i.indexname
		JOIN pg_index ix ON ix.indexrelid = c.oid
		JOIN pg_am am ON am.oid = c.relam
		WHERE i.schemaname = '%s' AND i.tablename = '%s'
		ORDER BY i.indexname
	`, schema, table)

	result, err := conn.ExecuteQuery(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to get indexes: %w", err)
	}

	indexes := make([]map[string]interface{}, len(result.Rows))
	for i, row := range result.Rows {
		index := map[string]interface{}{
			"name":       row[0],
			"type":       row[2],
			"is_unique":  row[3],
			"is_primary": row[4],
			"size":       row[5],
		}
		if includeDefinition && len(row) > 1 {
			index["definition"] = row[1]
		}
		indexes[i] = index
	}

	output, _ := json.MarshalIndent(map[string]interface{}{
		"table":   table,
		"schema":  schema,
		"indexes": indexes,
	}, "", "  ")

	return string(output), nil
}

func (t *AdvancedTools) getTableSize(ctx context.Context, args map[string]interface{}) (string, error) {
	// Get current connection
	conn, err := t.connGetter()
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	tableName, ok := args["table_name"].(string)
	if !ok {
		return "", fmt.Errorf("table_name parameter is required")
	}

	includeIndexes := true
	if val, ok := args["include_indexes"].(bool); ok {
		includeIndexes = val
	}

	schema, table := parseTableName(tableName)
	fullTableName := fmt.Sprintf("%s.%s", schema, table)

	query := fmt.Sprintf(`
		SELECT
			pg_size_pretty(pg_total_relation_size('%s')) as total_size,
			pg_size_pretty(pg_relation_size('%s')) as table_size,
			pg_size_pretty(pg_total_relation_size('%s') - pg_relation_size('%s')) as indexes_size,
			(SELECT COUNT(*) FROM %s) as row_count
	`, fullTableName, fullTableName, fullTableName, fullTableName, fullTableName)

	result, err := conn.ExecuteQuery(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to get table size: %w", err)
	}

	if len(result.Rows) == 0 {
		return "", fmt.Errorf("table not found: %s", fullTableName)
	}

	row := result.Rows[0]
	sizeInfo := map[string]interface{}{
		"table":        table,
		"schema":       schema,
		"total_size":   row[0],
		"table_size":   row[1],
		"indexes_size": row[2],
		"row_count":    row[3],
	}

	if includeIndexes {
		indexSizeQuery := fmt.Sprintf(`
			SELECT
				indexname,
				pg_size_pretty(pg_relation_size(indexname::regclass)) as size
			FROM pg_indexes
			WHERE schemaname = '%s' AND tablename = '%s'
		`, schema, table)

		indexResult, err := conn.ExecuteQuery(ctx, indexSizeQuery)
		if err == nil && len(indexResult.Rows) > 0 {
			indexSizes := make([]map[string]string, len(indexResult.Rows))
			for i, indexRow := range indexResult.Rows {
				indexSizes[i] = map[string]string{
					"name": indexRow[0],
					"size": indexRow[1],
				}
			}
			sizeInfo["individual_indexes"] = indexSizes
		}
	}

	output, _ := json.MarshalIndent(sizeInfo, "", "  ")
	return string(output), nil
}

func (t *AdvancedTools) explainQuery(ctx context.Context, args map[string]interface{}) (string, error) {
	// Get current connection
	conn, err := t.connGetter()
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	query, ok := args["query"].(string)
	if !ok {
		return "", fmt.Errorf("query parameter is required")
	}

	// Safety check: only allow SELECT queries
	trimmedQuery := strings.TrimSpace(strings.ToUpper(query))
	if !strings.HasPrefix(trimmedQuery, "SELECT") {
		return "", fmt.Errorf("only SELECT queries are allowed for EXPLAIN")
	}

	analyze := false
	if val, ok := args["analyze"].(bool); ok {
		analyze = val
	}

	format := "text"
	if val, ok := args["format"].(string); ok {
		format = val
	}

	explainCmd := "EXPLAIN"
	if analyze {
		explainCmd = "EXPLAIN ANALYZE"
	}

	if format == "json" {
		explainCmd += " (FORMAT JSON)"
	}

	fullQuery := fmt.Sprintf("%s %s", explainCmd, query)

	result, err := conn.ExecuteQuery(ctx, fullQuery)
	if err != nil {
		return "", fmt.Errorf("failed to explain query: %w", err)
	}

	// Format results
	var planOutput strings.Builder
	for _, row := range result.Rows {
		for _, col := range row {
			planOutput.WriteString(col)
			planOutput.WriteString("\n")
		}
	}

	output, _ := json.MarshalIndent(map[string]interface{}{
		"query":   query,
		"analyze": analyze,
		"format":  format,
		"plan":    planOutput.String(),
	}, "", "  ")

	return string(output), nil
}

// ============================================================================
// Tool Implementations - Category 3: Relationship & Constraint Tools
// ============================================================================

func (t *AdvancedTools) getForeignKeys(ctx context.Context, args map[string]interface{}) (string, error) {
	// Get current connection
	conn, err := t.connGetter()
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	tableName, ok := args["table_name"].(string)
	if !ok {
		return "", fmt.Errorf("table_name parameter is required")
	}

	direction := "both"
	if val, ok := args["direction"].(string); ok {
		direction = val
	}

	schema, table := parseTableName(tableName)

	var foreignKeys []map[string]interface{}

	// Outgoing FK (this table references other tables)
	if direction == "outgoing" || direction == "both" {
		outgoingQuery := fmt.Sprintf(`
			SELECT
				tc.constraint_name,
				kcu.column_name,
				ccu.table_schema AS foreign_schema,
				ccu.table_name AS foreign_table,
				ccu.column_name AS foreign_column
			FROM information_schema.table_constraints AS tc
			JOIN information_schema.key_column_usage AS kcu
				ON tc.constraint_name = kcu.constraint_name
				AND tc.table_schema = kcu.table_schema
			JOIN information_schema.constraint_column_usage AS ccu
				ON ccu.constraint_name = tc.constraint_name
				AND ccu.table_schema = tc.table_schema
			WHERE tc.constraint_type = 'FOREIGN KEY'
				AND tc.table_schema = '%s'
				AND tc.table_name = '%s'
		`, schema, table)

		result, err := conn.ExecuteQuery(ctx, outgoingQuery)
		if err == nil {
			for _, row := range result.Rows {
				foreignKeys = append(foreignKeys, map[string]interface{}{
					"direction":       "outgoing",
					"constraint_name": row[0],
					"column":          row[1],
					"references":      fmt.Sprintf("%s.%s(%s)", row[2], row[3], row[4]),
				})
			}
		}
	}

	// Incoming FK (other tables reference this table)
	if direction == "incoming" || direction == "both" {
		incomingQuery := fmt.Sprintf(`
			SELECT
				tc.constraint_name,
				tc.table_schema,
				tc.table_name,
				kcu.column_name,
				ccu.column_name AS referenced_column
			FROM information_schema.table_constraints AS tc
			JOIN information_schema.key_column_usage AS kcu
				ON tc.constraint_name = kcu.constraint_name
				AND tc.table_schema = kcu.table_schema
			JOIN information_schema.constraint_column_usage AS ccu
				ON ccu.constraint_name = tc.constraint_name
				AND ccu.table_schema = tc.table_schema
			WHERE tc.constraint_type = 'FOREIGN KEY'
				AND ccu.table_schema = '%s'
				AND ccu.table_name = '%s'
		`, schema, table)

		result, err := conn.ExecuteQuery(ctx, incomingQuery)
		if err == nil {
			for _, row := range result.Rows {
				foreignKeys = append(foreignKeys, map[string]interface{}{
					"direction":       "incoming",
					"constraint_name": row[0],
					"from_table":      fmt.Sprintf("%s.%s", row[1], row[2]),
					"from_column":     row[3],
					"references":      row[4],
				})
			}
		}
	}

	output, _ := json.MarshalIndent(map[string]interface{}{
		"table":        table,
		"schema":       schema,
		"foreign_keys": foreignKeys,
	}, "", "  ")

	return string(output), nil
}

func (t *AdvancedTools) getTableConstraints(ctx context.Context, args map[string]interface{}) (string, error) {
	// Get current connection
	conn, err := t.connGetter()
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	tableName, ok := args["table_name"].(string)
	if !ok {
		return "", fmt.Errorf("table_name parameter is required")
	}

	constraintType := "all"
	if val, ok := args["constraint_type"].(string); ok {
		constraintType = val
	}

	schema, table := parseTableName(tableName)
	fullTableName := fmt.Sprintf("%s.%s", schema, table)

	query := fmt.Sprintf(`
		SELECT
			conname,
			CASE contype
				WHEN 'c' THEN 'CHECK'
				WHEN 'f' THEN 'FOREIGN KEY'
				WHEN 'p' THEN 'PRIMARY KEY'
				WHEN 'u' THEN 'UNIQUE'
				WHEN 'x' THEN 'EXCLUDE'
			END AS constraint_type,
			pg_catalog.pg_get_constraintdef(oid, true) as definition
		FROM pg_catalog.pg_constraint
		WHERE conrelid = '%s'::regclass
		ORDER BY contype, conname
	`, fullTableName)

	result, err := conn.ExecuteQuery(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to get constraints: %w", err)
	}

	constraints := make([]map[string]interface{}, 0)
	for _, row := range result.Rows {
		if constraintType != "all" && row[1] != constraintType {
			continue
		}
		constraints = append(constraints, map[string]interface{}{
			"name":       row[0],
			"type":       row[1],
			"definition": row[2],
		})
	}

	output, _ := json.MarshalIndent(map[string]interface{}{
		"table":       table,
		"schema":      schema,
		"constraints": constraints,
	}, "", "  ")

	return string(output), nil
}

func (t *AdvancedTools) getTableDependencies(ctx context.Context, args map[string]interface{}) (string, error) {
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

	query := fmt.Sprintf(`
		SELECT DISTINCT
			dependent_ns.nspname as dependent_schema,
			dependent_view.relname as dependent_view,
			dependent_view.relkind as object_type
		FROM pg_depend
		JOIN pg_rewrite ON pg_depend.objid = pg_rewrite.oid
		JOIN pg_class as dependent_view ON pg_rewrite.ev_class = dependent_view.oid
		JOIN pg_class as source_table ON pg_depend.refobjid = source_table.oid
		JOIN pg_namespace dependent_ns ON dependent_ns.oid = dependent_view.relnamespace
		JOIN pg_namespace source_ns ON source_ns.oid = source_table.relnamespace
		WHERE source_ns.nspname = '%s'
			AND source_table.relname = '%s'
			AND dependent_view.relname != '%s'
			AND source_table.relkind = 'r'
			AND pg_depend.deptype = 'n'
		ORDER BY dependent_schema, dependent_view
	`, schema, table, table)

	result, err := conn.ExecuteQuery(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to get dependencies: %w", err)
	}

	dependencies := make([]map[string]interface{}, len(result.Rows))
	for i, row := range result.Rows {
		objectType := "unknown"
		switch row[2] {
		case "v":
			objectType = "view"
		case "m":
			objectType = "materialized view"
		case "f":
			objectType = "foreign table"
		}

		dependencies[i] = map[string]interface{}{
			"schema": row[0],
			"name":   row[1],
			"type":   objectType,
		}
	}

	output, _ := json.MarshalIndent(map[string]interface{}{
		"table":        table,
		"schema":       schema,
		"dependencies": dependencies,
	}, "", "  ")

	return string(output), nil
}

// ============================================================================
// Tool Implementations - Category 4: Trigger Tools
// ============================================================================

func (t *AdvancedTools) getTableTriggers(ctx context.Context, args map[string]interface{}) (string, error) {
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

	query := fmt.Sprintf(`
		SELECT
			trigger_name,
			event_manipulation as event,
			action_timing as timing,
			action_orientation as orientation,
			action_statement
		FROM information_schema.triggers
		WHERE event_object_schema = '%s'
			AND event_object_table = '%s'
		ORDER BY trigger_name
	`, schema, table)

	result, err := conn.ExecuteQuery(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to get triggers: %w", err)
	}

	triggers := make([]map[string]interface{}, len(result.Rows))
	for i, row := range result.Rows {
		triggers[i] = map[string]interface{}{
			"name":        row[0],
			"event":       row[1],
			"timing":      row[2],
			"orientation": row[3],
			"statement":   row[4],
		}
	}

	output, _ := json.MarshalIndent(map[string]interface{}{
		"table":    table,
		"schema":   schema,
		"triggers": triggers,
	}, "", "  ")

	return string(output), nil
}

func (t *AdvancedTools) getTriggerDefinition(ctx context.Context, args map[string]interface{}) (string, error) {
	// Get current connection
	conn, err := t.connGetter()
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	triggerName, ok := args["trigger_name"].(string)
	if !ok {
		return "", fmt.Errorf("trigger_name parameter is required")
	}

	tableName, ok := args["table_name"].(string)
	if !ok {
		return "", fmt.Errorf("table_name parameter is required")
	}

	schema, table := parseTableName(tableName)

	query := fmt.Sprintf(`
		SELECT
			pg_catalog.pg_get_triggerdef(t.oid, true) as trigger_def,
			p.prosrc as function_source
		FROM pg_catalog.pg_trigger t
		JOIN pg_catalog.pg_class c ON c.oid = t.tgrelid
		JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
		LEFT JOIN pg_catalog.pg_proc p ON p.oid = t.tgfoid
		WHERE n.nspname = '%s'
			AND c.relname = '%s'
			AND t.tgname = '%s'
			AND NOT t.tgisinternal
	`, schema, table, triggerName)

	result, err := conn.ExecuteQuery(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to get trigger definition: %w", err)
	}

	if len(result.Rows) == 0 {
		return "", fmt.Errorf("trigger not found: %s on table %s.%s", triggerName, schema, table)
	}

	row := result.Rows[0]
	output, _ := json.MarshalIndent(map[string]interface{}{
		"trigger":         triggerName,
		"table":           table,
		"schema":          schema,
		"definition":      row[0],
		"function_source": row[1],
	}, "", "  ")

	return string(output), nil
}

// ============================================================================
// Tool Implementations - Category 5: Statistics Tools
// ============================================================================

func (t *AdvancedTools) getColumnStats(ctx context.Context, args map[string]interface{}) (string, error) {
	// Get current connection
	conn, err := t.connGetter()
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	tableName, ok := args["table_name"].(string)
	if !ok {
		return "", fmt.Errorf("table_name parameter is required")
	}

	columnName, _ := args["column_name"].(string)
	schema, table := parseTableName(tableName)

	whereClause := fmt.Sprintf("schemaname = '%s' AND tablename = '%s'", schema, table)
	if columnName != "" {
		whereClause += fmt.Sprintf(" AND attname = '%s'", columnName)
	}

	query := fmt.Sprintf(`
		SELECT
			attname as column_name,
			null_frac * 100 as null_percent,
			n_distinct,
			avg_width as avg_bytes,
			most_common_vals::text as common_values,
			most_common_freqs::text as common_frequencies
		FROM pg_stats
		WHERE %s
		ORDER BY attname
	`, whereClause)

	result, err := conn.ExecuteQuery(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to get column stats: %w", err)
	}

	stats := make([]map[string]interface{}, len(result.Rows))
	for i, row := range result.Rows {
		stats[i] = map[string]interface{}{
			"column":              row[0],
			"null_percent":        row[1],
			"distinct_values":     row[2],
			"avg_bytes":           row[3],
			"common_values":       row[4],
			"common_frequencies":  row[5],
		}
	}

	output, _ := json.MarshalIndent(map[string]interface{}{
		"table":  table,
		"schema": schema,
		"stats":  stats,
	}, "", "  ")

	return string(output), nil
}

func (t *AdvancedTools) getTableStats(ctx context.Context, args map[string]interface{}) (string, error) {
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

	query := fmt.Sprintf(`
		SELECT
			n_live_tup as live_rows,
			n_dead_tup as dead_rows,
			n_tup_ins as inserts,
			n_tup_upd as updates,
			n_tup_del as deletes,
			last_vacuum,
			last_autovacuum,
			last_analyze,
			last_autoanalyze,
			vacuum_count,
			autovacuum_count,
			analyze_count,
			autoanalyze_count
		FROM pg_stat_user_tables
		WHERE schemaname = '%s' AND relname = '%s'
	`, schema, table)

	result, err := conn.ExecuteQuery(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to get table stats: %w", err)
	}

	if len(result.Rows) == 0 {
		return "", fmt.Errorf("table not found: %s.%s", schema, table)
	}

	row := result.Rows[0]
	output, _ := json.MarshalIndent(map[string]interface{}{
		"table":             table,
		"schema":            schema,
		"live_rows":         row[0],
		"dead_rows":         row[1],
		"inserts":           row[2],
		"updates":           row[3],
		"deletes":           row[4],
		"last_vacuum":       row[5],
		"last_autovacuum":   row[6],
		"last_analyze":      row[7],
		"last_autoanalyze":  row[8],
		"vacuum_count":      row[9],
		"autovacuum_count":  row[10],
		"analyze_count":     row[11],
		"autoanalyze_count": row[12],
	}, "", "  ")

	return string(output), nil
}

// ============================================================================
// Tool Implementations - Category 6: Discovery Tools
// ============================================================================

func (t *AdvancedTools) listSequences(ctx context.Context, args map[string]interface{}) (string, error) {
	// Get current connection
	conn, err := t.connGetter()
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	schemaFilter, _ := args["schema"].(string)

	whereClause := ""
	if schemaFilter != "" {
		whereClause = fmt.Sprintf("WHERE sequence_schema = '%s'", schemaFilter)
	}

	query := fmt.Sprintf(`
		SELECT
			sequence_schema,
			sequence_name,
			data_type,
			start_value,
			minimum_value,
			maximum_value,
			increment,
			cycle_option
		FROM information_schema.sequences
		%s
		ORDER BY sequence_schema, sequence_name
	`, whereClause)

	result, err := conn.ExecuteQuery(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to list sequences: %w", err)
	}

	sequences := make([]map[string]interface{}, len(result.Rows))
	for i, row := range result.Rows {
		sequences[i] = map[string]interface{}{
			"schema":    row[0],
			"name":      row[1],
			"data_type": row[2],
			"start":     row[3],
			"min":       row[4],
			"max":       row[5],
			"increment": row[6],
			"cycle":     row[7],
		}
	}

	output, _ := json.MarshalIndent(map[string]interface{}{
		"sequences": sequences,
	}, "", "  ")

	return string(output), nil
}

func (t *AdvancedTools) listMaterializedViews(ctx context.Context, args map[string]interface{}) (string, error) {
	// Get current connection
	conn, err := t.connGetter()
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	schemaFilter, _ := args["schema"].(string)

	whereClause := "WHERE schemaname != 'pg_catalog' AND schemaname != 'information_schema'"
	if schemaFilter != "" {
		whereClause += fmt.Sprintf(" AND schemaname = '%s'", schemaFilter)
	}

	query := fmt.Sprintf(`
		SELECT
			schemaname,
			matviewname,
			pg_size_pretty(pg_total_relation_size(schemaname||'.'||matviewname)) as size,
			definition
		FROM pg_matviews
		%s
		ORDER BY schemaname, matviewname
	`, whereClause)

	result, err := conn.ExecuteQuery(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to list materialized views: %w", err)
	}

	matviews := make([]map[string]interface{}, len(result.Rows))
	for i, row := range result.Rows {
		matviews[i] = map[string]interface{}{
			"schema":     row[0],
			"name":       row[1],
			"size":       row[2],
			"definition": row[3],
		}
	}

	output, _ := json.MarshalIndent(map[string]interface{}{
		"materialized_views": matviews,
	}, "", "  ")

	return string(output), nil
}

func (t *AdvancedTools) getTableReferences(ctx context.Context, args map[string]interface{}) (string, error) {
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

	// Get outgoing references (tables this table references)
	outgoingQuery := fmt.Sprintf(`
		SELECT DISTINCT
			ccu.table_schema,
			ccu.table_name
		FROM information_schema.table_constraints AS tc
		JOIN information_schema.constraint_column_usage AS ccu
			ON ccu.constraint_name = tc.constraint_name
			AND ccu.table_schema = tc.table_schema
		WHERE tc.constraint_type = 'FOREIGN KEY'
			AND tc.table_schema = '%s'
			AND tc.table_name = '%s'
		ORDER BY ccu.table_schema, ccu.table_name
	`, schema, table)

	outgoingResult, _ := conn.ExecuteQuery(ctx, outgoingQuery)

	outgoing := make([]string, len(outgoingResult.Rows))
	for i, row := range outgoingResult.Rows {
		outgoing[i] = fmt.Sprintf("%s.%s", row[0], row[1])
	}

	// Get incoming references (tables that reference this table)
	incomingQuery := fmt.Sprintf(`
		SELECT DISTINCT
			tc.table_schema,
			tc.table_name
		FROM information_schema.table_constraints AS tc
		JOIN information_schema.constraint_column_usage AS ccu
			ON ccu.constraint_name = tc.constraint_name
			AND ccu.table_schema = tc.table_schema
		WHERE tc.constraint_type = 'FOREIGN KEY'
			AND ccu.table_schema = '%s'
			AND ccu.table_name = '%s'
		ORDER BY tc.table_schema, tc.table_name
	`, schema, table)

	incomingResult, _ := conn.ExecuteQuery(ctx, incomingQuery)

	incoming := make([]string, len(incomingResult.Rows))
	for i, row := range incomingResult.Rows {
		incoming[i] = fmt.Sprintf("%s.%s", row[0], row[1])
	}

	output, _ := json.MarshalIndent(map[string]interface{}{
		"table":                  table,
		"schema":                 schema,
		"references_to":          outgoing,
		"referenced_by":          incoming,
		"total_references":       len(outgoing),
		"total_referenced_by":    len(incoming),
	}, "", "  ")

	return string(output), nil
}
