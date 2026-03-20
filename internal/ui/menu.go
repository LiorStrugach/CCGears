package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mryan/ccgears/internal/config"
	"github.com/mryan/ccgears/internal/preset"
	"github.com/mryan/ccgears/internal/preview"
	"github.com/mryan/ccgears/internal/snapshot"
)

const menuWidth = 40

// RunInteractiveMenu runs the main CCGears menu loop.
func RunInteractiveMenu(cfg *config.Config, projectDir string) error {
	if err := cfg.EnsureDirs(); err != nil {
		return err
	}

	// Enter raw mode for the entire session
	if err := EnableRawMode(); err != nil {
		return fmt.Errorf("failed to enable raw mode: %w", err)
	}
	defer DisableRawMode()

	menuLabels := []string{
		"Load preset",
		"Save preset",
		"Create preset",
		"Scan & import",
		"Delete preset",
		"Undo last load",
		"Exit",
	}

	lastSelected := 0

	for {
		// Clear screen
		if useColor {
			printf("\033[2J\033[H")
		}

		printf(Banner())
		println()
		printf("  %s\n", dim("Use ↑↓ arrows to navigate, Enter to select, q to quit"))
		println()

		choice := SelectMenuBoxed(menuLabels, menuWidth)
		if choice < 0 {
			choice = 6 // treat cancel as Exit
		}
		lastSelected = choice

		printf("\n")

		switch choice {
		case 0:
			flowLoad(cfg, projectDir)
		case 1:
			flowSave(cfg, projectDir)
		case 2:
			flowCreate(cfg, projectDir)
		case 3:
			flowScan(cfg)
		case 4:
			flowDelete(cfg)
		case 5:
			flowUndo(cfg)
		case 6:
			// Clear screen on exit
			if useColor {
				printf("\033[2J\033[H")
			}
			return nil
		}

		_ = lastSelected
		if choice != 6 {
			println()
			ReadLine(dim("Press Enter to continue...") + " ")
		}
	}
}

func flowScan(cfg *config.Config) {
	PrintHeader("Scan & Import")
	println()

	home, _ := os.UserHomeDir()
	defaultRoot := filepath.Dir(home)
	if home != "" {
		defaultRoot = home
	}

	rootDir := PromptText("Root directory to scan", defaultRoot)
	if rootDir == "" {
		Warn("Cancelled.")
		return
	}

	// Expand ~ if needed
	if strings.HasPrefix(rootDir, "~/") {
		rootDir = filepath.Join(home, rootDir[2:])
	} else if rootDir == "~" {
		rootDir = home
	}

	printf("\n  %s\n", dim("Scanning..."))

	results, err := preset.ScanForProjects(rootDir, 4)
	if err != nil {
		Error(fmt.Sprintf("Scan failed: %v", err))
		return
	}

	if len(results) == 0 {
		Warn("No projects with Claude Code configs found.")
		return
	}

	printf("  Found %s projects with Claude Code configs.\n\n", cyanBold(fmt.Sprintf("%d", len(results))))
	printf("  %s\n\n", dim("↑↓ navigate, Space toggle, a=all, n=none, Enter confirm, q=cancel"))

	// Build display items and default checked state
	items := make([]string, len(results))
	checked := make([]bool, len(results))
	for i, r := range results {
		shortPath := preset.ShortenPath(r.ProjectDir)

		// Build info line
		var parts []string
		if r.SkillCount > 0 {
			parts = append(parts, fmt.Sprintf("%d skills", r.SkillCount))
		}
		if r.ToolCount > 0 {
			parts = append(parts, fmt.Sprintf("%d tools", r.ToolCount))
		}
		if r.HasClaudeMD {
			parts = append(parts, "CLAUDE.md")
		}
		info := ""
		if len(parts) > 0 {
			info = "  " + dim(strings.Join(parts, ", "))
		}

		items[i] = fmt.Sprintf("%-18s %s%s", magBold(r.DirName), dim(shortPath), info)
		checked[i] = true // default all selected
	}

	selected := SelectMulti(items, checked)
	if selected == nil || len(selected) == 0 {
		Warn("No projects selected.")
		return
	}

	println()

	// Import selected projects
	imported := 0
	skipped := 0
	for _, idx := range selected {
		r := results[idx]
		baseName := preset.AutoPresetName(r.DirName)
		name := preset.UniquePresetName(cfg, baseName)

		// Check if the base name already existed (UniquePresetName changed it)
		if name != baseName {
			// Check if the original exists — it's a true collision
			if _, err := preset.Get(cfg, baseName); err == nil {
				printf("  %s %s already exists, skipping\n", yellowBold("⚠"), magBold(baseName))
				skipped++
				continue
			}
		}

		desc := fmt.Sprintf("Imported from %s", preset.ShortenPath(r.ProjectDir))
		_, count, err := preset.Create(cfg, name, desc, r.ProjectDir)
		if err != nil {
			printf("  %s %s: %v\n", redBold("✗"), r.DirName, err)
			skipped++
			continue
		}
		printf("  %s %s %s\n", greenBold("✓"), magBold(name), dim(fmt.Sprintf("(%d files)", count)))
		imported++
	}

	println()
	if imported > 0 {
		Success(fmt.Sprintf("Imported %d of %d presets.", imported, len(selected)))
	}
	if skipped > 0 {
		printf("  %s\n", dim(fmt.Sprintf("%d skipped.", skipped)))
	}
}

func flowCreate(cfg *config.Config, projectDir string) {
	PrintHeader("Create Preset")
	println()

	name := PromptText("Preset name "+dim("(lowercase, hyphens, max 48 chars)"), "")
	if name == "" {
		Warn("Cancelled.")
		return
	}

	description := PromptText("Description "+dim("(optional)"), "")

	// Show what will be captured
	claudeDir := fmt.Sprintf("%s/.claude", projectDir)
	toolsDir := fmt.Sprintf("%s/tools", projectDir)
	claudeMD := fmt.Sprintf("%s/CLAUDE.md", projectDir)

	println()
	println("  " + yellowBold("Will capture:"))
	if fi, err := os.Stat(claudeDir); err == nil && fi.IsDir() {
		count, size := snapshot.CountFiles(claudeDir, snapshot.DefaultExcludes)
		printf("    %s  %s\n", cyan(".claude/"), dim(fmt.Sprintf("(%d files, %s)", count, FormatSize(size))))
	}
	if fi, err := os.Stat(toolsDir); err == nil && fi.IsDir() {
		count, size := snapshot.CountFiles(toolsDir, snapshot.DefaultExcludes)
		printf("    %s  %s\n", cyan("tools/  "), dim(fmt.Sprintf("(%d files, %s)", count, FormatSize(size))))
	}
	if fi, err := os.Stat(claudeMD); err == nil && !fi.IsDir() {
		printf("    %s  %s\n", cyan("CLAUDE.md"), dim("(present)"))
	}
	println()

	if !Confirm(fmt.Sprintf("Create preset %s?", magBold(name))) {
		Warn("Cancelled.")
		return
	}

	meta, count, err := preset.Create(cfg, name, description, projectDir)
	if err != nil {
		Error(fmt.Sprintf("%v", err))
		return
	}
	println()
	Success(fmt.Sprintf("Preset %s created %s", magBold(meta.Name), dim(fmt.Sprintf("(%d files)", count))))
}

func flowLoad(cfg *config.Config, projectDir string) {
	PrintHeader("Load Preset")

	name := selectPreset(cfg, "Load")
	if name == "" {
		return
	}

	// Show preview
	presetDir := preset.Dir(cfg, name)
	p := preview.Build(presetDir)
	printPreview(name, p)
	println()

	if !Confirm(fmt.Sprintf("Load %s? Current state will be backed up.", magBold(name))) {
		Warn("Cancelled.")
		return
	}

	count, err := preset.Load(cfg, name, projectDir)
	if err != nil {
		Error(fmt.Sprintf("%v", err))
		return
	}
	println()
	Success(fmt.Sprintf("Loaded %s %s", magBold(name), dim(fmt.Sprintf("(%d files)", count))))
	printf("  Run %s to start a new session.\n", cyanBold("claude"))
}

func flowSave(cfg *config.Config, projectDir string) {
	PrintHeader("Save Preset")

	name := selectPreset(cfg, "Save to")
	if name == "" {
		return
	}

	// Compute diff
	presetDir := preset.Dir(cfg, name)
	diff := preset.ComputeDiff(presetDir, projectDir)

	if !diff.HasChanges() {
		Warn("No changes detected. Nothing to save.")
		return
	}

	println()
	printDiffSummary(name, diff)
	println()

	if !Confirm(fmt.Sprintf("Save these changes to %s?", magBold(name))) {
		Warn("Cancelled.")
		return
	}

	count, err := preset.Save(cfg, name, projectDir)
	if err != nil {
		Error(fmt.Sprintf("%v", err))
		return
	}
	println()
	Success(fmt.Sprintf("Preset %s saved %s", magBold(name), dim(fmt.Sprintf("(%d files)", count))))
}

func flowDelete(cfg *config.Config) {
	PrintHeader("Delete Preset")

	name := selectPreset(cfg, "Delete")
	if name == "" {
		return
	}

	meta, err := preset.Get(cfg, name)
	if err != nil {
		Error(fmt.Sprintf("%v", err))
		return
	}

	println()
	InfoLine("Name:       ", magBold(meta.Name))
	if meta.Description != "" {
		InfoLine("Description:", meta.Description)
	}
	InfoLine("Created:    ", dim(FormatDate(meta.Created)))
	println()
	printf("  %s\n", redBold("  This cannot be undone."))
	println()

	if !ConfirmByName(name) {
		Warn("Name does not match. Delete cancelled.")
		return
	}

	if err := preset.Delete(cfg, name); err != nil {
		Error(fmt.Sprintf("%v", err))
		return
	}
	println()
	Success(fmt.Sprintf("Preset %s deleted.", magBold(name)))
}

func flowUndo(cfg *config.Config) {
	if !snapshot.HasBackup(cfg.BackupPath) {
		Warn("No backup available. Nothing to undo.")
		return
	}

	PrintHeader("Undo Last Load")

	info, err := snapshot.GetBackupInfo(cfg.BackupPath)
	if err != nil {
		Error(fmt.Sprintf("%v", err))
		return
	}

	println()
	InfoLine("Preset:   ", magBold(info.PresetLoaded))
	InfoLine("Project:  ", info.ProjectDir)
	InfoLine("Timestamp:", dim(FormatDate(info.Timestamp)))
	println()

	if !Confirm("Undo this load and restore previous state?") {
		Warn("Cancelled.")
		return
	}

	dir, err := snapshot.RestoreBackup(cfg.BackupPath)
	if err != nil {
		Error(fmt.Sprintf("%v", err))
		return
	}
	println()
	Success(fmt.Sprintf("Previous state restored to %s", dir))
}

// selectPreset lists presets and lets the user pick with arrow keys.
func selectPreset(cfg *config.Config, action string) string {
	presets, err := preset.List(cfg)
	if err != nil {
		Error(fmt.Sprintf("%v", err))
		return ""
	}
	if len(presets) == 0 {
		Warn("No presets found. Create one first.")
		return ""
	}

	println()
	printf("  %s\n", dim("Use ↑↓ to select, Enter to confirm, q to cancel"))
	println()

	items := make([]string, len(presets))
	for i, p := range presets {
		desc := dim("(no description)")
		if p.Description != "" {
			desc = p.Description
		}
		lastUsed := ""
		if p.LastUsed != "" {
			lastUsed = "  " + dim(FormatDate(p.LastUsed))
		}
		items[i] = fmt.Sprintf("%-18s %s%s", magBold(p.Name), desc, lastUsed)
	}

	idx := SelectMenu(items, 0)
	if idx < 0 {
		return ""
	}
	return presets[idx].Name
}

func printPreview(name string, p *preview.PresetPreview) {
	const w = 48
	println()
	printf("  %s\n", BoxTop(magBold("Preview: "+name), w))
	printf("  %s\n", BoxEmpty(w))

	if p.Headline != "" {
		content := fmt.Sprintf("  %s  %s", cyan("CLAUDE.md"), cyanItalic("\""+p.Headline+"\""))
		visLen := 2 + 9 + 2 + len(p.Headline) + 2
		printf("  %s\n", BoxLine(content, w, visLen))
		printf("  %s\n", BoxEmpty(w))
	}

	if len(p.Skills) > 0 {
		header := fmt.Sprintf("  %s", yellowBold(fmt.Sprintf("Skills (%d)", len(p.Skills))))
		visLen := 2 + len(fmt.Sprintf("Skills (%d)", len(p.Skills)))
		printf("  %s\n", BoxLine(header, w, visLen))
		for _, s := range p.Skills {
			line := fmt.Sprintf("    %s %s", magenta("●"), bold(s.Name))
			visLen := 4 + 1 + 1 + len(s.Name)
			if s.Description != "" {
				short := s.Description
				maxDesc := w - visLen - 5
				if maxDesc > 0 && len(short) > maxDesc {
					short = short[:maxDesc-3] + "..."
				}
				line += "  " + dim(short)
				visLen += 2 + len(short)
			}
			printf("  %s\n", BoxLine(line, w, visLen))
		}
		printf("  %s\n", BoxEmpty(w))
	}

	hasPerm := p.BashPerms > 0 || len(p.MCPTools) > 0 || len(p.WebPerms) > 0
	if hasPerm {
		header := fmt.Sprintf("  %s", yellowBold("Permissions"))
		visLen := 2 + 11
		printf("  %s\n", BoxLine(header, w, visLen))
		if len(p.MCPTools) > 0 {
			line := fmt.Sprintf("    %s %s %d", blue("◆"), cyan("MCP tools:"), len(p.MCPTools))
			visLen := 4 + 1 + 1 + 10 + 1 + len(fmt.Sprintf("%d", len(p.MCPTools)))
			printf("  %s\n", BoxLine(line, w, visLen))
		}
		if p.BashPerms > 0 {
			line := fmt.Sprintf("    %s %s %d patterns", blue("◆"), cyan("Bash:     "), p.BashPerms)
			visLen := 4 + 1 + 1 + 10 + 1 + len(fmt.Sprintf("%d patterns", p.BashPerms))
			printf("  %s\n", BoxLine(line, w, visLen))
		}
		if len(p.WebPerms) > 0 {
			webStr := strings.Join(p.WebPerms, ", ")
			if len(webStr) > 24 {
				webStr = webStr[:21] + "..."
			}
			line := fmt.Sprintf("    %s %s %s", blue("◆"), cyan("Web:      "), webStr)
			visLen := 4 + 1 + 1 + 10 + 1 + len(webStr)
			printf("  %s\n", BoxLine(line, w, visLen))
		}
		printf("  %s\n", BoxEmpty(w))
	}

	if len(p.ToolScripts) > 0 {
		header := fmt.Sprintf("  %s", yellowBold(fmt.Sprintf("Tools (%d)", len(p.ToolScripts))))
		visLen := 2 + len(fmt.Sprintf("Tools (%d)", len(p.ToolScripts)))
		printf("  %s\n", BoxLine(header, w, visLen))
		scriptStr := strings.Join(p.ToolScripts, ", ")
		if len(scriptStr) > w-8 {
			scriptStr = scriptStr[:w-11] + "..."
		}
		line := "    " + scriptStr
		printf("  %s\n", BoxLine(line, w, 4+len(scriptStr)))
		printf("  %s\n", BoxEmpty(w))
	}

	totalLine := fmt.Sprintf("  %s", dim(fmt.Sprintf("Total: %d files, %s", p.TotalFiles, FormatSize(p.TotalSize))))
	totalVisLen := 2 + len(fmt.Sprintf("Total: %d files, %s", p.TotalFiles, FormatSize(p.TotalSize)))
	printf("  %s\n", BoxLine(totalLine, w, totalVisLen))
	printf("  %s\n", BoxEmpty(w))
	printf("  %s\n", BoxBottom(w))
}

func printDiffSummary(name string, diff *preset.DiffSummary) {
	printf("  %s %s\n", yellowBold("Changes to"), magBold(name)+":")
	println()
	for _, f := range diff.Added {
		printf("    %s %-40s %s\n", green("+"), f, dim("(added)"))
	}
	for _, f := range diff.Modified {
		printf("    %s %-40s %s\n", yellow("~"), f, dim("(modified)"))
	}
	for _, f := range diff.Removed {
		printf("    %s %-40s %s\n", red("-"), f, dim("(removed)"))
	}
	if diff.Unchanged > 0 {
		println()
		printf("    %s\n", dim(fmt.Sprintf("%d unchanged", diff.Unchanged)))
	}
}

// RunList outputs presets for non-interactive use.
func RunList(cfg *config.Config, jsonOutput bool) error {
	if err := cfg.EnsureDirs(); err != nil {
		return err
	}

	presets, err := preset.List(cfg)
	if err != nil {
		return err
	}

	if jsonOutput {
		if presets == nil {
			presets = []*preset.Meta{}
		}
		data, err := json.MarshalIndent(presets, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	if len(presets) == 0 {
		fmt.Println("No presets found.")
		return nil
	}

	for _, p := range presets {
		desc := ""
		if p.Description != "" {
			desc = "  " + p.Description
		}
		lastUsed := ""
		if p.LastUsed != "" {
			lastUsed = fmt.Sprintf("  (%s)", FormatDate(p.LastUsed))
		}
		fmt.Printf("%-20s%s%s\n", p.Name, desc, lastUsed)
	}
	return nil
}

// RunLoadNonInteractive loads a preset without interactive prompts.
func RunLoadNonInteractive(cfg *config.Config, name, projectDir string) error {
	if err := cfg.EnsureDirs(); err != nil {
		return err
	}

	count, err := preset.Load(cfg, name, projectDir)
	if err != nil {
		return err
	}
	if useColor {
		fmt.Printf("%s Loaded preset %s (%d files). Previous state backed up.\n",
			greenBold("✓"), magBold(name), count)
	} else {
		fmt.Printf("Loaded preset '%s' (%d files). Previous state backed up.\n", name, count)
	}
	return nil
}
