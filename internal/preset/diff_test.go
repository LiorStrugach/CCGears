package preset

import (
	"os"
	"path/filepath"
	"testing"
)

func setupDiffTest(t *testing.T) (presetDir, projectDir string) {
	t.Helper()
	presetDir = t.TempDir()
	projectDir = t.TempDir()

	// Preset has claude/settings.local.json and tools/script.py
	os.MkdirAll(filepath.Join(presetDir, "claude"), 0755)
	os.WriteFile(filepath.Join(presetDir, "claude", "settings.local.json"), []byte(`{"v":1}`), 0644)
	os.MkdirAll(filepath.Join(presetDir, "tools"), 0755)
	os.WriteFile(filepath.Join(presetDir, "tools", "script.py"), []byte("original"), 0644)
	os.WriteFile(filepath.Join(presetDir, "CLAUDE.md"), []byte("# Original"), 0644)

	// Project mirrors the preset
	os.MkdirAll(filepath.Join(projectDir, ".claude"), 0755)
	os.WriteFile(filepath.Join(projectDir, ".claude", "settings.local.json"), []byte(`{"v":1}`), 0644)
	os.MkdirAll(filepath.Join(projectDir, "tools"), 0755)
	os.WriteFile(filepath.Join(projectDir, "tools", "script.py"), []byte("original"), 0644)
	os.WriteFile(filepath.Join(projectDir, "CLAUDE.md"), []byte("# Original"), 0644)

	return presetDir, projectDir
}

func TestDiff_NoChanges(t *testing.T) {
	presetDir, projectDir := setupDiffTest(t)
	diff := ComputeDiff(presetDir, projectDir)

	if diff.HasChanges() {
		t.Errorf("expected no changes: added=%v removed=%v modified=%v",
			diff.Added, diff.Removed, diff.Modified)
	}
	if diff.Unchanged != 3 {
		t.Errorf("unchanged = %d, want 3", diff.Unchanged)
	}
}

func TestDiff_AddedFile(t *testing.T) {
	presetDir, projectDir := setupDiffTest(t)

	// Add a new file to project
	os.WriteFile(filepath.Join(projectDir, "tools", "new.py"), []byte("new"), 0644)

	diff := ComputeDiff(presetDir, projectDir)
	if len(diff.Added) != 1 || diff.Added[0] != "tools/new.py" {
		t.Errorf("Added = %v, want [tools/new.py]", diff.Added)
	}
}

func TestDiff_RemovedFile(t *testing.T) {
	presetDir, projectDir := setupDiffTest(t)

	// Remove a file from project
	os.Remove(filepath.Join(projectDir, "tools", "script.py"))

	diff := ComputeDiff(presetDir, projectDir)
	if len(diff.Removed) != 1 || diff.Removed[0] != "tools/script.py" {
		t.Errorf("Removed = %v, want [tools/script.py]", diff.Removed)
	}
}

func TestDiff_ModifiedFile(t *testing.T) {
	presetDir, projectDir := setupDiffTest(t)

	// Modify a file in project
	os.WriteFile(filepath.Join(projectDir, ".claude", "settings.local.json"), []byte(`{"v":2}`), 0644)

	diff := ComputeDiff(presetDir, projectDir)
	if len(diff.Modified) != 1 || diff.Modified[0] != ".claude/settings.local.json" {
		t.Errorf("Modified = %v, want [.claude/settings.local.json]", diff.Modified)
	}
}

func TestDiff_Mixed(t *testing.T) {
	presetDir, projectDir := setupDiffTest(t)

	// Add
	os.WriteFile(filepath.Join(projectDir, "tools", "new.py"), []byte("new"), 0644)
	// Modify
	os.WriteFile(filepath.Join(projectDir, ".claude", "settings.local.json"), []byte(`{"v":99}`), 0644)
	// Remove
	os.Remove(filepath.Join(projectDir, "CLAUDE.md"))

	diff := ComputeDiff(presetDir, projectDir)
	if len(diff.Added) != 1 {
		t.Errorf("Added count = %d, want 1", len(diff.Added))
	}
	if len(diff.Modified) != 1 {
		t.Errorf("Modified count = %d, want 1", len(diff.Modified))
	}
	if len(diff.Removed) != 1 {
		t.Errorf("Removed count = %d, want 1", len(diff.Removed))
	}
	if diff.Unchanged != 1 {
		t.Errorf("Unchanged = %d, want 1", diff.Unchanged)
	}
}

func TestDiff_MissingDirs(t *testing.T) {
	presetDir, projectDir := setupDiffTest(t)

	// Remove entire .claude from project
	os.RemoveAll(filepath.Join(projectDir, ".claude"))

	diff := ComputeDiff(presetDir, projectDir)
	if len(diff.Removed) != 1 {
		t.Errorf("Removed = %v, want 1 entry for settings.local.json", diff.Removed)
	}
}
