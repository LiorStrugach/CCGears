# CCGears

**Dynamic Preset Loader for Claude Code**

```
   ___  ___  ___
  / __|/ __|/ __|___ __ _ _ _ ___
 | (__| (__| (_ / -_) _` | '_(_-<
  \___|\___|\___|___\__,_|_| /__/
```

CCGears eliminates the context-switching tax in [Claude Code](https://docs.anthropic.com/en/docs/claude-code). Every task type requires a different combination of skills, MCP servers, and tool permissions. CCGears lets you snapshot a configuration (.claude folder) once, save it as a named preset, and switch to it instantly across all of your projects.

No more manually remembering which tools belong to which context. No more re-enabling the same permissions at the start of every session.

---

## Features

- **Preset management** — Create, load, save, and delete named configuration snapshots
- **Auto-scan** — Discover Claude Code configs across your filesystem and import them in bulk
- **Preview before load** — See skills, permissions, tools, and CLAUDE.md headline before switching
- **Diff before save** — Review what changed before updating a preset
- **Auto-backup** — Current state is backed up before every load, with one-click undo
- **Seamless Claude Code integration** — `/ccgears` and `/ccgears-save` commands work inside Claude Code sessions
- **Session continuity** — Switch presets mid-conversation and resume right where you left off
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

### Download a release (recommended)

Download the latest binary for your platform from the [Releases page](https://github.com/MrYanMYN/CCGears/releases).

| Platform | File |
|----------|------|
| macOS (Apple Silicon) | `ccgears_*_mac_arm64.tar.gz` |
| macOS (Intel) | `ccgears_*_mac_amd64.tar.gz` |
| Linux (x86_64) | `ccgears_*_linux_amd64.tar.gz` |
| Linux (ARM64) | `ccgears_*_linux_arm64.tar.gz` |
| Windows (x86_64) | `ccgears_*_windows_amd64.zip` |
| Windows (ARM64) | `ccgears_*_windows_arm64.zip` |

#### macOS / Linux

```bash
# Download and extract (example for macOS Apple Silicon)
tar -xzf ccgears_*_mac_arm64.tar.gz

# Move to a directory in your PATH
sudo mv ccgears /usr/local/bin/

# Verify
ccgears version
```

#### Windows

1. Download and extract the `.zip` file
2. Move `ccgears.exe` to a directory in your PATH (e.g. `C:\Users\<you>\bin\`)
3. Or add the extracted folder to your PATH via System Settings > Environment Variables
4. Open a new terminal and run `ccgears version`

### From source (requires Go 1.20+)

```bash
git clone https://github.com/MrYanMYN/CCGears.git
cd CCGears
go install ./cmd/ccgears
```

The binary is installed to `~/go/bin/ccgears`. Make sure it's in your PATH:

```bash
# Add to ~/.zshrc or ~/.bashrc
export PATH="$PATH:$HOME/go/bin"
```

---

## Quick Start

### 1. Create your first preset

Navigate to a project that has a `.claude/` directory:

```bash
cd ~/projects/my-infra-project
ccgears
```

Select **Create new preset +** at the bottom of the menu, enter a name like `infra-tools`, and an optional description. CCGears snapshots your current `.claude/`, `tools/`, and `CLAUDE.md` into `~/.ccgears/presets/infra-tools/`.

### 2. Load a preset

Navigate to any project directory and load the preset:

```bash
cd ~/projects/different-project
ccgears
```

Highlight a preset with arrow keys, press **Enter** to preview, then **Enter** again to load. Your `.claude/` and `tools/` are replaced with the preset's contents. The previous state is backed up automatically.

After loading, press **Enter** to launch Claude Code directly, or **q** to return to the menu.

### 3. Scan and import existing configs

Select **Scan & import** at the bottom of the menu, enter a root directory (e.g. `~/Documents`), and CCGears will find all projects with Claude Code configurations. Use Space to toggle which ones to import, then Enter to confirm.

### 4. Switch presets from inside Claude Code

While working in a Claude Code session, type:

```
/ccgears
```

This automatically exits the session, opens CCGears to pick a new preset, and resumes the same conversation after loading. No manual steps needed.

---

## Usage

### Interactive mode

```bash
ccgears
```

Launches the preset-centric TUI:

```
  Presets

  ▸ frigate              Imported from ~/Documents/ArguFrigate/frigate     2 days ago
    testmaker            Imported from ~/Documents/Hobby/TestMaker         just now
    watermelon           Creative pipeline config                          1 week ago
  ────────────────────────────────────────────────────────────────────────
    ⟳  Scan & import
    +  Create new preset

  Enter load  s save  r rename  d delete  u undo  p preview  q quit
```

Your presets are the main menu. Use hotkeys to act on the highlighted preset:

| Key | Action |
|-----|--------|
| `Enter` | Load the preset (preview first, then confirm) |
| `s` | Save current project state to this preset |
| `r` | Rename this preset |
| `d` | Delete this preset (type name to confirm) |
| `p` | Preview this preset (skills, tools, permissions) |
| `u` | Undo last load (works from anywhere) |
| `↑↓` / `j/k` | Navigate |
| `q` / `Esc` | Quit |

After loading a preset, CCGears launches `claude` as a subprocess. CCGears stays alive in the background — when Claude exits, CCGears can reopen for another preset switch.

### Non-interactive mode

For scripting and shell aliases:

```bash
# Load a preset directly
ccgears load infra-tools

# Save current state to the active preset
ccgears save

# Rename a preset
ccgears rename old-name new-name

# List all presets
ccgears list

# List as JSON (for scripts)
ccgears list --json

# Show help
ccgears help
```

### Claude Code slash commands

Two slash commands are available inside any Claude Code session:

| Command | What it does |
|---------|-------------|
| `/ccgears` | Switches presets. Auto-exits Claude, opens CCGears TUI, resumes the session after loading a new preset. |
| `/ccgears-save` | Saves current `.claude/` and `tools/` state back to the active preset. Runs inline — no session exit needed. |

These commands auto-execute without permission prompts. The `/ccgears` and `/ccgears-save` skills are automatically injected into every loaded preset, so they persist across switches.

### Shell alias examples

```bash
# Add to ~/.zshrc
alias cc-infra="cd ~/projects/infra && ccgears load infra-tools && claude"
alias cc-creative="cd ~/projects/creative && ccgears load watermelon && claude"
```

---

## The CCGears Loop

When launched interactively, CCGears acts as a **parent process** that manages Claude Code sessions:

```
                    ┌──────────────────────┐
                    │   CCGears TUI        │
                    │   Pick & load preset │
                    └──────┬───────────────┘
                           │ Enter
                           ▼
                    ┌──────────────────────┐
                    │   Claude Code        │
                    │   (subprocess)       │
                    │                      │
                    │   /ccgears ──────────┤──▶ touches marker
                    │   auto-exits         │    & kills claude
                    └──────┬───────────────┘
                           │
                    CCGears detects marker
                           │
                           ▼
                    ┌──────────────────────┐
                    │   CCGears TUI        │
                    │   Pick new preset    │
                    └──────┬───────────────┘
                           │ Enter
                           ▼
                    ┌──────────────────────┐
                    │   Claude Code        │
                    │   --resume <session> │
                    │   (same conversation)│
                    └──────────────────────┘
```

Key behaviors:
- **CCGears stays alive** while Claude runs as a subprocess
- **Session continuity** — after switching, `claude --resume <session-id>` picks up the exact conversation
- **Clean exit** — if you `/exit` Claude normally (without `/ccgears`), CCGears exits too
- **SIGINT isolation** — CCGears ignores Ctrl+C signals from the `/ccgears` skill so it survives

---

## Preset Actions

### Load (Enter)

1. Highlight a preset and press Enter
2. Preview its contents — skills, permissions, tools, CLAUDE.md headline
3. Press Enter to confirm — current `.claude/` and `tools/` are backed up, then replaced
4. Press Enter to launch `claude`, or q to return to the menu

### Save (s)

Hover over a preset and press `s` to save the current project state to it. Shows a diff summary before confirming:

```
  Changes to infra-tools:

    + tools/new_script.py              (added)
    ~ .claude/settings.local.json      (modified)
    - tools/old_script.py              (removed)

    2 unchanged
```

### Rename (r)

Hover over a preset and press `r` to rename it. Enter the new name (lowercase, hyphens, max 48 chars).

### Delete (d)

Hover over a preset and press `d` to delete it. Type the preset name to confirm.

### Preview (p)

Hover over a preset and press `p` to view its contents without loading.

### Undo (u)

Press `u` from anywhere to undo the last load. Restores the previous `.claude/` and `tools/` state. Only one undo level available.

### Create new preset (+)

Select **Create new preset +** at the bottom of the list. Enter a name and optional description. Captures current `.claude/`, `tools/`, and `CLAUDE.md`.

**Preset name rules:** lowercase letters, digits, and hyphens only. Max 48 characters. No leading or trailing hyphens.

### Scan & import

Select **Scan & import** at the bottom of the list:

1. Enter a root directory to scan (default: home directory)
2. CCGears walks up to 4 levels deep, skipping `node_modules`, `.git`, `.venv`, `vendor`, etc.
3. Multi-select which projects to import (Space to toggle, `a` for all, `n` for none)
4. Creates presets automatically with names derived from directory names

---

## Preset Storage

All data is stored locally in `~/.ccgears/`:

```
~/.ccgears/
├── config.json                  # Settings (active preset, default preset, store path)
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
4. Injects `/ccgears` and `/ccgears-save` skills + auto-approve permissions
5. Tracks the active preset in `config.json`
6. Updates the preset's "last used" timestamp

**Create/Save** copies files from your project into the preset store:
- Symlinks are dereferenced (target content is copied, not the link)
- `.DS_Store` and `__pycache__` are excluded automatically

**Undo** reverses the last load by restoring from the backup slot.

**Switch** (`ccgears switch`, called by `/ccgears` skill):
1. Creates a marker file at `/tmp/.ccgears-switch`
2. Lists available presets
3. Sends SIGINT to the Claude Code process (graceful exit, session saved)
4. CCGears parent process detects the marker, reopens TUI
5. After loading a new preset, launches `claude --resume <session-id>`

---

## Development

### Project structure

```
CCGears/
├── cmd/ccgears/main.go          # CLI entry point, subprocess management
├── internal/
│   ├── config/                   # Config loading, path resolution
│   ├── preset/                   # CRUD operations, scan, diff, skill injection
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

## CLI Reference

| Command | Description |
|---------|-------------|
| `ccgears` | Interactive TUI (launches claude after loading) |
| `ccgears load <name>` | Load a preset non-interactively |
| `ccgears save` | Save current state to the active preset |
| `ccgears rename <old> <new>` | Rename a preset |
| `ccgears list` | List all presets |
| `ccgears list --json` | List presets as JSON |
| `ccgears switch` | Internal: used by `/ccgears` skill to trigger preset switch |
| `ccgears version` | Show version |
| `ccgears help` | Show help |

---

## Limitations

- **Single backup slot** — Only the last load can be undone. Loading twice overwrites the first backup.
- **Preset names are global** — Names must be unique across all projects. Use descriptive names like `infra-frigate` instead of just `default`.
- **Large workspaces** — Skill evaluation workspaces (images, videos) are included in snapshots. Presets for projects with large eval artifacts will be correspondingly large.
- **Session resume** — The `/ccgears` flow finds the most recent session file by modification time. If multiple Claude sessions are active in the same directory, it may resume the wrong one.

---

## License

MIT
