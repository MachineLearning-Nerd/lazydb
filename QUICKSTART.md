# LazyDB Quick Start Guide

Get up and running with LazyDB in 5 minutes!

## Installation

```bash
# Clone and build
git clone https://github.com/yourusername/lazydb.git
cd lazydb
go build -o lazydb cmd/lazydb/main.go

# Run
./lazydb
```

## First Steps

### 1. Add a Connection

When you first launch LazyDB, you'll see three panels. Let's add your first database connection:

```
Press '1' → Focus Connections panel
Press 'a' → Open "Add Connection" dialog
```

Fill in the form:
- **Name**: `my-database`
- **Host**: `localhost` (or your database host)
- **Port**: `5432`
- **Database**: `postgres` (or your database name)
- **Username**: `postgres` (or your username)
- **Password**: [your password]
- **SSL Mode**: `disable` (for local) or `require` (for remote)
- **Environment**: Use `←/→` to select (Development/Staging/Production)

```
Press Enter → Save connection
```

### 2. Connect

```
Use j/k (or ↑/↓) → Select your connection
Press Enter → Connect to database
```

You'll see a green status indicator when connected successfully.

### 3. Run a Query

```
Press '2' → Focus Editor panel
Type your SQL → Example: SELECT * FROM pg_database;
Press Ctrl+R → Execute query
```

Results appear in the right panel automatically!

## Common Operations

### Navigate Between Panels

```
Press '1' → Connections panel
Press '2' → Editor panel
Press '3' → Results panel

Or use Tab/Shift+Tab to cycle through panels
```

### Explore Database Schema

```
Press '1' → Focus Connections panel
Press 's' → Open Schema Explorer
Use j/k → Navigate schemas, tables, views, functions
Press Enter/Space → Expand/collapse nodes
Press 'p' on a table → Generate preview query (SELECT * LIMIT 10)
Press Esc → Return to connections
```

The schema explorer shows:
- 📂 Schemas - All available schemas
- 📊 Tables - All tables with column details
- 👁 Views - Database views
- ⚙ Functions - Stored procedures and functions

### Search Database Objects

Quickly find tables, views, or functions:

```
Press '1' → Focus Connections panel
Press 's' → Open Schema Explorer
Press '/' → Enter search mode
Type 'user' → Filters to show only objects containing 'user'
  🔍 Search: user (5 matches)

  ▼ 📂 public
    ▼ 📊 Tables (2)
      ▶ 📋 users
      ▶ 📋 user_sessions

Use j/k → Navigate filtered results
Press 'p' → Preview selected table
Press Esc → Clear search and show all objects
```

**Search features:**
- Case-insensitive substring matching
- Real-time filtering as you type
- Shows match count
- Preserves tree structure
- Works on schemas, tables, views, functions, and columns

### Refresh Schema Data

Reload schema from database to see latest changes:

```
Press 's' → Open Schema Explorer
Press 'r' → Refresh
See loading indicator → Schema reloads with current data
```

**When to use refresh:**
- After creating/dropping tables
- After database schema changes
- To apply new filter settings (like excluding system tables)
- When schema data seems stale

### Edit Complex Queries in Neovim

```
Press '2' → Focus Editor panel
Press Ctrl+E → Opens Neovim with your query
[Edit in Neovim, save and quit]
Query is automatically updated in LazyDB!
```

### View PostgreSQL Help

```
Press '?' or F1 → Open help dialog
Use ←/→ → Browse categories
Use ↑/↓ → Navigate queries
Press Enter → Copy query to editor
```

Help categories include:
- Database & Schema info
- Table structure
- Indexes
- Functions & Procedures
- Sequences
- Triggers
- Data Types
- Performance queries
- Users & Permissions

### Manage Connections

```
# Edit a connection
Press '1' → Focus Connections
Select connection with j/k
Press 'e' → Edit
Make changes
Press Enter → Save

# Delete a connection
Press '1' → Focus Connections
Select connection with j/k
Press 'd' → Delete
Press 'y' → Confirm
```

### Environment Organization

LazyDB groups connections by environment:

```
▼ 🟢 Development
  ▶ dev-local
  > dev-docker

▼ 🔵 Staging
  > staging-db

▼ 🔴 Production
  > prod-master
```

Navigate across all environments seamlessly with j/k!

## Query History

Every query you execute is automatically logged to:

```
~/.lazydb/queries/Development_2025-01.sql
~/.lazydb/queries/Staging_2025-01.sql
~/.lazydb/queries/Production_2025-01.sql
```

Format:
```sql
-- Executed on: 2025-01-15 14:30:45 (Development)
SELECT * FROM users WHERE active = true;
```

Perfect for auditing and tracking what was run in each environment!

## Tips & Tricks

### Keyboard Shortcuts Cheat Sheet

```
Global:
  q / Ctrl+C → Quit
  ? / F1     → Help
  1/2/3      → Jump to panel
  Tab        → Next panel

Connections:
  j/k   → Navigate
  Enter → Connect
  a     → Add
  e     → Edit
  d     → Delete
  s     → Schema Explorer

Schema Explorer:
  j/k          → Navigate
  Enter/Space  → Expand/collapse
  /            → Search mode
  r            → Refresh schema
  p            → Preview table
  Esc          → Exit search / Back

Search Mode:
  [type]       → Filter results
  Backspace    → Delete character
  j/k          → Navigate results
  Esc          → Clear search

Editor:
  Ctrl+R → Execute
  Ctrl+E → Open Neovim
  F2     → Save query

Results:
  j/k → Scroll vertical
  h/l → Scroll horizontal
```

### Workflow Example

Here's a typical workflow:

1. **Morning**: Launch LazyDB
2. **Connect**: Press `1`, select connection, `Enter`
3. **Explore Schema**: Press `s` to open schema explorer
4. **Navigate**: Use j/k to browse schemas and tables
5. **Preview Table**: Navigate to a table, press `p` to preview
6. **Execute**: Press `Ctrl+R` to run the preview query
7. **Browse Results**: Press `3`, use j/k to scroll
8. **Complex Query**: Press `2`, `Ctrl+E`, write query in Neovim
9. **New Connection**: Press `1`, `a`, add staging database
10. **Switch**: Use j/k to select staging, `Enter` to connect
11. **Help**: Press `?` to find common PostgreSQL queries
12. **Copy & Run**: Navigate to query, `Enter` to copy, `Ctrl+R` to execute

### Security Best Practices

- Passwords are automatically encrypted (AES-256-GCM)
- Connection files have restrictive permissions (0600)
- Query history is plain text - be careful with sensitive data
- For production credentials, consider using environment variables

### Troubleshooting

**Can't connect to database?**
- Check host/port/credentials
- Verify PostgreSQL is running
- Check firewall rules
- Try `disable` for SSL Mode on local connections

**Neovim not working?**
- Install Neovim: `brew install neovim` (macOS) or `apt install neovim` (Linux)
- LazyDB will fall back to basic editor if Neovim isn't found

**Navigation not working?**
- Make sure you're in the right panel (press 1/2/3)
- Check if a dialog is open (press Esc to close)

## Next Steps

Now that you're up and running:

1. ✅ Add all your database connections
2. ✅ Organize them by environment
3. ✅ Explore the help dialog (`?`) for PostgreSQL tips
4. ✅ Set up Neovim for powerful query editing
5. ✅ Check your query history files for audit trails

## Need More Help?

- Full documentation: [README.md](README.md)
- Keybindings reference: Press `?` in LazyDB
- Report issues: [GitHub Issues](https://github.com/yourusername/lazydb/issues)

Happy querying! 🚀
