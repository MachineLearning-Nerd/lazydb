# 🎨 LazyDB UI Mockups

Comprehensive ASCII art mockups for all LazyDB screens and interactions.

---

## Main Application Screen

### Default View (No Connection)

```
┌────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ LazyDB v1.0.0                                                                                        [?] Help  [q] Quit    │
├──────────────────────────────┬────────────────────────────────────────────────────┬──────────────────────────────────────────┤
│                              │                                                    │                                          │
│  CONNECTIONS                 │  QUERY EDITOR                                      │  RESULTS                                 │
│                              │                                                    │                                          │
│  ▼ 🟢 Development            │  ┌───────────────────────────────────────────────┐│  ╔════════════════════════════════════╗ │
│    • dev-local               │  │ -- No connection active                       ││  ║  No results to display             ║ │
│    • dev-docker              │  │ -- Connect to a database first                ││  ║                                    ║ │
│                              │  │                                               ││  ║  Connect to a database and         ║ │
│  ▶ 🔵 Staging                │  │                                               ││  ║  execute a query to see results    ║ │
│    • staging-db              │  │                                               ││  ║                                    ║ │
│                              │  │                                               ││  ╚════════════════════════════════════╝ │
│  ▶ 🔴 Production             │  │                                               ││                                          │
│    • prod-master             │  │                                               ││                                          │
│    • prod-replica            │  └───────────────────────────────────────────────┘│                                          │
│                              │                                                    │                                          │
│  ─────────────────────────── │  Press Ctrl-E to edit in Neovim                   │                                          │
│  No connection active        │  Or start typing SQL here                         │                                          │
│                              │                                                    │                                          │
│  [a] Add connection          │  [Ctrl-R] Execute                                 │  [No keybindings available]              │
│  [d] Delete                  │  [Ctrl-S] Save query                              │                                          │
│  [e] Edit                    │  [Ctrl-L] Load query                              │                                          │
│  [Enter] Connect             │  [Ctrl-H] History                                 │                                          │
│                              │                                                    │                                          │
├──────────────────────────────┴────────────────────────────────────────────────────┴──────────────────────────────────────────┤
│  ▸ Not connected • Press Enter on a connection to connect                                      [1][2][3] Switch Panel      │
└────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

### Active Connection with Query

```
┌────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ LazyDB v1.0.0                                                     [dev] dev-local @ localhost:5432/mydb         [?] Help   │
├──────────────────────────────┬────────────────────────────────────────────────────┬──────────────────────────────────────────┤
│                              │                                                    │                                          │
│  CONNECTIONS                 │  QUERY EDITOR                                      │  RESULTS                                 │
│                              │                                                    │                                          │
│  ▼ 🟢 Development            │  ┌───────────────────────────────────────────────┐│  ╔═════╦══════════════╦═══════════════╗ │
│    ▸ dev-local ✓             │  │ 1  SELECT id, name, email, created_at         ││  ║  id ║ name         ║ email         ║ │
│      ▼ Tables (45)           │  │ 2  FROM users                                 ││  ╠═════╬══════════════╬═══════════════╣ │
│        • users               │  │ 3  WHERE active = true                        ││  ║  1  ║ Alice Smith  ║ alice@ex.com  ║ │
│        • posts               │  │ 4  ORDER BY created_at DESC                   ││  ║  2  ║ Bob Jones    ║ bob@exam.com  ║ │
│        • comments            │  │ 5  LIMIT 10;                                  ││  ║  3  ║ Carol White  ║ carol@ex.com  ║ │
│        • ...                 │  │ 6  █                                          ││  ║  4  ║ David Brown  ║ david@ex.com  ║ │
│      ▶ Views (3)             │  │                                               ││  ║  5  ║ Eva Green    ║ eva@examp.com ║ │
│      ▶ Functions (12)        │  └───────────────────────────────────────────────┘│  ║  6  ║ Frank Miller ║ frank@ex.com  ║ │
│    • dev-docker              │                                                    │  ║  7  ║ Grace Lee    ║ grace@ex.com  ║ │
│                              │  Ln 6, Col 1                                       │  ║  8  ║ Henry Davis  ║ henry@ex.com  ║ │
│  ▶ 🔵 Staging                │                                                    │  ║  9  ║ Iris Wilson  ║ iris@exam.com ║ │
│  ▶ 🔴 Production             │  [Ctrl-E] Edit in Neovim                           │  ║  10 ║ Jack Taylor  ║ jack@exam.com ║ │
│                              │  [Ctrl-R] Execute                                  │  ╚═════╩══════════════╩═══════════════╝ │
│  ─────────────────────────── │  [Ctrl-S] Save query                               │                                          │
│  Connected: dev-local        │  [Ctrl-L] Load query                               │  10 rows in 42ms                         │
│  Host: localhost:5432        │                                                    │                                          │
│  Database: mydb              │                                                    │  [j/k] Scroll rows                       │
│                              │                                                    │  [h/l] Scroll columns                    │
│  [r] Refresh schema          │                                                    │  [y] Copy cell                           │
│  [p] Preview data            │                                                    │  [yy] Copy row                           │
│  [d] Disconnect              │                                                    │  [e] Export                              │
│                              │                                                    │                                          │
├──────────────────────────────┴────────────────────────────────────────────────────┴──────────────────────────────────────────┤
│  ● Connected: dev-local @ localhost:5432/mydb • 10 rows returned                                [1][2][3] Switch Panel      │
└────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Connection Form Modal

### Add New Connection

```
┌────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ LazyDB v1.0.0                                                                                        [?] Help  [q] Quit    │
├──────────────────────────────┬────────────────────────────────────────────────────┬──────────────────────────────────────────┤
│                              │ ┌───────────────────────────────────────────────┐ │                                          │
│  CONNECTIONS                 │ │  Add New Connection                           │ │  RESULTS                                 │
│                              │ │                                               │ │                                          │
│  ▼ 🟢 Development            │ │  ┌─────────────────────────────────────────┐ │ │                                          │
│    • dev-local               │ │  │ Name:        [                         ]│ │ │                                          │
│    • dev-docker              │ │  │              ▲                          │ │ │                                          │
│                              │ │  │ Type:        [PostgreSQL ▼]             │ │ │                                          │
│  ▶ 🔵 Staging                │ │  │              │ PostgreSQL               │ │ │                                          │
│    • staging-db              │ │  │              │ MySQL                    │ │ │                                          │
│                              │ │  │              │ SQLite                   │ │ │                                          │
│  ▶ 🔴 Production             │ │  │              └──────────────────────────│ │ │                                          │
│    • prod-master             │ │  │ Host:        [localhost               ]│ │ │                                          │
│    • prod-replica            │ │  │ Port:        [5432                    ]│ │ │                                          │
│                              │ │  │ Database:    [                         ]│ │ │                                          │
│  ─────────────────────────── │ │  │ Username:    [                         ]│ │ │                                          │
│  [a] Add connection          │ │  │ Password:    [•••••••••••••••••••••••]│ │ │                                          │
│  [d] Delete                  │ │  │ SSL Mode:    [Prefer ▼]                │ │ │                                          │
│  [e] Edit                    │ │  │ Environment: [Development ▼]           │ │ │                                          │
│  [Enter] Connect             │ │  └─────────────────────────────────────────┘ │ │                                          │
│                              │ │                                               │ │                                          │
│                              │ │  [Tab] Next field    [Shift-Tab] Previous    │ │                                          │
│                              │ │  [Ctrl-T] Test       [Ctrl-S] Save           │ │                                          │
│                              │ │  [Esc] Cancel                                 │ │                                          │
│                              │ └───────────────────────────────────────────────┘ │                                          │
│                              │                                                    │                                          │
├──────────────────────────────┴────────────────────────────────────────────────────┴──────────────────────────────────────────┤
│  ▸ Not connected • Fill in connection details and press Ctrl-S to save                         [1][2][3] Switch Panel      │
└────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Query History Modal

```
┌────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ LazyDB v1.0.0                                                     [dev] dev-local @ localhost:5432/mydb         [?] Help   │
├──────────────────────────────┬────────────────────────────────────────────────────┬──────────────────────────────────────────┤
│                              │ ┌────────────────────────────────────────────────────────────────────────────────────────┐ │
│  CONNECTIONS                 │ │  Query History                                                      [Ctrl-H] Close     │ │
│                              │ │                                                                                        │ │
│  ▼ 🟢 Development            │ │  Search: [                                                                       ] 🔍 │ │
│    ▸ dev-local ✓             │ │                                                                                        │ │
│      ▼ Tables (45)           │ │  ┌──────────────────────────────────────────────────────────────────────────────────┐ │ │
│        • users               │ │  │ ▸ SELECT * FROM users WHERE active = true LIMIT 10;                              │ │ │
│        • posts               │ │  │   dev-local • 2024-01-15 14:23:45 • 42ms • 10 rows                              │ │ │
│        • comments            │ │  │                                                                                  │ │ │
│        • ...                 │ │  │ • UPDATE users SET last_login = NOW() WHERE id = 123;                            │ │ │
│      ▶ Views (3)             │ │  │   dev-local • 2024-01-15 14:15:22 • 15ms • 1 row                                │ │ │
│      ▶ Functions (12)        │ │  │                                                                                  │ │ │
│    • dev-docker              │ │  │ • INSERT INTO posts (title, content) VALUES ('Test', 'Content');                 │ │ │
│                              │ │  │   dev-local • 2024-01-15 13:45:10 • 8ms • 1 row                                 │ │ │
│  ▶ 🔵 Staging                │ │  │                                                                                  │ │ │
│  ▶ 🔴 Production             │ │  │ • DELETE FROM comments WHERE spam = true;                                        │ │ │
│                              │ │  │   dev-local • 2024-01-15 12:30:05 • 125ms • 42 rows                             │ │ │
│  ─────────────────────────── │ │  │                                                                                  │ │ │
│  Connected: dev-local        │ │  │ • SELECT COUNT(*) FROM users;                                                    │ │ │
│  Host: localhost:5432        │ │  │   dev-local • 2024-01-15 11:15:33 • 5ms • 1 row                                 │ │ │
│  Database: mydb              │ │  └──────────────────────────────────────────────────────────────────────────────────┘ │ │
│                              │ │                                                                                        │ │
│                              │ │  [j/k] Navigate  [Enter] Load query  [d] Delete  [/] Search  [Esc] Close             │ │
│                              │ └────────────────────────────────────────────────────────────────────────────────────────┘ │
│                              │                                                                    │                        │
├──────────────────────────────┴────────────────────────────────────────────────────┴──────────────────────────────────────────┤
│  ● Connected: dev-local @ localhost:5432/mydb • Viewing query history                          [1][2][3] Switch Panel      │
└────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Query Library Modal

```
┌────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ LazyDB v1.0.0                                                     [dev] dev-local @ localhost:5432/mydb         [?] Help   │
├──────────────────────────────┬────────────────────────────────────────────────────┬──────────────────────────────────────────┤
│                              │ ┌────────────────────────────────────────────────────────────────────────────────────────┐ │
│  CONNECTIONS                 │ │  Query Library                                                      [Ctrl-L] Close     │ │
│                              │ │                                                                                        │ │
│  ▼ 🟢 Development            │ │  Filter: [All ▼]  Tags: [              ]  Search: [                              ] 🔍│ │
│    ▸ dev-local ✓             │ │                                                                                        │ │
│      ▼ Tables (45)           │ │  ┌──────────────────────────────────────────────────────────────────────────────────┐ │ │
│        • users               │ │  │ ▸ 📌 Active Users Report                                                         │ │ │
│        • posts               │ │  │    Tags: [reports] [users]                                                       │ │ │
│        • comments            │ │  │    SELECT * FROM users WHERE active = true ORDER BY created_at DESC;             │ │ │
│        • ...                 │ │  │                                                                                  │ │ │
│      ▶ Views (3)             │ │  │ • User Login Update                                                              │ │ │
│      ▶ Functions (12)        │ │  │   Tags: [users] [maintenance]                                                    │ │ │
│    • dev-docker              │ │  │   UPDATE users SET last_login = NOW() WHERE id = $1;                             │ │ │
│                              │ │  │                                                                                  │ │ │
│  ▶ 🔵 Staging                │ │  │ • 📌 Post Count by Author                                                        │ │ │
│  ▶ 🔴 Production             │ │  │   Tags: [reports] [analytics]                                                    │ │ │
│                              │ │  │   SELECT author_id, COUNT(*) as post_count...                                    │ │ │
│  ─────────────────────────── │ │  │                                                                                  │ │ │
│  Connected: dev-local        │ │  │ • Clean Spam Comments                                                            │ │ │
│  Host: localhost:5432        │ │  │   Tags: [cleanup] [comments]                                                     │ │ │
│  Database: mydb              │ │  │   DELETE FROM comments WHERE spam = true;                                        │ │ │
│                              │ │  └──────────────────────────────────────────────────────────────────────────────────┘ │ │
│                              │ │                                                                                        │ │
│                              │ │  [j/k] Navigate  [Enter] Load  [d] Delete  [f] Toggle favorite  [Esc] Close          │ │
│                              │ └────────────────────────────────────────────────────────────────────────────────────────┘ │
│                              │                                                                    │                        │
├──────────────────────────────┴────────────────────────────────────────────────────┴──────────────────────────────────────────┤
│  ● Connected: dev-local @ localhost:5432/mydb • Browsing saved queries                         [1][2][3] Switch Panel      │
└────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Help Modal

```
┌────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ LazyDB v1.0.0                                                     [dev] dev-local @ localhost:5432/mydb         [?] Help   │
├──────────────────────────────┬────────────────────────────────────────────────────┬──────────────────────────────────────────┤
│                              │ ┌────────────────────────────────────────────────────────────────────────────────────────┐ │
│  CONNECTIONS                 │ │  LazyDB Keyboard Shortcuts                                       [?] or [Esc] Close   │ │
│                              │ │                                                                                        │ │
│  ▼ 🟢 Development            │ │  ╔═══════════════════════════════════════════════════════════════════════════════════╗│ │
│    ▸ dev-local ✓             │ │  ║ GLOBAL                                                                            ║│ │
│      ▼ Tables (45)           │ │  ╠═══════════════════════════════════════════════════════════════════════════════════╣│ │
│        • users               │ │  ║  q             Quit application                                                   ║│ │
│        • posts               │ │  ║  ?             Show/hide this help                                                ║│ │
│        • comments            │ │  ║  1, 2, 3       Focus connections, editor, results panel                          ║│ │
│        • ...                 │ │  ║  Tab           Next panel                                                         ║│ │
│      ▶ Views (3)             │ │  ║  Shift-Tab     Previous panel                                                     ║│ │
│      ▶ Functions (12)        │ │  ║                                                                                   ║│ │
│    • dev-docker              │ │  ╠═══════════════════════════════════════════════════════════════════════════════════╣│ │
│                              │ │  ║ CONNECTIONS PANEL                                                                 ║│ │
│  ▶ 🔵 Staging                │ │  ╠═══════════════════════════════════════════════════════════════════════════════════╣│ │
│  ▶ 🔴 Production             │ │  ║  j/k or ↑/↓    Navigate up/down                                                  ║│ │
│                              │ │  ║  Enter         Connect to database / Expand group                                ║│ │
│  ─────────────────────────── │ │  ║  a             Add new connection                                                 ║│ │
│  Connected: dev-local        │ │  ║  e             Edit selected connection                                           ║│ │
│  Host: localhost:5432        │ │  ║  d             Delete selected connection                                         ║│ │
│  Database: mydb              │ │  ║  t             Test connection                                                    ║│ │
│                              │ │  ║  r             Refresh schema                                                     ║│ │
│                              │ │  ║  /             Search connections                                                 ║│ │
│                              │ │  ║  p             Preview table data (when table selected)                          ║│ │
│                              │ │  ║                                                                                   ║│ │
│                              │ │  ╠═══════════════════════════════════════════════════════════════════════════════════╣│ │
│                              │ │  ║ QUERY EDITOR PANEL                                                                ║│ │
│                              │ │  ╠═══════════════════════════════════════════════════════════════════════════════════╣│ │
│                              │ │  ║  Ctrl-E        Edit in Neovim (full screen)                                      ║│ │
│                              │ │  ║  Ctrl-R        Execute query                                                      ║│ │
│                              │ │  ║  Ctrl-S        Save query to library                                              ║│ │
│                              │ │  ║  Ctrl-L        Load query from library                                            ║│ │
│                              │ │  ║  Ctrl-H        View query history                                                 ║│ │
│                              │ │  ║  Ctrl-F        Format SQL query                                                   ║│ │
│                              │ │  ║  Ctrl-T        Insert query template                                              ║│ │
│                              │ │  ║  Ctrl-C        Cancel running query                                               ║│ │
│                              │ │  ║                                                                                   ║│ │
│                              │ │  ╠═══════════════════════════════════════════════════════════════════════════════════╣│ │
│                              │ │  ║ RESULTS PANEL                                                                     ║│ │
│                              │ │  ╠═══════════════════════════════════════════════════════════════════════════════════╣│ │
│                              │ │  ║  j/k or ↑/↓    Scroll rows                                                        ║│ │
│                              │ │  ║  h/l or ←/→    Scroll columns                                                     ║│ │
│                              │ │  ║  g / G         Go to first/last row                                               ║│ │
│                              │ │  ║  y             Copy cell to clipboard                                             ║│ │
│                              │ │  ║  yy            Copy entire row                                                    ║│ │
│                              │ │  ║  yc            Copy column                                                        ║│ │
│                              │ │  ║  ya            Copy all results                                                   ║│ │
│                              │ │  ║  e             Export results (CSV, JSON, SQL)                                    ║│ │
│                              │ │  ║  s             Sort by current column                                             ║│ │
│                              │ │  ║  w             Toggle word wrap                                                   ║│ │
│                              │ │  ║  /             Search in results                                                  ║│ │
│                              │ │  ╚═══════════════════════════════════════════════════════════════════════════════════╝│ │
│                              │ └────────────────────────────────────────────────────────────────────────────────────────┘ │
│                              │                                                                    │                        │
├──────────────────────────────┴────────────────────────────────────────────────────┴──────────────────────────────────────────┤
│  ● Connected: dev-local @ localhost:5432/mydb • Press ? to close help                          [1][2][3] Switch Panel      │
└────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Error States

### Connection Failed

```
┌────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ LazyDB v1.0.0                                                                                        [?] Help  [q] Quit    │
├──────────────────────────────┬────────────────────────────────────────────────────┬──────────────────────────────────────────┤
│                              │                                                    │                                          │
│  CONNECTIONS                 │  QUERY EDITOR                                      │  RESULTS                                 │
│                              │                                                    │                                          │
│  ▼ 🟢 Development            │  ┌───────────────────────────────────────────────┐│                                          │
│    ▸ dev-local ✗             │  │ -- No connection active                       ││                                          │
│    • dev-docker              │  │ -- Connect to a database first                ││                                          │
│                              │  │                                               ││                                          │
│  ▶ 🔵 Staging                │  │                                               ││                                          │
│  ▶ 🔴 Production             │  │                                               ││                                          │
│                              │  └───────────────────────────────────────────────┘│                                          │
│  ─────────────────────────── │                                                    │                                          │
│  ╔═══════════════════════╗   │                                                    │                                          │
│  ║ ❌ Connection Failed  ║   │                                                    │                                          │
│  ╠═══════════════════════╣   │                                                    │                                          │
│  ║ Failed to connect to  ║   │                                                    │                                          │
│  ║ dev-local             ║   │                                                    │                                          │
│  ║                       ║   │                                                    │                                          │
│  ║ Error:                ║   │                                                    │                                          │
│  ║ Connection refused    ║   │                                                    │                                          │
│  ║ (localhost:5432)      ║   │                                                    │                                          │
│  ║                       ║   │                                                    │                                          │
│  ║ Possible causes:      ║   │                                                    │                                          │
│  ║ • DB server not running│  │                                                    │                                          │
│  ║ • Wrong host/port     ║   │                                                    │                                          │
│  ║ • Firewall blocking   ║   │                                                    │                                          │
│  ║                       ║   │                                                    │                                          │
│  ║ [r] Retry             ║   │                                                    │                                          │
│  ║ [e] Edit connection   ║   │                                                    │                                          │
│  ║ [Esc] Close           ║   │                                                    │                                          │
│  ╚═══════════════════════╝   │                                                    │                                          │
│                              │                                                    │                                          │
├──────────────────────────────┴────────────────────────────────────────────────────┴──────────────────────────────────────────┤
│  ✗ Connection failed: dev-local • Check host and port                                          [1][2][3] Switch Panel      │
└────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

### Query Error

```
┌────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ LazyDB v1.0.0                                                     [dev] dev-local @ localhost:5432/mydb         [?] Help   │
├──────────────────────────────┬────────────────────────────────────────────────────┬──────────────────────────────────────────┤
│                              │                                                    │                                          │
│  CONNECTIONS                 │  QUERY EDITOR                                      │  RESULTS                                 │
│                              │                                                    │                                          │
│  ▼ 🟢 Development            │  ┌───────────────────────────────────────────────┐│  ╔════════════════════════════════════╗ │
│    ▸ dev-local ✓             │  │ 1  SELCT id, name, email                      ││  ║  ❌ Query Failed                   ║ │
│      ▼ Tables (45)           │  │ 2  FROM users                                 ││  ║                                    ║ │
│        • users               │  │ 3  WHERE active = true;                       ││  ║  ERROR: syntax error at or near    ║ │
│        • posts               │  │ 4  █                                          ││  ║  "SELCT"                           ║ │
│        • comments            │  │                                               ││  ║                                    ║ │
│        • ...                 │  │  ⚠️  Line 1, Column 1                         ││  ║  LINE 1: SELCT id, name, email     ║ │
│      ▶ Views (3)             │  │      ^                                        ││  ║          ^                         ║ │
│      ▶ Functions (12)        │  │      Syntax error near "SELCT"               ││  ║                                    ║ │
│    • dev-docker              │  └───────────────────────────────────────────────┘│  ║  Hint: Did you mean "SELECT"?      ║ │
│                              │                                                    │  ║                                    ║ │
│  ▶ 🔵 Staging                │  Ln 4, Col 1                                       │  ║  Query executed in 5ms             ║ │
│  ▶ 🔴 Production             │                                                    │  ╚════════════════════════════════════╝ │
│                              │  [Ctrl-E] Edit in Neovim                           │                                          │
│  ─────────────────────────── │  [Ctrl-R] Execute                                  │                                          │
│  Connected: dev-local        │  [Ctrl-S] Save query                               │                                          │
│  Host: localhost:5432        │  [Ctrl-L] Load query                               │                                          │
│  Database: mydb              │                                                    │                                          │
│                              │                                                    │                                          │
├──────────────────────────────┴────────────────────────────────────────────────────┴──────────────────────────────────────────┤
│  ✗ Query failed: Syntax error at or near "SELCT"                                               [1][2][3] Switch Panel      │
└────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Loading States

### Query Executing

```
┌────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ LazyDB v1.0.0                                                     [dev] dev-local @ localhost:5432/mydb         [?] Help   │
├──────────────────────────────┬────────────────────────────────────────────────────┬──────────────────────────────────────────┤
│                              │                                                    │                                          │
│  CONNECTIONS                 │  QUERY EDITOR                                      │  RESULTS                                 │
│                              │                                                    │                                          │
│  ▼ 🟢 Development            │  ┌───────────────────────────────────────────────┐│  ╔════════════════════════════════════╗ │
│    ▸ dev-local ✓             │  │ 1  SELECT *                                   ││  ║                                    ║ │
│      ▼ Tables (45)           │  │ 2  FROM large_table                           ││  ║   ⏳ Executing query...            ║ │
│        • users               │  │ 3  WHERE created_at > '2024-01-01';           ││  ║                                    ║ │
│        • posts               │  │ 4  █                                          ││  ║      ▓▓▓▓▓▓▓░░░░░░░░░░  45%       ║ │
│        • comments            │  │                                               ││  ║                                    ║ │
│        • ...                 │  └───────────────────────────────────────────────┘│  ║   12,450 rows fetched              ║ │
│      ▶ Views (3)             │                                                    │  ║   Elapsed: 2.3s                    ║ │
│      ▶ Functions (12)        │  Ln 4, Col 1                                       │  ║                                    ║ │
│    • dev-docker              │                                                    │  ║   [Ctrl-C] Cancel query            ║ │
│                              │  [Ctrl-E] Edit in Neovim                           │  ║                                    ║ │
│  ▶ 🔵 Staging                │  [Ctrl-C] Cancel query                             │  ╚════════════════════════════════════╝ │
│  ▶ 🔴 Production             │  [Ctrl-S] Save query                               │                                          │
│                              │  [Ctrl-L] Load query                               │                                          │
│  ─────────────────────────── │                                                    │                                          │
│  Connected: dev-local        │                                                    │                                          │
│  Host: localhost:5432        │                                                    │                                          │
│  Database: mydb              │                                                    │                                          │
│                              │                                                    │                                          │
├──────────────────────────────┴────────────────────────────────────────────────────┴──────────────────────────────────────────┤
│  ⏳ Executing query... 12,450 rows fetched • 2.3s elapsed • Press Ctrl-C to cancel              [1][2][3] Switch Panel     │
└────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Export Dialog

```
┌────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ LazyDB v1.0.0                                                     [dev] dev-local @ localhost:5432/mydb         [?] Help   │
├──────────────────────────────┬────────────────────────────────────────────────────┬──────────────────────────────────────────┤
│                              │                                                    │                                          │
│  CONNECTIONS                 │  QUERY EDITOR                                      │  RESULTS                                 │
│                              │                                                    │                                          │
│  ▼ 🟢 Development            │  ┌───────────────────────────────┐                │  ╔═════╦══════════════╦═══════════════╗ │
│    ▸ dev-local ✓             │  │  Export Results               │                │  ║  id ║ name         ║ email         ║ │
│      ▼ Tables (45)           │  │                               │                │  ╠═════╬══════════════╬═══════════════╣ │
│        • users               │  │  Format:                      │                │  ║  1  ║ Alice Smith  ║ alice@ex.com  ║ │
│        • posts               │  │  ( ) CSV (with headers)       │                │  ║  2  ║ Bob Jones    ║ bob@exam.com  ║ │
│        • comments            │  │  ( ) CSV (no headers)         │                │  ║  3  ║ Carol White  ║ carol@ex.com  ║ │
│        • ...                 │  │  (•) JSON (array)             │                │  ║  4  ║ David Brown  ║ david@ex.com  ║ │
│      ▶ Views (3)             │  │  ( ) JSON (newline-delimited)│                │  ║  5  ║ Eva Green    ║ eva@examp.com ║ │
│      ▶ Functions (12)        │  │  ( ) SQL INSERT statements   │                │  ║  6  ║ Frank Miller ║ frank@ex.com  ║ │
│    • dev-docker              │  │  ( ) Markdown table          │                │  ║  7  ║ Grace Lee    ║ grace@ex.com  ║ │
│                              │  │                               │                │  ║  8  ║ Henry Davis  ║ henry@ex.com  ║ │
│  ▶ 🔵 Staging                │  │  Destination:                 │                │  ║  9  ║ Iris Wilson  ║ iris@exam.com ║ │
│  ▶ 🔴 Production             │  │  ( ) Clipboard                │                │  ║  10 ║ Jack Taylor  ║ jack@exam.com ║ │
│                              │  │  (•) File: [users.json]      │                │  ╚═════╩══════════════╩═══════════════╝ │
│  ─────────────────────────── │  │                               │                │                                          │
│  Connected: dev-local        │  │  [Enter] Export               │                │  10 rows in 42ms                         │
│  Host: localhost:5432        │  │  [Esc] Cancel                 │                │                                          │
│  Database: mydb              │  └───────────────────────────────┘                │                                          │
│                              │                                                    │                                          │
│                              │                                                    │                                          │
│                              │                                                    │                                          │
├──────────────────────────────┴────────────────────────────────────────────────────┴──────────────────────────────────────────┤
│  ● Connected: dev-local @ localhost:5432/mydb • Ready to export 10 rows                        [1][2][3] Switch Panel      │
└────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Large Result Set with Pagination

```
┌────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ LazyDB v1.0.0                                                     [dev] dev-local @ localhost:5432/mydb         [?] Help   │
├──────────────────────────────┬────────────────────────────────────────────────────┬──────────────────────────────────────────┤
│                              │                                                    │                                          │
│  CONNECTIONS                 │  QUERY EDITOR                                      │  RESULTS                                 │
│                              │                                                    │                                          │
│  ▼ 🟢 Development            │  ┌───────────────────────────────────────────────┐│  ╔═════╦══════════════╦═══════════════╗ │
│    ▸ dev-local ✓             │  │ 1  SELECT * FROM large_table;                 ││  ║  id ║ name         ║ created_at    ║ │
│      ▼ Tables (45)           │  │ 2  █                                          ││  ╠═════╬══════════════╬═══════════════╣ │
│        • users               │  │                                               ││  ║ 1001║ User 1001    ║ 2024-01-01... ║ │
│        • posts               │  │                                               ││  ║ 1002║ User 1002    ║ 2024-01-01... ║ │
│        • comments            │  │                                               ││  ║ 1003║ User 1003    ║ 2024-01-01... ║ │
│        • ...                 │  │                                               ││  ║ 1004║ User 1004    ║ 2024-01-01... ║ │
│      ▶ Views (3)             │  │                                               ││  ║ 1005║ User 1005    ║ 2024-01-01... ║ │
│      ▶ Functions (12)        │  │                                               ││  ║ ...                                ║ │
│    • dev-docker              │  └───────────────────────────────────────────────┘│  ║ 2000║ User 2000    ║ 2024-01-05... ║ │
│                              │                                                    │  ╚═════╩══════════════╩═══════════════╝ │
│  ▶ 🔵 Staging                │  Ln 2, Col 1                                       │                                          │
│  ▶ 🔴 Production             │                                                    │  Page 2 of 50 (1000 rows per page)      │
│                              │  [Ctrl-E] Edit in Neovim                           │  Total: 50,000 rows in 1.2s             │
│  ─────────────────────────── │  [Ctrl-R] Execute                                  │                                          │
│  Connected: dev-local        │  [Ctrl-S] Save query                               │  [j/k] Scroll  [n] Next page            │
│  Host: localhost:5432        │  [Ctrl-L] Load query                               │  [p] Previous  [g] First  [G] Last      │
│  Database: mydb              │                                                    │  [e] Export    [/] Search               │
│                              │                                                    │                                          │
├──────────────────────────────┴────────────────────────────────────────────────────┴──────────────────────────────────────────┤
│  ● Connected: dev-local @ localhost:5432/mydb • 50,000 rows • Page 2/50                        [1][2][3] Switch Panel      │
└────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

These mockups provide a comprehensive visual guide for all major UI screens and interactions in LazyDB, showing the full user experience from connection management to query execution and result viewing.
