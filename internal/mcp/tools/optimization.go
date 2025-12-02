package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/MachineLearning-Nerd/lazydb/internal/db"
	"github.com/MachineLearning-Nerd/lazydb/internal/mcp/server"
)

// OptimizationTools provides query optimization and performance analysis tools
type OptimizationTools struct {
	connGetter server.ConnectionGetter
}

// NewOptimizationTools creates a new OptimizationTools instance
func NewOptimizationTools(connGetter server.ConnectionGetter) *OptimizationTools {
	return &OptimizationTools{connGetter: connGetter}
}

// Register registers all optimization tools with the tool registry
func (t *OptimizationTools) Register(registry *server.ToolRegistry) {
	t.registerAnalysisTools(registry)
	t.registerComparisonTools(registry)
}

// ============================================================================
// Helper Types for Query Analysis
// ============================================================================

// Issue represents a detected performance issue
type Issue struct {
	Severity    string   `json:"severity"`    // critical, high, medium, low
	Type        string   `json:"type"`        // seq_scan, missing_index, high_cost, nested_loop, sort_without_index
	Description string   `json:"description"` // Human-readable description
	Table       string   `json:"table"`       // Affected table (schema.table)
	Columns     []string `json:"columns"`     // Affected columns
	Suggestion  string   `json:"suggestion"`  // Recommended action
}

// IndexSuggestion represents a suggested index
type IndexSuggestion struct {
	Table               string `json:"table"`                // schema.table
	Columns             []string `json:"columns"`            // Columns to index
	IndexType           string `json:"type"`                 // btree, hash, gin, gist
	Reason              string `json:"reason"`               // Why this index is suggested
	EstimatedImprovement string `json:"estimated_improvement"` // high, medium, low
	DDL                 string `json:"ddl"`                  // CREATE INDEX statement
}

// PlanNode represents a node in the EXPLAIN JSON output
type PlanNode struct {
	NodeType          string     `json:"Node Type"`
	RelationName      string     `json:"Relation Name,omitempty"`
	Schema            string     `json:"Schema,omitempty"`
	Alias             string     `json:"Alias,omitempty"`
	StartupCost       float64    `json:"Startup Cost,omitempty"`
	TotalCost         float64    `json:"Total Cost,omitempty"`
	PlanRows          float64    `json:"Plan Rows,omitempty"`
	PlanWidth         int        `json:"Plan Width,omitempty"`
	ActualStartupTime float64    `json:"Actual Startup Time,omitempty"`
	ActualTotalTime   float64    `json:"Actual Total Time,omitempty"`
	ActualRows        float64    `json:"Actual Rows,omitempty"`
	ActualLoops       int        `json:"Actual Loops,omitempty"`
	Filter            string     `json:"Filter,omitempty"`
	IndexName         string     `json:"Index Name,omitempty"`
	IndexCond         string     `json:"Index Cond,omitempty"`
	JoinType          string     `json:"Join Type,omitempty"`
	HashCond          string     `json:"Hash Cond,omitempty"`
	MergeCond         string     `json:"Merge Cond,omitempty"`
	SortKey           []string   `json:"Sort Key,omitempty"`
	SharedHitBlocks   int        `json:"Shared Hit Blocks,omitempty"`
	SharedReadBlocks  int        `json:"Shared Read Blocks,omitempty"`
	Plans             []PlanNode `json:"Plans,omitempty"`
}

// ExplainResult represents the top-level EXPLAIN JSON output
type ExplainResult struct {
	Plan          PlanNode `json:"Plan"`
	PlanningTime  float64  `json:"Planning Time,omitempty"`
	ExecutionTime float64  `json:"Execution Time,omitempty"`
}

// ============================================================================
// Tool Registration
// ============================================================================

func (t *OptimizationTools) registerAnalysisTools(registry *server.ToolRegistry) {
	// Tool 1: analyze_query_performance (Rule-Based)
	registry.Register(
		server.Tool{
			Name:        "analyze_query_performance",
			Description: "Detect performance issues: seq scans, missing indexes, high cost. Suggests CREATE INDEX.",
			Category:    "optimization",
			Tags:        []string{"performance", "analysis", "index", "optimization"},
			InputExamples: []map[string]interface{}{
				{"query": "SELECT * FROM users WHERE email = 'test@example.com'"},
				{"query": "SELECT * FROM orders WHERE status = 'pending'", "row_threshold": 5000},
				{"query": "SELECT u.*, o.* FROM users u JOIN orders o ON u.id = o.user_id", "include_ddl_suggestions": true},
			},
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"query": {
						Type:     "string",
						Examples: []string{"SELECT * FROM users WHERE id = 1", "SELECT * FROM orders WHERE status = 'pending'"},
					},
					"include_ddl_suggestions": {
						Type:    "boolean",
						Default: true,
					},
					"row_threshold": {
						Type:    "integer",
						Default: 10000,
					},
				},
				Required: []string{"query"},
			},
		},
		t.analyzeQueryPerformance,
	)

	// Tool 2: get_query_optimization_context (AI Context Provider)
	registry.Register(
		server.Tool{
			Name:        "get_query_optimization_context",
			Description: "Aggregate context for LLM optimization: EXPLAIN, DDLs, indexes, stats.",
			Category:    "optimization",
			Tags:        []string{"optimization", "context", "ai", "llm"},
			InputExamples: []map[string]interface{}{
				{"query": "SELECT * FROM users WHERE email = 'test@example.com'"},
				{"query": "SELECT * FROM orders WHERE created_at > NOW() - INTERVAL '7 days'", "include_sample_data": true},
				{"query": "SELECT u.*, o.* FROM users u JOIN orders o ON u.id = o.user_id", "include_related_tables": true},
			},
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"query": {
						Type:     "string",
						Examples: []string{"SELECT * FROM users WHERE id = 1", "SELECT * FROM orders JOIN users ON orders.user_id = users.id"},
					},
					"include_sample_data": {
						Type:    "boolean",
						Default: false,
					},
					"include_related_tables": {
						Type:    "boolean",
						Default: true,
					},
				},
				Required: []string{"query"},
			},
		},
		t.getQueryOptimizationContext,
	)
}

func (t *OptimizationTools) registerComparisonTools(registry *server.ToolRegistry) {
	// Tool 3: compare_query_performance
	registry.Register(
		server.Tool{
			Name:        "compare_query_performance",
			Description: "Compare before/after query performance with improvement %.",
			Category:    "optimization",
			Tags:        []string{"performance", "comparison", "benchmark", "optimization"},
			InputExamples: []map[string]interface{}{
				{
					"query_before": "SELECT * FROM users WHERE email = 'test@example.com'",
					"query_after":  "SELECT * FROM users WHERE email = 'test@example.com' AND id = 1",
				},
				{
					"query_before": "SELECT * FROM orders WHERE status = 'pending'",
					"query_after":  "SELECT id, status FROM orders WHERE status = 'pending'",
					"runs":         5,
				},
			},
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"query_before": {
						Type:     "string",
						Examples: []string{"SELECT * FROM users WHERE email = 'test@example.com'"},
					},
					"query_after": {
						Type:     "string",
						Examples: []string{"SELECT id, email FROM users WHERE email = 'test@example.com'"},
					},
					"runs": {
						Type:    "integer",
						Default: 3,
					},
				},
				Required: []string{"query_before", "query_after"},
			},
		},
		t.compareQueryPerformance,
	)
}

// ============================================================================
// Tool Implementations
// ============================================================================

func (t *OptimizationTools) analyzeQueryPerformance(ctx context.Context, args map[string]interface{}) (string, error) {
	conn, err := t.connGetter()
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	query, ok := args["query"].(string)
	if !ok {
		return "", fmt.Errorf("query parameter is required")
	}

	// Validate SELECT query
	if err := validateSelectQuery(query); err != nil {
		return "", err
	}

	includeDDL := true
	if val, ok := args["include_ddl_suggestions"].(bool); ok {
		includeDDL = val
	}

	rowThreshold := int64(10000)
	if val, ok := args["row_threshold"].(float64); ok {
		rowThreshold = int64(val)
	}

	// Run EXPLAIN ANALYZE with JSON format
	explainQuery := fmt.Sprintf("EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) %s", query)
	result, err := conn.ExecuteQuery(ctx, explainQuery)
	if err != nil {
		return "", fmt.Errorf("failed to explain query: %w", err)
	}

	// Parse EXPLAIN JSON output
	if len(result.Rows) == 0 || len(result.Rows[0]) == 0 {
		return "", fmt.Errorf("empty explain output")
	}

	var explainResults []ExplainResult
	if err := json.Unmarshal([]byte(result.Rows[0][0]), &explainResults); err != nil {
		return "", fmt.Errorf("failed to parse explain output: %w", err)
	}

	if len(explainResults) == 0 {
		return "", fmt.Errorf("empty explain results")
	}

	explainResult := explainResults[0]

	// Get table row counts for referenced tables
	tables := extractTablesFromPlan(&explainResult.Plan)
	tableRowCounts := make(map[string]int64)
	existingIndexes := make(map[string][]string)

	for _, tableName := range tables {
		// Validate table name to prevent SQL injection
		schema, table, err := safeTableName(tableName)
		if err != nil {
			continue // Skip invalid table names
		}
		fullName := fmt.Sprintf("%s.%s", schema, table)

		// Get row count estimate
		rowCountQuery := fmt.Sprintf(`
			SELECT reltuples::bigint as row_estimate
			FROM pg_class c
			JOIN pg_namespace n ON n.oid = c.relnamespace
			WHERE n.nspname = '%s' AND c.relname = '%s'
		`, schema, table)

		rowResult, err := conn.ExecuteQuery(ctx, rowCountQuery)
		if err == nil && len(rowResult.Rows) > 0 && len(rowResult.Rows[0]) > 0 {
			if count, err := strconv.ParseInt(rowResult.Rows[0][0], 10, 64); err == nil {
				tableRowCounts[fullName] = count
			}
		}

		// Get existing indexes
		indexQuery := fmt.Sprintf(`
			SELECT indexname, array_to_string(array_agg(attname), ',') as columns
			FROM (
				SELECT i.indexname, a.attname
				FROM pg_indexes i
				JOIN pg_class c ON c.relname = i.indexname
				JOIN pg_index ix ON ix.indexrelid = c.oid
				JOIN pg_attribute a ON a.attrelid = ix.indrelid AND a.attnum = ANY(ix.indkey)
				WHERE i.schemaname = '%s' AND i.tablename = '%s'
				ORDER BY i.indexname, array_position(ix.indkey, a.attnum)
			) sub
			GROUP BY indexname
		`, schema, table)

		indexResult, err := conn.ExecuteQuery(ctx, indexQuery)
		if err == nil {
			for _, row := range indexResult.Rows {
				if len(row) >= 2 {
					existingIndexes[fullName] = append(existingIndexes[fullName], row[1])
				}
			}
		}
	}

	// Detect issues
	issues := detectIssues(&explainResult.Plan, tableRowCounts, rowThreshold)

	// Generate index suggestions
	suggestions := generateIndexSuggestions(issues, existingIndexes, includeDDL)

	// Build warnings
	var warnings []string
	for _, table := range tables {
		schema, tableName := parseTableName(table)
		fullName := fmt.Sprintf("%s.%s", schema, tableName)
		if _, ok := tableRowCounts[fullName]; !ok {
			warnings = append(warnings, fmt.Sprintf("Could not get row count for %s", fullName))
		}
	}

	// Check if statistics might be stale
	if explainResult.Plan.ActualRows > 0 && explainResult.Plan.PlanRows > 0 {
		ratio := explainResult.Plan.ActualRows / explainResult.Plan.PlanRows
		if ratio > 10 || ratio < 0.1 {
			warnings = append(warnings, "Statistics may be outdated. Consider running ANALYZE on affected tables.")
		}
	}

	output, _ := json.MarshalIndent(map[string]interface{}{
		"query": query,
		"analysis": map[string]interface{}{
			"total_cost":        explainResult.Plan.TotalCost,
			"actual_time_ms":    explainResult.Plan.ActualTotalTime,
			"rows_returned":     explainResult.Plan.ActualRows,
			"planning_time_ms":  explainResult.PlanningTime,
			"execution_time_ms": explainResult.ExecutionTime,
			"issues":            issues,
		},
		"index_suggestions": suggestions,
		"warnings":          warnings,
	}, "", "  ")

	return string(output), nil
}

func (t *OptimizationTools) getQueryOptimizationContext(ctx context.Context, args map[string]interface{}) (string, error) {
	conn, err := t.connGetter()
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	query, ok := args["query"].(string)
	if !ok {
		return "", fmt.Errorf("query parameter is required")
	}

	// Validate SELECT query
	if err := validateSelectQuery(query); err != nil {
		return "", err
	}

	includeSampleData := false
	if val, ok := args["include_sample_data"].(bool); ok {
		includeSampleData = val
	}

	includeRelatedTables := true
	if val, ok := args["include_related_tables"].(bool); ok {
		includeRelatedTables = val
	}

	// Get EXPLAIN ANALYZE in both formats
	explainJSONQuery := fmt.Sprintf("EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON, VERBOSE) %s", query)
	explainTextQuery := fmt.Sprintf("EXPLAIN (ANALYZE, BUFFERS, VERBOSE) %s", query)

	jsonResult, err := conn.ExecuteQuery(ctx, explainJSONQuery)
	if err != nil {
		return "", fmt.Errorf("failed to explain query: %w", err)
	}

	textResult, err := conn.ExecuteQuery(ctx, explainTextQuery)
	if err != nil {
		return "", fmt.Errorf("failed to explain query (text): %w", err)
	}

	// Parse JSON result
	var explainResults []ExplainResult
	if len(jsonResult.Rows) > 0 && len(jsonResult.Rows[0]) > 0 {
		if err := json.Unmarshal([]byte(jsonResult.Rows[0][0]), &explainResults); err != nil {
			return "", fmt.Errorf("failed to parse EXPLAIN JSON: %w", err)
		}
	}

	// Build text output
	var textPlan strings.Builder
	for _, row := range textResult.Rows {
		for _, col := range row {
			textPlan.WriteString(col)
			textPlan.WriteString("\n")
		}
	}

	// Extract tables
	var tables []string
	if len(explainResults) > 0 {
		tables = extractTablesFromPlan(&explainResults[0].Plan)
	}

	// Build metrics
	metrics := map[string]interface{}{}
	if len(explainResults) > 0 {
		metrics["planning_time_ms"] = explainResults[0].PlanningTime
		metrics["execution_time_ms"] = explainResults[0].ExecutionTime
		metrics["total_cost"] = explainResults[0].Plan.TotalCost
		metrics["actual_rows"] = explainResults[0].Plan.ActualRows
		metrics["buffers_hit"] = explainResults[0].Plan.SharedHitBlocks
		metrics["buffers_read"] = explainResults[0].Plan.SharedReadBlocks
	}

	// Build table context
	tablesContext := make(map[string]interface{})
	if includeRelatedTables {
		for _, tableName := range tables {
			// Validate table name to prevent SQL injection
			schema, table, err := safeTableName(tableName)
			if err != nil {
				continue // Skip invalid table names
			}
			fullName := fmt.Sprintf("%s.%s", schema, table)

			tableInfo := make(map[string]interface{})

			// Get DDL
			columns, err := conn.GetTableColumns(ctx, schema, table)
			if err == nil {
				var ddl strings.Builder
				ddl.WriteString(fmt.Sprintf("CREATE TABLE %s.%s (\n", schema, table))
				for i, col := range columns {
					nullability := "NULL"
					if !col.Nullable {
						nullability = "NOT NULL"
					}
					ddl.WriteString(fmt.Sprintf("    %s %s %s", col.Name, col.Type, nullability))
					if i < len(columns)-1 {
						ddl.WriteString(",")
					}
					ddl.WriteString("\n")
				}
				ddl.WriteString(");")
				tableInfo["ddl"] = ddl.String()
			}

			// Get row count and size
			sizeQuery := fmt.Sprintf(`
				SELECT
					pg_size_pretty(pg_total_relation_size('%s.%s')) as total_size,
					(SELECT reltuples::bigint FROM pg_class WHERE relname = '%s' AND relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = '%s')) as row_count
			`, schema, table, table, schema)

			sizeResult, err := conn.ExecuteQuery(ctx, sizeQuery)
			if err == nil && len(sizeResult.Rows) > 0 {
				if len(sizeResult.Rows[0]) > 0 {
					tableInfo["size"] = sizeResult.Rows[0][0]
				}
				if len(sizeResult.Rows[0]) > 1 {
					tableInfo["row_count"] = sizeResult.Rows[0][1]
				}
			}

			// Get indexes
			indexQuery := fmt.Sprintf(`
				SELECT
					i.indexname,
					am.amname as index_type,
					ix.indisunique as is_unique,
					ix.indisprimary as is_primary,
					pg_size_pretty(pg_relation_size(i.indexname::regclass)) as size,
					i.indexdef as definition
				FROM pg_indexes i
				JOIN pg_class c ON c.relname = i.indexname
				JOIN pg_index ix ON ix.indexrelid = c.oid
				JOIN pg_am am ON am.oid = c.relam
				WHERE i.schemaname = '%s' AND i.tablename = '%s'
			`, schema, table)

			indexResult, err := conn.ExecuteQuery(ctx, indexQuery)
			if err == nil && len(indexResult.Rows) > 0 {
				indexes := make([]map[string]interface{}, len(indexResult.Rows))
				for i, row := range indexResult.Rows {
					indexes[i] = map[string]interface{}{
						"name":       row[0],
						"type":       row[1],
						"is_unique":  row[2],
						"is_primary": row[3],
						"size":       row[4],
						"definition": row[5],
					}
				}
				tableInfo["indexes"] = indexes
			}

			// Get column statistics
			statsQuery := fmt.Sprintf(`
				SELECT
					attname as column_name,
					null_frac * 100 as null_percent,
					n_distinct,
					avg_width as avg_bytes
				FROM pg_stats
				WHERE schemaname = '%s' AND tablename = '%s'
			`, schema, table)

			statsResult, err := conn.ExecuteQuery(ctx, statsQuery)
			if err == nil && len(statsResult.Rows) > 0 {
				colStats := make(map[string]interface{})
				for _, row := range statsResult.Rows {
					if len(row) >= 4 {
						colStats[row[0]] = map[string]interface{}{
							"null_percent":    row[1],
							"distinct_values": row[2],
							"avg_width_bytes": row[3],
						}
					}
				}
				tableInfo["column_statistics"] = colStats
			}

			// Get table statistics
			tableStatsQuery := fmt.Sprintf(`
				SELECT
					n_live_tup as live_rows,
					n_dead_tup as dead_rows,
					last_vacuum,
					last_analyze
				FROM pg_stat_user_tables
				WHERE schemaname = '%s' AND relname = '%s'
			`, schema, table)

			tableStatsResult, err := conn.ExecuteQuery(ctx, tableStatsQuery)
			if err == nil && len(tableStatsResult.Rows) > 0 && len(tableStatsResult.Rows[0]) >= 4 {
				tableInfo["table_statistics"] = map[string]interface{}{
					"live_rows":    tableStatsResult.Rows[0][0],
					"dead_rows":    tableStatsResult.Rows[0][1],
					"last_vacuum":  tableStatsResult.Rows[0][2],
					"last_analyze": tableStatsResult.Rows[0][3],
				}
			}

			// Get sample data if requested
			if includeSampleData {
				sampleQuery := fmt.Sprintf("SELECT * FROM %s.%s LIMIT 5", schema, table)
				sampleResult, err := conn.ExecuteQuery(ctx, sampleQuery)
				if err == nil && len(sampleResult.Rows) > 0 {
					tableInfo["sample_data"] = map[string]interface{}{
						"columns": sampleResult.Columns,
						"rows":    sampleResult.Rows,
					}
				}
			}

			tablesContext[fullName] = tableInfo
		}
	}

	// Detect potential issues
	var potentialIssues []string
	if len(explainResults) > 0 {
		issues := detectIssues(&explainResults[0].Plan, nil, 10000)
		for _, issue := range issues {
			potentialIssues = append(potentialIssues, issue.Description)
		}
	}

	// Generate optimization hints
	optimizationHints := []string{
		"Consider indexing frequently filtered columns",
		"Review join order for optimal performance",
		"Check if statistics are up-to-date with ANALYZE",
	}

	output, _ := json.MarshalIndent(map[string]interface{}{
		"query": map[string]interface{}{
			"original":          query,
			"tables_referenced": tables,
		},
		"execution_plan": map[string]interface{}{
			"text_format": textPlan.String(),
			"json_format": explainResults,
			"metrics":     metrics,
		},
		"tables":             tablesContext,
		"potential_issues":   potentialIssues,
		"optimization_hints": optimizationHints,
	}, "", "  ")

	return string(output), nil
}

func (t *OptimizationTools) compareQueryPerformance(ctx context.Context, args map[string]interface{}) (string, error) {
	conn, err := t.connGetter()
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	queryBefore, ok := args["query_before"].(string)
	if !ok {
		return "", fmt.Errorf("query_before parameter is required")
	}

	queryAfter, ok := args["query_after"].(string)
	if !ok {
		return "", fmt.Errorf("query_after parameter is required")
	}

	// Validate both queries
	if err := validateSelectQuery(queryBefore); err != nil {
		return "", fmt.Errorf("query_before: %w", err)
	}
	if err := validateSelectQuery(queryAfter); err != nil {
		return "", fmt.Errorf("query_after: %w", err)
	}

	runs := 3
	if val, ok := args["runs"].(float64); ok {
		runs = int(val)
		if runs > 5 {
			runs = 5
		}
		if runs < 1 {
			runs = 1
		}
	}

	// Run EXPLAIN ANALYZE multiple times and average
	beforeMetrics, err := runMultipleExplains(ctx, conn, queryBefore, runs)
	if err != nil {
		return "", fmt.Errorf("failed to analyze query_before: %w", err)
	}

	afterMetrics, err := runMultipleExplains(ctx, conn, queryAfter, runs)
	if err != nil {
		return "", fmt.Errorf("failed to analyze query_after: %w", err)
	}

	// Calculate improvements
	execTimeImprove := calculateImprovement(beforeMetrics["execution_time_ms"].(float64), afterMetrics["execution_time_ms"].(float64))
	costImprove := calculateImprovement(beforeMetrics["total_cost"].(float64), afterMetrics["total_cost"].(float64))

	// Buffer reads improvement (only if we have data)
	var bufferReadsImprove float64
	beforeReads := beforeMetrics["buffers_read"].(float64)
	afterReads := afterMetrics["buffers_read"].(float64)
	if beforeReads > 0 || afterReads > 0 {
		bufferReadsImprove = calculateImprovement(beforeReads, afterReads)
	}

	// Determine verdict
	improved := execTimeImprove < 0
	var summary string
	var recommendations []string

	if improved {
		summary = fmt.Sprintf("The optimized query is %.1f%% faster", -execTimeImprove)
		if bufferReadsImprove < 0 {
			summary += fmt.Sprintf(" with %.1f%% fewer disk reads", -bufferReadsImprove)
		}
	} else if execTimeImprove > 0 {
		summary = fmt.Sprintf("The optimized query is %.1f%% slower - consider reverting", execTimeImprove)
	} else {
		summary = "No significant performance difference detected"
	}

	// Generate recommendations
	if improved {
		recommendations = append(recommendations, "The optimized version is recommended")
	} else {
		recommendations = append(recommendations, "Consider keeping the original query")
	}

	beforeHitRatio := calculateBufferHitRatio(beforeMetrics)
	afterHitRatio := calculateBufferHitRatio(afterMetrics)
	if afterHitRatio > beforeHitRatio {
		recommendations = append(recommendations, fmt.Sprintf("Buffer hit ratio improved from %.1f%% to %.1f%%", beforeHitRatio, afterHitRatio))
	}

	output, _ := json.MarshalIndent(map[string]interface{}{
		"comparison": map[string]interface{}{
			"before": map[string]interface{}{
				"query":        queryBefore,
				"metrics":      beforeMetrics,
				"plan_summary": beforeMetrics["plan_summary"],
			},
			"after": map[string]interface{}{
				"query":        queryAfter,
				"metrics":      afterMetrics,
				"plan_summary": afterMetrics["plan_summary"],
			},
		},
		"improvement": map[string]interface{}{
			"execution_time_percent":       execTimeImprove,
			"cost_reduction_percent":       costImprove,
			"buffer_reads_reduction_percent": bufferReadsImprove,
		},
		"verdict": map[string]interface{}{
			"improved":        improved,
			"summary":         summary,
			"recommendations": recommendations,
		},
	}, "", "  ")

	return string(output), nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// validateSelectQuery ensures the query is a SELECT statement
func validateSelectQuery(query string) error {
	trimmed := strings.TrimSpace(strings.ToUpper(query))
	if !strings.HasPrefix(trimmed, "SELECT") && !strings.HasPrefix(trimmed, "WITH") {
		return fmt.Errorf("only SELECT queries (including CTEs with WITH) are allowed")
	}
	return nil
}

// validateIdentifier checks if a string is a valid PostgreSQL identifier
// This prevents SQL injection via schema/table names
func validateIdentifier(name string) error {
	if name == "" {
		return fmt.Errorf("identifier cannot be empty")
	}
	if len(name) > 63 {
		return fmt.Errorf("identifier too long (max 63 characters)")
	}
	// PostgreSQL identifier: starts with letter/underscore, followed by letters/digits/underscores
	matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, name)
	if !matched {
		return fmt.Errorf("invalid identifier: %s", name)
	}
	return nil
}

// safeTableName validates and returns schema.table, or error if invalid
func safeTableName(tableName string) (string, string, error) {
	schema, table := parseTableName(tableName)
	if err := validateIdentifier(schema); err != nil {
		return "", "", fmt.Errorf("invalid schema name: %w", err)
	}
	if err := validateIdentifier(table); err != nil {
		return "", "", fmt.Errorf("invalid table name: %w", err)
	}
	return schema, table, nil
}

// extractTablesFromPlan recursively extracts table names from EXPLAIN plan
func extractTablesFromPlan(plan *PlanNode) []string {
	tableSet := make(map[string]bool)
	extractTablesRecursive(plan, tableSet)

	tables := make([]string, 0, len(tableSet))
	for table := range tableSet {
		tables = append(tables, table)
	}
	return tables
}

func extractTablesRecursive(plan *PlanNode, tables map[string]bool) {
	if plan.RelationName != "" {
		schema := plan.Schema
		if schema == "" {
			schema = "public"
		}
		tables[fmt.Sprintf("%s.%s", schema, plan.RelationName)] = true
	}

	for i := range plan.Plans {
		extractTablesRecursive(&plan.Plans[i], tables)
	}
}

// detectIssues analyzes the plan and returns detected issues
func detectIssues(plan *PlanNode, tableRowCounts map[string]int64, threshold int64) []Issue {
	var issues []Issue
	detectIssuesRecursive(plan, tableRowCounts, threshold, &issues)
	return issues
}

func detectIssuesRecursive(plan *PlanNode, tableRowCounts map[string]int64, threshold int64, issues *[]Issue) {
	schema := plan.Schema
	if schema == "" {
		schema = "public"
	}
	tableName := fmt.Sprintf("%s.%s", schema, plan.RelationName)

	// Check for Sequential Scan on large tables
	if plan.NodeType == "Seq Scan" && plan.RelationName != "" {
		rowCount := int64(0)
		if tableRowCounts != nil {
			rowCount = tableRowCounts[tableName]
		} else {
			rowCount = int64(plan.PlanRows)
		}

		if rowCount >= threshold {
			// Extract columns from filter
			columns := extractColumnsFromFilter(plan.Filter)

			*issues = append(*issues, Issue{
				Severity:    getSeverity(rowCount, threshold),
				Type:        "seq_scan",
				Description: fmt.Sprintf("Sequential scan on table '%s' with ~%d rows", tableName, rowCount),
				Table:       tableName,
				Columns:     columns,
				Suggestion:  fmt.Sprintf("Consider adding an index on %s(%s)", tableName, strings.Join(columns, ", ")),
			})
		}
	}

	// Check for high cost operations
	if plan.TotalCost > 10000 {
		*issues = append(*issues, Issue{
			Severity:    "medium",
			Type:        "high_cost",
			Description: fmt.Sprintf("High cost operation: %s (cost: %.2f)", plan.NodeType, plan.TotalCost),
			Table:       tableName,
			Columns:     nil,
			Suggestion:  "Review query structure and consider optimization",
		})
	}

	// Check for nested loops on large tables
	if plan.NodeType == "Nested Loop" {
		if plan.PlanRows > 1000 || plan.ActualRows > 1000 {
			*issues = append(*issues, Issue{
				Severity:    "high",
				Type:        "nested_loop",
				Description: fmt.Sprintf("Nested loop with %.0f rows (consider hash or merge join)", plan.ActualRows),
				Table:       tableName,
				Columns:     nil,
				Suggestion:  "Consider rewriting join or adding indexes to enable hash/merge join",
			})
		}
	}

	// Check for Sort without index
	if plan.NodeType == "Sort" && len(plan.SortKey) > 0 {
		*issues = append(*issues, Issue{
			Severity:    "low",
			Type:        "sort_without_index",
			Description: fmt.Sprintf("Sort operation on %v (might benefit from index)", plan.SortKey),
			Table:       tableName,
			Columns:     plan.SortKey,
			Suggestion:  "Consider adding an index to support this sort order",
		})
	}

	// Recurse into child plans
	for i := range plan.Plans {
		detectIssuesRecursive(&plan.Plans[i], tableRowCounts, threshold, issues)
	}
}

// extractColumnsFromFilter extracts column names from a filter expression
func extractColumnsFromFilter(filter string) []string {
	if filter == "" {
		return []string{}
	}

	// Simple regex to extract column-like identifiers
	re := regexp.MustCompile(`\b([a-zA-Z_][a-zA-Z0-9_]*)\b`)
	matches := re.FindAllStringSubmatch(filter, -1)

	// Filter out common SQL keywords
	keywords := map[string]bool{
		"AND": true, "OR": true, "NOT": true, "IN": true, "IS": true,
		"NULL": true, "TRUE": true, "FALSE": true, "LIKE": true,
		"BETWEEN": true, "ANY": true, "ALL": true, "SOME": true,
	}

	columns := make([]string, 0)
	seen := make(map[string]bool)
	for _, match := range matches {
		col := match[1]
		if !keywords[strings.ToUpper(col)] && !seen[col] {
			columns = append(columns, col)
			seen[col] = true
		}
	}

	return columns
}

// getSeverity determines issue severity based on row count
func getSeverity(rowCount, threshold int64) string {
	if rowCount > threshold*10 {
		return "critical"
	} else if rowCount > threshold*5 {
		return "high"
	} else if rowCount > threshold {
		return "medium"
	}
	return "low"
}

// generateIndexSuggestions creates index suggestions from detected issues
func generateIndexSuggestions(issues []Issue, existingIndexes map[string][]string, includeDDL bool) []IndexSuggestion {
	suggestions := make([]IndexSuggestion, 0)
	seen := make(map[string]bool)

	for _, issue := range issues {
		if issue.Type != "seq_scan" || len(issue.Columns) == 0 {
			continue
		}

		// Check if index already exists
		key := fmt.Sprintf("%s:%s", issue.Table, strings.Join(issue.Columns, ","))
		if seen[key] {
			continue
		}
		seen[key] = true

		// Check against existing indexes
		if indexes, ok := existingIndexes[issue.Table]; ok {
			indexExists := false
			for _, existingCols := range indexes {
				if strings.Contains(existingCols, issue.Columns[0]) {
					indexExists = true
					break
				}
			}
			if indexExists {
				continue
			}
		}

		suggestion := IndexSuggestion{
			Table:     issue.Table,
			Columns:   issue.Columns,
			IndexType: "btree",
			Reason:    issue.Description,
			EstimatedImprovement: issue.Severity,
		}

		if includeDDL {
			suggestion.DDL = generateIndexDDL(issue.Table, issue.Columns)
		}

		suggestions = append(suggestions, suggestion)
	}

	return suggestions
}

// generateIndexDDL creates a CREATE INDEX statement
func generateIndexDDL(table string, columns []string) string {
	// Extract schema and table name
	parts := strings.Split(table, ".")
	schema := "public"
	tableName := table
	if len(parts) == 2 {
		schema = parts[0]
		tableName = parts[1]
	}

	// Generate index name
	indexName := fmt.Sprintf("idx_%s_%s", tableName, strings.Join(columns, "_"))
	if len(indexName) > 63 { // PostgreSQL identifier limit
		indexName = indexName[:63]
	}

	return fmt.Sprintf("CREATE INDEX CONCURRENTLY %s ON %s.%s (%s);",
		indexName, schema, tableName, strings.Join(columns, ", "))
}

// runMultipleExplains runs EXPLAIN ANALYZE multiple times and averages results
func runMultipleExplains(ctx context.Context, conn db.Connection, query string, runs int) (map[string]interface{}, error) {
	var totalExecTime, totalPlanTime, totalCost float64
	var totalBuffersHit, totalBuffersRead int
	var planSummary string
	successfulRuns := 0

	for i := 0; i < runs; i++ {
		explainQuery := fmt.Sprintf("EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) %s", query)

		result, err := conn.ExecuteQuery(ctx, explainQuery)
		if err != nil {
			return nil, err
		}

		if len(result.Rows) == 0 || len(result.Rows[0]) == 0 {
			continue
		}

		var explainResults []ExplainResult
		if err := json.Unmarshal([]byte(result.Rows[0][0]), &explainResults); err != nil {
			continue
		}

		if len(explainResults) > 0 {
			successfulRuns++
			totalExecTime += explainResults[0].ExecutionTime
			totalPlanTime += explainResults[0].PlanningTime
			totalCost += explainResults[0].Plan.TotalCost
			totalBuffersHit += explainResults[0].Plan.SharedHitBlocks
			totalBuffersRead += explainResults[0].Plan.SharedReadBlocks

			if planSummary == "" {
				planSummary = buildPlanSummary(&explainResults[0].Plan)
			}
		}
	}

	if successfulRuns == 0 {
		return nil, fmt.Errorf("no successful explain runs")
	}

	runsFloat := float64(successfulRuns)
	return map[string]interface{}{
		"planning_time_ms":  totalPlanTime / runsFloat,
		"execution_time_ms": totalExecTime / runsFloat,
		"total_cost":        totalCost / runsFloat,
		"buffers_hit":       float64(totalBuffersHit) / runsFloat,
		"buffers_read":      float64(totalBuffersRead) / runsFloat,
		"plan_summary":      planSummary,
	}, nil
}

// buildPlanSummary creates a human-readable summary of the plan
func buildPlanSummary(plan *PlanNode) string {
	var parts []string
	buildPlanSummaryRecursive(plan, &parts)
	return strings.Join(parts, " -> ")
}

func buildPlanSummaryRecursive(plan *PlanNode, parts *[]string) {
	nodeDesc := plan.NodeType
	if plan.RelationName != "" {
		nodeDesc += " on " + plan.RelationName
	}
	*parts = append(*parts, nodeDesc)

	for i := range plan.Plans {
		buildPlanSummaryRecursive(&plan.Plans[i], parts)
	}
}

// calculateImprovement calculates percentage improvement (negative = better)
func calculateImprovement(before, after float64) float64 {
	if before == 0 {
		if after == 0 {
			return 0
		}
		return 100 // Infinite increase
	}
	return ((after - before) / before) * 100
}

// calculateBufferHitRatio calculates the buffer cache hit ratio
func calculateBufferHitRatio(metrics map[string]interface{}) float64 {
	hit, hitOk := metrics["buffers_hit"].(float64)
	read, readOk := metrics["buffers_read"].(float64)
	if !hitOk || !readOk {
		return 100 // Assume good hit ratio if data unavailable
	}
	total := hit + read
	if total == 0 {
		return 100
	}
	return (hit / total) * 100
}
