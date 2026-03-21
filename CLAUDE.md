# CCGears

Go CLI tool for managing Claude Code configuration presets.

## Project structure

- `cmd/ccgears/main.go` — CLI entry point, subprocess management, `switch`/`save` subcommands
- `internal/ui/` — Bubble Tea TUI (app.go state machine, styles.go, components.go)
- `internal/preset/` — CRUD operations, scan, diff, skill injection
- `internal/config/` — Config loading, path resolution
- `internal/snapshot/` — File copying with exclusions, backup/restore
- `internal/preview/` — Parse SKILL.md frontmatter, permissions, CLAUDE.md
- `internal/validate/` — Preset name validation

## Build

```bash
go build -o ccgears ./cmd/ccgears
go test ./internal/... -v
```

## Key design decisions

- CCGears acts as a parent process that launches `claude` as a subprocess
- `/ccgears` skill uses `ccgears switch` which sends SIGINT to Claude Code
- Skills and permissions are injected into `.claude/` after every preset load via `ensureCCGearsSkill()`
- Session continuity uses `claude --resume <session-id>` found from the newest `.jsonl` in `~/.claude/projects/`
