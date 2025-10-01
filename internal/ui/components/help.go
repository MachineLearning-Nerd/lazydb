package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HelpCategory represents a category of help queries
type HelpCategory struct {
	Name    string
	Queries []HelpQuery
}

// HelpQuery represents a single query with description
type HelpQuery struct {
	Description string
	Query       string
}

// HelpDialog represents the PostgreSQL help reference dialog
type HelpDialog struct {
	categories      []HelpCategory
	selectedCat     int
	selectedQuery   int
	scrollOffset    int
	width           int
	height          int
	maxVisibleLines int
}

// NewHelpDialog creates a new help dialog
func NewHelpDialog() *HelpDialog {
	return &HelpDialog{
		categories:      getHelpCategories(),
		selectedCat:     0,
		selectedQuery:   0,
		scrollOffset:    0,
		width:           100,
		height:          35,
		maxVisibleLines: 20,
	}
}

// getHelpCategories returns all help categories with queries
func getHelpCategories() []HelpCategory {
	return []HelpCategory{
		{
			Name: "Database & Schema",
			Queries: []HelpQuery{
				{
					Description: "List all schemas",
					Query:       "SELECT schema_name FROM information_schema.schemata ORDER BY schema_name;",
				},
				{
					Description: "List all tables in current schema",
					Query:       "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name;",
				},
				{
					Description: "Get database size",
					Query:       "SELECT pg_size_pretty(pg_database_size(current_database()));",
				},
				{
					Description: "List all databases",
					Query:       "SELECT datname FROM pg_database WHERE datistemplate = false;",
				},
			},
		},
		{
			Name: "Table Structure",
			Queries: []HelpQuery{
				{
					Description: "Get table columns and types",
					Query:       "SELECT column_name, data_type, character_maximum_length, is_nullable, column_default FROM information_schema.columns WHERE table_name = 'your_table' ORDER BY ordinal_position;",
				},
				{
					Description: "Get table constraints",
					Query:       "SELECT constraint_name, constraint_type FROM information_schema.table_constraints WHERE table_name = 'your_table';",
				},
				{
					Description: "Get foreign key relationships",
					Query:       "SELECT tc.table_name, kcu.column_name, ccu.table_name AS foreign_table, ccu.column_name AS foreign_column FROM information_schema.table_constraints tc JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name JOIN information_schema.constraint_column_usage ccu ON ccu.constraint_name = tc.constraint_name WHERE tc.constraint_type = 'FOREIGN KEY';",
				},
			},
		},
		{
			Name: "Indexes",
			Queries: []HelpQuery{
				{
					Description: "List all indexes for a table",
					Query:       "SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'your_table';",
				},
				{
					Description: "Find unused indexes",
					Query:       "SELECT schemaname, tablename, indexname, idx_scan FROM pg_stat_user_indexes WHERE idx_scan = 0 AND indexrelname NOT LIKE 'pg_toast%';",
				},
				{
					Description: "Get index size",
					Query:       "SELECT pg_size_pretty(pg_relation_size('index_name'));",
				},
			},
		},
		{
			Name: "Functions & Procedures",
			Queries: []HelpQuery{
				{
					Description: "List all functions",
					Query:       "SELECT routine_name, routine_type FROM information_schema.routines WHERE routine_schema = 'public';",
				},
				{
					Description: "Get function definition",
					Query:       "SELECT pg_get_functiondef(oid) FROM pg_proc WHERE proname = 'function_name';",
				},
				{
					Description: "List function parameters",
					Query:       "SELECT parameter_name, data_type, parameter_mode FROM information_schema.parameters WHERE specific_name = 'function_name';",
				},
			},
		},
		{
			Name: "Sequences",
			Queries: []HelpQuery{
				{
					Description: "List all sequences",
					Query:       "SELECT sequence_name FROM information_schema.sequences;",
				},
				{
					Description: "Get current sequence value",
					Query:       "SELECT currval('sequence_name');",
				},
				{
					Description: "Get next sequence value",
					Query:       "SELECT nextval('sequence_name');",
				},
				{
					Description: "Get last value without incrementing",
					Query:       "SELECT last_value FROM sequence_name;",
				},
			},
		},
		{
			Name: "Triggers",
			Queries: []HelpQuery{
				{
					Description: "List all triggers",
					Query:       "SELECT trigger_name, event_manipulation, event_object_table FROM information_schema.triggers;",
				},
				{
					Description: "Get trigger definition",
					Query:       "SELECT pg_get_triggerdef(oid) FROM pg_trigger WHERE tgname = 'trigger_name';",
				},
			},
		},
		{
			Name: "Data Types",
			Queries: []HelpQuery{
				{
					Description: "List custom data types",
					Query:       "SELECT typname FROM pg_type WHERE typtype = 'c';",
				},
				{
					Description: "List enum types and values",
					Query:       "SELECT t.typname, e.enumlabel FROM pg_type t JOIN pg_enum e ON t.oid = e.enumtypid ORDER BY t.typname, e.enumsortorder;",
				},
			},
		},
		{
			Name: "Performance",
			Queries: []HelpQuery{
				{
					Description: "Table sizes (largest first)",
					Query:       "SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size FROM pg_tables WHERE schemaname NOT IN ('pg_catalog', 'information_schema') ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;",
				},
				{
					Description: "Active connections",
					Query:       "SELECT pid, usename, application_name, client_addr, state, query FROM pg_stat_activity WHERE state = 'active';",
				},
				{
					Description: "Long running queries (over 1 minute)",
					Query:       "SELECT pid, now() - query_start as duration, query FROM pg_stat_activity WHERE state = 'active' AND now() - query_start > interval '1 minute';",
				},
				{
					Description: "Cache hit ratio",
					Query:       "SELECT sum(heap_blks_read) as heap_read, sum(heap_blks_hit) as heap_hit, round(sum(heap_blks_hit) / (sum(heap_blks_hit) + sum(heap_blks_read)), 3) as ratio FROM pg_statio_user_tables;",
				},
			},
		},
		{
			Name: "Users & Permissions",
			Queries: []HelpQuery{
				{
					Description: "List all users/roles",
					Query:       "SELECT rolname FROM pg_roles;",
				},
				{
					Description: "List user permissions on tables",
					Query:       "SELECT grantee, table_schema, table_name, privilege_type FROM information_schema.table_privileges WHERE grantee = 'username';",
				},
				{
					Description: "Check current user",
					Query:       "SELECT current_user;",
				},
			},
		},
	}
}

// Navigate handles navigation within the help dialog
func (h *HelpDialog) Navigate(direction string) {
	switch direction {
	case "next_category":
		if h.selectedCat < len(h.categories)-1 {
			h.selectedCat++
			h.selectedQuery = 0
			h.scrollOffset = 0
		}
	case "prev_category":
		if h.selectedCat > 0 {
			h.selectedCat--
			h.selectedQuery = 0
			h.scrollOffset = 0
		}
	case "next_query":
		if h.selectedQuery < len(h.categories[h.selectedCat].Queries)-1 {
			h.selectedQuery++
			h.adjustScroll()
		}
	case "prev_query":
		if h.selectedQuery > 0 {
			h.selectedQuery--
			h.adjustScroll()
		}
	}
}

// adjustScroll adjusts the scroll offset to keep selected query visible
func (h *HelpDialog) adjustScroll() {
	if h.selectedQuery < h.scrollOffset {
		h.scrollOffset = h.selectedQuery
	} else if h.selectedQuery >= h.scrollOffset+h.maxVisibleLines {
		h.scrollOffset = h.selectedQuery - h.maxVisibleLines + 1
	}
}

// GetSelectedQuery returns the currently selected query
func (h *HelpDialog) GetSelectedQuery() string {
	if h.selectedCat >= 0 && h.selectedCat < len(h.categories) {
		queries := h.categories[h.selectedCat].Queries
		if h.selectedQuery >= 0 && h.selectedQuery < len(queries) {
			return queries[h.selectedQuery].Query
		}
	}
	return ""
}

// View renders the help dialog
func (h *HelpDialog) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("5")).
		Padding(0, 1)

	categoryStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Padding(0, 1)

	selectedCategoryStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("2")).
		Background(lipgloss.Color("236")).
		Padding(0, 1)

	queryDescStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("6"))

	selectedQueryStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("2")).
		Bold(true)

	queryStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7"))

	content := titleStyle.Render("PostgreSQL Quick Reference") + "\n\n"

	// Render category tabs
	var categoryTabs []string
	for i, cat := range h.categories {
		if i == h.selectedCat {
			categoryTabs = append(categoryTabs, selectedCategoryStyle.Render(cat.Name))
		} else {
			categoryTabs = append(categoryTabs, categoryStyle.Render(cat.Name))
		}
	}
	content += strings.Join(categoryTabs, " ") + "\n\n"

	// Render queries for selected category
	currentCategory := h.categories[h.selectedCat]
	visibleQueries := currentCategory.Queries[h.scrollOffset:]
	if len(visibleQueries) > h.maxVisibleLines {
		visibleQueries = visibleQueries[:h.maxVisibleLines]
	}

	for i, query := range visibleQueries {
		actualIndex := i + h.scrollOffset
		prefix := "  "
		if actualIndex == h.selectedQuery {
			prefix = "▶ "
			content += selectedQueryStyle.Render(prefix+query.Description) + "\n"
			// Wrap long queries
			wrappedQuery := wrapText(query.Query, h.width-6)
			content += selectedQueryStyle.Render("  " + wrappedQuery) + "\n\n"
		} else {
			content += queryDescStyle.Render(prefix+query.Description) + "\n"
			wrappedQuery := wrapText(query.Query, h.width-6)
			content += queryStyle.Render("  " + wrappedQuery) + "\n\n"
		}
	}

	// Instructions
	content += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(
		"[←/→] Category  [↑/↓] Query  [Enter] Copy to Editor  [Esc/?] Close",
	)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("5")).
		Padding(1, 2).
		Width(h.width).
		Height(h.height)

	return borderStyle.Render(content)
}

// wrapText wraps text to fit within the specified width
func wrapText(text string, width int) string {
	if len(text) <= width {
		return text
	}

	var lines []string
	for len(text) > width {
		// Find last space before width
		breakPoint := width
		for i := width; i > 0; i-- {
			if text[i] == ' ' {
				breakPoint = i
				break
			}
		}
		lines = append(lines, text[:breakPoint])
		text = text[breakPoint+1:]
	}
	if len(text) > 0 {
		lines = append(lines, text)
	}

	return strings.Join(lines, "\n  ")
}
