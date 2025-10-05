package config

// Config represents the entire LazyDB configuration
type Config struct {
	Version     int              `yaml:"version"`
	Keybindings KeybindingsConfig `yaml:"keybindings"`
	UI          UIConfig          `yaml:"ui"`
	Theme       ThemeConfig       `yaml:"theme"`
	AI          *AIConfig         `yaml:"ai,omitempty"`
}

// KeybindingsConfig contains all keybinding configurations
type KeybindingsConfig struct {
	Resize      ResizeKeybindings      `yaml:"resize"`
	Layout      LayoutKeybindings      `yaml:"layout"`
	Navigation  NavigationKeybindings  `yaml:"navigation"`
	Global      GlobalKeybindings      `yaml:"global"`
	Connections ConnectionsKeybindings `yaml:"connections"`
	Schema      SchemaKeybindings      `yaml:"schema"`
}

// ResizeKeybindings for panel resizing
type ResizeKeybindings struct {
	GrowEditorLeft    string `yaml:"grow_editor_left"`
	ShrinkEditorLeft  string `yaml:"shrink_editor_left"`
	GrowEditorRight   string `yaml:"grow_editor_right"`
	ShrinkEditorRight string `yaml:"shrink_editor_right"`
}

// LayoutKeybindings for layout presets
type LayoutKeybindings struct {
	PresetMode     string `yaml:"preset_mode"`
	PresetDefault  string `yaml:"preset_default"`
	PresetEditor   string `yaml:"preset_editor"`
	PresetResults  string `yaml:"preset_results"`
	PresetBalanced string `yaml:"preset_balanced"`
}

// NavigationKeybindings for panel navigation
type NavigationKeybindings struct {
	FocusConnections string `yaml:"focus_connections"`
	FocusEditor      string `yaml:"focus_editor"`
	FocusResults     string `yaml:"focus_results"`
	NextPanel        string `yaml:"next_panel"`
	PrevPanel        string `yaml:"prev_panel"`
}

// GlobalKeybindings for global actions
type GlobalKeybindings struct {
	Help         string `yaml:"help"`
	Quit         string `yaml:"quit"`
	ExecuteQuery string `yaml:"execute_query"`
	SaveQuery    string `yaml:"save_query"`
	OpenNeovim   string `yaml:"open_neovim"`
	AIAssistant  string `yaml:"ai_assistant"`
}

// ConnectionsKeybindings for connections panel
type ConnectionsKeybindings struct {
	Add            string `yaml:"add"`
	Edit           string `yaml:"edit"`
	Delete         string `yaml:"delete"`
	Connect        string `yaml:"connect"`
	SchemaExplorer string `yaml:"schema_explorer"`
}

// SchemaKeybindings for schema explorer
type SchemaKeybindings struct {
	NavigateDown string `yaml:"navigate_down"`
	NavigateUp   string `yaml:"navigate_up"`
	Expand       string `yaml:"expand"`
	Preview      string `yaml:"preview"`
	Search       string `yaml:"search"`
	Refresh      string `yaml:"refresh"`
	Exit         string `yaml:"exit"`
}

// UIConfig contains UI-related settings
type UIConfig struct {
	DefaultLayout   PanelRatios `yaml:"default_layout"`
	ResizeIncrement int         `yaml:"resize_increment"`
	MinPanelWidth   int         `yaml:"min_panel_width"`
	MaxPanelWidth   int         `yaml:"max_panel_width"`
}

// PanelRatios defines the default panel width ratios
type PanelRatios struct {
	Connections int `yaml:"connections"`
	Editor      int `yaml:"editor"`
	Results     int `yaml:"results"`
}

// ThemeConfig contains theme settings
type ThemeConfig struct {
	Name               string `yaml:"name"`
	SyntaxHighlighting bool   `yaml:"syntax_highlighting"`
	SQLLinting         bool   `yaml:"sql_linting"`
}

// AIConfig contains AI assistant settings
type AIConfig struct {
	Enabled          bool   `yaml:"enabled"`
	CLITool          string `yaml:"cli_tool"` // copilot-cli|claude-cli|sgpt|mods|llm
	InjectInNeovim   bool   `yaml:"inject_in_neovim"`
	MaxTables        int    `yaml:"max_tables"`
	IncludeRowCounts bool   `yaml:"include_row_counts"`
	IncludeIndexes   bool   `yaml:"include_indexes"`
	ContextFormat    string `yaml:"context_format"` // comments|markdown|minimal

	// MCP Server Configuration
	MCPEnabled       bool   `yaml:"mcp_enabled"`
	MCPSmartTools    bool   `yaml:"mcp_smart_tools"`    // Enable AI-powered smart tools
	MCPCacheEnabled  bool   `yaml:"mcp_cache_enabled"`
	MCPMaxCacheSize  int64  `yaml:"mcp_max_cache_size"`
	MCPAIProvider    string `yaml:"mcp_ai_provider"` // claude|gemini|openai
}
