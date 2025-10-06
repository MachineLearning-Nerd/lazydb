package panels

import (
	"context"
	"fmt"
	"regexp"
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
	connManager     *db.ConnectionManager
	currentQuery    string
	viewport        viewport.Model
	width           int
	height          int
	renderer        *glamour.TermRenderer
	useMCP          bool   // If true, use MCP tools instead of injecting schema
	mode            string // "input" or "response" - vim-like modal system
}

// NewAIAssistantPanel creates a new AI assistant panel
func NewAIAssistantPanel(provider ai.CLIProvider, connMgr *db.ConnectionManager, useMCP bool) *AIAssistantPanel {
	ti := textinput.New()
	ti.Placeholder = "e.g., optimize this query for large datasets"
	ti.CharLimit = 200
	ti.Width = 60

	vp := viewport.New(60, 10)

	// Initialize glamour renderer for markdown
	// Use notty style to avoid ANSI escape codes in Bubble Tea TUI
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle("notty"),
		glamour.WithWordWrap(60),
	)

	return &AIAssistantPanel{
		visible:         false,
		input:           ti,
		provider:        provider,
		connManager:     connMgr,
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
	p.mode = "input" // Start in input mode for typing question
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
			// Hierarchical ESC: Response Mode → Input Mode → Close Dialog
			if p.mode == "response" {
				// Response mode: ESC goes back to input mode for follow-up questions
				p.mode = "input"
				p.input.Focus()
				p.copyStatus = ""
				return nil
			} else {
				// Input mode: ESC closes dialog completely
				p.Hide()
				return nil
			}
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
		case "i":
			// Vim-like 'i' key: switch to input mode for follow-up questions
			if p.mode == "response" {
				p.mode = "input"
				p.input.Focus()
				p.copyStatus = ""
				return nil
			}
		case "c", "n", "p":
			// Modal key handling: behavior depends on mode
			if p.mode == "response" && len(p.sections) > 0 {
				// Response mode: n/p/c work for navigation and copying
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
			} else if p.mode == "input" && !p.loading {
				// Input mode: pass keys to textinput for typing
				p.input, cmd = p.input.Update(msg)
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
			// Switch to response mode - now n/p/c work for navigation
			p.mode = "response"
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

		// Get fresh active connection (same pattern as query editor)
		conn, err := p.connManager.GetActive()
		if err != nil || conn.Status() != db.StatusConnected {
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
			// Legacy Mode: Build full schema context from fresh connection
			var err error
			schemaCtx, err = ai.BuildSchemaContext(ctx, conn, 50, true)
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

// processMarkdown pre-processes markdown to handle bold, italic, and formatting issues
func processMarkdown(content string) string {
	// Define styles for formatted text
	boldStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252"))

	// Replace **text** with styled bold (handles **, ***, ****, etc.)
	// This fixes the issue where **stat_****tests** shows as literal **
	content = regexp.MustCompile(`\*\*+([^*]+)\*\*+`).ReplaceAllStringFunc(content, func(match string) string {
		// Extract text between asterisks
		text := regexp.MustCompile(`\*+`).ReplaceAllString(match, "")
		return boldStyle.Render(text)
	})

	return content
}

// updateViewport updates the viewport content with formatted sections
func (p *AIAssistantPanel) updateViewport() {
	// Recreate renderer with current viewport width for proper wrapping
	// Use notty style to avoid ANSI escape codes in Bubble Tea TUI
	if p.viewport.Width > 0 {
		p.renderer, _ = glamour.NewTermRenderer(
			glamour.WithStandardStyle("notty"),
			glamour.WithWordWrap(p.viewport.Width),
		)
	}

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

		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("51")) // Bright cyan
		if isSelected {
			headerStyle = headerStyle.Foreground(lipgloss.Color("46")) // Bright green for selected
		}

		header := fmt.Sprintf("%s Section %d: %s", indicator, section.Number, sectionTitle)
		content.WriteString(headerStyle.Render(header))
		content.WriteString("\n")

		// Add separator line under header for visual clarity
		separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		separator := strings.Repeat("─", 60)
		content.WriteString(separatorStyle.Render(separator))
		content.WriteString("\n\n") // Extra spacing after separator

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
			// Pre-process markdown to fix bold formatting (** → styled text)
			processedContent := processMarkdown(section.Content)

			// Render as markdown
			if p.renderer != nil {
				rendered, err := p.renderer.Render(processedContent)
				if err == nil {
					sectionContent = rendered
				} else {
					sectionContent = processedContent
				}
			} else {
				sectionContent = processedContent
			}
		}

		// Indent section content slightly for visual hierarchy
		indentedContent := "  " + strings.ReplaceAll(sectionContent, "\n", "\n  ")
		content.WriteString(indentedContent)
		content.WriteString("\n")

		// Add more spacing between sections
		if i < len(p.sections)-1 {
			content.WriteString("\n\n") // Double spacing between sections
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
		Foreground(lipgloss.Color("252")) // Bright white for better contrast

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

	// Database info
	dbName := "no connection"
	if p.connManager != nil {
		if conn, err := p.connManager.GetActive(); err == nil && conn != nil {
			dbName = conn.Config().Database
		}
	}

	// Build content
	var content strings.Builder

	content.WriteString(titleStyle.Render("🤖 AI Assistant"))
	content.WriteString(labelStyle.Render(fmt.Sprintf(" (%s)", providerName)))
	content.WriteString(" ")
	dbStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
	content.WriteString(dbStyle.Render(fmt.Sprintf("[%s]", dbName)))
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

	// Help text - show mode-specific commands
	if p.loading {
		content.WriteString(labelStyle.Render("[Esc] Cancel"))
	} else if p.mode == "response" && len(p.sections) > 0 {
		// Response mode with sections: show navigation and mode switching help
		content.WriteString(labelStyle.Render("RESPONSE MODE  [n/p] Sections  [c] Copy §  [Ctrl+C] Copy All  [i/Esc] New Question"))
	} else if p.mode == "response" {
		// Response mode without sections: show basic help and mode switching
		content.WriteString(labelStyle.Render("RESPONSE MODE  [↑/↓] Scroll  [Ctrl+C] Copy  [i/Esc] New Question"))
	} else {
		// Input mode: show typing help
		content.WriteString(labelStyle.Render("INPUT MODE  [Enter] Submit  [Esc] Close"))
	}

	return borderStyle.Render(content.String())
}

// AIResponseMsg is sent when AI response is ready
type AIResponseMsg struct {
	Response string
	Err      error
}
