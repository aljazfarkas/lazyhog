package components

import (
	"fmt"
	"strings"

	"github.com/aljazfarkas/lazyhog/internal/client"
	"github.com/aljazfarkas/lazyhog/internal/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

// RenderTable renders a query result as a table
func RenderTable(result *client.QueryResult, width, height int, scrollX, scrollY int) string {
	if result == nil || len(result.Results) == 0 {
		return styles.DimTextStyle.Render("No results")
	}

	// Calculate column widths
	colWidths := calculateColumnWidths(result, width)

	var sb strings.Builder

	// Header
	header := renderTableHeader(result.Columns, colWidths, scrollX)
	sb.WriteString(styles.HeaderStyle.Render(header))
	sb.WriteString("\n")

	// Separator
	separator := renderSeparator(colWidths, scrollX)
	sb.WriteString(styles.DimTextStyle.Render(separator))
	sb.WriteString("\n")

	// Rows
	maxRows := height - 4
	if maxRows < 1 {
		maxRows = 1
	}

	startRow := scrollY
	if startRow < 0 {
		startRow = 0
	}
	endRow := startRow + maxRows
	if endRow > len(result.Results) {
		endRow = len(result.Results)
	}

	for i := startRow; i < endRow; i++ {
		row := renderTableRow(result.Results[i], colWidths, scrollX)
		sb.WriteString(row)
		sb.WriteString("\n")
	}

	// Info footer
	if len(result.Results) > maxRows {
		info := fmt.Sprintf("Showing %d-%d of %d rows", startRow+1, endRow, len(result.Results))
		sb.WriteString("\n")
		sb.WriteString(styles.DimTextStyle.Render(info))
	}

	return sb.String()
}

func calculateColumnWidths(result *client.QueryResult, maxWidth int) []int {
	numCols := len(result.Columns)
	if numCols == 0 {
		return []int{}
	}

	colWidths := make([]int, numCols)

	// Start with column header widths
	for i, col := range result.Columns {
		colWidths[i] = len(col)
	}

	// Check data widths (sample first 100 rows for performance)
	sampleSize := 100
	if sampleSize > len(result.Results) {
		sampleSize = len(result.Results)
	}

	for i := 0; i < sampleSize; i++ {
		row := result.Results[i]
		for j := 0; j < numCols && j < len(row); j++ {
			cellWidth := len(fmt.Sprintf("%v", row[j]))
			if cellWidth > colWidths[j] {
				colWidths[j] = cellWidth
			}
		}
	}

	// Cap maximum column width
	maxColWidth := 40
	for i := range colWidths {
		if colWidths[i] > maxColWidth {
			colWidths[i] = maxColWidth
		}
		if colWidths[i] < 5 {
			colWidths[i] = 5
		}
	}

	return colWidths
}

func renderTableHeader(columns []string, widths []int, scrollX int) string {
	var sb strings.Builder

	startCol := scrollX
	if startCol < 0 {
		startCol = 0
	}
	if startCol >= len(columns) {
		startCol = len(columns) - 1
	}

	for i := startCol; i < len(columns); i++ {
		if i < len(widths) {
			col := columns[i]
			if len(col) > widths[i] {
				col = styles.TruncateString(col, widths[i])
			}
			sb.WriteString(fmt.Sprintf("%-*s", widths[i], col))
			if i < len(columns)-1 {
				sb.WriteString(" │ ")
			}
		}
	}

	return sb.String()
}

func renderSeparator(widths []int, scrollX int) string {
	var sb strings.Builder

	startCol := scrollX
	if startCol < 0 {
		startCol = 0
	}
	if startCol >= len(widths) {
		startCol = len(widths) - 1
	}

	for i := startCol; i < len(widths); i++ {
		sb.WriteString(strings.Repeat("─", widths[i]))
		if i < len(widths)-1 {
			sb.WriteString("─┼─")
		}
	}

	return sb.String()
}

func renderTableRow(row []interface{}, widths []int, scrollX int) string {
	var sb strings.Builder

	startCol := scrollX
	if startCol < 0 {
		startCol = 0
	}
	if startCol >= len(row) {
		startCol = len(row) - 1
	}

	for i := startCol; i < len(row) && i < len(widths); i++ {
		cell := fmt.Sprintf("%v", row[i])
		if len(cell) > widths[i] {
			cell = styles.TruncateString(cell, widths[i])
		}

		// Style based on type
		cellStyled := cell
		if row[i] == nil {
			cellStyled = styles.JSONNullStyle.Render("null")
		}

		sb.WriteString(fmt.Sprintf("%-*s", widths[i], cellStyled))
		if i < len(row)-1 && i < len(widths)-1 {
			sb.WriteString(" │ ")
		}
	}

	return sb.String()
}

// GetTableDimensions returns the number of visible columns and rows
func GetTableDimensions(result *client.QueryResult, width, height int) (cols, rows int) {
	if result == nil {
		return 0, 0
	}

	cols = len(result.Columns)
	rows = len(result.Results)

	maxRows := height - 4
	if rows > maxRows {
		rows = maxRows
	}

	return cols, rows
}
