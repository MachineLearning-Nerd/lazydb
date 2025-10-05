# LazyDB MCP Server - Implementation Plan

**Document Version:** 1.0
**Date:** January 2025
**Status:** Planning & Initial Implementation

---

## Executive Summary

LazyDB MCP Server provides intelligent, context-aware database schema information to AI coding assistants through the Model Context Protocol (MCP). This server solves the "context explosion" problem by using smart multi-turn analysis to deliver only relevant schema information instead of dumping entire databases.

**Key Innovation:** Hybrid approach combining standard MCP protocol with intelligent context optimization, reducing token usage by 70%+ while supporting databases of any size.

---

## Problem Statement

### Current Limitations

1. **Blind Context Injection**
   - Current: Send first 50 tables with full schemas (~15K tokens)
   - Problem: Irrelevant tables waste context, large DBs lose important tables

2. **No Intelligence**
   - Can't analyze what user actually needs
   - Can't dynamically fetch relevant schema
   - No relationship discovery

3. **Scale Issues**
   - 50-table limit arbitrary
   - Databases with 500+ tables unusable
   - No way to include all tables

### Solution: Smart MCP Server

**Three-Tier Intelligence:**
- **Tier 1 (Basic Tools):** Standard MCP tools for direct schema access
- **Tier 2 (Smart Tools):** AI-powered context optimization
- **Tier 3 (Specialized):** Advanced features (samples, relationships, stats)

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        MCP Clients                               │
│  (Claude Code, Gemini CLI, Qwen Code, GitHub Copilot)          │
└───────────────────────────┬─────────────────────────────────────┘
                            │ MCP Protocol (stdio/SSE)
┌───────────────────────────▼─────────────────────────────────────┐
│                   LazyDB MCP Server                              │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │           Tool Layer (MCP Interface)                      │  │
│  │  Basic Tools    │  Smart Tools   │  Specialized Tools    │  │
│  └────────┬─────────────────┬────────────────┬──────────────┘  │
│           │                 │                │                  │
│  ┌────────▼─────────────────▼────────────────▼──────────────┐  │
│  │         Intelligence Layer (Multi-turn Logic)             │  │
│  │  • Question Parser  • Context Optimizer  • AI Analyzer   │  │
│  └────────┬──────────────────────────────────────────────────┘  │
│           │                                                      │
│  ┌────────▼──────────────────────────────────────────────────┐  │
│  │              Database Access Layer                         │  │
│  │  • Schema Reader  • Query Executor  • Metadata Cache      │  │
│  └────────┬──────────────────────────────────────────────────┘  │
└───────────┼──────────────────────────────────────────────────────┘
            │
┌───────────▼──────────────────────────────────────────────────────┐
│              PostgreSQL Database                                 │
└──────────────────────────────────────────────────────────────────┘
```

---

## MCP Protocol Integration

### Supported Clients (2025)

| Client | MCP Support | Configuration |
|--------|-------------|---------------|
| Claude Code | ✅ Full | `~/.claude/mcp_servers.json` |
| Gemini CLI | ✅ Full | `~/.gemini/settings.json` |
| Qwen Code | ✅ Full | Based on Gemini CLI |
| GitHub Copilot | ✅ Full | VS Code, JetBrains, etc. |
| Aider | ⚠️ Via wrapper | Can wrap Aider as MCP server |

### MCP Protocol Flow

```
1. Client → Server: initialize
   Server → Client: capabilities (tools, resources)

2. Client → Server: tools/list
   Server → Client: [list_all_tables, get_table_schema, smart_schema_context, ...]

3. User asks question in AI client
   Client → AI: User question + available tools

4. AI decides to call tool
   Client → Server: tools/call {name: "smart_schema_context", arguments: {...}}

5. Server executes tool (with internal intelligence)
   Server → Client: Optimized schema context

6. AI generates answer
   Client → User: SQL query / explanation
```

---

## Tool Specifications

### Tier 1: Basic Tools

#### 1. `list_all_tables`
**Purpose:** Get complete list of all tables in database
**Input:** None
**Output:** JSON map of schema → table names
**Use Case:** AI wants to see what tables exist

```json
{
  "public": ["users", "orders", "products"],
  "analytics": ["reports", "metrics"]
}
```

#### 2. `get_table_schema`
**Purpose:** Get detailed schema for specific table
**Input:**
- `table_name` (required): "schema.table" or "table"
- `include_constraints` (optional, default: true)

**Output:** JSON with columns, types, constraints

```json
{
  "table": "users",
  "schema": "public",
  "columns": [
    {"name": "id", "type": "SERIAL", "nullable": false, "primary_key": true},
    {"name": "email", "type": "VARCHAR(255)", "nullable": false, "unique": true}
  ],
  "foreign_keys": [...]
}
```

#### 3. `search_tables`
**Purpose:** Find tables matching pattern
**Input:**
- `pattern` (required): SQL LIKE pattern
- `schema` (optional): Filter by schema

**Output:** List of matching table names

#### 4. `get_related_tables`
**Purpose:** Find tables related via foreign keys
**Input:**
- `table_name` (required)
- `depth` (optional, default: 1): Relationship depth

**Output:** Tables that reference or are referenced by this table

#### 5. `get_sample_data`
**Purpose:** Get sample rows to understand data patterns
**Input:**
- `table_name` (required)
- `limit` (optional, default: 5, max: 10)

**Output:** JSON array of sample rows

### Tier 2: Smart Tools

#### 1. `smart_schema_context` ⭐ **PRIMARY TOOL**
**Purpose:** Get optimized database context based on user's question
**Input:**
- `question` (required): User's question or task
- `mode` (optional): "auto", "minimal", "full"
- `max_tables` (optional, default: 10)

**Intelligence Process:**
1. Analyze question using AI to extract:
   - Referenced table names
   - Keywords
   - Operation type (SELECT/INSERT/UPDATE/DELETE)
   - Whether JOINs needed
   - Whether aggregation needed
2. Fetch full schema for referenced tables
3. If joins needed, fetch related tables (minimal info)
4. Include lightweight list of all other tables for reference

**Output:** Multi-tier context

```json
{
  "analysis": {
    "referenced_tables": ["orders"],
    "operation_type": "SELECT",
    "needs_joins": true,
    "confidence": 0.95
  },
  "schemas": [
    {
      "table": "orders",
      "schema": "public",
      "columns": [...],
      "tier": "referenced"
    },
    {
      "table": "users",
      "schema": "public",
      "columns": [{"name": "id"}, {"name": "name"}],
      "tier": "related"
    }
  ],
  "all_tables": ["public.products", "public.categories", ...]
}
```

**Token Efficiency:**
- Traditional: 50 tables × 300 tokens = 15,000 tokens
- Smart: 3 tables (full) + 10 tables (minimal) + 200 names = 4,000 tokens
- **Savings: 73%**

---

## Implementation Phases

### Phase 1: MCP Protocol Foundation (Days 1-3)

**Deliverables:**
- MCP protocol types and handlers
- Server initialization and request handling
- Tool registry system
- stdio transport layer

**Files:**
- `internal/mcp/server/protocol.go`
- `internal/mcp/server/server.go`
- `internal/mcp/server/registry.go`
- `internal/mcp/server/transport.go`

**Acceptance Criteria:**
- Server responds to `initialize` method
- Server responds to `tools/list` method
- Server can execute tool calls
- Works with Claude Code

### Phase 2: Basic Tools Implementation (Days 4-5)

**Deliverables:**
- All 5 basic tools functional
- Database connection integration
- Error handling
- JSON response formatting

**Files:**
- `internal/mcp/tools/basic.go`
- `internal/mcp/tools/basic_test.go`

**Acceptance Criteria:**
- `list_all_tables` returns valid JSON
- `get_table_schema` fetches correct schema
- Tools work with test database
- Unit tests pass

### Phase 3: Smart Tools with Intelligence (Days 6-8)

**Deliverables:**
- Question analyzer (AI-powered)
- Context optimizer
- `smart_schema_context` tool
- Multi-turn logic

**Files:**
- `internal/mcp/tools/smart.go`
- `internal/mcp/intelligence/analyzer.go`
- `internal/mcp/intelligence/optimizer.go`
- `internal/mcp/intelligence/parser.go`

**Acceptance Criteria:**
- Analyzer correctly identifies tables 80%+ accuracy
- Optimizer reduces context by 70%+
- Fallback works without AI
- Smart tool completes in <2 seconds

### Phase 4: CLI & Configuration (Day 9)

**Deliverables:**
- Standalone MCP server binary
- Configuration file support
- Connection management
- Signal handling

**Files:**
- `cmd/lazydb-mcp/main.go`
- Configuration updates

**Acceptance Criteria:**
- Binary runs standalone
- Reads LazyDB config
- Connects to database
- Graceful shutdown

### Phase 5: Testing & Documentation (Day 10)

**Deliverables:**
- Unit tests for all components
- Integration tests
- Documentation
- Example configurations

**Files:**
- Test files
- `docs/mcp-server.md`
- Configuration examples

**Acceptance Criteria:**
- 80%+ code coverage
- Integration tests pass
- Works with 3+ AI clients
- Documentation complete

---

## Intelligence Layer Design

### Question Analyzer

**Purpose:** Extract intent and requirements from user questions

**Algorithm:**
1. **AI-Powered Analysis (Primary)**
   - Send question to AI with structured prompt
   - Extract: table names, keywords, operation type, needs
   - Return structured analysis with confidence score

2. **Fallback Analysis (When AI unavailable)**
   - Keyword extraction (users, orders, products, etc.)
   - Pattern matching for operations (SELECT/INSERT/UPDATE/DELETE)
   - Heuristic detection (JOINs, aggregations, date filters)

**Example Analysis:**

```go
type QuestionAnalysis struct {
    ReferencedTables  []string  // ["orders", "users"]
    Keywords          []string  // ["last", "month", "show"]
    OperationType     string    // "SELECT"
    NeedsJoins        bool      // true
    NeedsAggregation  bool      // false
    DateFilters       bool      // true
    Confidence        float64   // 0.95
}
```

### Context Optimizer

**Purpose:** Build optimized schema context based on analysis

**Strategy:**
1. **Tier 1 - Referenced Tables (Full Details)**
   - Parse analysis.ReferencedTables
   - Fetch complete schemas with all columns
   - Include constraints and indexes
   - ~200-500 tokens per table

2. **Tier 2 - Related Tables (Medium Details)**
   - If analysis.NeedsJoins == true
   - Query foreign key relationships
   - Include table name + key columns only
   - ~50-100 tokens per table

3. **Tier 3 - All Other Tables (Names Only)**
   - List remaining tables as reference
   - Format: "schema.table_name"
   - ~5-10 tokens per table
   - Enables AI to discover additional tables

**Caching Strategy:**
- Cache table schemas (TTL: 5 minutes)
- Cache relationship maps (TTL: 10 minutes)
- Cache analysis results (TTL: 1 minute)
- Cache key: hash(question + connection)

---

## Configuration

### LazyDB Config (`~/.lazydb/config.yml`)

```yaml
ai:
  enabled: true
  cli_tool: claude-cli
  inject_in_neovim: true

  # MCP Server Configuration
  mcp_enabled: true
  mcp_smart_tools: true          # Enable AI-powered optimization
  mcp_cache_enabled: true
  mcp_max_cache_size: 104857600  # 100MB
  mcp_ai_provider: claude        # claude, gemini, openai
```

### MCP Client Configurations

#### Claude Code
`~/.claude/mcp_servers.json`:
```json
{
  "lazydb": {
    "command": "/path/to/lazydb-mcp",
    "args": ["--connection", "current"],
    "env": {
      "ANTHROPIC_API_KEY": "${ANTHROPIC_API_KEY}"
    }
  }
}
```

#### Gemini CLI
`~/.gemini/settings.json`:
```json
{
  "mcp": {
    "servers": {
      "lazydb": {
        "command": "/path/to/lazydb-mcp",
        "args": ["--connection", "current"]
      }
    }
  }
}
```

#### Qwen Code
Same as Gemini CLI (based on same architecture)

---

## Usage Examples

### Example 1: Simple Query

**User:** "Show me all users who placed orders last month"

**Flow:**
```
1. AI calls: smart_schema_context(question="Show me all users who placed orders last month")

2. Server analyzes:
   - Referenced tables: ["users", "orders"]
   - Needs JOINs: true
   - Date filter: true

3. Server fetches:
   - users table (full schema)
   - orders table (full schema)
   - All table names (reference)

4. AI receives optimized context and generates:
   SELECT DISTINCT u.*
   FROM users u
   JOIN orders o ON u.id = o.user_id
   WHERE o.created_at >= DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month')
```

**Token Usage:** ~4,000 tokens (vs. 15,000 with traditional approach)

### Example 2: Exploration

**User:** "What tables are related to payments?"

**Flow:**
```
1. AI calls: search_tables(pattern="%payment%")
   Returns: ["payments", "payment_methods", "payment_history"]

2. AI calls: get_related_tables(table_name="payments")
   Returns: {references: ["orders"], referenced_by: ["refunds"]}

3. AI responds: "The payments table is related to:
   - orders (via foreign key)
   - refunds (referenced by)"
```

### Example 3: Data Understanding

**User:** "What kind of data is in the products table?"

**Flow:**
```
1. AI calls: get_table_schema(table_name="products")
   Returns: columns with types

2. AI calls: get_sample_data(table_name="products", limit=3)
   Returns: 3 sample rows

3. AI responds: "The products table contains:
   - Product details (name, description)
   - Pricing (price, currency)
   - Sample data shows electronics products..."
```

---

## Performance Metrics

### Success Criteria

**Context Efficiency:**
- ✅ 70%+ reduction in token usage
- ✅ Support databases with 500+ tables
- ✅ Response time <2 seconds

**Intelligence Accuracy:**
- ✅ 80%+ table identification accuracy
- ✅ 90%+ operation type detection
- ✅ Fallback mode maintains 60%+ accuracy

**Reliability:**
- ✅ 95%+ uptime
- ✅ Cache hit rate >60%
- ✅ Error rate <1%

**Compatibility:**
- ✅ Works with Claude Code
- ✅ Works with Gemini CLI
- ✅ Works with Qwen Code
- ✅ Works with GitHub Copilot

---

## Testing Strategy

### Unit Tests
- Protocol handlers
- Tool implementations
- Analyzer logic
- Optimizer logic
- Cache operations

### Integration Tests
- End-to-end MCP flows
- Multi-tool scenarios
- Error handling
- Performance benchmarks

### Client Testing
- Test with Claude Code
- Test with Gemini CLI
- Test with different databases
- Test with large schemas (100+ tables)

### Test Script Example
```bash
#!/bin/bash
# Test MCP server with sample requests

echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./lazydb-mcp

echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | ./lazydb-mcp

echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"smart_schema_context","arguments":{"question":"Show me all users"}}}' | ./lazydb-mcp
```

---

## Timeline & Milestones

### Week 1: Foundation
- **Days 1-3:** MCP protocol implementation
- **Days 4-5:** Basic tools
- **Milestone:** Basic MCP server works with Claude Code

### Week 2: Intelligence
- **Days 6-8:** Smart tools and AI integration
- **Day 9:** CLI and configuration
- **Day 10:** Testing and documentation
- **Milestone:** Smart context optimization delivers 70%+ savings

### Week 3: Optimization
- Performance tuning
- Cache optimization
- Bug fixes
- **Milestone:** Production-ready release

### Week 4: Enhancement
- Advanced tools
- Query history
- Usage analytics
- **Milestone:** v1.0 release

---

## Future Enhancements

### Phase 2 Features (Month 2)
1. **Query Optimization Tool**
   - Analyze query performance
   - Suggest indexes
   - Explain query plans

2. **Schema Migration Assistant**
   - Generate migration scripts
   - Detect schema changes
   - Version tracking

3. **Web UI**
   - MCP server dashboard
   - Usage analytics
   - Performance monitoring

### Phase 3 Features (Month 3)
1. **Multi-Database Support**
   - MySQL support
   - SQLite support
   - MongoDB support

2. **Collaborative Features**
   - Share schema contexts
   - Team knowledge base
   - Query templates

3. **Advanced AI**
   - Learn from usage patterns
   - Auto-detect priority tables
   - Predictive context loading

---

## Risk Mitigation

### Technical Risks

**Risk:** AI API failures
**Mitigation:** Fallback to keyword-based analysis

**Risk:** Large database performance
**Mitigation:** Aggressive caching, lazy loading

**Risk:** MCP protocol changes
**Mitigation:** Version compatibility layer

### Operational Risks

**Risk:** API cost (AI calls)
**Mitigation:** Cache results, batch requests

**Risk:** Database connection issues
**Mitigation:** Connection pooling, retry logic

**Risk:** Client compatibility
**Mitigation:** Extensive testing, version matrix

---

## Success Metrics

### Adoption Metrics
- Number of active MCP server installations
- Daily active users
- Tools called per session

### Performance Metrics
- Average response time
- Cache hit rate
- Token reduction percentage

### Quality Metrics
- Table identification accuracy
- User satisfaction score
- Bug report rate

---

## Conclusion

The LazyDB MCP Server represents a significant advancement in database-aware AI tooling. By combining standard MCP protocol with intelligent context optimization, it solves the fundamental problem of context explosion while maintaining compatibility with all major AI coding assistants.

**Key Innovations:**
1. Smart multi-turn analysis for context optimization
2. Three-tier schema delivery (full/medium/names)
3. AI-powered question understanding with fallback
4. Universal compatibility via MCP standard

**Expected Impact:**
- 70%+ reduction in context token usage
- Support for databases of any size
- Better AI responses through relevant context
- Seamless integration with existing workflows

---

**Document Status:** ✅ Planning Complete, Ready for Implementation
**Next Step:** Begin Phase 1 - MCP Protocol Foundation
