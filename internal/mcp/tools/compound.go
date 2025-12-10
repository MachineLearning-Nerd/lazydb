package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/MachineLearning-Nerd/lazydb/internal/mcp/server"
)

// CompoundTools provides high-level tools that combine multiple operations
type CompoundTools struct {
	connGetter     server.ConnectionGetter
	sessionManager SessionManager
}

// NewCompoundTools creates a new CompoundTools instance
func NewCompoundTools(connGetter server.ConnectionGetter) *CompoundTools {
	return &CompoundTools{connGetter: connGetter, sessionManager: nil}
}

// NewCompoundToolsWithSession creates a new CompoundTools instance with session manager
func NewCompoundToolsWithSession(connGetter server.ConnectionGetter, sessionManager SessionManager) *CompoundTools {
	return &CompoundTools{connGetter: connGetter, sessionManager: sessionManager}
}

// Register registers all compound tools with the tool registry
func (t *CompoundTools) Register(registry *server.ToolRegistry) {
	// Tool 1: analyze_table_comprehensive
	registry.Register(
		server.Tool{
			Name:        "analyze_table_comprehensive",
			Description: "Full table analysis: schema, indexes, size, stats, constraints, references. Single call replaces 6 tools.",
			Category:    "compound",
			Tags:        []string{"analysis", "comprehensive", "all-in-one"},
			InputExamples: []map[string]interface{}{
				{"table_name": "users"},
				{"table_name": "orders", "include_sample_data": true},
				{"table_name": "public.products"},
			},
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"table_name": {
						Type:     "string",
						Format:   "schema.table or table",
						Examples: []string{"users", "orders", "products"},
					},
					"include_sample_data": {
						Type:    "boolean",
						Default: false,
					},
				},
				Required: []string{"table_name"},
			},
		},
		t.analyzeTableComprehensive,
	)

	// Tool 2: search_lazydb_tools
	registry.Register(
		server.Tool{
			Name:        "search_lazydb_tools",
			Description: "Find relevant LazyDB tools by keyword. Use when unsure which tool to use.",
			Category:    "meta",
			Tags:        []string{"search", "help", "discovery"},
			InputExamples: []map[string]interface{}{
				{"query": "performance"},
				{"query": "foreign keys"},
				{"query": "table structure"},
				{"query": "index"},
			},
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"query": {
						Type:     "string",
						Examples: []string{"performance", "indexes", "relationships", "schema"},
					},
				},
				Required: []string{"query"},
			},
		},
		t.searchTools,
	)
}

// analyzeTableComprehensive performs full table analysis combining multiple operations
func (t *CompoundTools) analyzeTableComprehensive(ctx context.Context, args map[string]interface{}) (string, error) {
	conn, err := t.connGetter()
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	tableName, ok := args["table_name"].(string)
	if !ok {
		return "", fmt.Errorf("table_name parameter is required")
	}

	includeSampleData := false
	if val, ok := args["include_sample_data"].(bool); ok {
		includeSampleData = val
	}

	schema, table := parseTableName(tableName)
	fullTableName := fmt.Sprintf("%s.%s", schema, table)

	// Execute all analyses in parallel
	var wg sync.WaitGroup
	results := make(map[string]interface{})
	var mu sync.Mutex
	errors := make([]string, 0)

	// 1. Get table schema (columns)
	wg.Add(1)
	go func() {
		defer wg.Done()
		columns, err := conn.GetTableColumns(ctx, schema, table)
		if err != nil {
			mu.Lock()
			errors = append(errors, fmt.Sprintf("schema: %v", err))
			mu.Unlock()
			return
		}
		cols := make([]map[string]interface{}, len(columns))
		for i, col := range columns {
			cols[i] = map[string]interface{}{
				"name":     col.Name,
				"type":     col.Type,
				"nullable": col.Nullable,
				"default":  col.Default,
			}
		}
		mu.Lock()
		results["columns"] = cols
		mu.Unlock()
	}()

	// 2. Get table size
	wg.Add(1)
	go func() {
		defer wg.Done()
		query := fmt.Sprintf(`
			SELECT
				pg_size_pretty(pg_total_relation_size('%s')) as total_size,
				pg_size_pretty(pg_relation_size('%s')) as table_size,
				pg_size_pretty(pg_total_relation_size('%s') - pg_relation_size('%s')) as indexes_size,
				(SELECT reltuples::bigint FROM pg_class WHERE relname = '%s' AND relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = '%s')) as row_count
		`, fullTableName, fullTableName, fullTableName, fullTableName, table, schema)

		result, err := conn.ExecuteQuery(ctx, query)
		if err != nil {
			mu.Lock()
			errors = append(errors, fmt.Sprintf("size: %v", err))
			mu.Unlock()
			return
		}
		if len(result.Rows) > 0 && len(result.Rows[0]) >= 4 {
			mu.Lock()
			results["size"] = map[string]interface{}{
				"total":   result.Rows[0][0],
				"table":   result.Rows[0][1],
				"indexes": result.Rows[0][2],
				"rows":    result.Rows[0][3],
			}
			mu.Unlock()
		}
	}()

	// 3. Get indexes
	wg.Add(1)
	go func() {
		defer wg.Done()
		query := fmt.Sprintf(`
			SELECT
				i.indexname,
				am.amname as index_type,
				ix.indisunique as is_unique,
				ix.indisprimary as is_primary,
				pg_size_pretty(pg_relation_size(i.indexname::regclass)) as size
			FROM pg_indexes i
			JOIN pg_class c ON c.relname = i.indexname
			JOIN pg_index ix ON ix.indexrelid = c.oid
			JOIN pg_am am ON am.oid = c.relam
			WHERE i.schemaname = '%s' AND i.tablename = '%s'
		`, schema, table)

		result, err := conn.ExecuteQuery(ctx, query)
		if err != nil {
			mu.Lock()
			errors = append(errors, fmt.Sprintf("indexes: %v", err))
			mu.Unlock()
			return
		}
		indexes := make([]map[string]interface{}, len(result.Rows))
		for i, row := range result.Rows {
			indexes[i] = map[string]interface{}{
				"name":       row[0],
				"type":       row[1],
				"is_unique":  row[2],
				"is_primary": row[3],
				"size":       row[4],
			}
		}
		mu.Lock()
		results["indexes"] = indexes
		mu.Unlock()
	}()

	// 4. Get table stats
	wg.Add(1)
	go func() {
		defer wg.Done()
		query := fmt.Sprintf(`
			SELECT
				n_live_tup as live_rows,
				n_dead_tup as dead_rows,
				last_vacuum,
				last_analyze
			FROM pg_stat_user_tables
			WHERE schemaname = '%s' AND relname = '%s'
		`, schema, table)

		result, err := conn.ExecuteQuery(ctx, query)
		if err != nil {
			mu.Lock()
			errors = append(errors, fmt.Sprintf("stats: %v", err))
			mu.Unlock()
			return
		}
		if len(result.Rows) > 0 && len(result.Rows[0]) >= 4 {
			mu.Lock()
			results["statistics"] = map[string]interface{}{
				"live_rows":    result.Rows[0][0],
				"dead_rows":    result.Rows[0][1],
				"last_vacuum":  result.Rows[0][2],
				"last_analyze": result.Rows[0][3],
			}
			mu.Unlock()
		}
	}()

	// 5. Get constraints
	wg.Add(1)
	go func() {
		defer wg.Done()
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
		`, fullTableName)

		result, err := conn.ExecuteQuery(ctx, query)
		if err != nil {
			mu.Lock()
			errors = append(errors, fmt.Sprintf("constraints: %v", err))
			mu.Unlock()
			return
		}
		constraints := make([]map[string]interface{}, len(result.Rows))
		for i, row := range result.Rows {
			constraints[i] = map[string]interface{}{
				"name":       row[0],
				"type":       row[1],
				"definition": row[2],
			}
		}
		mu.Lock()
		results["constraints"] = constraints
		mu.Unlock()
	}()

	// 6. Get foreign key references
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Outgoing FK
		outgoingQuery := fmt.Sprintf(`
			SELECT DISTINCT ccu.table_schema, ccu.table_name
			FROM information_schema.table_constraints AS tc
			JOIN information_schema.constraint_column_usage AS ccu
				ON ccu.constraint_name = tc.constraint_name
			WHERE tc.constraint_type = 'FOREIGN KEY'
				AND tc.table_schema = '%s' AND tc.table_name = '%s'
		`, schema, table)

		outResult, _ := conn.ExecuteQuery(ctx, outgoingQuery)
		outgoing := make([]string, len(outResult.Rows))
		for i, row := range outResult.Rows {
			outgoing[i] = fmt.Sprintf("%s.%s", row[0], row[1])
		}

		// Incoming FK
		incomingQuery := fmt.Sprintf(`
			SELECT DISTINCT tc.table_schema, tc.table_name
			FROM information_schema.table_constraints AS tc
			JOIN information_schema.constraint_column_usage AS ccu
				ON ccu.constraint_name = tc.constraint_name
			WHERE tc.constraint_type = 'FOREIGN KEY'
				AND ccu.table_schema = '%s' AND ccu.table_name = '%s'
		`, schema, table)

		inResult, _ := conn.ExecuteQuery(ctx, incomingQuery)
		incoming := make([]string, len(inResult.Rows))
		for i, row := range inResult.Rows {
			incoming[i] = fmt.Sprintf("%s.%s", row[0], row[1])
		}

		mu.Lock()
		results["references"] = map[string]interface{}{
			"references_to":  outgoing,
			"referenced_by":  incoming,
		}
		mu.Unlock()
	}()

	// 7. Get sample data if requested
	if includeSampleData {
		wg.Add(1)
		go func() {
			defer wg.Done()
			query := fmt.Sprintf("SELECT * FROM %s.%s LIMIT 5", schema, table)
			result, err := conn.ExecuteQuery(ctx, query)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Sprintf("sample_data: %v", err))
				mu.Unlock()
				return
			}
			mu.Lock()
			results["sample_data"] = map[string]interface{}{
				"columns": result.Columns,
				"rows":    result.Rows,
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	// Build final output
	output := map[string]interface{}{
		"table":  table,
		"schema": schema,
	}

	// Merge results
	for k, v := range results {
		output[k] = v
	}

	if len(errors) > 0 {
		output["warnings"] = errors
	}

	jsonOutput, _ := json.MarshalIndent(output, "", "  ")
	return string(jsonOutput), nil
}

// searchTools searches for tools by keyword
func (t *CompoundTools) searchTools(ctx context.Context, args map[string]interface{}) (string, error) {
	query, ok := args["query"].(string)
	if !ok {
		return "", fmt.Errorf("query parameter is required")
	}

	// Build enabled categories set for status lookup
	enabledCategories := make(map[string]bool)
	if t.sessionManager != nil {
		for _, cat := range t.sessionManager.GetEnabledCategories() {
			enabledCategories[cat] = true
		}
	}

	// Tool definitions with keywords for search
	toolInfo := []map[string]interface{}{
		// Basic Tools
		{"name": "list_all_tables", "category": "schema", "keywords": []string{"tables", "list", "overview", "database"}},
		{"name": "get_table_schema", "category": "schema", "keywords": []string{"columns", "types", "structure", "schema"}},
		{"name": "search_tables", "category": "discovery", "keywords": []string{"search", "find", "pattern", "like"}},
		{"name": "get_sample_data", "category": "discovery", "keywords": []string{"sample", "rows", "data", "preview"}},
		{"name": "get_table_count", "category": "discovery", "keywords": []string{"count", "rows", "size"}},
		// DDL Tools
		{"name": "get_table_ddl", "category": "schema", "keywords": []string{"ddl", "create", "definition", "sql"}},
		{"name": "get_view_definition", "category": "schema", "keywords": []string{"view", "select", "definition"}},
		{"name": "get_function_definition", "category": "schema", "keywords": []string{"function", "procedure", "source", "code"}},
		// Performance Tools
		{"name": "get_table_indexes", "category": "performance", "keywords": []string{"index", "indexes", "performance"}},
		{"name": "get_table_size", "category": "performance", "keywords": []string{"size", "disk", "storage", "space"}},
		{"name": "explain_query", "category": "performance", "keywords": []string{"explain", "plan", "performance", "query"}},
		// Relationship Tools
		{"name": "get_foreign_keys", "category": "relationships", "keywords": []string{"foreign", "key", "fk", "relationship", "reference"}},
		{"name": "get_table_constraints", "category": "relationships", "keywords": []string{"constraint", "primary", "unique", "check"}},
		{"name": "get_table_dependencies", "category": "relationships", "keywords": []string{"dependency", "dependent", "view", "function"}},
		{"name": "get_table_references", "category": "relationships", "keywords": []string{"reference", "relationship", "fk", "map"}},
		// Statistics Tools
		{"name": "get_column_stats", "category": "statistics", "keywords": []string{"statistics", "column", "distribution", "null"}},
		{"name": "get_table_stats", "category": "statistics", "keywords": []string{"statistics", "vacuum", "analyze", "dead"}},
		// Trigger Tools
		{"name": "get_table_triggers", "category": "triggers", "keywords": []string{"trigger", "event", "insert", "update", "delete"}},
		{"name": "get_trigger_definition", "category": "triggers", "keywords": []string{"trigger", "definition", "source", "function"}},
		// Discovery Tools
		{"name": "list_sequences", "category": "discovery", "keywords": []string{"sequence", "serial", "autoincrement", "id"}},
		{"name": "list_materialized_views", "category": "discovery", "keywords": []string{"materialized", "view", "cache", "refresh"}},
		// Optimization Tools
		{"name": "analyze_query_performance", "category": "optimization", "keywords": []string{"analyze", "performance", "slow", "index", "optimize"}},
		{"name": "get_query_optimization_context", "category": "optimization", "keywords": []string{"context", "llm", "ai", "optimize"}},
		{"name": "compare_query_performance", "category": "optimization", "keywords": []string{"compare", "benchmark", "before", "after"}},
		// Compound Tools
		{"name": "analyze_table_comprehensive", "category": "compound", "keywords": []string{"comprehensive", "full", "analysis", "all"}},
		// Management Tools (always-on)
		{"name": "lazydb_enable_category", "category": "meta", "keywords": []string{"enable", "add", "activate", "category"}, "always_on": true},
		{"name": "lazydb_disable_category", "category": "meta", "keywords": []string{"disable", "remove", "deactivate", "category"}, "always_on": true},
		{"name": "lazydb_list_categories", "category": "meta", "keywords": []string{"list", "categories", "status", "available"}, "always_on": true},
		{"name": "lazydb_reset_session", "category": "meta", "keywords": []string{"reset", "session", "preset", "default"}, "always_on": true},
		{"name": "search_lazydb_tools", "category": "meta", "keywords": []string{"search", "find", "tools", "discover"}, "always_on": true},
	}

	queryLower := strings.ToLower(query)
	// Split query into words for multi-word search support
	queryWords := strings.Fields(queryLower)
	matches := make([]map[string]interface{}, 0)

	for _, tool := range toolInfo {
		matched := false
		toolName := strings.ToLower(tool["name"].(string))
		toolCategory := strings.ToLower(tool["category"].(string))
		keywords := tool["keywords"].([]string)

		// Check each query word against tool name, category, and keywords
		for _, word := range queryWords {
			if matched {
				break
			}
			// Check if word matches tool name
			if strings.Contains(toolName, word) {
				matched = true
				break
			}
			// Check if word matches category
			if strings.Contains(toolCategory, word) {
				matched = true
				break
			}
			// Check if word matches any keyword
			for _, kw := range keywords {
				if strings.Contains(strings.ToLower(kw), word) || strings.Contains(word, strings.ToLower(kw)) {
					matched = true
					break
				}
			}
		}

		if matched {
			// Add enabled status
			isAlwaysOn := false
			if ao, ok := tool["always_on"].(bool); ok {
				isAlwaysOn = ao
			}

			enabled := isAlwaysOn || enabledCategories[tool["category"].(string)]

			matchResult := map[string]interface{}{
				"name":      tool["name"],
				"category":  tool["category"],
				"keywords":  tool["keywords"],
				"enabled":   enabled,
				"always_on": isAlwaysOn,
			}

			// Add hint for disabled tools
			if !enabled {
				matchResult["hint"] = fmt.Sprintf("Enable with: lazydb_enable_category(\"%s\")", tool["category"])
			}

			matches = append(matches, matchResult)
		}
	}

	// Count enabled vs disabled
	enabledCount := 0
	for _, m := range matches {
		if m["enabled"].(bool) {
			enabledCount++
		}
	}

	output, _ := json.MarshalIndent(map[string]interface{}{
		"query":         query,
		"matches":       matches,
		"count":         len(matches),
		"enabled_count": enabledCount,
	}, "", "  ")

	return string(output), nil
}
