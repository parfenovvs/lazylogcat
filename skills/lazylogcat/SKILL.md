---
name: lazylogcat
description: >
  How to use the lazylogcat CLI and TUI to view and filter Android logcat on a device or emulator.
  Use this skill whenever the user is debugging Android apps, reading crash or ANR logs, tailing
  logcat, filtering by package/tag/text, using lazylogcat or asking about "logcat in the terminal,"
  or lightweight log viewers outside Android Studio. Also use when the user needs help with
  lazylogcat config files or CLI flags. Do not use this skill for generic Android app development
  with no logging or devices involved, or for building or contributing to the lazylogcat project
  itself unless the user explicitly asks to change lazylogcat source code.
---

# lazylogcat (agent guide)

lazylogcat is a **terminal UI** for Android `logcat`. Your job is to help the user **run it correctly**, pick **filters**, and interpret **config**—not to describe how to compile lazylogcat from source unless they ask.

## Prerequisites

**Device or emulator**: `adb` must be installed and on `PATH` (`adb devices` should list the device). If `adb` is missing, say so and point to Android platform tools—do not assume lazylogcat will stream logs without it.

## What to run

| Goal | Command |
|------|---------|
| Interactive log viewer (default) | `lazylogcat` |
| Same with file debug log in cwd | `lazylogcat --debug` (writes `.lazylogcat.log`) |
| Override filters for this run | add `--pkg`, `--tag`, and/or `--text` (each is **contains** match) |
| Custom config file locations | `--config` (project config path), `--config-local` (local overrides path) |

## Configuration (for the user’s project)

Configs merge in order; **later files override earlier ones**:

1. OS config directory → `lazylogcat/config.json` (e.g. Linux `~/.config/lazylogcat/config.json`, macOS under `~/Library/Application Support`, Windows under `%AppData%`).
2. Project: `.lazylogcat/config.json`.
3. Local overrides: `.lazylogcat/config.local.json`.

If the user’s filters behave wrong, check which layer applies and prefer adjusting **project** or **local** files so machine-specific paths stay out of git when needed. JSON shape and keys: see the published schema URL in the lazylogcat repo’s `config.schema.json` if you need field names—do not invent keys.

## Behavior hints

- In some terminals, **mouse selection** in the TUI needs **Shift** held—mention this if the user says click/select does not work.
- `--debug` affects logging for troubleshooting lazylogcat itself, not Android log verbosity.

## Installing this skill (end users)

The binary can copy this skill into Cursor or Claude paths:

`lazylogcat skill install --agent cursor|claude --user` or `--project` (exactly one of `--user` / `--project`; `--project` uses the **current working directory**).

Print the path the command echoes so the user knows where files landed.
