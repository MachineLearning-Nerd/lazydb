package panels

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/atotto/clipboard"
	"github.com/MachineLearning-Nerd/lazydb/internal/db"
)

// yankRowMsg is sent when a row is successfully yanked to clipboard
type yankRowMsg struct {
	rowIndex int
	success  bool
}

// ResultsPanel represents the right panel showing query results
type ResultsPanel struct {
	width       int
	height      int
	result      *db.QueryResult
	hasData     bool
	scrollX     int // horizontal scroll offset
	scrollY     int // vertical scroll offset
}

// NewResultsPanel creates a new results panel
func NewResultsPanel() *ResultsPanel {
	return &ResultsPanel{
		hasData: false,
		scrollX: 0,
		scrollY: 0,
	}
}

// SetSize sets the panel dimensions
func (p *ResultsPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// SetResult sets the query result to display
func (p *ResultsPanel) SetResult(result db.QueryResult) {
	p.result = &result
	p.hasData = true
	p.scrollX = 0
	p.scrollY = 0
}

// Clear clears the current results
func (p *ResultsPanel) Clear() {
	p.result = nil
	p.hasData = false
	p.scrollX = 0
	p.scrollY = 0
}

// Update handles keyboard input for scrolling and yanking
func (p *ResultsPanel) Update(msg tea.Msg) tea.Cmd {
	if !p.hasData || p.result == nil {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if p.scrollX > 0 {
				p.scrollX -= 5
			}
		case "right", "l":
			p.scrollX += 5
		case "up", "k":
			if p.scrollY > 0 {
				p.scrollY--
			}
		case "down", "j":
			if p.result != nil && p.scrollY < len(p.result.Rows)-1 {
				p.scrollY++
			}
		case "home":
			p.scrollX = 0
		case "end":
			p.scrollX = 1000 // scroll far right
		case "pgup":
			p.scrollY -= 10
			if p.scrollY < 0 {
				p.scrollY = 0
			}
		case "pgdown":
			if p.result != nil {
				p.scrollY += 10
				if p.scrollY >= len(p.result.Rows) {
					p.scrollY = len(p.result.Rows) - 1
				}
			}
		case "y":
			// Yank current row to clipboard
			return p.yankCurrentRow()
		}
	}
	return nil
}

// View renders the results panel
func (p *ResultsPanel) View() string {
	if p.width == 0 || p.height == 0 {
		return ""
	}

	content := "RESULTS\n\n"

	if !p.hasData {
		content += "No results yet.\n\n"
		content += "Execute a query with Ctrl-R"
		return content
	}

	if p.result.Error != nil {
		content += fmt.Sprintf("❌ Error:\n%s", p.result.Error.Error())
		return content
	}

	// Display results in table format
	if len(p.result.Columns) == 0 {
		content += "Query executed successfully.\n"
		content += fmt.Sprintf("(%d rows affected, %dms)", p.result.RowCount, p.result.ExecutionMs)
		return content
	}

	// Calculate column widths (min 10, max 30 characters per column)
	colWidths := make([]int, len(p.result.Columns))
	for i, col := range p.result.Columns {
		colWidths[i] = max(10, len(col))
	}
	for _, row := range p.result.Rows {
		for i, cell := range row {
			if i < len(colWidths) {
				cellLen := len(cell)
				if cellLen > colWidths[i] && colWidths[i] < 30 {
					colWidths[i] = min(30, cellLen)
				}
			}
		}
	}

	// Build table as a string, then apply horizontal scroll
	var tableLines []string

	// Header row
	headerLine := "│ "
	for i, col := range p.result.Columns {
		headerLine += padOrTruncate(col, colWidths[i]) + " │ "
	}
	tableLines = append(tableLines, headerLine)

	// Separator
	separatorLine := "├─"
	for i := range p.result.Columns {
		separatorLine += strings.Repeat("─", colWidths[i]) + "─┼─"
	}
	separatorLine = separatorLine[:len(separatorLine)-2] + "┤"
	tableLines = append(tableLines, separatorLine)

	// Data rows
	maxDisplayRows := p.height - 8 // Leave room for header, footer, etc.
	if maxDisplayRows < 1 {
		maxDisplayRows = 5
	}

	startRow := p.scrollY
	endRow := min(startRow+maxDisplayRows, len(p.result.Rows))

	for rowIdx := startRow; rowIdx < endRow; rowIdx++ {
		row := p.result.Rows[rowIdx]
		rowLine := "│ "
		for i, cell := range row {
			if i < len(colWidths) {
				rowLine += padOrTruncate(cell, colWidths[i]) + " │ "
			}
		}
		tableLines = append(tableLines, rowLine)
	}

	// Apply horizontal scrolling
	scrolledLines := make([]string, len(tableLines))
	for i, line := range tableLines {
		if p.scrollX < len(line) {
			scrolledLines[i] = line[p.scrollX:]
		} else {
			scrolledLines[i] = ""
		}
		// Truncate to panel width
		if len(scrolledLines[i]) > p.width-4 {
			scrolledLines[i] = scrolledLines[i][:p.width-4]
		}
	}

	content += strings.Join(scrolledLines, "\n")
	content += "\n\n"

	// Add summary with scroll indicators
	scrollInfo := ""
	if p.scrollX > 0 {
		scrollInfo += "◄ "
	}
	if len(p.result.Rows) > endRow {
		scrollInfo += fmt.Sprintf("%d rows (showing %d-%d), %dms", p.result.RowCount, startRow+1, endRow, p.result.ExecutionMs)
	} else {
		scrollInfo += fmt.Sprintf("%d rows, %dms", p.result.RowCount, p.result.ExecutionMs)
	}
	if p.scrollX > 0 || len(p.result.Rows) > maxDisplayRows {
		scrollInfo += " ►"
	}

	content += scrollInfo

	return content
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// padOrTruncate pads or truncates a string to the specified width
func padOrTruncate(s string, width int) string {
	if len(s) > width {
		if width > 3 {
			return s[:width-3] + "..."
		}
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

// yankCurrentRow copies the current row to clipboard as tab-separated values
func (p *ResultsPanel) yankCurrentRow() tea.Cmd {
	return func() tea.Msg {
		if p.result == nil || len(p.result.Rows) == 0 {
			return yankRowMsg{rowIndex: -1, success: false}
		}

		// Ensure scrollY is within bounds
		if p.scrollY < 0 || p.scrollY >= len(p.result.Rows) {
			return yankRowMsg{rowIndex: -1, success: false}
		}

		// Get the current row
		row := p.result.Rows[p.scrollY]
		
		// Join cells with tabs for easy pasting into spreadsheets
		textToCopy := strings.Join(row, "\t")
		
		// Copy to clipboard
		err := clipboard.WriteAll(textToCopy)
		if err != nil {
			return yankRowMsg{rowIndex: p.scrollY, success: false}
		}

		return yankRowMsg{rowIndex: p.scrollY, success: true}
	}
}

// Help returns help text for the results panel
func (p *ResultsPanel) Help() string {
	return "[←→] scroll horizontal  [↑↓] scroll vertical  [Home/End] jump  [y] yank row"
}
