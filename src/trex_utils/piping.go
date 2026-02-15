package trex_utils

import (
	"fmt"
	"sort"
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
				// Join all remaining tokens and split properly by comma
				fullArg := strings.TrimSpace(strings.Join(args, " "))
				if fullArg == "" || fullArg == "*" {
					// select * → keep everything
					break
				}
				var fields []string
				for _, f := range strings.Split(fullArg, ",") {
					trimmed := strings.TrimSpace(f)
					if trimmed != "" {
						fields = append(fields, trimmed)
					}
				}
				if len(fields) > 0 {
					selected := make(map[string]interface{})
					for _, field := range fields {
						if val, exists := obj[field]; exists {
							selected[field] = val
						} else {
							// Optional: log missing field
							// fmt.Fprintf(os.Stderr, "warning: field %q not found\n", field)
						}
					}
					current["output"] = selected
				}
			}
		case "pp":
			current["__pretty_print"] = true
		case "tt":
			current["__table_print"] = true
		}
	}

	return current, nil
}

// SplitCommaFields splits comma-separated list and trims each item
func SplitCommaFields(s string) []string {
	var result []string
	for _, f := range strings.Split(s, ",") {
		trimmed := strings.TrimSpace(f)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// TablePrint formats data as a pretty table — vertical for single object, horizontal for array
func TablePrint(data interface{}) string {
	switch v := data.(type) {
	case map[string]interface{}:
		// Single object → vertical table (one row per key-value)
		return formatMapAsVerticalTable(v)
	case []interface{}:
		// Array of objects → classic horizontal table
		return formatTable(v)
	default:
		return PrettyPrint(data)
	}
}

// formatMapAsVerticalTable prints a single map as a vertical key-value table
func formatMapAsVerticalTable(m map[string]interface{}) string {
	if len(m) == 0 {
		return "(empty)\n"
	}

	var result strings.Builder
	var keys []string
	maxKeyLen := 0

	// Collect and sort keys for consistent order
	for key := range m {
		keys = append(keys, key)
		if len(key) > maxKeyLen {
			maxKeyLen = len(key)
		}
	}
	sort.Strings(keys)

	// Top border
	result.WriteString("┌")
	result.WriteString(strings.Repeat("─", maxKeyLen+2))
	result.WriteString("┬")
	result.WriteString(strings.Repeat("─", 42))
	result.WriteString("┐\n")

	// Header
	result.WriteString("│ ")
	result.WriteString(padRight("Field", maxKeyLen))
	result.WriteString(" │ ")
	result.WriteString(padRight("Value", 40))
	result.WriteString(" │\n")

	// Separator
	result.WriteString("├")
	result.WriteString(strings.Repeat("─", maxKeyLen+2))
	result.WriteString("┼")
	result.WriteString(strings.Repeat("─", 42))
	result.WriteString("┤\n")

	// One row per field
	for _, key := range keys {
		valStr := fmt.Sprintf("%v", m[key])
		if len(valStr) > 40 {
			valStr = valStr[:37] + "..."
		}

		result.WriteString("│ ")
		result.WriteString(padRight(key, maxKeyLen))
		result.WriteString(" │ ")
		result.WriteString(padRight(valStr, 40))
		result.WriteString(" │\n")
	}

	// Bottom border
	result.WriteString("└")
	result.WriteString(strings.Repeat("─", maxKeyLen+2))
	result.WriteString("┴")
	result.WriteString(strings.Repeat("─", 42))
	result.WriteString("┘\n")

	return result.String()
}

// formatTable formats array of objects as horizontal table
func formatTable(arr []interface{}) string {
	if len(arr) == 0 {
		return "(empty)\n"
	}

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

	sort.Strings(columns)

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

	// Separator
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
