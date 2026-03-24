# Execute Query Tool + Key Dispatch Extraction

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a read-only `execute_query` MCP tool and extract key dispatch from the 1,247-line main.go god file into a focused `input.go`.

**Architecture:** The execute_query tool follows the existing BasicTools pattern in `basic.go`, using `pg_query_go` to parse and validate that only SELECT/WITH statements are allowed (including blocking `SELECT INTO`). The key dispatch extraction moves all `tea.KeyMsg` routing into `cmd/lazydb/input.go` as a `handleKeyMsg()` method on the model, eliminating duplicated handlers between the INSERT mode key sink and global keybindings block.

**Tech Stack:** Go, pg_query_go (SQL parsing), bubbletea (TUI framework), MCP protocol

---

## Task 1: Add `IsReadOnlyQuery` to SQL Validator

**Files:**
- Modify: `internal/db/validator.go` (add exported validation function)

- [ ] **Step 1: Add `IsReadOnlyQuery` function to validator.go**

Add at the bottom of `internal/db/validator.go`. This is the canonical location since it already imports `pg_query` and handles SQL validation:

```go
// IsReadOnlyQuery validates that a SQL query is read-only (SELECT or WITH/CTE only).
// Uses pg_query_go to parse the AST and reject any non-SELECT statements.
// Also rejects SELECT INTO (which creates a table).
func IsReadOnlyQuery(query string) error {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return fmt.Errorf("query cannot be empty")
	}

	result, err := pg_query.Parse(trimmed)
	if err != nil {
		return fmt.Errorf("invalid SQL: %w", err)
	}

	if len(result.Stmts) == 0 {
		return fmt.Errorf("no SQL statements found")
	}

	if len(result.Stmts) > 1 {
		return fmt.Errorf("only single statements allowed, found %d", len(result.Stmts))
	}

	stmt := result.Stmts[0].Stmt
	if stmt == nil {
		return fmt.Errorf("empty statement")
	}

	selectNode, ok := stmt.Node.(*pg_query.Node_SelectStmt)
	if !ok {
		return fmt.Errorf("only SELECT queries are allowed (got %T)", stmt.Node)
	}

	// Block SELECT INTO (creates a new table)
	if selectNode.SelectStmt.IntoClause != nil {
		return fmt.Errorf("SELECT INTO is not allowed (creates a table)")
	}

	return nil
}

// HasOuterLimit checks if the outermost SELECT statement has a LIMIT clause.
// Uses AST inspection instead of string matching to avoid false positives
// from column/table names containing "limit".
func HasOuterLimit(query string) bool {
	result, err := pg_query.Parse(strings.TrimSpace(query))
	if err != nil || len(result.Stmts) == 0 {
		return false
	}
	selectNode, ok := result.Stmts[0].Stmt.Node.(*pg_query.Node_SelectStmt)
	if !ok {
		return false
	}
	return selectNode.SelectStmt.LimitCount != nil
}
```

- [ ] **Step 2: Build to verify compilation**

Run: `go build ./internal/db/...`
Expected: SUCCESS

- [ ] **Step 3: Commit**

```bash
git add internal/db/validator.go
git commit -m "feat(db): add IsReadOnlyQuery and HasOuterLimit SQL validation"
```

---

## Task 2: Add Unit Tests for Read-Only Validation

**Files:**
- Create: `tests/unit/execute_query_test.go`

- [ ] **Step 1: Write comprehensive tests**

```go
package unit

import (
	"testing"

	"github.com/MachineLearning-Nerd/lazydb/internal/db"
)

func TestIsReadOnlyQuery_AllowsSelect(t *testing.T) {
	allowed := []string{
		"SELECT * FROM users",
		"SELECT id, name FROM users WHERE id = 1",
		"SELECT * FROM users LIMIT 10",
		"SELECT u.*, o.total FROM users u JOIN orders o ON u.id = o.user_id",
		"WITH cte AS (SELECT * FROM users) SELECT * FROM cte",
		"WITH RECURSIVE tree AS (SELECT id, parent_id FROM categories WHERE parent_id IS NULL UNION ALL SELECT c.id, c.parent_id FROM categories c JOIN tree t ON c.parent_id = t.id) SELECT * FROM tree",
		"SELECT 1",
		"(SELECT 1) UNION (SELECT 2)",
		"SELECT 1 /* ; DROP TABLE users; */",
		"SELECT * FROM rate_limit_events",
	}
	for _, q := range allowed {
		if err := db.IsReadOnlyQuery(q); err != nil {
			t.Errorf("expected query to be allowed: %q, got error: %v", q, err)
		}
	}
}

func TestIsReadOnlyQuery_RejectsDML(t *testing.T) {
	rejected := []struct {
		query string
		desc  string
	}{
		{"INSERT INTO users (name) VALUES ('test')", "INSERT"},
		{"UPDATE users SET name = 'test' WHERE id = 1", "UPDATE"},
		{"DELETE FROM users WHERE id = 1", "DELETE"},
		{"DROP TABLE users", "DROP TABLE"},
		{"CREATE TABLE test (id int)", "CREATE TABLE"},
		{"ALTER TABLE users ADD COLUMN email text", "ALTER TABLE"},
		{"TRUNCATE users", "TRUNCATE"},
		{"", "empty query"},
		{"SELECT 1; DROP TABLE users", "multi-statement injection"},
		{"SELECT * INTO new_table FROM users", "SELECT INTO (creates table)"},
		{"CREATE TABLE evil AS SELECT * FROM users", "CTAS"},
		{"EXPLAIN SELECT * FROM users", "EXPLAIN"},
		{"COPY users TO '/tmp/data.csv'", "COPY"},
	}
	for _, tc := range rejected {
		if err := db.IsReadOnlyQuery(tc.query); err == nil {
			t.Errorf("expected query to be rejected (%s): %q", tc.desc, tc.query)
		}
	}
}

func TestHasOuterLimit(t *testing.T) {
	tests := []struct {
		query    string
		expected bool
	}{
		{"SELECT * FROM users LIMIT 10", true},
		{"SELECT * FROM users", false},
		{"SELECT * FROM rate_limit_events", false},
		{"SELECT * FROM (SELECT * FROM users LIMIT 5) sub", false},
		{"WITH cte AS (SELECT * FROM users LIMIT 1) SELECT * FROM cte", false},
	}
	for _, tc := range tests {
		got := db.HasOuterLimit(tc.query)
		if got != tc.expected {
			t.Errorf("HasOuterLimit(%q) = %v, want %v", tc.query, got, tc.expected)
		}
	}
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./tests/unit/... -run "TestIsReadOnlyQuery|TestHasOuterLimit" -v`
Expected: ALL PASS

- [ ] **Step 3: Commit**

```bash
git add tests/unit/execute_query_test.go
git commit -m "test: add unit tests for read-only query validation"
```

---

## Task 3: Add `execute_query` MCP Tool

**Files:**
- Modify: `internal/mcp/tools/basic.go` (add registration + handler)
- Modify: `internal/mcp/server/categories.go` (add to CategorySchema)

- [ ] **Step 1: Add tool registration in basic.go**

Add before the closing `}` of `func (t *BasicTools) Register(registry *server.ToolRegistry)` (after `get_table_count` registration, before line 154):

```go
	// Execute read-only SQL queries (SELECT/WITH only)
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
```

- [ ] **Step 2: Add the handler method**

```go
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
```

- [ ] **Step 3: Add `db` import to basic.go**

Add to the import block:

```go
"github.com/MachineLearning-Nerd/lazydb/internal/db"
```

- [ ] **Step 4: Add `execute_query` to CategoryTools in categories.go**

In `internal/mcp/server/categories.go`, add `"execute_query"` to the `CategorySchema` slice:

```go
CategorySchema: {
    "list_all_tables",
    "get_table_schema",
    "search_tables",
    "get_table_ddl",
    "get_view_definition",
    "get_function_definition",
    "execute_query",
},
```

- [ ] **Step 5: Build and verify both binaries**

Run: `go build ./cmd/lazydb-mcp/... && go build -o lazydb ./cmd/lazydb`
Expected: SUCCESS

- [ ] **Step 6: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

- [ ] **Step 7: Commit**

```bash
git add internal/mcp/tools/basic.go internal/mcp/server/categories.go
git commit -m "feat(mcp): add execute_query tool for read-only SQL execution"
```

---

## Task 4: Extract Key Dispatch — Create `input.go`

**Files:**
- Create: `cmd/lazydb/input.go`

- [ ] **Step 1: Create `cmd/lazydb/input.go` with the unified key dispatch**

Both files share the `main` package, so `input.go` methods directly access model fields. No interface needed.

Extract ALL key handling from `main.go` into `handleKeyMsg()`. Unify duplicated handlers via shared helper methods. Add `return m, nil` after Help dialog to prevent fall-through.

```go
package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/MachineLearning-Nerd/lazydb/internal/db"
	"github.com/MachineLearning-Nerd/lazydb/internal/editor"
	"github.com/MachineLearning-Nerd/lazydb/internal/storage"
	"github.com/MachineLearning-Nerd/lazydb/internal/ui/components"
)

// handleKeyMsg routes keyboard input to appropriate handlers.
// NOTE: This is a pointer receiver called from Update() which has a value receiver.
// Go takes &m of the Update() local copy. Mutations work because bubbletea uses
// the returned tea.Model, not the original. All code paths must return m.
func (m *model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	var cmd tea.Cmd

	// KEY SINK: Editor in INSERT mode captures all keys except whitelisted globals.
	if m.focusedPanel == FocusEditor && m.editorPanel.IsInInsertMode() && !m.layoutPresetMode {
		switch key {
		case m.config.Keybindings.Global.Quit, "ctrl+c":
			return m.handleQuit()
		case m.config.Keybindings.Global.ExecuteQuery:
			return m, m.executeQuery()
		case m.config.Keybindings.Global.SaveQuery:
			return m, m.saveQuery()
		case m.config.Keybindings.Global.OpenNeovim:
			return m.handleOpenNeovim()
		case m.config.Keybindings.Global.AIAssistant:
			return m.handleAIAssistant()
		default:
			cmd = m.editorPanel.Update(msg)
			return m, cmd
		}
	}

	// Resize keys
	increment := m.config.UI.ResizeIncrement
	switch key {
	case m.config.Keybindings.Resize.ShrinkEditorLeft:
		m.resizePanel("connections", increment)
		m.statusMessage = m.layoutStatus("")
		return m, nil
	case m.config.Keybindings.Resize.GrowEditorLeft:
		m.resizePanel("editor-left", increment)
		m.statusMessage = m.layoutStatus("")
		return m, nil
	case m.config.Keybindings.Resize.ShrinkEditorRight:
		m.resizePanel("results", increment)
		m.statusMessage = m.layoutStatus("")
		return m, nil
	case m.config.Keybindings.Resize.GrowEditorRight:
		m.resizePanel("editor-right", increment)
		m.statusMessage = m.layoutStatus("")
		return m, nil
	}

	// Layout preset mode entry
	if key == m.config.Keybindings.Layout.PresetMode {
		m.layoutPresetMode = true
		m.statusMessage = "Layout preset: [1] Default  [2] Editor  [3] Results  [4] Balanced"
		return m, nil
	}

	// Layout preset selection
	if m.layoutPresetMode {
		switch key {
		case m.config.Keybindings.Layout.PresetDefault:
			m.setLayoutPreset(1)
			m.statusMessage = m.layoutStatus("Default")
			m.layoutPresetMode = false
			return m, nil
		case m.config.Keybindings.Layout.PresetEditor:
			m.setLayoutPreset(2)
			m.statusMessage = m.layoutStatus("Editor Focus")
			m.layoutPresetMode = false
			return m, nil
		case m.config.Keybindings.Layout.PresetResults:
			m.setLayoutPreset(3)
			m.statusMessage = m.layoutStatus("Results Focus")
			m.layoutPresetMode = false
			return m, nil
		case m.config.Keybindings.Layout.PresetBalanced:
			m.setLayoutPreset(4)
			m.statusMessage = m.layoutStatus("Balanced")
			m.layoutPresetMode = false
			return m, nil
		case "esc":
			m.layoutPresetMode = false
			m.statusMessage = ""
			return m, nil
		}
	}

	// Global keybindings
	switch key {
	case m.config.Keybindings.Global.Quit, "ctrl+c":
		return m.handleQuit()
	case m.config.Keybindings.Global.Help:
		m.helpDialog = components.NewHelpDialog()
		m.showDialog = components.DialogTypeHelp
		return m, nil
	case m.config.Keybindings.Global.ExecuteQuery:
		return m, m.executeQuery()
	case m.config.Keybindings.Global.SaveQuery:
		return m, m.saveQuery()
	case m.config.Keybindings.Global.OpenNeovim:
		return m.handleOpenNeovim()
	case m.config.Keybindings.Global.AIAssistant:
		return m.handleAIAssistant()
	}

	// Navigation keybindings
	switch key {
	case m.config.Keybindings.Navigation.NextPanel:
		return m.handleNextPanel()
	case m.config.Keybindings.Navigation.PrevPanel:
		return m.handlePrevPanel()
	case m.config.Keybindings.Navigation.FocusConnections:
		if m.focusedPanel == FocusEditor {
			m.editorPanel.Blur()
		}
		m.focusedPanel = FocusConnections
		return m, nil
	case m.config.Keybindings.Navigation.FocusEditor:
		if m.focusedPanel != FocusEditor {
			m.focusedPanel = FocusEditor
			return m, m.editorPanel.Focus()
		}
		return m, nil
	case m.config.Keybindings.Navigation.FocusResults:
		if m.focusedPanel == FocusEditor {
			m.editorPanel.Blur()
		}
		m.focusedPanel = FocusResults
		return m, nil
	}

	// Connections panel keybindings
	if m.focusedPanel == FocusConnections && !m.connectionsPanel.IsInSchemaSearchMode() {
		switch key {
		case m.config.Keybindings.Connections.Add:
			m.connectionForm = components.NewConnectionFormDialog(components.DialogTypeAdd, nil)
			m.showDialog = components.DialogTypeAdd
			return m, nil
		case m.config.Keybindings.Connections.Edit:
			selectedConn := m.connectionsPanel.GetSelectedConnection()
			if selectedConn != "" {
				conn, err := m.connManager.GetConnection(selectedConn)
				if err == nil {
					connConfig := conn.Config()
					m.connectionForm = components.NewConnectionFormDialog(components.DialogTypeEdit, &connConfig)
					m.showDialog = components.DialogTypeEdit
				}
			}
			return m, nil
		case m.config.Keybindings.Connections.Delete:
			selectedConn := m.connectionsPanel.GetSelectedConnection()
			if selectedConn != "" {
				m.deleteTarget = selectedConn
				m.confirmDialog = components.NewConfirmationDialog(
					fmt.Sprintf("Delete connection '%s'?", selectedConn),
				)
				m.showDialog = components.DialogTypeDelete
			}
			return m, nil
		case m.config.Keybindings.Connections.Connect:
			selectedConn := m.connectionsPanel.GetSelectedConnection()
			if selectedConn != "" {
				m.connManager.SetActive(selectedConn)
				allConns := m.connManager.ListConnections()
				configs := make([]db.ConnectionConfig, 0, len(allConns))
				for _, connName := range allConns {
					if conn, err := m.connManager.GetConnection(connName); err == nil {
						configs = append(configs, conn.Config())
					}
				}
				if err := storage.SaveConnections(configs, selectedConn); err != nil {
					m.logDebug("[WARN] Failed to save active connection: %v", err)
				}
				return m, m.connectToDatabase()
			}
		}
	}

	// Delegate unmatched keys to focused panel
	switch m.focusedPanel {
	case FocusConnections:
		cmd = m.connectionsPanel.Update(msg)
	case FocusEditor:
		cmd = m.editorPanel.Update(msg)
	case FocusResults:
		cmd = m.resultsPanel.Update(msg)
	}

	return m, cmd
}

// --- Helper methods that eliminate duplication ---

func (m *model) handleQuit() (tea.Model, tea.Cmd) {
	m.saveConnections()
	return m, tea.Quit
}

func (m *model) handleOpenNeovim() (tea.Model, tea.Cmd) {
	if editor.IsNvimAvailable() {
		query := m.editorPanel.GetQuery()
		activeConn, err := m.connManager.GetActive()
		var conn db.Connection
		if err == nil {
			conn = activeConn
		}
		injectSchema := true
		if m.config.AI != nil {
			injectSchema = m.config.AI.InjectInNeovim
		}
		return m, editor.OpenInNeovimCmd(query, conn, injectSchema)
	}
	m.statusMessage = "Neovim not found. Please install nvim."
	return m, nil
}

func (m *model) handleAIAssistant() (tea.Model, tea.Cmd) {
	if m.config.AI != nil && m.config.AI.Enabled {
		query := m.editorPanel.GetQuery()
		m.aiAssistant.Show(query)
	} else {
		m.statusMessage = "AI Assistant is disabled. Enable in config.yml"
	}
	return m, nil
}

func (m *model) handleNextPanel() (tea.Model, tea.Cmd) {
	if m.focusedPanel == FocusEditor {
		m.editorPanel.Blur()
	}
	maxPanel := 3
	m.focusedPanel = PanelFocus((int(m.focusedPanel) + 1) % maxPanel)
	if m.focusedPanel == FocusEditor {
		return m, m.editorPanel.Focus()
	}
	return m, nil
}

func (m *model) handlePrevPanel() (tea.Model, tea.Cmd) {
	if m.focusedPanel == FocusEditor {
		m.editorPanel.Blur()
	}
	maxPanel := 3
	m.focusedPanel = PanelFocus((int(m.focusedPanel) + maxPanel - 1) % maxPanel)
	if m.focusedPanel == FocusEditor {
		return m, m.editorPanel.Focus()
	}
	return m, nil
}

func (m *model) layoutStatus(label string) string {
	if label == "" {
		return fmt.Sprintf("Layout: %d%% | %d%% | %d%%", m.connectionRatio, m.editorRatio, m.resultsRatio)
	}
	return fmt.Sprintf("Layout: %d%% | %d%% | %d%% (%s)", m.connectionRatio, m.editorRatio, m.resultsRatio, label)
}
```

- [ ] **Step 2: Move `resizePanel()` and `setLayoutPreset()` from main.go to input.go**

Cut these methods from `main.go` and paste into `input.go`.

- [ ] **Step 3: Verify it compiles (expect failure — Update still references old code)**

Run: `go build -o lazydb ./cmd/lazydb`
Expected: Compile error — proceed to Task 5

---

## Task 5: Refactor `main.go` Update() to Delegate to `handleKeyMsg`

**Files:**
- Modify: `cmd/lazydb/main.go`

- [ ] **Step 1: Replace the `tea.KeyMsg` case in Update()**

Replace the key handling block with delegation to `handleKeyMsg`:

```go
case tea.KeyMsg:
    // Handle AI Assistant if visible (highest priority)
    if m.aiAssistant.IsVisible() {
        cmd = m.aiAssistant.Update(msg)
        return m, cmd
    }

    // Handle dialog interactions
    if m.showDialog != components.DialogTypeNone {
        return m.handleDialogKeys(msg)
    }

    // Delegate to key dispatch (input.go)
    return m.handleKeyMsg(msg)
```

- [ ] **Step 2: KEEP the panel delegation block in Update()**

**IMPORTANT:** Do NOT remove lines 582-595 (`// Update panels based on focus`). This block runs for ALL message types (not just KeyMsg) and handles non-key message propagation to focused panels (e.g., `tea.WindowSizeMsg` forwarding). The `handleKeyMsg` panel delegation only covers `KeyMsg`.

- [ ] **Step 3: Remove the old `resizePanel()` and `setLayoutPreset()` from main.go**

These now live in `input.go`.

- [ ] **Step 4: Clean up imports in main.go**

Remove imports only used by the moved key dispatch code (e.g., `storage`, `editor` if not used elsewhere in main.go). The Go compiler will tell you which imports are unused.

- [ ] **Step 5: Build and verify**

Run: `go build -o lazydb ./cmd/lazydb`
Expected: SUCCESS

- [ ] **Step 6: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

- [ ] **Step 7: Commit**

```bash
git add cmd/lazydb/input.go cmd/lazydb/main.go
git commit -m "refactor: extract key dispatch from main.go into input.go"
```

---

## Task 6: Final Verification

- [ ] **Step 1: Build both binaries**

```bash
go build -o lazydb ./cmd/lazydb
go build -o bin/lazydb-mcp ./cmd/lazydb-mcp
```

- [ ] **Step 2: Run all tests**

```bash
go test ./... -v
```

- [ ] **Step 3: Verify line counts**

```bash
wc -l cmd/lazydb/main.go cmd/lazydb/input.go
```

Expected: main.go ~850 lines, input.go ~370 lines

- [ ] **Step 4: Manual smoke test for TUI** (if possible)

- Launch `./lazydb`
- INSERT mode: type `SELECT * FROM t WHERE x = 1 AND arr[0] = ?` — all chars appear
- ESC to NORMAL: `=`/`-`/`[`/`]` resize, `?` opens help, `1`/`2`/`3` navigate
- `L` then `1`/`2`/`3`/`4` for layout presets
- `ctrl+r` executes query, `ctrl+q` quits

- [ ] **Step 5: Manual smoke test for MCP** (if possible)

Test execute_query via MCP:
- Valid: `{"query": "SELECT 1"}` → returns JSON with columns, rows, row_count, execution_ms
- Rejected: `{"query": "INSERT INTO users VALUES (1)"}` → error
- Rejected: `{"query": "DROP TABLE users"}` → error
- Rejected: `{"query": "SELECT 1; DROP TABLE users"}` → error (multi-statement)
- Rejected: `{"query": "SELECT * INTO new_table FROM users"}` → error (SELECT INTO)
- Auto-limit: `{"query": "SELECT * FROM users"}` → LIMIT 50 injected
- No false positive: `{"query": "SELECT * FROM rate_limit_events"}` → LIMIT injected (AST check, not string match)
