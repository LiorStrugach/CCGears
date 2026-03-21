# CCGears

**Dynamic Preset Loader for Claude Code**

```
   ___  ___  ___
  / __|/ __|/ __|___ __ _ _ _ ___
 | (__| (__| (_ / -_) _` | '_(_-<
  \___|\___|\___|___\__,_|_| /__/
```

CCGears eliminates the context-switching tax in [Claude Code](https://docs.anthropic.com/en/docs/claude-code). Every task type requires a different combination of skills, MCP servers, and tool permissions. CCGears lets you snapshot a configuration once, save it as a named preset, and switch to it instantly.

No more manually remembering which tools belong to which context. No more re-enabling the same permissions at the start of every session.

---

## Features

- **Preset management** — Create, load, save, and delete named configuration snapshots
- **Auto-scan** — Discover Claude Code configs across your filesystem and import them in bulk
- **Preview before load** — See skills, permissions, tools, and CLAUDE.md headline before switching
- **Diff before save** — Review what changed before updating a preset
- **Auto-backup** — Current state is backed up before every load, with one-click undo
- **Launch integration** — Optionally launch `claude` directly after loading a preset
- **Interactive TUI** — Arrow-key navigation powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Cross-platform** — Builds for macOS, Linux, and Windows
- **Zero config** — Works out of the box, stores presets in `~/.ccgears/`

## What gets captured

A preset snapshots three things from your project directory:

| Source | Stored as | Contains |
|--------|-----------|----------|
| `.claude/` | `claude/` | `settings.local.json` (permissions), `skills/` (custom skills with SKILL.md) |
| `tools/` | `tools/` | CLI scripts (.py, .sh, .js) that Claude Code calls |
| `CLAUDE.md` | `CLAUDE.md` | Project-level instructions for Claude Code |

Excluded automatically: `.DS_Store`, `__pycache__`

---

## Installation

### From source (requires Go 1.20+)

```bash
git clone https://github.com/mryan/ccgears.git
cd ccgears
go install ./cmd/ccgears
```

The binary is installed to `$GOPATH/bin/ccgears` (typically `~/go/bin/ccgears`).

Make sure `~/go/bin` is in your PATH:

```bash
# Add to ~/.zshrc or ~/.bashrc
export PATH="$PATH:$HOME/go/bin"
```

### Build locally

```bash
go build -o ccgears ./cmd/ccgears
./ccgears
```

### Cross-compile

```bash
GOOS=linux GOARCH=amd64 go build -o ccgears-linux ./cmd/ccgears
GOOS=windows GOARCH=amd64 go build -o ccgears.exe ./cmd/ccgears
```

---

## Quick Start

### 1. Create your first preset

Navigate to a project that has a `.claude/` directory:

```bash
cd ~/projects/my-infra-project
ccgears
```

Select **Create preset**, enter a name like `infra-tools`, and an optional description. CCGears snapshots your current `.claude/`, `tools/`, and `CLAUDE.md` into `~/.ccgears/presets/infra-tools/`.

### 2. Load a preset

Navigate to any project directory and load the preset:

```bash
cd ~/projects/different-project
ccgears
```

Select **Load preset**, pick `infra-tools`, review the preview, and confirm. Your `.claude/` and `tools/` are replaced with the preset's contents. The previous state is backed up automatically.

After loading, press **Enter** to launch Claude Code directly, or **q** to return to the menu.

### 3. Scan and import existing configs

Select **Scan & import** from the menu, enter a root directory (e.g. `~/Documents`), and CCGears will find all projects with Claude Code configurations. Use Space to toggle which ones to import, then Enter to confirm.

---

## Usage

### Interactive mode

```bash
ccgears
```

Launches the TUI with arrow-key navigation:

```
╭──────────────────────────────────╮
│                                  │
│   ▸ Load preset                  │
│     Save preset                  │
│     Create preset                │
│     Scan & import                │
│     Delete preset                │
│     Undo last load               │
│     Exit                         │
│                                  │
╰──────────────────────────────────╯
```

**Controls:** `↑↓` or `j/k` to navigate, `Enter` to select, `q` or `Esc` to go back.

### Non-interactive mode

For scripting and shell aliases:

```bash
# Load a preset directly
ccgears load infra-tools

# List all presets
ccgears list

# List as JSON (for scripts)
ccgears list --json

# Show help
ccgears help
```

### Shell alias example

```bash
# Add to ~/.zshrc
alias cc-infra="cd ~/projects/infra && ccgears load infra-tools && claude"
alias cc-creative="cd ~/projects/creative && ccgears load watermelon && claude"
```

---

## Menu Options

### Load preset

1. Select a preset from the list
2. Preview its contents — skills, permissions, tools, CLAUDE.md headline
3. Confirm — current `.claude/` and `tools/` are backed up, then replaced
4. Press Enter to launch `claude`, or q to return to the menu

### Save preset

Updates an existing preset with your current project state. Shows a diff summary before confirming:

```
  Changes to infra-tools:

    + tools/new_script.py              (added)
    ~ .claude/settings.local.json      (modified)
    - tools/old_script.py              (removed)

    2 unchanged
```

### Create preset

Captures the current `.claude/`, `tools/`, and `CLAUDE.md` into a new named preset. Shows file counts and sizes before confirming.

**Preset name rules:** lowercase letters, digits, and hyphens only. Max 48 characters. No leading or trailing hyphens.

### Scan & import

Discovers Claude Code configurations across your filesystem:

1. Enter a root directory to scan (default: home directory)
2. CCGears walks up to 4 levels deep, skipping `node_modules`, `.git`, `.venv`, `vendor`, etc.
3. Multi-select which projects to import (Space to toggle, `a` for all, `n` for none)
4. Creates presets automatically with names derived from directory names

### Delete preset

Removes a preset permanently. Requires typing the preset name to confirm.

### Undo last load

Restores the state from before the most recent Load operation. Only one undo level is available (the last load).

---

## Preset Storage

All data is stored locally in `~/.ccgears/`:

```
~/.ccgears/
├── config.json                  # Settings (default preset, custom store path)
├── presets/
│   ├── infra-tools/
│   │   ├── preset.json          # Metadata: name, description, timestamps
│   │   ├── claude/              # Snapshot of .claude/
│   │   ├── tools/               # Snapshot of tools/
│   │   └── CLAUDE.md            # Snapshot of CLAUDE.md (if present)
│   └── watermelon/
│       └── ...
└── backup/
    └── last/                    # Auto-backup before each load
        ├── backup_meta.json
        ├── claude/
        ├── tools/
        └── CLAUDE.md
```

### Custom storage location

Override with the `CCGEARS_HOME` environment variable:

```bash
export CCGEARS_HOME="/path/to/custom/location"
```

Or set `store_path` in `~/.ccgears/config.json`:

```json
{
  "default_preset": "infra-tools",
  "store_path": "/path/to/custom/presets"
}
```

---

## How It Works

CCGears operates purely via the filesystem. It does not modify Claude Code internals.

**Load** copies files from the preset store into your project directory:
1. Backs up current `.claude/`, `tools/`, `CLAUDE.md` to `~/.ccgears/backup/last/`
2. Removes existing `.claude/` and `tools/` from the project
3. Copies the preset's stored directories into the project
4. Updates the preset's "last used" timestamp

**Create/Save** copies files from your project into the preset store:
- Symlinks are dereferenced (target content is copied, not the link)
- `.DS_Store` and `__pycache__` are excluded automatically

**Undo** reverses the last load by restoring from the backup slot.

---

## Development

### Project structure

```
CCGears/
├── cmd/ccgears/main.go          # CLI entry point
├── internal/
│   ├── config/                   # Config loading, path resolution
│   ├── preset/                   # CRUD operations, scan, diff
│   ├── preview/                  # Parse skills, permissions, CLAUDE.md
│   ├── snapshot/                 # File copying, backup/restore
│   ├── ui/                       # Bubble Tea TUI (app, styles, components)
│   └── validate/                 # Name validation, path checks
├── go.mod
├── Makefile
└── README.md
```

### Build and test

```bash
make build     # Build binary
make test      # Run all tests
make install   # Install to $GOPATH/bin
make clean     # Remove binary
```

### Run tests

```bash
go test ./internal/... -v
```

### Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — Terminal UI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Style definitions
- [Bubbles](https://github.com/charmbracelet/bubbles) — TUI components (text input)

All other functionality uses the Go standard library.

---

## Limitations

- **Single backup slot** — Only the last load can be undone. Loading twice overwrites the first backup.
- **No hot-reload** — Presets are loaded before starting Claude Code, not during a session. Restart `claude` to pick up changes.
- **Preset names are global** — Names must be unique across all projects. Use descriptive names like `infra-frigate` instead of just `default`.
- **Large workspaces** — Skill evaluation workspaces (images, videos) are included in snapshots. Presets for projects with large eval artifacts will be correspondingly large.

---

## License

MIT
