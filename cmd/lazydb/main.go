package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/MachineLearning-Nerd/lazydb/internal/ai"
	"github.com/MachineLearning-Nerd/lazydb/internal/config"
	"github.com/MachineLearning-Nerd/lazydb/internal/db"
	"github.com/MachineLearning-Nerd/lazydb/internal/editor"
	"github.com/MachineLearning-Nerd/lazydb/internal/storage"
	"github.com/MachineLearning-Nerd/lazydb/internal/ui"
	"github.com/MachineLearning-Nerd/lazydb/internal/ui/components"
	"github.com/MachineLearning-Nerd/lazydb/internal/ui/panels"
)

// PanelFocus represents which panel currently has focus
type PanelFocus int

const (
	FocusConnections PanelFocus = iota
	FocusEditor
	FocusResults
	// FocusHistory // TODO: Implement HistoryPanel
)

// Message types for async operations
type connectionResultMsg struct {
	err error
}

type queryResultMsg struct {
	result db.QueryResult
}

type saveResultMsg struct {
	success  bool
	filepath string
	err      error
}

type yankRowMsg struct {
	rowIndex int
	success  bool
}

type historyClipboardCopyMsg struct {
	success bool
	query   string
	err     error
}

type historyLoadQueryInEditorMsg struct {
	query string
}


// model represents the application state
type model struct {
	width             int
	height            int
	focusedPanel      PanelFocus
	connectionsPanel  *panels.ConnectionsPanel
	editorPanel       *panels.EditorPanel
	resultsPanel      *panels.ResultsPanel
	// historyPanel      *panels.HistoryPanel // TODO: Implement HistoryPanel
	aiAssistant       *panels.AIAssistantPanel
	connManager       *db.ConnectionManager
	config            *config.Config
	theme             ui.Theme
	styles            ui.Styles
	statusMessage     string
	showDialog        components.DialogType
	connectionForm    *components.ConnectionFormDialog
	confirmDialog     *components.ConfirmationDialog
	helpDialog        *components.HelpDialog
	deleteTarget      string // Connection name to delete (for confirmation)
	debugLog          *os.File // Debug log file
	// historyVisible    bool   // Whether history panel is currently visible // TODO: Implement HistoryPanel
	// Panel ratios (must sum to 100 when history is hidden, 100% when visible)
	connectionRatio   int    // Default: 20%
	editorRatio       int    // Default: 40%
	resultsRatio      int    // Default: 40%
	// historyRatio      int    // Default: 20% // TODO: Implement HistoryPanel
	layoutPresetMode  bool   // True when waiting for preset number after Shift+L
}

// logDebug writes a debug message to the log file
func (m *model) logDebug(format string, args ...interface{}) {
	if m.debugLog != nil {
		msg := fmt.Sprintf(format, args...)
		m.debugLog.WriteString(msg + "\n")
		m.debugLog.Sync() // Flush immediately for tail -f
	}
}


// Init initializes the model
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updatePanelSizes()
		m.aiAssistant.SetSize(msg.Width, msg.Height)

	case connectionResultMsg:
		if msg.err != nil {
			m.statusMessage = fmt.Sprintf("Connection failed: %v", msg.err)
		} else {
			activeConn := m.connManager.ActiveName()
			m.statusMessage = fmt.Sprintf("Connected to %s", activeConn)
			// Update history panel with new database context
			// TODO: Implement HistoryPanel
			// if m.historyVisible {
			// 	m.historyPanel.SetDatabaseFromConnection()
			// }
		}

	case queryResultMsg:
		m.resultsPanel.SetResult(msg.result)
		if msg.result.Error != nil {
			m.statusMessage = fmt.Sprintf("Query error: %v", msg.result.Error)
		} else {
			m.statusMessage = fmt.Sprintf("Query executed: %d rows in %dms", msg.result.RowCount, msg.result.ExecutionMs)
		}

	case saveResultMsg:
		if msg.err != nil {
			m.statusMessage = fmt.Sprintf("Save failed: %v", msg.err)
		} else {
			m.statusMessage = fmt.Sprintf("Query saved to %s", msg.filepath)
		}

	case yankRowMsg:
		if msg.success {
			m.statusMessage = fmt.Sprintf("Row %d yanked to clipboard", msg.rowIndex+1)
		} else {
			if msg.rowIndex == -1 {
				m.statusMessage = "No row to yank"
			} else {
				m.statusMessage = fmt.Sprintf("Failed to yank row %d", msg.rowIndex+1)
			}
		}

	case historyClipboardCopyMsg:
		if msg.success {
			m.statusMessage = "Query copied to clipboard"
		} else {
			m.statusMessage = fmt.Sprintf("Failed to copy query: %v", msg.err)
		}

	case historyLoadQueryInEditorMsg:
		m.editorPanel.SetQuery(msg.query)
		m.focusedPanel = FocusEditor
		m.statusMessage = "Query loaded in editor"

	case panels.TablePreviewMsg:
		// Generate preview query: SELECT * FROM schema.table LIMIT 10
		previewQuery := fmt.Sprintf("SELECT * FROM %s.%s LIMIT 10;", msg.Schema, msg.Table)
		m.editorPanel.SetQuery(previewQuery)
		m.statusMessage = fmt.Sprintf("Preview query for %s.%s loaded", msg.Schema, msg.Table)
		m.focusedPanel = FocusEditor

	case editor.NvimErrorMsg:
		m.statusMessage = fmt.Sprintf("Neovim error: %v", msg.Err)

	case editor.NvimSuccessMsg:
		if msg.Text != "" {
			m.editorPanel.SetQuery(msg.Text)
			m.statusMessage = "Query updated from Neovim"
		}

	case panels.AIResponseMsg:
		// Route to AI assistant
		return m, m.aiAssistant.Update(msg)

	case tea.KeyMsg:
		// Handle AI Assistant if visible (highest priority)
		if m.aiAssistant.IsVisible() {
			cmd = m.aiAssistant.Update(msg)
			return m, cmd
		}

		// Handle dialog interactions
		if m.showDialog != components.DialogTypeNone {
			return m.handleDialogKeys(msg)
		}

		// Delegate to key dispatch (input.go)
		return m.handleKeyMsg(msg)

	}

	// Update panels based on focus
	if m.focusedPanel == FocusConnections {
		cmd = m.connectionsPanel.Update(msg)
	} else if m.focusedPanel == FocusEditor {
		cmd = m.editorPanel.Update(msg)
	} else if m.focusedPanel == FocusResults {
		cmd = m.resultsPanel.Update(msg)
	}
	// TODO: Implement HistoryPanel
	// else if m.focusedPanel == FocusHistory {
	// 	cmd = m.historyPanel.Update(msg)
	// }

	return m, cmd
}

// handleDialogKeys handles keyboard input when a dialog is open
func (m *model) handleDialogKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.showDialog {
	case components.DialogTypeAdd, components.DialogTypeEdit:
		switch msg.String() {
		case "esc":
			// Cancel dialog
			m.showDialog = components.DialogTypeNone
			m.connectionForm = nil
			return m, nil

		case "enter":
			// Submit form
			config, err := m.connectionForm.GetConfig()
			if err != nil {
				m.statusMessage = fmt.Sprintf("Validation error: %v", err)
				return m, nil
			}

			if m.showDialog == components.DialogTypeAdd {
				// Add new connection
				pgConn := db.NewPostgresConnection(config)
				m.connManager.AddConnection(config.Name, pgConn)
				m.statusMessage = fmt.Sprintf("Connection '%s' added", config.Name)
			} else {
				// Edit existing connection
				// Remove old connection and add updated one
				m.connManager.RemoveConnection(m.connectionForm.Config.Name)
				pgConn := db.NewPostgresConnection(config)
				m.connManager.AddConnection(config.Name, pgConn)
				m.statusMessage = fmt.Sprintf("Connection '%s' updated", config.Name)
			}

			// Save connections to file
			m.saveConnections()

			m.showDialog = components.DialogTypeNone
			m.connectionForm = nil
			return m, nil

		default:
			// Update form input
			var cmd tea.Cmd
			m.connectionForm, cmd = m.connectionForm.Update(msg)
			return m, cmd
		}

	case components.DialogTypeDelete:
		switch msg.String() {
		case "y", "enter":
			// Confirm deletion
			err := m.connManager.RemoveConnection(m.deleteTarget)
			if err != nil {
				m.statusMessage = fmt.Sprintf("Delete failed: %v", err)
			} else {
				m.statusMessage = fmt.Sprintf("Connection '%s' deleted", m.deleteTarget)
				// Save connections to file
				m.saveConnections()
			}
			m.showDialog = components.DialogTypeNone
			m.confirmDialog = nil
			m.deleteTarget = ""
			return m, nil

		case "n", "esc":
			// Cancel deletion
			m.showDialog = components.DialogTypeNone
			m.confirmDialog = nil
			m.deleteTarget = ""
			return m, nil
		}

	case components.DialogTypeHelp:
		switch msg.String() {
		case "esc", "?", "f1":
			// Close help dialog
			m.showDialog = components.DialogTypeNone
			m.helpDialog = nil
			return m, nil

		case "left":
			// Previous category
			m.helpDialog.Navigate("prev_category")
			return m, nil

		case "right":
			// Next category
			m.helpDialog.Navigate("next_category")
			return m, nil

		case "up", "k":
			// Previous query
			m.helpDialog.Navigate("prev_query")
			return m, nil

		case "down", "j":
			// Next query
			m.helpDialog.Navigate("next_query")
			return m, nil

		case "enter":
			// Copy selected query to editor
			query := m.helpDialog.GetSelectedQuery()
			if query != "" {
				m.editorPanel.SetQuery(query)
				m.statusMessage = "Query copied to editor"
			}
			m.showDialog = components.DialogTypeNone
			m.helpDialog = nil
			return m, nil
		}
	}

	return m, nil
}

// saveConnections saves all connections to file
func (m *model) saveConnections() {
	configs := m.connManager.GetAllConfigs()
	activeConn := m.connManager.ActiveName()

	err := storage.SaveConnections(configs, activeConn)
	if err != nil {
		// Log error but don't interrupt user flow
		m.statusMessage = fmt.Sprintf("Warning: Failed to save connections: %v", err)
	}
}

// connectToDatabase attempts to connect to the active database
func (m *model) connectToDatabase() tea.Cmd {
	return func() tea.Msg {
		conn, err := m.connManager.GetActive()
		if err != nil {
			return connectionResultMsg{err: err}
		}

		ctx := context.Background()
		err = conn.Connect(ctx)
		return connectionResultMsg{err: err}
	}
}

// executeQuery executes the current query in the editor
func (m *model) executeQuery() tea.Cmd {
	return func() tea.Msg {
		// Get the query text
		query := m.editorPanel.GetQuery()
		if query == "" {
			return queryResultMsg{
				result: db.QueryResult{
					Error: fmt.Errorf("no query to execute"),
				},
			}
		}

		// Get the active connection
		conn, err := m.connManager.GetActive()
		if err != nil {
			return queryResultMsg{
				result: db.QueryResult{
					Error: fmt.Errorf("not connected to database"),
				},
			}
		}

		// Check if connected
		if conn.Status() != db.StatusConnected {
			return queryResultMsg{
				result: db.QueryResult{
					Error: fmt.Errorf("not connected to database"),
				},
			}
		}

		// Execute the query with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		result, err := conn.ExecuteQuery(ctx, query)
		if err != nil {
			return queryResultMsg{
				result: db.QueryResult{
					Error: err,
				},
			}
		}

		// If query executed successfully, log it to environment's history file
		if result.Error == nil {
			config := conn.Config()
			environment := config.Environment
			if environment == "" {
				environment = db.EnvDevelopment // Default to Development
			}
			// Ignore logging errors to not interrupt user flow
			_ = storage.AppendQueryToHistory(query, environment)
		}

		return queryResultMsg{result: result}
	}
}

// saveQuery saves the current query to a file
func (m *model) saveQuery() tea.Cmd {
	return func() tea.Msg {
		query := m.editorPanel.GetQuery()
		if query == "" {
			return saveResultMsg{
				success: false,
				err:     fmt.Errorf("no query to save"),
			}
		}

		// Save with auto-generated filename
		err := storage.SaveQuery(query, "")
		if err != nil {
			return saveResultMsg{
				success: false,
				err:     err,
			}
		}

		// Get the queries directory to show in message
		queriesDir, _ := storage.GetQueriesDir()
		return saveResultMsg{
			success:  true,
			filepath: queriesDir,
		}
	}
}


// updatePanelSizes calculates and sets the size for each panel
func (m *model) updatePanelSizes() {
	// Reserve space for header and footer
	headerHeight := 3
	footerHeight := 3
	availableHeight := m.height - headerHeight - footerHeight

	if availableHeight < 10 {
		availableHeight = 10
	}

	// Panel widths using dynamic ratios
	totalWidth := m.width
	connectionWidth := (totalWidth * m.connectionRatio / 100)
	editorWidth := (totalWidth * m.editorRatio / 100)
	resultsWidth := (totalWidth * m.resultsRatio / 100)

	// Content height (subtract padding)
	contentHeight := availableHeight - 4 // 4 for border and padding

	// Set sizes with border/padding adjustment
	m.connectionsPanel.SetSize(connectionWidth-6, contentHeight)
	m.editorPanel.SetSize(editorWidth-6, contentHeight)
	m.resultsPanel.SetSize(resultsWidth-6, contentHeight)
}

// View renders the UI
func (m model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	// Render main view
	mainView := m.renderMainView()

	// If AI Assistant is visible, render it on top
	if m.aiAssistant.IsVisible() {
		aiView := m.aiAssistant.View()
		return m.overlayDialog(mainView, aiView)
	}

	// If dialog is open, render it on top
	if m.showDialog != components.DialogTypeNone {
		dialogView := m.renderDialog()
		// Center dialog on screen
		return m.overlayDialog(mainView, dialogView)
	}

	return mainView
}

// renderMainView renders the main 3-panel layout
func (m model) renderMainView() string {

	// Header with connection info
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.PrimaryColor).
		Padding(0, 1)

	connInfoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("6")). // Cyan for connection info
		Padding(0, 1)

	helpStyle := lipgloss.NewStyle().
		Foreground(m.theme.HelpColor).
		Padding(0, 1)

	// Build header parts
	title := titleStyle.Render("LazyDB v1.0.0")

	// Connection info (center)
	connInfo := ""
	if activeConn, err := m.connManager.GetActive(); err == nil {
		config := activeConn.Config()
		status := activeConn.Status()
		if status == db.StatusConnected {
			connInfo = connInfoStyle.Render(fmt.Sprintf("[%s] %s@%s:%d/%s",
				config.Name, config.Username, config.Host, config.Port, config.Database))
		} else {
			connInfo = connInfoStyle.Render(fmt.Sprintf("[%s] %s", config.Name, status.String()))
		}
	}

	help := helpStyle.Render("[?] Help")

	// Calculate spacing to distribute across width
	usedWidth := lipgloss.Width(title) + lipgloss.Width(connInfo) + lipgloss.Width(help)
	leftPad := (m.width - usedWidth) / 2
	if leftPad < 0 {
		leftPad = 1
	}

	header := title +
		strings.Repeat(" ", leftPad) +
		connInfo +
		strings.Repeat(" ", leftPad) +
		help

	// Calculate panel dimensions
	headerHeight := 3
	footerHeight := 3
	availableHeight := m.height - headerHeight - footerHeight
	if availableHeight < 10 {
		availableHeight = 10
	}

	var panels string
	// var historyView string // TODO: Implement HistoryPanel

	// TODO: Implement HistoryPanel - 4-panel layout disabled
	// if m.historyVisible {
	// 	// 4-panel layout
	// 	// Recalculate ratios to include history panel
	// 	totalRatio := m.connectionRatio + m.editorRatio + m.resultsRatio + m.historyRatio
	//
	// 	connectionWidth := (m.width * m.connectionRatio / 100) / totalRatio * 100
	// 	editorWidth := (m.width * m.editorRatio / 100) / totalRatio * 100
	// 	resultsWidth := (m.width * m.resultsRatio / 100) / totalRatio * 100
	// 	historyWidth := (m.width * m.historyRatio / 100) / totalRatio * 100
	//
	// 	// Render each panel with appropriate styling and dimensions
	// 	connStyle := m.styles.Panel.Width(connectionWidth - 6).Height(availableHeight - 4)
	// 	editStyle := m.styles.Panel.Width(editorWidth - 6).Height(availableHeight - 4)
	// 	resStyle := m.styles.Panel.Width(resultsWidth - 6).Height(availableHeight - 4)
	// 	histStyle := m.styles.Panel.Width(historyWidth - 6).Height(availableHeight - 4)
	//
	// 	// Highlight focused panel
	// 	switch m.focusedPanel {
	// 	case FocusConnections:
	// 		connStyle = m.styles.ActivePanel.Width(connectionWidth - 6).Height(availableHeight - 4)
	// 	case FocusEditor:
	// 		editStyle = m.styles.ActivePanel.Width(editorWidth - 6).Height(availableHeight - 4)
	// 	case FocusResults:
	// 		resStyle = m.styles.ActivePanel.Width(resultsWidth - 6).Height(availableHeight - 4)
	// 	case FocusHistory:
	// 		histStyle = m.styles.ActivePanel.Width(historyWidth - 6).Height(availableHeight - 4)
	// 	}
	//
	// 	connectionsView := connStyle.Render(m.connectionsPanel.View())
	// 	editorView := editStyle.Render(m.editorPanel.View())
	// 	resultsView := resStyle.Render(m.resultsPanel.View())
	// 	historyView = histStyle.Render(m.historyPanel.View())
	//
	// 	// Arrange panels horizontally
	// 	panels = lipgloss.JoinHorizontal(
	// 		lipgloss.Top,
	// 		connectionsView,
	// 		editorView,
	// 		resultsView,
	// 		historyView,
	// 	)
	// } else {
	{
		// 3-panel layout (original)
		totalRatio := m.connectionRatio + m.editorRatio + m.resultsRatio

		// Calculate panel widths as percentage of terminal width
		connectionWidth := m.width * m.connectionRatio / totalRatio
		editorWidth := m.width * m.editorRatio / totalRatio
		resultsWidth := m.width * m.resultsRatio / totalRatio

		// Render each panel with appropriate styling and dimensions
		connStyle := m.styles.Panel.Width(connectionWidth - 6).Height(availableHeight - 4)
		editStyle := m.styles.Panel.Width(editorWidth - 6).Height(availableHeight - 4)
		resStyle := m.styles.Panel.Width(resultsWidth - 6).Height(availableHeight - 4)

		// Highlight focused panel
		switch m.focusedPanel {
		case FocusConnections:
			connStyle = m.styles.ActivePanel.Width(connectionWidth - 6).Height(availableHeight - 4)
		case FocusEditor:
			editStyle = m.styles.ActivePanel.Width(editorWidth - 6).Height(availableHeight - 4)
		case FocusResults:
			resStyle = m.styles.ActivePanel.Width(resultsWidth - 6).Height(availableHeight - 4)
		}

		connectionsView := connStyle.Render(m.connectionsPanel.View())
		editorView := editStyle.Render(m.editorPanel.View())
		resultsView := resStyle.Render(m.resultsPanel.View())

		// Arrange panels horizontally
		panels = lipgloss.JoinHorizontal(
			lipgloss.Top,
			connectionsView,
			editorView,
			resultsView,
		)
	}

	// Footer with panel-specific help
	var helpText string
	switch m.focusedPanel {
	case FocusConnections:
		helpText = m.connectionsPanel.Help()
	case FocusEditor:
		helpText = m.editorPanel.Help()
	case FocusResults:
		helpText = m.resultsPanel.Help()
	// TODO: Implement HistoryPanel
	// case FocusHistory:
	// 	helpText = m.historyPanel.Help()
	}

	// Connection status with database info
	statusText := "▸ Not connected"
	if m.statusMessage != "" {
		statusText = "▸ " + m.statusMessage
	} else if activeConn, err := m.connManager.GetActive(); err == nil {
		config := activeConn.Config()
		status := activeConn.Status()
		if status == db.StatusConnected {
			statusText = fmt.Sprintf("▸ Connected: %s @ %s:%d/%s",
				config.Name, config.Host, config.Port, config.Database)
		} else {
			statusText = fmt.Sprintf("▸ %s - %s", config.Name, status.String())
		}
	}

	// Panel focus indicators with highlighting
	focusIndicators := ""
	maxPanels := 3
	// TODO: Implement HistoryPanel - would make this 4
	// if m.historyVisible {
	// 	maxPanels = 4
	// }
	
	for i := 1; i <= maxPanels; i++ {
		panelNum := PanelFocus(i - 1)
		if panelNum == m.focusedPanel {
			// Highlighted focused panel
			focusIndicators += lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("2")). // Green for focused
				Render(fmt.Sprintf("[%d]", i))
		} else {
			// Regular unfocused panel
			focusIndicators += lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")). // Gray for unfocused
				Render(fmt.Sprintf(" %d ", i))
		}
	}

	// Add layout info to footer
	layoutInfo := fmt.Sprintf("%d%%│%d%%│%d%%", m.connectionRatio, m.editorRatio, m.resultsRatio)

	footer := m.styles.StatusBar.Render(
		fmt.Sprintf("%s     %s Focus  [Tab] Next  [Ctrl+←/→] Resize  [Ctrl+Q] Quit  │ %s │ %s", statusText, focusIndicators, layoutInfo, helpText),
	)

	// Combine all parts
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"\n",
		panels,
		"\n",
		footer,
	)
}

// renderDialog renders the appropriate dialog
func (m model) renderDialog() string {
	switch m.showDialog {
	case components.DialogTypeAdd, components.DialogTypeEdit:
		return m.connectionForm.View()
	case components.DialogTypeDelete:
		return m.confirmDialog.View()
	case components.DialogTypeHelp:
		return m.helpDialog.View()
	default:
		return ""
	}
}

// overlayDialog overlays the dialog on top of the main view
func (m model) overlayDialog(mainView, dialogView string) string {
	// Render main view as background
	// Then place dialog in center on top

	// For simplicity, just center the dialog - the TUI will handle the background
	centered := lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		dialogView)

	return centered
}

func main() {
	// Create debug log file
	configDir, _ := config.GetConfigDir()
	debugLogPath := filepath.Join(configDir, "debug.log")
	debugLog, err := os.OpenFile(debugLogPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to create debug log: %v\n", err)
		debugLog = nil
	}
	defer func() {
		if debugLog != nil {
			debugLog.Close()
		}
	}()

	// Log startup
	if debugLog != nil {
		debugLog.WriteString(fmt.Sprintf("[STARTUP] Debug logging initialized at %s\n", debugLogPath))
	}

	// Load configuration
	cfg, loadErr := config.LoadConfig()
	if loadErr != nil {
		msg := fmt.Sprintf("Warning: Failed to load config: %v\n", loadErr)
		fmt.Fprint(os.Stderr, msg)
		if debugLog != nil {
			debugLog.WriteString("[STARTUP] " + msg)
		}
		fmt.Fprintf(os.Stderr, "Using default configuration\n")
		cfg = config.DefaultConfig()
	}

	// DEBUG LOG 2: Show loaded config values
	if debugLog != nil {
		debugLog.WriteString(fmt.Sprintf("[STARTUP] Config loaded - Resize keys: grow='%s' shrink='%s' growR='%s' shrinkR='%s'\n",
			cfg.Keybindings.Resize.GrowEditorLeft,
			cfg.Keybindings.Resize.ShrinkEditorLeft,
			cfg.Keybindings.Resize.GrowEditorRight,
			cfg.Keybindings.Resize.ShrinkEditorRight,
		))
		debugLog.Sync()
	}

	// Initialize theme and styles
	theme := ui.DefaultTheme()
	styles := ui.NewStyles(theme)

	// Create connection manager
	connMgr := db.NewConnectionManager()

	// Load saved connections from file
	savedConfig, loadErr := storage.LoadConnections()
	if loadErr != nil {
		// If loading fails, just start with empty connections
		// User can add connections manually
	} else {
		// Add all saved connections to manager
		for _, connConfig := range savedConfig.Connections {
			pgConn := db.NewPostgresConnection(connConfig)
			connMgr.AddConnection(connConfig.Name, pgConn)
		}

		// Set active connection if one was saved
		if savedConfig.ActiveConnection != "" {
			connMgr.SetActive(savedConfig.ActiveConnection)
		}
	}

	// If no connections were loaded, add default dev-local connection
	if len(connMgr.ListConnections()) == 0 {
		pgConfig := db.ConnectionConfig{
			Name:     "dev-local",
			Host:     "localhost",
			Port:     5432,
			Database: "postgres",
			Username: "postgres",
			Password: "postgres",
			SSLMode:  "disable",
		}
		pgConn := db.NewPostgresConnection(pgConfig)
		connMgr.AddConnection("dev-local", pgConn)
		connMgr.SetActive("dev-local")
	}

	// Initialize AI provider (if enabled)
	var aiProvider ai.CLIProvider
	if cfg.AI != nil && cfg.AI.Enabled {
		aiProvider = ai.DetectAvailableCLI(cfg.AI.CLITool)
		if aiProvider != nil && debugLog != nil {
			debugLog.WriteString(fmt.Sprintf("[STARTUP] AI provider detected: %s\n", aiProvider.Name()))
		}
	}

	// Determine if MCP mode should be used for AI assistant
	useMCP := false
	if cfg.AI != nil && cfg.AI.MCPEnabled {
		useMCP = true
	}

	// Initialize the model
	ctx := context.Background()
	m := model{
		focusedPanel:     FocusEditor,
		connectionsPanel: panels.NewConnectionsPanel(connMgr, ctx),
		editorPanel:      panels.NewEditorPanel(),
		resultsPanel:     panels.NewResultsPanel(),
		// historyPanel:     panels.NewHistoryPanel(connMgr), // TODO: Implement HistoryPanel
		aiAssistant:      panels.NewAIAssistantPanel(aiProvider, connMgr, useMCP),
		connManager:      connMgr,
		config:           cfg,
		debugLog:         debugLog,
		theme:            theme,
		styles:           styles,
		statusMessage:    "",
		showDialog:       components.DialogTypeNone,
		// Initialize panel ratios from config
		connectionRatio:  cfg.UI.DefaultLayout.Connections,
		editorRatio:      cfg.UI.DefaultLayout.Editor,
		resultsRatio:     cfg.UI.DefaultLayout.Results,
		// historyRatio:     20, // Default 20% for history panel // TODO: Implement HistoryPanel
	}

	// Create the program with alternate screen
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
