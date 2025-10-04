package panels

import (
	"context"
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/MachineLearning-Nerd/lazydb/internal/ai"
	"github.com/MachineLearning-Nerd/lazydb/internal/db"
)

// AIAssistantPanel represents the AI assistant dialog
type AIAssistantPanel struct {
	visible      bool
	input        textinput.Model
	response     string
	loading      bool
	error        string
	copyStatus   string // Feedback for copy operation
	provider     ai.CLIProvider
	conn         db.Connection
	currentQuery string
	viewport     viewport.Model
	width        int
	height       int
}

// NewAIAssistantPanel creates a new AI assistant panel
func NewAIAssistantPanel(provider ai.CLIProvider, conn db.Connection) *AIAssistantPanel {
	ti := textinput.New()
	ti.Placeholder = "e.g., optimize this query for large datasets"
	ti.CharLimit = 200
	ti.Width = 60

	vp := viewport.New(60, 10)

	return &AIAssistantPanel{
		visible:  false,
		input:    ti,
		provider: provider,
		conn:     conn,
		viewport: vp,
	}
}

// Show displays the AI assistant panel
func (p *AIAssistantPanel) Show(query string) {
	p.visible = true
	p.currentQuery = query
	p.response = ""
	p.error = ""
	p.copyStatus = ""
	p.loading = false
	p.input.SetValue("")
	p.input.Focus()
}

// Hide closes the AI assistant panel
func (p *AIAssistantPanel) Hide() {
	p.visible = false
	p.input.Blur()
}

// IsVisible returns whether the panel is currently shown
func (p *AIAssistantPanel) IsVisible() bool {
	return p.visible
}

// SetSize updates the panel dimensions
func (p *AIAssistantPanel) SetSize(width, height int) {
	p.width = width
	p.height = height

	// Update viewport size (leave room for borders and input)
	vpWidth := width - 10
	vpHeight := height - 15
	if vpWidth < 20 {
		vpWidth = 20
	}
	if vpHeight < 5 {
		vpHeight = 5
	}
	p.viewport.Width = vpWidth
	p.viewport.Height = vpHeight
}

// SetProvider updates the AI provider
func (p *AIAssistantPanel) SetProvider(provider ai.CLIProvider) {
	p.provider = provider
}

// SetConnection updates the database connection
func (p *AIAssistantPanel) SetConnection(conn db.Connection) {
	p.conn = conn
}

// Update handles input events
func (p *AIAssistantPanel) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle keys when visible
		if !p.visible {
			return nil
		}

		switch msg.String() {
		case "esc":
			p.Hide()
			return nil
		case "enter":
			// Submit the request
			if !p.loading && p.input.Value() != "" {
				task := p.input.Value()
				p.loading = true
				p.error = ""
				return p.invokeAI(task)
			}
		case "ctrl+c":
			// Copy response to clipboard
			if p.response != "" {
				err := clipboard.WriteAll(p.response)
				if err != nil {
					p.copyStatus = "⚠ Copy failed: clipboard not available"
				} else {
					p.copyStatus = "✓ Copied to clipboard"
				}
			} else {
				p.copyStatus = "⚠ No response to copy"
			}
			return nil
		case "up", "down", "pgup", "pgdown":
			// Scroll viewport
			p.viewport, cmd = p.viewport.Update(msg)
			return cmd
		}

		// Pass to input if not loading
		if !p.loading {
			p.input, cmd = p.input.Update(msg)
		}

	case AIResponseMsg:
		p.loading = false
		if msg.Err != nil {
			p.error = msg.Err.Error()
		} else {
			p.response = msg.Response
			p.viewport.SetContent(p.response)
		}
		return nil
	}

	return cmd
}

// invokeAI calls the AI CLI tool
func (p *AIAssistantPanel) invokeAI(task string) tea.Cmd {
	return func() tea.Msg {
		// Check if provider is available
		if p.provider == nil {
			return AIResponseMsg{
				Err: fmt.Errorf("no AI CLI provider available. Please install one of: copilot-cli, claude-cli, sgpt, mods, llm"),
			}
		}

		// Check if connection is available
		if p.conn == nil || p.conn.Status() != db.StatusConnected {
			return AIResponseMsg{
				Err: fmt.Errorf("no active database connection"),
			}
		}

		// Build schema context
		ctx := context.Background()
		schemaCtx, err := ai.BuildSchemaContext(ctx, p.conn, 50, true)
		if err != nil {
			return AIResponseMsg{
				Err: fmt.Errorf("failed to build schema context: %w", err),
			}
		}

		// Invoke AI provider
		response, err := ai.InvokeCLI(ctx, p.provider, schemaCtx, p.currentQuery, task)
		if err != nil {
			return AIResponseMsg{
				Err: err,
			}
		}

		return AIResponseMsg{
			Response: response,
		}
	}
}

// View renders the AI assistant panel
func (p *AIAssistantPanel) View() string {
	if !p.visible {
		return ""
	}

	// Styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212"))

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	loadingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("3"))

	// Provider info
	providerName := "none"
	if p.provider != nil {
		providerName = p.provider.Name()
	}

	// Build content
	var content strings.Builder

	content.WriteString(titleStyle.Render("🤖 AI Assistant"))
	content.WriteString(labelStyle.Render(fmt.Sprintf(" (%s)", providerName)))
	content.WriteString("\n\n")

	// Input field
	content.WriteString(labelStyle.Render("Ask: "))
	content.WriteString(p.input.View())
	content.WriteString("\n\n")

	// Show loading indicator
	if p.loading {
		content.WriteString(loadingStyle.Render("⏳ Thinking..."))
		content.WriteString("\n")
	}

	// Show error if any
	if p.error != "" {
		content.WriteString(errorStyle.Render("Error: "))
		content.WriteString(p.error)
		content.WriteString("\n")
	}

	// Show response
	if p.response != "" {
		content.WriteString(labelStyle.Render("Response:"))
		content.WriteString("\n\n")
		content.WriteString(p.viewport.View())
		content.WriteString("\n")
	}

	// Show copy status if present
	if p.copyStatus != "" {
		copyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46")) // Green for success
		if strings.HasPrefix(p.copyStatus, "⚠") {
			copyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("220")) // Yellow for warning
		}
		content.WriteString(copyStyle.Render(p.copyStatus))
		content.WriteString("\n")
	}

	// Help text
	if p.loading {
		content.WriteString(labelStyle.Render("[Esc] Cancel"))
	} else if p.response != "" {
		content.WriteString(labelStyle.Render("[↑/↓] Scroll  [Esc] Close  [Ctrl+C] Copy"))
	} else {
		content.WriteString(labelStyle.Render("[Enter] Submit  [Esc] Close"))
	}

	return borderStyle.Render(content.String())
}

// AIResponseMsg is sent when AI response is ready
type AIResponseMsg struct {
	Response string
	Err      error
}
