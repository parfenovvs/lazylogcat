# Contributing to lazylogcat

Thank you for your interest in contributing!

## Before You Start

- **Bug fixes & small improvements**: Go ahead and open a PR
- **New features or major refactoring**: Open an issue first to discuss

## Quick Start

1. Fork and clone the repository
2. Create a branch from `trunk`
3. Make your changes
4. Run tests and checks:
   ```bash
   go test -v -race ./...
   go fmt ./...
   go vet ./...
   ```
5. **Match CI** (recommended before opening a PR): the workflow also builds the embedded web UI. If you change lazylogcat behavior that the agent should know about, update `skills/lazylogcat/SKILL.md` in the same PR. From the repo root:
   ```bash
   (cd web-ui && bun install)
   go generate ./internal/web/...
   go build -v ./...
   go test -v -race ./...
   ```
6. Commit using [Conventional Commits](https://www.conventionalcommits.org/) format:
   ```
   feat: add device selection screen
   fix: handle empty logcat buffer
   docs: update installation instructions
   ```
7. Push and open a PR against `trunk`

## Code Style

- Format with `go fmt`
- Group imports: stdlib, external, internal (blank line separated)
- Use table-driven tests with `t.Run()` for test cases
- Wrap errors with context: `fmt.Errorf("context: %w", err)`

## Bug Reports

- Search existing issues first
- Include: Go version, OS, steps to reproduce, expected vs actual behavior

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
