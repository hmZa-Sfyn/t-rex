package trex_utils

import (
	"fmt"
	"strings"
)

// Pipeline represents a piped command sequence
type Pipeline struct {
	commands []string
}

// NewPipeline creates a new pipeline from command line
func NewPipeline(line string) *Pipeline {
	parts := strings.Split(line, "|")
	var commands []string
	for _, part := range parts {
		commands = append(commands, strings.TrimSpace(part))
	}
	return &Pipeline{commands: commands}
}

// HasPipe checks if line contains pipe
func HasPipe(line string) bool {
	return strings.Contains(line, "|")
}

// SelectFields filters JSON data to specified fields
func SelectFields(data map[string]interface{}, fields []string) map[string]interface{} {
	result := make(map[string]interface{})
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if val, exists := data[field]; exists {
			result[field] = val
		}
	}
	return result
}

// FilterArray filters an array of objects to specified fields
func FilterArray(arr []interface{}, fields []string) []map[string]interface{} {
	var results []map[string]interface{}
	for _, item := range arr {
		if obj, ok := item.(map[string]interface{}); ok {
			results = append(results, SelectFields(obj, fields))
		}
	}
	return results
}

// PrettyPrint formats output for pretty printing
func PrettyPrint(data interface{}) string {
	switch v := data.(type) {
	case map[string]interface{}:
		return formatMap(v, 0)
	case []interface{}:
		return formatArray(v, 0)
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

func formatMap(m map[string]interface{}, indent int) string {
	var result strings.Builder
	prefix := strings.Repeat("  ", indent)

	for key, val := range m {
		result.WriteString(prefix)
		result.WriteString(ColoredText(key, Cyan) + ": ")

		switch v := val.(type) {
		case map[string]interface{}:
			result.WriteString("\n")
			result.WriteString(formatMap(v, indent+1))
		case []interface{}:
			result.WriteString("\n")
			result.WriteString(formatArray(v, indent+1))
		case string:
			result.WriteString(ColoredText("\""+v+"\"", Green) + "\n")
		case float64:
			result.WriteString(ColoredText(fmt.Sprintf("%g", v), Yellow) + "\n")
		case bool:
			result.WriteString(ColoredText(fmt.Sprintf("%v", v), Magenta) + "\n")
		case nil:
			result.WriteString(ColoredText("null", Dim) + "\n")
		default:
			result.WriteString(fmt.Sprintf("%v\n", v))
		}
	}

	return result.String()
}

func formatArray(arr []interface{}, indent int) string {
	var result strings.Builder
	prefix := strings.Repeat("  ", indent)

	for i, val := range arr {
		result.WriteString(prefix)
		result.WriteString(ColoredText(fmt.Sprintf("[%d]", i), Blue) + ": ")

		switch v := val.(type) {
		case map[string]interface{}:
			result.WriteString("\n")
			result.WriteString(formatMap(v, indent+1))
		case []interface{}:
			result.WriteString("\n")
			result.WriteString(formatArray(v, indent+1))
		case string:
			result.WriteString(ColoredText("\""+v+"\"", Green) + "\n")
		case float64:
			result.WriteString(ColoredText(fmt.Sprintf("%g", v), Yellow) + "\n")
		case bool:
			result.WriteString(ColoredText(fmt.Sprintf("%v", v), Magenta) + "\n")
		case nil:
			result.WriteString(ColoredText("null", Dim) + "\n")
		default:
			result.WriteString(fmt.Sprintf("%v\n", v))
		}
	}

	return result.String()
}

// ExecutePipeline processes a pipeline of commands
func ExecutePipeline(data map[string]interface{}, pipeline *Pipeline) (map[string]interface{}, error) {
	current := data

	for i, cmd := range pipeline.commands {
		if i == 0 {
			continue // Skip the first command, it's already executed
		}

		parts := ParseCommand(cmd)
		if len(parts) == 0 {
			continue
		}

		op := parts[0]
		args := parts[1:]

		switch op {
		case "select":
			if obj, ok := current["output"].(map[string]interface{}); ok {
				current["output"] = SelectFields(obj, args)
			}
		case "pp":
			current["__pretty_print"] = true
		case "tt":
			current["__table_print"] = true
		}
	}

	return current, nil
}

// TablePrint formats data as a pretty table
func TablePrint(data interface{}) string {
	switch v := data.(type) {
	case []interface{}:
		return formatTable(v)
	case map[string]interface{}:
		return formatMapAsTable(v)
	default:
		return PrettyPrint(data)
	}
}

func formatTable(arr []interface{}) string {
	if len(arr) == 0 {
		return "(empty)\n"
	}

	// Check if all items are maps
	var items []map[string]interface{}
	var columns []string
	columnWidths := make(map[string]int)

	for _, item := range arr {
		if obj, ok := item.(map[string]interface{}); ok {
			items = append(items, obj)
			for key := range obj {
				if !contains(columns, key) {
					columns = append(columns, key)
				}
				width := len(fmt.Sprintf("%v", obj[key]))
				if width > columnWidths[key] {
					columnWidths[key] = width
				}
			}
		}
	}

	if len(items) == 0 {
		return PrettyPrint(arr)
	}

	// Ensure minimum column widths
	for _, col := range columns {
		if columnWidths[col] < len(col) {
			columnWidths[col] = len(col)
		}
	}

	var result strings.Builder

	// Top border
	result.WriteString("┌")
	for i, col := range columns {
		result.WriteString(strings.Repeat("─", columnWidths[col]+2))
		if i < len(columns)-1 {
			result.WriteString("┬")
		}
	}
	result.WriteString("┐\n")

	// Header
	result.WriteString("│")
	for _, col := range columns {
		result.WriteString(" " + padRight(col, columnWidths[col]) + " ")
		result.WriteString("│")
	}
	result.WriteString("\n")

	// Header separator
	result.WriteString("├")
	for i, col := range columns {
		result.WriteString(strings.Repeat("─", columnWidths[col]+2))
		if i < len(columns)-1 {
			result.WriteString("┼")
		}
	}
	result.WriteString("┤\n")

	// Data rows
	for _, item := range items {
		result.WriteString("│")
		for _, col := range columns {
			val := fmt.Sprintf("%v", item[col])
			result.WriteString(" " + padRight(val, columnWidths[col]) + " ")
			result.WriteString("│")
		}
		result.WriteString("\n")
	}

	// Bottom border
	result.WriteString("└")
	for i, col := range columns {
		result.WriteString(strings.Repeat("─", columnWidths[col]+2))
		if i < len(columns)-1 {
			result.WriteString("┴")
		}
	}
	result.WriteString("┘\n")

	return result.String()
}

func formatMapAsTable(m map[string]interface{}) string {
	var result strings.Builder

	maxKeyLen := 0
	for key := range m {
		if len(key) > maxKeyLen {
			maxKeyLen = len(key)
		}
	}

	result.WriteString("┌" + strings.Repeat("─", maxKeyLen+2) + "┬" + strings.Repeat("─", 40) + "┐\n")

	for key, val := range m {
		valStr := fmt.Sprintf("%v", val)
		if len(valStr) > 38 {
			valStr = valStr[:35] + "..."
		}
		result.WriteString("│ " + padRight(key, maxKeyLen) + " │ " + padRight(valStr, 38) + " │\n")
	}

	result.WriteString("└" + strings.Repeat("─", maxKeyLen+2) + "┴" + strings.Repeat("─", 40) + "┘\n")

	return result.String()
}

func padRight(s string, length int) string {
	if len(s) >= length {
		return s[:length]
	}
	return s + strings.Repeat(" ", length-len(s))
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
