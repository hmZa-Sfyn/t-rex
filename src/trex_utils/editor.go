package trex_utils

import (
	"fmt"
	"os"
	"os/exec"
)

// LineEditor handles interactive line editing with history
type LineEditor struct {
	history      *History
	historyIndex int
	tempLine     string
}

// NewLineEditor creates a new line editor
func NewLineEditor(hist *History) *LineEditor {
	return &LineEditor{
		history:      hist,
		historyIndex: -1,
	}
}

// ReadLine reads a line from input with history support
func (le *LineEditor) ReadLine(prompt string) (string, error) {
	fmt.Print(prompt)

	// enable raw mode so we can read arrow keys immediately
	// fall back to cooked mode if stty is not available
	rawCmd := exec.Command("sh", "-c", "stty raw -echo -icanon min 1 time 0")
	restoreCmd := exec.Command("sh", "-c", "stty sane")
	_ = rawCmd.Run()
	defer restoreCmd.Run()

	// Read input character by character to handle arrow keys
	var line []rune
	var cursorPos int

	for {
		// Read one byte
		buf := make([]byte, 1)
		_, err := os.Stdin.Read(buf)
		if err != nil {
			return "", err
		}

		ch := buf[0]

		// Handle escape sequences (arrow keys, etc)
		if ch == 27 { // ESC
			// Read the next byte (should be '[' or 'O')
			b1 := make([]byte, 1)
			n, err := os.Stdin.Read(b1)
			if err != nil || n == 0 {
				continue
			}

			if b1[0] != '[' && b1[0] != 'O' {
				continue
			}

			// Read the final byte that indicates the arrow
			b2 := make([]byte, 1)
			n, err = os.Stdin.Read(b2)
			if err != nil || n == 0 {
				continue
			}

			switch b2[0] {
			case 'A': // Up arrow - previous history
				if le.historyIndex == -1 {
					le.tempLine = string(line)
				}
				le.historyPrevious()
				line = []rune(le.getHistoryLine())
				cursorPos = len(line)
				le.redrawLine(prompt, line, cursorPos)

			case 'B': // Down arrow - next history
				le.historyNext()
				line = []rune(le.getHistoryLine())
				cursorPos = len(line)
				le.redrawLine(prompt, line, cursorPos)

			case 'C': // Right arrow - move cursor right
				if cursorPos < len(line) {
					cursorPos++
					fmt.Print("\033[C")
				}

			case 'D': // Left arrow - move cursor left
				if cursorPos > 0 {
					cursorPos--
					fmt.Print("\033[D")
				}
			}
			continue
		}

		// Handle backspace
		if ch == 127 || ch == 8 {
			if cursorPos > 0 {
				line = append(line[:cursorPos-1], line[cursorPos:]...)
				cursorPos--
				le.redrawLine(prompt, line, cursorPos)
			}
			continue
		}

		// Handle Enter
		if ch == '\n' || ch == '\r' {
			fmt.Println()
			// clear temporary saved line when accepting input
			le.tempLine = ""
			return string(line), nil
		}

		// Handle Ctrl+C
		if ch == 3 {
			fmt.Println("^C")
			return "", fmt.Errorf("interrupted")
		}

		// Handle regular character
		if ch >= 32 && ch < 127 {
			if cursorPos == len(line) {
				line = append(line, rune(ch))
				fmt.Print(string(ch))
			} else {
				line = append(line[:cursorPos+1], line[cursorPos:]...)
				line[cursorPos] = rune(ch)
				// Redraw line from cursor position
				for i := cursorPos; i < len(line); i++ {
					fmt.Print(string(line[i]))
				}
				// Move cursor back to proper position
				fmt.Printf("\033[%dD", len(line)-cursorPos-1)
			}
			cursorPos++

			// If user types while browsing history, keep tempLine until they return
		}
	}
}

// historyPrevious moves to previous history entry
func (le *LineEditor) historyPrevious() {
	historyEntries := le.history.GetAll()
	if le.historyIndex < len(historyEntries)-1 {
		le.historyIndex++
	}
}

// historyNext moves to next history entry
func (le *LineEditor) historyNext() {
	if le.historyIndex > 0 {
		le.historyIndex--
	} else if le.historyIndex == 0 {
		le.historyIndex = -1
	}
}

// getHistoryLine returns the current history line
func (le *LineEditor) getHistoryLine() string {
	if le.historyIndex == -1 {
		// Return the temporarily-saved current line when not in history
		return le.tempLine
	}

	historyEntries := le.history.GetAll()
	if le.historyIndex >= len(historyEntries) {
		return ""
	}

	return historyEntries[len(historyEntries)-1-le.historyIndex]
}

// redrawLine redraws the current line
func (le *LineEditor) redrawLine(prompt string, line []rune, cursorPos int) {
	// Move to start of line
	fmt.Print("\r")

	// Clear the line
	fmt.Print("\033[K")

	// Print prompt and line
	fmt.Print(prompt)
	fmt.Print(string(line))

	// Move cursor to correct position
	if cursorPos < len(line) {
		moveBack := len(line) - cursorPos
		fmt.Printf("\033[%dD", moveBack)
	}
}
