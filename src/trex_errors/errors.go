package trex_errors

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// ANSI color codes (avoid importing trex_utils to prevent cycles)
const (
	ansiReset = "\033[0m"
	ansiRed   = "\033[31m"
	ansiCyan  = "\033[36m"
	ansiBold  = "\033[1m"
	ansiDim   = "\033[2m"
)

// ErrorType represents different error categories
type ErrorType string

const (
	ParseError    ErrorType = "PARSE_ERROR"
	RuntimeError  ErrorType = "RUNTIME_ERROR"
	ModuleError   ErrorType = "MODULE_ERROR"
	ConfigError   ErrorType = "CONFIG_ERROR"
	FileNotFound  ErrorType = "FILE_NOT_FOUND"
	InvalidModule ErrorType = "INVALID_MODULE"
)

// TRexError represents a T-Rex shell error with rich formatting
type TRexError struct {
	Type    ErrorType
	Message string
	File    string
	Line    int
	Context string
	Hint    string
}

// NewError creates a new T-Rex error
func NewError(t ErrorType, msg string) *TRexError {
	return &TRexError{
		Type:    t,
		Message: msg,
	}
}

// WithLocation adds file and line information
func (e *TRexError) WithLocation(file string, line int) *TRexError {
	e.File = file
	e.Line = line
	return e
}

// WithContext adds code context
func (e *TRexError) WithContext(ctx string) *TRexError {
	e.Context = ctx
	return e
}

// WithHint adds a helpful hint
func (e *TRexError) WithHint(hint string) *TRexError {
	e.Hint = hint
	return e
}

// Format returns a Rust-style error message
func (e *TRexError) Format() string {
	builder := strings.Builder{}
	// Header with colored type
	builder.WriteString(ansiRed + ansiBold + "× " + string(e.Type) + ansiReset + "\n\n")
	builder.WriteString("  " + e.Message + "\n")

	if e.File != "" && e.Line > 0 {
		builder.WriteString(fmt.Sprintf("\n   ╭─[%s:%d:1]\n", e.File, e.Line))
		builder.WriteString(fmt.Sprintf(" %d │ %s\n", e.Line, e.Context))
		// Use rune count for correct width when context contains Unicode
		ctxWidth := utf8.RuneCountInString(e.Context)
		builder.WriteString("   │ " + strings.Repeat("─", ctxWidth) + "\n")
		builder.WriteString("   ╰────\n")
	}

	if e.Hint != "" {
		builder.WriteString(fmt.Sprintf("  %s %s\n", ansiCyan+"help:"+ansiReset, e.Hint))
	}

	return builder.String()
}

// Error implements error interface
func (e *TRexError) Error() string {
	return e.Format()
}
