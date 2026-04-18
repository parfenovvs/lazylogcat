---
name: lazylogcat
description: >
  How to use the lazylogcat CLI to capture, filter, and inspect Android logcat non-interactively.
  Use this skill whenever an AI agent needs to: stream or dump Android logs, filter by package/tag/text,
  parse log output programmatically, check connected devices, read or validate lazylogcat config,
  or install this skill into an agent environment.
  Also use when the user asks about lazylogcat CLI flags, config files, or automating logcat workflows
  from scripts or agent pipelines. Do not use for generic Android app development with no logging
  involved, or for contributing to lazylogcat source unless the user explicitly asks.
---

# lazylogcat (AI agent guide)

lazylogcat is an Android `logcat` tool. As an AI agent you will always use it **non-interactively via CLI**—never launch the bare `lazylogcat` command (that opens an interactive TUI which blocks). Use the subcommands below instead.

## Prerequisites

`adb` must be installed and on `PATH`. Verify with `adb devices` before running any log commands. If `adb` is missing, tell the user to install Android platform tools.

## CLI subcommands

### Capture and dump logs

```bash
# Dump recent logs to stdout (non-blocking, use for scripting)
lazylogcat logs dump [--pkg <pkg>] [--tag <tag>] [--text <text>] [--lines <n>]

# Parse a saved logcat file
lazylogcat logs parse <file>
```

Both `--pkg`, `--tag`, and `--text` are **contains** matches. Combine freely.

### List connected devices

```bash
lazylogcat devices
```

Returns connected ADB devices as a list. Use this before log commands to confirm a device is reachable.

### Inspect resolved config

```bash
lazylogcat config show
```

Prints the fully merged config (all layers applied) as JSON. Use to debug unexpected filter behavior.

### Debug logging

Add `--debug` to any command to write a `.lazylogcat.log` file in the cwd. This traces lazylogcat internals, not Android log verbosity.

## Configuration

Configs merge in order; **later files override earlier ones**:

1. OS config dir → `lazylogcat/config.json` (Linux: `~/.config/lazylogcat/config.json`; macOS: `~/Library/Application Support`; Windows: `%AppData%`)
2. Project: `.lazylogcat/config.json`
3. Local overrides: `.lazylogcat/config.local.json`

Override config paths per-run:
```bash
lazylogcat logs dump --config <path> --config-local <path>
```

