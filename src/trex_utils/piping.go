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

// SelectFields filters JSON data to specified fields.
// Supports nested access with "::" separator, e.g. "data::user::name"
// and aliasing with "as", e.g. "data::user::name as username"
func SelectFields(data map[string]interface{}, fields []string) map[string]interface{} {
	result := make(map[string]interface{})
	for _, field := range fields {
		field = strings.TrimSpace(field)

		// Support aliasing: "data::user::name as username"
		alias := ""
		if idx := indexOfWord(field, "as"); idx >= 0 {
			alias = strings.TrimSpace(field[idx+3:])
			field = strings.TrimSpace(field[:idx])
		}

		if strings.Contains(field, "::") {
			// Nested field path
			val, ok := getNestedField(data, strings.Split(field, "::"))
			if ok {
				key := alias
				if key == "" {
					// Use last segment as key by default
					parts := strings.Split(field, "::")
					key = parts[len(parts)-1]
				}
				result[key] = val
			}
		} else {
			if val, exists := data[field]; exists {
				key := alias
				if key == "" {
					key = field
				}
				result[key] = val
			}
		}
	}
	return result
}

// getNestedField traverses a map using a path of keys
func getNestedField(data interface{}, path []string) (interface{}, bool) {
	if len(path) == 0 {
		return data, true
	}

	key := strings.TrimSpace(path[0])

	switch v := data.(type) {
	case map[string]interface{}:
		val, exists := v[key]
		if !exists {
			return nil, false
		}
		return getNestedField(val, path[1:])
	case []interface{}:
		// Allow numeric index like "0", "1", etc.
		var idx int
		if _, err := fmt.Sscanf(key, "%d", &idx); err == nil {
			if idx >= 0 && idx < len(v) {
				return getNestedField(v[idx], path[1:])
			}
		}
		return nil, false
	default:
		if len(path) == 0 {
			return v, true
		}
		return nil, false
	}
}

// indexOfWord finds the byte index of a whole word in s, or -1 if not found.
func indexOfWord(s, word string) int {
	lower := strings.ToLower(s)
	target := " " + word + " "
	idx := strings.Index(lower, target)
	if idx >= 0 {
		return idx
	}
	return -1
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

	keys := sortedKeys(m)
	for _, key := range keys {
		val := m[key]
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
			// Join all remaining tokens and split by comma
			fullArg := strings.TrimSpace(strings.Join(args, " "))
			if fullArg == "" || fullArg == "*" {
				break // select * → keep everything
			}

			fields := SplitCommaFields(fullArg)
			if len(fields) == 0 {
				break
			}

			// Determine the source — could be root, or "output" wrapper
			var source interface{}
			if out, ok := current["output"]; ok {
				source = out
			} else {
				source = current
			}

			switch src := source.(type) {
			case map[string]interface{}:
				current["output"] = SelectFields(src, fields)

			case []interface{}:
				// Apply select to every element of the array
				var selected []interface{}
				for _, item := range src {
					if obj, ok := item.(map[string]interface{}); ok {
						selected = append(selected, SelectFields(obj, fields))
					}
				}
				current["output"] = selected

			default:
				// Scalar or unknown — nothing to select from
			}

		case "pp":
			current["__pretty_print"] = true

		case "tt":
			current["__table_print"] = true
		}
	}

	return current, nil
}

// SplitCommaFields splits a comma-separated list and trims each item.
// Respects "as" aliases that may contain spaces, e.g. "data::user::name as uname"
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

// ─────────────────────────────────────────────
// Table printing
// ─────────────────────────────────────────────

const maxColWidth = 48 // max display width for any single column value

// TablePrint formats data as a pretty table.
// Single object → vertical key/value table.
// Array of objects → horizontal table.
func TablePrint(data interface{}) string {
	switch v := data.(type) {
	case map[string]interface{}:
		return formatMapAsVerticalTable(v)
	case []interface{}:
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

	keys := sortedKeys(m)

	maxKeyLen := 5 // minimum "Field" header width
	for _, key := range keys {
		if len(key) > maxKeyLen {
			maxKeyLen = len(key)
		}
	}

	valWidth := 40 // fixed value column width

	var result strings.Builder

	writeBorder := func(left, mid, right, horiz string) {
		result.WriteString(left)
		result.WriteString(strings.Repeat(horiz, maxKeyLen+2))
		result.WriteString(mid)
		result.WriteString(strings.Repeat(horiz, valWidth+2))
		result.WriteString(right + "\n")
	}

	writeBorder("┌", "┬", "┐", "─")

	// Header
	result.WriteString("│ " + padRight("Field", maxKeyLen) + " │ " + padRight("Value", valWidth) + " │\n")
	writeBorder("├", "┼", "┤", "─")

	for _, key := range keys {
		valStr := flattenValue(m[key])
		// Wrap long values across multiple lines
		lines := wrapString(valStr, valWidth)
		for li, line := range lines {
			if li == 0 {
				result.WriteString("│ " + padRight(key, maxKeyLen) + " │ " + padRight(line, valWidth) + " │\n")
			} else {
				result.WriteString("│ " + padRight("", maxKeyLen) + " │ " + padRight(line, valWidth) + " │\n")
			}
		}
	}

	writeBorder("└", "┴", "┘", "─")
	return result.String()
}

// formatTable formats an array of objects as a horizontal table
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
				w := len(flattenValue(obj[key]))
				if w > columnWidths[key] {
					columnWidths[key] = w
				}
			}
		}
	}

	if len(items) == 0 {
		return PrettyPrint(arr)
	}

	sort.Strings(columns)

	// Clamp each column and ensure header fits
	for _, col := range columns {
		w := columnWidths[col]
		if w > maxColWidth {
			w = maxColWidth
		}
		if len(col) > w {
			w = len(col)
		}
		columnWidths[col] = w
	}

	var result strings.Builder

	writeBorderRow := func(left, mid, right, horiz string) {
		result.WriteString(left)
		for i, col := range columns {
			result.WriteString(strings.Repeat(horiz, columnWidths[col]+2))
			if i < len(columns)-1 {
				result.WriteString(mid)
			}
		}
		result.WriteString(right + "\n")
	}

	writeBorderRow("┌", "┬", "┐", "─")

	// Header row
	result.WriteString("│")
	for _, col := range columns {
		result.WriteString(" " + padRight(col, columnWidths[col]) + " │")
	}
	result.WriteString("\n")

	writeBorderRow("├", "┼", "┤", "─")

	// Data rows
	for _, item := range items {
		// Collect wrapped lines per column
		colLines := make(map[string][]string)
		maxLines := 1
		for _, col := range columns {
			lines := wrapString(flattenValue(item[col]), columnWidths[col])
			colLines[col] = lines
			if len(lines) > maxLines {
				maxLines = len(lines)
			}
		}

		for li := 0; li < maxLines; li++ {
			result.WriteString("│")
			for _, col := range columns {
				lines := colLines[col]
				cell := ""
				if li < len(lines) {
					cell = lines[li]
				}
				result.WriteString(" " + padRight(cell, columnWidths[col]) + " │")
			}
			result.WriteString("\n")
		}
	}

	writeBorderRow("└", "┴", "┘", "─")
	return result.String()
}

// ─────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────

// flattenValue converts any value to a display string, truncating nested objects
func flattenValue(v interface{}) string {
	switch val := v.(type) {
	case nil:
		return "null"
	case string:
		return val
	case float64:
		return fmt.Sprintf("%g", val)
	case bool:
		return fmt.Sprintf("%v", val)
	case map[string]interface{}:
		// Inline compact JSON-like representation
		parts := make([]string, 0, len(val))
		for k, vv := range val {
			parts = append(parts, fmt.Sprintf("%s:%v", k, vv))
		}
		sort.Strings(parts)
		return "{" + strings.Join(parts, ", ") + "}"
	case []interface{}:
		strs := make([]string, 0, len(val))
		for _, item := range val {
			strs = append(strs, flattenValue(item))
		}
		return "[" + strings.Join(strs, ", ") + "]"
	default:
		return fmt.Sprintf("%v", val)
	}
}

// wrapString splits s into lines of at most width runes
func wrapString(s string, width int) []string {
	if width <= 0 {
		return []string{s}
	}
	var lines []string
	runes := []rune(s)
	for len(runes) > width {
		lines = append(lines, string(runes[:width]))
		runes = runes[width:]
	}
	lines = append(lines, string(runes))
	return lines
}

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func padRight(s string, length int) string {
	runes := []rune(s)
	if len(runes) >= length {
		return string(runes[:length])
	}
	return s + strings.Repeat(" ", length-len(runes))
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// formatMapAsTable is kept for backward compatibility
func formatMapAsTable(m map[string]interface{}) string {
	return formatMapAsVerticalTable(m)
}
