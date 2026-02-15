package trex_utils

import (
	"bufio"
	"fmt"
	"os"

	"golang.org/x/term"
)

// History is assumed to be an interface or type with this method.
// You must provide it (example stub below if you don't have one yet).
type History interface {
	GetAll() []string // returns history entries, oldest first
	Add(string)       // optional – called when you want to save the final line
}

// LineEditor provides basic readline-like editing with history
type LineEditor struct {
	history      History
	historyIndex int    // -1 = current (editing) line
	tempLine     string // saved editing line when navigating history
}

// NewLineEditor creates a new LineEditor
func NewLineEditor(hist History) *LineEditor {
	return &LineEditor{
		history:      hist,
		historyIndex: -1,
		tempLine:     "",
	}
}

// ReadLine reads one line with basic editing and history support
func (le *LineEditor) ReadLine(prompt string) (string, error) {
	fmt.Print(prompt)

	// Enter raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return "", fmt.Errorf("failed to enter raw mode: %w", err)
	}
	defer func() {
		_ = term.Restore(int(os.Stdin.Fd()), oldState)
	}()

	reader := bufio.NewReader(os.Stdin)

	var line []rune
	var cursor int // cursor position (0..len(line))

	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return "", err
		}

		switch r {
		// ──────────────── Arrow keys ────────────────
		case 27: // ESC → possible escape sequence
			b1, err := reader.ReadByte()
			if err != nil || b1 != '[' {
				continue // plain ESC → ignore
			}

			b2, err := reader.ReadByte()
			if err != nil {
				continue
			}

			switch b2 {
			case 'A': // Up
				if le.historyIndex == -1 {
					le.tempLine = string(line)
				}
				le.historyPrevious()
				line = []rune(le.getHistoryEntry())
				cursor = len(line)
				le.redraw(prompt, line, cursor)

			case 'B': // Down
				le.historyNext()
				line = []rune(le.getHistoryEntry())
				cursor = len(line)
				le.redraw(prompt, line, cursor)

			case 'C': // Right
				if cursor < len(line) {
					cursor++
					fmt.Print("\x1b[C")
				}

			case 'D': // Left
				if cursor > 0 {
					cursor--
					fmt.Print("\x1b[D")
				}
			}

		// ──────────────── Backspace ────────────────
		case 127, 8:
			if cursor > 0 {
				line = append(line[:cursor-1], line[cursor:]...)
				cursor--
				le.redraw(prompt, line, cursor)
			}

		// ──────────────── Enter ────────────────
		case '\r', '\n':
			fmt.Println()
			final := string(line)
			// Optional: save to history here (if you want)
			// le.history.Add(final)
			le.tempLine = ""
			return final, nil

		// ──────────────── Ctrl+C ────────────────
		case 3:
			fmt.Println("^C")
			return "", fmt.Errorf("user interrupted")

		// ──────────────── Printable characters ────────────────
		default:
			if r >= 32 && r != 127 {
				// Insert at cursor position
				line = append(line[:cursor], append([]rune{r}, line[cursor:]...)...)
				cursor++
				le.redraw(prompt, line, cursor)
			}
		}
	}
}

func (le *LineEditor) historyPrevious() {
	entries := le.history.GetAll()
	if le.historyIndex < len(entries)-1 {
		le.historyIndex++
	}
}

func (le *LineEditor) historyNext() {
	if le.historyIndex > 0 {
		le.historyIndex--
	} else {
		le.historyIndex = -1 // back to current line
	}
}

func (le *LineEditor) getHistoryEntry() string {
	if le.historyIndex == -1 {
		return le.tempLine
	}

	entries := le.history.GetAll()
	if le.historyIndex >= len(entries) {
		return ""
	}

	// Assuming GetAll() returns oldest → newest
	return entries[len(entries)-1-le.historyIndex]
}

func (le *LineEditor) redraw(prompt string, runes []rune, pos int) {
	// Return to start of line
	fmt.Print("\r")

	// Clear to end of line
	fmt.Print("\x1b[K")

	// Print prompt + current content
	fmt.Print(prompt)
	fmt.Print(string(runes))

	// Move cursor back to correct position
	if pos < len(runes) {
		distance := len(runes) - pos
		fmt.Printf("\x1b[%dD", distance)
	}
}
