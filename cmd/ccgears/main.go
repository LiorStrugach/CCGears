package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mryan/ccgears/internal/config"
	"github.com/mryan/ccgears/internal/ui"
)

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
		default:
			_, _ = fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", args[0])
			printUsage()
			os.Exit(1)
		}
	}

	// Interactive mode — Bubble Tea
	cfg := mustLoadConfig()
	if err := cfg.EnsureDirs(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	projectDir := mustGetwd()

	app := ui.NewApp(cfg, projectDir)
	p := tea.NewProgram(app, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Check if we should launch claude
	if m, ok := finalModel.(ui.App); ok && m.ShouldLaunchClaude() {
		launchClaude()
	}
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
    ccgears              Interactive menu
    ccgears load <name>  Load a preset (non-interactive)
    ccgears list         List all presets
    ccgears list --json  List presets as JSON
    ccgears version      Show version
    ccgears help         Show this help

  Presets are stored in ~/.ccgears/presets/
  Override with CCGEARS_HOME environment variable.
`)
}

func launchClaude() {
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		fmt.Printf("  'claude' not found in PATH. Start it manually.\n")
		return
	}
	env := os.Environ()
	if err := syscall.Exec(claudePath, []string{"claude"}, env); err != nil {
		fmt.Printf("  Failed to launch claude: %v. Start it manually.\n", err)
	}
}
