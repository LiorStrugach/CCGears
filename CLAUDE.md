# CCGears

Go CLI tool for managing Claude Code configuration presets.

## Project structure

- `cmd/ccgears/main.go` — CLI entry point, subprocess management, `switch`/`save`/`rename` subcommands
- `cmd/ccgears/signal_unix.go` — Unix signal handling (SIGINT ignore/reset, process tree kill)
- `cmd/ccgears/signal_windows.go` — Windows signal handling (taskkill)
- `internal/ui/` — Bubble Tea TUI (app.go state machine, styles.go, components.go)
- `internal/preset/` — CRUD operations, scan, diff, rename, skill injection
- `internal/config/` — Config loading, path resolution, active preset tracking
- `internal/snapshot/` — File copying with exclusions, backup/restore
- `internal/preview/` — Parse SKILL.md frontmatter, permissions, CLAUDE.md
- `internal/validate/` — Preset name validation

## Build

```bash
go build -o ccgears ./cmd/ccgears
go test ./internal/... -v
```

## Key design decisions

- Main menu is preset-centric: presets are listed directly, actions via hotkeys (Enter=load, s=save, r=rename, d=delete, p=preview)
- CCGears acts as a parent process that launches `claude` as a subprocess
- `/ccgears` skill uses `ccgears switch` which sends SIGINT to Claude Code
- `/ccgears-save` skill uses `ccgears save` to persist changes inline
- Skills and permissions are injected into `.claude/` after every preset load via `ensureCCGearsSkill()`
- Session continuity uses `claude --resume <session-id>` found from the newest `.jsonl` in `~/.claude/projects/`
- Platform-specific code in build-tagged files (`signal_unix.go`, `signal_windows.go`)
