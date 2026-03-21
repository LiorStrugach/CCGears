package preset

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mryan/ccgears/internal/config"
	"github.com/mryan/ccgears/internal/snapshot"
	"github.com/mryan/ccgears/internal/validate"
)

var (
	ErrNotFound      = errors.New("preset not found")
	ErrAlreadyExists = errors.New("preset already exists")
	ErrEmptyProject  = errors.New("no .claude/, tools/, or CLAUDE.md found in project directory")
)

const metaFileName = "preset.json"

// Meta holds preset metadata.
type Meta struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	Created       string `json:"created"`
	LastUsed      string `json:"last_used"`
	SourceProject string `json:"source_project"`
}

// Create captures the current project's .claude/, tools/, and CLAUDE.md into a new preset.
func Create(cfg *config.Config, name, description, projectDir string) (*Meta, int, error) {
	if err := validate.PresetName(name); err != nil {
		return nil, 0, err
	}

	presetDir := filepath.Join(cfg.StorePath, name)
	if _, err := os.Stat(presetDir); err == nil {
		return nil, 0, fmt.Errorf("%w: %s", ErrAlreadyExists, name)
	}

	// Check at least one source exists
	claudeDir := filepath.Join(projectDir, ".claude")
	toolsDir := filepath.Join(projectDir, "tools")
	claudeMD := filepath.Join(projectDir, "CLAUDE.md")

	hasContent := validate.DirExists(claudeDir) || validate.DirExists(toolsDir) || validate.FileExists(claudeMD)
	if !hasContent {
		return nil, 0, ErrEmptyProject
	}

	if err := os.MkdirAll(presetDir, 0755); err != nil {
		return nil, 0, fmt.Errorf("creating preset directory: %w", err)
	}

	totalFiles := 0

	// Copy .claude/ -> claude/ (no dot)
	if validate.DirExists(claudeDir) {
		n, err := snapshot.CopyDir(claudeDir, filepath.Join(presetDir, "claude"), snapshot.DefaultExcludes)
		if err != nil {
			os.RemoveAll(presetDir)
			return nil, 0, fmt.Errorf("copying .claude/: %w", err)
		}
		totalFiles += n
	}

	// Copy tools/ -> tools/
	if validate.DirExists(toolsDir) {
		n, err := snapshot.CopyDir(toolsDir, filepath.Join(presetDir, "tools"), snapshot.DefaultExcludes)
		if err != nil {
			os.RemoveAll(presetDir)
			return nil, 0, fmt.Errorf("copying tools/: %w", err)
		}
		totalFiles += n
	}

	// Copy CLAUDE.md
	if validate.FileExists(claudeMD) {
		data, err := os.ReadFile(claudeMD)
		if err == nil {
			os.WriteFile(filepath.Join(presetDir, "CLAUDE.md"), data, 0644)
			totalFiles++
		}
	}

	now := time.Now().Format(time.RFC3339)
	meta := &Meta{
		Name:          name,
		Description:   description,
		Created:       now,
		LastUsed:      now,
		SourceProject: projectDir,
	}
	if err := writeMeta(presetDir, meta); err != nil {
		os.RemoveAll(presetDir)
		return nil, 0, err
	}

	return meta, totalFiles, nil
}

// Get reads metadata for a single preset. Returns ErrNotFound if it doesn't exist.
func Get(cfg *config.Config, name string) (*Meta, error) {
	presetDir := filepath.Join(cfg.StorePath, name)
	if _, err := os.Stat(presetDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return readMeta(presetDir)
}

// Delete removes a preset from the store.
func Delete(cfg *config.Config, name string) error {
	presetDir := filepath.Join(cfg.StorePath, name)
	if _, err := os.Stat(presetDir); os.IsNotExist(err) {
		return fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return os.RemoveAll(presetDir)
}

// Dir returns the filesystem path for a preset.
func Dir(cfg *config.Config, name string) string {
	return filepath.Join(cfg.StorePath, name)
}

func writeMeta(presetDir string, meta *Meta) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(presetDir, metaFileName)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func readMeta(presetDir string) (*Meta, error) {
	data, err := os.ReadFile(filepath.Join(presetDir, metaFileName))
	if err != nil {
		// Return partial meta with just the name from the directory
		return &Meta{Name: filepath.Base(presetDir)}, nil
	}
	var meta Meta
	if err := json.Unmarshal(data, &meta); err != nil {
		return &Meta{Name: filepath.Base(presetDir)}, nil
	}
	return &meta, nil
}

// Load restores a preset's .claude/, tools/, and CLAUDE.md into projectDir.
// Automatically backs up the current state first.
func Load(cfg *config.Config, name, projectDir string) (int, error) {
	presetDir := filepath.Join(cfg.StorePath, name)
	if _, err := os.Stat(presetDir); os.IsNotExist(err) {
		return 0, fmt.Errorf("%w: %s", ErrNotFound, name)
	}

	// Backup current state
	if err := snapshot.CreateBackup(cfg.BackupPath, projectDir, name); err != nil {
		return 0, fmt.Errorf("creating backup: %w", err)
	}

	// Remove current state
	snapshot.RemoveIfExists(filepath.Join(projectDir, ".claude"))
	snapshot.RemoveIfExists(filepath.Join(projectDir, "tools"))

	totalFiles := 0

	// Copy claude/ -> .claude/
	claudeSrc := filepath.Join(presetDir, "claude")
	if validate.DirExists(claudeSrc) {
		n, err := snapshot.CopyDir(claudeSrc, filepath.Join(projectDir, ".claude"), nil)
		if err != nil {
			return 0, fmt.Errorf("loading .claude/: %w", err)
		}
		totalFiles += n
	}

	// Copy tools/ -> tools/
	toolsSrc := filepath.Join(presetDir, "tools")
	if validate.DirExists(toolsSrc) {
		n, err := snapshot.CopyDir(toolsSrc, filepath.Join(projectDir, "tools"), nil)
		if err != nil {
			return 0, fmt.Errorf("loading tools/: %w", err)
		}
		totalFiles += n
	}

	// Copy CLAUDE.md
	claudeMDSrc := filepath.Join(presetDir, "CLAUDE.md")
	if validate.FileExists(claudeMDSrc) {
		snapshot.RemoveIfExists(filepath.Join(projectDir, "CLAUDE.md"))
		data, err := os.ReadFile(claudeMDSrc)
		if err == nil {
			os.WriteFile(filepath.Join(projectDir, "CLAUDE.md"), data, 0644)
			totalFiles++
		}
	}

	// Ensure ccgears skill persists across preset switches
	ensureCCGearsSkill(filepath.Join(projectDir, ".claude"))

	// Update last_used
	updateLastUsed(cfg, name)

	return totalFiles, nil
}

// Save updates an existing preset from the current project state.
// Returns the number of files written.
func Save(cfg *config.Config, name, projectDir string) (int, error) {
	presetDir := filepath.Join(cfg.StorePath, name)
	if _, err := os.Stat(presetDir); os.IsNotExist(err) {
		return 0, fmt.Errorf("%w: %s", ErrNotFound, name)
	}

	totalFiles := 0

	// Replace claude/
	claudeDst := filepath.Join(presetDir, "claude")
	claudeSrc := filepath.Join(projectDir, ".claude")
	snapshot.RemoveIfExists(claudeDst)
	if validate.DirExists(claudeSrc) {
		n, err := snapshot.CopyDir(claudeSrc, claudeDst, snapshot.DefaultExcludes)
		if err != nil {
			return 0, fmt.Errorf("saving .claude/: %w", err)
		}
		totalFiles += n
	}

	// Replace tools/
	toolsDst := filepath.Join(presetDir, "tools")
	toolsSrc := filepath.Join(projectDir, "tools")
	snapshot.RemoveIfExists(toolsDst)
	if validate.DirExists(toolsSrc) {
		n, err := snapshot.CopyDir(toolsSrc, toolsDst, snapshot.DefaultExcludes)
		if err != nil {
			return 0, fmt.Errorf("saving tools/: %w", err)
		}
		totalFiles += n
	}

	// Replace CLAUDE.md
	claudeMDDst := filepath.Join(presetDir, "CLAUDE.md")
	claudeMDSrc := filepath.Join(projectDir, "CLAUDE.md")
	snapshot.RemoveIfExists(claudeMDDst)
	if validate.FileExists(claudeMDSrc) {
		data, err := os.ReadFile(claudeMDSrc)
		if err == nil {
			os.WriteFile(claudeMDDst, data, 0644)
			totalFiles++
		}
	}

	// Update metadata timestamp
	meta, _ := readMeta(presetDir)
	meta.LastUsed = time.Now().Format(time.RFC3339)
	meta.SourceProject = projectDir
	writeMeta(presetDir, meta)

	return totalFiles, nil
}

func updateLastUsed(cfg *config.Config, name string) error {
	meta, err := Get(cfg, name)
	if err != nil {
		return err
	}
	meta.LastUsed = time.Now().Format(time.RFC3339)
	return writeMeta(filepath.Join(cfg.StorePath, name), meta)
}

// ccgearsSkillContent is the embedded SKILL.md for the /ccgears slash command.
// This gets injected into every loaded preset so the command persists across switches.
const ccgearsSkillContent = `---
name: ccgears
description: >
  Switch to the CCGears preset manager. Automatically exits this session
  and opens CCGears. Session resumes after preset switch.
allowed-tools: Bash(touch *), Bash(export *), Bash(ccgears *), Bash(kill *)
---

# CCGears — Switch Presets

When this skill is invoked, IMMEDIATELY run this single Bash command.
Do NOT ask questions. Do NOT present options. Just execute this command:

` + "```" + `bash
touch /tmp/.ccgears-switch && export PATH="$PATH:$HOME/go/bin" && ccgears list && kill -INT $PPID
` + "```" + `

This will list the available presets, then exit this session. CCGears will
open automatically and this session will resume after the preset switch.
`

// ensureCCGearsSkill writes the ccgears skill into .claude/skills/ccgears/
// so the /ccgears command persists after every preset load.
func ensureCCGearsSkill(claudeDir string) {
	skillDir := filepath.Join(claudeDir, "skills", "ccgears")
	skillFile := filepath.Join(skillDir, "SKILL.md")

	os.MkdirAll(skillDir, 0755)
	os.WriteFile(skillFile, []byte(ccgearsSkillContent), 0644)
}
