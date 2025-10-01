# 🔧 Neovim Integration Technical Specification

## Overview

LazyDB integrates Neovim as an external query editor, giving users the full power of their Neovim configuration (plugins, LSP, snippets, etc.) for editing SQL queries.

**Design Philosophy**: Keep it simple. Spawn Neovim, let user edit, read result. No complex RPC, no embedded processes (at least for MVP).

---

## Architecture Decision

### Approach A: Spawned Process (MVP - Weeks 1-8)

**How it works**:
1. User presses `Ctrl-E` in query editor
2. LazyDB writes current query to temporary file
3. LazyDB suspends its UI (releases terminal control)
4. Neovim is spawned with the temp file
5. User edits in full-screen Neovim
6. User saves and quits (`:wq`)
7. LazyDB reads edited content from temp file
8. LazyDB resumes UI with updated query
9. Temp file is cleaned up

**Pros**:
- ✅ Simple to implement
- ✅ Works with any Neovim configuration
- ✅ User gets full terminal control
- ✅ No complex IPC/RPC
- ✅ Cross-platform compatible

**Cons**:
- ❌ Full-screen switch (leaves LazyDB UI)
- ❌ Slightly jarring UX (context switch)
- ❌ Can't show LazyDB data while editing

**Implementation Complexity**: Low (1-2 days)

---

### Approach B: Embedded RPC (Future - Phase 4)

**How it works**:
1. Neovim runs as embedded process (`nvim --embed`)
2. Communication via msgpack-rpc
3. Neovim renders in LazyDB's editor panel
4. More integrated experience

**Pros**:
- ✅ No full-screen switch
- ✅ Integrated in LazyDB UI
- ✅ Can show results while editing
- ✅ More polished experience

**Cons**:
- ❌ Complex implementation
- ❌ Requires msgpack-rpc handling
- ❌ May conflict with user's Neovim config
- ❌ Harder to debug

**Implementation Complexity**: High (2-3 weeks)

---

## MVP Implementation (Approach A)

### Go Implementation

```go
package editor

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
)

// NvimEditor handles Neovim integration
type NvimEditor struct {
    executable string        // Path to nvim binary
    tempDir    string        // Directory for temporary files
    config     *NvimConfig   // Editor configuration
}

// NvimConfig holds configuration for Neovim integration
type NvimConfig struct {
    // Neovim arguments to pass
    Args []string

    // Whether to set filetype to SQL
    SetSQLFiletype bool

    // Whether to disable swap file
    NoSwapFile bool

    // Custom init commands
    InitCommands []string
}

// New creates a new NvimEditor instance
func New() (*NvimEditor, error) {
    // Find nvim executable
    executable, err := exec.LookPath("nvim")
    if err != nil {
        // Try common paths
        paths := []string{
            "/usr/bin/nvim",
            "/usr/local/bin/nvim",
            "/opt/homebrew/bin/nvim",
        }

        for _, path := range paths {
            if _, err := os.Stat(path); err == nil {
                executable = path
                break
            }
        }

        if executable == "" {
            return nil, fmt.Errorf("neovim not found in PATH")
        }
    }

    // Get temp directory
    tempDir := filepath.Join(os.TempDir(), "lazydb")
    if err := os.MkdirAll(tempDir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create temp directory: %w", err)
    }

    return &NvimEditor{
        executable: executable,
        tempDir:    tempDir,
        config: &NvimConfig{
            SetSQLFiletype: true,
            NoSwapFile:     true,
        },
    }, nil
}

// Edit opens Neovim for editing SQL content
func (ne *NvimEditor) Edit(initialContent string) (string, error) {
    // 1. Create temporary file
    tmpFile, err := os.CreateTemp(ne.tempDir, "query-*.sql")
    if err != nil {
        return "", fmt.Errorf("failed to create temp file: %w", err)
    }
    tmpFilePath := tmpFile.Name()
    defer os.Remove(tmpFilePath) // Cleanup on exit

    // 2. Write initial content
    if _, err := tmpFile.WriteString(initialContent); err != nil {
        return "", fmt.Errorf("failed to write temp file: %w", err)
    }

    // Important: Close file before opening in Neovim
    if err := tmpFile.Close(); err != nil {
        return "", fmt.Errorf("failed to close temp file: %w", err)
    }

    // 3. Build Neovim command
    args := ne.buildArgs(tmpFilePath)
    cmd := exec.Command(ne.executable, args...)

    // 4. Give Neovim full terminal control
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    // 5. Run Neovim (blocks until user exits)
    if err := cmd.Run(); err != nil {
        return "", fmt.Errorf("neovim exited with error: %w", err)
    }

    // 6. Read edited content
    editedContent, err := os.ReadFile(tmpFilePath)
    if err != nil {
        return "", fmt.Errorf("failed to read edited file: %w", err)
    }

    return string(editedContent), nil
}

// buildArgs constructs Neovim command arguments
func (ne *NvimEditor) buildArgs(filepath string) []string {
    args := []string{}

    // Set SQL filetype
    if ne.config.SetSQLFiletype {
        args = append(args, "-c", "set filetype=sql")
    }

    // Disable swap file
    if ne.config.NoSwapFile {
        args = append(args, "-c", "set noswapfile")
    }

    // Add custom init commands
    for _, cmd := range ne.config.InitCommands {
        args = append(args, "-c", cmd)
    }

    // Add the file to edit
    args = append(args, "--", filepath)

    return args
}

// IsAvailable checks if Neovim is available
func (ne *NvimEditor) IsAvailable() bool {
    _, err := exec.LookPath(ne.executable)
    return err == nil
}

// GetVersion returns Neovim version
func (ne *NvimEditor) GetVersion() (string, error) {
    cmd := exec.Command(ne.executable, "--version")
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }

    // Parse version from output
    // Example: "NVIM v0.9.5"
    lines := string(output)
    if len(lines) > 0 {
        return lines[:50], nil // First line
    }

    return "unknown", nil
}
```

---

## Integration with Bubbletea

### Suspending and Resuming UI

```go
package ui

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/yourusername/lazydb/internal/editor"
)

// nvimEditMsg signals that Neovim editing is complete
type nvimEditMsg struct {
    content string
    err     error
}

// openNvimCmd opens Neovim for editing
func openNvimCmd(editor *editor.NvimEditor, currentQuery string) tea.Cmd {
    return func() tea.Msg {
        // This function runs in a goroutine
        // It will block the entire program while Neovim runs

        content, err := editor.Edit(currentQuery)

        return nvimEditMsg{
            content: content,
            err:     err,
        }
    }
}

// In Model.Update()
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {

    case tea.KeyMsg:
        // Check for Ctrl-E in editor panel
        if m.activePanel == EditorPanel {
            if msg.Type == tea.KeyCtrlE {
                // Suspend Bubbletea program
                m.program.Send(tea.Suspend)

                // Open Neovim (this blocks)
                return m, openNvimCmd(m.nvimEditor, m.editorContent)
            }
        }

    case nvimEditMsg:
        // Neovim editing complete
        if msg.err != nil {
            // Handle error
            m.errorMessage = msg.err.Error()
        } else {
            // Update editor content with edited query
            m.editorContent = msg.content
        }

        return m, nil
    }

    return m, nil
}
```

**Important**: The `tea.Suspend` message isn't a real Bubbletea feature. We need a different approach.

### Better Approach: Alt Screen

```go
// openNvim opens Neovim and suspends Bubbletea
func (m *Model) openNvim() tea.Cmd {
    return tea.Batch(
        // Exit alt screen (returns to main terminal)
        tea.ExitAltScreen,

        // Run Neovim editing
        func() tea.Msg {
            content, err := m.nvimEditor.Edit(m.editorContent)
            return nvimEditMsg{content: content, err: err}
        },

        // Re-enter alt screen after Neovim exits
        tea.EnterAltScreen,
    )
}
```

**Problem**: Bubbletea keeps running while Neovim is open. This can cause conflicts.

### Correct Approach: Program Quit/Restart

```go
// When user presses Ctrl-E:
case tea.KeyCtrlE:
    // 1. Save current state
    if err := m.saveState(); err != nil {
        return m, showError(err)
    }

    // 2. Quit Bubbletea
    return m, tea.Sequence(
        func() tea.Msg {
            // Run Neovim (blocks)
            content, err := m.nvimEditor.Edit(m.editorContent)
            if err != nil {
                return errorMsg{err}
            }

            // Save edited content to state
            m.editorContent = content
            m.saveState()

            return nvimEditCompleteMsg{}
        },
        tea.Quit,
    )

// In main():
func main() {
    app := NewApp()

    for {
        // Run Bubbletea
        p := tea.NewProgram(app, tea.WithAltScreen())
        if _, err := p.Run(); err != nil {
            log.Fatal(err)
        }

        // Check if we need to restart after Neovim
        if app.needsRestart {
            app.needsRestart = false
            continue
        }

        break
    }
}
```

**This is too complex**. Let's use the simplest approach:

### Simplest Approach: Exec Directly

```go
case tea.KeyCtrlE:
    // Just call editor directly (blocks everything)
    content, err := m.nvimEditor.Edit(m.editorContent)
    if err != nil {
        m.errorMessage = err.Error()
        return m, nil
    }

    m.editorContent = content
    return m, nil
```

**The key insight**: It's OK if LazyDB blocks while Neovim runs! The user expects this behavior. They pressed Ctrl-E to edit in Neovim, so they want the full terminal for Neovim.

---

## User Experience Flow

### Typical Workflow

```
1. User writes partial query in LazyDB editor:
   SELECT * FROM users WHERE

2. User presses Ctrl-E

3. LazyDB suspends (screen clears)

4. Neovim opens full-screen with query:
   ┌─────────────────────────────────┐
   │ query-20240115.sql              │
   ├─────────────────────────────────┤
   │ SELECT * FROM users WHERE█      │
   │                                 │
   │                                 │
   │ ~                               │
   │ ~                               │
   └─────────────────────────────────┘

5. User completes query in Neovim:
   SELECT * FROM users WHERE active = true
   ORDER BY created_at DESC
   LIMIT 10;

6. User saves and quits: :wq

7. LazyDB resumes with updated query

8. User presses Ctrl-R to execute
```

### Helpful Features in Neovim

**Auto-set SQL filetype**:
```bash
nvim -c "set filetype=sql" query.sql
```

**Enable SQL syntax highlighting**:
- Neovim does this automatically with filetype=sql

**Load user's config**:
- Neovim loads `~/.config/nvim/init.lua` by default
- User's plugins, LSP, snippets all work automatically

**Optional: Custom Neovim config for LazyDB**:
```bash
# User can create ~/.config/lazydb/nvim-init.lua
nvim -u ~/.config/lazydb/nvim-init.lua query.sql
```

---

## Configuration

### LazyDB Config

```toml
[editor]
# Editor to use (nvim, vim, nano, etc.)
command = "nvim"

# Arguments to pass to editor
args = [
    "-c", "set filetype=sql",
    "-c", "set noswapfile"
]

# Custom Neovim config (optional)
nvim_config = "~/.config/lazydb/nvim-init.lua"

# Whether to show hints before opening
show_hints = true
```

### User's Neovim Config

Users can add LazyDB-specific config in their `init.lua`:

```lua
-- ~/.config/nvim/init.lua

-- Detect LazyDB temp files
vim.api.nvim_create_autocmd("BufRead", {
  pattern = "*/lazydb/query-*.sql",
  callback = function()
    -- LazyDB query file detected

    -- Set some helpful options
    vim.bo.textwidth = 0  -- No line wrapping
    vim.wo.wrap = false

    -- Load SQL snippets
    require("luasnip").filetype_extend("sql", {"postgres"})

    -- Start LSP if available
    vim.cmd("LspStart")

    -- Helpful keymaps
    vim.keymap.set("n", "<leader>x", ":wq<CR>", {
      buffer = true,
      desc = "Save and return to LazyDB"
    })
  end
})
```

---

## Error Handling

### Neovim Not Found

```go
if !editor.IsAvailable() {
    // Show error message
    return showError("Neovim not found. Please install Neovim or configure a different editor.")
}
```

### Fallback to Built-in Editor

```go
func (m *Model) editQuery() tea.Cmd {
    // Try Neovim first
    if m.nvimEditor.IsAvailable() {
        return m.openNvim()
    }

    // Fall back to built-in editor
    return m.useBuiltinEditor()
}
```

### Neovim Exits with Error

```go
if err := cmd.Run(); err != nil {
    // Check exit code
    if exitErr, ok := err.(*exec.ExitError); ok {
        code := exitErr.ExitCode()

        if code == 1 {
            // User likely pressed :cq (quit without saving)
            return "", fmt.Errorf("editing cancelled")
        }

        return "", fmt.Errorf("neovim exited with code %d", code)
    }

    return "", err
}
```

### Temp File Issues

```go
// Ensure temp directory exists
if err := os.MkdirAll(tempDir, 0755); err != nil {
    return fmt.Errorf("failed to create temp directory: %w", err)
}

// Check disk space before writing
stat, err := os.Stat(tempDir)
if err != nil {
    return fmt.Errorf("failed to check temp directory: %w", err)
}

// Clean up old temp files periodically
m.cleanupOldTempFiles()
```

---

## Alternative Editors

### Support Other Editors

```go
type EditorType int

const (
    EditorNeovim EditorType = iota
    EditorVim
    EditorNano
    EditorVSCode
    EditorCustom
)

func (et EditorType) Command() string {
    switch et {
    case EditorNeovim:
        return "nvim"
    case EditorVim:
        return "vim"
    case EditorNano:
        return "nano"
    case EditorVSCode:
        return "code"
    default:
        return "nvim"
    }
}

func (et EditorType) Args(filepath string) []string {
    switch et {
    case EditorNeovim, EditorVim:
        return []string{
            "-c", "set filetype=sql",
            filepath,
        }
    case EditorNano:
        return []string{"-Y", "sql", filepath}
    case EditorVSCode:
        return []string{"--wait", filepath}
    default:
        return []string{filepath}
    }
}
```

### Auto-detect Editor

```go
func detectEditor() EditorType {
    // Check EDITOR environment variable
    if editor := os.Getenv("EDITOR"); editor != "" {
        if strings.Contains(editor, "nvim") {
            return EditorNeovim
        }
        if strings.Contains(editor, "vim") {
            return EditorVim
        }
        if strings.Contains(editor, "nano") {
            return EditorNano
        }
    }

    // Try in order of preference
    editors := []EditorType{
        EditorNeovim,
        EditorVim,
        EditorNano,
    }

    for _, et := range editors {
        if _, err := exec.LookPath(et.Command()); err == nil {
            return et
        }
    }

    // No editor found
    return EditorCustom
}
```

---

## Testing

### Unit Tests

```go
func TestNvimEditor_Edit(t *testing.T) {
    // Skip if Neovim not available
    if !isNvimAvailable() {
        t.Skip("Neovim not available")
    }

    editor, err := New()
    require.NoError(t, err)

    initial := "SELECT 1;"

    // Mock user editing
    // In real test, we'd programmatically edit the file

    result, err := editor.Edit(initial)
    require.NoError(t, err)
    assert.Contains(t, result, "SELECT")
}
```

### Integration Tests

Test with actual Neovim:

```go
func TestNvimIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Create test file
    content := "SELECT * FROM test;"

    // Use expect/pty library to automate Neovim
    // 1. Open Neovim
    // 2. Send keystrokes
    // 3. Save and quit
    // 4. Verify output
}
```

---

## Future Enhancements (Phase 4)

### Embedded Neovim with RPC

See ARCHITECTURE.md for details on embedded RPC approach.

### Split View

Show LazyDB results while editing query:

```
┌──────────────────────────────────────┐
│ NEOVIM                               │
│ SELECT * FROM users█                 │
│                                      │
├──────────────────────────────────────┤
│ LIVE RESULTS                         │
│ ╔══╦══════╦══════╗                   │
│ ║id║ name ║email ║                   │
│ ╠══╬══════╬══════╣                   │
│ ║1 ║Alice ║a@... ║                   │
│ ╚══╩══════╩══════╝                   │
└──────────────────────────────────────┘
```

This requires embedded Neovim with RPC.

---

## Summary

**For MVP (Phase 1)**:
- Use simple spawned process approach
- Let user edit in full-screen Neovim
- Read result from temp file
- Simple, reliable, works everywhere

**For Future (Phase 4)**:
- Consider embedded RPC approach
- More integrated experience
- Split view with live results
- More complex but better UX

**Decision**: Start with Approach A, upgrade to Approach B later if needed.
