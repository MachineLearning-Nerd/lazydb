package panels

import (
	"fmt"

	"github.com/MachineLearning-Nerd/lazydb/internal/db"
)

// ConnectionsPanel represents the left panel showing database connections
type ConnectionsPanel struct {
	width     int
	height    int
	connMgr   *db.ConnectionManager
}

// NewConnectionsPanel creates a new connections panel
func NewConnectionsPanel(connMgr *db.ConnectionManager) *ConnectionsPanel {
	return &ConnectionsPanel{
		connMgr: connMgr,
	}
}

// SetSize sets the panel dimensions
func (p *ConnectionsPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
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

	if len(connNames) == 0 {
		content += "No connections configured\n"
		content += "\nPress 'a' to add a connection"
	} else {
		for _, name := range connNames {
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

			// Mark active connection
			prefix := "  "
			if name == activeConn {
				prefix = "▶ "
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
