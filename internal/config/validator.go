package config

import (
	"fmt"
	"strings"
)

// ValidateConfig validates the configuration
func ValidateConfig(cfg *Config) error {
	// Validate panel ratios sum to 100
	total := cfg.UI.DefaultLayout.Connections + cfg.UI.DefaultLayout.Editor + cfg.UI.DefaultLayout.Results
	if total != 100 {
		return fmt.Errorf("panel ratios must sum to 100, got %d", total)
	}

	// Validate panel width constraints
	if cfg.UI.MinPanelWidth < 10 || cfg.UI.MinPanelWidth > 50 {
		return fmt.Errorf("min_panel_width must be between 10 and 50, got %d", cfg.UI.MinPanelWidth)
	}

	if cfg.UI.MaxPanelWidth < 50 || cfg.UI.MaxPanelWidth > 90 {
		return fmt.Errorf("max_panel_width must be between 50 and 90, got %d", cfg.UI.MaxPanelWidth)
	}

	if cfg.UI.MinPanelWidth >= cfg.UI.MaxPanelWidth {
		return fmt.Errorf("min_panel_width must be less than max_panel_width")
	}

	// Validate resize increment
	if cfg.UI.ResizeIncrement < 1 || cfg.UI.ResizeIncrement > 20 {
		return fmt.Errorf("resize_increment must be between 1 and 20, got %d", cfg.UI.ResizeIncrement)
	}

	// Check for duplicate keybindings
	if err := checkDuplicateKeys(cfg); err != nil {
		return err
	}

	return nil
}

// checkDuplicateKeys checks for duplicate keybindings
func checkDuplicateKeys(cfg *Config) error {
	keys := make(map[string]string)

	// Helper to check and add key
	addKey := func(key, context string) error {
		if key == "" {
			return nil // Empty keys are allowed (disabled bindings)
		}

		normalizedKey := strings.ToLower(key)
		if existing, found := keys[normalizedKey]; found {
			return fmt.Errorf("duplicate keybinding '%s' found in both %s and %s", key, existing, context)
		}
		keys[normalizedKey] = context
		return nil
	}

	// Check resize keys
	if err := addKey(cfg.Keybindings.Resize.GrowEditorLeft, "resize.grow_editor_left"); err != nil {
		return err
	}
	if err := addKey(cfg.Keybindings.Resize.ShrinkEditorLeft, "resize.shrink_editor_left"); err != nil {
		return err
	}
	if err := addKey(cfg.Keybindings.Resize.GrowEditorRight, "resize.grow_editor_right"); err != nil {
		return err
	}
	if err := addKey(cfg.Keybindings.Resize.ShrinkEditorRight, "resize.shrink_editor_right"); err != nil {
		return err
	}

	// Check layout keys
	if err := addKey(cfg.Keybindings.Layout.PresetMode, "layout.preset_mode"); err != nil {
		return err
	}
	// Note: Preset numbers can overlap with navigation keys (1,2,3) - this is intentional

	// Check global keys
	if err := addKey(cfg.Keybindings.Global.Help, "global.help"); err != nil {
		return err
	}
	if err := addKey(cfg.Keybindings.Global.Quit, "global.quit"); err != nil {
		return err
	}
	if err := addKey(cfg.Keybindings.Global.ExecuteQuery, "global.execute_query"); err != nil {
		return err
	}
	if err := addKey(cfg.Keybindings.Global.SaveQuery, "global.save_query"); err != nil {
		return err
	}
	if err := addKey(cfg.Keybindings.Global.OpenNeovim, "global.open_neovim"); err != nil {
		return err
	}

	// Check connections keys
	if err := addKey(cfg.Keybindings.Connections.Add, "connections.add"); err != nil {
		return err
	}
	if err := addKey(cfg.Keybindings.Connections.Edit, "connections.edit"); err != nil {
		return err
	}
	if err := addKey(cfg.Keybindings.Connections.Delete, "connections.delete"); err != nil {
		return err
	}
	if err := addKey(cfg.Keybindings.Connections.Connect, "connections.connect"); err != nil {
		return err
	}
	if err := addKey(cfg.Keybindings.Connections.SchemaExplorer, "connections.schema_explorer"); err != nil {
		return err
	}

	// Check schema keys
	if err := addKey(cfg.Keybindings.Schema.NavigateDown, "schema.navigate_down"); err != nil {
		return err
	}
	if err := addKey(cfg.Keybindings.Schema.NavigateUp, "schema.navigate_up"); err != nil {
		return err
	}
	if err := addKey(cfg.Keybindings.Schema.Expand, "schema.expand"); err != nil {
		return err
	}
	if err := addKey(cfg.Keybindings.Schema.Preview, "schema.preview"); err != nil {
		return err
	}
	if err := addKey(cfg.Keybindings.Schema.Search, "schema.search"); err != nil {
		return err
	}
	if err := addKey(cfg.Keybindings.Schema.Refresh, "schema.refresh"); err != nil {
		return err
	}
	if err := addKey(cfg.Keybindings.Schema.Exit, "schema.exit"); err != nil {
		return err
	}

	return nil
}
