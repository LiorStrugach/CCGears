package preview

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFrontmatter_Simple(t *testing.T) {
	content := "---\nname: my-skill\ndescription: A simple skill\n---\n# Content"
	name, desc := ParseFrontmatter(content)
	if name != "my-skill" {
		t.Errorf("name = %q, want %q", name, "my-skill")
	}
	if desc != "A simple skill" {
		t.Errorf("desc = %q, want %q", desc, "A simple skill")
	}
}

func TestParseFrontmatter_MultilineDescription(t *testing.T) {
	content := `---
name: chess-reel
description: >
  Run the Chess Reel Pipeline to produce an engaging 9:16 short-form video
  (Instagram Reels / YouTube Shorts) from a chess opening name.
---

# Chess Reel Pipeline`
	name, desc := ParseFrontmatter(content)
	if name != "chess-reel" {
		t.Errorf("name = %q, want %q", name, "chess-reel")
	}
	expected := "Run the Chess Reel Pipeline to produce an engaging 9:16 short-form video (Instagram Reels / YouTube Shorts) from a chess opening name."
	if desc != expected {
		t.Errorf("desc = %q, want %q", desc, expected)
	}
}

func TestParseFrontmatter_NoFrontmatter(t *testing.T) {
	content := "# Just a heading\nNo frontmatter here"
	name, desc := ParseFrontmatter(content)
	if name != "" || desc != "" {
		t.Errorf("expected empty, got name=%q desc=%q", name, desc)
	}
}

func TestParseFrontmatter_EmptyFrontmatter(t *testing.T) {
	content := "---\n---\n# Content"
	name, desc := ParseFrontmatter(content)
	if name != "" || desc != "" {
		t.Errorf("expected empty, got name=%q desc=%q", name, desc)
	}
}

func TestParseFrontmatter_NameOnly(t *testing.T) {
	content := "---\nname: test\n---\n"
	name, desc := ParseFrontmatter(content)
	if name != "test" {
		t.Errorf("name = %q, want %q", name, "test")
	}
	if desc != "" {
		t.Errorf("desc = %q, want empty", desc)
	}
}

func TestBuild_FullPreset(t *testing.T) {
	presetDir := t.TempDir()

	// Create claude/settings.local.json
	os.MkdirAll(filepath.Join(presetDir, "claude", "skills", "my-skill"), 0755)
	os.WriteFile(filepath.Join(presetDir, "claude", "settings.local.json"), []byte(`{
		"permissions": {
			"allow": [
				"Bash(python3:*)",
				"Bash(pip:*)",
				"WebSearch",
				"WebFetch(domain:example.com)",
				"mcp__docker__run"
			]
		}
	}`), 0644)

	// Create a skill
	os.WriteFile(filepath.Join(presetDir, "claude", "skills", "my-skill", "SKILL.md"),
		[]byte("---\nname: my-skill\ndescription: A test skill\n---\n"), 0644)

	// Create tools
	os.MkdirAll(filepath.Join(presetDir, "tools"), 0755)
	os.WriteFile(filepath.Join(presetDir, "tools", "script.py"), []byte("print()"), 0644)
	os.WriteFile(filepath.Join(presetDir, "tools", "helper.sh"), []byte("#!/bin/sh"), 0644)

	// Create CLAUDE.md
	os.WriteFile(filepath.Join(presetDir, "CLAUDE.md"), []byte("# My Project\nInstructions"), 0644)

	p := Build(presetDir)

	if len(p.Skills) != 1 || p.Skills[0].Name != "my-skill" {
		t.Errorf("skills = %v, want 1 skill named my-skill", p.Skills)
	}
	if len(p.MCPTools) != 1 || p.MCPTools[0] != "mcp__docker__run" {
		t.Errorf("MCPTools = %v, want [mcp__docker__run]", p.MCPTools)
	}
	if p.BashPerms != 2 {
		t.Errorf("BashPerms = %d, want 2", p.BashPerms)
	}
	if len(p.WebPerms) != 2 {
		t.Errorf("WebPerms = %v, want 2 entries", p.WebPerms)
	}
	if p.Headline != "My Project" {
		t.Errorf("Headline = %q, want %q", p.Headline, "My Project")
	}
	if len(p.ToolScripts) != 2 {
		t.Errorf("ToolScripts = %v, want 2 scripts", p.ToolScripts)
	}
	if p.TotalFiles < 5 {
		t.Errorf("TotalFiles = %d, want >= 5", p.TotalFiles)
	}
}

func TestBuild_EmptyPreset(t *testing.T) {
	presetDir := t.TempDir()
	p := Build(presetDir)

	if len(p.Skills) != 0 || len(p.MCPTools) != 0 || len(p.ToolScripts) != 0 {
		t.Error("empty preset should have no skills, MCPs, or tools")
	}
	if p.Headline != "" {
		t.Errorf("Headline = %q, want empty", p.Headline)
	}
}
