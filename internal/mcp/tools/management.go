package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/MachineLearning-Nerd/lazydb/internal/mcp/server"
)

// SessionManager interface for session management operations
type SessionManager interface {
	EnableCategory(category string) error
	DisableCategory(category string) error
	GetEnabledCategories() []string
	IsCategoryEnabled(category string) bool
	ResetSession(preset string)
}

// ManagementTools provides tools for dynamic tool category management
type ManagementTools struct {
	sessionManager SessionManager
}

// NewManagementTools creates a new ManagementTools instance
func NewManagementTools(sessionManager SessionManager) *ManagementTools {
	return &ManagementTools{sessionManager: sessionManager}
}

// Register registers all management tools with the tool registry
func (t *ManagementTools) Register(registry *server.ToolRegistry) {
	// Tool 1: lazydb_enable_category
	registry.Register(
		server.Tool{
			Name:        "lazydb_enable_category",
			Description: "Enable a tool category for the current session. Use to add new capabilities on-demand.",
			Category:    "meta",
			Tags:        []string{"management", "dynamic", "enable", "session"},
			InputExamples: []map[string]interface{}{
				{"category": "performance"},
				{"category": "optimization"},
				{"category": "statistics"},
			},
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"category": {
						Type:        "string",
						Description: "Category to enable. Available: schema, performance, relationships, statistics, discovery, triggers, optimization, meta, compound",
						Enum:        []string{"schema", "performance", "relationships", "statistics", "discovery", "triggers", "optimization", "meta", "compound"},
						Examples:    []string{"performance", "optimization", "statistics"},
					},
				},
				Required: []string{"category"},
			},
		},
		t.enableCategory,
	)

	// Tool 2: lazydb_disable_category
	registry.Register(
		server.Tool{
			Name:        "lazydb_disable_category",
			Description: "Disable a tool category for the current session. Use to reduce context window size.",
			Category:    "meta",
			Tags:        []string{"management", "dynamic", "disable", "session"},
			InputExamples: []map[string]interface{}{
				{"category": "triggers"},
				{"category": "statistics"},
			},
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"category": {
						Type:        "string",
						Description: "Category to disable. Note: 'schema' and 'meta' are recommended to keep enabled.",
						Enum:        []string{"schema", "performance", "relationships", "statistics", "discovery", "triggers", "optimization", "meta", "compound"},
						Examples:    []string{"triggers", "statistics"},
					},
				},
				Required: []string{"category"},
			},
		},
		t.disableCategory,
	)

	// Tool 3: lazydb_list_categories
	registry.Register(
		server.Tool{
			Name:        "lazydb_list_categories",
			Description: "List all available tool categories with their enabled status and tool count.",
			Category:    "meta",
			Tags:        []string{"management", "discovery", "list", "session"},
			InputExamples: []map[string]interface{}{
				{},
			},
			InputSchema: server.InputSchema{
				Type:       "object",
				Properties: map[string]server.Property{},
				Required:   []string{},
			},
		},
		t.listCategories,
	)

	// Tool 4: lazydb_reset_session
	registry.Register(
		server.Tool{
			Name:        "lazydb_reset_session",
			Description: "Reset session to a preset configuration. Presets: minimal (~450 tokens), standard (~1200 tokens), performance (~1100 tokens), full (~2400 tokens).",
			Category:    "meta",
			Tags:        []string{"management", "reset", "preset", "session"},
			InputExamples: []map[string]interface{}{
				{},
				{"preset": "minimal"},
				{"preset": "performance"},
				{"preset": "full"},
			},
			InputSchema: server.InputSchema{
				Type: "object",
				Properties: map[string]server.Property{
					"preset": {
						Type:        "string",
						Description: "Preset to reset to. Default: minimal",
						Enum:        []string{"minimal", "standard", "performance", "full"},
						Default:     "minimal",
						Examples:    []string{"minimal", "standard", "performance", "full"},
					},
				},
				Required: []string{},
			},
		},
		t.resetSession,
	)
}

// enableCategory enables a tool category for the current session
func (t *ManagementTools) enableCategory(ctx context.Context, args map[string]interface{}) (string, error) {
	category, ok := args["category"].(string)
	if !ok {
		return "", fmt.Errorf("category parameter is required")
	}

	err := t.sessionManager.EnableCategory(category)
	if err != nil {
		return "", err
	}

	// Get updated enabled categories
	enabledCategories := t.sessionManager.GetEnabledCategories()
	sort.Strings(enabledCategories)

	// Get tools in the enabled category
	categoryTools := server.GetCategoryTools(category)

	output := map[string]interface{}{
		"success":            true,
		"action":             "enabled",
		"category":           category,
		"tools_added":        categoryTools,
		"tools_count":        len(categoryTools),
		"enabled_categories": enabledCategories,
		"message":            fmt.Sprintf("Category '%s' enabled. %d tools now available.", category, len(categoryTools)),
	}

	jsonOutput, _ := json.MarshalIndent(output, "", "  ")
	return string(jsonOutput), nil
}

// disableCategory disables a tool category for the current session
func (t *ManagementTools) disableCategory(ctx context.Context, args map[string]interface{}) (string, error) {
	category, ok := args["category"].(string)
	if !ok {
		return "", fmt.Errorf("category parameter is required")
	}

	// Get tools that will be removed
	categoryTools := server.GetCategoryTools(category)

	err := t.sessionManager.DisableCategory(category)
	if err != nil {
		return "", err
	}

	// Get updated enabled categories
	enabledCategories := t.sessionManager.GetEnabledCategories()
	sort.Strings(enabledCategories)

	output := map[string]interface{}{
		"success":            true,
		"action":             "disabled",
		"category":           category,
		"tools_removed":      categoryTools,
		"tools_count":        len(categoryTools),
		"enabled_categories": enabledCategories,
		"message":            fmt.Sprintf("Category '%s' disabled. %d tools removed from session.", category, len(categoryTools)),
	}

	jsonOutput, _ := json.MarshalIndent(output, "", "  ")
	return string(jsonOutput), nil
}

// listCategories lists all available categories with their status
func (t *ManagementTools) listCategories(ctx context.Context, args map[string]interface{}) (string, error) {
	allCategories := server.AllCategories()
	enabledCategories := t.sessionManager.GetEnabledCategories()

	// Build enabled set for fast lookup
	enabledSet := make(map[string]bool)
	for _, cat := range enabledCategories {
		enabledSet[cat] = true
	}

	categories := make([]map[string]interface{}, 0, len(allCategories))
	totalEnabled := 0
	totalTools := 0

	for _, cat := range allCategories {
		tools := server.GetCategoryTools(cat)
		description := server.CategoryDescriptions[cat]
		enabled := enabledSet[cat]

		if enabled {
			totalEnabled++
			totalTools += len(tools)
		}

		categories = append(categories, map[string]interface{}{
			"name":        cat,
			"description": description,
			"tools":       tools,
			"tool_count":  len(tools),
			"enabled":     enabled,
		})
	}

	// Add preset information
	presets := map[string][]string{
		"minimal":     server.GetPresetCategories("minimal"),
		"standard":    server.GetPresetCategories("standard"),
		"performance": server.GetPresetCategories("performance"),
		"full":        server.GetPresetCategories("full"),
	}

	output := map[string]interface{}{
		"categories":            categories,
		"total_categories":      len(allCategories),
		"enabled_count":         totalEnabled,
		"total_enabled_tools":   totalTools,
		"available_presets":     presets,
		"always_on_tools":       server.AlwaysOnTools,
		"always_on_tools_count": len(server.AlwaysOnTools),
	}

	jsonOutput, _ := json.MarshalIndent(output, "", "  ")
	return string(jsonOutput), nil
}

// resetSession resets the session to a preset configuration
func (t *ManagementTools) resetSession(ctx context.Context, args map[string]interface{}) (string, error) {
	preset := "minimal"
	if p, ok := args["preset"].(string); ok && p != "" {
		preset = p
	}

	// Validate preset
	validPresets := []string{"minimal", "standard", "performance", "full"}
	valid := false
	for _, p := range validPresets {
		if p == preset {
			valid = true
			break
		}
	}
	if !valid {
		return "", fmt.Errorf("invalid preset: %s. Valid presets: minimal, standard, performance, full", preset)
	}

	t.sessionManager.ResetSession(preset)

	// Get updated enabled categories
	enabledCategories := t.sessionManager.GetEnabledCategories()
	sort.Strings(enabledCategories)

	// Count total tools
	totalTools := 0
	for _, cat := range enabledCategories {
		totalTools += len(server.GetCategoryTools(cat))
	}

	// Token estimates by preset
	tokenEstimates := map[string]string{
		"minimal":     "~450 tokens",
		"standard":    "~1,200 tokens",
		"performance": "~1,100 tokens",
		"full":        "~2,400 tokens",
	}

	output := map[string]interface{}{
		"success":            true,
		"action":             "reset",
		"preset":             preset,
		"enabled_categories": enabledCategories,
		"total_tools":        totalTools,
		"token_estimate":     tokenEstimates[preset],
		"message":            fmt.Sprintf("Session reset to '%s' preset with %d categories and %d tools.", preset, len(enabledCategories), totalTools),
	}

	jsonOutput, _ := json.MarshalIndent(output, "", "  ")
	return string(jsonOutput), nil
}
