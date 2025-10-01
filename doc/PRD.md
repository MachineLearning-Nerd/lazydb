# 📋 PRD: LazyDB - Standalone Terminal Database Browser

## Executive Summary

**Product Name**: LazyDB
**Tagline**: "A lazier way to manage databases"
**Description**: A cross-platform, standalone TUI (Terminal User Interface) database management tool written in Go, inspired by lazygit/lazydocker, with embedded Neovim for query editing.

---

## Vision & Goals

### Primary Vision
Create the most intuitive, keyboard-driven database client for developers who live in the terminal, combining the speed of a TUI with the power of Neovim for SQL editing.

### Core Goals
1. **Zero Configuration**: Works out of the box, connects in seconds
2. **Keyboard-First**: 100% keyboard-driven workflow (mouse optional)
3. **Neovim Integration**: Full Neovim power for query editing
4. **Beautiful UI**: Modern, colorful, informative terminal interface
5. **Cross-Platform**: Single binary for macOS, Linux, Windows
6. **Multi-Database**: PostgreSQL, MySQL, SQLite support from day one

---

## Architecture Overview

### Technology Stack

**Core Application**:
- **Language**: Go 1.21+
- **TUI Framework**: Bubbletea (Elm architecture) + Bubbles (components) + Lipgloss (styling)
- **Alternative**: gocui (lazygit-style) if more control needed
- **Database Drivers**: pgx (PostgreSQL), go-sql-driver (MySQL), mattn/go-sqlite3

**Editor Integration**:
- **Editor**: Neovim (spawned as subprocess)
- **Communication**: RPC via msgpack, or simple file-based exchange
- **Fallback**: Built-in basic SQL editor if Neovim not available

**Distribution**:
- **Packaging**: GoReleaser for cross-platform binaries
- **Installation**: Homebrew, apt/yum repos, cargo install, direct download
- **Size Target**: <10MB binary

### Application Structure

```
lazydb/
├── cmd/
│   └── lazydb/
│       └── main.go              # Entry point
├── internal/
│   ├── app/
│   │   ├── app.go               # Main application controller
│   │   └── state.go             # Global application state
│   ├── ui/
│   │   ├── panels/
│   │   │   ├── connections.go   # Left panel: connection tree
│   │   │   ├── editor.go        # Center: Neovim integration
│   │   │   └── results.go       # Right: query results table
│   │   ├── components/
│   │   │   ├── table.go         # Reusable table widget
│   │   │   ├── tree.go          # Tree view widget
│   │   │   ├── statusbar.go     # Bottom status bar
│   │   │   └── modal.go         # Modal dialogs
│   │   └── theme.go             # Color scheme and styling
│   ├── db/
│   │   ├── connection.go        # Connection management
│   │   ├── postgres.go          # PostgreSQL driver
│   │   ├── mysql.go             # MySQL driver
│   │   ├── sqlite.go            # SQLite driver
│   │   └── query.go             # Query execution engine
│   ├── editor/
│   │   ├── nvim.go              # Neovim spawning and communication
│   │   └── fallback.go          # Basic editor fallback
│   ├── storage/
│   │   ├── config.go            # Load/save config (TOML)
│   │   ├── history.go           # Query history (SQLite)
│   │   └── connections.go       # Saved connections (encrypted)
│   └── keybindings/
│       └── keys.go              # Global keyboard shortcuts
├── assets/
│   ├── logo.txt                 # ASCII art logo
│   └── themes/                  # Color themes
└── README.md
```

---

## UI Design (Lazygit-Inspired)

### Screen Layout

```
┌────────────────────────────────────────────────────────────────────────────┐
│ LazyDB v1.0.0                      [dev] postgres@localhost      [?] Help  │
├───────────────────┬────────────────────────────────┬────────────────────────┤
│ CONNECTIONS       │  QUERY EDITOR                  │  RESULTS               │
│                   │                                │                        │
│ ▼ 🟢 Development  │  -- Press 'e' to edit in Nvim  │  ╔═══╦═══════╦═══════╗│
│   • dev-local ✓   │  -- Or type here directly      │  ║ id║ name  ║ email ║│
│   • dev-docker    │                                │  ╠═══╬═══════╬═══════╣│
│                   │  SELECT * FROM users           │  ║ 1 ║ Alice ║ a@... ║│
│ ▼ 🔵 Staging      │  WHERE active = true           │  ║ 2 ║ Bob   ║ b@... ║│
│   • staging-db    │  LIMIT 10;                     │  ╚═══╩═══════╩═══════╝│
│                   │                                │                        │
│ ▶ 🔴 Production   │                                │  10 rows (42ms)        │
│   • prod-master   │                                │                        │
│   • prod-replica  │                                │  [j/k] scroll          │
│                   │                                │  [y] copy              │
│ ──────────────────│                                │  [e] export            │
│ [a] add           │  [Ctrl-E] Edit in Neovim       │                        │
│ [d] delete        │  [Ctrl-R] Execute              │                        │
│ [e] edit          │  [Ctrl-S] Save query           │                        │
│ [Enter] connect   │  [Ctrl-L] Load query           │                        │
├───────────────────┴────────────────────────────────┴────────────────────────┤
│ ▸ Connected: dev-local @ localhost:5432/mydb      [1][2][3] Focus Panel    │
└────────────────────────────────────────────────────────────────────────────┘
```

---

## Core Features

See PRD sections above for details on:
- Connection Management
- Query Execution
- Neovim Integration
- Query History & Library
- Schema Explorer
- Export & Import
- Configuration

---

## Development Phases

### Phase 1: MVP (6-8 weeks)
Basic functional TUI with PostgreSQL support and Neovim integration

### Phase 2: Enhanced (4-6 weeks)
Multi-database support and query management

### Phase 3: Advanced (4-6 weeks)
Professional features, export, themes, and distribution

### Phase 4: Future
Visual query builder, ER diagrams, plugins, cloud database support

---

## Key Differentiators

**vs. DBeaver**: Faster, keyboard-driven, lightweight, terminal-native
**vs. psql/mysql CLI**: Visual TUI, easier, multi-database, better results
**vs. Gobang/lazysql**: Neovim integration, better UX, richer features

---

## Success Metrics

- **Adoption**: 1000+ GitHub stars in 3 months, 10,000+ downloads in 6 months
- **Quality**: <0.1% crash rate, <100ms startup time, >80% test coverage
- **UX**: <60s time to first query, 100% keyboard-only workflow

---

## Next Steps

1. Create project repository
2. Setup Go module
3. Create all documentation files
4. Prototype UI with Bubbletea
5. Test Neovim spawning

---

For complete details, see full PRD sections above.
