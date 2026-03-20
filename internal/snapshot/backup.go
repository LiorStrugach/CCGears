package snapshot

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mryan/ccgears/internal/validate"
)

var ErrNoBackup = errors.New("no backup available")

const backupMetaFile = "backup_meta.json"

// BackupMeta records what was backed up and where it came from.
type BackupMeta struct {
	ProjectDir   string `json:"project_dir"`
	PresetLoaded string `json:"preset_loaded"`
	Timestamp    string `json:"timestamp"`
	HadClaude    bool   `json:"had_claude"`
	HadTools     bool   `json:"had_tools"`
	HadClaudeMD  bool   `json:"had_claude_md"`
}

// CreateBackup snapshots the current .claude/, tools/, and CLAUDE.md from projectDir
// into backupPath. Overwrites any previous backup.
func CreateBackup(backupPath, projectDir, presetName string) error {
	// Clear previous backup
	os.RemoveAll(backupPath)
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return fmt.Errorf("creating backup directory: %w", err)
	}

	claudeDir := filepath.Join(projectDir, ".claude")
	toolsDir := filepath.Join(projectDir, "tools")
	claudeMD := filepath.Join(projectDir, "CLAUDE.md")

	meta := BackupMeta{
		ProjectDir:   projectDir,
		PresetLoaded: presetName,
		Timestamp:    time.Now().Format(time.RFC3339),
		HadClaude:    validate.DirExists(claudeDir),
		HadTools:     validate.DirExists(toolsDir),
		HadClaudeMD:  validate.FileExists(claudeMD),
	}

	if meta.HadClaude {
		if _, err := CopyDir(claudeDir, filepath.Join(backupPath, "claude"), DefaultExcludes); err != nil {
			return fmt.Errorf("backing up .claude/: %w", err)
		}
	}
	if meta.HadTools {
		if _, err := CopyDir(toolsDir, filepath.Join(backupPath, "tools"), DefaultExcludes); err != nil {
			return fmt.Errorf("backing up tools/: %w", err)
		}
	}
	if meta.HadClaudeMD {
		data, err := os.ReadFile(claudeMD)
		if err == nil {
			os.WriteFile(filepath.Join(backupPath, "CLAUDE.md"), data, 0644)
		}
	}

	// Write metadata
	data, _ := json.MarshalIndent(meta, "", "  ")
	return os.WriteFile(filepath.Join(backupPath, backupMetaFile), data, 0644)
}

// RestoreBackup undoes the last load by restoring backed-up files to the original project directory.
func RestoreBackup(backupPath string) (string, error) {
	meta, err := GetBackupInfo(backupPath)
	if err != nil {
		return "", err
	}

	projectDir := meta.ProjectDir

	// Remove current state
	RemoveIfExists(filepath.Join(projectDir, ".claude"))
	RemoveIfExists(filepath.Join(projectDir, "tools"))
	RemoveIfExists(filepath.Join(projectDir, "CLAUDE.md"))

	// Restore what was backed up
	if meta.HadClaude {
		src := filepath.Join(backupPath, "claude")
		if validate.DirExists(src) {
			if _, err := CopyDir(src, filepath.Join(projectDir, ".claude"), nil); err != nil {
				return projectDir, fmt.Errorf("restoring .claude/: %w", err)
			}
		}
	}
	if meta.HadTools {
		src := filepath.Join(backupPath, "tools")
		if validate.DirExists(src) {
			if _, err := CopyDir(src, filepath.Join(projectDir, "tools"), nil); err != nil {
				return projectDir, fmt.Errorf("restoring tools/: %w", err)
			}
		}
	}
	if meta.HadClaudeMD {
		src := filepath.Join(backupPath, "CLAUDE.md")
		if data, err := os.ReadFile(src); err == nil {
			os.WriteFile(filepath.Join(projectDir, "CLAUDE.md"), data, 0644)
		}
	}

	// Clear backup after successful restore
	os.RemoveAll(backupPath)

	return projectDir, nil
}

// HasBackup returns true if a backup exists.
func HasBackup(backupPath string) bool {
	return validate.FileExists(filepath.Join(backupPath, backupMetaFile))
}

// GetBackupInfo reads backup metadata.
func GetBackupInfo(backupPath string) (*BackupMeta, error) {
	data, err := os.ReadFile(filepath.Join(backupPath, backupMetaFile))
	if err != nil {
		return nil, ErrNoBackup
	}
	var meta BackupMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("corrupt backup metadata: %w", err)
	}
	return &meta, nil
}
