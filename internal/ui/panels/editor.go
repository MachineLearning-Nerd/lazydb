package panels

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textarea"
)

// EditorPanel represents the center panel for query editing
type EditorPanel struct {
	width    int
	height   int
	textarea textarea.Model
}

// NewEditorPanel creates a new editor panel
func NewEditorPanel() *EditorPanel {
	ta := textarea.New()
	ta.Placeholder = "Enter SQL query here..."
	ta.SetValue("SELECT * FROM pg_database;")
	ta.ShowLineNumbers = true
	ta.CharLimit = 10000 // Reasonable limit for SQL queries
	ta.Focus()

	// Set some style preferences
	ta.FocusedStyle.CursorLine = ta.FocusedStyle.CursorLine

	return &EditorPanel{
		textarea: ta,
	}
}

// SetSize sets the panel dimensions
func (p *EditorPanel) SetSize(width, height int) {
	p.width = width
	p.height = height

	// Update textarea size (leave room for title)
	if height > 4 {
		p.textarea.SetHeight(height - 2)
	}
	if width > 4 {
		p.textarea.SetWidth(width - 2)
	}
}

// Update handles textarea updates
func (p *EditorPanel) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	p.textarea, cmd = p.textarea.Update(msg)
	return cmd
}

// View renders the editor panel
func (p *EditorPanel) View() string {
	if p.width == 0 || p.height == 0 {
		return ""
	}

	content := "QUERY EDITOR\n\n"
	content += p.textarea.View()

	return content
}

// GetQuery returns the current query text
func (p *EditorPanel) GetQuery() string {
	return p.textarea.Value()
}

// SetQuery sets the query text
func (p *EditorPanel) SetQuery(query string) {
	p.textarea.SetValue(query)
}

// Focus sets focus on the textarea
func (p *EditorPanel) Focus() tea.Cmd {
	return p.textarea.Focus()
}

// Blur removes focus from the textarea
func (p *EditorPanel) Blur() {
	p.textarea.Blur()
}

// Help returns help text for the editor panel
func (p *EditorPanel) Help() string {
	return "[Ctrl-R] Execute  [F2] Save  [Ctrl-A/E] line start/end  [Ctrl-K/U] delete line"
}
