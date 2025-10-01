package panels

// ResultsPanel represents the right panel showing query results
type ResultsPanel struct {
	width  int
	height int
}

// NewResultsPanel creates a new results panel
func NewResultsPanel() *ResultsPanel {
	return &ResultsPanel{}
}

// SetSize sets the panel dimensions
func (p *ResultsPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// View renders the results panel
func (p *ResultsPanel) View() string {
	if p.width == 0 || p.height == 0 {
		return ""
	}

	content := "RESULTS\n\n"
	content += "╔═══╦═══════╦═══════╗\n"
	content += "║ id║ name  ║ email ║\n"
	content += "╠═══╬═══════╬═══════╣\n"
	content += "║ 1 ║ Alice ║ a@... ║\n"
	content += "║ 2 ║ Bob   ║ b@... ║\n"
	content += "╚═══╩═══════╩═══════╝\n\n"
	content += "10 rows (42ms)\n"

	return content
}

// Help returns help text for the results panel
func (p *ResultsPanel) Help() string {
	return "[j/k] scroll  [y] copy  [e] export"
}
