package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mryan/ccgears/internal/config"
	"github.com/mryan/ccgears/internal/ui"
)

const markerFile = "/tmp/.ccgears-switch"

func main() {
	args := os.Args[1:]

	if len(args) > 0 {
		switch args[0] {
		case "version", "--version":
			fmt.Printf("ccgears v%s\n", ui.GetVersion())
			return
		case "help", "--help", "-h":
			printUsage()
			return
		case "load":
			if len(args) < 2 {
				_, _ = fmt.Fprintln(os.Stderr, "Usage: ccgears load <preset-name>")
				os.Exit(1)
			}
			cfg := mustLoadConfig()
			projectDir := mustGetwd()
			if err := ui.RunLoadNonInteractive(cfg, args[1], projectDir); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "list":
			jsonOutput := len(args) > 1 && args[1] == "--json"
			cfg := mustLoadConfig()
			if err := ui.RunList(cfg, jsonOutput); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "switch":
			// Called by /ccgears skill inside Claude Code.
			// Creates marker, lists presets, finds and kills the claude process.
			_ = os.WriteFile(markerFile, []byte{}, 0644)
			cfg := mustLoadConfig()
			_ = cfg.EnsureDirs()
			_ = ui.RunList(cfg, false)
			fmt.Println("\nSwitching to CCGears...")
			killClaudeParent()
			return
		default:
			_, _ = fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", args[0])
			printUsage()
			os.Exit(1)
		}
	}

	// Interactive mode — CCGears ↔ Claude Code loop
	cfg := mustLoadConfig()
	if err := cfg.EnsureDirs(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	projectDir := mustGetwd()
	resume := false

	for {
		// Show CCGears TUI
		app := ui.NewApp(cfg, projectDir)
		p := tea.NewProgram(app, tea.WithAltScreen())

		finalModel, err := p.Run()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		m, ok := finalModel.(ui.App)
		if !ok || !m.ShouldLaunchClaude() {
			break // user chose Exit
		}

		// Launch claude as subprocess (CCGears stays alive)
		runClaudeSession(resume)

		// Claude exited. Check if /ccgears was invoked (marker file).
		if _, err := os.Stat(markerFile); err != nil {
			break // no marker → normal exit, CCGears exits too
		}

		// Marker exists → user wants to switch presets
		os.Remove(markerFile)
		resume = true // next claude launch will --resume
		// Loop back to TUI
	}
}

// runClaudeSession launches claude as a subprocess and waits for it to exit.
// CCGears stays alive as the parent process.
// SIGINT is ignored so that when /ccgears skill sends kill -INT to Claude,
// CCGears survives and can check the marker file.
func runClaudeSession(resume bool) {
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		fmt.Println("  'claude' not found in PATH. Install Claude Code first.")
		return
	}

	args := []string{}
	if resume {
		args = append(args, "--resume")
	}

	// Ignore SIGINT so CCGears survives when /ccgears skill kills Claude
	signal.Ignore(syscall.SIGINT)

	cmd := exec.Command(claudePath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	_ = cmd.Run() // blocks until claude exits

	// Restore SIGINT handling
	signal.Reset(syscall.SIGINT)
}

// killClaudeParent walks up the process tree from our PID to find the
// claude process, then sends it SIGINT. Only kills that specific process.
func killClaudeParent() {
	pid := os.Getpid()
	for {
		ppid := getParentPID(pid)
		if ppid <= 1 {
			break // reached init, stop
		}
		name := getProcessName(ppid)
		if name == "claude" || name == "node" {
			// Found it — send SIGINT to just this process
			syscall.Kill(ppid, syscall.SIGINT)
			return
		}
		pid = ppid
	}
	// Fallback: kill direct parent
	syscall.Kill(syscall.Getppid(), syscall.SIGINT)
}

// getParentPID reads the parent PID of a given PID via ps.
func getParentPID(pid int) int {
	out, err := exec.Command("ps", "-o", "ppid=", "-p", fmt.Sprintf("%d", pid)).Output()
	if err != nil {
		return 0
	}
	var ppid int
	_, _ = fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &ppid)
	return ppid
}

// getProcessName reads the process name of a given PID via ps.
func getProcessName(pid int) string {
	out, err := exec.Command("ps", "-o", "comm=", "-p", fmt.Sprintf("%d", pid)).Output()
	if err != nil {
		return ""
	}
	name := strings.TrimSpace(string(out))
	// ps returns full path on some systems, get just the basename
	return filepath.Base(name)
}

func mustLoadConfig() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
	return cfg
}

func mustGetwd() string {
	dir, err := os.Getwd()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
		os.Exit(1)
	}
	return dir
}

func printUsage() {
	fmt.Print(ui.Banner())
	fmt.Printf(`
  Usage:
    ccgears              Interactive menu (launches claude after loading)
    ccgears load <name>  Load a preset (non-interactive)
    ccgears list         List all presets
    ccgears list --json  List presets as JSON
    ccgears version      Show version
    ccgears help         Show this help

  Inside Claude Code, type /ccgears to switch presets seamlessly.
  CCGears will reopen automatically after Claude exits.

  Presets are stored in ~/.ccgears/presets/
  Override with CCGEARS_HOME environment variable.
`)
}
