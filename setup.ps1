# T-Rex Shell Setup Script for Windows
# This script sets up T-Rex with configuration directories and installs the binary

$ErrorActionPreference = "Stop"

Write-Host "🦖 T-Rex Shell Setup" -ForegroundColor Cyan
Write-Host "====================" -ForegroundColor Cyan
Write-Host ""

# Detect Windows and set T-Rex home
$TREX_HOME = "$env:USERPROFILE\.t-rex-windows"

Write-Host "📍 Detected: Windows" -ForegroundColor Green
Write-Host "📁 Setting up T-Rex home directory: $TREX_HOME" -ForegroundColor Yellow

# Create T-Rex home directory structure
@("$TREX_HOME", "$TREX_HOME\modules", "$TREX_HOME\bin") | ForEach-Object {
    if (-not (Test-Path $_)) {
        New-Item -ItemType Directory -Path $_ -Force | Out-Null
    }
}

Write-Host "✓ Created directories" -ForegroundColor Green

# Copy default configuration if it doesn't exist
$configPath = "$TREX_HOME\.trexrc"
if (-not (Test-Path $configPath)) {
    $config = @"
# T-Rex Shell Configuration
module_paths=~/.t-rex-windows/modules
use_colors=true
theme=default
history_enabled=true
history_size=1000
prompt_symbol=❯
prompt_color=cyan
python_executable=python3
"@
    Set-Content -Path $configPath -Value $config -Encoding UTF8
    Write-Host "✓ Created .trexrc configuration file" -ForegroundColor Green
} else {
    Write-Host "✓ .trexrc already exists" -ForegroundColor Green
}

# Copy example modules
$modulesPath = ".\modules"
if (Test-Path $modulesPath) {
    Write-Host "📦 Installing example modules..." -ForegroundColor Yellow
    Get-ChildItem "$modulesPath\*.py" | ForEach-Object {
        Copy-Item $_.FullName "$TREX_HOME\modules\" -Force
    }
    Write-Host "✓ Modules installed" -ForegroundColor Green
}

# Build the binary if main.go exists
if (Test-Path "main.go") {
    Write-Host "🔨 Building T-Rex binary..." -ForegroundColor Yellow
    
    $env:GO111MODULE = "off"
    $env:GOPATH = (Get-Location).Path
    
    & go build -o "$TREX_HOME\bin\t-rex.exe" main.go
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Binary built successfully" -ForegroundColor Green
    } else {
        Write-Host "✗ Build failed" -ForegroundColor Red
        exit 1
    }
}

Write-Host ""
Write-Host "✅ Setup complete!" -ForegroundColor Green
Write-Host ""
Write-Host "🚀 Usage:" -ForegroundColor Cyan
Write-Host "   & `"$TREX_HOME\bin\t-rex.exe`"       # Run T-Rex directly"
Write-Host "   & `"~\.t-rex-windows\bin\t-rex.exe`" # Or from home directory"
Write-Host ""
Write-Host "📝 To add T-Rex to PATH:" -ForegroundColor Yellow
Write-Host "   1. Open Environment Variables (System Properties)"
Write-Host "   2. Add: $TREX_HOME\bin"
Write-Host ""
Write-Host "📖 Configuration:" -ForegroundColor Cyan
Write-Host "   Edit: $TREX_HOME\.trexrc"
Write-Host ""
Write-Host "🐍 Add custom Python modules to:" -ForegroundColor Cyan
Write-Host "   $TREX_HOME\modules\"
Write-Host ""
Write-Host "📜 History is saved to:" -ForegroundColor Cyan
Write-Host "   $TREX_HOME\history"
Write-Host ""

# Offer to add to PATH
$addPath = Read-Host "Would you like to add T-Rex to your PATH? (Y/N)"
if ($addPath -eq "Y" -or $addPath -eq "y") {
    $currentPath = [System.Environment]::GetEnvironmentVariable("PATH", "User")
    $newPath = "$TREX_HOME\bin;$currentPath"
    [System.Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
    Write-Host "✓ Added to PATH. Restart terminal to apply changes." -ForegroundColor Green
}
