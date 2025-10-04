package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/MachineLearning-Nerd/lazydb/internal/db"
)

// SchemaContext holds database schema information for AI context
type SchemaContext struct {
	ConnectionName string
	DBType         string // PostgreSQL, MySQL, etc.
	Database       string
	Schemas        []SchemaInfo
}

// SchemaInfo holds information about a database schema
type SchemaInfo struct {
	Name      string
	Tables    []TableInfo
	Views     []ViewInfo
	Functions []FunctionInfo
}

// TableInfo holds information about a table
type TableInfo struct {
	Name    string
	Schema  string
	Columns []ColumnInfo
}

// ViewInfo holds information about a view
type ViewInfo struct {
	Name   string
	Schema string
}

// FunctionInfo holds information about a function
type FunctionInfo struct {
	Name   string
	Schema string
}

// ColumnInfo holds information about a column
type ColumnInfo struct {
	Name     string
	Type     string
	Nullable bool
	Default  string
}

// ContextFormat represents the output format for schema context
type ContextFormat string

const (
	FormatSQLComments ContextFormat = "comments"
	FormatMarkdown    ContextFormat = "markdown"
	FormatPlainText   ContextFormat = "plaintext"
	FormatMinimal     ContextFormat = "minimal"
)

// BuildSchemaContext extracts schema information from a database connection
func BuildSchemaContext(ctx context.Context, conn db.Connection, maxTables int, includeColumns bool) (*SchemaContext, error) {
	config := conn.Config()

	schemaCtx := &SchemaContext{
		ConnectionName: config.Name,
		DBType:         "PostgreSQL", // TODO: Detect database type
		Database:       config.Database,
		Schemas:        []SchemaInfo{},
	}

	// Get all schemas
	schemaNames, err := conn.ListSchemas(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list schemas: %w", err)
	}

	// Limit schemas to common ones if there are too many
	filteredSchemas := filterCommonSchemas(schemaNames)

	tablesProcessed := 0
	for _, schemaName := range filteredSchemas {
		if maxTables > 0 && tablesProcessed >= maxTables {
			break
		}

		schemaInfo := SchemaInfo{
			Name:      schemaName,
			Tables:    []TableInfo{},
			Views:     []ViewInfo{},
			Functions: []FunctionInfo{},
		}

		// Get tables
		tables, err := conn.ListTables(ctx, schemaName)
		if err == nil {
			for _, table := range tables {
				if maxTables > 0 && tablesProcessed >= maxTables {
					break
				}

				tableInfo := TableInfo{
					Name:    table.Name,
					Schema:  schemaName,
					Columns: []ColumnInfo{},
				}

				// Get columns if requested
				if includeColumns {
					columns, err := conn.GetTableColumns(ctx, schemaName, table.Name)
					if err == nil {
						for _, col := range columns {
							tableInfo.Columns = append(tableInfo.Columns, ColumnInfo{
								Name:     col.Name,
								Type:     col.Type,
								Nullable: col.Nullable,
								Default:  col.Default,
							})
						}
					}
				}

				schemaInfo.Tables = append(schemaInfo.Tables, tableInfo)
				tablesProcessed++
			}
		}

		// Get views
		views, err := conn.ListViews(ctx, schemaName)
		if err == nil {
			for _, view := range views {
				schemaInfo.Views = append(schemaInfo.Views, ViewInfo{
					Name:   view.Name,
					Schema: schemaName,
				})
			}
		}

		// Get functions (limit to avoid clutter)
		functions, err := conn.ListFunctions(ctx, schemaName)
		if err == nil {
			for i, fn := range functions {
				if i >= 10 { // Limit functions to 10
					break
				}
				schemaInfo.Functions = append(schemaInfo.Functions, FunctionInfo{
					Name:   fn.Name,
					Schema: schemaName,
				})
			}
		}

		// Only add schema if it has content
		if len(schemaInfo.Tables) > 0 || len(schemaInfo.Views) > 0 {
			schemaCtx.Schemas = append(schemaCtx.Schemas, schemaInfo)
		}
	}

	return schemaCtx, nil
}

// filterCommonSchemas filters out system schemas and returns common ones
func filterCommonSchemas(schemas []string) []string {
	var filtered []string
	skipSchemas := map[string]bool{
		"pg_catalog":        true,
		"information_schema": true,
		"pg_toast":          true,
		"pg_temp_1":         true,
	}

	for _, schema := range schemas {
		if !skipSchemas[schema] {
			filtered = append(filtered, schema)
		}
	}

	// If we filtered everything, return public at least
	if len(filtered) == 0 && len(schemas) > 0 {
		filtered = append(filtered, "public")
	}

	return filtered
}

// FormatAsComments formats the schema context as SQL comments
func (sc *SchemaContext) FormatAsComments() string {
	var sb strings.Builder

	sb.WriteString("-- ╔════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("-- ║           LAZYDB SCHEMA CONTEXT                           ║\n")
	sb.WriteString("-- ╚════════════════════════════════════════════════════════════╝\n")
	sb.WriteString(fmt.Sprintf("-- Connection: %s (%s)\n", sc.ConnectionName, sc.DBType))
	sb.WriteString(fmt.Sprintf("-- Database: %s\n", sc.Database))
	sb.WriteString("--\n")

	for _, schema := range sc.Schemas {
		sb.WriteString(fmt.Sprintf("-- Schema: %s\n", schema.Name))

		if len(schema.Tables) > 0 {
			sb.WriteString("--   Tables:\n")
			for _, table := range schema.Tables {
				if len(table.Columns) > 0 {
					// Format with columns
					columnStrs := make([]string, len(table.Columns))
					for i, col := range table.Columns {
						nullStr := ""
						if !col.Nullable {
							nullStr = " NOT NULL"
						}
						columnStrs[i] = fmt.Sprintf("%s %s%s", col.Name, col.Type, nullStr)
					}
					sb.WriteString(fmt.Sprintf("--     - %s (%s)\n", table.Name, strings.Join(columnStrs, ", ")))
				} else {
					sb.WriteString(fmt.Sprintf("--     - %s\n", table.Name))
				}
			}
		}

		if len(schema.Views) > 0 {
			sb.WriteString("--   Views:\n")
			for _, view := range schema.Views {
				sb.WriteString(fmt.Sprintf("--     - %s\n", view.Name))
			}
		}

		if len(schema.Functions) > 0 {
			sb.WriteString("--   Functions:\n")
			for _, fn := range schema.Functions {
				sb.WriteString(fmt.Sprintf("--     - %s()\n", fn.Name))
			}
		}

		sb.WriteString("--\n")
	}

	sb.WriteString("-- ══════════════════════════════════════════════════════════════\n")
	sb.WriteString("\n")

	return sb.String()
}

// FormatAsMarkdown formats the schema context as Markdown
func (sc *SchemaContext) FormatAsMarkdown() string {
	var sb strings.Builder

	sb.WriteString("# Database Schema Context\n\n")
	sb.WriteString(fmt.Sprintf("**Connection:** %s (%s)  \n", sc.ConnectionName, sc.DBType))
	sb.WriteString(fmt.Sprintf("**Database:** %s\n\n", sc.Database))

	for _, schema := range sc.Schemas {
		sb.WriteString(fmt.Sprintf("## Schema: `%s`\n\n", schema.Name))

		if len(schema.Tables) > 0 {
			sb.WriteString("### Tables\n\n")
			for _, table := range schema.Tables {
				if len(table.Columns) > 0 {
					sb.WriteString(fmt.Sprintf("**`%s`**\n", table.Name))
					sb.WriteString("| Column | Type | Nullable |\n")
					sb.WriteString("|--------|------|----------|\n")
					for _, col := range table.Columns {
						nullable := "✓"
						if !col.Nullable {
							nullable = "✗"
						}
						sb.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n", col.Name, col.Type, nullable))
					}
					sb.WriteString("\n")
				} else {
					sb.WriteString(fmt.Sprintf("- `%s`\n", table.Name))
				}
			}
			sb.WriteString("\n")
		}

		if len(schema.Views) > 0 {
			sb.WriteString("### Views\n\n")
			for _, view := range schema.Views {
				sb.WriteString(fmt.Sprintf("- `%s`\n", view.Name))
			}
			sb.WriteString("\n")
		}

		if len(schema.Functions) > 0 {
			sb.WriteString("### Functions\n\n")
			for _, fn := range schema.Functions {
				sb.WriteString(fmt.Sprintf("- `%s()`\n", fn.Name))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// FormatAsPlainText formats the schema context as plain text
func (sc *SchemaContext) FormatAsPlainText() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Database: %s (%s %s)\n", sc.Database, sc.ConnectionName, sc.DBType))
	sb.WriteString("Schemas:\n")

	for _, schema := range sc.Schemas {
		sb.WriteString(fmt.Sprintf("\n%s:\n", schema.Name))

		if len(schema.Tables) > 0 {
			sb.WriteString("  Tables:\n")
			for _, table := range schema.Tables {
				if len(table.Columns) > 0 {
					columnStrs := make([]string, len(table.Columns))
					for i, col := range table.Columns {
						columnStrs[i] = fmt.Sprintf("%s:%s", col.Name, col.Type)
					}
					sb.WriteString(fmt.Sprintf("    %s (%s)\n", table.Name, strings.Join(columnStrs, ", ")))
				} else {
					sb.WriteString(fmt.Sprintf("    %s\n", table.Name))
				}
			}
		}

		if len(schema.Views) > 0 {
			sb.WriteString("  Views: ")
			viewNames := make([]string, len(schema.Views))
			for i, view := range viewNames {
				viewNames[i] = view
			}
			sb.WriteString(strings.Join(viewNames, ", "))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// FormatAsMinimal formats the schema context in a compact format
func (sc *SchemaContext) FormatAsMinimal() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("DB: %s | ", sc.Database))

	var tableNames []string
	for _, schema := range sc.Schemas {
		for _, table := range schema.Tables {
			if schema.Name == "public" {
				tableNames = append(tableNames, table.Name)
			} else {
				tableNames = append(tableNames, fmt.Sprintf("%s.%s", schema.Name, table.Name))
			}
		}
	}

	sb.WriteString(fmt.Sprintf("Tables: %s", strings.Join(tableNames, ", ")))

	return sb.String()
}

// Format formats the schema context using the specified format
func (sc *SchemaContext) Format(format ContextFormat) string {
	switch format {
	case FormatSQLComments:
		return sc.FormatAsComments()
	case FormatMarkdown:
		return sc.FormatAsMarkdown()
	case FormatPlainText:
		return sc.FormatAsPlainText()
	case FormatMinimal:
		return sc.FormatAsMinimal()
	default:
		return sc.FormatAsComments()
	}
}

// StripContextComments removes LazyDB schema context comments from SQL
func StripContextComments(sql string) string {
	lines := strings.Split(sql, "\n")
	var result []string
	inContext := false

	for _, line := range lines {
		// Detect start of context (look for box-drawing header or text marker)
		// This catches the top border line and prevents accumulation
		if !inContext && (strings.Contains(line, "-- ╔") || strings.Contains(line, "-- ║") ||
			strings.Contains(line, "-- ╚") || strings.Contains(line, "LAZYDB SCHEMA CONTEXT")) {
			inContext = true
			continue
		}

		// Detect end of context
		if inContext && strings.Contains(line, "══════════════════════════════════════════════════════════════") {
			inContext = false
			continue
		}

		// Skip lines within context block
		if inContext {
			continue
		}

		result = append(result, line)
	}

	// Remove all leading comment-only lines (leftover from context)
	// This prevents accumulation of blank "--" lines
	for len(result) > 0 {
		trimmed := strings.TrimSpace(result[0])
		if trimmed == "" || trimmed == "--" {
			result = result[1:] // Remove leading blank or comment-only line
		} else {
			break
		}
	}

	// Join and trim leading whitespace
	return strings.TrimSpace(strings.Join(result, "\n"))
}
