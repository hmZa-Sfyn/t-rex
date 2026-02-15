package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"trex_modules"
	"trex_utils"
)

const Version = "1.0.0"

func main() {
	// Define command-line flags
	pathFlag := flag.String("path", "", "Path to custom modules directory")
	bannerFlag := flag.Bool("banner", false, "Show banner and exit")
	execFlag := flag.String("exec", "", "Execute a command and exit")
	versionFlag := flag.Bool("version", false, "Show version information")

	flag.Parse()

	// If a non-flag positional argument is provided and it's a file, execute it as script
	args := flag.Args()

	// Handle version flag
	if *versionFlag {
		showVersion()
		os.Exit(0)
	}

	// Handle banner flag
	if *bannerFlag {
		trex_utils.PrintBanner()
		os.Exit(0)
	}

	shell := NewShell()

	// If a file path was passed as positional arg, execute file and exit
	if len(args) > 0 {
		candidate := args[0]
		if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
			shell.ExecuteFile(candidate)
			os.Exit(0)
		}
	}

	// Handle custom module path
	if *pathFlag != "" {
		if err := shell.SetModulePath(*pathFlag); err != nil {
			trex_utils.PrintError("Invalid module path: " + err.Error())
			os.Exit(1)
		}
	}

	// Handle exec flag - execute command and exit
	if *execFlag != "" {
		shell.ExecuteOnce(*execFlag)
		os.Exit(0)
	}

	// Otherwise run interactive shell
	shell.Run()
}

// showVersion displays version information
func showVersion() {
	fmt.Printf("T-Rex Shell v%s\n", Version)
	fmt.Println("A JSON-based command execution shell")
	fmt.Println("https://github.com/yourusername/t-rex")
}

// Shell represents the T-Rex shell
type Shell struct {
	history        *trex_utils.History
	executor       *trex_utils.PythonExecutor
	loader         *trex_modules.Loader
	moduleDir      string
	useColors      bool
	promptColor    trex_utils.Color
	promptTemplate string
	vars           map[string]string
}

// NewShell creates a new shell instance
func NewShell() *Shell {
	homeDir, _ := os.UserHomeDir()
	moduleDir := filepath.Join(homeDir, ".t-rex", "modules")

	shell := &Shell{
		history:        trex_utils.NewHistory(1000),
		executor:       trex_utils.NewPythonExecutor(moduleDir),
		loader:         trex_modules.NewLoader(moduleDir),
		moduleDir:      moduleDir,
		useColors:      true,
		promptColor:    trex_utils.Cyan,
		promptTemplate: "❯",
		vars:           make(map[string]string),
	}

	loadConfig(shell)
	return shell
}

// SetModulePath sets a custom module path
func (s *Shell) SetModulePath(path string) error {
	// Verify path exists
	if _, err := os.Stat(path); err != nil {
		return err
	}

	s.moduleDir = path
	s.loader = trex_modules.NewLoader(path)
	s.executor = trex_utils.NewPythonExecutor(path)

	return nil
}

// ExecuteOnce executes a single command and returns
func (s *Shell) ExecuteOnce(line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}

	if err := s.executeCommand(line); err != nil {
		// print brief error (detailed logging handled elsewhere)
		trex_utils.PrintError(err.Error())
	}
}

// Run starts the interactive shell
func (s *Shell) Run() {
	trex_utils.PrintBanner()
	fmt.Println()

	// Initialize editor
	editor := trex_utils.NewLineEditor(s.history)

	for {
		prompt := trex_utils.BuildPrompt("❯", s.promptColor, true, true, false)
		line, err := editor.ReadLine(prompt)
		if err != nil {
			fmt.Println()
			trex_utils.PrintExit("Goodbye! 👋")
			os.Exit(0)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for exit
		if line == "exit" || line == "quit" {
			trex_utils.PrintExit("Goodbye! 👋")
			os.Exit(0)
		}

		s.history.Add(line)
		if err := s.executeCommand(line); err != nil {
			// error already logged/printed by lower-level handlers
			trex_utils.PrintError(err.Error())
		}
	}
}

// executeCommand processes a command
func (s *Shell) executeCommand(line string) error {
	// If the line is a file path, execute file
	if fi, err := os.Stat(line); err == nil && !fi.IsDir() {
		s.ExecuteFile(line)
		return nil
	}

	// variable assignment: set NAME VALUE  or let NAME = VALUE
	parts := trex_utils.ParseCommand(line)
	if len(parts) > 0 {
		if parts[0] == "set" || parts[0] == "let" {
			if len(parts) >= 3 {
				name := parts[1]
				if strings.HasPrefix(name, "$") {
					name = name[1:]
				}
				// support optional '=' token
				valParts := parts[2:]
				if valParts[0] == "=" && len(valParts) > 1 {
					valParts = valParts[1:]
				}
				val := strings.Join(valParts, " ")
				s.vars[name] = val
				return nil
			}
		}
	}

	// Check for forloop pattern: forloop RANGE as $var do { ... }
	handled, err := s.handleForLoop(line)
	if err != nil {
		return err
	}
	if handled {
		return nil
	}

	// Check for pipes
	if trex_utils.HasPipe(line) {
		return s.executePipeline(line)
	}

	if len(parts) == 0 {
		return nil
	}

	cmd := parts[0]
	args := parts[1:]

	// Expand variables in args
	for i, a := range args {
		args[i] = s.expandVars(a)
	}

	// Try to execute as Python module
	var result map[string]interface{}
	result, err = s.executeModule(cmd, args)
	if err != nil {
		return err
	}
	if result != nil {
		s.printResult(result)
	}

	return nil
}

// executePipeline handles piped commands
func (s *Shell) executePipeline(line string) error {
	parts := strings.SplitN(line, "|", 2)
	firstCmd := strings.TrimSpace(parts[0])

	// Execute first command
	cmdParts := trex_utils.ParseCommand(firstCmd)
	if len(cmdParts) == 0 {
		return nil
	}

	cmd := cmdParts[0]
	args := cmdParts[1:]

	result, err := s.executeModule(cmd, args)
	if err != nil {
		return err
	}
	if result == nil {
		return nil
	}

	// Process remaining pipes
	pipeRest := strings.TrimSpace(parts[1])
	pipeParts := strings.Split(pipeRest, "|")

	for _, pipe := range pipeParts {
		pipe = strings.TrimSpace(pipe)
		cmdParts := trex_utils.ParseCommand(pipe)
		if len(cmdParts) == 0 {
			continue
		}

		op := cmdParts[0]
		args := cmdParts[1:]

		switch op {
		case "select":
			if output, ok := result["output"].(map[string]interface{}); ok {
				result["output"] = trex_utils.SelectFields(output, args)
			}
		case "pp":
			result["__pretty_print"] = true
		case "tt":
			result["__table_print"] = true
		default:
			// If the op corresponds to a module, execute it with previous output passed as argument
			if modulePath, err := s.loader.FindModule(op); err == nil && modulePath != "" {
				// prepare previous output as string argument
				var prevArg string
				if out, ok := result["output"]; ok {
					// if complex, marshal to JSON string
					switch v := out.(type) {
					case string:
						prevArg = v
					default:
						if b, err := json.Marshal(v); err == nil {
							prevArg = string(b)
						} else {
							prevArg = fmt.Sprintf("%v", v)
						}
					}
				} else {
					prevArg = ""
				}

				callArgs := append(args, prevArg)
				modResult, err := s.executeModule(op, callArgs)
				if err != nil {
					return err
				}
				if modResult == nil {
					// if module failed, stop pipeline
					return nil
				}
				// Replace current result with module result for next steps
				result = modResult
				continue
			}
		}
	}

	s.printResult(result)
	return nil
}

// expandVars replaces $var and ${var} in the input string using shell variables
func (s *Shell) expandVars(input string) string {
	// quick regex replacement
	re := regexp.MustCompile(`\$(?:\{([A-Za-z_][A-Za-z0-9_]*)\}|([A-Za-z_][A-Za-z0-9_]*))`)
	return re.ReplaceAllStringFunc(input, func(m string) string {
		// extract name
		sub := ""
		if strings.HasPrefix(m, "${") && strings.HasSuffix(m, "}") {
			sub = m[2 : len(m)-1]
		} else if strings.HasPrefix(m, "$") {
			sub = m[1:]
		}
		if v, ok := s.vars[sub]; ok {
			return v
		}
		return ""
	})
}

// executeModule executes a Python module
func (s *Shell) executeModule(cmd string, args []string) (map[string]interface{}, error) {
	modulePath, err := s.loader.FindModule(cmd)
	if err != nil {
		s.printModuleNotFound(cmd)
		return nil, os.ErrNotExist
	}

	result, err := s.executor.Execute(cmd, args)
	if err != nil {
		s.printExecutionError(cmd, modulePath, err)
		return nil, err
	}

	return result, nil
}

// ExecuteFile executes commands from a script file (one command per line)
func (s *Shell) ExecuteFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		trex_utils.PrintError("Failed to read script: " + err.Error())
		return
	}

	lines := strings.Split(string(data), "\n")
	fmt.Printf("Running script: %s\n", path)
	for idx, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Prefix with line number for easier debugging
		fmt.Printf("[%d] $ %s\n", idx+1, line)
		s.history.Add(line)
		if err := s.executeCommand(line); err != nil {
			// Write enhanced error info including file and line number
			if home, herr := os.UserHomeDir(); herr == nil {
				trexDir := filepath.Join(home, ".t-rex")
				os.MkdirAll(trexDir, 0755)
				logPath := filepath.Join(trexDir, "error.log")
				f, ferr := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if ferr == nil {
					defer f.Close()
					entry := fmt.Sprintf("TIME: %s\nSCRIPT: %s\nLINE: %d\nCOMMAND: %s\nERROR: %s\n---\n",
						strings.TrimSpace(trex_utils.Timestamp()), path, idx+1, line, err.Error())
					f.WriteString(entry)
				}
			}
			// Print a concise rust-style error for script context
			printRustStyleError("SCRIPT_ERROR", "Error running script", []string{
				fmt.Sprintf("File: %s", path),
				fmt.Sprintf("Line: %d", idx+1),
				fmt.Sprintf("Command: %s", line),
				fmt.Sprintf("Error: %s", err.Error()),
			}, "Check the command and module output for errors")
			return
		}
	}
}

// handleForLoop matches and executes constructs like:
// forloop 0..5 as $x do { echo "192.168.0.$x" }
func (s *Shell) handleForLoop(line string) (bool, error) {
	re := regexp.MustCompile(`(?s)^\s*forloop\s+([^\s]+)\s+as\s+\$([A-Za-z_][A-Za-z0-9_]*)\s+do\s*\{(.*)\}\s*$`)
	m := re.FindStringSubmatch(line)
	if m == nil {
		return false, nil
	}
	rangeExpr := m[1]
	varName := m[2]
	body := m[3]

	var values []string
	if strings.Contains(rangeExpr, "..") {
		parts := strings.SplitN(rangeExpr, "..", 2)
		start, err1 := strconv.Atoi(parts[0])
		end, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			return true, fmt.Errorf("invalid range: %s", rangeExpr)
		}
		if start <= end {
			for i := start; i <= end; i++ {
				values = append(values, strconv.Itoa(i))
			}
		} else {
			for i := start; i >= end; i-- {
				values = append(values, strconv.Itoa(i))
			}
		}
	} else if strings.Contains(rangeExpr, ",") {
		for _, part := range strings.Split(rangeExpr, ",") {
			values = append(values, strings.TrimSpace(part))
		}
	} else {
		// single number -> 0..n-1
		if n, err := strconv.Atoi(rangeExpr); err == nil {
			for i := 0; i < n; i++ {
				values = append(values, strconv.Itoa(i))
			}
		} else {
			return true, fmt.Errorf("invalid range expression: %s", rangeExpr)
		}
	}

	// split body into commands by ';' or newlines
	var cmds []string
	for _, part := range strings.Split(body, ";") {
		part = strings.TrimSpace(part)
		if part != "" {
			cmds = append(cmds, part)
		}
	}

	for _, val := range values {
		s.vars[varName] = val
		for _, cmd := range cmds {
			expanded := s.expandVars(cmd)
			if err := s.executeCommand(expanded); err != nil {
				return true, err
			}
		}
	}
	// remove loop variable
	delete(s.vars, varName)
	return true, nil
}

// printModuleNotFound prints a pretty error for missing module
func (s *Shell) printModuleNotFound(cmd string) {
	fmt.Println()
	printRustStyleError("MODULE_ERROR", "Module not found", []string{
		fmt.Sprintf("Module '%s' not found in:", cmd),
		fmt.Sprintf("  %s", s.moduleDir),
	}, "Make sure the module file exists in your modules directory")
	fmt.Println()
}

// printExecutionError prints a pretty error for module execution failure
func (s *Shell) printExecutionError(cmd string, modulePath string, err error) {
	fmt.Println()
	ctx := []string{
		fmt.Sprintf("Module: %s", cmd),
		fmt.Sprintf("File: %s", modulePath),
		fmt.Sprintf("Error: %s", err.Error()),
	}
	printRustStyleError("EXECUTION_ERROR", "Failed to execute module", ctx, "Ensure the module returns valid JSON output")

	// Also write error details to a log file under ~/.t-rex/error.log
	if home, herr := os.UserHomeDir(); herr == nil {
		trexDir := filepath.Join(home, ".t-rex")
		os.MkdirAll(trexDir, 0755)
		logPath := filepath.Join(trexDir, "error.log")
		f, ferr := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if ferr == nil {
			defer f.Close()
			// Write a structured entry
			entry := fmt.Sprintf("TIME: %s\nCOMMAND: %s\nFILE: %s\nERROR: %s\nHINT: %s\n---\n",
				strings.TrimSpace(trex_utils.Timestamp()), cmd, modulePath, err.Error(), "Ensure the module returns valid JSON output")
			f.WriteString(entry)
		}
	}
	fmt.Println()
}

// printRustStyleError prints an error in Rust-style format
func printRustStyleError(errorType string, title string, context []string, hint string) {
	// Error header
	fmt.Printf("%s %s\n", trex_utils.ColoredText("×", trex_utils.Red), trex_utils.ColoredText(errorType, trex_utils.Red))
	fmt.Println()

	// Error message
	fmt.Printf("  %s\n", title)
	fmt.Println()

	// Context
	for _, line := range context {
		fmt.Printf("  %s\n", line)
	}
	fmt.Println()

	// Hint
	if hint != "" {
		fmt.Printf("  %s %s\n", trex_utils.ColoredText("help:", trex_utils.Cyan), hint)
	}
}

// printResult prints command result
func (s *Shell) printResult(result map[string]interface{}) {
	if result == nil {
		return
	}

	prettyPrint := false
	tablePrint := false

	if pp, exists := result["__pretty_print"]; exists {
		prettyPrint = pp.(bool)
	}
	if tt, exists := result["__table_print"]; exists {
		tablePrint = tt.(bool)
	}

	fmt.Println()

	if tablePrint {
		if output, exists := result["output"]; exists {
			fmt.Print(trex_utils.TablePrint(output))
		}
	} else if prettyPrint {
		if output, exists := result["output"]; exists {
			fmt.Print(trex_utils.PrettyPrint(output))
		}
	} else {
		// Print as formatted JSON
		if data, err := json.MarshalIndent(result, "", "  "); err == nil {
			fmt.Println(string(data))
		}
	}

	fmt.Println()
}

// loadConfig loads configuration from .trexrc
func loadConfig(s *Shell) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	trexDir := filepath.Join(homeDir, ".t-rex")
	os.MkdirAll(trexDir, 0755)

	configPath := filepath.Join(trexDir, ".trexrc")
	data, err := os.ReadFile(configPath)
	if err != nil {
		// Create default config if doesn't exist
		createDefaultConfig(configPath)
		return
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])

			if key == "use_colors" && val == "false" {
				s.useColors = false
			}
			if key == "prompt_symbol" {
				s.promptTemplate = val
			}
			if key == "prompt_template" {
				s.promptTemplate = val
			}
		}
	}
}

// createDefaultConfig creates a default .trexrc file
func createDefaultConfig(configPath string) {
	config := `# T-Rex Shell Configuration
module_paths=~/.t-rex/modules
use_colors=true
theme=default
history_enabled=true
history_size=1000

# Prompt customization - use format: prompt_template=%u@%h:%D❯
# %u = username
# %h = hostname
# %w = full working directory
# %d = full working directory (same as %w)
# %D = working directory basename only
# %~ = home directory relative path
prompt_symbol=❯
prompt_template=❯
prompt_color=cyan
python_executable=python3
`
	os.WriteFile(configPath, []byte(config), 0644)
}
