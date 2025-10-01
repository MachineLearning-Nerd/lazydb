package panels

// ConnectionsPanel represents the left panel showing database connections
type ConnectionsPanel struct {
	width  int
	height int
}

// NewConnectionsPanel creates a new connections panel
func NewConnectionsPanel() *ConnectionsPanel {
	return &ConnectionsPanel{}
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
	content += "▼ 🟢 Development\n"
	content += "  • dev-local ✓\n"
	content += "  • dev-docker\n\n"
	content += "▶ 🔵 Staging\n"
	content += "  • staging-db\n\n"
	content += "▶ 🔴 Production\n"
	content += "  • prod-master\n"
	content += "  • prod-replica\n"

	return content
}

// Help returns help text for the connections panel
func (p *ConnectionsPanel) Help() string {
	return "[a] add  [d] delete  [e] edit  [Enter] connect"
}
