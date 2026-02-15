package trex_utils

import (
	"bufio"
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

	// enable raw mode with stty so we can read escape sequences reliably
	rawCmd := exec.Command("sh", "-c", "stty raw -echo -icanon min 1 time 0")
	restoreCmd := exec.Command("sh", "-c", "stty sane")
	_ = rawCmd.Run()
	defer restoreCmd.Run()

	reader := bufio.NewReader(os.Stdin)
	var line []rune
	var cursorPos int

	for {
		// read a rune to correctly handle UTF-8 input
		r, _, err := reader.ReadRune()
		if err != nil {
			return "", err
		}

		// ESC sequence handling
		if r == 27 {
			// next should be '[' or 'O'
			b1, err := reader.ReadByte()
			if err != nil {
				continue
			}
			if b1 != '[' && b1 != 'O' {
				continue
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
				line = []rune(le.getHistoryLine())
				cursorPos = len(line)
				le.redrawLine(prompt, line, cursorPos)
			case 'B': // Down
				le.historyNext()
				line = []rune(le.getHistoryLine())
				cursorPos = len(line)
				le.redrawLine(prompt, line, cursorPos)
			case 'C': // Right
				if cursorPos < len(line) {
					cursorPos++
					fmt.Print("\033[C")
				}
			case 'D': // Left
				if cursorPos > 0 {
					cursorPos--
					fmt.Print("\033[D")
				}
			}
			continue
		}

		// Backspace
		if r == 127 || r == 8 {
			if cursorPos > 0 {
				line = append(line[:cursorPos-1], line[cursorPos:]...)
				cursorPos--
				le.redrawLine(prompt, line, cursorPos)
			}
			continue
		}

		// Enter
		if r == '\n' || r == '\r' {
			fmt.Println()
			le.tempLine = ""
			return string(line), nil
		}

		// Ctrl+C
		if r == 3 {
			fmt.Println("^C")
			return "", fmt.Errorf("interrupted")
		}

		// Printable
		if r >= 32 {
			if cursorPos == len(line) {
				line = append(line, r)
				fmt.Print(string(r))
			} else {
				// insert at cursor
				line = append(line[:cursorPos], append([]rune{r}, line[cursorPos:]...)...)
				le.redrawLine(prompt, line, cursorPos+1)
			}
			cursorPos++
		}
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
