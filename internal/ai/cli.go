package ai

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// CLIProvider represents an AI CLI tool interface
type CLIProvider interface {
	Name() string
	IsAvailable() bool
	BuildCommand(ctx context.Context, schemaCtx *SchemaContext, query string, task string) (*exec.Cmd, error)
	ParseResponse(output string) string
}

// DetectAvailableCLI detects which AI CLI tools are installed
func DetectAvailableCLI(preferred string) CLIProvider {
	providers := []CLIProvider{
		&CopilotCLI{},
		&ClaudeCLI{},
		&ShellGPT{},
		&Mods{},
		&LLM{},
	}

	// First try the preferred provider if specified
	if preferred != "" {
		for _, provider := range providers {
			if strings.EqualFold(provider.Name(), preferred) && provider.IsAvailable() {
				return provider
			}
		}
	}

	// Fallback to first available provider
	for _, provider := range providers {
		if provider.IsAvailable() {
			return provider
		}
	}

	return nil
}

// GetAvailableProviders returns a list of all available CLI providers
func GetAvailableProviders() []string {
	providers := []CLIProvider{
		&CopilotCLI{},
		&ClaudeCLI{},
		&ShellGPT{},
		&Mods{},
		&LLM{},
	}

	var available []string
	for _, provider := range providers {
		if provider.IsAvailable() {
			available = append(available, provider.Name())
		}
	}

	return available
}

// CopilotCLI implements GitHub Copilot CLI
type CopilotCLI struct{}

func (c *CopilotCLI) Name() string {
	return "copilot-cli"
}

func (c *CopilotCLI) IsAvailable() bool {
	_, err := exec.LookPath("copilot")
	return err == nil
}

func (c *CopilotCLI) BuildCommand(ctx context.Context, schemaCtx *SchemaContext, query string, task string) (*exec.Cmd, error) {
	// Build prompt
	prompt := fmt.Sprintf("%s\n\nSchema Context:\n%s\n\nCurrent Query:\n%s",
		task,
		schemaCtx.FormatAsPlainText(),
		query)

	cmd := exec.CommandContext(ctx, "copilot", "suggest", prompt)
	return cmd, nil
}

func (c *CopilotCLI) ParseResponse(output string) string {
	// Copilot output is usually clean, just trim whitespace
	return strings.TrimSpace(output)
}

// ClaudeCLI implements Anthropic Claude CLI
type ClaudeCLI struct{}

func (c *ClaudeCLI) Name() string {
	return "claude-cli"
}

func (c *ClaudeCLI) IsAvailable() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

func (c *ClaudeCLI) BuildCommand(ctx context.Context, schemaCtx *SchemaContext, query string, task string) (*exec.Cmd, error) {
	var prompt string

	// Check if MCP mode is enabled (schema context will be nil or explicitly disabled)
	if schemaCtx == nil || schemaCtx.UseMCP {
		// MCP Mode: Don't inject schema, let MCP tools handle it
		prompt = fmt.Sprintf(`Task: %s

Current SQL Query:
%s

Please use the available lazydb MCP tools to access the database schema, inspect database objects, analyze queries, and provide an improved version of the query or explanation as requested.

The lazydb MCP server provides comprehensive database inspection tools including:
- Schema exploration (tables, views, functions, sequences)
- DDL generation and object definitions
- Index and constraint analysis
- Query performance analysis (EXPLAIN)
- Table statistics and storage metrics
- Foreign key relationships and dependencies
- Trigger inspection

Use these tools as needed to understand the database structure and provide the best solution.`,
			task,
			query)
	} else {
		// Legacy Mode: Inject schema context directly
		prompt = fmt.Sprintf(`Task: %s

Database Schema:
%s

Current SQL Query:
%s

Please provide an improved version of the query or explanation as requested.`,
			task,
			schemaCtx.FormatAsMarkdown(),
			query)
	}

	// Bypass permissions for MCP tools when in MCP mode
	if schemaCtx != nil && schemaCtx.UseMCP {
		cmd := exec.CommandContext(ctx, "claude", "--permission-mode", "bypassPermissions", "-p", prompt)
		return cmd, nil
	}

	cmd := exec.CommandContext(ctx, "claude", "-p", prompt)
	return cmd, nil
}

func (c *ClaudeCLI) ParseResponse(output string) string {
	// Claude output is usually well-formatted, just trim
	return strings.TrimSpace(output)
}

// ShellGPT implements shell-gpt (sgpt)
type ShellGPT struct{}

func (s *ShellGPT) Name() string {
	return "sgpt"
}

func (s *ShellGPT) IsAvailable() bool {
	_, err := exec.LookPath("sgpt")
	return err == nil
}

func (s *ShellGPT) BuildCommand(ctx context.Context, schemaCtx *SchemaContext, query string, task string) (*exec.Cmd, error) {
	// Build prompt
	prompt := fmt.Sprintf(`PostgreSQL Query Task: %s

Database Schema:
%s

Current Query:
%s`,
		task,
		schemaCtx.FormatAsPlainText(),
		query)

	cmd := exec.CommandContext(ctx, "sgpt", "--shell", prompt)
	return cmd, nil
}

func (s *ShellGPT) ParseResponse(output string) string {
	return strings.TrimSpace(output)
}

// Mods implements charm's mods CLI
type Mods struct{}

func (m *Mods) Name() string {
	return "mods"
}

func (m *Mods) IsAvailable() bool {
	_, err := exec.LookPath("mods")
	return err == nil
}

func (m *Mods) BuildCommand(ctx context.Context, schemaCtx *SchemaContext, query string, task string) (*exec.Cmd, error) {
	// Mods works best with stdin for context
	prompt := fmt.Sprintf(`%s

Schema: %s

Query: %s`,
		task,
		schemaCtx.FormatAsMinimal(),
		query)

	cmd := exec.CommandContext(ctx, "mods", prompt)
	return cmd, nil
}

func (m *Mods) ParseResponse(output string) string {
	return strings.TrimSpace(output)
}

// LLM implements Simon Willison's llm CLI
type LLM struct{}

func (l *LLM) Name() string {
	return "llm"
}

func (l *LLM) IsAvailable() bool {
	_, err := exec.LookPath("llm")
	return err == nil
}

func (l *LLM) BuildCommand(ctx context.Context, schemaCtx *SchemaContext, query string, task string) (*exec.Cmd, error) {
	// Build prompt
	prompt := fmt.Sprintf(`%s

Database Schema:
%s

SQL Query:
%s`,
		task,
		schemaCtx.FormatAsPlainText(),
		query)

	cmd := exec.CommandContext(ctx, "llm", prompt)
	return cmd, nil
}

func (l *LLM) ParseResponse(output string) string {
	return strings.TrimSpace(output)
}

// InvokeCLI is a helper function to invoke a CLI provider and get the response
func InvokeCLI(ctx context.Context, provider CLIProvider, schemaCtx *SchemaContext, query string, task string) (string, error) {
	if provider == nil {
		return "", fmt.Errorf("no AI CLI provider available")
	}

	cmd, err := provider.BuildCommand(ctx, schemaCtx, query, task)
	if err != nil {
		return "", fmt.Errorf("failed to build command: %w", err)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute %s: %w\nOutput: %s", provider.Name(), err, string(output))
	}

	response := provider.ParseResponse(string(output))
	return response, nil
}
