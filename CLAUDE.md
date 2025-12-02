# LazyDB - Claude Project Memory

## Critical Build Information

### ⚠️ CORRECT BUILD COMMAND
```bash
go build -o lazydb ./cmd/lazydb
```

**IMPORTANT**:
- Binary output: `./lazydb` (project root)
- **NOT** `./bin/lazydb` or any other location
- Run with: `./lazydb`

## Project Structure

```
LazyDB/
├── cmd/lazydb/main.go          # Main application entry point
├── internal/
│   ├── config/                 # Configuration system (YAML-based)
│   │   ├── config.go          # Type definitions
│   │   ├── defaults.go        # Default values
│   │   ├── loader.go          # Load/save logic
│   │   └── validator.go       # Validation & duplicate key checking
│   ├── db/                     # Database layer
│   │   ├── connection.go      # Connection management
│   │   ├── query.go           # Query execution
│   │   └── validator.go       # SQL validation (pg_query_go)
│   ├── editor/                 # Neovim integration
│   ├── storage/                # Persistence layer
│   └── ui/                     # TUI components
│       ├── components/         # Reusable UI components
│       │   └── highlighter.go # SQL syntax highlighting (Chroma v2)
│       └── panels/             # Main panels
│           ├── connections.go
│           ├── editor.go       # Vim-style modal editor
│           └── results.go
├── lazydb                      # Built binary (gitignored)
└── ~/.lazydb/                  # User data directory
    ├── config.yml             # User configuration
    ├── connections.json       # Encrypted connections
    ├── debug.log              # Debug output
    └── queries/               # Query history by environment
```

## Configuration System

### Location
- Primary: `~/.lazydb/config.yml`
- XDG: `$XDG_CONFIG_HOME/lazydb/config.yml`
- Auto-generated on first run

### Structure
```yaml
version: 1
keybindings:
  resize:                      # Panel resizing keys
    grow_editor_left: "="
    shrink_editor_left: "-"
    grow_editor_right: "["
    shrink_editor_right: "]"
  layout:                      # Layout preset keys
    preset_mode: "L"
    preset_default: "1"
    preset_editor: "2"
    preset_results: "3"
    preset_balanced: "4"
  navigation:                  # Panel focus keys
    focus_connections: "1"
    focus_editor: "2"
    focus_results: "3"
    next_panel: "tab"
    prev_panel: "shift+tab"
  global:                      # Global actions
    help: "?"
    quit: "ctrl+q"             # CHANGED from "q" - now Ctrl+Q to quit app
    execute_query: "ctrl+r"
    save_query: "f2"
    open_neovim: "ctrl+e"
  connections:                 # Connections panel
    add: "a"
    edit: "e"
    delete: "d"
    connect: "enter"
    schema_explorer: "s"
  schema:                      # Schema explorer
    navigate_down: "j"
    navigate_up: "k"
    expand: "enter"
    preview: "p"
    search: "/"
    refresh: "r"
    exit: "esc"
ui:
  default_layout:              # Panel width ratios (must sum to 100)
    connections: 20
    editor: 40
    results: 40
  resize_increment: 5          # Resize step size
  min_panel_width: 15          # Minimum panel width %
  max_panel_width: 70          # Maximum panel width %
theme:
  name: "monokai"
  syntax_highlighting: true
  sql_linting: true
```

## Debug Logging System

### Viewing Logs
```bash
# Terminal 1: Watch logs in real-time
tail -f ~/.lazydb/debug.log

# Terminal 2: Run LazyDB
./lazydb
```

### Log Markers
- `[STARTUP]` - App initialization, config loading
- `[LOG1]` - Key press entry point in main.go
- `[LOG3]` - Keybinding comparison checks (shows true/false matches)
- `[LOG4]` - Before panel.Update() call
- `[LOG5]` - Inside editor panel (shows INSERT/NORMAL mode)

### Log Output Example
```
[STARTUP] Debug logging initialized at /Users/dineshjinjala/.lazydb/debug.log
[STARTUP] Config loaded - Resize keys: grow='=' shrink='-' growR='[' shrinkR=']'
[LOG1] Key='=' Panel=Editor Dialog=false
[LOG3] Key='=' | Cfg: grow='='(true) shrink='-'(false) growR='['(false) shrinkR=']'(false)
[LOG4] About to call Editor.Update() with key='='
[LOG5] EditorPanel received key='=' mode=INSERT
[LOG5] INSERT mode - passing key='=' to textarea (will be typed)
```

## Recent Issues Fixed

### Navigation Keys (1, 2, 3) in Editor INSERT Mode ✅ FIXED

**Problem**: When typing in the query editor in INSERT mode, pressing "1", "2", or "3" would trigger panel navigation instead of typing the characters.

**Root Cause**: In main.go:396-439, navigation keybindings were processed without checking if the editor was in INSERT mode, causing keys to trigger navigation commands instead of being passed to the textarea.

**Solution** (main.go:397):
- Wrapped navigation keybindings with INSERT mode check: `if !(m.focusedPanel == FocusEditor && m.editorPanel.IsInInsertMode())`
- Follows the same pattern used for "L" key (layout preset mode) in main.go:308
- Navigation keys now only work when editor is in NORMAL mode or when other panels are focused

**Expected Behavior After Fix**:
- ✅ Editor in INSERT mode: Typing "1", "2", "3" adds characters to query
- ✅ Editor in NORMAL mode: Keys switch panel focus (1=Connections, 2=Editor, 3=Results)
- ✅ Other panels focused: Keys switch panel focus as expected
- ✅ Tab/Shift+Tab navigation unaffected

**Files Modified**:
- `cmd/lazydb/main.go`: Added INSERT mode guard around navigation keybindings

### Schema Search Mode Key Conflicts ✅ FIXED

#### Issue 1: Connection Panel Keys (a, e, d, s)
**Problem**: When searching in schema explorer (typing "da"), keys like "d" were triggering delete dialog instead of being added to search input.

**Root Cause**: In main.go, connections panel keybindings were checked BEFORE the panel's Update() method, so search mode couldn't intercept the keys.

**Solution**:
- Added `IsInSchemaSearchMode()` method to ConnectionsPanel
- Modified main.go line 402 to check search mode before handling keybindings
- Now search input is properly isolated from command keybindings

#### Issue 2: Schema Command Keys (p, r) ✅ FIXED
**Problem**: When searching for tables containing "p" or "r", these keys triggered preview/refresh commands instead of being added to search input.

**Root Cause**: In connections.go line 122-133, "p" was explicitly handled for preview even in search mode.

**Solution**:
- Removed "p" case from search mode switch statement
- All printable characters now go to search input when in search mode
- To use preview/refresh: exit search (ESC) first, then press p/r
- Follows standard search UX patterns

### Vim-Style Schema Search with Auto-Expand ✅ IMPLEMENTED

**Feature**: Three-state search system with automatic deep search across all schemas

**Key Hierarchy**:
- **Application Level**: `Ctrl+Q` quits LazyDB (NOT 'q')
- **Schema View Level**: `q` exits schema view → returns to connections
- **Search Level**: `ESC` clears search/filter → returns to normal mode (stays in schema view)

**Three States**:

1. **Normal Mode** (full list)
   - Press `/` to enter Search Input Mode
   - All commands work: p (preview), r (refresh), j/k (navigate)
   - `q` exits schema view → connections
   - `ESC` does nothing in normal mode

2. **Search Input Mode** (actively typing, auto-expanding all nodes)
   - **Auto-expand**: Automatically expands ALL schemas/tables/views for deep search
   - Shows loading indicator: `🔍 Search: term_ (Expanding all schemas...)`
   - Type to add characters (live filtering as you type)
   - Backspace to delete characters
   - **Enter** commits search → Search Results Mode
   - **ESC** cancels search → Normal Mode (clears search, stays in schema view)
   - **q** exits schema view → connections
   - j/k navigate filtered results while typing
   - Visual: `🔍 Search: term_` (cyan, with cursor)

3. **Search Results Mode** (filter committed, commands work on filtered data)
   - Filter remains active: `🔍 Filter: term` (green)
   - All commands work on filtered results: p, r, j/k, enter (expand)
   - **ESC** clears filter → Normal Mode (stays in schema view)
   - **q** exits schema view → connections
   - `/` modifies search → re-enters Search Input Mode

**Flow Diagram**:
```
Connections View
    ↓ [s]
Schema Normal Mode (full list)
    ↓ [/]
Search Input Mode (typing, auto-expanding all nodes)
    ↓ [Enter]
Search Results Mode (filter active, all commands work)
    ↓ [ESC]
Schema Normal Mode (full list, nodes still expanded)
    ↓ [q]
Connections View
    ↓ [Ctrl+Q]
Quit Application
```

**Benefits**:
- ✅ No manual expansion needed - search works across entire tree
- ✅ ESC clears search but stays in schema view (one level back)
- ✅ q exits schema view entirely (two levels back)
- ✅ No accidental app quit with 'q' (now Ctrl+Q)
- ✅ Clear hierarchy: ESC = back one level, q = exit view
- ✅ Can use p/r commands on filtered results
- ✅ Vim-like workflow familiar to users

**Files Modified**:
- `internal/config/defaults.go`: Changed Global.Quit from "q" to "ctrl+q"
- `cmd/lazydb/main.go`: Updated footer text to show [Ctrl+Q] Quit
- `internal/ui/components/schema.go`:
  - Added `searchLoading` field, `expandAll()` method
  - Updated `EnterSearchMode()` to auto-expand all nodes
  - Updated `View()` to show loading indicator
- `internal/ui/panels/connections.go`:
  - Removed early ESC interception bug (line 92-96)
  - Added 'q' to exit schema view in all 3 modes
  - Fixed ESC in Search Results Mode to only clear filter
  - Updated Help() text for all modes

## Current Issues Under Investigation

### Resize Keys Not Working
**Problem**: Pressing `=`, `-`, `[`, `]` doesn't resize panels

**Investigation Status**:
- ✅ Config system implemented
- ✅ Keybindings loaded from YAML
- ✅ Debug logging added at 5 critical points
- ⏳ Awaiting log analysis to identify root cause

**Likely Causes**:
1. Keys being consumed by editor in INSERT mode
2. String comparison failing in switch statement
3. Config values not matching pressed keys
4. Fall-through to panel.Update() instead of resize logic

**Next Steps**:
1. Run app with `tail -f ~/.lazydb/debug.log`
2. Press resize keys
3. Analyze LOG3 output to see which config value matches
4. Fix identified issue

## Key Features Implemented

### v1.0 Features
- ✅ PostgreSQL connection management
- ✅ 3-panel TUI (Connections | Editor | Results)
- ✅ Vim-style modal editor (INSERT/NORMAL modes)
- ✅ SQL syntax highlighting (Chroma v2)
- ✅ Real-time SQL linting (pg_query_go v6)
- ✅ Multi-statement query execution
- ✅ Schema explorer with lazy loading
- ✅ Neovim integration (Ctrl+E)
- ✅ Configurable keybindings (YAML)
- ✅ Dynamic panel resizing
- ✅ Layout presets (4 presets)
- ✅ Query history by environment
- ✅ Encrypted password storage

### Key Design Decisions

**Multi-statement Query Support**:
- Uses `conn.Exec()` for DDL/DML and multiple statements
- Uses `conn.Query()` only for single SELECT statements
- Prevents "cannot insert multiple commands into prepared statement" error

**Keybinding Priority**:
1. Dialog keys (highest)
2. Resize keys
3. Layout preset keys
4. Global keys
5. Navigation keys
6. Panel-specific keys (lowest)

**Tmux Compatibility**:
- Avoided `Ctrl+Arrow` and `Alt+Arrow` (conflict with tmux)
- Using `=`, `-`, `[`, `]` for panel resizing instead

## Testing

### Unit Tests
```bash
go test ./tests/unit/... -v
```

### Integration Tests
```bash
# Start test database
docker run --name test-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 -d postgres

# Run tests
TEST_POSTGRES_DSN="postgres://postgres:postgres@localhost:5432/postgres" \
  go test ./tests/integration/... -v
```

## Dependencies

### Key Libraries
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `github.com/jackc/pgx/v5` - PostgreSQL driver
- `github.com/alecthomas/chroma/v2` - Syntax highlighting
- `github.com/pganalyze/pg_query_go/v6` - SQL parsing/validation
- `gopkg.in/yaml.v3` - YAML config parsing

### Go Version
- Minimum: Go 1.21+

## Common Tasks

```bash
# Build
go build -o lazydb ./cmd/lazydb

# Run
./lazydb

# Debug with logs
tail -f ~/.lazydb/debug.log  # Terminal 1
./lazydb                      # Terminal 2

# Edit config
vim ~/.lazydb/config.yml

# View connections
cat ~/.lazydb/connections.json

# Clean build
rm lazydb && go build -o lazydb ./cmd/lazydb
```

## Important Notes

- **BREAKING CHANGE**: Quit key changed from `q` to `Ctrl+Q` (prevents accidental quit)
- `q` now exits schema view → returns to connections
- `ESC` clears search/filter → stays in schema view
- Editor starts in INSERT mode by default (not NORMAL mode)
- Resize keys should work globally, regardless of focused panel
- Config file auto-generates on first run (will use new Ctrl+Q default)
- All passwords encrypted with AES-256-GCM
- Query history stored per environment per month
- System tables (pg_*) filtered out from schema explorer
- Schema search auto-expands all nodes for deep search
