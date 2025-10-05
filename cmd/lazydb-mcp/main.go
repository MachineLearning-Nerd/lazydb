package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/MachineLearning-Nerd/lazydb/internal/config"
	"github.com/MachineLearning-Nerd/lazydb/internal/db"
	"github.com/MachineLearning-Nerd/lazydb/internal/mcp/server"
	"github.com/MachineLearning-Nerd/lazydb/internal/mcp/tools"
	"github.com/MachineLearning-Nerd/lazydb/internal/storage"
)

const (
	serverName    = "lazydb-mcp"
	serverVersion = "1.0.0"
)

func main() {
	// Parse command-line flags
	connectionName := flag.String("connection", "", "Connection name from LazyDB config (uses active if not specified)")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	// Setup logging to stderr (stdout is reserved for MCP protocol)
	if *verbose {
		fmt.Fprintf(os.Stderr, "Starting LazyDB MCP Server v%s\n", serverVersion)
	}

	// Load LazyDB configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Create connection manager
	connMgr := db.NewConnectionManager()

	// Load connections from storage
	savedConfig, err := storage.LoadConnections()
	if err != nil {
		if *verbose {
			fmt.Fprintf(os.Stderr, "Warning: Failed to load connections: %v\n", err)
		}
	} else {
		// Add all saved connections to the manager
		for _, connConfig := range savedConfig.Connections {
			conn := db.NewPostgresConnection(connConfig)
			connMgr.AddConnection(connConfig.Name, conn)
		}

		// Set active connection
		if savedConfig.ActiveConnection != "" {
			if err := connMgr.SetActive(savedConfig.ActiveConnection); err != nil {
				if *verbose {
					fmt.Fprintf(os.Stderr, "Warning: Failed to set active connection: %v\n", err)
				}
			}
		}
	}

	// Get database connection
	var conn db.Connection
	if *connectionName != "" {
		conn, err = connMgr.GetConnection(*connectionName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Connection '%s' not found\n", *connectionName)
			os.Exit(1)
		}
	} else {
		conn, err = connMgr.GetActive()
		if err != nil {
			fmt.Fprintf(os.Stderr, "No active connection found. Please specify --connection or set an active connection in LazyDB\n")
			os.Exit(1)
		}
	}

	// Connect to database if not already connected
	if conn.Status() != db.StatusConnected {
		ctx := context.Background()
		if err := conn.Connect(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
			os.Exit(1)
		}
	}

	if *verbose {
		connCfg := conn.Config()
		fmt.Fprintf(os.Stderr, "Connected to: %s (%s)\n", connCfg.Name, connCfg.Database)
	}

	// Create MCP server configuration
	mcpConfig := &server.Config{
		ServerName:    serverName,
		ServerVersion: serverVersion,
		EnableCache:   cfg.AI != nil && cfg.AI.MCPCacheEnabled,
		MaxCacheSize:  100 * 1024 * 1024, // 100MB default
		AIProvider:    getAIProvider(cfg),
		AIAPIKey:      getAIAPIKey(),
	}

	// Create MCP server
	mcpServer := server.NewMCPServer(conn, mcpConfig)

	// Register basic tools
	basicTools := tools.NewBasicTools(conn)
	basicTools.Register(mcpServer.GetRegistry())

	// Register advanced tools
	advancedTools := tools.NewAdvancedTools(conn)
	advancedTools.Register(mcpServer.GetRegistry())

	if *verbose {
		fmt.Fprintf(os.Stderr, "Registered %d tools (5 basic + 16 advanced)\n", mcpServer.GetRegistry().Count())
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		if *verbose {
			fmt.Fprintf(os.Stderr, "\nReceived signal %v, shutting down...\n", sig)
		}
		cancel()
	}()

	// Start MCP server (reads from stdin, writes to stdout)
	if *verbose {
		fmt.Fprintf(os.Stderr, "MCP server started, listening on stdin/stdout\n")
	}

	if err := mcpServer.Start(ctx, os.Stdin, os.Stdout); err != nil {
		if err != context.Canceled {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "MCP server stopped\n")
	}
}

// getAIProvider returns the AI provider from config
func getAIProvider(cfg *config.Config) string {
	if cfg.AI != nil && cfg.AI.MCPAIProvider != "" {
		return cfg.AI.MCPAIProvider
	}
	if cfg.AI != nil && cfg.AI.CLITool != "" {
		return cfg.AI.CLITool
	}
	return "claude"
}

// getAIAPIKey returns the AI API key from environment
func getAIAPIKey() string {
	// Try multiple environment variables
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		return key
	}
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		return key
	}
	if key := os.Getenv("GOOGLE_API_KEY"); key != "" {
		return key
	}
	return ""
}
