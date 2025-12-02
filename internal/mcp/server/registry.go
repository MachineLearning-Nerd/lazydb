package server

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// ToolHandler is a function that executes a tool
type ToolHandler func(ctx context.Context, args map[string]interface{}) (string, error)

// ToolRegistry manages available tools and their handlers
type ToolRegistry struct {
	mu       sync.RWMutex
	tools    map[string]Tool
	handlers map[string]ToolHandler
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools:    make(map[string]Tool),
		handlers: make(map[string]ToolHandler),
	}
}

// Register adds a tool and its handler to the registry
func (r *ToolRegistry) Register(tool Tool, handler ToolHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tools[tool.Name] = tool
	r.handlers[tool.Name] = handler
}

// GetAllTools returns all registered tools
func (r *ToolRegistry) GetAllTools() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// GetTool returns a specific tool by name
func (r *ToolRegistry) GetTool(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, ok := r.tools[name]
	return tool, ok
}

// ExecuteTool executes a tool by name with given arguments
func (r *ToolRegistry) ExecuteTool(ctx context.Context, name string, args map[string]interface{}) (string, error) {
	r.mu.RLock()
	handler, ok := r.handlers[name]
	r.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("tool not found: %s", name)
	}

	return handler(ctx, args)
}

// HasTool checks if a tool is registered
func (r *ToolRegistry) HasTool(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.tools[name]
	return ok
}

// Count returns the number of registered tools
func (r *ToolRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.tools)
}

// GetToolsByCategory returns tools filtered by category
func (r *ToolRegistry) GetToolsByCategory(categories []string) []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(categories) == 0 {
		// No filter, return all tools
		tools := make([]Tool, 0, len(r.tools))
		for _, tool := range r.tools {
			tools = append(tools, tool)
		}
		return tools
	}

	// Build category set for fast lookup
	categorySet := make(map[string]bool)
	for _, cat := range categories {
		categorySet[cat] = true
	}

	tools := make([]Tool, 0)
	for _, tool := range r.tools {
		if categorySet[tool.Category] {
			tools = append(tools, tool)
		}
	}
	return tools
}

// GetToolsByTags returns tools matching any of the tags
func (r *ToolRegistry) GetToolsByTags(tags []string) []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(tags) == 0 {
		return r.GetAllTools()
	}

	// Build tag set for fast lookup
	tagSet := make(map[string]bool)
	for _, tag := range tags {
		tagSet[strings.ToLower(tag)] = true
	}

	tools := make([]Tool, 0)
	for _, tool := range r.tools {
		for _, toolTag := range tool.Tags {
			if tagSet[strings.ToLower(toolTag)] {
				tools = append(tools, tool)
				break
			}
		}
	}
	return tools
}

// SearchTools returns tools matching keyword search in name, description, or tags
func (r *ToolRegistry) SearchTools(query string) []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if query == "" {
		return r.GetAllTools()
	}

	queryLower := strings.ToLower(query)
	tools := make([]Tool, 0)

	for _, tool := range r.tools {
		// Search in name
		if strings.Contains(strings.ToLower(tool.Name), queryLower) {
			tools = append(tools, tool)
			continue
		}
		// Search in description
		if strings.Contains(strings.ToLower(tool.Description), queryLower) {
			tools = append(tools, tool)
			continue
		}
		// Search in tags
		for _, tag := range tool.Tags {
			if strings.Contains(strings.ToLower(tag), queryLower) {
				tools = append(tools, tool)
				break
			}
		}
	}
	return tools
}

// GetCategories returns all unique categories from registered tools
func (r *ToolRegistry) GetCategories() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	categorySet := make(map[string]bool)
	for _, tool := range r.tools {
		if tool.Category != "" {
			categorySet[tool.Category] = true
		}
	}

	categories := make([]string, 0, len(categorySet))
	for cat := range categorySet {
		categories = append(categories, cat)
	}
	return categories
}

// GetTags returns all unique tags from registered tools
func (r *ToolRegistry) GetTags() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tagSet := make(map[string]bool)
	for _, tool := range r.tools {
		for _, tag := range tool.Tags {
			tagSet[tag] = true
		}
	}

	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	return tags
}
