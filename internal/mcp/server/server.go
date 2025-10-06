package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/MachineLearning-Nerd/lazydb/internal/db"
)

// Config holds MCP server configuration
type Config struct {
	ServerName    string
	ServerVersion string
	EnableCache   bool
	MaxCacheSize  int64
	AIProvider    string // For smart tools (claude, gemini, openai)
	AIAPIKey      string
}

// ConnectionGetter is a function that returns the current database connection
type ConnectionGetter func() (db.Connection, error)

// MCPServer is the main MCP server
type MCPServer struct {
	conn         db.Connection     // Deprecated: use connGetter instead
	connGetter   ConnectionGetter  // Dynamic connection getter
	toolRegistry *ToolRegistry
	config       *Config
	initialized  bool
}

// NewMCPServer creates a new MCP server instance (legacy, uses static connection)
func NewMCPServer(conn db.Connection, config *Config) *MCPServer {
	return &MCPServer{
		conn:         conn,
		connGetter:   nil,
		toolRegistry: NewToolRegistry(),
		config:       config,
		initialized:  false,
	}
}

// NewMCPServerWithGetter creates a new MCP server with dynamic connection getter
func NewMCPServerWithGetter(connGetter ConnectionGetter, config *Config) *MCPServer {
	return &MCPServer{
		conn:         nil,
		connGetter:   connGetter,
		toolRegistry: NewToolRegistry(),
		config:       config,
		initialized:  false,
	}
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

	tools := s.toolRegistry.GetAllTools()

	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}
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
