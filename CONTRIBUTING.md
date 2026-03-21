# Contributing to CCGears

Thanks for your interest in contributing to CCGears!

## Getting started

1. Fork and clone the repo
2. Install Go 1.20+
3. Build: `go build -o ccgears ./cmd/ccgears`
4. Run tests: `go test ./internal/... -v`

## Development workflow

1. Create a branch for your feature or fix
2. Make your changes
3. Run tests: `go test ./internal/...`
4. Run vet: `go vet ./...`
5. Submit a pull request

## Project structure

```
cmd/ccgears/         CLI entry point
internal/
  config/            Config loading, path resolution
  preset/            CRUD operations, scan, diff
  preview/           Parse skills, permissions, CLAUDE.md
  snapshot/          File copying, backup/restore
  ui/                Bubble Tea TUI
  validate/          Name validation
```

## Code style

- Standard Go formatting (`gofmt`)
- No external dependencies beyond Bubble Tea / Lip Gloss / Bubbles
- Tests use `t.TempDir()` for filesystem isolation
- Handle errors explicitly — no silent failures

## Adding features

- **New menu option**: Add to `menuLabels` in `app.go`, add state handling in `Update()` and `View()`
- **New CLI subcommand**: Add to the switch in `main.go`
- **New skill**: Add template in `preset.go`, update `ensureCCGearsSkill()`

## Reporting issues

Open an issue on GitHub with:
- What you expected to happen
- What actually happened
- Your OS and Go version
- Steps to reproduce
