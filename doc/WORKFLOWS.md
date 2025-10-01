# 🔄 User Workflows - LazyDB

Comprehensive user workflow diagrams and procedures for LazyDB operations.

---

## Overview

This document provides detailed, step-by-step workflows for all major LazyDB operations. Each workflow includes:
- **Flow Diagrams**: Visual representation of the process
- **Step-by-Step Instructions**: Detailed procedures with keyboard shortcuts
- **Edge Cases**: Alternative paths and error scenarios
- **Tips & Tricks**: Pro tips for efficiency

---

## Table of Contents

1. [First-Time Setup Workflow](#first-time-setup-workflow)
2. [Connection Management Workflow](#connection-management-workflow)
3. [Query Execution Workflow](#query-execution-workflow)
4. [Schema Exploration Workflow](#schema-exploration-workflow)
5. [Data Export Workflow](#data-export-workflow)
6. [Query History & Library Workflow](#query-history--library-workflow)
7. [Transaction Workflow](#transaction-workflow)
8. [Error Handling Workflow](#error-handling-workflow)

---

## First-Time Setup Workflow

### Initial Launch Experience

```
┌─────────────────────────────────────────────────────────────┐
│                    First Launch Flow                         │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Launch LazyDB                                               │
│      ↓                                                       │
│  Empty State Screen                                          │
│      ├─ "No connections configured"                          │
│      ├─ "Press 'a' to add your first connection"            │
│      └─ Quick start guide hint                               │
│      ↓                                                       │
│  User presses 'a'                                            │
│      ↓                                                       │
│  Connection Form Modal                                       │
│      ├─ Fill in connection details                           │
│      ├─ Test connection (optional)                           │
│      └─ Save connection                                      │
│      ↓                                                       │
│  Connection appears in sidebar                               │
│      ↓                                                       │
│  User presses Enter on connection                            │
│      ↓                                                       │
│  Connected! Status bar shows "Connected to dev-local"        │
│      ↓                                                       │
│  Ready to query                                              │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Step-by-Step: First Connection Setup

**Step 1: Launch LazyDB**
```bash
$ lazydb
```

**Step 2: Add First Connection**
- Press `a` to open connection form
- **Empty state helper**: Form appears with focus on "Name" field

**Step 3: Fill Connection Details**

| Field | Action | Example |
|-------|--------|---------|
| **Name** | Type connection name | `dev-local` |
| `Tab` | Move to Type | |
| **Type** | Select database | `PostgreSQL` |
| `Tab` | Move to Host | |
| **Host** | Type hostname | `localhost` |
| `Tab` | Move to Port | |
| **Port** | Enter port (defaults auto-fill) | `5432` |
| `Tab` | Move to Database | |
| **Database** | Database name | `myapp_dev` |
| `Tab` | Move to Username | |
| **Username** | Your username | `postgres` |
| `Tab` | Move to Password | |
| **Password** | Your password | `••••••••` |
| `Tab` | Move to Environment | |
| **Environment** | Select env | `🟢 Development` |

**Step 4: Test Connection (Optional)**
- Press `Ctrl-T` to test connection
- Wait for confirmation: "✅ Connection successful"
- If fails: See error message and fix details

**Step 5: Save Connection**
- Press `Enter` to save
- Connection appears in left sidebar under "🟢 Development"

**Step 6: Connect to Database**
- Press `Enter` on the connection
- Status bar shows: "▸ Connected: dev-local @ localhost:5432/myapp_dev"
- Schema tree loads (if enabled)

**Step 7: Run Your First Query**
- Press `2` to focus editor panel
- Type: `SELECT version();`
- Press `Ctrl-R` to execute
- Results appear in right panel

🎉 **Success!** You're now ready to use LazyDB.

---

## Connection Management Workflow

### Adding a New Connection

```
┌─────────────────────────────────────────────────────────────┐
│                 Add Connection Flow                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Focus Connections Panel (press '1')                         │
│      ↓                                                       │
│  Press 'a' (Add Connection)                                  │
│      ↓                                                       │
│  Connection Form Opens                                       │
│      ├─ Fill required fields                                 │
│      ├─ Optional: Test connection (Ctrl-T)                   │
│      │     ├─ Success → Continue                             │
│      │     └─ Failure → Fix and retry                        │
│      └─ Save (Enter) or Cancel (Esc)                         │
│      ↓                                                       │
│  Connection Added to List                                    │
│      ├─ Appears in appropriate env group                     │
│      └─ Saved to ~/.config/lazydb/connections.toml          │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Editing an Existing Connection

```
┌─────────────────────────────────────────────────────────────┐
│                 Edit Connection Flow                         │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Select Connection (use j/k)                                 │
│      ↓                                                       │
│  Press 'e' (Edit)                                            │
│      ↓                                                       │
│  Connection Form Opens (Pre-filled)                          │
│      ├─ Modify fields as needed                              │
│      ├─ Password field shows ••••••••                        │
│      │   (Leave unchanged or enter new)                      │
│      └─ Save (Enter) or Cancel (Esc)                         │
│      ↓                                                       │
│  Connection Updated                                          │
│      ├─ If connected: Prompt to reconnect                    │
│      └─ Changes saved immediately                            │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Deleting a Connection

```
┌─────────────────────────────────────────────────────────────┐
│                Delete Connection Flow                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Select Connection to Delete                                 │
│      ↓                                                       │
│  Press 'x' (Delete)                                          │
│      ↓                                                       │
│  Confirmation Dialog                                         │
│      "Delete connection 'dev-local'? (y/n)"                  │
│      ├─ Press 'y' → Delete                                   │
│      └─ Press 'n' or Esc → Cancel                            │
│      ↓                                                       │
│  If Active Connection                                        │
│      ├─ Warning: "This connection is active!"                │
│      ├─ "Disconnect first? (y/n)"                            │
│      │     ├─ 'y' → Disconnect, then delete                  │
│      │     └─ 'n' → Cancel delete                            │
│      └─                                                      │
│      ↓                                                       │
│  Connection Removed                                          │
│      ├─ Removed from sidebar                                 │
│      └─ Removed from config file                             │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Connecting to a Database

```
┌─────────────────────────────────────────────────────────────┐
│                  Connect Workflow                            │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Select Connection (j/k to navigate)                         │
│      ↓                                                       │
│  Press Enter                                                 │
│      ↓                                                       │
│  Connection Attempt                                          │
│      ├─ Show loading indicator "Connecting..."               │
│      ├─ Establish connection                                 │
│      └─ Connection pool created                              │
│      ↓                                                       │
│  Success?                                                    │
│      ├─ Yes → Connected State                                │
│      │     ├─ Status bar: "Connected: dev-local"             │
│      │     ├─ Connection marked with ✓                       │
│      │     ├─ Schema tree loads (if enabled)                 │
│      │     └─ Ready to query                                 │
│      │                                                       │
│      └─ No → Error State                                     │
│            ├─ Show error message                             │
│            ├─ Common issues:                                 │
│            │   • Wrong credentials → Edit & retry            │
│            │   • Server down → Check server                  │
│            │   • Network issue → Check connection            │
│            └─ Remain disconnected                            │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Switching Connections

```
┌─────────────────────────────────────────────────────────────┐
│               Switch Connection Flow                         │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Currently Connected to Connection A                         │
│      ↓                                                       │
│  Select Connection B                                         │
│      ↓                                                       │
│  Press Enter                                                 │
│      ↓                                                       │
│  Unsaved Query?                                              │
│      ├─ Yes → "Save query before switching? (y/n/c)"         │
│      │     ├─ 'y' → Save query, continue                     │
│      │     ├─ 'n' → Discard, continue                        │
│      │     └─ 'c' → Cancel switch                            │
│      └─ No → Continue                                        │
│      ↓                                                       │
│  Disconnect from Connection A                                │
│      ├─ Close connection pool                                │
│      ├─ Clear results panel                                  │
│      └─ Update status                                        │
│      ↓                                                       │
│  Connect to Connection B                                     │
│      ├─ Establish new connection                             │
│      ├─ Load schema                                          │
│      └─ Update UI                                            │
│      ↓                                                       │
│  Connected to Connection B                                   │
│      ├─ Status bar updated                                   │
│      └─ Ready for queries                                    │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Query Execution Workflow

### Basic Query Execution

```
┌─────────────────────────────────────────────────────────────┐
│                  Query Execution Flow                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Focus Editor Panel (press '2')                              │
│      ↓                                                       │
│  Write/Edit Query                                            │
│      ├─ Type query directly                                  │
│      │   OR                                                  │
│      └─ Press Ctrl-E to edit in Neovim                       │
│      ↓                                                       │
│  Execute Query (Ctrl-R)                                      │
│      ↓                                                       │
│  Validation Check                                            │
│      ├─ Connected? → Yes, continue                           │
│      ├─ Not connected? → Error "Not connected"               │
│      ├─ Empty query? → Error "Empty query"                   │
│      └─ Valid → Continue                                     │
│      ↓                                                       │
│  Execution                                                   │
│      ├─ Show loading indicator in results panel              │
│      ├─ Send query to database                               │
│      ├─ Status bar: "Executing..."                           │
│      └─ Timer starts                                         │
│      ↓                                                       │
│  Wait for Response                                           │
│      ├─ User can press Ctrl-C to cancel (if supported)       │
│      └─ Timeout after 30s (configurable)                     │
│      ↓                                                       │
│  Success?                                                    │
│      ├─ Yes → Display Results                                │
│      │     ├─ Parse result set                               │
│      │     ├─ Format table in results panel                  │
│      │     ├─ Show metadata: "10 rows (42ms)"                │
│      │     └─ Focus results panel (auto)                     │
│      │                                                       │
│      └─ No → Display Error                                   │
│            ├─ Show error message in results panel            │
│            ├─ Highlight problematic line (if available)      │
│            ├─ Suggest fix (if common error)                  │
│            └─ Keep focus on editor                           │
│      ↓                                                       │
│  Save to History                                             │
│      ├─ Query saved to history.db                            │
│      ├─ Include: query, result, timestamp, duration          │
│      └─ Accessible via Ctrl-H                                │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Editing Query in Neovim

```
┌─────────────────────────────────────────────────────────────┐
│               Neovim Editing Workflow                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Editor Panel Active with Query                              │
│      ↓                                                       │
│  Press Ctrl-E                                                │
│      ↓                                                       │
│  LazyDB Suspends TUI                                         │
│      ├─ Save current query to temp file                      │
│      ├─ Path: /tmp/lazydb-query-XXXXXX.sql                   │
│      └─ Suspend Bubbletea UI                                 │
│      ↓                                                       │
│  Neovim Spawns                                               │
│      ├─ Opens temp file                                      │
│      ├─ Filetype set to 'sql'                                │
│      ├─ User's Neovim config loads                           │
│      │   (LSP, completion, snippets, etc.)                   │
│      └─ Full terminal control                                │
│      ↓                                                       │
│  User Edits in Neovim                                        │
│      ├─ Full Vim power                                       │
│      ├─ Syntax highlighting                                  │
│      ├─ LSP features (if configured)                         │
│      └─ Custom keybindings                                   │
│      ↓                                                       │
│  User Saves and Quits (:wq)                                  │
│      ↓                                                       │
│  Neovim Exits                                                │
│      ├─ Edited query saved to temp file                      │
│      └─ Control returns to LazyDB                            │
│      ↓                                                       │
│  LazyDB Resumes                                              │
│      ├─ Read edited query from temp file                     │
│      ├─ Update editor panel content                          │
│      ├─ Clean up temp file                                   │
│      └─ Restore TUI                                          │
│      ↓                                                       │
│  Query Updated in Editor                                     │
│      ├─ Ready to execute (Ctrl-R)                            │
│      └─ Or edit further (Ctrl-E again)                       │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Multi-Query Execution (Future)

```
┌─────────────────────────────────────────────────────────────┐
│             Multi-Query Execution Flow                       │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Editor Contains Multiple Queries                            │
│      SELECT * FROM users;                                    │
│      SELECT * FROM orders;                                   │
│      SELECT * FROM products;                                 │
│      ↓                                                       │
│  Press Ctrl-Shift-R (Execute All)                            │
│      ↓                                                       │
│  Split into Individual Queries                               │
│      ├─ Detect semicolon separators                          │
│      └─ Create query list                                    │
│      ↓                                                       │
│  Execute Sequentially                                        │
│      ├─ Query 1 → Execute → Result 1                         │
│      ├─ Query 2 → Execute → Result 2                         │
│      └─ Query 3 → Execute → Result 3                         │
│      ↓                                                       │
│  Display Results                                             │
│      ├─ Tabbed interface (one tab per result)                │
│      ├─ Tab labels: "users (10 rows)", "orders (25 rows)"    │
│      └─ Switch tabs with p/n keys                            │
│      ↓                                                       │
│  Summary in Status Bar                                       │
│      "3 queries executed (125ms total)"                      │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Schema Exploration Workflow

### Browsing Tables

```
┌─────────────────────────────────────────────────────────────┐
│                Schema Exploration Flow                       │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Connect to Database                                         │
│      ↓                                                       │
│  Press 's' (Toggle Schema)                                   │
│      ↓                                                       │
│  Schema Tree Loads                                           │
│      ├─ Show loading indicator                               │
│      ├─ Query: ListTables(), ListViews()                     │
│      └─ Build tree structure                                 │
│      ↓                                                       │
│  Schema Tree Displayed                                       │
│      ▼ 📚 Tables (25)                                        │
│        • users                                               │
│        • orders                                              │
│        • products                                            │
│      ▶ 👁️ Views (3)                                          │
│      ▶ ⚡ Functions (10)                                      │
│      ↓                                                       │
│  Navigate with j/k                                           │
│      ↓                                                       │
│  Select Table (e.g., "users")                                │
│      ↓                                                       │
│  Available Actions                                           │
│      ├─ Enter → Preview data                                 │
│      ├─ 'i' → Show table info                                │
│      ├─ 'c' → Copy table name                                │
│      └─ Expand → Show columns                                │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Table Preview

```
┌─────────────────────────────────────────────────────────────┐
│                 Table Preview Flow                           │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Select Table in Schema Tree                                 │
│      ↓                                                       │
│  Press Enter                                                 │
│      ↓                                                       │
│  Generate Preview Query                                      │
│      SELECT * FROM users LIMIT 10;                           │
│      ↓                                                       │
│  Execute Query                                               │
│      ├─ Show in editor panel (read-only)                     │
│      └─ Results in results panel                             │
│      ↓                                                       │
│  User Reviews Data                                           │
│      ├─ Navigate with hjkl                                   │
│      ├─ Scroll through columns                               │
│      └─ Copy data if needed                                  │
│      ↓                                                       │
│  User Can Modify Query                                       │
│      ├─ Press '2' to focus editor                            │
│      ├─ Modify LIMIT or add WHERE clause                     │
│      └─ Re-execute with Ctrl-R                               │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Table Structure

```
┌─────────────────────────────────────────────────────────────┐
│               View Table Structure Flow                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Select Table                                                │
│      ↓                                                       │
│  Press 'i' (Info)                                            │
│      ↓                                                       │
│  Query Table Schema                                          │
│      ├─ PostgreSQL: information_schema.columns               │
│      ├─ MySQL: DESCRIBE table                                │
│      └─ SQLite: PRAGMA table_info(table)                     │
│      ↓                                                       │
│  Display Structure in Modal                                  │
│      ┌────────────────────────────────────────┐             │
│      │     Table: users                       │             │
│      ├────────────────────────────────────────┤             │
│      │ Column    | Type      | Nullable | PK  │             │
│      ├────────────────────────────────────────┤             │
│      │ id        | INTEGER   | NO       | YES │             │
│      │ username  | VARCHAR   | NO       |     │             │
│      │ email     | VARCHAR   | NO       |     │             │
│      │ created_at| TIMESTAMP | YES      |     │             │
│      └────────────────────────────────────────┘             │
│      ↓                                                       │
│  Additional Info                                             │
│      ├─ Indexes: idx_users_email (UNIQUE)                    │
│      ├─ Constraints: FK to profiles table                    │
│      └─ Row count: ~1,245 rows                               │
│      ↓                                                       │
│  Press Esc to Close                                          │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Data Export Workflow

### Export to CSV

```
┌─────────────────────────────────────────────────────────────┐
│                  CSV Export Flow                             │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Results Displayed in Results Panel                          │
│      ↓                                                       │
│  Press 'e' (Export Menu)                                     │
│      ↓                                                       │
│  Export Options Modal                                        │
│      • CSV (selected)                                        │
│      • JSON                                                  │
│      • SQL                                                   │
│      • Markdown                                              │
│      • Clipboard                                             │
│      ↓                                                       │
│  Press Enter (CSV selected)                                  │
│      ↓                                                       │
│  CSV Options Dialog                                          │
│      ├─ Include headers? [✓] Yes                             │
│      ├─ Delimiter: [,] Comma                                 │
│      ├─ Quote style: ["] Double quote                        │
│      └─ File path: ~/Downloads/results.csv                   │
│      ↓                                                       │
│  Enter File Path or Accept Default                           │
│      ↓                                                       │
│  Export Process                                              │
│      ├─ Show progress: "Exporting... (0/100 rows)"           │
│      ├─ Write CSV file                                       │
│      └─ Close file handle                                    │
│      ↓                                                       │
│  Success!                                                    │
│      ├─ Message: "✅ Exported 100 rows to results.csv"       │
│      └─ File saved to specified path                         │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Copy to Clipboard

```
┌─────────────────────────────────────────────────────────────┐
│                Clipboard Copy Flow                           │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Results Displayed                                           │
│      ↓                                                       │
│  Copy Options                                                │
│      ├─ 'y' → Copy single cell                               │
│      ├─ 'Y' → Copy entire row                                │
│      ├─ Ctrl-Y → Copy column                                 │
│      └─ Ctrl-A → Copy all results                            │
│      ↓                                                       │
│  Example: Copy Row (Press 'Y')                               │
│      ↓                                                       │
│  Format Row Data                                             │
│      "1    Alice    alice@example.com    2024-01-15"         │
│      (tab-separated by default)                              │
│      ↓                                                       │
│  Copy to System Clipboard                                    │
│      ├─ Use pbcopy (macOS)                                   │
│      ├─ Use xclip (Linux)                                    │
│      └─ Use clip.exe (Windows)                               │
│      ↓                                                       │
│  Success Notification                                        │
│      "✅ 1 row copied to clipboard"                          │
│      ↓                                                       │
│  Paste Anywhere (Ctrl-V)                                     │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Export to SQL

```
┌─────────────────────────────────────────────────────────────┐
│                  SQL Export Flow                             │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Select SQL Export Option                                    │
│      ↓                                                       │
│  SQL Export Options                                          │
│      ├─ Table name: users                                    │
│      ├─ Include CREATE TABLE? [✓] Yes                        │
│      ├─ Include DROP TABLE? [ ] No                           │
│      └─ File path: ~/Downloads/users.sql                     │
│      ↓                                                       │
│  Generate SQL Statements                                     │
│      CREATE TABLE users (...);                               │
│      INSERT INTO users (id, name, email) VALUES             │
│        (1, 'Alice', 'alice@example.com'),                    │
│        (2, 'Bob', 'bob@example.com'),                        │
│        ...                                                   │
│      ↓                                                       │
│  Write to File                                               │
│      ↓                                                       │
│  Success!                                                    │
│      "✅ Exported 100 rows as SQL INSERT statements"         │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Query History & Library Workflow

### Accessing Query History

```
┌─────────────────────────────────────────────────────────────┐
│                Query History Flow                            │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Press Ctrl-H (anywhere)                                     │
│      ↓                                                       │
│  History Modal Opens                                         │
│      ┌────────────────────────────────────────┐             │
│      │         Query History                  │             │
│      ├────────────────────────────────────────┤             │
│      │ 2024-01-15 14:32  SELECT * FROM users  │             │
│      │ 2024-01-15 14:30  SELECT COUNT(*) ...  │             │
│      │ 2024-01-15 14:25  INSERT INTO orders   │             │
│      │ ...                                    │             │
│      └────────────────────────────────────────┘             │
│      ↓                                                       │
│  Navigate with j/k                                           │
│      ↓                                                       │
│  Preview Query Details                                       │
│      ├─ Full query text                                      │
│      ├─ Execution time                                       │
│      ├─ Row count                                            │
│      └─ Connection used                                      │
│      ↓                                                       │
│  Actions Available                                           │
│      ├─ Enter → Load query to editor                         │
│      ├─ 'd' → Delete from history                            │
│      ├─ '/' → Search history                                 │
│      └─ Esc → Close modal                                    │
│      ↓                                                       │
│  Select Query and Press Enter                                │
│      ↓                                                       │
│  Query Loaded to Editor                                      │
│      ├─ Replaces current editor content                      │
│      │   (with confirmation if unsaved)                      │
│      └─ Ready to modify and execute                          │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Saving to Query Library

```
┌─────────────────────────────────────────────────────────────┐
│                Save to Library Flow                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Query in Editor                                             │
│      ↓                                                       │
│  Press Ctrl-S                                                │
│      ↓                                                       │
│  Save Query Dialog                                           │
│      ┌────────────────────────────────────────┐             │
│      │       Save Query to Library            │             │
│      ├────────────────────────────────────────┤             │
│      │ Name: [Get active users              ] │             │
│      │ Tags: [users, active                 ] │             │
│      │ Notes: [Returns all users with...    ] │             │
│      └────────────────────────────────────────┘             │
│      ↓                                                       │
│  Fill in Metadata                                            │
│      ├─ Name: Descriptive name (required)                    │
│      ├─ Tags: Comma-separated tags (optional)                │
│      └─ Notes: Description (optional)                        │
│      ↓                                                       │
│  Press Enter to Save                                         │
│      ↓                                                       │
│  Query Saved                                                 │
│      ├─ Stored in ~/.config/lazydb/library/                  │
│      ├─ Accessible via Ctrl-O                                │
│      └─ Searchable by name/tags                              │
│      ↓                                                       │
│  Success Message                                             │
│      "✅ Query saved to library"                             │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Loading from Query Library

```
┌─────────────────────────────────────────────────────────────┐
│               Load from Library Flow                         │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Press Ctrl-O (Open)                                         │
│      ↓                                                       │
│  Library Modal Opens                                         │
│      ┌────────────────────────────────────────┐             │
│      │         Query Library                  │             │
│      ├────────────────────────────────────────┤             │
│      │ Get active users          [users]     │             │
│      │ Monthly sales report      [reports]   │             │
│      │ User registration stats   [analytics] │             │
│      │ ...                                    │             │
│      └────────────────────────────────────────┘             │
│      ↓                                                       │
│  Search/Filter                                               │
│      ├─ Press '/' to search                                  │
│      ├─ Type "sales" → filters list                          │
│      └─ Esc to clear filter                                  │
│      ↓                                                       │
│  Select Query                                                │
│      ↓                                                       │
│  Preview Panel                                               │
│      ├─ Full query text                                      │
│      ├─ Tags: reports, monthly                               │
│      ├─ Notes: "Generates monthly sales report"              │
│      └─ Last used: 2024-01-15                                │
│      ↓                                                       │
│  Press Enter to Load                                         │
│      ↓                                                       │
│  Query Loaded to Editor                                      │
│      ├─ Ready to modify                                      │
│      └─ Ready to execute                                     │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Transaction Workflow

### Manual Transaction Control

```
┌─────────────────────────────────────────────────────────────┐
│                Transaction Workflow                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Connected to Database                                       │
│      ↓                                                       │
│  Start Transaction                                           │
│      Press Ctrl-Shift-B (BEGIN)                              │
│      ↓                                                       │
│  Transaction Started                                         │
│      ├─ Status bar shows: "🔄 Transaction Active"            │
│      ├─ Visual indicator in UI (orange border)               │
│      └─ All queries run in transaction                       │
│      ↓                                                       │
│  Execute Queries                                             │
│      UPDATE products SET price = 100 WHERE id = 1;           │
│      INSERT INTO audit_log (...);                            │
│      DELETE FROM cache WHERE expired = true;                 │
│      ↓                                                       │
│  Review Results                                              │
│      ├─ Check if all queries succeeded                       │
│      ├─ Verify data changes                                  │
│      └─ Decide: COMMIT or ROLLBACK                           │
│      ↓                                                       │
│  Commit or Rollback                                          │
│      ├─ Ctrl-Shift-C → COMMIT                                │
│      │     ├─ Changes persisted                              │
│      │     ├─ Status: "✅ Transaction committed"             │
│      │     └─ Transaction ends                               │
│      │                                                       │
│      └─ Ctrl-Shift-R → ROLLBACK                              │
│            ├─ Changes discarded                              │
│            ├─ Status: "↩️ Transaction rolled back"           │
│            └─ Transaction ends                               │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Auto-Rollback on Error

```
┌─────────────────────────────────────────────────────────────┐
│              Auto-Rollback Workflow                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Transaction Active                                          │
│      ↓                                                       │
│  Execute Query                                               │
│      UPDATE users SET email = 'invalid' WHERE id = 999;      │
│      ↓                                                       │
│  Query Fails                                                 │
│      Error: "No rows affected"                               │
│      ↓                                                       │
│  Error Modal                                                 │
│      ┌────────────────────────────────────────┐             │
│      │  ⚠️  Query Failed in Transaction       │             │
│      ├────────────────────────────────────────┤             │
│      │ Error: No rows affected                │             │
│      │                                        │             │
│      │ Transaction is still active.           │             │
│      │ What would you like to do?             │             │
│      │                                        │             │
│      │  [C] Commit anyway                     │             │
│      │  [R] Rollback (recommended)            │             │
│      │  [I] Ignore and continue               │             │
│      └────────────────────────────────────────┘             │
│      ↓                                                       │
│  User Chooses                                                │
│      ├─ 'R' → ROLLBACK → Safe state restored                 │
│      ├─ 'C' → COMMIT → Partial changes saved                 │
│      └─ 'I' → Continue → Risk of inconsistency               │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Error Handling Workflow

### Connection Errors

```
┌─────────────────────────────────────────────────────────────┐
│              Connection Error Flow                           │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Attempt to Connect                                          │
│      ↓                                                       │
│  Connection Fails                                            │
│      ↓                                                       │
│  Identify Error Type                                         │
│      ├─ Authentication Failed                                │
│      ├─ Host Unreachable                                     │
│      ├─ Database Not Found                                   │
│      ├─ Timeout                                              │
│      └─ SSL/TLS Error                                        │
│      ↓                                                       │
│  Error Modal with Diagnosis                                  │
│      ┌────────────────────────────────────────┐             │
│      │  ❌ Connection Failed                   │             │
│      ├────────────────────────────────────────┤             │
│      │ Error: Authentication failed           │             │
│      │                                        │             │
│      │ Diagnosis:                             │             │
│      │ • Username or password incorrect       │             │
│      │                                        │             │
│      │ Suggestions:                           │             │
│      │ 1. Check credentials                   │             │
│      │ 2. Verify user exists in database      │             │
│      │ 3. Check user permissions              │             │
│      │                                        │             │
│      │ [E] Edit Connection  [R] Retry  [C]ancel│             │
│      └────────────────────────────────────────┘             │
│      ↓                                                       │
│  User Action                                                 │
│      ├─ 'E' → Edit connection form                           │
│      ├─ 'R' → Retry with same credentials                    │
│      └─ 'C' → Cancel, remain disconnected                    │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Query Syntax Errors

```
┌─────────────────────────────────────────────────────────────┐
│               Syntax Error Flow                              │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Execute Query                                               │
│      SELET * FROM users;  -- Typo: SELET                     │
│      ↓                                                       │
│  Database Returns Error                                      │
│      ↓                                                       │
│  Parse Error Message                                         │
│      "syntax error at or near 'SELET'"                       │
│      ↓                                                       │
│  Error Display in Results Panel                              │
│      ┌────────────────────────────────────────┐             │
│      │  ❌ Syntax Error                        │             │
│      ├────────────────────────────────────────┤             │
│      │ syntax error at or near "SELET"        │             │
│      │                                        │             │
│      │ Line 1, Position 1                     │             │
│      │ ───────────────────────────            │             │
│      │ SELET * FROM users;                    │             │
│      │ ^^^^^                                  │             │
│      │                                        │             │
│      │ Did you mean: SELECT?                  │             │
│      │                                        │             │
│      │ [F] Fix automatically  [E]dit manually │             │
│      └────────────────────────────────────────┘             │
│      ↓                                                       │
│  User Action                                                 │
│      ├─ 'F' → Auto-fix to SELECT                             │
│      │     └─ Query updated in editor                        │
│      │                                                       │
│      └─ 'E' or Esc → Edit manually                           │
│            └─ Cursor positioned at error                     │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Query Timeout

```
┌─────────────────────────────────────────────────────────────┐
│                  Timeout Flow                                │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Long-Running Query                                          │
│      SELECT * FROM huge_table;  -- Takes > 30s               │
│      ↓                                                       │
│  Execution Status                                            │
│      Status bar: "Executing... (15s elapsed)"                │
│      ↓                                                       │
│  User Notices Slow Query                                     │
│      ↓                                                       │
│  Options                                                     │
│      ├─ Wait: Let query finish                               │
│      └─ Cancel: Press Ctrl-C                                 │
│      ↓                                                       │
│  If Timeout Reached (30s)                                    │
│      ↓                                                       │
│  Timeout Modal                                               │
│      ┌────────────────────────────────────────┐             │
│      │  ⏱️  Query Timeout                      │             │
│      ├────────────────────────────────────────┤             │
│      │ Query exceeded 30s timeout limit       │             │
│      │                                        │             │
│      │ The query may still be running on      │             │
│      │ the database server.                   │             │
│      │                                        │             │
│      │ Suggestions:                           │             │
│      │ 1. Add LIMIT clause                    │             │
│      │ 2. Add WHERE clause to filter          │             │
│      │ 3. Check indexes on table              │             │
│      │ 4. Increase timeout in config          │             │
│      │                                        │             │
│      │ [K] Kill query  [W]ait longer  [C]ancel│             │
│      └────────────────────────────────────────┘             │
│      ↓                                                       │
│  User Choice                                                 │
│      ├─ 'K' → Send CANCEL signal (if supported)              │
│      ├─ 'W' → Wait another 30s                               │
│      └─ 'C' → Give up, return to editor                      │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Connection Lost

```
┌─────────────────────────────────────────────────────────────┐
│              Connection Lost Flow                            │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Query in Progress                                           │
│      ↓                                                       │
│  Connection Interrupted                                      │
│      • Network issue                                         │
│      • Database server restart                               │
│      • Connection killed by DBA                              │
│      ↓                                                       │
│  Error Detected                                              │
│      "connection lost"                                       │
│      ↓                                                       │
│  Update UI State                                             │
│      ├─ Mark connection as disconnected                      │
│      ├─ Remove ✓ from connection                             │
│      └─ Status bar: "❌ Connection lost"                     │
│      ↓                                                       │
│  Error Modal                                                 │
│      ┌────────────────────────────────────────┐             │
│      │  ⚠️  Connection Lost                    │             │
│      ├────────────────────────────────────────┤             │
│      │ Lost connection to dev-local           │             │
│      │                                        │             │
│      │ Possible causes:                       │             │
│      │ • Network interruption                 │             │
│      │ • Database server restart              │             │
│      │ • Idle timeout exceeded                │             │
│      │                                        │             │
│      │ [R] Reconnect  [S]witch connection     │             │
│      └────────────────────────────────────────┘             │
│      ↓                                                       │
│  User Action                                                 │
│      ├─ 'R' → Attempt reconnect                              │
│      │     ├─ Success → Resume work                          │
│      │     └─ Failure → Show error again                     │
│      │                                                       │
│      └─ 'S' → Switch to different connection                 │
│            └─ Connection picker modal                        │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Power User Workflows

### Rapid Fire Query Development

```
┌─────────────────────────────────────────────────────────────┐
│           Rapid Development Workflow                         │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Connected to dev database                                   │
│      ↓                                                       │
│  Write base query                                            │
│      SELECT * FROM users                                     │
│      ↓                                                       │
│  Ctrl-R → Execute → See results                              │
│      ↓                                                       │
│  Refine query (add WHERE)                                    │
│      SELECT * FROM users WHERE active = true                 │
│      ↓                                                       │
│  Ctrl-R → Execute → See filtered results                     │
│      ↓                                                       │
│  Refine again (add ORDER BY)                                 │
│      SELECT * FROM users WHERE active = true                 │
│      ORDER BY created_at DESC                                │
│      ↓                                                       │
│  Ctrl-R → Execute → See sorted results                       │
│      ↓                                                       │
│  Perfect! Save to library                                    │
│      Ctrl-S → Name: "Active users by date"                   │
│      ↓                                                       │
│  Total time: < 1 minute 🚀                                   │
│                                                              │
│  Key to speed:                                               │
│  • No modal switches                                         │
│  • Keyboard-only workflow                                    │
│  • Instant feedback                                          │
│  • History tracking automatic                                │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Schema-Driven Development

```
┌─────────────────────────────────────────────────────────────┐
│          Schema-Driven Development Flow                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Press '1' → Focus connections panel                         │
│      ↓                                                       │
│  Press 's' → Show schema tree                                │
│      ↓                                                       │
│  Navigate schema                                             │
│      ▼ Tables                                                │
│        ▼ users (expand)                                      │
│          • id                                                │
│          • username                                          │
│          • email                                             │
│        ▶ orders                                              │
│      ↓                                                       │
│  Press 'c' on "users" → Copy table name                      │
│      ↓                                                       │
│  Press '2' → Jump to editor                                  │
│      ↓                                                       │
│  Type query with autocomplete context                        │
│      SELECT * FROM users  -- users in clipboard              │
│      ↓                                                       │
│  Expand "users" in schema                                    │
│      See columns: id, username, email                        │
│      ↓                                                       │
│  Refine query                                                │
│      SELECT id, username FROM users WHERE...                 │
│      ↓                                                       │
│  Execute → Results                                           │
│      ↓                                                       │
│  Schema knowledge drives development                         │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Best Practices

### Connection Management Best Practices

1. **Organization**
   - Use environments: 🟢 Dev, 🔵 Staging, 🔴 Production
   - Descriptive names: `dev-local-postgres`, not `db1`
   - Add connection notes in description field

2. **Security**
   - Never save production passwords in config
   - Use SSH tunnels for remote connections
   - Set up read-only users for production

3. **Performance**
   - Configure appropriate connection pool sizes
   - Set reasonable timeout values
   - Use persistent connections for frequently accessed DBs

### Query Development Best Practices

1. **Start Small**
   - Begin with `SELECT * FROM table LIMIT 10`
   - Add filters incrementally
   - Test each modification

2. **Use EXPLAIN**
   - Check query plans before executing large queries
   - Identify missing indexes
   - Optimize slow queries

3. **Transaction Safety**
   - Always use transactions for multi-query operations
   - Test on dev before running on prod
   - Have rollback plan ready

### Keyboard Efficiency

1. **Master Core Shortcuts**
   - `1`, `2`, `3` for panel navigation
   - `Ctrl-R` for execute
   - `Ctrl-E` for Neovim
   - `Ctrl-H` for history

2. **Learn Advanced Shortcuts**
   - `Tab`/`Shift-Tab` for panel cycling
   - `g`/`G` for jump to top/bottom
   - `y`/`Y` for copy cell/row
   - `/` for search in any context

3. **Customize Keybindings**
   - Map frequently-used actions to convenient keys
   - Use `<leader>` for custom commands
   - Document your custom mappings

---

## Troubleshooting Common Issues

### "Command Not Found: lazydb"

**Problem**: LazyDB not in PATH

**Solution**:
```bash
# Add to ~/.bashrc or ~/.zshrc
export PATH="$PATH:/path/to/lazydb"

# Or create symlink
ln -s /path/to/lazydb /usr/local/bin/lazydb
```

### Slow Query Execution

**Problem**: Queries taking too long

**Investigation Workflow**:
1. Check table size: `SELECT COUNT(*) FROM table`
2. Check indexes: Press 'i' on table in schema
3. Use EXPLAIN: `EXPLAIN SELECT...`
4. Add appropriate indexes or LIMIT clauses

### Connection Timeout

**Problem**: Connections timing out frequently

**Solution**:
```toml
# ~/.config/lazydb/config.toml
[connections]
timeout = 60  # Increase from 30s default
max_idle_time = 300  # 5 minutes
```

### Neovim Not Found

**Problem**: Ctrl-E doesn't work

**Check**:
```bash
which nvim
```

**Solution**:
```toml
# ~/.config/lazydb/config.toml
[editor]
nvim_path = "/opt/homebrew/bin/nvim"  # Custom path
fallback_enabled = true  # Use built-in editor if nvim not found
```

---

## See Also

- [Keybindings Reference](./KEYBINDINGS.md) - Complete keyboard shortcuts
- [UI Mockups](./UI_MOCKUPS.md) - Visual UI guide
- [Neovim Integration](./NEOVIM_INTEGRATION.md) - Editor integration details
- [Architecture](./ARCHITECTURE.md) - System architecture
- [Implementation Plan](./IMPLEMENTATION_PLAN.md) - Development roadmap

---

**Document Version**: v1.0
**Last Updated**: 2024-01
**Applies to LazyDB**: v0.1.0-dev (MVP Phase)
