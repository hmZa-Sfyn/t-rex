# Configuration Guide

## .trexrc Configuration File

The `.trexrc` file is the main configuration file for T-Rex shell. It's located in your T-Rex home directory:
- **Linux/macOS**: `~/.t-rex/.trexrc`
- **Windows**: `~/.t-rex-windows/.trexrc`

### Configuration Options

```bash
# Module paths - colon-separated paths where T-Rex looks for modules
module_paths=~/.t-rex/modules:/usr/local/lib/t-rex/modules

# Enable or disable colored output
use_colors=true

# Color theme (default is only available theme for now)
theme=default

# History settings
history_enabled=true
history_size=1000

# Prompt customization
prompt_symbol=❯
prompt_color=cyan

# Python3 interpreter location
python_executable=python3
```

### Available Colors for prompt_color
- red
- green
- yellow
- blue
- magenta
- cyan
- white

### Prompt Symbols
You can use any Unicode character:
```bash
prompt_symbol=$        # Dollar sign
prompt_symbol=→        # Arrow
prompt_symbol=>>       # Double angle
prompt_symbol=🦖       # T-Rex emoji
```

## Environment Variables

Set these in your shell's rc file (`.bashrc`, `.zshrc`, etc.):

```bash
# Required for T-Rex development
export GO111MODULE=off
export GOPATH=/path/to/t-rex

# T-Rex specific variables
export TREX_HOME=~/.t-rex
export TREX_MODULES=~/.t-rex/modules
```

## Home Directory Structure

T-Rex creates the following structure on first run:

```
~/.t-rex/          (Linux/macOS)
├── .trexrc         # Main configuration
├── history         # Command history (auto-generated)
└── modules/        # User modules

~/.t-rex-windows/   (Windows)
├── .trexrc
├── history
└── modules/
```

## Module Paths

Modules are searched in this order:
1. Paths in `module_paths` from `.trexrc`
2. Current working directory
3. Default `~/.t-rex/modules`

Add multiple paths by separating with colons:
```bash
module_paths=~/.t-rex/modules:/opt/t-rex/modules:/usr/local/t-rex/modules
```

## Customizing Colors

Edit `.trexrc` to change the prompt color:

```bash
# Cyan prompt (default)
prompt_color=cyan
prompt_symbol=❯

# Red prompt with arrow
prompt_color=red
prompt_symbol=→

# Green prompt with dollar sign
prompt_color=green
prompt_symbol=$
```

## Disabling Color Output

To disable colors:
```bash
use_colors=false
```

This is useful for:
- Piping output to files
- Automation scripts
- Accessibility reasons

## History Settings

```bash
# Enable command history (recommended)
history_enabled=true

# How many commands to keep (default 1000)
history_size=2000

# History is automatically saved to ~/.t-rex/history
```

## Advanced Configuration

### Custom Python Interpreter
If you have multiple Python versions:
```bash
python_executable=python3.11
```

### System-wide Configuration
You can place a system-wide config at:
```
/usr/local/etc/t-rex/.trexrc
/etc/t-rex/.trexrc
```

T-Rex will load in order:
1. System-wide config
2. User config (`~/.t-rex/.trexrc`)
3. Local config (`./.trexrc` in current directory)

## Configuration Format

The `.trexrc` file uses simple key=value format:
- Lines starting with `#` are comments
- Empty lines are ignored
- Whitespace around `=` is trimmed
- Values with spaces should be quoted

```bash
# This is a comment
key=value
module_paths=/path/one:/path/two
prompt_symbol=❯
# prompt_color=red  # This is disabled (commented out)
```

## Example Configurations

### Development Setup
```bash
module_paths=~/.t-rex/modules:./modules:../modules
use_colors=true
history_enabled=true
history_size=5000
prompt_symbol=❯ dev
prompt_color=cyan
```

### Production Setup
```bash
module_paths=/usr/local/lib/t-rex/modules
use_colors=false
history_enabled=true
history_size=100
prompt_symbol=$
python_executable=python3
```

### Minimal Setup
```bash
module_paths=~/.t-rex/modules
use_colors=true
prompt_symbol=❯
```

## Reloading Configuration

Configuration is loaded when T-Rex starts. To apply changes:
1. Edit `.trexrc`
2. Exit current session (`exit`)
3. Start a new T-Rex session

Changes take effect immediately on startup.

## Troubleshooting Configuration

### Colors not working
- Check `use_colors=true`
- Ensure terminal supports ANSI colors
- Try `use_colors=false` to verify color parsing isn't the issue

### Modules not found
- Verify paths in `module_paths` exist
- Check Python files have `.py` extension
- Ensure modules return valid JSON

### History not saving
- Check `history_enabled=true`
- Verify `~/.t-rex/` directory exists
- Check directory permissions: `chmod 755 ~/.t-rex`

## Configuration Best Practices

1. **Keep it simple** - Start with defaults, customize as needed
2. **Comment your changes** - Use `#` to document custom settings
3. **Test modules** - Run modules directly before expecting them to work in T-Rex
4. **Backup config** - Keep a copy of your `.trexrc` before major changes
5. **Document paths** - If using custom module paths, document their purpose

## Next Steps

- See [learn.md](learn.md) to create custom modules
- Check [README.md](README.md) for usage examples
- Run `help` in T-Rex for built-in commands
