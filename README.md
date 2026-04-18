# lazylogcat

[![Build Status](https://github.com/parfenovvs/lazylogcat/actions/workflows/auto.yml/badge.svg?branch=trunk)](https://github.com/parfenovvs/lazylogcat/actions/workflows/auto.yml)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.25-00ADD8?logo=go)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/parfenovvs/lazylogcat)](https://github.com/parfenovvs/lazylogcat/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/parfenovvs/lazylogcat)](https://goreportcard.com/report/github.com/parfenovvs/lazylogcat)

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

```bash
brew tap parfenovvs/lazylogcat && brew install lazylogcat
```

## Usage

```bash
# Launch the TUI
lazylogcat

# Enable debug logging (to .lazylogcat.log file in working directory)
lazylogcat --debug
```

### Agent skill (Cursor / Claude Code)

The binary embeds a skill you can install for coding agents:

```bash
lazylogcat skill install --agent cursor --user
lazylogcat skill install --agent claude --project
```

Use `--user` for a global install under your home directory, or `--project` to install next to the current project (the shell working directory matters). Requires `--agent` and exactly one of `--user` or `--project`.

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

This project wouldn't exist without these open source projects. The TUI stack uses **Bubble Tea v2** and related libraries published on `charm.land` (successors to the original [Charm](https://github.com/charmbracelet) GitHub modules):

- **[Bubble Tea](https://pkg.go.dev/charm.land/bubbletea/v2)** (`charm.land/bubbletea/v2`)
- **[Bubbles](https://pkg.go.dev/charm.land/bubbles/v2)** (`charm.land/bubbles/v2`)
- **[Lip Gloss](https://pkg.go.dev/charm.land/lipgloss/v2)** (`charm.land/lipgloss/v2`)
- **[Cobra](https://github.com/spf13/cobra)** by spf13
- **[clipboard](https://github.com/atotto/clipboard)** by atotto

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
