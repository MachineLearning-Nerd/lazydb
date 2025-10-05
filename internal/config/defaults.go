package config

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Version:     1,
		Keybindings: DefaultKeybindings(),
		UI:          DefaultUIConfig(),
		Theme:       DefaultThemeConfig(),
		AI:          DefaultAIConfig(),
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
			AIAssistant:  "ctrl+a",
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

// DefaultAIConfig returns the default AI configuration
func DefaultAIConfig() *AIConfig {
	return &AIConfig{
		Enabled:          true,
		CLITool:          "claude-cli", // Default to claude-cli, will auto-detect if not available
		InjectInNeovim:   true,
		MaxTables:        50,
		IncludeRowCounts: false,
		IncludeIndexes:   true,
		ContextFormat:    "comments",

		// MCP Server defaults
		MCPEnabled:      true,
		MCPSmartTools:   true,
		MCPCacheEnabled: true,
		MCPMaxCacheSize: 104857600, // 100MB
		MCPAIProvider:   "claude",
	}
}
