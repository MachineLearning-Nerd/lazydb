package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/MachineLearning-Nerd/lazydb/internal/db"
	"github.com/MachineLearning-Nerd/lazydb/internal/editor"
	"github.com/MachineLearning-Nerd/lazydb/internal/storage"
	"github.com/MachineLearning-Nerd/lazydb/internal/ui/components"
)

// handleKeyMsg routes keyboard input to appropriate handlers.
// NOTE: This is a pointer receiver called from Update() which has a value receiver.
// Go takes &m of the Update() local copy. Mutations work because bubbletea uses
// the returned tea.Model, not the original. All code paths must return m.
func (m *model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// KEY SINK: Editor in INSERT mode captures all keys except whitelisted globals.
	// This prevents resize, layout, navigation, and other keybindings from
	// intercepting characters the user is trying to type in the query editor.
	if m.focusedPanel == FocusEditor && m.editorPanel.IsInInsertMode() && !m.layoutPresetMode {
		switch key {
		case m.config.Keybindings.Global.Quit, "ctrl+c":
			return m.handleQuit()
		case m.config.Keybindings.Global.ExecuteQuery: // ctrl+r
			return m, m.executeQuery()
		case m.config.Keybindings.Global.SaveQuery: // f2
			return m, m.saveQuery()
		case m.config.Keybindings.Global.OpenNeovim: // ctrl+e
			return m.handleOpenNeovim()
		case m.config.Keybindings.Global.AIAssistant: // ctrl+a
			return m.handleAIAssistant()
		default:
			// All other keys go to editor textarea for typing
			cmd := m.editorPanel.Update(msg)
			return m, cmd
		}
	}

	// HIGHEST PRIORITY: Panel resizing (must be before dialog/panel checks)
	increment := m.config.UI.ResizeIncrement

	switch key {
	case m.config.Keybindings.Resize.ShrinkEditorLeft:
		// Shrink editor left (grow connections)
		m.resizePanel("connections", increment)
		m.statusMessage = m.layoutStatus("")
		return m, nil
	case m.config.Keybindings.Resize.GrowEditorLeft:
		// Grow editor left (shrink connections)
		m.resizePanel("editor-left", increment)
		m.statusMessage = m.layoutStatus("")
		return m, nil
	case m.config.Keybindings.Resize.ShrinkEditorRight:
		// Shrink editor right (grow results)
		m.resizePanel("results", increment)
		m.statusMessage = m.layoutStatus("")
		return m, nil
	case m.config.Keybindings.Resize.GrowEditorRight:
		// Grow editor right (shrink results)
		m.resizePanel("editor-right", increment)
		m.statusMessage = m.layoutStatus("")
		return m, nil
	}

	// Layout preset mode handling
	// INSERT mode guard removed -- key sink above already routes L to editor in INSERT mode
	if key == m.config.Keybindings.Layout.PresetMode {
		m.layoutPresetMode = true
		m.statusMessage = "Layout preset: [1] Default  [2] Editor  [3] Results  [4] Balanced"
		return m, nil
	}

	// Check if in layout preset mode
	if m.layoutPresetMode {
		switch key {
		case m.config.Keybindings.Layout.PresetDefault:
			m.setLayoutPreset(1)
			m.statusMessage = m.layoutStatus("Default")
			m.layoutPresetMode = false
			return m, nil
		case m.config.Keybindings.Layout.PresetEditor:
			m.setLayoutPreset(2)
			m.statusMessage = m.layoutStatus("Editor Focus")
			m.layoutPresetMode = false
			return m, nil
		case m.config.Keybindings.Layout.PresetResults:
			m.setLayoutPreset(3)
			m.statusMessage = m.layoutStatus("Results Focus")
			m.layoutPresetMode = false
			return m, nil
		case m.config.Keybindings.Layout.PresetBalanced:
			m.setLayoutPreset(4)
			m.statusMessage = m.layoutStatus("Balanced")
			m.layoutPresetMode = false
			return m, nil
		case "esc":
			m.layoutPresetMode = false
			m.statusMessage = ""
			return m, nil
		}
	}

	// Global keybindings
	switch key {
	case m.config.Keybindings.Global.Quit, "ctrl+c":
		return m.handleQuit()
	case m.config.Keybindings.Global.Help:
		// Open help dialog
		m.helpDialog = components.NewHelpDialog()
		m.showDialog = components.DialogTypeHelp
		return m, nil
	case m.config.Keybindings.Global.ExecuteQuery:
		return m, m.executeQuery()
	case m.config.Keybindings.Global.SaveQuery:
		return m, m.saveQuery()
	case m.config.Keybindings.Global.OpenNeovim:
		return m.handleOpenNeovim()
	case m.config.Keybindings.Global.AIAssistant:
		return m.handleAIAssistant()
	// TODO: Implement HistoryPanel
	// case "h":
	// 	// Toggle history panel visibility
	// 	m.historyVisible = !m.historyVisible
	// 	if m.historyVisible {
	// 		m.historyPanel.SetDatabaseFromConnection()
	// 		m.statusMessage = "History panel opened"
	// 	} else {
	// 		m.statusMessage = "History panel closed"
	// 	}
	}

	// Navigation keybindings
	// INSERT mode guard removed -- key sink above already routes all keys to editor in INSERT mode
	switch key {
	case m.config.Keybindings.Navigation.NextPanel:
		return m.handleNextPanel()
	case m.config.Keybindings.Navigation.PrevPanel:
		return m.handlePrevPanel()
	case m.config.Keybindings.Navigation.FocusConnections:
		if m.focusedPanel == FocusEditor {
			m.editorPanel.Blur()
		}
		m.focusedPanel = FocusConnections
		return m, nil
	case m.config.Keybindings.Navigation.FocusEditor:
		if m.focusedPanel != FocusEditor {
			m.focusedPanel = FocusEditor
			return m, m.editorPanel.Focus()
		}
		return m, nil
	case m.config.Keybindings.Navigation.FocusResults:
		if m.focusedPanel == FocusEditor {
			m.editorPanel.Blur()
		}
		m.focusedPanel = FocusResults
		return m, nil
		// TODO: Implement HistoryPanel
		// case "4":
		// 	// Panel focus to history (only if visible)
		// 	if m.historyVisible {
		// 		if m.focusedPanel == FocusEditor {
		// 			m.editorPanel.Blur()
		// 		}
		// 		m.focusedPanel = FocusHistory
		// 	}
	}

	// Connections panel keybindings (only when connections panel is focused)
	// BUT skip if in schema search mode (let panel handle search input)
	if m.focusedPanel == FocusConnections && !m.connectionsPanel.IsInSchemaSearchMode() {
		switch key {
		case m.config.Keybindings.Connections.Add:
			// Add new connection
			m.connectionForm = components.NewConnectionFormDialog(components.DialogTypeAdd, nil)
			m.showDialog = components.DialogTypeAdd
			return m, nil
		case m.config.Keybindings.Connections.Edit:
			// Edit selected connection
			selectedConn := m.connectionsPanel.GetSelectedConnection()
			if selectedConn != "" {
				conn, err := m.connManager.GetConnection(selectedConn)
				if err == nil {
					connConfig := conn.Config()
					m.connectionForm = components.NewConnectionFormDialog(components.DialogTypeEdit, &connConfig)
					m.showDialog = components.DialogTypeEdit
				}
			}
			return m, nil
		case m.config.Keybindings.Connections.Delete:
			// Delete selected connection
			selectedConn := m.connectionsPanel.GetSelectedConnection()
			if selectedConn != "" {
				m.deleteTarget = selectedConn
				m.confirmDialog = components.NewConfirmationDialog(
					fmt.Sprintf("Delete connection '%s'?", selectedConn),
				)
				m.showDialog = components.DialogTypeDelete
			}
			return m, nil
		case m.config.Keybindings.Connections.Connect:
			// Connect to selected database
			selectedConn := m.connectionsPanel.GetSelectedConnection()
			if selectedConn != "" {
				// Set as active and connect
				m.connManager.SetActive(selectedConn)

				// Save active connection to file for MCP server
				allConns := m.connManager.ListConnections()
				configs := make([]db.ConnectionConfig, 0, len(allConns))
				for _, connName := range allConns {
					if conn, err := m.connManager.GetConnection(connName); err == nil {
						configs = append(configs, conn.Config())
					}
				}
				if err := storage.SaveConnections(configs, selectedConn); err != nil {
					m.logDebug("[WARN] Failed to save active connection: %v", err)
				}

				return m, m.connectToDatabase()
			}
			return m, nil
		}
	}

	// Panel delegation: unmatched keys go to the focused panel
	var cmd tea.Cmd
	switch m.focusedPanel {
	case FocusConnections:
		cmd = m.connectionsPanel.Update(msg)
	case FocusEditor:
		cmd = m.editorPanel.Update(msg)
	case FocusResults:
		cmd = m.resultsPanel.Update(msg)
	// TODO: Implement HistoryPanel
	// case FocusHistory:
	// 	cmd = m.historyPanel.Update(msg)
	}
	return m, cmd
}

// --- Helper methods ---

// handleQuit saves connections and returns tea.Quit.
func (m *model) handleQuit() (tea.Model, tea.Cmd) {
	m.saveConnections()
	return m, tea.Quit
}

// handleOpenNeovim checks nvim availability, gets query + connection, and opens nvim.
func (m *model) handleOpenNeovim() (tea.Model, tea.Cmd) {
	if editor.IsNvimAvailable() {
		query := m.editorPanel.GetQuery()
		activeConn, err := m.connManager.GetActive()
		var conn db.Connection
		if err == nil {
			conn = activeConn
		}
		injectSchema := true
		if m.config.AI != nil {
			injectSchema = m.config.AI.InjectInNeovim
		}
		return m, editor.OpenInNeovimCmd(query, conn, injectSchema)
	}
	m.statusMessage = "Neovim not found. Please install nvim."
	return m, nil
}

// handleAIAssistant checks AI enabled state and shows the assistant with the current query.
func (m *model) handleAIAssistant() (tea.Model, tea.Cmd) {
	if m.config.AI != nil && m.config.AI.Enabled {
		query := m.editorPanel.GetQuery()
		m.aiAssistant.Show(query)
	} else {
		m.statusMessage = "AI Assistant is disabled. Enable in config.yml"
	}
	return m, nil
}

// handleNextPanel blurs the editor if focused and cycles focus forward.
func (m *model) handleNextPanel() (tea.Model, tea.Cmd) {
	if m.focusedPanel == FocusEditor {
		m.editorPanel.Blur()
	}
	// Cycle through panels (3 panels only - history panel not implemented)
	maxPanel := 3
	// TODO: Implement HistoryPanel - would make this 4
	// if m.historyVisible {
	// 	maxPanel = 4
	// }
	m.focusedPanel = PanelFocus((int(m.focusedPanel) + 1) % maxPanel)
	var cmd tea.Cmd
	if m.focusedPanel == FocusEditor {
		cmd = m.editorPanel.Focus()
	}
	return m, cmd
}

// handlePrevPanel blurs the editor if focused and cycles focus backward.
func (m *model) handlePrevPanel() (tea.Model, tea.Cmd) {
	if m.focusedPanel == FocusEditor {
		m.editorPanel.Blur()
	}
	// Cycle backwards through panels (3 panels only - history panel not implemented)
	maxPanel := 3
	// TODO: Implement HistoryPanel - would make this 4
	// if m.historyVisible {
	// 	maxPanel = 4
	// }
	m.focusedPanel = PanelFocus((int(m.focusedPanel) + maxPanel - 1) % maxPanel)
	var cmd tea.Cmd
	if m.focusedPanel == FocusEditor {
		cmd = m.editorPanel.Focus()
	}
	return m, cmd
}

// layoutStatus formats a status string showing the current panel layout ratios.
// If label is non-empty, it is appended in parentheses (e.g. "Default").
func (m *model) layoutStatus(label string) string {
	base := fmt.Sprintf("Layout: %d%% | %d%% | %d%%", m.connectionRatio, m.editorRatio, m.resultsRatio)
	if label != "" {
		return base + " (" + label + ")"
	}
	return base
}

// resizePanel adjusts panel ratios with constraints.
func (m *model) resizePanel(panel string, delta int) {
	minRatio := m.config.UI.MinPanelWidth
	maxRatio := m.config.UI.MaxPanelWidth

	switch panel {
	case "connections":
		newConn := m.connectionRatio + delta
		newEditor := m.editorRatio - delta
		if newConn < minRatio || newConn > maxRatio {
			return
		}
		if newEditor < minRatio || newEditor > maxRatio {
			return
		}
		m.connectionRatio = newConn
		m.editorRatio = newEditor

	case "editor-left":
		newConn := m.connectionRatio - delta
		newEditor := m.editorRatio + delta
		if newConn < minRatio || newConn > maxRatio {
			return
		}
		if newEditor < minRatio || newEditor > maxRatio {
			return
		}
		m.connectionRatio = newConn
		m.editorRatio = newEditor

	case "editor-right":
		newEditor := m.editorRatio + delta
		newResults := m.resultsRatio - delta
		if newEditor < minRatio || newEditor > maxRatio {
			return
		}
		if newResults < minRatio || newResults > maxRatio {
			return
		}
		m.editorRatio = newEditor
		m.resultsRatio = newResults

	case "results":
		newEditor := m.editorRatio - delta
		newResults := m.resultsRatio + delta
		if newEditor < minRatio || newEditor > maxRatio {
			return
		}
		if newResults < minRatio || newResults > maxRatio {
			return
		}
		m.editorRatio = newEditor
		m.resultsRatio = newResults
	}

	m.updatePanelSizes()
}

// setLayoutPreset sets a predefined layout.
func (m *model) setLayoutPreset(preset int) {
	switch preset {
	case 1: // Default
		m.connectionRatio = 20
		m.editorRatio = 40
		m.resultsRatio = 40
	case 2: // Editor focus
		m.connectionRatio = 15
		m.editorRatio = 55
		m.resultsRatio = 30
	case 3: // Results focus
		m.connectionRatio = 15
		m.editorRatio = 30
		m.resultsRatio = 55
	case 4: // Balanced
		m.connectionRatio = 25
		m.editorRatio = 50
		m.resultsRatio = 25
	}
	m.updatePanelSizes()
}
