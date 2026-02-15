package trex_utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// PythonExecutor executes Python3 modules
type PythonExecutor struct {
	pythonPath string
	modulePath string
}

// NewPythonExecutor creates a new Python executor
func NewPythonExecutor(modulePath string) *PythonExecutor {
	pythonPath := "python3"
	return &PythonExecutor{
		pythonPath: pythonPath,
		modulePath: modulePath,
	}
}

// Execute runs a Python module with given arguments
func (p *PythonExecutor) Execute(moduleName string, args []string) (map[string]interface{}, error) {
	cmd := exec.Command(p.pythonPath, "-m", moduleName)

	if p.modulePath != "" {
		cmd.Env = append(os.Environ(), "PYTHONPATH="+p.modulePath)
	}

	cmd.Args = append(cmd.Args, args...)

	// Stream stderr to the parent process so modules can print live progress there,
	// and capture stdout fully for JSON unmarshalling.
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// Read stderr and print to our stderr as live output
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			fmt.Fprintln(os.Stderr, scanner.Text())
		}
	}()

	// Read all stdout into buffer
	var outBuf bytes.Buffer
	_, err = io.Copy(&outBuf, stdoutPipe)
	if err != nil {
		// ensure process reaped
		cmd.Wait()
		return nil, err
	}

	// Wait for process to finish
	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(outBuf.Bytes(), &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ExecuteInline executes a Python command directly
func (p *PythonExecutor) ExecuteInline(code string) (map[string]interface{}, error) {
	cmd := exec.Command(p.pythonPath, "-c", code)

	if p.modulePath != "" {
		cmd.Env = append(os.Environ(), "PYTHONPATH="+p.modulePath)
	}

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(stdout.Bytes(), &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ParseCommand parses a command line into parts
func ParseCommand(line string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false

	for i, ch := range line {
		if ch == '"' {
			inQuotes = !inQuotes
		} else if ch == ' ' && !inQuotes {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(ch)
		}

		if i == len(line)-1 && current.Len() > 0 {
			parts = append(parts, current.String())
		}
	}

	return parts
}
