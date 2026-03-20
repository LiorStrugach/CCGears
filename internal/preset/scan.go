package preset

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mryan/ccgears/internal/config"
	"github.com/mryan/ccgears/internal/validate"
)

// ScanResult represents a discovered project with Claude Code configuration.
type ScanResult struct {
	ProjectDir  string
	DirName     string
	HasClaude   bool
	HasTools    bool
	HasClaudeMD bool
	SkillCount  int
	ToolCount   int
	Headline    string
}

var skipDirs = map[string]bool{
	"node_modules": true,
	".git":         true,
	".venv":        true,
	"__pycache__":  true,
	"vendor":       true,
	"dist":         true,
	"build":        true,
	".next":        true,
	".cache":       true,
	".ccgears":     true,
	".Trash":       true,
}

// ScanForProjects walks rootDir up to maxDepth levels, finding directories
// that contain .claude/, tools/, or CLAUDE.md.
func ScanForProjects(rootDir string, maxDepth int) ([]ScanResult, error) {
	rootDir, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, err
	}

	rootDepth := strings.Count(rootDir, string(filepath.Separator))
	var results []ScanResult
	seen := make(map[string]bool)

	filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}

		name := d.Name()

		// Skip noise directories
		if skipDirs[name] {
			return filepath.SkipDir
		}

		// Don't recurse into .claude/ itself
		if name == ".claude" {
			return filepath.SkipDir
		}

		// Enforce max depth
		depth := strings.Count(path, string(filepath.Separator)) - rootDepth
		if depth > maxDepth {
			return filepath.SkipDir
		}

		// Check if this directory is a project with CC config
		claudeDir := filepath.Join(path, ".claude")
		toolsDir := filepath.Join(path, "tools")
		claudeMD := filepath.Join(path, "CLAUDE.md")

		hasClaude := validate.DirExists(claudeDir)
		hasTools := validate.DirExists(toolsDir)
		hasClaudeMD := validate.FileExists(claudeMD)

		if !hasClaude && !hasTools && !hasClaudeMD {
			return nil
		}

		// Avoid duplicates
		if seen[path] {
			return nil
		}
		seen[path] = true

		result := ScanResult{
			ProjectDir:  path,
			DirName:     filepath.Base(path),
			HasClaude:   hasClaude,
			HasTools:    hasTools,
			HasClaudeMD: hasClaudeMD,
		}

		if hasClaude {
			result.SkillCount = countSkills(filepath.Join(claudeDir, "skills"))
		}
		if hasTools {
			result.ToolCount = countScripts(toolsDir)
		}
		if hasClaudeMD {
			result.Headline = readHeadline(claudeMD)
		}

		results = append(results, result)
		return nil
	})

	return results, nil
}

// AutoPresetName converts a directory name to a valid preset name.
func AutoPresetName(dirName string) string {
	name := strings.ToLower(dirName)
	// Replace non-alphanumeric with hyphens
	re := regexp.MustCompile(`[^a-z0-9]+`)
	name = re.ReplaceAllString(name, "-")
	// Trim leading/trailing hyphens
	name = strings.Trim(name, "-")
	if len(name) > 48 {
		name = name[:48]
		name = strings.TrimRight(name, "-")
	}
	if name == "" {
		name = "unnamed"
	}
	return name
}

// UniquePresetName returns a name that doesn't collide with existing presets.
func UniquePresetName(cfg *config.Config, baseName string) string {
	name := baseName
	if _, err := os.Stat(filepath.Join(cfg.StorePath, name)); os.IsNotExist(err) {
		return name
	}
	for i := 2; i <= 99; i++ {
		candidate := baseName + "-" + strings.Repeat("", 0) + itoa(i)
		if _, err := os.Stat(filepath.Join(cfg.StorePath, candidate)); os.IsNotExist(err) {
			return candidate
		}
	}
	return baseName
}

// ShortenPath replaces $HOME prefix with ~ for display.
func ShortenPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}
	return path
}

func countSkills(skillsDir string) int {
	if !validate.DirExists(skillsDir) {
		return 0
	}
	count := 0
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return 0
	}
	for _, e := range entries {
		if e.IsDir() && validate.FileExists(filepath.Join(skillsDir, e.Name(), "SKILL.md")) {
			count++
		}
	}
	return count
}

func countScripts(toolsDir string) int {
	count := 0
	entries, err := os.ReadDir(toolsDir)
	if err != nil {
		return 0
	}
	for _, e := range entries {
		if !e.IsDir() {
			ext := strings.ToLower(filepath.Ext(e.Name()))
			if ext == ".py" || ext == ".sh" || ext == ".js" {
				count++
			}
		}
	}
	return count
}

func readHeadline(path string) string {
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

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return itoa(n/10) + string(rune('0'+n%10))
}
