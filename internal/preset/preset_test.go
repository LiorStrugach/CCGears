package preset

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/mryan/ccgears/internal/config"
	"github.com/mryan/ccgears/internal/snapshot"
)

func setupTest(t *testing.T) (*config.Config, string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("CCGEARS_HOME", home)
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.EnsureDirs(); err != nil {
		t.Fatal(err)
	}

	// Create a project directory with .claude/ and tools/
	project := t.TempDir()
	if err := os.MkdirAll(filepath.Join(project, ".claude", "skills", "my-skill"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".claude", "settings.local.json"), []byte(`{"permissions":{}}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".claude", "skills", "my-skill", "SKILL.md"), []byte("---\nname: my-skill\n---\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(project, "tools"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, "tools", "script.py"), []byte("print('hello')"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, "CLAUDE.md"), []byte("# My Project\nInstructions here"), 0644); err != nil {
		t.Fatal(err)
	}

	return cfg, project
}

func TestCreate_Success(t *testing.T) {
	cfg, project := setupTest(t)

	meta, count, err := Create(cfg, "test-preset", "A test preset", project)
	if err != nil {
		t.Fatal(err)
	}
	if meta.Name != "test-preset" {
		t.Errorf("Name = %q, want %q", meta.Name, "test-preset")
	}
	if count < 3 {
		t.Errorf("count = %d, want >= 3", count)
	}

	// Verify files in store
	presetDir := filepath.Join(cfg.StorePath, "test-preset")
	if _, err := os.Stat(filepath.Join(presetDir, "claude", "settings.local.json")); err != nil {
		t.Error("settings.local.json not in preset")
	}
	if _, err := os.Stat(filepath.Join(presetDir, "tools", "script.py")); err != nil {
		t.Error("script.py not in preset")
	}
	if _, err := os.Stat(filepath.Join(presetDir, "CLAUDE.md")); err != nil {
		t.Error("CLAUDE.md not in preset")
	}
	if _, err := os.Stat(filepath.Join(presetDir, "preset.json")); err != nil {
		t.Error("preset.json not in preset")
	}
}

func TestCreate_AlreadyExists(t *testing.T) {
	cfg, project := setupTest(t)

	_, _, _ = Create(cfg, "test-preset", "first", project)
	_, _, err := Create(cfg, "test-preset", "second", project)
	if !errors.Is(err, ErrAlreadyExists) {
		t.Errorf("err = %v, want ErrAlreadyExists", err)
	}
}

func TestCreate_InvalidName(t *testing.T) {
	cfg, project := setupTest(t)

	_, _, err := Create(cfg, "INVALID", "bad", project)
	if err == nil {
		t.Error("expected error for invalid name")
	}
}

func TestCreate_EmptyProject(t *testing.T) {
	cfg, _ := setupTest(t)
	emptyDir := t.TempDir()

	_, _, err := Create(cfg, "empty", "nothing here", emptyDir)
	if !errors.Is(err, ErrEmptyProject) {
		t.Errorf("err = %v, want ErrEmptyProject", err)
	}
}

func TestGet_Success(t *testing.T) {
	cfg, project := setupTest(t)
	_, _, _ = Create(cfg, "test-preset", "desc", project)

	meta, err := Get(cfg, "test-preset")
	if err != nil {
		t.Fatal(err)
	}
	if meta.Description != "desc" {
		t.Errorf("Description = %q, want %q", meta.Description, "desc")
	}
}

func TestGet_NotFound(t *testing.T) {
	cfg, _ := setupTest(t)

	_, err := Get(cfg, "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestDelete_Success(t *testing.T) {
	cfg, project := setupTest(t)
	_, _, _ = Create(cfg, "test-preset", "desc", project)

	if err := Delete(cfg, "test-preset"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(cfg.StorePath, "test-preset")); !os.IsNotExist(err) {
		t.Error("preset directory should be removed")
	}
}

func TestDelete_NotFound(t *testing.T) {
	cfg, _ := setupTest(t)

	err := Delete(cfg, "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestList_Empty(t *testing.T) {
	cfg, _ := setupTest(t)

	presets, err := List(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(presets) != 0 {
		t.Errorf("len = %d, want 0", len(presets))
	}
}

func TestList_Multiple(t *testing.T) {
	cfg, project := setupTest(t)

	_, _, _ = Create(cfg, "beta", "second", project)
	_, _, _ = Create(cfg, "alpha", "first", project)

	presets, err := List(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(presets) != 2 {
		t.Fatalf("len = %d, want 2", len(presets))
	}
	if presets[0].Name != "alpha" {
		t.Errorf("first = %q, want %q", presets[0].Name, "alpha")
	}
	if presets[1].Name != "beta" {
		t.Errorf("second = %q, want %q", presets[1].Name, "beta")
	}
}

func TestLoad_Success(t *testing.T) {
	cfg, project := setupTest(t)
	_, _, _ = Create(cfg, "test-preset", "desc", project)

	targetDir := t.TempDir()
	count, err := Load(cfg, "test-preset", targetDir)
	if err != nil {
		t.Fatal(err)
	}
	if count < 3 {
		t.Errorf("count = %d, want >= 3", count)
	}

	if _, err := os.Stat(filepath.Join(targetDir, ".claude", "settings.local.json")); err != nil {
		t.Error("settings.local.json not loaded")
	}
	if _, err := os.Stat(filepath.Join(targetDir, "tools", "script.py")); err != nil {
		t.Error("script.py not loaded")
	}
	if _, err := os.Stat(filepath.Join(targetDir, "CLAUDE.md")); err != nil {
		t.Error("CLAUDE.md not loaded")
	}
}

func TestLoad_CreatesBackup(t *testing.T) {
	cfg, project := setupTest(t)
	_, _, _ = Create(cfg, "test-preset", "desc", project)

	_, _ = Load(cfg, "test-preset", project)

	if !snapshot.HasBackup(cfg.BackupPath) {
		t.Error("backup should exist after load")
	}
}

func TestLoad_NotFound(t *testing.T) {
	cfg, _ := setupTest(t)

	_, err := Load(cfg, "nonexistent", t.TempDir())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestSave_Success(t *testing.T) {
	cfg, project := setupTest(t)
	_, _, _ = Create(cfg, "test-preset", "desc", project)

	if err := os.WriteFile(filepath.Join(project, "tools", "new-tool.py"), []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}

	count, err := Save(cfg, "test-preset", project)
	if err != nil {
		t.Fatal(err)
	}
	if count < 4 {
		t.Errorf("count = %d, want >= 4", count)
	}

	presetDir := filepath.Join(cfg.StorePath, "test-preset")
	if _, err := os.Stat(filepath.Join(presetDir, "tools", "new-tool.py")); err != nil {
		t.Error("new-tool.py not saved to preset")
	}
}

func TestSave_NotFound(t *testing.T) {
	cfg, _ := setupTest(t)

	_, err := Save(cfg, "nonexistent", t.TempDir())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestRename_Success(t *testing.T) {
	cfg, project := setupTest(t)
	_, _, _ = Create(cfg, "old-name", "test", project)

	err := Rename(cfg, "old-name", "new-name")
	if err != nil {
		t.Fatal(err)
	}

	_, err = Get(cfg, "old-name")
	if !errors.Is(err, ErrNotFound) {
		t.Error("old name should not exist after rename")
	}

	meta, err := Get(cfg, "new-name")
	if err != nil {
		t.Fatal(err)
	}
	if meta.Name != "new-name" {
		t.Errorf("Name = %q, want %q", meta.Name, "new-name")
	}
	if meta.Description != "test" {
		t.Errorf("Description = %q, want %q", meta.Description, "test")
	}
}

func TestRename_UpdatesActivePreset(t *testing.T) {
	cfg, project := setupTest(t)
	_, _, _ = Create(cfg, "active-one", "test", project)
	cfg.ActivePreset = "active-one"
	_ = cfg.Save()

	err := Rename(cfg, "active-one", "active-two")
	if err != nil {
		t.Fatal(err)
	}

	cfg2, _ := config.Load()
	if cfg2.ActivePreset != "active-two" {
		t.Errorf("ActivePreset = %q, want %q", cfg2.ActivePreset, "active-two")
	}
}

func TestRename_NotFound(t *testing.T) {
	cfg, _ := setupTest(t)

	err := Rename(cfg, "nonexistent", "new-name")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestRename_AlreadyExists(t *testing.T) {
	cfg, project := setupTest(t)
	_, _, _ = Create(cfg, "first", "one", project)
	_, _, _ = Create(cfg, "second", "two", project)

	err := Rename(cfg, "first", "second")
	if !errors.Is(err, ErrAlreadyExists) {
		t.Errorf("err = %v, want ErrAlreadyExists", err)
	}
}

func TestRename_InvalidName(t *testing.T) {
	cfg, project := setupTest(t)
	_, _, _ = Create(cfg, "valid", "test", project)

	err := Rename(cfg, "valid", "INVALID")
	if err == nil {
		t.Error("expected error for invalid new name")
	}
}
