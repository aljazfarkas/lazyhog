package miller

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aljazfarkas/lazyhog/internal/ui/styles"
)

// toggleJSONFoldAtCursor toggles the fold state of the JSON object at the current scroll position
func (m *Model) toggleJSONFoldAtCursor() {
	// For now, we'll implement a simplified version that toggles all folding
	// A full implementation would track which specific line/path the cursor is on
	m.allFolded = !m.allFolded

	// Clear the fold state map to reset individual folds
	m.jsonFoldState = make(map[string]bool)
}

// jsonFoldAll toggles all top-level JSON keys folded/expanded
func (m *Model) jsonFoldAll() {
	m.allFolded = !m.allFolded

	// Clear individual fold states
	m.jsonFoldState = make(map[string]bool)
}

// renderFoldedJSON renders JSON with fold indicators
// Returns a slice of lines that can be rendered with scrolling
func (m Model) renderFoldedJSON(data interface{}, scrollOffset int) []string {
	if data == nil {
		return []string{styles.DimTextStyle.Render("(no data)")}
	}

	// Convert to JSON first to ensure we can parse it
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return []string{styles.ErrorTextStyle.Render(fmt.Sprintf("Error: %v", err))}
	}

	// If not folded, return the pretty-printed JSON with syntax highlighting
	if !m.allFolded {
		return m.renderColoredJSON(string(jsonBytes))
	}

	// If folded, show compact representation
	return m.renderFoldedCompact(data)
}

// renderColoredJSON renders JSON with basic syntax highlighting
func (m Model) renderColoredJSON(jsonStr string) []string {
	lines := strings.Split(jsonStr, "\n")
	var result []string

	for _, line := range lines {
		// Apply basic coloring based on content
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "\"") && strings.Contains(trimmed, ":") {
			// This is likely a key
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := styles.JSONKeyStyle.Render(parts[0])
				value := parts[1]
				result = append(result, key+":"+value)
				continue
			}
		}

		// Default styling
		result = append(result, line)
	}

	return result
}

// renderFoldedCompact renders JSON in a compact folded format
func (m Model) renderFoldedCompact(data interface{}) []string {
	var lines []string

	// Add fold indicator
	lines = append(lines, styles.DimTextStyle.Render("▶ {...} (folded - press Space or Shift+Z to expand)"))
	lines = append(lines, "")

	// Try to convert to map to show top-level keys
	if mapData, ok := data.(map[string]interface{}); ok {
		lines = append(lines, styles.JSONKeyStyle.Render("Top-level keys:"))
		for key := range mapData {
			lines = append(lines, "  • "+styles.DimTextStyle.Render(key))
		}
	} else {
		// For non-map data, show type
		jsonBytes, _ := json.Marshal(data)
		preview := string(jsonBytes)
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		lines = append(lines, styles.DimTextStyle.Render(preview))
	}

	return lines
}

// getJSONLineCount returns the total number of lines in the rendered JSON
// Used for calculating scroll bounds
func (m Model) getJSONLineCount() int {
	if m.inspectorData == nil {
		return 0
	}

	lines := m.renderFoldedJSON(m.inspectorData, 0)
	return len(lines)
}
