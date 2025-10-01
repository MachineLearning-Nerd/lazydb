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
3. **Quick Query**: Press `2`, type `SELECT * FROM users LIMIT 10;`, `Ctrl+R`
4. **Complex Query**: Press `Ctrl+E`, write query in Neovim, save and quit
5. **Execute**: Query auto-loads, press `Ctrl+R`
6. **Browse Results**: Press `3`, use j/k to scroll
7. **New Connection**: Press `1`, `a`, add staging database
8. **Switch**: Use j/k to select staging, `Enter` to connect
9. **Help**: Press `?` to find the query for listing all tables
10. **Copy & Run**: Navigate to query, `Enter` to copy, `Ctrl+R` to execute

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
