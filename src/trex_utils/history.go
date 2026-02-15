package trex_utils

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// History manages command history
type History struct {
	commands []string
	filePath string
	maxSize  int
}

// NewHistory creates a new history manager
func NewHistory(maxSize int) *History {
	h := &History{
		commands: []string{},
		maxSize:  maxSize,
	}

	homeDir, err := os.UserHomeDir()
	if err == nil {
		trexDir := filepath.Join(homeDir, ".t-rex")
		os.MkdirAll(trexDir, 0755)
		h.filePath = filepath.Join(trexDir, "history")
		h.loadHistory()
	}

	return h
}

// Add adds a command to history
func (h *History) Add(cmd string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return
	}
	h.commands = append(h.commands, cmd)
	if len(h.commands) > h.maxSize {
		h.commands = h.commands[1:]
	}
	h.saveHistory()
}

// loadHistory loads history from file
func (h *History) loadHistory() {
	if h.filePath == "" {
		return
	}

	file, err := os.Open(h.filePath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		h.commands = append(h.commands, scanner.Text())
	}

	if len(h.commands) > h.maxSize {
		h.commands = h.commands[len(h.commands)-h.maxSize:]
	}
}

// saveHistory saves history to file
func (h *History) saveHistory() {
	if h.filePath == "" {
		return
	}

	file, err := os.Create(h.filePath)
	if err != nil {
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, cmd := range h.commands {
		writer.WriteString(cmd + "\n")
	}
	writer.Flush()
}

// GetAll returns all commands
func (h *History) GetAll() []string {
	return h.commands
}

// GetLast returns the last n commands
func (h *History) GetLast(n int) []string {
	if n > len(h.commands) {
		n = len(h.commands)
	}
	return h.commands[len(h.commands)-n:]
}
