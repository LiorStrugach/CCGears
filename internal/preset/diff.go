package preset

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mryan/ccgears/internal/snapshot"
	"github.com/mryan/ccgears/internal/validate"
)

// DiffSummary describes file-level differences between a preset and the current project.
type DiffSummary struct {
	Added     []string // files in project but not in preset
	Removed   []string // files in preset but not in project
	Modified  []string // files in both but content differs
	Unchanged int
}

// HasChanges returns true if there are any differences.
func (d *DiffSummary) HasChanges() bool {
	return len(d.Added) > 0 || len(d.Removed) > 0 || len(d.Modified) > 0
}

// ComputeDiff compares the current project state against a stored preset.
func ComputeDiff(presetDir, projectDir string) *DiffSummary {
	diff := &DiffSummary{}

	// Map: logical path -> pairs of (presetAbsPath, projectAbsPath)
	// We compare:
	//   preset/claude/*  vs  project/.claude/*
	//   preset/tools/*   vs  project/tools/*
	//   preset/CLAUDE.md vs  project/CLAUDE.md

	type dirPair struct {
		presetSub  string // subdir in preset (e.g., "claude")
		projectSub string // subdir in project (e.g., ".claude")
	}
	pairs := []dirPair{
		{"claude", ".claude"},
		{"tools", "tools"},
	}

	presetFiles := make(map[string]string)  // relative path -> absolute path in preset
	projectFiles := make(map[string]string) // relative path -> absolute path in project

	for _, p := range pairs {
		pDir := filepath.Join(presetDir, p.presetSub)
		jDir := filepath.Join(projectDir, p.projectSub)

		if validate.DirExists(pDir) {
			for _, f := range listFilesRel(pDir) {
				key := p.projectSub + "/" + f
				presetFiles[key] = filepath.Join(pDir, f)
			}
		}
		if validate.DirExists(jDir) {
			for _, f := range listFilesRel(jDir) {
				key := p.projectSub + "/" + f
				projectFiles[key] = filepath.Join(jDir, f)
			}
		}
	}

	// Handle CLAUDE.md
	presetMD := filepath.Join(presetDir, "CLAUDE.md")
	projectMD := filepath.Join(projectDir, "CLAUDE.md")
	if validate.FileExists(presetMD) {
		presetFiles["CLAUDE.md"] = presetMD
	}
	if validate.FileExists(projectMD) {
		projectFiles["CLAUDE.md"] = projectMD
	}

	// All unique keys
	allKeys := make(map[string]bool)
	for k := range presetFiles {
		allKeys[k] = true
	}
	for k := range projectFiles {
		allKeys[k] = true
	}

	for key := range allKeys {
		pPath, inPreset := presetFiles[key]
		jPath, inProject := projectFiles[key]

		switch {
		case inProject && !inPreset:
			diff.Added = append(diff.Added, key)
		case inPreset && !inProject:
			diff.Removed = append(diff.Removed, key)
		default:
			if filesContentDiffer(pPath, jPath) {
				diff.Modified = append(diff.Modified, key)
			} else {
				diff.Unchanged++
			}
		}
	}

	sort.Strings(diff.Added)
	sort.Strings(diff.Removed)
	sort.Strings(diff.Modified)

	return diff
}

func listFilesRel(dir string) []string {
	var files []string
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if snapshot.ShouldExclude(d.Name(), snapshot.DefaultExcludes) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.IsDir() {
			rel, _ := filepath.Rel(dir, path)
			files = append(files, filepath.ToSlash(rel))
		}
		return nil
	})
	return files
}

func filesContentDiffer(a, b string) bool {
	infoA, errA := os.Stat(a)
	infoB, errB := os.Stat(b)
	if errA != nil || errB != nil {
		return true
	}
	if infoA.Size() != infoB.Size() {
		return true
	}

	dataA, errA := os.ReadFile(a)
	dataB, errB := os.ReadFile(b)
	if errA != nil || errB != nil {
		return true
	}
	return !bytes.Equal(dataA, dataB)
}

// FormatRelPath cleans up a diff path for display, replacing forward slashes.
func FormatRelPath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
