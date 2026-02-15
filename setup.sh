#!/bin/bash

# T-Rex Shell Setup Script for Linux/macOS
# This script sets up T-Rex with configuration directories and installs the binary

set -e

echo "🦖 T-Rex Shell Setup"
echo "===================="
echo ""

# Detect OS
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    TREX_HOME="$HOME/.t-rex"
    echo "📍 Detected: Linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    TREX_HOME="$HOME/.t-rex"
    echo "📍 Detected: macOS"
else
    echo "❌ Unsupported OS. Please use setup.ps1 on Windows"
    exit 1
fi

echo "📁 Setting up T-Rex home directory: $TREX_HOME"

# Create T-Rex home directory structure
mkdir -p "$TREX_HOME"
mkdir -p "$TREX_HOME/modules"
mkdir -p "$TREX_HOME/bin"

echo "✓ Created directories"

# Copy default configuration if it doesn't exist
if [ ! -f "$TREX_HOME/.trexrc" ]; then
    cat > "$TREX_HOME/.trexrc" << 'EOF'
# T-Rex Shell Configuration
module_paths=~/.t-rex/modules
use_colors=true
theme=default
history_enabled=true
history_size=1000
prompt_symbol=❯
prompt_color=cyan
python_executable=python3
EOF
    echo "✓ Created .trexrc configuration file"
else
    echo "✓ .trexrc already exists"
fi

# Copy example modules
if [ -d "modules" ]; then
    echo "📦 Installing example modules..."
    for module in modules/*.py; do
        if [ -f "$module" ]; then
            cp "$module" "$TREX_HOME/modules/"
            chmod +x "$TREX_HOME/modules/$(basename $module)"
        fi
    done
    echo "✓ Modules installed"
fi

# Build the binary if main.go exists
if [ -f "main.go" ]; then
    echo "🔨 Building T-Rex binary..."
    export GO111MODULE=off
    export GOPATH=$(pwd)
    go build -o "$TREX_HOME/bin/t-rex" main.go
    chmod +x "$TREX_HOME/bin/t-rex"
    echo "✓ Binary built successfully"
fi

# Create symlink in /usr/local/bin if possible
if [ -w "/usr/local/bin" ]; then
    ln -sf "$TREX_HOME/bin/t-rex" /usr/local/bin/t-rex
    echo "✓ Created symlink: /usr/local/bin/t-rex"
    echo ""
    echo "🚀 You can now run: t-rex"
else
    echo ""
    echo "📝 To use t-rex from anywhere, add to your ~/.bashrc or ~/.zshrc:"
    echo "   export PATH=\"$TREX_HOME/bin:\$PATH\""
fi

echo ""
echo "✅ Setup complete!"
echo ""
echo "📖 Usage:"
echo "   $TREX_HOME/bin/t-rex       # Run T-Rex directly"
echo "   ~/.t-rex/bin/t-rex         # Or from home directory"
echo ""
echo "📚 Configuration:"
echo "   Edit: $TREX_HOME/.trexrc"
echo ""
echo "🐍 Add custom Python modules to:"
echo "   $TREX_HOME/modules/"
echo ""
echo "📜 History is saved to:"
echo "   $TREX_HOME/history"
echo ""

# Add or update alias in ~/.bashrc and ~/.zshrc
add_or_update_alias() {
    local name="$1"
    local cmd="$2"
    if [ -z "$name" ] || [ -z "$cmd" ]; then
        echo "Usage: add_or_update_alias <name> <command>"
        return 1
    fi
    local files=("$HOME/.bashrc" "$HOME/.zshrc")
    local alias_line="alias ${name}='${cmd}'"
    for file in "${files[@]}"; do
        mkdir -p "$(dirname "$file")"
        touch "$file"
        if grep -qE "^[[:space:]]*alias[[:space:]]+${name}=" "$file"; then
            sed -i.bak -E "s|^[[:space:]]*alias[[:space:]]+${name}=.*|${alias_line}|" "$file" && rm -f "${file}.bak"
            echo "Updated alias in $file"
        else
            echo "" >> "$file"
            echo "# Added by t-rex setup on $(date)" >> "$file"
            echo "${alias_line}" >> "$file"
            echo "Added alias to $file"
        fi
    done
}

# If a built binary exists, create/update a convenient alias
if [ -n "$TREX_HOME" ] && [ -x "$TREX_HOME/bin/t-rex" ]; then
    add_or_update_alias t-rex "$TREX_HOME/bin/t-rex" || true
fi
