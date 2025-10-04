package components

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// SQLHighlighter provides syntax highlighting for SQL queries
type SQLHighlighter struct {
	lexer     chroma.Lexer
	formatter chroma.Formatter
	style     *chroma.Style
}

// NewSQLHighlighter creates a new SQL syntax highlighter
func NewSQLHighlighter() *SQLHighlighter {
	// Get PostgreSQL lexer
	lexer := lexers.Get("postgresql")
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	// Use terminal formatter with true color support
	formatter := formatters.Get("terminal16m")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	// Use dracula style (brighter, better contrast)
	style := styles.Get("dracula")
	if style == nil {
		style = styles.Fallback
	}

	return &SQLHighlighter{
		lexer:     lexer,
		formatter: formatter,
		style:     style,
	}
}

// Highlight applies syntax highlighting to SQL code
func (h *SQLHighlighter) Highlight(code string) (string, error) {
	// Tokenize the code
	iterator, err := h.lexer.Tokenise(nil, code)
	if err != nil {
		return code, err
	}

	// Format with syntax highlighting
	var buf bytes.Buffer
	err = h.formatter.Format(&buf, h.style, iterator)
	if err != nil {
		return code, err
	}

	return buf.String(), nil
}

// HighlightLines applies syntax highlighting line by line (for editor display)
func (h *SQLHighlighter) HighlightLines(code string) []string {
	highlighted, err := h.Highlight(code)
	if err != nil {
		// Fallback to original code if highlighting fails
		return strings.Split(code, "\n")
	}
	return strings.Split(highlighted, "\n")
}

// HighlightWithCursor applies syntax highlighting and injects a visible cursor
func (h *SQLHighlighter) HighlightWithCursor(code string, cursorLine, cursorCol int) (string, error) {
	// Get highlighted text first
	highlighted, err := h.Highlight(code)
	if err != nil {
		return code, err
	}

	// Split into lines
	lines := strings.Split(highlighted, "\n")

	// Check bounds
	if cursorLine < 0 || cursorLine >= len(lines) {
		return highlighted, nil
	}

	// Get the line where cursor is
	line := lines[cursorLine]

	// Remove ANSI codes to find actual character position
	cleanLine := stripANSI(line)

	// Check column bounds
	if cursorCol < 0 || cursorCol > len(cleanLine) {
		return highlighted, nil
	}

	// Build line with cursor
	// We need to find the position in the ANSI-coded string that corresponds to cursorCol
	result := injectCursorAtPosition(line, cleanLine, cursorCol)
	lines[cursorLine] = result

	return strings.Join(lines, "\n"), nil
}

// stripANSI removes ANSI escape codes from text
func stripANSI(text string) string {
	// ANSI escape sequence pattern: ESC[...m
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiRegex.ReplaceAllString(text, "")
}

// injectCursorAtPosition injects cursor styling at the correct position
func injectCursorAtPosition(ansiLine, cleanLine string, col int) string {
	if col >= len(cleanLine) {
		// Cursor at end of line - add cursor block
		return ansiLine + "\x1b[48;5;51m \x1b[0m" // Bright cyan background space
	}

	// Count through the ANSI string to find where the character at col is
	var result strings.Builder
	cleanIdx := 0
	inEscape := false
	escapeSeq := ""

	i := 0
	for i < len(ansiLine) {
		ch := ansiLine[i]

		if ch == '\x1b' {
			// Start of ANSI escape sequence
			inEscape = true
			escapeSeq = string(ch)
			i++
			continue
		}

		if inEscape {
			escapeSeq += string(ch)
			if ch == 'm' {
				// End of escape sequence
				result.WriteString(escapeSeq)
				inEscape = false
				escapeSeq = ""
			}
			i++
			continue
		}

		// Regular character
		if cleanIdx == col {
			// This is where we insert the cursor!
			// Reset any previous formatting, apply cursor style, write char, reset
			result.WriteString("\x1b[0m\x1b[7m") // Reset + Inverse
			result.WriteByte(ch)
			result.WriteString("\x1b[27m") // Un-inverse
			// Re-apply any color that was active
			i++
			cleanIdx++

			// Continue with rest of line
			result.WriteString(ansiLine[i:])
			return result.String()
		}

		result.WriteByte(ch)
		cleanIdx++
		i++
	}

	return result.String()
}

// SetTheme changes the color scheme
func (h *SQLHighlighter) SetTheme(themeName string) {
	style := styles.Get(themeName)
	if style != nil {
		h.style = style
	}
}

// Available themes: monokai, dracula, nord, github, vim, vs, etc.
