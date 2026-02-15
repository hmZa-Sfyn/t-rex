# T-Rex Shell - Quick Start Guide

Welcome to T-Rex Shell! 🦖 A modern, JSON-based command execution shell built in Go.

## Installation

### Linux/macOS

```bash
cd /home/hmza/workspaces/hamza/t-rex
bash setup.sh
```

The setup script will:
- Create `~/.t-rex/` directory structure
- Create `.trexrc` configuration file
- Copy example Python modules
- Build the T-Rex binary
- Install modules to `~/.t-rex/modules/`

### Windows

```powershell
cd path\to\t-rex
powershell -ExecutionPolicy Bypass -File setup.ps1
```

The setup script will:
- Create `~/.t-rex-windows/` directory structure
- Create `.trexrc` configuration file
- Copy example Python modules
- Build the T-Rex binary
- Optionally add T-Rex to PATH

## Running T-Rex

After setup:

```bash
# Linux/macOS
~/.t-rex/bin/t-rex

# Or add to PATH
export PATH="$HOME/.t-rex/bin:$PATH"
t-rex
```

```powershell
# Windows
~/.t-rex-windows/bin/t-rex.exe
```

## First Commands

```bash
❯ echo "hello world"
{
  "output": "hello world",
  "status": "success"
}

❯ sysinfo
{
  "output": {...},
  "status": "success"
}

❯ random 1 100
{
  "output": 42,
  "range": {"min": 1, "max": 100},
  "status": "success"
}
```

## Pretty Print Output

Use the `pp` pipe to display output in a readable format:

```bash
❯ sysinfo | pp
hostname: 0root
platform: Linux
architecture: x86_64
python_version: 3.13.11
```

## Filter Fields with Select

```bash
❯ sysinfo | select hostname,platform
{
  "output": {"hostname": "0root", "platform": "Linux"},
  "status": "success"
}
```

## Directory Structure

```
~/.t-rex/                    (Linux/macOS)
├── bin/
│   └── t-rex               # The shell executable
├── modules/                # Your Python3 modules
│   ├── echo.py
│   ├── random.py
│   ├── sysinfo.py
│   └── ... (your custom modules)
├── .trexrc                 # Configuration file
└── history                 # Command history (auto-created)

~/.t-rex-windows/           (Windows)
├── bin/
│   └── t-rex.exe           # The shell executable
├── modules/                # Your Python3 modules
├── .trexrc
└── history
```

## Configuration

Edit `~/.t-rex/.trexrc` to customize T-Rex:

```bash
# Module search paths (colon-separated)
module_paths=~/.t-rex/modules

# Enable colored output
use_colors=true

# Shell theme
theme=default

# History settings
history_enabled=true
history_size=1000

# Prompt customization
prompt_symbol=❯
prompt_color=cyan

# Python interpreter
python_executable=python3
```

## Creating Custom Modules

See [learn.md](learn.md) for detailed instructions on creating custom Python3 modules.

Quick example:

```python
#!/usr/bin/env python3
"""greet - Say hello to someone"""
import json
import sys

name = sys.argv[1] if len(sys.argv) > 1 else "World"
result = {
    "output": f"Hello, {name}!",
    "status": "success"
}
print(json.dumps(result))
```

Place in `~/.t-rex/modules/greet.py` and run:

```bash
❯ greet Alice
{
  "output": "Hello, Alice!",
  "status": "success"
}
```

## Piping and Composition

Chain commands together:

```bash
❯ command1 | select field1,field2 | pp
```

- `select` - Filter JSON fields
- `pp` - Pretty print output

## Built-in Features

- **Command History** - Automatically saved to `~/.t-rex/history`
- **Color Output** - Beautiful ANSI colored terminal output
- **JSON Output** - All commands return structured JSON
- **Piping** - Chain commands with `|` operator
- **Configuration** - Easy-to-edit `.trexrc` file
- **Cross-Platform** - Works on Linux, macOS, and Windows

## Troubleshooting

### Module Not Found
- Ensure Python modules are in `~/.t-rex/modules/`
- Check module name matches filename (without .py)
- Verify modules return valid JSON output

### Execution Error
- Run the Python module directly to test: `python3 module.py`
- Check Python3 is installed: `python3 --version`
- Ensure module syntax is correct

### Color Issues
- Set `use_colors=false` in `.trexrc` if colors don't display
- Some terminals may need explicit color support enabled

## Development

### Building from Source

```bash
cd /path/to/t-rex
export GO111MODULE=off
export GOPATH=$(pwd)
go build -o t-rex main.go
```

### Project Structure

```
t-rex/
├── main.go                      # Main shell REPL
├── setup.sh                     # Linux/macOS setup
├── setup.ps1                    # Windows setup
├── src/
│   ├── trex_errors/            # Error handling
│   ├── trex_modules/           # Module loader
│   ├── trex_utils/             # Utilities (colors, history, piping)
│   └── trex_help/              # Help system
└── modules/                     # Example modules
    ├── echo.py
    ├── random.py
    └── sysinfo.py
```

## Tips & Tricks

1. **Add to PATH**: `export PATH="$HOME/.t-rex/bin:$PATH"` in your shell profile
2. **Use Aliases**: Create alias in shell: `alias trex='~/.t-rex/bin/t-rex'`
3. **Backup Config**: Keep a copy of `.trexrc` before modifying
4. **Check History**: View recent commands with `~/.t-rex/history`
5. **Exit Codes**: Use exit or quit to leave the shell

## Next Steps

- Read [learn.md](learn.md) to create custom modules
- Check [CONFIG.md](CONFIG.md) for detailed configuration options
- Explore example modules in `modules/` directory

## Support

For issues or questions, check the documentation files included in the project.

Happy shell scripting! 🦖
