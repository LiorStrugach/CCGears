package preset

import (
	"os"
	"path/filepath"
	"testing"
)

func setupScanFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	// Project 1: has .claude/ with skills and tools/
	p1 := filepath.Join(root, "project-alpha")
	os.MkdirAll(filepath.Join(p1, ".claude", "skills", "my-skill"), 0755)
	os.WriteFile(filepath.Join(p1, ".claude", "settings.local.json"), []byte(`{}`), 0644)
	os.WriteFile(filepath.Join(p1, ".claude", "skills", "my-skill", "SKILL.md"), []byte("---\nname: my-skill\n---\n"), 0644)
	os.MkdirAll(filepath.Join(p1, "tools"), 0755)
	os.WriteFile(filepath.Join(p1, "tools", "script.py"), []byte("print()"), 0644)
	os.WriteFile(filepath.Join(p1, "tools", "helper.sh"), []byte("#!/bin/sh"), 0644)
	os.WriteFile(filepath.Join(p1, "CLAUDE.md"), []byte("# Alpha Project\nInstructions"), 0644)

	// Project 2: has only .claude/
	p2 := filepath.Join(root, "project-beta")
	os.MkdirAll(filepath.Join(p2, ".claude"), 0755)
	os.WriteFile(filepath.Join(p2, ".claude", "settings.local.json"), []byte(`{}`), 0644)

	// Project 3: inside node_modules (should be skipped)
	p3 := filepath.Join(root, "project-alpha", "node_modules", "some-pkg")
	os.MkdirAll(filepath.Join(p3, ".claude"), 0755)
	os.WriteFile(filepath.Join(p3, ".claude", "settings.local.json"), []byte(`{}`), 0644)

	// Project 4: too deep (depth 5 from root)
	deep := filepath.Join(root, "a", "b", "c", "d", "e")
	os.MkdirAll(filepath.Join(deep, ".claude"), 0755)
	os.WriteFile(filepath.Join(deep, ".claude", "settings.local.json"), []byte(`{}`), 0644)

	// Not a project: no .claude/ or tools/ or CLAUDE.md
	np := filepath.Join(root, "not-a-project")
	os.MkdirAll(np, 0755)
	os.WriteFile(filepath.Join(np, "README.md"), []byte("hello"), 0644)

	return root
}

func TestScanForProjects(t *testing.T) {
	root := setupScanFixture(t)

	results, err := ScanForProjects(root, 4)
	if err != nil {
		t.Fatal(err)
	}

	// Should find project-alpha and project-beta, but NOT node_modules or too-deep
	if len(results) != 2 {
		names := make([]string, len(results))
		for i, r := range results {
			names[i] = r.DirName
		}
		t.Fatalf("expected 2 results, got %d: %v", len(results), names)
	}

	// Find project-alpha
	var alpha *ScanResult
	for i := range results {
		if results[i].DirName == "project-alpha" {
			alpha = &results[i]
		}
	}
	if alpha == nil {
		t.Fatal("project-alpha not found")
	}
	if !alpha.HasClaude {
		t.Error("project-alpha should have .claude/")
	}
	if !alpha.HasTools {
		t.Error("project-alpha should have tools/")
	}
	if !alpha.HasClaudeMD {
		t.Error("project-alpha should have CLAUDE.md")
	}
	if alpha.SkillCount != 1 {
		t.Errorf("SkillCount = %d, want 1", alpha.SkillCount)
	}
	if alpha.ToolCount != 2 {
		t.Errorf("ToolCount = %d, want 2", alpha.ToolCount)
	}
	if alpha.Headline != "Alpha Project" {
		t.Errorf("Headline = %q, want %q", alpha.Headline, "Alpha Project")
	}
}

func TestScanForProjects_SkipsNodeModules(t *testing.T) {
	root := setupScanFixture(t)

	results, _ := ScanForProjects(root, 10) // high depth to catch everything

	for _, r := range results {
		if r.DirName == "some-pkg" {
			t.Error("should not find projects inside node_modules")
		}
	}
}

func TestScanForProjects_MaxDepth(t *testing.T) {
	root := setupScanFixture(t)

	// With depth 2, the deep project (depth 5) should not be found
	results, _ := ScanForProjects(root, 2)
	for _, r := range results {
		if r.DirName == "e" {
			t.Error("should not find projects beyond max depth")
		}
	}
}

func TestAutoPresetName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"TestMaker", "testmaker"},
		{"event_sampling", "event-sampling"},
		{"My Project", "my-project"},
		{"CLAUDE-config", "claude-config"},
		{"---leading---", "leading"},
		{"a", "a"},
		{"123abc", "123abc"},
	}
	for _, tt := range tests {
		got := AutoPresetName(tt.input)
		if got != tt.want {
			t.Errorf("AutoPresetName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestUniquePresetName(t *testing.T) {
	home := t.TempDir()
	t.Setenv("CCGEARS_HOME", home)
	cfg, project := setupTest(t)
	_ = project

	// No collision
	name := UniquePresetName(cfg, "test")
	if name != "test" {
		t.Errorf("name = %q, want %q", name, "test")
	}

	// Create the preset dir to cause collision
	os.MkdirAll(filepath.Join(cfg.StorePath, "test"), 0755)
	name = UniquePresetName(cfg, "test")
	if name != "test-2" {
		t.Errorf("name = %q, want %q", name, "test-2")
	}
}

func TestShortenPath(t *testing.T) {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, "Documents", "Hobby", "TestMaker")
	short := ShortenPath(path)
	if short != "~/Documents/Hobby/TestMaker" {
		t.Errorf("ShortenPath = %q, want %q", short, "~/Documents/Hobby/TestMaker")
	}

	// Non-home path unchanged
	other := "/tmp/foo"
	if ShortenPath(other) != other {
		t.Errorf("ShortenPath(%q) should be unchanged", other)
	}
}
