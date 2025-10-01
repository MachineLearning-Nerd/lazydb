package panels

// EditorPanel represents the center panel for query editing
type EditorPanel struct {
	width  int
	height int
	query  string
}

// NewEditorPanel creates a new editor panel
func NewEditorPanel() *EditorPanel {
	return &EditorPanel{
		query: "-- Press 'e' to edit in Nvim\n-- Or type here directly\n\nSELECT * FROM users\nWHERE active = true\nLIMIT 10;",
	}
}

// SetSize sets the panel dimensions
func (p *EditorPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// View renders the editor panel
func (p *EditorPanel) View() string {
	if p.width == 0 || p.height == 0 {
		return ""
	}

	content := "QUERY EDITOR\n\n"
	content += p.query

	return content
}

// Help returns help text for the editor panel
func (p *EditorPanel) Help() string {
	return "[Ctrl-E] Edit in Neovim  [Ctrl-R] Execute  [Ctrl-S] Save query"
}
