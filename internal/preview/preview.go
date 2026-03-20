package preview

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mryan/ccgears/internal/snapshot"
	"github.com/mryan/ccgears/internal/validate"
)

// SkillInfo holds parsed skill metadata.
type SkillInfo struct {
	Name        string
	Description string
}

// PresetPreview summarizes a preset's contents for display before loading.
type PresetPreview struct {
	Skills      []SkillInfo
	MCPTools    []string
	BashPerms   int
	WebPerms    []string
	Headline    string // first # heading from CLAUDE.md
	ToolScripts []string
	TotalFiles  int
	TotalSize   int64
}

// Build generates a preview from a preset directory.
func Build(presetDir string) *PresetPreview {
	p := &PresetPreview{}

	claudeDir := filepath.Join(presetDir, "claude")
	toolsDir := filepath.Join(presetDir, "tools")
	claudeMD := filepath.Join(presetDir, "CLAUDE.md")

	// Parse skills
	if validate.DirExists(claudeDir) {
		p.Skills = parseSkills(filepath.Join(claudeDir, "skills"))
		parsePermissions(filepath.Join(claudeDir, "settings.local.json"), p)
	}

	// Parse tool scripts
	if validate.DirExists(toolsDir) {
		p.ToolScripts = listScripts(toolsDir)
	}

	// Parse CLAUDE.md headline
	if validate.FileExists(claudeMD) {
		p.Headline = extractHeadline(claudeMD)
	}

	// Count total files and size
	p.TotalFiles, p.TotalSize = snapshot.CountFiles(presetDir, snapshot.DefaultExcludes)

	return p
}

func parseSkills(skillsDir string) []SkillInfo {
	if !validate.DirExists(skillsDir) {
		return nil
	}

	var skills []SkillInfo
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		skillMD := filepath.Join(skillsDir, e.Name(), "SKILL.md")
		if !validate.FileExists(skillMD) {
			continue
		}

		data, err := os.ReadFile(skillMD)
		if err != nil {
			continue
		}

		name, desc := ParseFrontmatter(string(data))
		if name == "" {
			name = e.Name()
		}
		skills = append(skills, SkillInfo{Name: name, Description: desc})
	}

	return skills
}

// ParseFrontmatter extracts name and description from YAML frontmatter.
// Handles single-line values and multi-line > folded blocks.
func ParseFrontmatter(content string) (name, description string) {
	lines := strings.Split(content, "\n")
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != "---" {
		return "", ""
	}

	// Find closing ---
	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end < 0 {
		return "", ""
	}

	fmLines := lines[1:end]

	var currentKey string
	var currentVal strings.Builder

	flush := func() {
		if currentKey == "" {
			return
		}
		val := strings.TrimSpace(currentVal.String())
		switch currentKey {
		case "name":
			name = val
		case "description":
			description = val
		}
	}

	for _, line := range fmLines {
		// New key: starts with non-whitespace and contains ':'
		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' && strings.Contains(line, ":") {
			flush()
			parts := strings.SplitN(line, ":", 2)
			currentKey = strings.TrimSpace(parts[0])
			currentVal.Reset()
			val := strings.TrimSpace(parts[1])
			if val != ">" && val != "|" && val != "" {
				currentVal.WriteString(val)
			}
		} else if currentKey != "" {
			// Continuation line (indented)
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				if currentVal.Len() > 0 {
					currentVal.WriteString(" ")
				}
				currentVal.WriteString(trimmed)
			}
		}
	}
	flush()

	return name, description
}

func parsePermissions(settingsPath string, p *PresetPreview) {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return
	}

	var settings struct {
		Permissions struct {
			Allow []string `json:"allow"`
		} `json:"permissions"`
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		return
	}

	for _, perm := range settings.Permissions.Allow {
		switch {
		case strings.HasPrefix(perm, "mcp__"):
			p.MCPTools = append(p.MCPTools, perm)
		case strings.HasPrefix(perm, "Bash("):
			p.BashPerms++
		case strings.HasPrefix(perm, "WebSearch") || strings.HasPrefix(perm, "WebFetch"):
			p.WebPerms = append(p.WebPerms, perm)
		}
	}
}

func listScripts(toolsDir string) []string {
	var scripts []string
	filepath.WalkDir(toolsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && (strings.HasSuffix(d.Name(), ".py") || strings.HasSuffix(d.Name(), ".sh") || strings.HasSuffix(d.Name(), ".js")) {
			scripts = append(scripts, d.Name())
		}
		return nil
	})
	return scripts
}

func extractHeadline(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return ""
}
