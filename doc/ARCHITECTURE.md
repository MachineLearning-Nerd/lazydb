# 🏗️ LazyDB Architecture

## System Architecture

### High-Level Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                         LazyDB TUI                                │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │                    Bubbletea Application                    │  │
│  │  ┌──────────────┬──────────────────┬──────────────────┐   │  │
│  │  │ Connections  │  Query Editor    │    Results       │   │  │
│  │  │   Panel      │     Panel        │     Panel        │   │  │
│  │  └──────────────┴──────────────────┴──────────────────┘   │  │
│  │                                                              │  │
│  │  ┌────────────────────────────────────────────────────────┐│  │
│  │  │                 Status Bar                              ││  │
│  │  └────────────────────────────────────────────────────────┘│  │
│  └────────────────────────────────────────────────────────────┘  │
│                                │                                  │
│                                ▼                                  │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │                    Application State                        │  │
│  │  • Active Connection    • Query History                     │  │
│  │  • Current Query        • Schema Cache                      │  │
│  │  • Results Data         • UI State                          │  │
│  └────────────────────────────────────────────────────────────┘  │
└────────────────────────┬─────────────────┬───────────────────────┘
                         │                 │
                         ▼                 ▼
        ┌────────────────────────┐  ┌──────────────────┐
        │  Database Abstraction  │  │  Neovim Spawner  │
        │   ┌────────────────┐   │  │  (External)      │
        │   │   PostgreSQL   │   │  └──────────────────┘
        │   │     MySQL      │   │
        │   │     SQLite     │   │
        │   └────────────────┘   │
        └────────────────────────┘
                 │
                 ▼
        ┌────────────────────────┐
        │    Actual Databases    │
        │  (Local/Remote)        │
        └────────────────────────┘
```

---

## Component Architecture

### 1. UI Layer (Bubbletea)

#### Bubbletea Model-View-Update Pattern

```go
type Model struct {
    // State
    activePanel    PanelType           // connections, editor, results
    connections    *ConnectionsPanel
    editor         *EditorPanel
    results        *ResultsPanel
    statusBar      *StatusBar

    // Application state
    currentConn    *db.Connection
    currentQuery   string
    queryResults   *db.ResultSet

    // UI state
    width, height  int
    helpVisible    bool
}

func (m Model) Init() tea.Cmd {
    return tea.Batch(
        loadConnections(),
        loadHistory(),
        tick(),
    )
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return m.handleKeyPress(msg)
    case tea.WindowSizeMsg:
        return m.handleResize(msg)
    case queryResultMsg:
        return m.handleQueryResult(msg)
    }
    return m, nil
}

func (m Model) View() string {
    return lipgloss.JoinVertical(
        lipgloss.Left,
        m.renderHeader(),
        m.renderMainContent(),
        m.renderStatusBar(),
    )
}
```

#### Panel Architecture

**BasePanel Interface**:
```go
type Panel interface {
    Init() tea.Cmd
    Update(msg tea.Msg) (Panel, tea.Cmd)
    View() string
    Focus()
    Blur()
    HandleKey(key tea.KeyMsg) (Panel, tea.Cmd)
}
```

**ConnectionsPanel**:
```go
type ConnectionsPanel struct {
    tree       *Tree              // Hierarchical connection list
    connections []db.Connection
    selected   int
    expanded   map[string]bool    // Which groups are expanded
    focused    bool
}
```

**EditorPanel**:
```go
type EditorPanel struct {
    content    string             // Current query text
    cursor     int
    nvimActive bool              // Is Neovim currently open?
    history    []string           // Recent queries
    focused    bool
}
```

**ResultsPanel**:
```go
type ResultsPanel struct {
    results    *db.ResultSet
    table      *Table             // Rendered table widget
    offset     int                // Scroll offset
    selected   struct{ row, col int }
    focused    bool
}
```

### 2. Database Abstraction Layer

#### Database Interface

```go
type Database interface {
    Connect(config ConnectionConfig) error
    Disconnect() error
    Execute(query string) (*ResultSet, error)
    ExecuteWithParams(query string, params []interface{}) (*ResultSet, error)
    ListTables() ([]Table, error)
    ListViews() ([]View, error)
    DescribeTable(tableName string) (*TableSchema, error)
    GetVersion() (string, error)
    Ping() error
}
```

#### Connection Pool

```go
type ConnectionPool struct {
    mu          sync.RWMutex
    connections map[string]*sql.DB
    configs     map[string]ConnectionConfig
    active      string  // Current active connection ID
}

func (cp *ConnectionPool) Get(id string) (*sql.DB, error)
func (cp *ConnectionPool) Set(id string, db *sql.DB, config ConnectionConfig)
func (cp *ConnectionPool) Switch(id string) error
func (cp *ConnectionPool) Close(id string) error
func (cp *ConnectionPool) CloseAll() error
```

#### Query Executor

```go
type QueryExecutor struct {
    db      Database
    timeout time.Duration
    cancel  context.CancelFunc
}

func (qe *QueryExecutor) Execute(query string) (*ResultSet, error) {
    ctx, cancel := context.WithTimeout(context.Background(), qe.timeout)
    defer cancel()

    // Execute query with cancellation support
    resultChan := make(chan *ResultSet)
    errChan := make(chan error)

    go func() {
        result, err := qe.db.ExecuteContext(ctx, query)
        if err != nil {
            errChan <- err
            return
        }
        resultChan <- result
    }()

    select {
    case result := <-resultChan:
        return result, nil
    case err := <-errChan:
        return nil, err
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

func (qe *QueryExecutor) Cancel() {
    if qe.cancel != nil {
        qe.cancel()
    }
}
```

### 3. Neovim Integration

#### Spawned Process Approach (Phase 1)

```go
type NvimEditor struct {
    executable string        // Path to nvim binary
    tempDir    string        // Directory for temp files
    config     NvimConfig    // Editor configuration
}

func (ne *NvimEditor) Edit(initialContent string) (string, error) {
    // 1. Create temporary file
    tmpFile, err := os.CreateTemp(ne.tempDir, "lazydb-query-*.sql")
    if err != nil {
        return "", err
    }
    defer os.Remove(tmpFile.Name())

    // 2. Write initial content
    if _, err := tmpFile.WriteString(initialContent); err != nil {
        return "", err
    }
    tmpFile.Close()

    // 3. Build Neovim command
    args := []string{
        "-c", "set filetype=sql",           // Set SQL filetype
        "-c", "set noswapfile",             // Disable swap file
        "-c", "autocmd VimLeave * :wq",     // Auto-save on exit
        "--", tmpFile.Name(),
    }

    cmd := exec.Command(ne.executable, args...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    // 4. Run Neovim (blocks until user exits)
    if err := cmd.Run(); err != nil {
        return "", err
    }

    // 5. Read edited content
    content, err := os.ReadFile(tmpFile.Name())
    if err != nil {
        return "", err
    }

    return string(content), nil
}

func (ne *NvimEditor) IsAvailable() bool {
    _, err := exec.LookPath(ne.executable)
    return err == nil
}
```

#### Embedded RPC Approach (Phase 4 - Future)

```go
type EmbeddedNvim struct {
    process *exec.Cmd
    msgpack *msgpack.Conn
    session nvim.Session
}

func (en *EmbeddedNvim) Start() error {
    // Spawn Neovim with --embed flag
    en.process = exec.Command("nvim", "--embed", "--headless")

    // Setup msgpack-rpc communication
    stdin, _ := en.process.StdinPipe()
    stdout, _ := en.process.StdoutPipe()

    en.msgpack = msgpack.NewConn(stdin, stdout)

    return en.process.Start()
}

func (en *EmbeddedNvim) SetBuffer(content string) error {
    lines := strings.Split(content, "\n")
    return en.session.SetBufferLines(0, 0, -1, true, lines)
}

func (en *EmbeddedNvim) GetBuffer() (string, error) {
    lines, err := en.session.GetBufferLines(0, 0, -1, true)
    if err != nil {
        return "", err
    }
    return strings.Join(lines, "\n"), nil
}
```

### 4. Storage Layer

#### Configuration Storage

```go
type ConfigStorage struct {
    path string  // ~/.config/lazydb/config.toml
}

type Config struct {
    General struct {
        Theme        string `toml:"theme"`
        PageSize     int    `toml:"page_size"`
        QueryTimeout int    `toml:"query_timeout"`
        Editor       string `toml:"editor"`
    }
    UI struct {
        ConnectionsPanelWidth int  `toml:"connections_panel_width"`
        ResultsPanelWidth     int  `toml:"results_panel_width"`
        ShowLineNumbers       bool `toml:"show_line_numbers"`
    }
    Keybindings map[string]string `toml:"keybindings"`
}

func (cs *ConfigStorage) Load() (*Config, error)
func (cs *ConfigStorage) Save(config *Config) error
```

#### Connection Storage (Encrypted)

```go
type ConnectionStorage struct {
    path    string  // ~/.config/lazydb/connections.toml
    keyring keyring.Keyring  // System keyring for passwords
}

type ConnectionConfig struct {
    Name        string `toml:"name"`
    Type        string `toml:"type"`  // postgres, mysql, sqlite
    Host        string `toml:"host"`
    Port        int    `toml:"port"`
    Database    string `toml:"database"`
    Username    string `toml:"username"`
    Password    string `toml:"-"`  // Not stored in file
    SSLMode     string `toml:"ssl_mode"`
    Environment string `toml:"environment"`  // dev, staging, prod
}

func (cs *ConnectionStorage) Save(conn ConnectionConfig) error {
    // Save password to system keyring
    err := cs.keyring.Set("lazydb", conn.Name, conn.Password)
    if err != nil {
        return err
    }

    // Save connection (without password) to TOML file
    conn.Password = ""
    return cs.saveToFile(conn)
}

func (cs *ConnectionStorage) Load(name string) (*ConnectionConfig, error) {
    // Load from TOML file
    conn, err := cs.loadFromFile(name)
    if err != nil {
        return nil, err
    }

    // Retrieve password from keyring
    password, err := cs.keyring.Get("lazydb", name)
    if err != nil {
        return nil, err
    }
    conn.Password = password

    return conn, nil
}
```

#### Query History Storage

```go
type HistoryStorage struct {
    db *sql.DB  // SQLite database
}

type HistoryEntry struct {
    ID          int64
    Query       string
    Connection  string
    Database    string
    ExecutedAt  time.Time
    Duration    time.Duration
    RowsReturned int
    Success     bool
    Error       string
}

func (hs *HistoryStorage) Add(entry HistoryEntry) error
func (hs *HistoryStorage) Search(query string, limit int) ([]HistoryEntry, error)
func (hs *HistoryStorage) GetRecent(limit int) ([]HistoryEntry, error)
func (hs *HistoryStorage) Clear() error
```

---

## Data Flow

### Query Execution Flow

```
User presses Ctrl-R
    ↓
EditorPanel.HandleKey()
    ↓
Sends queryExecuteMsg to Update()
    ↓
Model.Update() receives queryExecuteMsg
    ↓
Creates QueryExecutor with current connection
    ↓
Executor.Execute() (async in goroutine)
    ↓
Query sent to database driver
    ↓
Results streamed back
    ↓
Send queryResultMsg to Update()
    ↓
Model.Update() receives queryResultMsg
    ↓
Updates ResultsPanel with data
    ↓
Saves to history storage
    ↓
View() re-renders with new results
```

### Neovim Edit Flow

```
User presses Ctrl-E in editor panel
    ↓
EditorPanel.HandleKey()
    ↓
Sends openNvimMsg to Update()
    ↓
Model.Update() suspends Bubbletea
    ↓
NvimEditor.Edit() called with current query
    ↓
Temporary SQL file created
    ↓
Neovim process spawned with file
    ↓
User edits in full-screen Neovim
    ↓
User saves and quits (:wq)
    ↓
Neovim exits, LazyDB resumes
    ↓
Edited content read from temp file
    ↓
EditorPanel updated with new content
    ↓
View() re-renders with edited query
```

---

## Performance Considerations

### 1. Async Query Execution

- Queries run in goroutines to avoid blocking UI
- Cancellable via context.Context
- Results streamed for large datasets
- Progress indicator shown during execution

### 2. Result Set Pagination

```go
type ResultSet struct {
    Columns  []string
    Rows     [][]interface{}
    Total    int           // Total row count (if known)
    PageSize int
    Page     int
}

func (rs *ResultSet) NextPage() error
func (rs *ResultSet) PrevPage() error
func (rs *ResultSet) GoToPage(page int) error
```

### 3. Schema Caching

```go
type SchemaCache struct {
    mu        sync.RWMutex
    cache     map[string]*Schema  // connection ID -> schema
    ttl       time.Duration
    lastFetch map[string]time.Time
}

func (sc *SchemaCache) Get(connID string) (*Schema, error) {
    sc.mu.RLock()
    defer sc.mu.RUnlock()

    // Check if cache is still valid
    if lastFetch, ok := sc.lastFetch[connID]; ok {
        if time.Since(lastFetch) < sc.ttl {
            return sc.cache[connID], nil
        }
    }

    return nil, ErrCacheMiss
}

func (sc *SchemaCache) Set(connID string, schema *Schema) {
    sc.mu.Lock()
    defer sc.mu.Unlock()

    sc.cache[connID] = schema
    sc.lastFetch[connID] = time.Now()
}
```

### 4. Connection Pooling

- Keep connections alive across queries
- Reuse connections for same database
- Max connection pool size configurable
- Idle connection timeout

---

## Error Handling Strategy

### Error Categories

1. **Connection Errors**: Network issues, auth failures
2. **Query Errors**: Syntax errors, constraint violations
3. **System Errors**: File I/O, permission issues
4. **User Errors**: Invalid input, cancelled operations

### Error Handling Pattern

```go
type LazyDBError struct {
    Category  ErrorCategory
    Message   string
    Cause     error
    Retryable bool
    Hint      string  // Suggestion for user
}

func (e *LazyDBError) Error() string {
    if e.Hint != "" {
        return fmt.Sprintf("%s\nHint: %s", e.Message, e.Hint)
    }
    return e.Message
}

// Example usage
if err != nil {
    return &LazyDBError{
        Category:  CategoryQuery,
        Message:   "Failed to execute query",
        Cause:     err,
        Retryable: false,
        Hint:      "Check your SQL syntax near 'SELCT'",
    }
}
```

---

## Testing Strategy

### Unit Tests

```go
// Database abstraction tests
func TestPostgresConnect(t *testing.T)
func TestQueryExecution(t *testing.T)
func TestResultPagination(t *testing.T)

// UI component tests
func TestConnectionsPanelNavigation(t *testing.T)
func TestTableRendering(t *testing.T)
func TestKeybindingHandling(t *testing.T)
```

### Integration Tests

```go
// Database integration (with Docker)
func TestPostgresIntegration(t *testing.T) {
    // Start PostgreSQL container
    db := setupTestDB(t)
    defer db.Close()

    // Test queries
    result, err := db.Execute("SELECT 1")
    require.NoError(t, err)
    assert.Equal(t, 1, result.RowCount())
}
```

### E2E Tests

```go
// Full workflow tests
func TestCompleteQueryWorkflow(t *testing.T) {
    app := NewTestApp(t)
    defer app.Cleanup()

    // Simulate user actions
    app.SendKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
    app.SendKey(tea.KeyMsg{Type: tea.KeyEnter})
    app.TypeQuery("SELECT * FROM users")
    app.SendKey(tea.KeyMsg{Type: tea.KeyCtrlR})

    // Verify results
    assert.Eventually(t, func() bool {
        return app.ResultsPanel.RowCount() > 0
    }, time.Second, time.Millisecond*100)
}
```

---

## Security Considerations

### 1. Password Storage
- **Never** store passwords in plain text
- Use system keyring (keychain on macOS, Secret Service on Linux)
- Encrypt config files if keyring unavailable
- Support for SSH key-based auth

### 2. SQL Injection Prevention
- Use prepared statements with parameter binding
- Warn user about unsafe queries
- No automatic query modification

### 3. Audit Logging (Optional)
- Log all executed queries with timestamp
- Store connection used
- Optional email alerts for production queries
- Redact sensitive data from logs

---

## Future Enhancements

### Plugin System

```go
type Plugin interface {
    Name() string
    Init(app *App) error
    Hooks() []Hook
}

type Hook interface {
    OnQueryExecute(query string) error
    OnConnectionChange(conn *Connection) error
    OnResultsReady(results *ResultSet) error
}
```

### Visual Query Builder

```
┌─ Visual Query Builder ─────────────────────┐
│                                             │
│  Tables:   [ users ▼ ]                     │
│  Columns:  [x] id  [x] name  [ ] email     │
│  Where:    name LIKE '%smith%'             │
│  Order:    name ASC                        │
│  Limit:    10                              │
│                                             │
│  Generated SQL:                            │
│  SELECT id, name FROM users                │
│  WHERE name LIKE '%smith%'                 │
│  ORDER BY name ASC                         │
│  LIMIT 10;                                 │
│                                             │
│  [Execute]  [Copy]  [Cancel]               │
└─────────────────────────────────────────────┘
```

### ER Diagram Viewer

- Parse schema relationships
- Render ASCII art ER diagram
- Interactive navigation

---

This architecture is designed to be:
- **Modular**: Easy to swap components
- **Testable**: Clear interfaces and separation of concerns
- **Performant**: Async operations, caching, streaming
- **Maintainable**: Clear patterns and documentation
- **Extensible**: Plugin system for future growth
