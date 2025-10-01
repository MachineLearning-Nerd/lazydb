package panels

import (
	"fmt"
	"strings"

	"github.com/MachineLearning-Nerd/lazydb/internal/db"
)

// ResultsPanel represents the right panel showing query results
type ResultsPanel struct {
	width   int
	height  int
	result  *db.QueryResult
	hasData bool
}

// NewResultsPanel creates a new results panel
func NewResultsPanel() *ResultsPanel {
	return &ResultsPanel{
		hasData: false,
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
}

// Clear clears the current results
func (p *ResultsPanel) Clear() {
	p.result = nil
	p.hasData = false
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

	// Calculate column widths
	colWidths := make([]int, len(p.result.Columns))
	for i, col := range p.result.Columns {
		colWidths[i] = len(col)
	}
	for _, row := range p.result.Rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Limit column widths to fit in panel
	maxColWidth := 20
	for i := range colWidths {
		if colWidths[i] > maxColWidth {
			colWidths[i] = maxColWidth
		}
	}

	// Build header
	header := "║"
	separator := "╠"
	for i, col := range p.result.Columns {
		header += " " + padOrTruncate(col, colWidths[i]) + " ║"
		separator += strings.Repeat("═", colWidths[i]+2) + "╬"
	}
	separator = separator[:len(separator)-1] + "╣"

	// Build top border
	topBorder := "╔"
	for i := range p.result.Columns {
		topBorder += strings.Repeat("═", colWidths[i]+2) + "╦"
	}
	topBorder = topBorder[:len(topBorder)-1] + "╗"

	content += topBorder + "\n"
	content += header + "\n"
	content += separator + "\n"

	// Build rows (limit to fit in panel)
	maxRows := 10
	displayRows := p.result.Rows
	if len(displayRows) > maxRows {
		displayRows = displayRows[:maxRows]
	}

	for _, row := range displayRows {
		rowStr := "║"
		for i, cell := range row {
			if i < len(colWidths) {
				rowStr += " " + padOrTruncate(cell, colWidths[i]) + " ║"
			}
		}
		content += rowStr + "\n"
	}

	// Build bottom border
	bottomBorder := "╚"
	for i := range p.result.Columns {
		bottomBorder += strings.Repeat("═", colWidths[i]+2) + "╩"
	}
	bottomBorder = bottomBorder[:len(bottomBorder)-1] + "╝"

	content += bottomBorder + "\n\n"

	// Add summary
	if len(p.result.Rows) > maxRows {
		content += fmt.Sprintf("%d rows (showing %d), %dms", p.result.RowCount, maxRows, p.result.ExecutionMs)
	} else {
		content += fmt.Sprintf("%d rows, %dms", p.result.RowCount, p.result.ExecutionMs)
	}

	return content
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

// Help returns help text for the results panel
func (p *ResultsPanel) Help() string {
	return "[j/k] scroll  [y] copy  [e] export"
}
