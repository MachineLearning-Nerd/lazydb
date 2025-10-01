package components

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/MachineLearning-Nerd/lazydb/internal/db"
)

// DialogType represents the type of dialog
type DialogType int

const (
	DialogTypeNone DialogType = iota
	DialogTypeAdd
	DialogTypeEdit
	DialogTypeDelete
)

// ConnectionFormDialog represents a form for adding/editing connections
type ConnectionFormDialog struct {
	inputs     []textinput.Model
	focusIndex int
	mode       DialogType
	Config     db.ConnectionConfig // For edit mode (exported for access)
	width      int
	height     int
}

// Field indices
const (
	fieldName = iota
	fieldHost
	fieldPort
	fieldDatabase
	fieldUsername
	fieldPassword
	fieldSSLMode
	fieldEnvironment
)

// NewConnectionFormDialog creates a new connection form dialog
func NewConnectionFormDialog(mode DialogType, config *db.ConnectionConfig) *ConnectionFormDialog {
	inputs := make([]textinput.Model, 8)

	// Name
	inputs[fieldName] = textinput.New()
	inputs[fieldName].Placeholder = "Connection name"
	inputs[fieldName].Focus()
	inputs[fieldName].CharLimit = 50
	inputs[fieldName].Width = 40

	// Host
	inputs[fieldHost] = textinput.New()
	inputs[fieldHost].Placeholder = "localhost"
	inputs[fieldHost].CharLimit = 100
	inputs[fieldHost].Width = 40

	// Port
	inputs[fieldPort] = textinput.New()
	inputs[fieldPort].Placeholder = "5432"
	inputs[fieldPort].CharLimit = 5
	inputs[fieldPort].Width = 40

	// Database
	inputs[fieldDatabase] = textinput.New()
	inputs[fieldDatabase].Placeholder = "postgres"
	inputs[fieldDatabase].CharLimit = 50
	inputs[fieldDatabase].Width = 40

	// Username
	inputs[fieldUsername] = textinput.New()
	inputs[fieldUsername].Placeholder = "postgres"
	inputs[fieldUsername].CharLimit = 50
	inputs[fieldUsername].Width = 40

	// Password
	inputs[fieldPassword] = textinput.New()
	inputs[fieldPassword].Placeholder = "password"
	inputs[fieldPassword].EchoMode = textinput.EchoPassword
	inputs[fieldPassword].EchoCharacter = '•'
	inputs[fieldPassword].CharLimit = 100
	inputs[fieldPassword].Width = 40

	// SSL Mode
	inputs[fieldSSLMode] = textinput.New()
	inputs[fieldSSLMode].Placeholder = "disable"
	inputs[fieldSSLMode].CharLimit = 20
	inputs[fieldSSLMode].Width = 40

	// Environment (display-only, cycled with left/right arrows)
	inputs[fieldEnvironment] = textinput.New()
	inputs[fieldEnvironment].Placeholder = "Development"
	inputs[fieldEnvironment].CharLimit = 20
	inputs[fieldEnvironment].Width = 40

	dialog := &ConnectionFormDialog{
		inputs:     inputs,
		focusIndex: 0,
		mode:       mode,
		width:      60,
		height:     22,
	}

	// If editing, pre-fill with existing config
	if mode == DialogTypeEdit && config != nil {
		dialog.Config = *config
		inputs[fieldName].SetValue(config.Name)
		inputs[fieldHost].SetValue(config.Host)
		inputs[fieldPort].SetValue(strconv.Itoa(config.Port))
		inputs[fieldDatabase].SetValue(config.Database)
		inputs[fieldUsername].SetValue(config.Username)
		inputs[fieldPassword].SetValue(config.Password)
		inputs[fieldSSLMode].SetValue(config.SSLMode)
		if config.Environment != "" {
			inputs[fieldEnvironment].SetValue(string(config.Environment))
		} else {
			inputs[fieldEnvironment].SetValue(string(db.EnvDevelopment))
		}
	} else {
		// Default to Development for new connections
		inputs[fieldEnvironment].SetValue(string(db.EnvDevelopment))
	}

	return dialog
}

// Update handles input events
func (d *ConnectionFormDialog) Update(msg tea.Msg) (*ConnectionFormDialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			// Move to next input
			d.inputs[d.focusIndex].Blur()
			d.focusIndex = (d.focusIndex + 1) % len(d.inputs)
			d.inputs[d.focusIndex].Focus()
			return d, nil

		case "shift+tab", "up":
			// Move to previous input
			d.inputs[d.focusIndex].Blur()
			d.focusIndex--
			if d.focusIndex < 0 {
				d.focusIndex = len(d.inputs) - 1
			}
			d.inputs[d.focusIndex].Focus()
			return d, nil

		case "left", "right":
			// Cycle environment when focused on environment field
			if d.focusIndex == fieldEnvironment {
				currentEnv := d.inputs[fieldEnvironment].Value()
				var newEnv string
				switch currentEnv {
				case string(db.EnvDevelopment):
					newEnv = string(db.EnvStaging)
				case string(db.EnvStaging):
					newEnv = string(db.EnvProduction)
				case string(db.EnvProduction):
					newEnv = string(db.EnvDevelopment)
				default:
					newEnv = string(db.EnvDevelopment)
				}
				d.inputs[fieldEnvironment].SetValue(newEnv)
				return d, nil
			}
		}
	}

	// Update focused input (but not for environment field which is read-only)
	if d.focusIndex != fieldEnvironment {
		var cmd tea.Cmd
		d.inputs[d.focusIndex], cmd = d.inputs[d.focusIndex].Update(msg)
		return d, cmd
	}

	return d, nil
}

// View renders the dialog
func (d *ConnectionFormDialog) View() string {
	title := "Add Connection"
	if d.mode == DialogTypeEdit {
		title = "Edit Connection"
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("5")).
		Padding(0, 1)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).
		Width(12)

	content := titleStyle.Render(title) + "\n\n"

	labels := []string{
		"Name:",
		"Host:",
		"Port:",
		"Database:",
		"Username:",
		"Password:",
		"SSL Mode:",
		"Environment:",
	}

	for i, input := range d.inputs {
		label := labelStyle.Render(labels[i]) + " "
		if i == fieldEnvironment {
			// Add hint for environment field
			envValue := input.Value()
			var envIcon string
			switch envValue {
			case string(db.EnvDevelopment):
				envIcon = "🟢"
			case string(db.EnvStaging):
				envIcon = "🔵"
			case string(db.EnvProduction):
				envIcon = "🔴"
			}
			content += label + envIcon + " " + input.View() + " [←/→ to change]\n"
		} else {
			content += label + input.View() + "\n"
		}
	}

	content += "\n[Tab] Next field  [Enter] Save  [Esc] Cancel\n"

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("5")).
		Padding(1, 2)

	return borderStyle.Render(content)
}

// GetConfig returns the connection config from the form
func (d *ConnectionFormDialog) GetConfig() (db.ConnectionConfig, error) {
	port, err := strconv.Atoi(d.inputs[fieldPort].Value())
	if err != nil || port <= 0 || port > 65535 {
		return db.ConnectionConfig{}, fmt.Errorf("invalid port number")
	}

	config := db.ConnectionConfig{
		Name:        d.inputs[fieldName].Value(),
		Host:        d.inputs[fieldHost].Value(),
		Port:        port,
		Database:    d.inputs[fieldDatabase].Value(),
		Username:    d.inputs[fieldUsername].Value(),
		Password:    d.inputs[fieldPassword].Value(),
		SSLMode:     d.inputs[fieldSSLMode].Value(),
		Environment: db.Environment(d.inputs[fieldEnvironment].Value()),
	}

	// Validation
	if config.Name == "" {
		return db.ConnectionConfig{}, fmt.Errorf("name is required")
	}
	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Database == "" {
		return db.ConnectionConfig{}, fmt.Errorf("database is required")
	}
	if config.Username == "" {
		return db.ConnectionConfig{}, fmt.Errorf("username is required")
	}
	if config.SSLMode == "" {
		config.SSLMode = "disable"
	}
	if config.Environment == "" {
		config.Environment = db.EnvDevelopment
	}

	return config, nil
}

// ConfirmationDialog represents a yes/no confirmation dialog
type ConfirmationDialog struct {
	message string
	width   int
	height  int
}

// NewConfirmationDialog creates a new confirmation dialog
func NewConfirmationDialog(message string) *ConfirmationDialog {
	return &ConfirmationDialog{
		message: message,
		width:   50,
		height:  10,
	}
}

// View renders the confirmation dialog
func (d *ConfirmationDialog) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("9")).
		Padding(0, 1)

	content := titleStyle.Render("Confirm") + "\n\n"
	content += d.message + "\n\n"
	content += "[y] Yes  [n/Esc] No\n"

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("9")).
		Padding(1, 2)

	return borderStyle.Render(content)
}
