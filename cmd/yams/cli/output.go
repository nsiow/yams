package cli

import (
	"fmt"
	"os"
	"strings"
)

const (
	FormatJSON  = "json"
	FormatTable = "table"
)

// TableWriter formats data as an aligned table
type TableWriter struct {
	headers []string
	rows    [][]string
	widths  []int
}

// NewTableWriter creates a new table writer with the given headers
func NewTableWriter(headers ...string) *TableWriter {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	return &TableWriter{
		headers: headers,
		widths:  widths,
	}
}

// AddRow adds a row to the table
func (t *TableWriter) AddRow(values ...string) {
	// Pad with empty strings if needed
	for len(values) < len(t.headers) {
		values = append(values, "")
	}
	// Truncate if too many values
	if len(values) > len(t.headers) {
		values = values[:len(t.headers)]
	}

	t.rows = append(t.rows, values)
	for i, v := range values {
		if len(v) > t.widths[i] {
			t.widths[i] = len(v)
		}
	}
}

// Render outputs the table to stdout
func (t *TableWriter) Render() {
	// Print headers
	for i, h := range t.headers {
		if i > 0 {
			fmt.Print("  ")
		}
		fmt.Printf("%-*s", t.widths[i], strings.ToUpper(h))
	}
	fmt.Println()

	// Print separator
	for i, w := range t.widths {
		if i > 0 {
			fmt.Print("  ")
		}
		fmt.Print(strings.Repeat("-", w))
	}
	fmt.Println()

	// Print rows
	for _, row := range t.rows {
		for i, v := range row {
			if i > 0 {
				fmt.Print("  ")
			}
			fmt.Printf("%-*s", t.widths[i], v)
		}
		fmt.Println()
	}
}

// OutputJSON writes JSON data to stdout
func OutputJSON(data []byte) {
	os.Stdout.Write(data)
}

// Truncate shortens a string to max length, adding ellipsis if needed
func Truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
