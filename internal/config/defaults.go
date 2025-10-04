package config

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Version:     1,
		Keybindings: DefaultKeybindings(),
		UI:          DefaultUIConfig(),
		Theme:       DefaultThemeConfig(),
	}
}

// DefaultKeybindings returns the default keybinding configuration
func DefaultKeybindings() KeybindingsConfig {
	return KeybindingsConfig{
		Resize: ResizeKeybindings{
			GrowEditorLeft:    "=",
			ShrinkEditorLeft:  "-",
			GrowEditorRight:   "[",
			ShrinkEditorRight: "]",
		},
		Layout: LayoutKeybindings{
			PresetMode:     "L",
			PresetDefault:  "1",
			PresetEditor:   "2",
			PresetResults:  "3",
			PresetBalanced: "4",
		},
		Navigation: NavigationKeybindings{
			FocusConnections: "1",
			FocusEditor:      "2",
			FocusResults:     "3",
			NextPanel:        "tab",
			PrevPanel:        "shift+tab",
		},
		Global: GlobalKeybindings{
			Help:         "?",
			Quit:         "ctrl+q",
			ExecuteQuery: "ctrl+r",
			SaveQuery:    "f2",
			OpenNeovim:   "ctrl+e",
		},
		Connections: ConnectionsKeybindings{
			Add:            "a",
			Edit:           "e",
			Delete:         "d",
			Connect:        "enter",
			SchemaExplorer: "s",
		},
		Schema: SchemaKeybindings{
			NavigateDown: "j",
			NavigateUp:   "k",
			Expand:       "enter",
			Preview:      "p",
			Search:       "/",
			Refresh:      "r",
			Exit:         "esc",
		},
	}
}

// DefaultUIConfig returns the default UI configuration
func DefaultUIConfig() UIConfig {
	return UIConfig{
		DefaultLayout: PanelRatios{
			Connections: 20,
			Editor:      40,
			Results:     40,
		},
		ResizeIncrement: 5,
		MinPanelWidth:   15,
		MaxPanelWidth:   70,
	}
}

// DefaultThemeConfig returns the default theme configuration
func DefaultThemeConfig() ThemeConfig {
	return ThemeConfig{
		Name:               "monokai",
		SyntaxHighlighting: true,
		SQLLinting:         true,
	}
}
