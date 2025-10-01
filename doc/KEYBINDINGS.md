# ⌨️ Keybindings Reference - LazyDB

Complete keyboard shortcuts reference for LazyDB, organized by context and panel.

---

## Philosophy

LazyDB follows the **100% keyboard-driven workflow** philosophy inspired by lazygit:

1. **Vim-Style Navigation**: `hjkl` for movement, intuitive mnemonics for actions
2. **Context-Aware**: Keybindings change based on active panel and state
3. **Discoverable**: Status bar shows relevant shortcuts, `?` for help
4. **Customizable**: All keybindings can be remapped in config
5. **Consistent**: Similar actions use similar keys across contexts

---

## Quick Reference

### Most Important Keys

| Key | Action | Context |
|-----|--------|---------|
| `?` | Toggle help | Global |
| `q` | Quit | Global |
| `1` / `2` / `3` | Jump to panel 1/2/3 | Global |
| `Tab` / `Shift-Tab` | Cycle panels | Global |
| `Ctrl-R` | Execute query | Editor |
| `Ctrl-E` | Edit in Neovim | Editor |
| `Enter` | Connect to database | Connections |
| `a` | Add new connection | Connections |
| `e` | Export results | Results |

---

## Global Keybindings

**Available in all contexts**

### Panel Navigation

| Key | Action | Description |
|-----|--------|-------------|
| `1` | Focus connections panel | Jump to left panel |
| `2` | Focus editor panel | Jump to center panel |
| `3` | Focus results panel | Jump to right panel |
| `Tab` | Next panel | Cycle through panels forward |
| `Shift-Tab` | Previous panel | Cycle through panels backward |

### Application Control

| Key | Action | Description |
|-----|--------|-------------|
| `q` | Quit application | Exit LazyDB (with confirmation if unsaved) |
| `Ctrl-C` | Force quit | Emergency exit without saving |
| `?` | Toggle help | Show/hide keybindings help overlay |
| `Ctrl-L` | Redraw screen | Refresh UI (fixes rendering issues) |

### History & Library

| Key | Action | Description |
|-----|--------|-------------|
| `Ctrl-H` | Query history | Open query history modal |
| `Ctrl-L` | Query library | Open saved queries modal |
| `:` | Command mode | Open command palette (future) |

---

## Connections Panel Keybindings

**Active when connections panel is focused (Panel 1)**

### Navigation

| Key | Action | Description |
|-----|--------|-------------|
| `j` / `↓` | Move down | Select next connection/group |
| `k` / `↑` | Move up | Select previous connection/group |
| `g` / `Home` | Jump to top | Go to first item |
| `G` / `End` | Jump to bottom | Go to last item |
| `h` / `←` | Collapse group | Close expanded group |
| `l` / `→` | Expand group | Open collapsed group |
| `Space` | Toggle expand | Expand/collapse current group |

### Connection Actions

| Key | Action | Description |
|-----|--------|-------------|
| `Enter` | Connect | Connect to selected database |
| `d` | Disconnect | Disconnect from active database |
| `a` | Add connection | Open new connection form |
| `e` | Edit connection | Edit selected connection |
| `x` | Delete connection | Delete selected connection (with confirmation) |
| `t` | Test connection | Test selected connection without connecting |
| `r` | Refresh | Reload connection list |

### Schema Navigation

| Key | Action | Description |
|-----|--------|-------------|
| `s` | Toggle schema | Show/hide schema tree for connected database |
| `Enter` | Preview table | View table data (SELECT * LIMIT 10) |
| `i` | Table info | Show table structure (DESCRIBE) |
| `c` | Copy name | Copy table/column name to clipboard |

### Filtering & Search

| Key | Action | Description |
|-----|--------|-------------|
| `/` | Search connections | Filter connections by name |
| `Esc` | Clear search | Clear search filter |
| `n` | Next match | Jump to next search result |
| `N` | Previous match | Jump to previous search result |

### Grouping & Organization

| Key | Action | Description |
|-----|--------|-------------|
| `f` | Filter by env | Show only dev/staging/prod connections |
| `F` | Clear filter | Show all connections |
| `m` | Move to group | Move connection to different environment group |

---

## Editor Panel Keybindings

**Active when editor panel is focused (Panel 2)**

### Basic Editing

| Key | Action | Description |
|-----|--------|-------------|
| `Ctrl-E` | Edit in Neovim | Open current query in Neovim |
| `Esc` | Exit editor | Return to normal mode (if in insert mode) |
| `i` | Insert mode | Start typing query (if not already) |

**Note**: Built-in editor uses basic text input. For full Vim keybindings, use `Ctrl-E` to edit in Neovim.

### Query Execution

| Key | Action | Description |
|-----|--------|-------------|
| `Ctrl-R` | Execute query | Run current query |
| `Ctrl-Enter` | Execute query | Alternative execute shortcut |
| `Ctrl-C` | Cancel query | Stop running query (if supported) |
| `Ctrl-T` | Execute selection | Run highlighted portion (future) |

### Query Management

| Key | Action | Description |
|-----|--------|-------------|
| `Ctrl-S` | Save query | Save to query library |
| `Ctrl-O` | Open query | Load query from library |
| `Ctrl-N` | New query | Clear editor, start fresh |
| `Ctrl-D` | Duplicate query | Copy current query to new tab (future) |

### Query History

| Key | Action | Description |
|-----|--------|-------------|
| `Ctrl-H` | Query history | Open history modal |
| `Ctrl-↑` | Previous query | Load previous query from history |
| `Ctrl-↓` | Next query | Load next query from history |

### Templates & Snippets

| Key | Action | Description |
|-----|--------|-------------|
| `Ctrl-T` | Templates menu | Show common query templates |
| `Ctrl-Space` | Autocomplete | Trigger table/column completion (future) |

**Available Templates**:
- `select` → `SELECT * FROM table_name LIMIT 10;`
- `insert` → `INSERT INTO table_name (col1) VALUES (val1);`
- `update` → `UPDATE table_name SET col1 = val1 WHERE id = 1;`
- `delete` → `DELETE FROM table_name WHERE id = 1;`
- `join` → `SELECT * FROM t1 JOIN t2 ON t1.id = t2.id;`
- `create` → `CREATE TABLE table_name (...);`

### Formatting & Validation

| Key | Action | Description |
|-----|--------|-------------|
| `Ctrl-F` | Format query | Auto-format SQL (future) |
| `Ctrl-V` | Validate query | Check SQL syntax without executing |
| `Ctrl-/` | Toggle comment | Comment/uncomment current line |

---

## Results Panel Keybindings

**Active when results panel is focused (Panel 3)**

### Navigation

| Key | Action | Description |
|-----|--------|-------------|
| `j` / `↓` | Move down | Scroll down one row |
| `k` / `↑` | Move up | Scroll up one row |
| `h` / `←` | Scroll left | Move left one column |
| `l` / `→` | Scroll right | Move right one column |
| `g` / `Home` | Jump to top | Go to first row |
| `G` / `End` | Jump to bottom | Go to last row |
| `Ctrl-F` / `Page Down` | Page down | Scroll down one page |
| `Ctrl-B` / `Page Up` | Page up | Scroll up one page |
| `0` / `^` | Jump to first column | Go to leftmost column |
| `$` | Jump to last column | Go to rightmost column |

### Data Actions

| Key | Action | Description |
|-----|--------|-------------|
| `y` | Copy cell | Copy selected cell to clipboard |
| `Y` | Copy row | Copy entire row (tab-separated) |
| `Ctrl-Y` | Copy column | Copy entire column |
| `Ctrl-A` | Copy all | Copy entire result set |
| `v` | Visual mode | Start selecting cells (future) |

### Export & Sharing

| Key | Action | Description |
|-----|--------|-------------|
| `e` | Export menu | Open export options |
| `Ctrl-E` | Export to CSV | Quick export to CSV file |
| `Ctrl-J` | Export to JSON | Quick export to JSON file |
| `Ctrl-S` | Export to SQL | Export as INSERT statements |

### Filtering & Sorting

| Key | Action | Description |
|-----|--------|-------------|
| `/` | Search results | Filter results by text |
| `Esc` | Clear search | Clear search filter |
| `s` | Sort column | Sort by current column (toggle asc/desc) |
| `S` | Clear sort | Remove all sorting |
| `f` | Filter column | Add filter to current column (future) |
| `F` | Clear filters | Remove all filters (future) |

### View Options

| Key | Action | Description |
|-----|--------|-------------|
| `w` | Toggle wrap | Wrap/truncate long text in cells |
| `n` | Toggle numbers | Show/hide row numbers |
| `#` | Toggle grid | Show/hide table borders |
| `Ctrl-+` | Increase width | Widen current column |
| `Ctrl--` | Decrease width | Narrow current column |
| `=` | Auto-fit columns | Adjust column widths to content |

### Results Management

| Key | Action | Description |
|-----|--------|-------------|
| `r` | Refresh | Re-execute last query |
| `c` | Clear results | Clear current results |
| `p` | Previous result | Switch to previous result tab (future) |
| `n` | Next result | Switch to next result tab (future) |

---

## Modal Keybindings

**Active when modal dialogs are open**

### Connection Form

| Key | Action | Description |
|-----|--------|-------------|
| `Tab` | Next field | Move to next input field |
| `Shift-Tab` | Previous field | Move to previous input field |
| `Enter` | Submit | Save connection |
| `Esc` | Cancel | Close form without saving |
| `Ctrl-T` | Test connection | Test before saving |

**Form Fields Navigation Order**:
1. Name
2. Type (PostgreSQL/MySQL/SQLite)
3. Host
4. Port
5. Database
6. Username
7. Password
8. Environment (dev/staging/prod)

### Query History Modal

| Key | Action | Description |
|-----|--------|-------------|
| `j` / `↓` | Next query | Select next query |
| `k` / `↑` | Previous query | Select previous query |
| `Enter` | Load query | Load selected query to editor |
| `/` | Search history | Filter queries |
| `d` | Delete query | Remove from history |
| `Esc` | Close | Close modal |

### Query Library Modal

| Key | Action | Description |
|-----|--------|-------------|
| `j` / `↓` | Next query | Select next query |
| `k` / `↑` | Previous query | Select previous query |
| `Enter` | Load query | Load selected query to editor |
| `/` | Search library | Filter saved queries |
| `e` | Edit metadata | Edit query name/tags |
| `d` | Delete query | Remove from library |
| `Esc` | Close | Close modal |

### Export Dialog

| Key | Action | Description |
|-----|--------|-------------|
| `j` / `↓` | Next option | Select next export format |
| `k` / `↑` | Previous option | Select previous export format |
| `Enter` | Export | Export to selected format |
| `Tab` | Toggle options | Cycle through export options |
| `Esc` | Cancel | Close dialog |

**Export Options**:
- CSV (with/without headers)
- JSON (array/objects)
- SQL (INSERT statements)
- Markdown table
- TSV (tab-separated)
- Clipboard

### Help Modal

| Key | Action | Description |
|-----|--------|-------------|
| `j` / `↓` | Scroll down | View more keybindings |
| `k` / `↑` | Scroll up | Scroll back |
| `g` / `Home` | Jump to top | Go to beginning |
| `G` / `End` | Jump to bottom | Go to end |
| `/` | Search help | Filter keybindings |
| `?` / `Esc` | Close | Close help modal |

### Confirmation Dialogs

| Key | Action | Description |
|-----|--------|-------------|
| `y` / `Enter` | Confirm | Proceed with action |
| `n` / `Esc` | Cancel | Cancel action |
| `Tab` | Toggle | Switch between Yes/No |

---

## Transaction Mode Keybindings

**Active when transaction is in progress**

| Key | Action | Description |
|-----|--------|-------------|
| `Ctrl-Shift-B` | BEGIN | Start transaction |
| `Ctrl-Shift-C` | COMMIT | Commit transaction |
| `Ctrl-Shift-R` | ROLLBACK | Rollback transaction |

**Transaction Status**: Shown in status bar when active.

---

## Advanced Features

### Multi-Query Execution

| Key | Action | Description |
|-----|--------|-------------|
| `Ctrl-Shift-R` | Execute all | Run all queries in editor |
| `Ctrl-Alt-R` | Execute to cursor | Run queries up to cursor position |

### Tabs & Windows (Future)

| Key | Action | Description |
|-----|--------|-------------|
| `Ctrl-T` | New tab | Open new query tab |
| `Ctrl-W` | Close tab | Close current tab |
| `Ctrl-Tab` | Next tab | Switch to next tab |
| `Ctrl-Shift-Tab` | Previous tab | Switch to previous tab |
| `Alt-1` to `Alt-9` | Jump to tab N | Jump to specific tab |

### Compare Mode (Future)

| Key | Action | Description |
|-----|--------|-------------|
| `Ctrl-D` | Diff mode | Compare two result sets |
| `]c` | Next diff | Jump to next difference |
| `[c` | Previous diff | Jump to previous difference |

---

## Customization

### Config File Location

Keybindings are customized in:

```bash
~/.config/lazydb/config.toml
```

### Keybinding Configuration

```toml
[keybindings]

# Global
quit = "q"
help = "?"
panel_1 = "1"
panel_2 = "2"
panel_3 = "3"

# Connections Panel
connect = "<enter>"
add_connection = "a"
edit_connection = "e"
delete_connection = "x"
test_connection = "t"

# Editor Panel
execute_query = "<C-r>"
edit_in_nvim = "<C-e>"
save_query = "<C-s>"
open_query = "<C-o>"
new_query = "<C-n>"

# Results Panel
export_menu = "e"
copy_cell = "y"
copy_row = "Y"
search_results = "/"

# Navigation
move_down = "j"
move_up = "k"
move_left = "h"
move_right = "l"
jump_top = "g"
jump_bottom = "G"
```

### Key Syntax

| Syntax | Description | Example |
|--------|-------------|---------|
| `a` | Single key | Press `a` |
| `<C-r>` | Ctrl + key | Ctrl+R |
| `<S-tab>` | Shift + key | Shift+Tab |
| `<M-x>` / `<A-x>` | Alt + key | Alt+X |
| `<C-S-r>` | Ctrl + Shift + key | Ctrl+Shift+R |
| `<enter>` | Enter key | Enter |
| `<esc>` | Escape key | Esc |
| `<space>` | Space bar | Space |
| `<tab>` | Tab key | Tab |
| `<up>` / `<down>` | Arrow keys | ↑ / ↓ |
| `<pageup>` / `<pagedown>` | Page keys | PgUp / PgDn |
| `<home>` / `<end>` | Home/End keys | Home / End |

### Vim-Style Key Notation

LazyDB uses Vim-style key notation for consistency with Neovim:

- `<C-x>` = Ctrl+X
- `<M-x>` or `<A-x>` = Alt+X
- `<S-x>` = Shift+X
- `<CR>` = Enter/Return
- `<Esc>` = Escape
- `<BS>` = Backspace
- `<Del>` = Delete
- `<Space>` = Space

### Creating Custom Keybindings

**Example**: Add custom export shortcuts

```toml
[keybindings.custom]
export_csv = "<C-e>"
export_json = "<C-j>"
export_sql = "<C-s>"
quick_select = "<leader>s"
quick_describe = "<leader>d"
```

### Keybinding Presets

LazyDB includes preset configurations:

```toml
[keybindings]
preset = "vim"  # Options: "vim" (default), "emacs", "custom"
```

**Vim Preset** (default):
- Navigation: `hjkl`
- Jump to top/bottom: `g`/`G`
- Visual mode: `v`
- Yank/copy: `y`

**Emacs Preset**:
- Navigation: Arrow keys / `Ctrl-N/P/F/B`
- Jump to top/bottom: `Alt-<` / `Alt->`
- Kill/copy: `Ctrl-W` / `Alt-W`
- Cancel: `Ctrl-G`

---

## Keyboard Shortcuts Cheat Sheet

### Quick Start

```
┌─────────────────────────────────────────────────────────────┐
│                   LazyDB Quick Reference                     │
├─────────────────────────────────────────────────────────────┤
│ NAVIGATION                                                   │
│  1,2,3    Focus panel     Tab      Next panel               │
│  hjkl     Move cursor     g/G      Top/Bottom               │
│                                                              │
│ CONNECTIONS                                                  │
│  Enter    Connect         a        Add connection           │
│  e        Edit            x        Delete                    │
│  s        Toggle schema   t        Test connection          │
│                                                              │
│ EDITOR                                                       │
│  Ctrl-R   Execute query   Ctrl-E   Edit in Neovim           │
│  Ctrl-S   Save query      Ctrl-O   Open query               │
│  Ctrl-H   Query history   Ctrl-N   New query                │
│                                                              │
│ RESULTS                                                      │
│  y        Copy cell       Y        Copy row                 │
│  e        Export menu     /        Search                   │
│  s        Sort column     w        Toggle wrap              │
│                                                              │
│ GLOBAL                                                       │
│  ?        Help            q        Quit                      │
│  Ctrl-L   Redraw          Esc      Cancel/Close             │
└─────────────────────────────────────────────────────────────┘
```

---

## Platform-Specific Notes

### macOS

- `Cmd` key can be mapped to `<D-x>` in config (experimental)
- Terminal.app may not support all Ctrl combinations
- Use iTerm2 or Alacritty for best compatibility

### Linux

- All keybindings work in most terminal emulators
- Some terminals may intercept `Ctrl-Shift` combinations
- Configure terminal to pass through all keys to LazyDB

### Windows

- Use Windows Terminal or Alacritty for best experience
- Some Ctrl combinations may conflict with Windows shortcuts
- PowerShell may require special configuration

### Terminal Compatibility

**Recommended Terminals**:
- ✅ Alacritty (best compatibility)
- ✅ iTerm2 (macOS)
- ✅ Windows Terminal
- ✅ Kitty
- ⚠️ Terminal.app (limited)
- ⚠️ GNOME Terminal (most keys work)
- ❌ PuTTY (poor Ctrl key support)

---

## Conflicts & Resolution

### Common Keybinding Conflicts

| Key | LazyDB | System | Resolution |
|-----|--------|--------|------------|
| `Ctrl-C` | Cancel query | SIGINT | Use `Ctrl-K` for cancel in config |
| `Ctrl-Z` | (unused) | Suspend | LazyDB handles SIGTSTP properly |
| `Ctrl-S` | Save query | Terminal flow control | Disable XON/XOFF in terminal settings |
| `Ctrl-Q` | (unused) | Terminal flow control | Disable XON/XOFF in terminal settings |

### Disable Terminal Flow Control

Add to `~/.bashrc` or `~/.zshrc`:

```bash
# Disable XON/XOFF flow control
stty -ixon
```

---

## Accessibility

### Screen Reader Support

LazyDB provides aria-label equivalents in status messages:

- Current panel and selection announced
- Action results spoken
- Error messages read aloud

### High Contrast Mode

Enable in config:

```toml
[ui]
high_contrast = true
```

Changes:
- Brighter colors
- Bolder borders
- Higher contrast text

### Reduced Motion

Disable animations:

```toml
[ui]
animations = false
```

---

## Learning Path

### Beginner (Day 1)

Learn these essential keys first:
- `1`, `2`, `3` - Panel navigation
- `Enter` - Connect to database
- `Ctrl-R` - Execute query
- `q` - Quit
- `?` - Help

### Intermediate (Week 1)

Add these shortcuts:
- `hjkl` - Vim-style navigation
- `a`, `e`, `x` - Manage connections
- `y`, `Y` - Copy results
- `Ctrl-E` - Edit in Neovim
- `/` - Search

### Advanced (Month 1)

Master these features:
- `Ctrl-S`, `Ctrl-O` - Save/load queries
- `Ctrl-H` - Query history
- `s`, `f` - Sort and filter results
- `e` - Export options
- Transaction mode keys

---

## Troubleshooting

### Keybinding Not Working

**Symptom**: Key press has no effect

**Solutions**:
1. Check if terminal supports the key combination
2. Verify key not intercepted by terminal emulator
3. Check for conflicts in terminal settings
4. Try alternative key binding
5. Enable debug mode to see key events:
   ```bash
   lazydb --debug-keys
   ```

### Key Pressed but Wrong Action

**Symptom**: Different action than expected

**Solutions**:
1. Verify active panel (check status bar)
2. Check if modal is open (keys work in modal context)
3. Review custom keybindings in config
4. Reset to default keybindings:
   ```bash
   lazydb --reset-keys
   ```

### Ctrl/Alt Keys Not Working

**Symptom**: Modified keys (Ctrl/Alt) don't work

**Solutions**:
1. **Terminal Settings**: Configure terminal to pass through all keys
2. **macOS Terminal.app**: Switch to iTerm2 or Alacritty
3. **tmux/screen**: May intercept some keys - check tmux.conf
4. **SSH**: Ensure SSH client passes through all keys

### Special Characters

**Symptom**: Non-ASCII characters cause issues

**Solutions**:
1. Ensure terminal uses UTF-8 encoding
2. Set `LANG=en_US.UTF-8` in environment
3. Check terminal font supports Unicode

---

## Future Enhancements

### Planned Keybindings

**v0.6** (Schema Explorer):
- `Ctrl-P` - Table picker (fuzzy find)
- `Ctrl-Shift-F` - Find in all tables
- `Alt-Enter` - Quick preview table

**v0.7** (Advanced Export):
- `Ctrl-Shift-E` - Export wizard
- `Ctrl-Shift-P` - Print query results
- `Ctrl-Shift-S` - Schedule query (future)

**v0.8** (Collaboration):
- `Ctrl-Shift-Share` - Share query snippet
- `Ctrl-Shift-C` - Copy shareable link

**v1.0** (Visual Query Builder):
- `Ctrl-B` - Toggle query builder
- `Ctrl-Shift-J` - Join builder
- `Ctrl-Shift-W` - WHERE clause builder

---

## See Also

- [UI Mockups](./UI_MOCKUPS.md) - Visual guide to UI layout
- [Neovim Integration](./NEOVIM_INTEGRATION.md) - Neovim editor details
- [Implementation Plan](./IMPLEMENTATION_PLAN.md) - Development timeline
- [Configuration Guide](./CONFIGURATION.md) - Customization options (future)

---

## Feedback & Contributions

Have suggestions for better keybindings? Open an issue:

```bash
# GitHub Issues
https://github.com/yourusername/lazydb/issues

# Keybinding requests should include:
# - Use case description
# - Proposed key combination
# - Context (which panel)
# - Conflicts with existing keys
```

---

**Document Version**: v1.0
**Last Updated**: 2024-01
**Applies to LazyDB**: v0.1.0-dev (MVP Phase)
