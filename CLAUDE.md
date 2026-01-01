# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
# Build binary (creates ./pomodoro)
make

# Install to GOPATH
make install

# Run all tests (unit + acceptance)
make test

# Run only Go unit tests
make test-unit

# Run only bats acceptance tests
make test-acceptance

# Run a single unit test
go test -run TestFunctionName ./cmd/...

# Run a single bats test file
./test/.bats/bats-core/bin/bats test/start.bats

# Clean build artifacts
make clean
```

## Architecture

This is a Go CLI application using the [Open Pomodoro Format](https://github.com/open-pomodoro/open-pomodoro-format) for data storage. The CLI is built with [cobra](https://github.com/spf13/cobra) for command handling.

### Core Dependencies

- **github.com/open-pomodoro/go-openpomodoro**: Core library handling Pomodoro data format, persistence (`~/.pomodoro/`), and settings parsing
- **github.com/spf13/cobra**: CLI command framework

### Package Structure

- `cmd/` - All CLI commands as separate files (start.go, status.go, etc.). Each file registers its command in `init()` via `RootCmd.AddCommand()`
- `cmd/root.go` - Root command setup, initializes the `client` (openpomodoro.Client) and `settings` in `PersistentPreRun`
- `format/` - Status output formatting with format string substitution (`%r`, `%d`, `%t`, etc.)
- `hook/` - Shell hook execution for start/stop/break events (runs scripts from `~/.pomodoro/hooks/`)

### Key Patterns

- Commands access shared `client` and `settings` variables from `cmd/root.go`
- Most commands that modify state call `hook.Run(client, "hookname")` after changes
- Status display uses `format.Format(state, formatFlag)` for customizable output
- Version is injected via `-ldflags` at build time: `-X main.Version=$(VERSION)`

### Data Storage

All data is stored in `~/.pomodoro/` (or path specified via `--directory` flag):
- `current` - Active pomodoro state
- `history` - Completed pomodoros log
- `settings` - User configuration (logfmt format)
- `hooks/` - Executable scripts for lifecycle events

### Testing

- Unit tests: Standard Go tests in `*_test.go` files
- Acceptance tests: Bats (Bash Automated Testing System) in `test/*.bats`
- Test helper (`test/test_helper.bash`) provides `pomodoro()` wrapper that uses isolated temp directories

## Progress Log

### 2026-01-01: MCP Server Extension

- Created MCP server extension for Claude Code integration
- Added `mcp/` package with `server.go` and `tools.go`
- Implemented 10 MCP tools: start_pomodoro, get_status, finish_pomodoro, cancel_pomodoro, clear_pomodoro, start_break, repeat_pomodoro, amend_pomodoro, get_history, get_settings
- Added `cmd/pomodoro-mcp/main.go` entry point for stdio transport
- Changed module path from `github.com/open-pomodoro/openpomodoro-cli` to `github.com/BuddhiLW/openpomodoro-cli`
- Updated all cmd/*.go imports to use new module path
- Updated Makefile with MCP binary targets
- Fixed `.gitignore` to use `/pomodoro-mcp` (only ignore binary at root)
- Released v0.4.1 with MCP support
- Install: `go install github.com/BuddhiLW/openpomodoro-cli/cmd/pomodoro-mcp@v0.4.1`
