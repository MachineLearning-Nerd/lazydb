package panels

import (
	"context"
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/glamour"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/MachineLearning-Nerd/lazydb/internal/ai"
	"github.com/MachineLearning-Nerd/lazydb/internal/db"
)

// AIAssistantPanel represents the AI assistant dialog
type AIAssistantPanel struct {
	visible         bool
	input           textinput.Model
	response        string
	sections        []ai.ResponseSection
	selectedSection int // Current selected section (0-indexed)
	loading         bool
	error           string
	copyStatus      string // Feedback for copy operation
	provider        ai.CLIProvider
	conn            db.Connection
	currentQuery    string
	viewport        viewport.Model
	width           int
	height          int
	renderer        *glamour.TermRenderer
	useMCP          bool // If true, use MCP tools instead of injecting schema
}

// NewAIAssistantPanel creates a new AI assistant panel
func NewAIAssistantPanel(provider ai.CLIProvider, conn db.Connection, useMCP bool) *AIAssistantPanel {
	ti := textinput.New()
	ti.Placeholder = "e.g., optimize this query for large datasets"
	ti.CharLimit = 200
	ti.Width = 60

	vp := viewport.New(60, 10)

	// Initialize glamour renderer for markdown
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(60),
	)

	return &AIAssistantPanel{
		visible:         false,
		input:           ti,
		provider:        provider,
		conn:            conn,
		viewport:        vp,
		renderer:        renderer,
		sections:        []ai.ResponseSection{},
		selectedSection: 0,
		useMCP:          useMCP,
	}
}

// Show displays the AI assistant panel
func (p *AIAssistantPanel) Show(query string) {
	p.visible = true
	p.currentQuery = query
	p.response = ""
	p.sections = []ai.ResponseSection{}
	p.selectedSection = 0
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
			// Copy entire response to clipboard
			if p.response != "" {
				err := clipboard.WriteAll(p.response)
				if err != nil {
					p.copyStatus = "⚠ Copy failed: clipboard not available"
				} else {
					p.copyStatus = "✓ Copied entire response to clipboard"
				}
			} else {
				p.copyStatus = "⚠ No response to copy"
			}
			p.copyStatus += " "
			return nil
		case "c", "n", "p":
			// Handle section navigation only when viewing response
			// Otherwise, pass to input so user can type these letters
			if p.response != "" && len(p.sections) > 0 {
				switch msg.String() {
				case "c":
					// Copy current section only
					section := p.sections[p.selectedSection]
					err := clipboard.WriteAll(section.Content)
					if err != nil {
						p.copyStatus = "⚠ Copy failed: clipboard not available"
					} else {
						p.copyStatus = fmt.Sprintf("✓ Copied section %d (%s)", section.Number, ai.FormatSectionTitle(section))
					}
				case "n":
					// Next section
					if p.selectedSection < len(p.sections)-1 {
						p.selectedSection++
						p.updateViewport()
						p.copyStatus = ""
					}
				case "p":
					// Previous section
					if p.selectedSection > 0 {
						p.selectedSection--
						p.updateViewport()
						p.copyStatus = ""
					}
				}
				return nil
			} else {
				// No response yet - pass to input so user can type
				if !p.loading {
					p.input, cmd = p.input.Update(msg)
				}
				return cmd
			}
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
			// Parse response into sections
			p.sections = ai.ParseResponse(msg.Response)
			p.selectedSection = 0
			// Update viewport with formatted response
			p.updateViewport()
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

		// Build schema context (or skip if using MCP)
		ctx := context.Background()
		var schemaCtx *ai.SchemaContext

		if p.useMCP {
			// MCP Mode: Create minimal schema context with UseMCP flag
			schemaCtx = &ai.SchemaContext{
				UseMCP: true,
			}
		} else {
			// Legacy Mode: Build full schema context
			var err error
			schemaCtx, err = ai.BuildSchemaContext(ctx, p.conn, 50, true)
			if err != nil {
				return AIResponseMsg{
					Err: fmt.Errorf("failed to build schema context: %w", err),
				}
			}
			schemaCtx.UseMCP = false
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

// updateViewport updates the viewport content with formatted sections
func (p *AIAssistantPanel) updateViewport() {
	if len(p.sections) == 0 {
		p.viewport.SetContent(p.response)
		return
	}

	var content strings.Builder

	for i, section := range p.sections {
		isSelected := (i == p.selectedSection)

		// Section header with indicator
		sectionTitle := ai.FormatSectionTitle(section)
		indicator := "□"
		if isSelected {
			indicator = "▶"
		}

		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
		if isSelected {
			headerStyle = headerStyle.Foreground(lipgloss.Color("46")) // Green for selected
		}

		header := fmt.Sprintf("%s Section %d: %s", indicator, section.Number, sectionTitle)
		content.WriteString(headerStyle.Render(header))
		content.WriteString("\n")

		// Render section content with glamour for code/markdown
		var sectionContent string
		if section.Type == "code" || section.Type == "query" {
			// Render code block with syntax highlighting
			codeBlock := fmt.Sprintf("```sql\n%s\n```", section.Content)
			if p.renderer != nil {
				rendered, err := p.renderer.Render(codeBlock)
				if err == nil {
					sectionContent = rendered
				} else {
					sectionContent = section.Content
				}
			} else {
				sectionContent = section.Content
			}
		} else {
			// Render as markdown
			if p.renderer != nil {
				rendered, err := p.renderer.Render(section.Content)
				if err == nil {
					sectionContent = rendered
				} else {
					sectionContent = section.Content
				}
			} else {
				sectionContent = section.Content
			}
		}

		content.WriteString(sectionContent)
		content.WriteString("\n")

		// Add spacing between sections
		if i < len(p.sections)-1 {
			content.WriteString("\n")
		}
	}

	p.viewport.SetContent(content.String())
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
	} else if p.response != "" && len(p.sections) > 0 {
		// Show section navigation help
		content.WriteString(labelStyle.Render("[n/p] Sections  [c] Copy §  [Ctrl+C] Copy All  [↑/↓] Scroll  [Esc] Close"))
	} else if p.response != "" {
		// No sections, show basic help
		content.WriteString(labelStyle.Render("[↑/↓] Scroll  [Ctrl+C] Copy  [Esc] Close"))
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
