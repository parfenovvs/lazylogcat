# lazylogcat

[![Build Status](https://github.com/parfenovvs/lazylogcat/actions/workflows/auto.yml/badge.svg?branch=trunk)](https://github.com/parfenovvs/lazylogcat/actions/workflows/auto.yml)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.24-00ADD8?logo=go)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/parfenovvs/lazylogcat)](https://github.com/parfenovvs/lazylogcat/releases/latest)

A TUI for viewing Android logcat logs.

## Motivation

Reading Android logcat logs shouldn't require heavy tooling:
- **adb CLI is cumbersome**: Raw `adb logcat` is difficult to use
- **IDEs are resource-heavy**: Android Studio consumes significant memory just to view logs
- **Alternative editors lack logcat support**: Developers using Zed, VS Code, Neovim, or other editors have no integrated option

lazylogcat provides a lightweight, focused TUI for viewing and filtering logcat logs without the overhead.

![demo](demo.svg)

## Prerequisites

- `adb` (Android Debug Bridge) installed and in PATH (see [ADB Installation Guide](https://developer.android.com/tools/adb))

## Installation

### Option 1: Homebrew (Recommended)

```bash
brew tap parfenovvs/lazylogcat && brew install lazylogcat
```

### Option 2: Install via go install

Requires Go 1.24 or higher.

```bash
go install github.com/parfenovvs/lazylogcat@latest
```

**Important:** This installs the binary to your `$GOBIN` directory (or `$GOPATH/bin` if `GOBIN` is not set). Ensure this directory is in your `PATH`:

```bash
# Check where the binary is installed
go env GOBIN    # If empty, defaults to $(go env GOPATH)/bin

# Verify installation
which lazylogcat
```

If `lazylogcat` is not found, add the Go bin directory to your `PATH` environment variable.

### Option 3: Build from Source

Requires Go 1.24 or higher.

```bash
# Clone the repository
git clone https://github.com/parfenovvs/lazylogcat.git
cd lazylogcat

# Build the binary
go build -o lazylogcat .

# Run directly
./lazylogcat

Optional: Move the binary to a directory in your PATH
```

## Usage

```bash
# Launch the TUI
lazylogcat

# Enable debug logging (to .lazylogcat.log file in working directory)
lazylogcat --debug
```

### Tips

- Select with mouse requires holding `Shift` key in some terminal emulators

## Configuration

Configuration is automatically discovered and merged in layers:

`<config-dir>/lazylogcat/config.json` → `.lazylogcat/config.json` → `.lazylogcat/config.local.json`

Where `<config-dir>` is the OS-specific user config directory:
- **Linux**: `~/.config`
- **macOS**: `~/Library/Application Support`
- **Windows**: `%AppData%`

Each layer overrides the previous. All fields are optional. See the [JSON Schema](config.schema.json) for editor autocompletion.

```json
{
  "$schema": "https://github.com/parfenovvs/lazylogcat/raw/trunk/config.schema.json",
  "display": {
    "color": true,
    "wrap": true,
    "columns": {
      "date": false,
      "time": true,
      "pid": false,
      "tid": false,
      "level": true,
      "tag": true,
      "message": true
    }
  },
  "filter": {
    "package_name": "com.example.app",
    "log_tag": "MainActivity",
    "log_text": "error"
  }
}
```

## Acknowledgments

This project wouldn't exist without these incredible open source projects:

- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** by Charm
- **[Bubbles](https://github.com/charmbracelet/bubbles)** by Charm
- **[Lip Gloss](https://github.com/charmbracelet/lipgloss)** by Charm
- **[Cobra](https://github.com/spf13/cobra)** by spf13
- **[clipboard](https://github.com/atotto/clipboard)** by atotto

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
