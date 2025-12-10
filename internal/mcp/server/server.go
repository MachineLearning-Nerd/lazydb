package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/MachineLearning-Nerd/lazydb/internal/db"
)

// Config holds MCP server configuration
type Config struct {
	ServerName    string
	ServerVersion string
	EnableCache   bool
	MaxCacheSize  int64
	AIProvider    string   // For smart tools (claude, gemini, openai)
	AIAPIKey      string
	Categories    []string // Tool categories to enable (empty = all)
}

// ConnectionGetter is a function that returns the current database connection
type ConnectionGetter func() (db.Connection, error)

// AlwaysOnTools lists tools that are always available regardless of session categories
var AlwaysOnTools = []string{
	"lazydb_enable_category",
	"lazydb_disable_category",
	"lazydb_list_categories",
	"lazydb_reset_session",
	"search_lazydb_tools", // Tool discovery is also always available
}

// MCPServer is the main MCP server
type MCPServer struct {
	conn         db.Connection     // Deprecated: use connGetter instead
	connGetter   ConnectionGetter  // Dynamic connection getter
	toolRegistry *ToolRegistry
	config       *Config
	initialized  bool

	// Session state for dynamic tool management (Docker-style)
	sessionCategories map[string]bool // Currently enabled categories
	sessionMu         sync.RWMutex    // Protect session state
}

// NewMCPServer creates a new MCP server instance (legacy, uses static connection)
func NewMCPServer(conn db.Connection, config *Config) *MCPServer {
	server := &MCPServer{
		conn:              conn,
		connGetter:        nil,
		toolRegistry:      NewToolRegistry(),
		config:            config,
		initialized:       false,
		sessionCategories: make(map[string]bool),
	}
	// Initialize with minimal preset by default
	server.ResetSession("minimal")
	return server
}

// NewMCPServerWithGetter creates a new MCP server with dynamic connection getter
func NewMCPServerWithGetter(connGetter ConnectionGetter, config *Config) *MCPServer {
	server := &MCPServer{
		conn:              nil,
		connGetter:        connGetter,
		toolRegistry:      NewToolRegistry(),
		config:            config,
		initialized:       false,
		sessionCategories: make(map[string]bool),
	}
	// Initialize with minimal preset by default
	server.ResetSession("minimal")
	return server
}

// GetRegistry returns the tool registry for registration
func (s *MCPServer) GetRegistry() *ToolRegistry {
	return s.toolRegistry
}

// GetConnection returns the database connection (supports both static and dynamic)
func (s *MCPServer) GetConnection() db.Connection {
	// Use dynamic connection getter if available
	if s.connGetter != nil {
		conn, err := s.connGetter()
		if err != nil {
			return nil
		}
		return conn
	}
	// Fallback to static connection (legacy)
	return s.conn
}

// GetConfig returns the server configuration
func (s *MCPServer) GetConfig() *Config {
	return s.config
}

// EnableCategory enables a tool category for the current session
func (s *MCPServer) EnableCategory(category string) error {
	// Validate category exists
	categories := AllCategories()
	valid := false
	for _, c := range categories {
		if c == category {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid category: %s", category)
	}

	s.sessionMu.Lock()
	defer s.sessionMu.Unlock()
	s.sessionCategories[category] = true
	return nil
}

// DisableCategory disables a tool category for the current session
func (s *MCPServer) DisableCategory(category string) error {
	// Validate category exists
	categories := AllCategories()
	valid := false
	for _, c := range categories {
		if c == category {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid category: %s", category)
	}

	s.sessionMu.Lock()
	defer s.sessionMu.Unlock()
	delete(s.sessionCategories, category)
	return nil
}

// GetEnabledCategories returns currently enabled categories
func (s *MCPServer) GetEnabledCategories() []string {
	s.sessionMu.RLock()
	defer s.sessionMu.RUnlock()

	categories := make([]string, 0, len(s.sessionCategories))
	for cat := range s.sessionCategories {
		categories = append(categories, cat)
	}
	return categories
}

// IsCategoryEnabled checks if a category is currently enabled
func (s *MCPServer) IsCategoryEnabled(category string) bool {
	s.sessionMu.RLock()
	defer s.sessionMu.RUnlock()
	return s.sessionCategories[category]
}

// IsAlwaysOnTool checks if a tool is in the always-on list
func (s *MCPServer) IsAlwaysOnTool(name string) bool {
	for _, t := range AlwaysOnTools {
		if t == name {
			return true
		}
	}
	return false
}

// ResetSession resets the session to a preset (default: "minimal")
func (s *MCPServer) ResetSession(preset string) {
	if preset == "" {
		preset = "minimal"
	}

	presetCategories := GetPresetCategories(preset)
	if presetCategories == nil {
		// Fallback to minimal if preset not found
		presetCategories = GetPresetCategories("minimal")
	}

	s.sessionMu.Lock()
	defer s.sessionMu.Unlock()

	// Clear existing categories
	s.sessionCategories = make(map[string]bool)

	// Enable preset categories
	for _, cat := range presetCategories {
		s.sessionCategories[cat] = true
	}
}

// Start runs the MCP server main loop (stdin/stdout)
func (s *MCPServer) Start(ctx context.Context, reader io.Reader, writer io.Writer) error {
	decoder := json.NewDecoder(reader)
	encoder := json.NewEncoder(writer)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			var req MCPRequest
			if err := decoder.Decode(&req); err != nil {
				if err == io.EOF {
					return nil
				}
				// Log error but continue
				continue
			}

			// Handle request
			resp := s.handleRequest(ctx, &req)

			// Send response
			if err := encoder.Encode(resp); err != nil {
				return fmt.Errorf("failed to encode response: %w", err)
			}
		}
	}
}

// handleRequest processes an MCP request and returns a response
func (s *MCPServer) handleRequest(ctx context.Context, req *MCPRequest) *MCPResponse {
	switch req.Method {
	case MethodInitialize:
		return s.handleInitialize(req)
	case MethodListTools:
		return s.handleListTools(req)
	case MethodCallTool:
		return s.handleCallTool(ctx, req)
	case MethodListResources:
		return s.handleListResources(req)
	case MethodListPrompts:
		return s.handleListPrompts(req)
	default:
		return s.errorResponse(req.ID, ErrorCodeMethodNotFound, fmt.Sprintf("Method not found: %s", req.Method))
	}
}

// handleInitialize processes the initialize request
func (s *MCPServer) handleInitialize(req *MCPRequest) *MCPResponse {
	s.initialized = true

	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		ServerInfo: ServerInfo{
			Name:    s.config.ServerName,
			Version: s.config.ServerVersion,
		},
		Capabilities: map[string]interface{}{
			"tools": map[string]bool{
				"list": true,
				"call": true,
			},
		},
	}

	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// handleListTools processes the tools/list request
func (s *MCPServer) handleListTools(req *MCPRequest) *MCPResponse {
	if !s.initialized {
		return s.errorResponse(req.ID, ErrorCodeInvalidRequest, "Server not initialized")
	}

	var tools []Tool
	var runtimeFilterUsed bool

	// Check for category filter in params (runtime override)
	if req.Params != nil {
		if categories, ok := req.Params["categories"].([]interface{}); ok && len(categories) > 0 {
			// Convert []interface{} to []string
			categoryStrings := make([]string, 0, len(categories))
			for _, c := range categories {
				if str, ok := c.(string); ok {
					categoryStrings = append(categoryStrings, str)
				}
			}
			tools = s.toolRegistry.GetToolsByCategory(categoryStrings)
			runtimeFilterUsed = true
		} else if tags, ok := req.Params["tags"].([]interface{}); ok && len(tags) > 0 {
			// Support tag-based filtering
			tagStrings := make([]string, 0, len(tags))
			for _, t := range tags {
				if str, ok := t.(string); ok {
					tagStrings = append(tagStrings, str)
				}
			}
			tools = s.toolRegistry.GetToolsByTags(tagStrings)
			runtimeFilterUsed = true
		} else if query, ok := req.Params["search"].(string); ok && query != "" {
			// Support search-based filtering
			tools = s.toolRegistry.SearchTools(query)
			runtimeFilterUsed = true
		}
	}

	// Use session-enabled categories if no runtime filter was used
	if !runtimeFilterUsed {
		enabledCategories := s.GetEnabledCategories()
		if len(enabledCategories) > 0 {
			tools = s.toolRegistry.GetToolsByCategory(enabledCategories)
		} else {
			// Fallback to all tools if no categories enabled (shouldn't happen with minimal preset)
			tools = s.toolRegistry.GetAllTools()
		}
	}

	// Always include always-on tools (management + discovery tools)
	tools = s.ensureAlwaysOnTools(tools)

	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}
}

// ensureAlwaysOnTools ensures management tools are always included in the tool list
func (s *MCPServer) ensureAlwaysOnTools(tools []Tool) []Tool {
	// Build a set of existing tool names for fast lookup
	existingTools := make(map[string]bool)
	for _, tool := range tools {
		existingTools[tool.Name] = true
	}

	// Add always-on tools if not already present
	for _, name := range AlwaysOnTools {
		if !existingTools[name] {
			if tool, ok := s.toolRegistry.GetTool(name); ok {
				tools = append(tools, tool)
			}
		}
	}

	return tools
}

// handleCallTool processes the tools/call request
func (s *MCPServer) handleCallTool(ctx context.Context, req *MCPRequest) *MCPResponse {
	if !s.initialized {
		return s.errorResponse(req.ID, ErrorCodeInvalidRequest, "Server not initialized")
	}

	// Extract tool name and arguments
	params := req.Params
	toolName, ok := params["name"].(string)
	if !ok {
		return s.errorResponse(req.ID, ErrorCodeInvalidParams, "Missing or invalid 'name' parameter")
	}

	arguments, _ := params["arguments"].(map[string]interface{})
	if arguments == nil {
		arguments = make(map[string]interface{})
	}

	// Execute tool
	result, err := s.toolRegistry.ExecuteTool(ctx, toolName, arguments)
	if err != nil {
		return s.errorResponse(req.ID, ErrorCodeInternalError, fmt.Sprintf("Tool execution failed: %v", err))
	}

	// Return result as MCP content block
	toolResult := ToolCallResult{
		Content: []ContentBlock{
			{
				Type: "text",
				Text: result,
			},
		},
	}

	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  toolResult,
	}
}

// handleListResources processes the resources/list request (not implemented yet)
func (s *MCPServer) handleListResources(req *MCPRequest) *MCPResponse {
	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"resources": []interface{}{},
		},
	}
}

// handleListPrompts processes the prompts/list request (not implemented yet)
func (s *MCPServer) handleListPrompts(req *MCPRequest) *MCPResponse {
	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"prompts": []interface{}{},
		},
	}
}

// errorResponse creates an error response
func (s *MCPServer) errorResponse(id interface{}, code int, message string) *MCPResponse {
	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
	}
}
