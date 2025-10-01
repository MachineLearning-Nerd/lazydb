package panels

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textarea"
)

// EditorMode represents the current editing mode
type EditorMode int

const (
	ModeNormal EditorMode = iota
	ModeInsert
)

// EditorPanel represents the center panel for query editing
type EditorPanel struct {
	width     int
	height    int
	textarea  textarea.Model
	mode      EditorMode
	clipboard string // For yank/paste operations
}

// NewEditorPanel creates a new editor panel
func NewEditorPanel() *EditorPanel {
	ta := textarea.New()
	ta.Placeholder = "Enter SQL query here... (Press ESC for Vim Normal mode, i for Insert mode)"
	ta.SetValue("SELECT * FROM pg_database;")
	ta.ShowLineNumbers = true
	ta.CharLimit = 10000 // Reasonable limit for SQL queries
	ta.Focus()

	// Set some style preferences
	ta.FocusedStyle.CursorLine = ta.FocusedStyle.CursorLine

	return &EditorPanel{
		textarea:  ta,
		mode:      ModeInsert, // Start in insert mode for easier use
		clipboard: "",
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

// Update handles textarea updates with Vim modal editing
func (p *EditorPanel) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// ESC always switches to normal mode
		if msg.String() == "esc" {
			p.mode = ModeNormal
			return nil
		}

		// Handle keys based on current mode
		if p.mode == ModeNormal {
			return p.handleNormalMode(msg)
		} else {
			// In insert mode, pass all keys to textarea
			p.textarea, cmd = p.textarea.Update(msg)
			return cmd
		}

	case deleteLineMsg:
		p.deleteLine()
		return nil

	case yankLineMsg:
		p.yankLine()
		return nil

	case gotoFirstLineMsg:
		p.moveCursorToStart()
		return nil
	}

	// For non-key messages, always update textarea
	p.textarea, cmd = p.textarea.Update(msg)
	return cmd
}

// handleNormalMode handles Vim normal mode keybindings
func (p *EditorPanel) handleNormalMode(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd

	switch msg.String() {
	// Mode switching
	case "i":
		p.mode = ModeInsert
		return nil
	case "a":
		p.mode = ModeInsert
		// Send right arrow key to textarea to move after current char
		p.textarea, cmd = p.textarea.Update(tea.KeyMsg{Type: tea.KeyRight})
		return cmd

	// Basic movement - pass through to textarea
	case "h":
		p.textarea, cmd = p.textarea.Update(tea.KeyMsg{Type: tea.KeyLeft})
		return cmd
	case "j":
		p.textarea, cmd = p.textarea.Update(tea.KeyMsg{Type: tea.KeyDown})
		return cmd
	case "k":
		p.textarea, cmd = p.textarea.Update(tea.KeyMsg{Type: tea.KeyUp})
		return cmd
	case "l":
		p.textarea, cmd = p.textarea.Update(tea.KeyMsg{Type: tea.KeyRight})
		return cmd
	case "left":
		p.textarea, cmd = p.textarea.Update(tea.KeyMsg{Type: tea.KeyLeft})
		return cmd
	case "down":
		p.textarea, cmd = p.textarea.Update(tea.KeyMsg{Type: tea.KeyDown})
		return cmd
	case "up":
		p.textarea, cmd = p.textarea.Update(tea.KeyMsg{Type: tea.KeyUp})
		return cmd
	case "right":
		p.textarea, cmd = p.textarea.Update(tea.KeyMsg{Type: tea.KeyRight})
		return cmd

	// Line movement
	case "0":
		p.textarea, cmd = p.textarea.Update(tea.KeyMsg{Type: tea.KeyHome})
		return cmd
	case "$":
		p.textarea, cmd = p.textarea.Update(tea.KeyMsg{Type: tea.KeyEnd})
		return cmd

	// Word movement (approximate with Ctrl+Left/Right)
	case "w":
		p.textarea, cmd = p.textarea.Update(tea.KeyMsg{Type: tea.KeyRight, Alt: true})
		return cmd
	case "b":
		p.textarea, cmd = p.textarea.Update(tea.KeyMsg{Type: tea.KeyLeft, Alt: true})
		return cmd

	// Delete operations
	case "d":
		// Wait for next 'd' to delete line
		return p.waitForSecondD()

	// Yank (copy) operations
	case "y":
		// Wait for next 'y' to yank line
		return p.waitForSecondY()

	// Paste
	case "p":
		p.pasteLine()
		return nil

	// Document movement
	case "g":
		return p.waitForSecondG()
	case "G":
		p.moveCursorToEnd()
		return nil
	}

	return nil
}

// Vim operation helpers
func (p *EditorPanel) deleteLine() {
	// Get current content
	value := p.textarea.Value()
	lines := strings.Split(value, "\n")

	if len(lines) == 0 {
		return
	}

	// For simplicity, delete first line and save to clipboard
	// (Proper implementation would track cursor position)
	if len(lines) > 0 {
		p.clipboard = lines[0]
		if len(lines) > 1 {
			p.textarea.SetValue(strings.Join(lines[1:], "\n"))
		} else {
			p.textarea.SetValue("")
		}
	}
}

func (p *EditorPanel) yankLine() {
	// Get current content
	value := p.textarea.Value()
	lines := strings.Split(value, "\n")

	// For simplicity, yank first line
	if len(lines) > 0 {
		p.clipboard = lines[0]
	}
}

func (p *EditorPanel) pasteLine() {
	if p.clipboard == "" {
		return
	}

	// Append clipboard content to the end
	value := p.textarea.Value()
	if value == "" {
		p.textarea.SetValue(p.clipboard)
	} else {
		p.textarea.SetValue(value + "\n" + p.clipboard)
	}
}

func (p *EditorPanel) moveCursorToStart() {
	// Move to beginning of document
	value := p.textarea.Value()
	lines := strings.Split(value, "\n")

	// Send up key multiple times to reach top
	for i := 0; i < len(lines)*2; i++ { // *2 to ensure we reach top
		p.textarea, _ = p.textarea.Update(tea.KeyMsg{Type: tea.KeyUp})
	}
	// Then move to start of line
	p.textarea, _ = p.textarea.Update(tea.KeyMsg{Type: tea.KeyHome})
}

func (p *EditorPanel) moveCursorToEnd() {
	// Move to end of document
	value := p.textarea.Value()
	lines := strings.Split(value, "\n")

	// Send down key multiple times to reach bottom
	for i := 0; i < len(lines)*2; i++ { // *2 to ensure we reach bottom
		p.textarea, _ = p.textarea.Update(tea.KeyMsg{Type: tea.KeyDown})
	}
	// Then move to end of line
	p.textarea, _ = p.textarea.Update(tea.KeyMsg{Type: tea.KeyEnd})
}

// Command handlers for multi-key sequences
func (p *EditorPanel) waitForSecondD() tea.Cmd {
	return func() tea.Msg {
		return deleteLineMsg{}
	}
}

func (p *EditorPanel) waitForSecondY() tea.Cmd {
	return func() tea.Msg {
		return yankLineMsg{}
	}
}

func (p *EditorPanel) waitForSecondG() tea.Cmd {
	return func() tea.Msg {
		return gotoFirstLineMsg{}
	}
}

// Message types for Vim operations
type deleteLineMsg struct{}
type yankLineMsg struct{}
type gotoFirstLineMsg struct{}

// View renders the editor panel with mode indicator
func (p *EditorPanel) View() string {
	if p.width == 0 || p.height == 0 {
		return ""
	}

	// Show mode indicator
	modeIndicator := ""
	if p.mode == ModeNormal {
		modeIndicator = " -- NORMAL --"
	} else {
		modeIndicator = " -- INSERT --"
	}

	content := "QUERY EDITOR" + modeIndicator + "\n\n"
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
	if p.mode == ModeNormal {
		return "[Ctrl-R] Execute  [F2] Save  [Ctrl-E] Neovim  [i/a] Insert  [hjkl] Move  [dd] Delete  [yy] Yank  [p] Paste"
	}
	return "[Ctrl-R] Execute  [F2] Save  [Ctrl-E] Neovim  [ESC] Normal mode"
}
