package trex_utils

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// Color codes for terminal output
type Color string

const (
	Reset   Color = "\033[0m"
	Red     Color = "\033[31m"
	Green   Color = "\033[32m"
	Yellow  Color = "\033[33m"
	Blue    Color = "\033[34m"
	Magenta Color = "\033[35m"
	Cyan    Color = "\033[36m"
	White   Color = "\033[37m"
	Bold    Color = "\033[1m"
	Dim     Color = "\033[2m"
)

// ColoredText returns text with ANSI color codes
func ColoredText(text string, color Color) string {
	return string(color) + text + string(Reset)
}

// Banner prints the T-Rex banner
func PrintBanner() {
	banner := `
  dP                                       
  88                                       
d8888P          88d888b. .d8888b. dP.  .dP 
  88   88888888 88'  '88 88ooood8  '8bd8'  
  88            88       88.  ...  .d88b.  
  dP            dP       '88888P' dP'  'dP

⠀⠀⠀⠀⠀⣾⣽⣿⣿⡇⠀⠀ (repo: https://github.com/hmZa-Sfyn/t-rex)
⢀⠀⠀⠀⢀⣿⣿⠯⠍⠁⠀⠀⠀(author: https://github.com/hmZa-Sfyn)⠀
⢸⣦⣤⣾⣿⣿⣿⠓⠀⠀⠀⠀ (version: beta-whatever)
⠀⠙⢿⣿⣿⡿⠃⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠸⠏⠈⠇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
`
	fmt.Println(ColoredText(banner, Cyan))
}

// PrintPrompt prints the shell prompt
func PrintPrompt(symbol string, color Color) string {
	return string(color) + symbol + string(Reset) + " "
}

// BuildPrompt builds a dynamic prompt with user info
func BuildPrompt(symbol string, color Color, showUser, showDir, showEnv bool) string {
	var prompt strings.Builder

	prompt.WriteString(string(color))

	if showUser {
		currentUser, err := user.Current()
		if err == nil {
			prompt.WriteString(currentUser.Username)
			prompt.WriteString("@")
		}
	}

	if showDir {
		cwd, err := os.Getwd()
		if err == nil {
			home, _ := os.UserHomeDir()
			if strings.HasPrefix(cwd, home) {
				cwd = strings.Replace(cwd, home, "~", 1)
			}
			prompt.WriteString(cwd)
		}
	}

	if showUser || showDir {
		prompt.WriteString(" ")
	}

	prompt.WriteString(symbol)
	prompt.WriteString(string(Reset))
	prompt.WriteString(" ")

	return prompt.String()
}

// PrintError prints an error in red
func PrintError(msg string) {
	fmt.Println(ColoredText("× "+msg, Red))
}

// PrintSuccess prints success message in green
func PrintSuccess(msg string) {
	fmt.Println(ColoredText("✓ "+msg, Green))
}

// PrintInfo prints info message in blue
func PrintInfo(msg string) {
	fmt.Println(ColoredText("ℹ "+msg, Blue))
}

// PrintWarning prints warning in yellow
func PrintWarning(msg string) {
	fmt.Println(ColoredText("⚠ "+msg, Yellow))
}

// PrintExit prints exit message
func PrintExit(msg string) {
	fmt.Println(ColoredText("👋 "+msg, Cyan))
}

// ExpandPrompt expands prompt variables like %u for username, %w for working dir
func ExpandPrompt(template string) string {
	result := template

	// %u - username
	if strings.Contains(result, "%u") {
		if currentUser, err := user.Current(); err == nil {
			result = strings.ReplaceAll(result, "%u", currentUser.Username)
		}
	}

	// %h - hostname
	if strings.Contains(result, "%h") {
		if hostname, err := os.Hostname(); err == nil {
			result = strings.ReplaceAll(result, "%h", hostname)
		}
	}

	// %w or %d - working directory (full path)
	if strings.Contains(result, "%w") || strings.Contains(result, "%d") {
		if wd, err := os.Getwd(); err == nil {
			result = strings.ReplaceAll(result, "%w", wd)
			result = strings.ReplaceAll(result, "%d", wd)
		}
	}

	// %D - working directory (basename only)
	if strings.Contains(result, "%D") {
		if wd, err := os.Getwd(); err == nil {
			dir := filepath.Base(wd)
			if dir == "/" {
				dir = "/"
			}
			result = strings.ReplaceAll(result, "%D", dir)
		}
	}

	// %~ - home directory relative path
	if strings.Contains(result, "%~") {
		if wd, err := os.Getwd(); err == nil {
			if homeDir, err := os.UserHomeDir(); err == nil {
				if strings.HasPrefix(wd, homeDir) {
					rel := strings.TrimPrefix(wd, homeDir)
					if rel == "" {
						result = strings.ReplaceAll(result, "%~", "~")
					} else {
						result = strings.ReplaceAll(result, "%~", "~"+rel)
					}
				} else {
					result = strings.ReplaceAll(result, "%~", wd)
				}
			}
		}
	}

	// %t - time (HH:MM)
	if strings.Contains(result, "%t") {
		// Skip time for now - would require time parsing
	}

	return result
}
