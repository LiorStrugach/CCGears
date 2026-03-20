package snapshot

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateBackup_WritesMetaAndContents(t *testing.T) {
	backupPath := filepath.Join(t.TempDir(), "backup")
	projectDir := t.TempDir()

	// Set up project
	os.MkdirAll(filepath.Join(projectDir, ".claude"), 0755)
	os.WriteFile(filepath.Join(projectDir, ".claude", "settings.local.json"), []byte(`{}`), 0644)
	os.MkdirAll(filepath.Join(projectDir, "tools"), 0755)
	os.WriteFile(filepath.Join(projectDir, "tools", "script.py"), []byte("print()"), 0644)
	os.WriteFile(filepath.Join(projectDir, "CLAUDE.md"), []byte("# Test"), 0644)

	err := CreateBackup(backupPath, projectDir, "test-preset")
	if err != nil {
		t.Fatal(err)
	}

	// Check meta exists
	meta, err := GetBackupInfo(backupPath)
	if err != nil {
		t.Fatal(err)
	}
	if meta.ProjectDir != projectDir {
		t.Errorf("ProjectDir = %q, want %q", meta.ProjectDir, projectDir)
	}
	if !meta.HadClaude || !meta.HadTools || !meta.HadClaudeMD {
		t.Errorf("had flags wrong: claude=%v tools=%v md=%v", meta.HadClaude, meta.HadTools, meta.HadClaudeMD)
	}

	// Check files backed up
	if _, err := os.Stat(filepath.Join(backupPath, "claude", "settings.local.json")); err != nil {
		t.Error("settings.local.json not backed up")
	}
	if _, err := os.Stat(filepath.Join(backupPath, "tools", "script.py")); err != nil {
		t.Error("script.py not backed up")
	}
	if _, err := os.Stat(filepath.Join(backupPath, "CLAUDE.md")); err != nil {
		t.Error("CLAUDE.md not backed up")
	}
}

func TestCreateBackup_ClearsPrevious(t *testing.T) {
	backupPath := filepath.Join(t.TempDir(), "backup")
	projectDir := t.TempDir()

	os.MkdirAll(filepath.Join(projectDir, ".claude"), 0755)
	os.WriteFile(filepath.Join(projectDir, ".claude", "first.txt"), []byte("first"), 0644)
	CreateBackup(backupPath, projectDir, "first-load")

	// Remove first.txt, add second.txt
	os.Remove(filepath.Join(projectDir, ".claude", "first.txt"))
	os.WriteFile(filepath.Join(projectDir, ".claude", "second.txt"), []byte("second"), 0644)
	CreateBackup(backupPath, projectDir, "second-load")

	// Old file should be gone
	if _, err := os.Stat(filepath.Join(backupPath, "claude", "first.txt")); !os.IsNotExist(err) {
		t.Error("old backup file should be cleared")
	}
	// New file should be present
	if _, err := os.Stat(filepath.Join(backupPath, "claude", "second.txt")); err != nil {
		t.Error("new backup file missing")
	}
}

func TestCreateBackup_EmptyProject(t *testing.T) {
	backupPath := filepath.Join(t.TempDir(), "backup")
	projectDir := t.TempDir() // Empty

	err := CreateBackup(backupPath, projectDir, "test")
	if err != nil {
		t.Fatal(err)
	}

	meta, err := GetBackupInfo(backupPath)
	if err != nil {
		t.Fatal(err)
	}
	if meta.HadClaude || meta.HadTools || meta.HadClaudeMD {
		t.Error("empty project should have all had flags false")
	}
}

func TestRestoreBackup_Success(t *testing.T) {
	backupPath := filepath.Join(t.TempDir(), "backup")
	projectDir := t.TempDir()

	// Set up original project state
	os.MkdirAll(filepath.Join(projectDir, ".claude"), 0755)
	os.WriteFile(filepath.Join(projectDir, ".claude", "original.json"), []byte(`{"original":true}`), 0644)
	os.WriteFile(filepath.Join(projectDir, "CLAUDE.md"), []byte("# Original"), 0644)

	// Create backup
	CreateBackup(backupPath, projectDir, "test")

	// Simulate a load: replace project state
	os.RemoveAll(filepath.Join(projectDir, ".claude"))
	os.Remove(filepath.Join(projectDir, "CLAUDE.md"))
	os.MkdirAll(filepath.Join(projectDir, ".claude"), 0755)
	os.WriteFile(filepath.Join(projectDir, ".claude", "loaded.json"), []byte(`{"loaded":true}`), 0644)

	// Restore
	dir, err := RestoreBackup(backupPath)
	if err != nil {
		t.Fatal(err)
	}
	if dir != projectDir {
		t.Errorf("restored dir = %q, want %q", dir, projectDir)
	}

	// Original file should be back
	data, err := os.ReadFile(filepath.Join(projectDir, ".claude", "original.json"))
	if err != nil {
		t.Fatal("original.json not restored")
	}
	if string(data) != `{"original":true}` {
		t.Errorf("content = %q, want original", string(data))
	}

	// Loaded file should be gone
	if _, err := os.Stat(filepath.Join(projectDir, ".claude", "loaded.json")); !os.IsNotExist(err) {
		t.Error("loaded.json should be removed after restore")
	}

	// CLAUDE.md should be restored
	data, err = os.ReadFile(filepath.Join(projectDir, "CLAUDE.md"))
	if err != nil || string(data) != "# Original" {
		t.Error("CLAUDE.md not restored correctly")
	}

	// Backup should be cleaned up
	if HasBackup(backupPath) {
		t.Error("backup should be removed after restore")
	}
}

func TestRestoreBackup_NoBackup(t *testing.T) {
	backupPath := filepath.Join(t.TempDir(), "backup")
	_, err := RestoreBackup(backupPath)
	if err == nil {
		t.Error("expected error for missing backup")
	}
}

func TestHasBackup(t *testing.T) {
	backupPath := filepath.Join(t.TempDir(), "backup")
	if HasBackup(backupPath) {
		t.Error("should be false before backup")
	}

	projectDir := t.TempDir()
	CreateBackup(backupPath, projectDir, "test")
	if !HasBackup(backupPath) {
		t.Error("should be true after backup")
	}
}
