package server

import (
	"context"
	"fmt"
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
