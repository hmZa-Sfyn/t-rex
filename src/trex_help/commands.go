package trex_help

// Commands contains all built-in command help
var Commands = map[string]string{
	"help": `help [command]
  Display help information. Use 'help <command>' for specific command help.`,

	"exit": `exit
  Exit the T-Rex shell.`,

	"clear": `clear
  Clear the terminal screen.`,

	"history": `history [n]
  Show command history. Use 'history 10' to show last 10 commands.`,

	"load": `load -path <path>
  Load Python modules from specified directory.`,

	"modules": `modules
  List all loaded modules.`,

	"select": `select <field1> [field2] ...
  Filter JSON output to specified fields (pipe operation).
  Example: ls | select name,owner`,

	"pp": `pp
  Pretty print JSON output without standard formatting (pipe operation).
  Example: ls | pp`,

	"tt": `tt
  Tabular print JSON output without standard formatting (pipe operation).
  Example: ls | tt`,
}

// GetHelp returns help text for a command
func GetHelp(cmd string) string {
	if help, exists := Commands[cmd]; exists {
		return help
	}
	return "No help available for: " + cmd
}

// GetAllHelp returns help for all commands
func GetAllHelp() string {
	help := "T-REX SHELL - Built-in Commands\n\n"
	for cmd := range Commands {
		help += "  " + cmd + "\n"
	}
	help += "\nUse 'help <command>' for detailed information.\n"
	return help
}
