package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
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

	s.executeCommand(line)
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
		s.executeCommand(line)
	}
}

// executeCommand processes a command
func (s *Shell) executeCommand(line string) {
	// Check for pipes
	if trex_utils.HasPipe(line) {
		s.executePipeline(line)
		return
	}

	parts := trex_utils.ParseCommand(line)
	if len(parts) == 0 {
		return
	}

	cmd := parts[0]
	args := parts[1:]

	// Try to execute as Python module
	result := s.executeModule(cmd, args)
	if result != nil {
		s.printResult(result)
	}
}

// executePipeline handles piped commands
func (s *Shell) executePipeline(line string) {
	parts := strings.SplitN(line, "|", 2)
	firstCmd := strings.TrimSpace(parts[0])

	// Execute first command
	cmdParts := trex_utils.ParseCommand(firstCmd)
	if len(cmdParts) == 0 {
		return
	}

	cmd := cmdParts[0]
	args := cmdParts[1:]

	result := s.executeModule(cmd, args)
	if result == nil {
		return
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
				modResult := s.executeModule(op, callArgs)
				if modResult == nil {
					// if module failed, stop pipeline
					return
				}
				// Replace current result with module result for next steps
				result = modResult
				continue
			}
		}
	}

	s.printResult(result)
}

// executeModule executes a Python module
func (s *Shell) executeModule(cmd string, args []string) map[string]interface{} {
	modulePath, err := s.loader.FindModule(cmd)
	if err != nil {
		s.printModuleNotFound(cmd)
		return nil
	}

	result, err := s.executor.Execute(cmd, args)
	if err != nil {
		s.printExecutionError(cmd, modulePath, err)
		return nil
	}

	return result
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
		s.executeCommand(line)
	}
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
