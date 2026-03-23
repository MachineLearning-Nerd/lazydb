# Centralized Key Sink for INSERT Mode

**Date:** 2026-03-23
**Status:** Draft
**Scope:** `cmd/lazydb/main.go` key dispatch refactor

## Problem

The editor's INSERT mode allows users to type SQL queries. However, several keybindings in `main.go` intercept keys **before** they reach the editor, making certain characters impossible to type:

| Key | Expected (INSERT) | Actual | Root Cause |
|-----|-------------------|--------|------------|
| `=` | Type `=` | Resizes panel | No INSERT guard (line 329) |
| `-` | Type `-` | Resizes panel | No INSERT guard (line 330) |
| `[` | Type `[` | Resizes panel | No INSERT guard (line 345) |
| `]` | Type `]` | Resizes panel | No INSERT guard (line 340) |
| `?` | Type `?` | Opens help dialog | No INSERT guard (line 404) |

Additionally, the existing INSERT mode guards are **scattered** across two locations (line 356 for `L`, line 455 for navigation). Every new keybinding added requires remembering to add a guard — a fragile pattern that has already produced 5 bugs.

## Solution

Insert a single **key sink** block at the top of the key dispatch (after AI/dialog checks, before all keybinding blocks). When the editor is focused and in INSERT mode, this block:

1. Checks a **whitelist** of always-global keys
2. Routes whitelisted keys to their handlers
3. Routes **everything else** directly to the editor and returns

This inverts the guard pattern: instead of N blocks each needing a guard, one block at the top handles the entire INSERT mode.

## Key Dispatch Flow (After)

```
tea.KeyMsg received
  |
  v
[1] AI Assistant visible? --> route to AI, return
  |
  v
[2] Dialog open? --> handleDialogKeys(), return
  |
  v
[3] KEY SINK: Editor focused + INSERT mode + NOT in layoutPresetMode?
  |   YES --> check whitelist:
  |     Global.Quit/ctrl+c  --> quit
  |     ctrl+r               --> execute query
  |     f2                   --> save query
  |     ctrl+e               --> open neovim
  |     ctrl+a               --> AI assistant
  |     (all others)         --> editorPanel.Update(msg), return
  |
  |   NO --> fall through
  v
[4] Resize keys (=, -, [, ])
[5] Layout preset mode (L, 1-4)
[6] Global keybindings (ctrl+q, ?, ctrl+r, f2, ctrl+e, ctrl+a)
[7] Navigation (tab, shift+tab, 1, 2, 3)
[8] Connections panel (a, e, d, enter)
[9] Panel.Update() for focused panel
```

### layoutPresetMode Exemption

The key sink condition includes `!m.layoutPresetMode`. Although entering preset mode from INSERT mode is unlikely (the `L` key guard prevents it), it is theoretically possible via programmatic focus changes. When `layoutPresetMode` is active, keys 1-4 and ESC must reach the layout preset handler rather than being swallowed by the key sink. Once the preset is selected or cancelled, the key sink resumes normal operation.

## Whitelist

Keys that remain functional during INSERT mode:

| Key | Action | Rationale |
|-----|--------|-----------|
| `Global.Quit` (default: `ctrl+q`) | Quit app | Must always be able to exit; uses config value, not hardcoded |
| `ctrl+c` | Quit app | Standard interrupt (always hardcoded) |
| `ctrl+r` | Execute query | Primary action, must work while typing |
| `f2` | Save query | Function key, no typing conflict |
| `ctrl+e` | Open Neovim | Modifier key, intentional action |
| `ctrl+a` | AI Assistant | Modifier key, intentional action |

Keys that are **not** whitelisted (routed to editor for typing):

| Key | Was intercepted by | Now types in editor |
|-----|--------------------|-------------------|
| `=` | Resize (grow editor left) | `=` character |
| `-` | Resize (shrink editor left) | `-` character |
| `[` | Resize (grow editor right) | `[` character |
| `]` | Resize (shrink editor right) | `]` character |
| `?` | Help dialog | `?` character |
| `L` | Layout preset mode | `L` character |
| `1`, `2`, `3` | Panel navigation | Number characters |
| `tab` | Next panel | Tab in editor (behavior-preserving: currently the INSERT guard at line 455 already prevents tab navigation, and tab reaches the editor via the catch-all at line 586) |
| `shift+tab` | Previous panel | Shift-tab in editor (same as tab — behavior-preserving) |

**Note on `esc`:** Handled internally by the editor panel (switches INSERT to NORMAL mode), so it does not need special handling in the whitelist — it falls through to `editorPanel.Update(msg)`.

**Note on `Global.Help` (`?`):** Intentionally not whitelisted because `?` is a typeable character needed in SQL queries (e.g., parameterized queries). Users who remap Help to a modifier key (e.g., `ctrl+h`) should be aware it will not function during INSERT mode — they must ESC to NORMAL mode first.

## Implementation

### Step 1: Add key sink block

Insert after line 306 (dialog check), before line 308 (resize keys):

```go
// KEY SINK: Editor in INSERT mode captures all keys except whitelisted globals.
// This prevents resize, layout, navigation, and other keybindings from
// intercepting characters the user is trying to type in the query editor.
if m.focusedPanel == FocusEditor && m.editorPanel.IsInInsertMode() && !m.layoutPresetMode {
    switch key {
    case m.config.Keybindings.Global.Quit, "ctrl+c":
        m.saveConnections()
        return m, tea.Quit
    case m.config.Keybindings.Global.ExecuteQuery: // ctrl+r
        return m, m.executeQuery()
    case m.config.Keybindings.Global.SaveQuery: // f2
        return m, m.saveQuery()
    case m.config.Keybindings.Global.OpenNeovim: // ctrl+e
        if editor.IsNvimAvailable() {
            query := m.editorPanel.GetQuery()
            activeConn, err := m.connManager.GetActive()
            var conn db.Connection
            if err == nil {
                conn = activeConn
            }
            injectSchema := true
            if m.config.AI != nil {
                injectSchema = m.config.AI.InjectInNeovim
            }
            return m, editor.OpenInNeovimCmd(query, conn, injectSchema)
        }
        m.statusMessage = "Neovim not found. Please install nvim."
        return m, nil
    case m.config.Keybindings.Global.AIAssistant: // ctrl+a
        if m.config.AI != nil && m.config.AI.Enabled {
            query := m.editorPanel.GetQuery()
            m.aiAssistant.Show(query)
        } else {
            m.statusMessage = "AI Assistant is disabled. Enable in config.yml"
        }
        return m, nil
    default:
        // All other keys go to editor textarea for typing
        cmd = m.editorPanel.Update(msg)
        return m, cmd
    }
}
```

### Step 2: Remove redundant INSERT mode guards

**Remove lines 356-360** (L key INSERT guard):
```go
// DELETE: This guard is now handled by the key sink above
if m.focusedPanel == FocusEditor && m.editorPanel.IsInInsertMode() {
    cmd = m.editorPanel.Update(msg)
    return m, cmd
}
```

**Simplify line 455** (navigation guard):
```go
// BEFORE:
if !(m.focusedPanel == FocusEditor && m.editorPanel.IsInInsertMode()) {

// AFTER (remove the INSERT mode check, keep as simple block):
// Navigation keybindings
switch key {
```

The INSERT mode check is no longer needed because the key sink at the top already routed all INSERT mode keys to the editor. By the time execution reaches this point, we know the editor is NOT in INSERT mode.

### Step 3: Clean up debug logging

Remove debug logging that was added specifically to diagnose resize key conflicts:

- **main.go lines 290-295**: LOG1 resize-specific key press logging
- **main.go lines 312-327**: LOG3 resize key comparison logging
- **main.go lines 574-579**: LOG4 pre-panel-update logging

The editor-internal LOG5 debug logs in `editor.go` (lines 133-135) are out of scope for this change — they serve a different diagnostic purpose (tracking INSERT/NORMAL mode routing within the editor).

## Files Modified

| File | Change |
|------|--------|
| `cmd/lazydb/main.go` | Add key sink block, remove redundant guards, clean debug logs |

### Behavioral Note on Catch-All (lines 582-589)

After the key sink is in place, the catch-all `editorPanel.Update(msg)` at line 586 will never be reached for INSERT mode `KeyMsg` events — the key sink handles them all. This is intentional. The catch-all continues to serve:
- Editor in NORMAL mode (vim commands not matched by global handlers)
- Non-KeyMsg messages (window resize, custom messages, etc.)

## Testing

### Manual Test Cases

1. **INSERT mode typing**: Focus editor, ensure INSERT mode, type `SELECT * FROM t WHERE x = 1 AND y = [?]`
   - All characters including `=`, `-`, `[`, `]`, `?` should appear in editor
2. **INSERT mode whitelisted keys**: In INSERT mode, press `ctrl+r` (should execute query), `f2` (should save), `ctrl+q` (should quit), `ctrl+a` (should open AI if enabled)
3. **NORMAL mode commands**: Press `ESC` to enter NORMAL mode, then `=`/`-`/`[`/`]` should resize panels, `?` should open help, `1`/`2`/`3` should navigate, `L` should enter layout preset mode
4. **NORMAL mode vim commands**: In NORMAL mode, verify `i` enters INSERT, `dd` deletes line, `yy` yanks, `h/j/k/l` moves cursor, `w/b` moves by word, `0/$` moves to line start/end, `gg/G` moves to file start/end
5. **Mode transition**: Type in INSERT mode, press `ESC`, immediately press `=` — should resize panel (proving key sink is no longer active)
6. **ctrl+c in INSERT mode**: Verify `ctrl+c` in INSERT mode quits the app and saves connections
7. **Schema search**: Navigate to connections, press `s` for schema, `/` for search — typing should work as before
8. **Connections panel**: When connections focused, `a`/`e`/`d` should still work
9. **Layout preset from NORMAL mode**: Press `ESC` to NORMAL, press `L`, then `1`/`2`/`3`/`4` — presets should apply

### Regression Checks

- Tab navigation works in NORMAL mode
- Layout presets (L + 1/2/3/4) work in NORMAL mode
- Dialog keys unaffected (dialog check is before key sink)
- AI assistant overlay unaffected (AI check is before key sink)
- Rapid mode switching (INSERT → ESC → command key) works without lag
