package components

import (
	"bytes"
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

	// Use monokai style (popular dark theme)
	style := styles.Get("monokai")
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

// SetTheme changes the color scheme
func (h *SQLHighlighter) SetTheme(themeName string) {
	style := styles.Get(themeName)
	if style != nil {
		h.style = style
	}
}

// Available themes: monokai, dracula, nord, github, vim, vs, etc.
