# LazyDB 🚀

**A fast, keyboard-driven TUI (Terminal User Interface) for managing PostgreSQL databases with Neovim integration.**

LazyDB brings the power and convenience of terminal-based database management with vim-style navigation, environment-based connection organization, and automatic query logging.

![Version](https://img.shields.io/badge/version-1.0.0--alpha-blue)
![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8?logo=go)
![License](https://img.shields.io/badge/license-MIT-green)

## ✨ Features

### Core Features
- **🎯 3-Panel TUI**: Connections | Editor | Results
- **⌨️  Vim-style Navigation**: Full keyboard control with hjkl movement
- **🔌 PostgreSQL Support**: Connect and query PostgreSQL databases
- **📝 Neovim Integration**: Edit complex queries in your favorite editor (Ctrl+E)
- **🌍 Environment-Based Organization**: Organize connections by Development/Staging/Production
- **🔐 Encrypted Storage**: AES-256-GCM password encryption for connection credentials
- **📜 Automatic Query History**: Per-environment monthly query logs
- **❓ PostgreSQL Quick Reference**: Built-in help with common queries (Press `?`)

### Query Management
- ✅ Multi-line SQL editor
- ✅ Execute queries with `Ctrl+R`
- ✅ Auto-save query history by environment
- ✅ Syntax-highlighted results table
- ✅ Scrollable results with vim bindings

### Connection Management
- ✅ Add/Edit/Delete connections
- ✅ Group by environment (Dev/Staging/Prod)
- ✅ Visual connection status indicators
- ✅ Persistent connection storage
- ✅ Encrypted password storage

## 📦 Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/lazydb.git
cd lazydb

# Build
go build -o lazydb cmd/lazydb/main.go

# Install (optional)
sudo mv lazydb /usr/local/bin/
```

### Requirements

- Go 1.21 or higher
- PostgreSQL database (for testing)
- Neovim (optional, for advanced editing)

## 🚀 Quick Start

1. **Launch LazyDB**
   ```bash
   ./lazydb
   ```

2. **Add Your First Connection**
   - Press `1` to focus Connections panel
   - Press `a` to add a new connection
   - Fill in the form:
     - Name: `dev-local`
     - Host: `localhost`
     - Port: `5432`
     - Database: `postgres`
     - Username: `postgres`
     - Password: `your-password`
     - SSL Mode: `disable`
     - Environment: Use `←/→` to select `Development`
   - Press `Enter` to save

3. **Connect to Database**
   - Select your connection with `j/k`
   - Press `Enter` to connect

4. **Run Your First Query**
   - Press `2` to focus Editor panel
   - Type a query: `SELECT * FROM pg_database;`
   - Press `Ctrl+R` to execute
   - Results appear in the right panel

## ⌨️ Keybindings

### Global
| Key | Action |
|-----|--------|
| `1`, `2`, `3` | Focus panel (Connections, Editor, Results) |
| `Tab` | Cycle to next panel |
| `Shift+Tab` | Cycle to previous panel |
| `?` or `F1` | Open PostgreSQL help reference |
| `q` or `Ctrl+C` | Quit application |

### Connections Panel
| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Enter` | Connect to selected database |
| `a` | Add new connection |
| `e` | Edit selected connection |
| `d` | Delete selected connection |

### Editor Panel
| Key | Action |
|-----|--------|
| `Ctrl+R` | Execute query |
| `Ctrl+E` | Open in Neovim |
| `F2` | Save query to file |

### Results Panel
| Key | Action |
|-----|--------|
| `j` / `↓` | Scroll down |
| `k` / `↑` | Scroll up |
| `h` / `←` | Scroll left |
| `l` / `→` | Scroll right |

### Help Dialog
| Key | Action |
|-----|--------|
| `←/→` | Switch category |
| `↑/↓` | Navigate queries |
| `Enter` | Copy query to editor |
| `Esc` / `?` | Close help |

## 📁 File Structure

LazyDB stores its configuration and data in `~/.lazydb/`:

```
~/.lazydb/
├── connections.json          # Encrypted connection configs
└── queries/
    ├── Development_2025-01.sql    # January dev queries
    ├── Staging_2025-01.sql        # January staging queries
    └── Production_2025-01.sql     # January production queries
```

### Query History Format

Each executed query is automatically logged:

```sql
-- Executed on: 2025-01-15 14:30:45 (Development)
SELECT * FROM users WHERE active = true;

-- Executed on: 2025-01-15 15:22:10 (Development)
UPDATE products SET price = price * 1.1;
```

## 🔒 Security

- **Password Encryption**: All database passwords are encrypted using AES-256-GCM before storage
- **File Permissions**: Config files are stored with `0600` permissions (user read/write only)
- **Key Derivation**: Encryption key derived from username + static salt using SHA-256

⚠️ **Note**: This encryption protects against casual file access. For production use, consider using environment variables or a secrets manager.

## 🧪 Testing

### Run Unit Tests
```bash
go test ./tests/unit/... -v
```

### Run Integration Tests
Requires a running PostgreSQL instance:

```bash
# Start test database
docker run --name test-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 -d postgres

# Run tests
TEST_POSTGRES_DSN="postgres://postgres:postgres@localhost:5432/postgres" \
  go test ./tests/integration/... -v
```

## 🗺️ Roadmap

### v1.0 (Current)
- [x] PostgreSQL support
- [x] Connection management
- [x] Query execution
- [x] Neovim integration
- [x] Environment organization
- [x] Password encryption
- [x] Query history logging
- [x] Help reference dialog

### v1.1 (Planned)
- [ ] MySQL support
- [ ] SQLite support
- [ ] Schema explorer (tables, views, functions)
- [ ] Query library with templates
- [ ] Export results (CSV, JSON, SQL)

### v2.0 (Future)
- [ ] Transaction support
- [ ] Query cancellation
- [ ] Custom themes
- [ ] Configurable keybindings
- [ ] Multi-tab support

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

MIT License - see [LICENSE](LICENSE) file for details

## 🙏 Acknowledgments

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [pgx](https://github.com/jackc/pgx) - PostgreSQL driver

Inspired by:
- [lazygit](https://github.com/jesseduffield/lazygit)
- [lazydocker](https://github.com/jesseduffield/lazydocker)

## 📧 Contact

- GitHub: [@yourusername](https://github.com/yourusername)
- Issues: [GitHub Issues](https://github.com/yourusername/lazydb/issues)

---

**Made with ❤️ for developers who live in the terminal**
