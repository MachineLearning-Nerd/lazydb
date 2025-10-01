package ui

import "github.com/charmbracelet/lipgloss"

// Theme contains color definitions for the UI
type Theme struct {
	PrimaryColor   lipgloss.Color
	SecondaryColor lipgloss.Color
	AccentColor    lipgloss.Color
	BorderColor    lipgloss.Color
	ActiveBorder   lipgloss.Color
	TitleColor     lipgloss.Color
	HelpColor      lipgloss.Color
}

// DefaultTheme returns the default color theme
func DefaultTheme() Theme {
	return Theme{
		PrimaryColor:   lipgloss.Color("86"),  // Cyan
		SecondaryColor: lipgloss.Color("212"), // Pink
		AccentColor:    lipgloss.Color("220"), // Yellow
		BorderColor:    lipgloss.Color("238"), // Gray
		ActiveBorder:   lipgloss.Color("86"),  // Cyan
		TitleColor:     lipgloss.Color("86"),  // Cyan
		HelpColor:      lipgloss.Color("241"), // Dark Gray
	}
}

// Styles contains all the lipgloss styles used in the application
type Styles struct {
	Panel        lipgloss.Style
	ActivePanel  lipgloss.Style
	PanelTitle   lipgloss.Style
	HelpText     lipgloss.Style
	StatusBar    lipgloss.Style
}

// NewStyles creates a new Styles instance with the given theme
func NewStyles(theme Theme) Styles {
	return Styles{
		Panel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.BorderColor).
			Padding(1, 2),
		ActivePanel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.ActiveBorder).
			Padding(1, 2),
		PanelTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.TitleColor),
		HelpText: lipgloss.NewStyle().
			Foreground(theme.HelpColor),
		StatusBar: lipgloss.NewStyle().
			Foreground(theme.HelpColor).
			Padding(0, 1),
	}
}
