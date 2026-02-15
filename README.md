# T-Rex Shell 🦖

A modern, user-friendly shell interface written in Go that executes Python3-based commands and returns JSON-formatted output with support for piping, filtering, and pretty printing.

## Features

- 🐍 **Python3 Module System** - Load and execute Python3 commands from configurable directories
- 📦 **JSON Output** - All commands return structured JSON output for easy parsing
- 🔗 **Piping Support** - Chain commands with `|` operator
- 🎨 **Colored Output** - Beautiful ANSI color-coded terminal output
- 📜 **Command History** - Automatic history saving across sessions
- ⚡ **Rust-style Errors** - Clear, helpful error messages with context
- 🔧 **Easy Customization** - Configuration files and modular architecture
- 💻 **Cross-Platform** - Works on Linux, macOS, and Windows

## Quick Start

### Installation

**Linux/macOS:**
```bash
bash setup.sh
~/.t-rex/bin/t-rex
```

**Windows:**
```powershell
powershell -ExecutionPolicy Bypass -File setup.ps1
~\.t-rex-windows\bin\t-rex.exe
```

### First Commands

```bash
❯ echo "hello world"
{
  "output": "hello world",
  "status": "success"
}

❯ sysinfo | pp
hostname: 0root
platform: Linux
architecture: x86_64
python_version: 3.13.11
```

## Usage

### Running Modules

All modules are Python3 scripts that return JSON output:

```bash
❯ echo "Hello T-Rex"
{"output":"Hello T-Rex","status":"success"}

❯ random 1 100
{"output":42,"range":{"min":1,"max":100},"status":"success"}
```

### Piping Commands

Use the pipe operator `|` to chain operations:

```bash
❯ sysinfo | select hostname,platform
{"output":{"hostname":"mypc","platform":"Linux"},"status":"success"}

❯ sysinfo | pp
hostname: mypc
platform: Linux
python_version: 3.9.0
```

### Available Operations

- `select <field1> [field2]` - Filter JSON fields (pipe operation)
- `pp` - Pretty print output without JSON formatting (pipe operation)
- `exit` or `quit` - Exit the shell

## Project Structure

```
t-rex/
├── main.go                 # Main shell REPL
├── setup.sh               # Linux/macOS setup script
├── setup.ps1              # Windows setup script
├── README.md              # This file
├── QUICKSTART.md          # Quick start guide
├── learn.md               # Guide for creating custom modules
├── CONFIG.md              # Configuration reference
├── .trexrc               # Configuration file
├── .env.example          # Environment variables example
├── src/
│   ├── trex_errors/      # Error handling system
│   ├── trex_modules/     # Module loader
│   ├── trex_utils/       # Color, history, piping utilities
│   └── trex_help/        # Help system
└── modules/              # Example Python3 modules
    ├── echo.py
    ├── random.py
    └── sysinfo.py
```

## Configuration

Edit `~/.t-rex/.trexrc` to customize T-Rex:

```bash
# Module paths - colon-separated
module_paths=~/.t-rex/modules

# Color support
use_colors=true
theme=default

# History settings
history_enabled=true
history_size=1000

# Prompt customization
prompt_symbol=❯
prompt_color=cyan

# Python3 interpreter
python_executable=python3
```

## Home Directory Structure

T-Rex creates the following structure on first run:

**Linux/macOS:**
```
~/.t-rex/
├── bin/t-rex            # Shell executable
├── modules/             # User modules
├── .trexrc             # Configuration
└── history             # Command history
```

**Windows:**
```
~\.t-rex-windows\
├── bin\t-rex.exe       # Shell executable
├── modules\            # User modules
├── .trexrc            # Configuration
└── history            # Command history
```

## Creating Custom Modules

See [learn.md](learn.md) for detailed instructions on creating custom Python3 modules.

### Quick Example

Create `~/.t-rex/modules/greet.py`:

```python
#!/usr/bin/env python3
"""greet - Say hello"""
import json
import sys

name = sys.argv[1] if len(sys.argv) > 1 else "World"
result = {
    "output": f"Hello, {name}!",
    "status": "success"
}
print(json.dumps(result))
```

Use it:
```bash
❯ greet Alice
{"output":"Hello, Alice!","status":"success"}
```

## Error Handling

Errors are displayed with helpful information:

```
  ✗ Module Error
    Module 'mymodule' not found
    Hint: Make sure 'mymodule' exists in your modules directory
    Location: /home/user/.t-rex/modules
```

## Features Explained

### JSON Output
All modules return structured JSON for easy parsing and piping:
```json
{
  "output": "result here",
  "status": "success"
}
```

### Pretty Printing
Use `pp` to display output in readable format:
```bash
❯ sysinfo | pp
```

### Field Selection
Use `select` to filter specific fields:
```bash
❯ sysinfo | select hostname,platform
```

### Command History
Automatically saved to `~/.t-rex/history` for easy access.

## Tips & Tricks

1. **Add to PATH**: `export PATH="$HOME/.t-rex/bin:$PATH"`
2. **Create Aliases**: `alias trex='~/.t-rex/bin/t-rex'`
3. **Test Modules**: Run Python modules directly: `python3 module.py`
4. **Backup Config**: Keep a copy of `.trexrc` before major changes
5. **Check Errors**: Look at module's stderr with proper error handling

## Development

### Building from Source

```bash
cd /path/to/t-rex
export GO111MODULE=off
export GOPATH=$(pwd)
go build -o t-rex main.go
```

### Code Organization

- **main.go** - Shell REPL and command execution (< 200 lines)
- **src/trex_errors/** - Error handling with Rust-style formatting
- **src/trex_modules/** - Module discovery and loading
- **src/trex_utils/** - Colors, history, parsing, piping
- **src/trex_help/** - Built-in help system
- **modules/** - Example Python3 modules

### Design Principles

1. **Modular** - Each component is independent
2. **Readable** - Files kept under 200 lines of code
3. **Extensible** - Easy to add new modules and features
4. **Cross-platform** - Works on Linux, macOS, and Windows
5. **User-friendly** - Clear errors and good documentation

## Documentation

- [QUICKSTART.md](QUICKSTART.md) - Quick start guide
- [learn.md](learn.md) - Creating custom modules
- [CONFIG.md](CONFIG.md) - Configuration reference
- [README.md](README.md) - This file

## Troubleshooting

### Module Not Found
- Check module exists in `~/.t-rex/modules/`
- Ensure module name matches filename (without .py)
- Run `python3 module.py` directly to test

### Execution Error
- Verify Python3 is installed: `python3 --version`
- Check module returns valid JSON
- Look for syntax errors in module code

### Color Issues
- Set `use_colors=false` in `.trexrc`
- Check terminal supports ANSI colors

## System Requirements

- **Go** 1.19+ (for building from source)
- **Python3** 3.6+ (for running modules)
- **Linux/macOS/Windows** operating system

## Contributing

To contribute:
1. Create a new Python module in `modules/`
2. Ensure it returns valid JSON output
3. Test with the shell
4. Include documentation

## License

Created with ❤️ for the command-line enthusiast community

## Support

For issues, questions, or contributions, refer to the included documentation files.

---

**Happy shell scripting! 🦖**

