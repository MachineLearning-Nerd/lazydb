package panels

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/MachineLearning-Nerd/lazydb/internal/db"
)

// ConnectionsPanel represents the left panel showing database connections
type ConnectionsPanel struct {
	width         int
	height        int
	connMgr       *db.ConnectionManager
	selectedIndex int // Currently selected connection (for navigation)
}

// NewConnectionsPanel creates a new connections panel
func NewConnectionsPanel(connMgr *db.ConnectionManager) *ConnectionsPanel {
	return &ConnectionsPanel{
		connMgr:       connMgr,
		selectedIndex: 0,
	}
}

// SetSize sets the panel dimensions
func (p *ConnectionsPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// Update handles key events for the connections panel
func (p *ConnectionsPanel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		connNames := p.connMgr.ListConnections()
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
	return nil
}

// GetSelectedConnection returns the name of the currently selected connection
func (p *ConnectionsPanel) GetSelectedConnection() string {
	connNames := p.connMgr.ListConnections()
	if len(connNames) == 0 || p.selectedIndex >= len(connNames) {
		return ""
	}
	return connNames[p.selectedIndex]
}

// View renders the connections panel
func (p *ConnectionsPanel) View() string {
	if p.width == 0 || p.height == 0 {
		return ""
	}

	content := "CONNECTIONS\n\n"

	// List all connections
	connNames := p.connMgr.ListConnections()
	activeConn := p.connMgr.ActiveName()

	// Clamp selectedIndex to valid range
	if len(connNames) > 0 && p.selectedIndex >= len(connNames) {
		p.selectedIndex = len(connNames) - 1
	}
	if p.selectedIndex < 0 {
		p.selectedIndex = 0
	}

	if len(connNames) == 0 {
		content += "No connections configured\n"
		content += "\nPress 'a' to add a connection"
	} else {
		for i, name := range connNames {
			conn, err := p.connMgr.GetConnection(name)
			if err != nil {
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
			if i == p.selectedIndex {
				prefix = "> " // Selected (highlighted)
			}
			if name == activeConn {
				prefix = "▶ " // Active (connected)
			}

			config := conn.Config()
			content += fmt.Sprintf("%s%s %s%s\n", prefix, statusIcon, config.Name, statusText)
		}
	}

	return content
}

// Help returns help text for the connections panel
func (p *ConnectionsPanel) Help() string {
	return "[a] add  [d] delete  [e] edit  [Enter] connect"
}
