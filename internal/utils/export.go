package utils

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/aljazfarkas/lazyhog/internal/client"
)

// ExportToCSV exports query results to a CSV file
func ExportToCSV(result *client.QueryResult, filename string) error {
	if result == nil {
		return fmt.Errorf("no result to export")
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write(result.Columns); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write rows
	for _, row := range result.Results {
		strRow := make([]string, len(row))
		for i, cell := range row {
			strRow[i] = fmt.Sprintf("%v", cell)
		}
		if err := writer.Write(strRow); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}
