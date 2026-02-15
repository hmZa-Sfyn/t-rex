package trex_utils

import (
	"bufio"
	"fmt"
	"os"
	"syscall"
	"unsafe"
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

	// ─── Enter raw mode using syscall/ioctl ───────────────────────
	oldTermios, err := enableRawMode()
	if err != nil {
		return "", fmt.Errorf("failed to enter raw mode: %w", err)
	}
	defer func() {
		_ = restoreTermios(oldTermios)
	}()

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
			// next should be '['
			b1, err := reader.ReadByte()
			if err != nil || b1 != '[' {
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
		if r >= 32 && r != 127 {
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
		return le.tempLine
	}

	historyEntries := le.history.GetAll()
	if le.historyIndex >= len(historyEntries) {
		return ""
	}

	// Most recent entry is at the end
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

// ────────────────────────────────────────────────
// Raw mode implementation (Linux/macOS/Unix only)
// ────────────────────────────────────────────────

type termios struct {
	Iflag  uint32
	Oflag  uint32
	Cflag  uint32
	Lflag  uint32
	Line   uint8
	Cc     [32]uint8
	Ispeed uint32
	Ospeed uint32
}

const (
	TCGETS = 0x5401
	TCSETS = 0x5402

	IGNBRK = 0000001
	BRKINT = 0000002
	PARMRK = 0000010
	ISTRIP = 0000020
	INLCR  = 0000040
	IGNCR  = 0000100
	ICRNL  = 0000200
	IXON   = 0002000
	OPOST  = 0000001
	ECHO   = 0000010
	ECHONL = 0000100
	ICANON = 0000002
	ISIG   = 0000001
	IEXTEN = 0100000
	CS8    = 0000060
	CSIZE  = 0000060
	PARENB = 0000400
)

func enableRawMode() (*termios, error) {
	fd := int(os.Stdin.Fd())

	var old termios
	_, _, e := syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(fd),
		TCGETS,
		uintptr(unsafe.Pointer(&old)),
		0, 0, 0,
	)
	if e != 0 {
		return nil, syscall.Errno(e)
	}

	raw := old
	raw.Iflag &^= uint32(IGNBRK | BRKINT | PARMRK | ISTRIP | INLCR | IGNCR | ICRNL | IXON)
	raw.Oflag &^= uint32(OPOST)
	raw.Lflag &^= uint32(ECHO | ECHONL | ICANON | ISIG | IEXTEN)
	raw.Cflag &^= uint32(CSIZE | PARENB)
	raw.Cflag |= uint32(CS8)
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0

	_, _, e = syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(fd),
		TCSETS,
		uintptr(unsafe.Pointer(&raw)),
		0, 0, 0,
	)
	if e != 0 {
		return nil, syscall.Errno(e)
	}

	return &old, nil
}

func restoreTermios(state *termios) error {
	fd := int(os.Stdin.Fd())
	_, _, e := syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(fd),
		TCSETS,
		uintptr(unsafe.Pointer(state)),
		0, 0, 0,
	)
	if e != 0 {
		return syscall.Errno(e)
	}
	return nil
}
