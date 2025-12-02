# LazyDB Agent Integration Guide

A comprehensive guide for building an external AI agent that integrates with LazyDB for intelligent database operations.

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Why External Agent?](#why-external-agent)
4. [Integration Points](#integration-points)
5. [MCP Tools Reference](#mcp-tools-reference)
6. [Implementation Guide](#implementation-guide)
7. [Care Points](#care-points)
8. [Code Templates](#code-templates)
9. [Testing](#testing)
10. [Deployment](#deployment)

---

## Overview

### What is LazyDB Agent?

LazyDB Agent is an external AI-powered assistant designed to help developers work with databases more efficiently. It integrates with LazyDB (a PostgreSQL TUI client) via the Model Context Protocol (MCP) to provide:

- Natural language to SQL conversion
- Query optimization suggestions
- Schema exploration and analysis
- Performance troubleshooting
- Migration generation
- Index recommendations

### Goals

| Goal | Description |
|------|-------------|
| **Speed** | Faster response times than large models (target: <2s) |
| **Accuracy** | High-quality SQL generation with schema awareness |
| **Integration** | Seamless integration with LazyDB TUI |
| **Extensibility** | Easy to add new tools and capabilities |

---

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              LazyDB TUI                                 │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌──────────────────┐  │
│  │ Connections│  │   Editor   │  │  Results   │  │  AI Assistant    │  │
│  │   Panel    │  │   Panel    │  │   Panel    │  │  (Modal Dialog)  │  │
│  └────────────┘  └────────────┘  └────────────┘  └────────┬─────────┘  │
└───────────────────────────────────────────────────────────┼─────────────┘
                                                            │ stdio
                                                            ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                         LazyDB Agent (External Process)                 │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                        Agent Core                                │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐   │   │
│  │  │ Task Router  │  │ Tool Manager │  │  Response Formatter  │   │   │
│  │  └──────────────┘  └──────────────┘  └──────────────────────┘   │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                    │                                    │
│  ┌─────────────────────────────────┴─────────────────────────────────┐ │
│  │                         MCP Client                                 │ │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐               │ │
│  │  │ list_tables │  │ get_schema  │  │ sample_data │  ...          │ │
│  │  └─────────────┘  └─────────────┘  └─────────────┘               │ │
│  └───────────────────────────────────────────────────────────────────┘ │
│                                    │                                    │
│  ┌─────────────────────────────────┴─────────────────────────────────┐ │
│  │                         LLM Client                                 │ │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐               │ │
│  │  │ Claude Haiku│  │ GPT-4o-mini │  │ Ollama/Local│               │ │
│  │  └─────────────┘  └─────────────┘  └─────────────┘               │ │
│  └───────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼ stdio
┌─────────────────────────────────────────────────────────────────────────┐
│                         LazyDB MCP Server                               │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                      Tool Registry                               │   │
│  │  list_all_tables | get_table_schema | search_tables | ...       │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                    │                                    │
│  ┌─────────────────────────────────┴─────────────────────────────────┐ │
│  │                      PostgreSQL Connection                        │ │
│  └───────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────┘
```

### Data Flow

```
1. User asks question in LazyDB AI Assistant
                    │
                    ▼
2. LazyDB spawns Agent process with task + context
                    │
                    ▼
3. Agent analyzes task, decides which tools to use
                    │
                    ▼
4. Agent calls MCP tools to gather schema/data info
                    │
                    ▼
5. Agent sends context + task to LLM
                    │
                    ▼
6. LLM generates response (SQL, analysis, etc.)
                    │
                    ▼
7. Agent formats response and returns to LazyDB
                    │
                    ▼
8. LazyDB displays response in AI Assistant panel
```

---

## Why External Agent?

### Advantages

| Advantage | Description |
|-----------|-------------|
| **Language Flexibility** | Build in Python, Go, Rust, or any language with good LLM libraries |
| **Easy Iteration** | Update agent without recompiling LazyDB |
| **Independent Scaling** | Agent can be deployed separately if needed |
| **Testing Isolation** | Test agent independently of LazyDB UI |
| **LLM Library Access** | Use mature Python libraries (LangChain, LlamaIndex, etc.) |

### Tradeoffs

| Tradeoff | Mitigation |
|----------|------------|
| Process overhead | Minimal (~10-50ms startup) |
| IPC complexity | Use simple stdio protocol |
| State management | Read connection state from shared file |

---

## Integration Points

### 1. CLI Provider Interface

LazyDB calls external agents via the CLI Provider interface:

```go
// Location: internal/ai/cli.go

type CLIProvider interface {
    // Name returns the provider identifier
    Name() string

    // IsAvailable checks if the provider binary exists
    IsAvailable() bool

    // BuildCommand creates the exec.Cmd to run the agent
    BuildCommand(ctx context.Context, schemaCtx *SchemaContext,
                 query string, task string) (*exec.Cmd, error)

    // ParseResponse processes the agent output
    ParseResponse(output string) string
}
```

### 2. Agent Invocation

LazyDB invokes your agent with these inputs:

```bash
lazydb-agent \
  --task "Write a query to find all users who signed up last month" \
  --query "SELECT * FROM users" \           # Current editor content
  --connection "production-db" \            # Active connection name
  --schema-context '{"tables": [...]}' \    # Optional: minimal schema info
  --use-mcp                                 # Flag to use MCP for schema
```

### 3. Expected Output Format

Your agent should output markdown that the AI Assistant panel can parse:

```markdown
## Analysis
Your analysis of the user's request...

## SQL Query
```sql
SELECT u.id, u.email, u.created_at
FROM users u
WHERE u.created_at >= DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month')
  AND u.created_at < DATE_TRUNC('month', CURRENT_DATE);
```

## Explanation
- Uses DATE_TRUNC for precise month boundaries
- Avoids timezone issues with date comparisons
- Index-friendly query pattern

## Recommendations
- Ensure index exists on `users.created_at`
- Consider adding LIMIT for large tables
```

### 4. Response Section Types

The AI Assistant panel recognizes these section types:

| Type | Detection | Use Case |
|------|-----------|----------|
| `code` | Triple backticks | SQL queries, code snippets |
| `header` | `##` or `**text**:` | Section titles |
| `list` | `-`, `*`, or `1.` | Recommendations, steps |
| `query` | SQL keywords | Executable SQL |
| `text` | Default | Explanations |

---

## MCP Tools Reference

### Available Tools

Your agent can call these MCP tools via the LazyDB MCP server:

#### Schema Discovery

| Tool | Purpose | Input | Output |
|------|---------|-------|--------|
| `list_all_tables` | Get all tables grouped by schema | `{schema?: string}` | Tables with row counts |
| `get_table_schema` | Detailed column info | `{table_name: string, include_constraints?: bool}` | Columns, types, constraints, FKs |
| `search_tables` | Find tables by pattern | `{pattern: string, schema?: string}` | Matching table names |
| `get_table_references` | FK relationships | `{table_name: string}` | Incoming/outgoing FKs |

#### Data Inspection

| Tool | Purpose | Input | Output |
|------|---------|-------|--------|
| `get_sample_data` | Preview rows | `{table_name: string, limit?: int}` | Sample rows (max 10) |
| `get_table_count` | Row count | `{table_name: string}` | Count |
| `get_column_stats` | Column statistics | `{table_name: string, column_name?: string}` | Nulls, distinct, distribution |

#### Query Analysis

| Tool | Purpose | Input | Output |
|------|---------|-------|--------|
| `explain_query` | Query plan | `{query: string, analyze?: bool}` | EXPLAIN output |
| `get_table_indexes` | Index info | `{table_name: string}` | Index definitions |
| `get_table_size` | Storage info | `{table_name: string}` | Size, bloat estimate |

#### Advanced Tools

| Tool | Purpose | Input | Output |
|------|---------|-------|--------|
| `get_table_ddl` | Generate CREATE statement | `{table_name: string}` | DDL |
| `get_view_definition` | View SQL | `{view_name: string}` | CREATE VIEW statement |
| `get_function_definition` | Function source | `{function_name: string}` | Function code |
| `list_sequences` | All sequences | `{schema?: string}` | Sequence info |
| `list_materialized_views` | MVs info | `{schema?: string}` | MV definitions |

### MCP Protocol

#### Request Format

```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "get_table_schema",
    "arguments": {
      "table_name": "public.users",
      "include_constraints": true
    }
  },
  "id": 1
}
```

#### Response Format

```json
{
  "jsonrpc": "2.0",
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Table: public.users\n\nColumns:\n- id: integer (PK, NOT NULL)\n- email: varchar(255) (NOT NULL, UNIQUE)\n..."
      }
    ]
  },
  "id": 1
}
```

---

## Implementation Guide

### Directory Structure

```
lazydb-agent/
├── cmd/
│   └── lazydb-agent/
│       └── main.go              # Entry point (or main.py)
├── internal/
│   ├── agent/
│   │   ├── agent.go             # Core agent logic
│   │   ├── router.go            # Task routing
│   │   └── formatter.go         # Response formatting
│   ├── llm/
│   │   ├── client.go            # LLM client interface
│   │   ├── anthropic.go         # Claude Haiku implementation
│   │   ├── openai.go            # GPT-4o-mini implementation
│   │   └── ollama.go            # Local model implementation
│   ├── mcp/
│   │   ├── client.go            # MCP client
│   │   ├── tools.go             # Tool definitions
│   │   └── protocol.go          # JSON-RPC protocol
│   └── prompts/
│       ├── system.go            # System prompts
│       ├── sql.go               # SQL-specific prompts
│       └── analysis.go          # Analysis prompts
├── configs/
│   └── config.yaml              # Agent configuration
├── go.mod
└── README.md
```

### Step-by-Step Implementation

#### Step 1: Create Entry Point

```go
// cmd/lazydb-agent/main.go
package main

import (
    "flag"
    "fmt"
    "os"

    "lazydb-agent/internal/agent"
)

func main() {
    // Parse flags
    task := flag.String("task", "", "User task/question")
    query := flag.String("query", "", "Current SQL query in editor")
    connection := flag.String("connection", "", "Active connection name")
    useMCP := flag.Bool("use-mcp", true, "Use MCP for schema discovery")
    flag.Parse()

    if *task == "" {
        fmt.Fprintln(os.Stderr, "Error: --task is required")
        os.Exit(1)
    }

    // Create and run agent
    a := agent.New(agent.Config{
        Connection: *connection,
        UseMCP:     *useMCP,
    })

    response, err := a.Process(*task, *query)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    // Output response to stdout (LazyDB reads this)
    fmt.Print(response)
}
```

#### Step 2: Implement Agent Core

```go
// internal/agent/agent.go
package agent

import (
    "context"
    "lazydb-agent/internal/llm"
    "lazydb-agent/internal/mcp"
    "lazydb-agent/internal/prompts"
)

type Agent struct {
    config    Config
    llmClient llm.Client
    mcpClient *mcp.Client
}

type Config struct {
    Connection string
    UseMCP     bool
    LLMModel   string // "haiku", "gpt-4o-mini", "ollama"
}

func New(cfg Config) *Agent {
    return &Agent{
        config:    cfg,
        llmClient: llm.NewClient(cfg.LLMModel),
        mcpClient: mcp.NewClient(),
    }
}

func (a *Agent) Process(task, currentQuery string) (string, error) {
    ctx := context.Background()

    // Step 1: Analyze task to determine required tools
    tools := a.analyzeTask(task)

    // Step 2: Gather context via MCP tools
    schemaContext, err := a.gatherContext(ctx, tools)
    if err != nil {
        return "", err
    }

    // Step 3: Build prompt with context
    prompt := prompts.BuildPrompt(task, currentQuery, schemaContext)

    // Step 4: Call LLM
    response, err := a.llmClient.Complete(ctx, prompt)
    if err != nil {
        return "", err
    }

    // Step 5: Format response
    return a.formatResponse(response), nil
}

func (a *Agent) analyzeTask(task string) []string {
    // Simple keyword-based routing (enhance with LLM for complex cases)
    tools := []string{}

    // Always get basic schema info
    tools = append(tools, "list_all_tables")

    // Add tools based on task keywords
    if containsAny(task, []string{"column", "schema", "structure", "field"}) {
        tools = append(tools, "get_table_schema")
    }
    if containsAny(task, []string{"slow", "performance", "optimize", "explain"}) {
        tools = append(tools, "explain_query", "get_table_indexes")
    }
    if containsAny(task, []string{"sample", "example", "data", "preview"}) {
        tools = append(tools, "get_sample_data")
    }
    if containsAny(task, []string{"relationship", "foreign", "join", "related"}) {
        tools = append(tools, "get_table_references")
    }

    return tools
}

func (a *Agent) gatherContext(ctx context.Context, tools []string) (string, error) {
    var context strings.Builder

    for _, tool := range tools {
        result, err := a.mcpClient.CallTool(ctx, tool, nil)
        if err != nil {
            continue // Non-fatal, try other tools
        }
        context.WriteString(result)
        context.WriteString("\n\n")
    }

    return context.String(), nil
}
```

#### Step 3: Implement MCP Client

```go
// internal/mcp/client.go
package mcp

import (
    "bufio"
    "context"
    "encoding/json"
    "fmt"
    "os/exec"
)

type Client struct {
    serverPath string
    connection string
}

func NewClient() *Client {
    return &Client{
        serverPath: "lazydb-mcp", // Assumes in PATH
    }
}

type Request struct {
    JSONRPC string      `json:"jsonrpc"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params"`
    ID      int         `json:"id"`
}

type Response struct {
    JSONRPC string `json:"jsonrpc"`
    Result  struct {
        Content []struct {
            Type string `json:"type"`
            Text string `json:"text"`
        } `json:"content"`
    } `json:"result"`
    Error *struct {
        Code    int    `json:"code"`
        Message string `json:"message"`
    } `json:"error"`
    ID int `json:"id"`
}

func (c *Client) CallTool(ctx context.Context, name string, args map[string]interface{}) (string, error) {
    // Start MCP server process
    cmd := exec.CommandContext(ctx, c.serverPath)
    stdin, _ := cmd.StdinPipe()
    stdout, _ := cmd.StdoutPipe()

    if err := cmd.Start(); err != nil {
        return "", fmt.Errorf("failed to start MCP server: %w", err)
    }
    defer cmd.Process.Kill()

    // Send initialization
    initReq := Request{
        JSONRPC: "2.0",
        Method:  "initialize",
        Params: map[string]interface{}{
            "protocolVersion": "2024-11-05",
            "capabilities":    map[string]interface{}{},
            "clientInfo": map[string]string{
                "name":    "lazydb-agent",
                "version": "1.0.0",
            },
        },
        ID: 1,
    }
    json.NewEncoder(stdin).Encode(initReq)

    // Read init response
    scanner := bufio.NewScanner(stdout)
    scanner.Scan() // Skip init response

    // Send tool call
    toolReq := Request{
        JSONRPC: "2.0",
        Method:  "tools/call",
        Params: map[string]interface{}{
            "name":      name,
            "arguments": args,
        },
        ID: 2,
    }
    json.NewEncoder(stdin).Encode(toolReq)

    // Read tool response
    scanner.Scan()
    var resp Response
    if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
        return "", err
    }

    if resp.Error != nil {
        return "", fmt.Errorf("MCP error: %s", resp.Error.Message)
    }

    if len(resp.Result.Content) > 0 {
        return resp.Result.Content[0].Text, nil
    }

    return "", nil
}
```

#### Step 4: Implement LLM Client

```go
// internal/llm/client.go
package llm

import "context"

type Client interface {
    Complete(ctx context.Context, prompt string) (string, error)
}

func NewClient(model string) Client {
    switch model {
    case "haiku":
        return NewAnthropicClient("claude-3-haiku-20240307")
    case "gpt-4o-mini":
        return NewOpenAIClient("gpt-4o-mini")
    case "ollama":
        return NewOllamaClient("llama3.1:8b")
    default:
        return NewAnthropicClient("claude-3-haiku-20240307")
    }
}
```

```go
// internal/llm/anthropic.go
package llm

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
)

type AnthropicClient struct {
    model  string
    apiKey string
}

func NewAnthropicClient(model string) *AnthropicClient {
    return &AnthropicClient{
        model:  model,
        apiKey: os.Getenv("ANTHROPIC_API_KEY"),
    }
}

func (c *AnthropicClient) Complete(ctx context.Context, prompt string) (string, error) {
    reqBody := map[string]interface{}{
        "model":      c.model,
        "max_tokens": 4096,
        "messages": []map[string]string{
            {"role": "user", "content": prompt},
        },
    }

    body, _ := json.Marshal(reqBody)
    req, _ := http.NewRequestWithContext(ctx, "POST",
        "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
    req.Header.Set("x-api-key", c.apiKey)
    req.Header.Set("anthropic-version", "2023-06-01")
    req.Header.Set("content-type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result struct {
        Content []struct {
            Text string `json:"text"`
        } `json:"content"`
    }
    json.NewDecoder(resp.Body).Decode(&result)

    if len(result.Content) > 0 {
        return result.Content[0].Text, nil
    }
    return "", fmt.Errorf("empty response")
}
```

#### Step 5: Implement Prompts

```go
// internal/prompts/system.go
package prompts

import "fmt"

const SystemPrompt = `You are a PostgreSQL database expert assistant integrated with LazyDB.
Your role is to help developers write efficient SQL queries, understand database schemas,
and optimize database performance.

Guidelines:
- Always use proper SQL formatting with clear indentation
- Consider query performance and suggest indexes when relevant
- Explain your reasoning briefly but clearly
- Use the schema context provided to write accurate queries
- Prefer standard PostgreSQL syntax

Response Format:
- Start with a brief analysis of the request
- Provide the SQL query in a code block
- Add explanations and recommendations as needed`

func BuildPrompt(task, currentQuery, schemaContext string) string {
    prompt := SystemPrompt + "\n\n"

    if schemaContext != "" {
        prompt += fmt.Sprintf("## Database Schema Context\n\n%s\n\n", schemaContext)
    }

    if currentQuery != "" {
        prompt += fmt.Sprintf("## Current Query in Editor\n\n```sql\n%s\n```\n\n", currentQuery)
    }

    prompt += fmt.Sprintf("## User Request\n\n%s", task)

    return prompt
}
```

---

## Care Points

### 1. Response Format Compatibility

**Critical**: Your agent's output must be parseable by LazyDB's `response_parser.go`.

```go
// LazyDB expects these section types
type ResponseSection struct {
    Type    string  // "code", "text", "list", "header", "query"
    Content string
    Number  int
    Title   string
}
```

**DO**:
```markdown
## Analysis
Your analysis here...

## SQL Query
```sql
SELECT * FROM users;
```
```

**DON'T**:
```json
{"analysis": "...", "query": "SELECT * FROM users"}
```

### 2. Connection State Handling

The active connection can change. Your agent should:

```go
// Read current connection from LazyDB's state file
func getActiveConnection() (*ConnectionConfig, error) {
    home, _ := os.UserHomeDir()
    data, err := os.ReadFile(filepath.Join(home, ".lazydb", "connections.json"))
    if err != nil {
        return nil, err
    }
    // Parse and return active connection
}
```

### 3. Error Handling

Return user-friendly errors that display well in the AI panel:

```go
func formatError(err error) string {
    return fmt.Sprintf(`## Error

Unable to complete the request:

> %s

### Suggestions
- Check your database connection
- Verify the table names exist
- Try a simpler query first
`, err.Error())
}
```

### 4. Timeout Management

LazyDB has default timeouts. Handle them gracefully:

```go
func (a *Agent) Process(task string) (string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Use ctx for all operations
    result, err := a.llmClient.Complete(ctx, prompt)
    if ctx.Err() == context.DeadlineExceeded {
        return "Request timed out. Try a simpler query or check your connection.", nil
    }
    return result, err
}
```

### 5. Token Efficiency

MCP tools can return large amounts of data. Be selective:

```go
// Instead of getting all tables, filter to relevant ones
func (a *Agent) gatherRelevantSchema(ctx context.Context, task string) string {
    // Extract table names mentioned in task
    tables := extractTableNames(task)

    if len(tables) == 0 {
        // Fall back to listing all, but with limits
        return a.mcpClient.CallTool(ctx, "list_all_tables", map[string]interface{}{
            "limit": 20,
        })
    }

    // Get schema only for mentioned tables
    var schemas []string
    for _, t := range tables {
        schema, _ := a.mcpClient.CallTool(ctx, "get_table_schema", map[string]interface{}{
            "table_name": t,
        })
        schemas = append(schemas, schema)
    }
    return strings.Join(schemas, "\n\n")
}
```

### 6. Security Considerations

Never execute arbitrary SQL from LLM output. LazyDB handles execution separately:

```go
// Your agent returns SQL as text, NOT executes it
// LazyDB user decides whether to run it

// DON'T do this:
func (a *Agent) Process(task string) string {
    query := llm.GenerateSQL(task)
    result := db.Execute(query)  // DANGEROUS!
    return result
}

// DO this:
func (a *Agent) Process(task string) string {
    query := llm.GenerateSQL(task)
    return formatSQLResponse(query)  // Return as text only
}
```

### 7. Streaming Support (Future)

For better UX, consider implementing streaming:

```go
// Future enhancement - streaming responses
type StreamingAgent interface {
    ProcessStream(task string, ch chan<- string) error
}

// LazyDB would need UI updates to support this
```

---

## Code Templates

### Python Implementation

If you prefer Python (better LLM library support):

```python
# lazydb_agent/main.py
import argparse
import json
import subprocess
import anthropic

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--task", required=True)
    parser.add_argument("--query", default="")
    parser.add_argument("--connection", default="")
    parser.add_argument("--use-mcp", action="store_true")
    args = parser.parse_args()

    agent = DBAgent(args.connection, args.use_mcp)
    response = agent.process(args.task, args.query)
    print(response)

class DBAgent:
    def __init__(self, connection: str, use_mcp: bool):
        self.connection = connection
        self.use_mcp = use_mcp
        self.client = anthropic.Anthropic()

    def process(self, task: str, query: str) -> str:
        # Gather schema context
        schema_ctx = self.gather_context(task)

        # Build prompt
        prompt = self.build_prompt(task, query, schema_ctx)

        # Call LLM
        response = self.client.messages.create(
            model="claude-3-haiku-20240307",
            max_tokens=4096,
            messages=[{"role": "user", "content": prompt}]
        )

        return response.content[0].text

    def gather_context(self, task: str) -> str:
        if not self.use_mcp:
            return ""

        # Call MCP tools
        tools_needed = self.analyze_task(task)
        context_parts = []

        for tool in tools_needed:
            result = self.call_mcp_tool(tool, {})
            if result:
                context_parts.append(result)

        return "\n\n".join(context_parts)

    def call_mcp_tool(self, name: str, args: dict) -> str:
        request = {
            "jsonrpc": "2.0",
            "method": "tools/call",
            "params": {"name": name, "arguments": args},
            "id": 1
        }

        proc = subprocess.Popen(
            ["lazydb-mcp"],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE
        )

        # Send init + tool call
        init_req = {
            "jsonrpc": "2.0",
            "method": "initialize",
            "params": {"protocolVersion": "2024-11-05", "capabilities": {}},
            "id": 0
        }

        proc.stdin.write(json.dumps(init_req).encode() + b"\n")
        proc.stdin.write(json.dumps(request).encode() + b"\n")
        proc.stdin.flush()

        # Read responses
        proc.stdout.readline()  # Skip init response
        response = json.loads(proc.stdout.readline())
        proc.terminate()

        if "result" in response and response["result"]["content"]:
            return response["result"]["content"][0]["text"]
        return ""

if __name__ == "__main__":
    main()
```

### Configuration File

```yaml
# configs/config.yaml
agent:
  name: "lazydb-agent"
  version: "1.0.0"

llm:
  provider: "anthropic"  # anthropic, openai, ollama
  model: "claude-3-haiku-20240307"
  max_tokens: 4096
  temperature: 0.1

mcp:
  server_path: "lazydb-mcp"
  timeout_seconds: 30

prompts:
  system_prompt_path: "prompts/system.txt"

logging:
  level: "info"
  file: "~/.lazydb/agent.log"
```

---

## Testing

### Unit Tests

```go
// internal/agent/agent_test.go
func TestAnalyzeTask(t *testing.T) {
    agent := New(Config{})

    tests := []struct {
        task     string
        expected []string
    }{
        {
            task:     "show me the columns in users table",
            expected: []string{"list_all_tables", "get_table_schema"},
        },
        {
            task:     "why is this query slow",
            expected: []string{"list_all_tables", "explain_query", "get_table_indexes"},
        },
    }

    for _, tt := range tests {
        tools := agent.analyzeTask(tt.task)
        assert.ElementsMatch(t, tt.expected, tools)
    }
}
```

### Integration Tests

```bash
# Test agent end-to-end
./lazydb-agent \
  --task "List all tables" \
  --connection "test-db" \
  --use-mcp

# Expected output:
## Database Tables

Based on the schema, here are all tables:

| Schema | Table | Rows |
|--------|-------|------|
| public | users | 1000 |
| public | orders | 5000 |
...
```

### Mock MCP Server

```go
// For testing without real database
type MockMCPClient struct {
    responses map[string]string
}

func (m *MockMCPClient) CallTool(ctx context.Context, name string, args map[string]interface{}) (string, error) {
    if resp, ok := m.responses[name]; ok {
        return resp, nil
    }
    return "", fmt.Errorf("unknown tool: %s", name)
}
```

---

## Deployment

### Building

```bash
# Go
cd lazydb-agent
go build -o lazydb-agent ./cmd/lazydb-agent

# Python
pip install -r requirements.txt
pyinstaller --onefile lazydb_agent/main.py -n lazydb-agent
```

### Installation

```bash
# Add to PATH
cp lazydb-agent /usr/local/bin/

# Or symlink
ln -s $(pwd)/lazydb-agent /usr/local/bin/lazydb-agent
```

### Environment Variables

```bash
# Required for cloud LLMs
export ANTHROPIC_API_KEY="sk-ant-..."
# or
export OPENAI_API_KEY="sk-..."

# Optional
export LAZYDB_AGENT_MODEL="haiku"
export LAZYDB_AGENT_TIMEOUT="30"
```

### LazyDB Integration

Add your agent as a CLI provider in LazyDB:

```go
// internal/ai/providers/lazydb_agent.go
package providers

type LazyDBAgentProvider struct{}

func (p *LazyDBAgentProvider) Name() string {
    return "lazydb-agent"
}

func (p *LazyDBAgentProvider) IsAvailable() bool {
    _, err := exec.LookPath("lazydb-agent")
    return err == nil
}

func (p *LazyDBAgentProvider) BuildCommand(ctx context.Context,
    schemaCtx *SchemaContext, query string, task string) (*exec.Cmd, error) {

    args := []string{
        "--task", task,
        "--query", query,
        "--connection", schemaCtx.ConnectionName,
    }

    if schemaCtx.UseMCP {
        args = append(args, "--use-mcp")
    }

    return exec.CommandContext(ctx, "lazydb-agent", args...), nil
}

func (p *LazyDBAgentProvider) ParseResponse(output string) string {
    return output // Already formatted correctly
}
```

---

## Appendix

### LLM Model Comparison

| Model | Speed | Quality | Cost | Best For |
|-------|-------|---------|------|----------|
| Claude 3 Haiku | ~1-2s | Good | $0.25/1M | Default choice |
| GPT-4o-mini | ~1-2s | Good | $0.15/1M | Cost-sensitive |
| Llama 3.1 8B (Ollama) | ~2-5s | Moderate | Free | Offline/Privacy |
| Claude 3.5 Sonnet | ~3-5s | Excellent | $3/1M | Complex analysis |

### Useful Resources

- [MCP Protocol Spec](https://modelcontextprotocol.io/)
- [LazyDB Repository](https://github.com/your-repo/lazydb)
- [Anthropic API Docs](https://docs.anthropic.com/)
- [pgx PostgreSQL Driver](https://github.com/jackc/pgx)

---

## Contributing

1. Fork the repository
2. Create feature branch: `git checkout -b feature/my-feature`
3. Implement with tests
4. Submit PR with description

## License

MIT License - See LICENSE file
