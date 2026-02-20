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

	"trex_errors"
	"trex_modules"
	"trex_utils"
)

const Version = "1.0.0"

func main() {
	// Define command-line flags
	pathFlag := flag.String("path", "", "Path to custom modules directory")
	bannerFlag := flag.Bool("banner", false, "Show banner and exit (default: false)")
	execFlag := flag.String("exec", "", "Execute a command and exit")
	versionFlag := flag.Bool("version", false, "Show version information (default: false)")

	verbose_flag := flag.Bool("vv", false, "Verbse to show logs and descripeted error messages (default: false)")

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

	vv := false

	// Handle banner flag
	if *verbose_flag {
		vv = true
	}

	shell := NewShell()

	// If a file path was passed as positional arg, execute file and exit
	if len(args) > 0 {
		candidate := args[0]
		if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
			shell.ExecuteFile(candidate, vv)
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
		shell.ExecuteOnce(*execFlag, vv)
		os.Exit(0)
	}

	// Otherwise run interactive shell
	shell.Run(vv)
}

// showVersion displays version information
func showVersion() {
	trex_utils.PrintBanner()
	os.Exit(0)
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
func (s *Shell) ExecuteOnce(line string, verbose bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}

	if err := s.executeCommand(line, verbose); err != nil {
		// print brief error (detailed logging handled elsewhere)
		trex_utils.PrintError(err.Error())
	}
}

// Run starts the interactive shell
func (s *Shell) Run(verbose bool) {
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
		if err := s.executeCommand(line, verbose); err != nil {
			// error already logged/printed by lower-level handlers
			trex_utils.PrintError(err.Error())
		}
	}
}

// executeCommand processes a command
func (s *Shell) executeCommand(line string, verbose bool) error {
	// If the line is a file path, execute file
	if fi, err := os.Stat(line); err == nil && !fi.IsDir() {
		s.ExecuteFile(line, verbose)
		return nil
	}

	// variable assignment: support multiple syntaxes
	//  - set NAME VALUE
	//  - let NAME = VALUE
	//  - NAME=VALUE or export NAME=VALUE
	parts := trex_utils.ParseCommand(line)
	if len(parts) > 0 {
		// set/let handlers
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
				fmt.Printf("%s=%s\n", name, val)
				return nil
			}
		}

		// assignment shorthand: NAME=VALUE or export NAME=VALUE
		assignRe := regexp.MustCompile(`^\s*(?:export\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.+)$`)
		if assignRe.MatchString(line) {
			m := assignRe.FindStringSubmatch(line)
			name := m[1]
			val := strings.TrimSpace(m[2])
			s.vars[name] = val
			fmt.Printf("%s=%s\n", name, val)
			return nil
		}

		// shorthand: $name value  -> set variable
		if strings.HasPrefix(parts[0], "$") && len(parts) >= 2 {
			name := parts[0][1:]
			val := strings.Join(parts[1:], " ")
			s.vars[name] = val
			fmt.Printf("%s=%s\n", name, val)
			return nil
		}
	}

	// Check for forloop pattern: forloop RANGE as $var do { ... }
	handled, err := s.handleForLoop(line, verbose)
	if err != nil {
		return err
	}
	if handled {
		return nil
	}

	// Check foreach
	fhandled, ferr := s.handleForeach(line, verbose)
	if ferr != nil {
		return ferr
	}
	if fhandled {
		return nil
	}

	// Check for pipes
	if trex_utils.HasPipe(line) {
		return s.executePipeline(line)
	}

	if len(parts) == 0 {
		return nil
	}

	// expand variables in command name as well
	cmd := s.expandVars(parts[0])
	args := parts[1:]

	// Expand variables in args
	for i, a := range args {
		args[i] = s.expandVars(a)
	}

	// Try to execute as Python module
	var result map[string]interface{}
	result, err = s.executeModule(strings.Split(cmd, " "), args)
	if err != nil {
		return err
	}
	if result != nil {
		s.printResult(result)
	}

	return nil
}

// executePipeline executes a command pipeline that may start with a literal value,
// array literal, or regular command, and supports piping through modules or special operators.
func (s *Shell) executePipeline(line string) error {
	// Split into first part and the rest after first |
	parts := strings.SplitN(line, "|", 2)
	firstPart := strings.TrimSpace(parts[0])
	if firstPart == "" {
		return nil
	}

	var result map[string]interface{}

	// ────────────────────────────────────────────────
	// Case 1: Array literal [ ... ]
	// ────────────────────────────────────────────────
	firstTrim := strings.TrimSpace(firstPart)
	if strings.HasPrefix(firstTrim, "[") && strings.HasSuffix(firstTrim, "]") {
		inner := strings.TrimSpace(firstTrim[1 : len(firstTrim)-1])
		items := trex_utils.ParseCommand(inner) // respects quotes, your existing parser
		result = map[string]interface{}{
			"output": items,
			"status": "success",
		}
	} else
	// ────────────────────────────────────────────────
	// Case 2: Quoted string literal "..." or '...'
	// ────────────────────────────────────────────────
	if (strings.HasPrefix(firstTrim, `"`) && strings.HasSuffix(firstTrim, `"`)) ||
		(strings.HasPrefix(firstTrim, `'`) && strings.HasSuffix(firstTrim, `'`)) {
		strVal := firstTrim[1 : len(firstTrim)-1]
		// Note: this is basic quote removal — no full unescaping yet
		result = map[string]interface{}{
			"output": strVal,
			"status": "success",
		}
	} else
	// ────────────────────────────────────────────────
	// Case 3: Number literal (int or float)
	// ────────────────────────────────────────────────
	if num, err := strconv.ParseFloat(firstTrim, 64); err == nil {
		// Prefer float64 for generality (most modules expect strings anyway)
		result = map[string]interface{}{
			"output": num,
			"status": "success",
		}
	} else if intVal, err := strconv.ParseInt(firstTrim, 10, 64); err == nil {
		result = map[string]interface{}{
			"output": intVal,
			"status": "success",
		}
	} else {
		// ────────────────────────────────────────────────
		// Case 4: Regular command
		// ────────────────────────────────────────────────
		cmdParts := trex_utils.ParseCommand(firstPart)
		if len(cmdParts) == 0 {
			return nil
		}

		cmd := cmdParts[0]
		args := cmdParts[1:]

		// Variable expansion
		for i := range args {
			args[i] = s.expandVars(args[i])
		}
		cmd = s.expandVars(cmd)

		var err error
		result, err = s.executeModule(strings.Split(cmd, " "), args)
		if err != nil {
			return err
		}
		if result == nil {
			return nil
		}
	}

	// No more pipes → just print and return
	if len(parts) < 2 {
		s.printResult(result)
		return nil
	}

	// ────────────────────────────────────────────────
	// Process piped stages
	// ────────────────────────────────────────────────
	pipeRest := strings.TrimSpace(parts[1])
	pipeParts := strings.Split(pipeRest, "|")

	for _, pipe := range pipeParts {
		pipe = strings.TrimSpace(pipe)
		if pipe == "" {
			continue
		}

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
			// Assume it's a module name
			modulePath, err := s.loader.FindModule(op)
			if err != nil || modulePath == "" {
				return fmt.Errorf("unknown pipeline operator or module: %s", op)
			}

			// Prepare arguments: user-supplied args + previous output
			callArgs := make([]string, 0, len(args)+1)
			callArgs = append(callArgs, args...)

			// Append previous output intelligently
			if out, ok := result["output"]; ok && out != nil {
				switch v := out.(type) {
				case string:
					callArgs = append(callArgs, v)

				case float64, int, int64, bool:
					callArgs = append(callArgs, fmt.Sprintf("%v", v))

				case []interface{}:
					for _, item := range v {
						callArgs = append(callArgs, fmt.Sprintf("%v", item))
					}

				case []string:
					callArgs = append(callArgs, v...)

				default:
					// JSON fallback for maps, structs, etc.
					if b, merr := json.Marshal(v); merr == nil {
						callArgs = append(callArgs, string(b))
					} else {
						callArgs = append(callArgs, fmt.Sprintf("%v", v))
					}
				}
			}

			// Execute next module
			modResult, err := s.executeModule(trex_utils.ParseCommand(pipe), callArgs)
			if err != nil {
				return err
			}
			if modResult == nil {
				return nil // module decided to stop / silent fail
			}

			// Carry forward the new result
			result = modResult
		}
	}

	// Final output
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
func (s *Shell) executeModule(cmdA []string, args []string) (map[string]interface{}, error) {
	cmd := cmdA[0]

	modulePath, err := s.loader.FindModule(cmd)
	if err != nil {
		s.printModuleNotFound(cmdA)
		return nil, os.ErrNotExist
	}

	result, err := s.executor.Execute(cmd, args)
	if err != nil {
		s.printExecutionError(cmdA, modulePath, err)
		return nil, err
	}

	return result, nil
}

// ExecuteFile executes commands from a script file (one command per line)
func (s *Shell) ExecuteFile(path string, verbose bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		trex_utils.PrintError("Failed to read script: " + err.Error())
		return
	}

	lines := strings.Split(string(data), "\n")

	if verbose == true {
		fmt.Printf("Running script: %s\n", path)
	}

	for idx, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Prefix with line number for easier debugging
		if verbose == true {
			fmt.Printf(" %d $ %s\n", idx+1, line)
		}
		s.history.Add(line)
		if err := s.executeCommand(line, verbose); err != nil {
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
			// Print a rich rust-style error with file/line/context
			e := trex_errors.NewError(trex_errors.ErrorType("SCRIPT_ERROR"), "Error running script").WithLocation(path, idx+1).WithContext(line).WithHint("Check the command and module output for errors")
			fmt.Print(e.Format())
			return
		}
	}
}

// handleForLoop matches and executes constructs like:
// forloop 0..5 as $x do { echo "192.168.0.$x" }
func (s *Shell) handleForLoop(line string, verbose bool) (bool, error) {
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
			if err := s.executeCommand(expanded, verbose); err != nil {
				return true, err
			}
		}
	}
	// remove loop variable
	delete(s.vars, varName)
	return true, nil
}

// handleForeach handles constructs like:
// foreach "sha256"|"sha512" as $x do { echo $x }
func (s *Shell) handleForeach(line string, verbose bool) (bool, error) {
	re := regexp.MustCompile(`(?s)^\s*foreach\s+(.+?)\s+as\s+\$([A-Za-z_][A-Za-z0-9_]*)\s+do\s*\{(.*)\}\s*$`)
	m := re.FindStringSubmatch(line)
	if m == nil {
		return false, nil
	}

	listExpr := strings.TrimSpace(m[1])
	varName := m[2]
	body := m[3]

	var items []string
	// array literal
	if strings.HasPrefix(listExpr, "[") && strings.HasSuffix(listExpr, "]") {
		inner := strings.TrimSpace(listExpr[1 : len(listExpr)-1])
		items = trex_utils.ParseCommand(inner)
	} else if strings.Contains(listExpr, "|") {
		parts := strings.Split(listExpr, "|")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			// strip quotes
			if (strings.HasPrefix(p, "\"") && strings.HasSuffix(p, "\"")) || (strings.HasPrefix(p, "'") && strings.HasSuffix(p, "'")) {
				p = p[1 : len(p)-1]
			}
			items = append(items, p)
		}
	} else {
		// single token
		items = append(items, strings.Trim(listExpr, " \t\n\r\"'"))
	}

	// split body into commands by ';'
	var cmds []string
	for _, part := range strings.Split(body, ";") {
		part = strings.TrimSpace(part)
		if part != "" {
			cmds = append(cmds, part)
		}
	}

	for _, it := range items {
		s.vars[varName] = it
		for _, cmd := range cmds {
			expanded := s.expandVars(cmd)
			if err := s.executeCommand(expanded, verbose); err != nil {
				return true, err
			}
		}
	}
	delete(s.vars, varName)
	return true, nil
}

// printRustStyleError prints a diagnostic message in a rustc-like style
// using only standard library + ANSI escape codes (no external dependencies)

// ANSI color codes
const (
	reset  = "\x1b[0m"
	bold   = "\x1b[1m"
	red    = "\x1b[31m"
	yellow = "\x1b[33m"
	cyan   = "\x1b[36m"
	green  = "\x1b[32m"
	gray   = "\x1b[90m"
)

// printRustStyleError prints a diagnostic message in a rustc-like style
func printRustStyleError(
	level string, // "ERROR", "WARNING", "NOTE"
	title string, // e.g. "module not found"
	location string, // "file.trex:9:5" or "<interactive>" or ""
	codeContext string, // the offending source line (or "")
	underlineStart int, // 0-based column
	underlineLen int, // how many characters to underline
	message string, // main error message
	hint string, // optional hint
	notes ...string, // additional notes
) {
	var levelColor string
	switch strings.ToUpper(level) {
	case "ERROR":
		levelColor = red
	case "WARNING":
		levelColor = yellow
	case "NOTE":
		levelColor = cyan
	default:
		levelColor = red
	}

	// ────────────────────────────────────────────────
	// Header
	// ────────────────────────────────────────────────

	// Location (with nicer spacing)
	header := fmt.Sprintf("<%s%s%s%s> %s",
		bold, levelColor, level, reset, bold+title+reset)

	if location != "" {
		fmt.Fprintf(os.Stderr, " %s-->%s %s %s\n", cyan, reset, location, header)
	} else {
		location = "entry:repl"
		fmt.Fprintf(os.Stderr, "%s--->%s %s %s\n", cyan, reset, location, header)
	}

	// Separator line
	//fmt.Fprintf(os.Stderr, " %s│%s\n", cyan, reset)

	// Code context + underline
	if codeContext != "" {
		// Show the source line
		fmt.Fprintf(os.Stderr, " %s│%s\n", cyan, reset)
		if location == "entry:repl" {
			fmt.Fprintf(os.Stderr, "%s0│%s %s\n", cyan, reset, codeContext)
		} else {
			fmt.Fprintf(os.Stderr, " %s│%s %s\n", cyan, reset, codeContext)
		}

		// Underline (only if meaningful)
		if underlineLen > 0 && underlineStart >= 0 {
			spaces := strings.Repeat("", underlineStart)
			underline := strings.Repeat("^", underlineLen) // ^ is more common in modern rustc
			fmt.Fprintf(os.Stderr, " %s│%s %s%s%s %s\n",
				cyan, reset,
				spaces,
				red+bold+underline+reset,
				" "+message, reset,
			)
		} else {
			// No underline → message right below line
			fmt.Fprintf(os.Stderr, " %s│%s  %s\n", cyan, reset, message)
		}
	} else {
		// No code → just message after separator
		fmt.Fprintf(os.Stderr, " %s│%s\n", cyan, reset)
		fmt.Fprintf(os.Stderr, " %s│%s %s\n", cyan, reset, message)
	}

	// Hint (if any)
	if hint != "" {
		fmt.Fprintf(os.Stderr, " %s│%s\n", cyan, reset)
		fmt.Fprintf(os.Stderr, " %s│%s %shint:%s %s\n", cyan, reset, bold, reset, hint)
	}

	// Notes
	for _, note := range notes {
		fmt.Fprintf(os.Stderr, " %s│%s %snote:%s %s\n", cyan, reset, bold, reset, note)
	}

	fmt.Fprintln(os.Stderr)
}

// printModuleNotFound – wrapper for module-not-found case
func (s *Shell) printModuleNotFound(cmd []string) {
	// You can improve this later by reading context from s.currentScript etc.
	// For now — keeping it simple as per original signature limitation

	printRustStyleError(
		"err_module_not_found",
		"module not found",
		"entry#repl",           // location
		strings.Join(cmd, " "), // code context
		0, len(cmd[0]),
		fmt.Sprintf("cannot find module %s'%s' %s", bold, cmd[0], reset),
		fmt.Sprintf("expected to find %s.py / %s.json / %s.yaml (or similar) in the modules directory", cmd[0], cmd[0], cmd[0]),
		fmt.Sprintf("current search path: %s", s.moduleDir),
		fmt.Sprintf("run %sls -la %s%s to see available modules", bold, s.moduleDir, reset),
	)
	//println(len(strings.Split(cmd, " ")[0]))
}

// printExecutionError – wrapper for module runtime / output errors
func (s *Shell) printExecutionError(cmd []string, modulePath string, err error) {
	// Try to make module path relative (from ~/.t-rex/modules)
	relPath := modulePath
	if home, _ := os.UserHomeDir(); home != "" {
		base := filepath.Join(home, ".t-rex", "modules")
		if r, err := filepath.Rel(base, modulePath); err == nil && !strings.HasPrefix(r, "..") {
			relPath = r
		}
	}

	printRustStyleError(
		"ERROR",
		"module execution failed",
		relPath, // using module file as "location" for now
		"",      // no source line context (would need shell state)
		0, 0,
		fmt.Sprintf("%s: %v", cmd[0], err),
		"module must print **valid JSON** to stdout and nothing else",
		fmt.Sprintf("no stray prints, debug output, tracebacks, or syntax errors allowed"),
		fmt.Sprintf("full path: %s", modulePath),
		"check Python syntax, imports, and use json.dumps(...) correctly",
	)

	// ─── Append structured log entry ─────────────────────────────────────
	home, _ := os.UserHomeDir()
	if home == "" {
		return
	}

	trexDir := filepath.Join(home, ".t-rex")
	_ = os.MkdirAll(trexDir, 0755)

	logPath := filepath.Join(trexDir, "error.log")
	f, ferr := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if ferr != nil {
		return
	}
	defer f.Close()

	timestamp := strings.TrimSpace(trex_utils.Timestamp()) // assuming this helper exists
	entry := fmt.Sprintf(
		"[%s] EXECUTION_ERROR\n"+
			"  Command : %s\n"+
			"  Module  : %s\n"+
			"  Error   : %v\n"+
			"  Hint    : Ensure module outputs only valid JSON\n"+
			"----------------------------------------\n",
		timestamp, cmd, modulePath, err,
	)
	_, _ = f.WriteString(entry)
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
		data, _ = os.ReadFile(configPath)

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
		} else {
			s.executeCommand(line, false)
		}
		//fmt.Println(line)
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
