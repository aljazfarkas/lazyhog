package components

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aljazfarkas/lazyhog/internal/ui/styles"
)

// FormatJSON formats JSON data for display with syntax highlighting
func FormatJSON(data interface{}) string {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error formatting JSON: %v", err)
	}

	return string(jsonBytes)
}

// FormatJSONWithColors formats JSON with color syntax highlighting
func FormatJSONWithColors(data interface{}, maxLines int) string {
	jsonStr := FormatJSON(data)
	lines := strings.Split(jsonStr, "\n")

	if maxLines > 0 && len(lines) > maxLines {
		lines = lines[:maxLines]
		lines = append(lines, styles.DimTextStyle.Render(fmt.Sprintf("... (%d more lines)", len(strings.Split(jsonStr, "\n"))-maxLines)))
	}

	var colored strings.Builder
	for _, line := range lines {
		colored.WriteString(colorizeJSONLine(line))
		colored.WriteString("\n")
	}

	return colored.String()
}

func colorizeJSONLine(line string) string {
	// Simple colorization - can be enhanced
	trimmed := strings.TrimSpace(line)

	// Keys (text before colon)
	if strings.Contains(line, ":") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]

			// Color the key
			key = styles.JSONKeyStyle.Render(key)

			// Color the value based on type
			valueTrimmed := strings.TrimSpace(value)
			if strings.HasPrefix(valueTrimmed, "\"") {
				value = styles.JSONStringStyle.Render(value)
			} else if valueTrimmed == "true" || valueTrimmed == "false" {
				value = styles.JSONBoolStyle.Render(value)
			} else if valueTrimmed == "null" || valueTrimmed == "null," {
				value = styles.JSONNullStyle.Render(value)
			} else if len(valueTrimmed) > 0 && (valueTrimmed[0] >= '0' && valueTrimmed[0] <= '9') || valueTrimmed[0] == '-' {
				value = styles.JSONNumberStyle.Render(value)
			}

			return key + ":" + value
		}
	}

	// Brackets and braces
	if trimmed == "{" || trimmed == "}" || trimmed == "[" || trimmed == "]" ||
		trimmed == "{," || trimmed == "}," || trimmed == "[," || trimmed == "]," {
		return styles.DimTextStyle.Render(line)
	}

	return line
}

// RenderJSONBox renders JSON in a bordered box
func RenderJSONBox(title string, data interface{}, width int, maxLines int) string {
	jsonContent := FormatJSONWithColors(data, maxLines)

	titleStyled := styles.HeaderStyle.Width(width - 4).Render(title)

	content := styles.BorderStyle.
		Width(width - 4).
		Render(jsonContent)

	return titleStyled + "\n" + content
}
