package panels

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/MachineLearning-Nerd/lazydb/internal/db"
	"github.com/MachineLearning-Nerd/lazydb/internal/ui/components"
)

// ViewMode represents what's being displayed in the connections panel
type ViewMode int

const (
	ViewConnections ViewMode = iota
	ViewSchema
)

// ConnectionsPanel represents the left panel showing database connections
type ConnectionsPanel struct {
	width         int
	height        int
	connMgr       *db.ConnectionManager
	selectedIndex int // Currently selected connection (for navigation)
	viewMode      ViewMode
	schemaTree    *components.SchemaTree
	ctx           context.Context
}

// NewConnectionsPanel creates a new connections panel
func NewConnectionsPanel(connMgr *db.ConnectionManager, ctx context.Context) *ConnectionsPanel {
	return &ConnectionsPanel{
		connMgr:       connMgr,
		selectedIndex: 0,
		viewMode:      ViewConnections,
		ctx:           ctx,
	}
}

// SetSize sets the panel dimensions
func (p *ConnectionsPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// Update handles key events for the connections panel
func (p *ConnectionsPanel) Update(msg tea.Msg) tea.Cmd {
	// Handle schema messages
	switch msg := msg.(type) {
	case components.SchemasLoadedMsg:
		if p.schemaTree != nil {
			p.schemaTree.HandleSchemasLoaded(msg.Schemas)
		}
		return nil
	case components.SchemaObjectsLoadedMsg:
		if p.schemaTree != nil {
			p.schemaTree.HandleSchemaObjectsLoaded(msg.Schema, msg.Tables, msg.Views, msg.Functions)
		}
		return nil
	case components.TableColumnsLoadedMsg:
		if p.schemaTree != nil {
			p.schemaTree.HandleTableColumnsLoaded(msg.Schema, msg.Table, msg.Columns)
		}
		return nil
	case components.SchemaErrorMsg:
		// Handle error - could add error display
		return nil
	}

	// Handle keyboard events
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Toggle between connections and schema view
		if msg.String() == "s" && p.viewMode == ViewConnections {
			// Check if we have an active connection
			activeConn, err := p.connMgr.GetActive()
			if err == nil && activeConn.Status() == db.StatusConnected {
				p.viewMode = ViewSchema
				p.schemaTree = components.NewSchemaTree(activeConn)
				// Calculate visible rows (leave space for header)
				visibleRows := p.height - 4
				if visibleRows < 5 {
					visibleRows = 5
				}
				p.schemaTree.SetMaxVisibleRows(visibleRows)
				return p.schemaTree.LoadSchemas(p.ctx)
			}
			return nil
		}

		if msg.String() == "esc" && p.viewMode == ViewSchema {
			p.viewMode = ViewConnections
			p.schemaTree = nil
			return nil
		}

		// Route events based on view mode
		if p.viewMode == ViewSchema && p.schemaTree != nil {
			// Check if in search mode
			if p.schemaTree.IsSearchMode() {
				switch msg.String() {
				case "esc":
					// Exit search mode
					p.schemaTree.ExitSearchMode()
					return nil
				case "backspace":
					// Delete last character
					p.schemaTree.DeleteSearchChar()
					return nil
				case "j", "down":
					// Navigate through filtered results
					p.schemaTree.MoveDown()
					return nil
				case "k", "up":
					// Navigate through filtered results
					p.schemaTree.MoveUp()
					return nil
				case "enter", " ":
					// Can still expand/collapse in search mode
					return p.schemaTree.Toggle(p.ctx)
				default:
					// Handle regular character input for search
					// All printable characters (including p, r, etc.) go to search input
					if len(msg.String()) == 1 && msg.Type == tea.KeyRunes {
						p.schemaTree.AddSearchChar(rune(msg.String()[0]))
					}
					return nil
				}
			}

			// Normal navigation mode (not in search)
			switch msg.String() {
			case "/":
				// Enter search mode
				p.schemaTree.EnterSearchMode()
				return nil
			case "r":
				// Refresh schema data from database
				return p.schemaTree.RefreshSchemas(p.ctx)
			case "j", "down":
				p.schemaTree.MoveDown()
			case "k", "up":
				p.schemaTree.MoveUp()
			case "enter", " ":
				return p.schemaTree.Toggle(p.ctx)
			case "p":
				// Generate preview query for selected table
				selected := p.schemaTree.GetSelected()
				if selected != nil && selected.Type == "table" {
					return func() tea.Msg {
						return TablePreviewMsg{
							Schema: selected.Schema,
							Table:  selected.Name,
						}
					}
				}
			}
			return nil
		}

		// Connections view navigation
		if p.viewMode == ViewConnections {
			connNames := p.getConnectionsInDisplayOrder()
			if len(connNames) == 0 {
				return nil
			}

			switch msg.String() {
			case "j", "down":
				// Move selection down
				if p.selectedIndex < len(connNames)-1 {
					p.selectedIndex++
				}
			case "k", "up":
				// Move selection up
				if p.selectedIndex > 0 {
					p.selectedIndex--
				}
			}
		}
	}
	return nil
}

// getConnectionsInDisplayOrder returns connections in the order they're displayed (grouped by environment)
func (p *ConnectionsPanel) getConnectionsInDisplayOrder() []string {
	connNames := p.connMgr.ListConnections()

	// Group connections by environment
	envGroups := make(map[db.Environment][]string)
	for _, name := range connNames {
		conn, err := p.connMgr.GetConnection(name)
		if err != nil {
			continue
		}
		config := conn.Config()
		env := config.Environment
		if env == "" {
			env = db.EnvDevelopment // Default to Development
		}
		envGroups[env] = append(envGroups[env], name)
	}

	// Build ordered list: Development, Staging, Production
	var orderedConnections []string
	envOrder := []db.Environment{db.EnvDevelopment, db.EnvStaging, db.EnvProduction}

	for _, env := range envOrder {
		connections, exists := envGroups[env]
		if exists {
			orderedConnections = append(orderedConnections, connections...)
		}
	}

	return orderedConnections
}

// GetSelectedConnection returns the name of the currently selected connection
func (p *ConnectionsPanel) GetSelectedConnection() string {
	connNames := p.getConnectionsInDisplayOrder()
	if len(connNames) == 0 || p.selectedIndex >= len(connNames) {
		return ""
	}
	return connNames[p.selectedIndex]
}

// IsInSchemaSearchMode returns true if schema explorer is active and in search mode
func (p *ConnectionsPanel) IsInSchemaSearchMode() bool {
	return p.viewMode == ViewSchema && p.schemaTree != nil && p.schemaTree.IsSearchMode()
}

// View renders the connections panel
func (p *ConnectionsPanel) View() string {
	if p.width == 0 || p.height == 0 {
		return ""
	}

	// Render schema view if active
	if p.viewMode == ViewSchema && p.schemaTree != nil {
		content := "SCHEMA EXPLORER\n"
		content += "Press [Esc] to return to connections\n\n"
		content += p.schemaTree.View()
		return content
	}

	// Render connections view
	content := "CONNECTIONS\n\n"

	// Get connections in display order
	orderedConnections := p.getConnectionsInDisplayOrder()
	activeConn := p.connMgr.ActiveName()

	// Clamp selectedIndex to valid range
	if len(orderedConnections) > 0 && p.selectedIndex >= len(orderedConnections) {
		p.selectedIndex = len(orderedConnections) - 1
	}
	if p.selectedIndex < 0 {
		p.selectedIndex = 0
	}

	if len(orderedConnections) == 0 {
		content += "No connections configured\n"
		content += "\nPress 'a' to add a connection"
	} else {
		// Group connections by environment for display
		envGroups := make(map[db.Environment][]string)
		for _, name := range orderedConnections {
			conn, err := p.connMgr.GetConnection(name)
			if err != nil {
				continue
			}
			config := conn.Config()
			env := config.Environment
			if env == "" {
				env = db.EnvDevelopment // Default to Development
			}
			envGroups[env] = append(envGroups[env], name)
		}

		// Render each environment group
		envOrder := []db.Environment{db.EnvDevelopment, db.EnvStaging, db.EnvProduction}
		currentIndex := 0 // Track overall index for selection

		for _, env := range envOrder {
			connections, exists := envGroups[env]
			if !exists || len(connections) == 0 {
				continue
			}

			// Environment header with icon
			var envIcon string
			switch env {
			case db.EnvDevelopment:
				envIcon = "🟢"
			case db.EnvStaging:
				envIcon = "🔵"
			case db.EnvProduction:
				envIcon = "🔴"
			}
			content += fmt.Sprintf("▼ %s %s\n", envIcon, env)

			// Render connections in this environment
			for _, name := range connections {
				conn, err := p.connMgr.GetConnection(name)
				if err != nil {
					currentIndex++
					continue
				}

				// Determine status icon
				statusIcon := "⚪"
				statusText := ""
				switch conn.Status() {
				case db.StatusConnected:
					statusIcon = "🟢"
					statusText = " ✓"
				case db.StatusConnecting:
					statusIcon = "🟡"
					statusText = " ⟳"
				case db.StatusError:
					statusIcon = "🔴"
					statusText = " ✗"
				case db.StatusDisconnected:
					statusIcon = "⚪"
				}

				// Mark selected connection (for navigation) and active connection
				prefix := "  "
				if currentIndex == p.selectedIndex {
					prefix = "> " // Selected (highlighted)
				}
				if name == activeConn {
					prefix = "▶ " // Active (connected)
				}

				config := conn.Config()
				content += fmt.Sprintf("  %s%s %s%s\n", prefix, statusIcon, config.Name, statusText)
				currentIndex++
			}

			content += "\n" // Add spacing between environment groups
		}
	}

	return content
}

// Help returns help text for the connections panel
func (p *ConnectionsPanel) Help() string {
	if p.viewMode == ViewSchema {
		// Check if in search mode
		if p.schemaTree != nil && p.schemaTree.IsSearchMode() {
			return "[type to search]  [Esc] clear search  [j/k] navigate  [Enter] expand  [p] preview"
		}
		return "[j/k] navigate  [Enter/Space] expand  [/] search  [r] refresh  [p] preview  [Esc] back"
	}
	return "[a] add  [d] delete  [e] edit  [Enter] connect  [s] schema"
}

// TablePreviewMsg is sent when user requests a table preview
type TablePreviewMsg struct {
	Schema string
	Table  string
}
