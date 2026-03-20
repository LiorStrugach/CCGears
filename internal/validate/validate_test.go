package validate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPresetName(t *testing.T) {
	valid := []string{
		"a",
		"z9",
		"my-preset",
		"argu-infra",
		"chess-reel-dev",
		"a1b2c3",
		"abcdefghijklmnopqrstuvwxyz0123456789abcdefghijkl", // 48 chars
	}
	for _, name := range valid {
		if err := PresetName(name); err != nil {
			t.Errorf("PresetName(%q) = %v, want nil", name, err)
		}
	}

	invalid := []string{
		"",
		"-leading",
		"trailing-",
		"UPPERCASE",
		"has space",
		"has_underscore",
		"has.dot",
		"abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklm", // 49 chars
	}
	for _, name := range invalid {
		if err := PresetName(name); err == nil {
			t.Errorf("PresetName(%q) = nil, want error", name)
		}
	}
}

func TestDirExists(t *testing.T) {
	dir := t.TempDir()
	if !DirExists(dir) {
		t.Errorf("DirExists(%q) = false, want true", dir)
	}
	if DirExists(filepath.Join(dir, "nope")) {
		t.Error("DirExists(nonexistent) = true, want false")
	}
	// File is not a dir
	f := filepath.Join(dir, "file.txt")
	os.WriteFile(f, []byte("hi"), 0644)
	if DirExists(f) {
		t.Errorf("DirExists(file) = true, want false")
	}
}

func TestFileExists(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "file.txt")
	os.WriteFile(f, []byte("hi"), 0644)
	if !FileExists(f) {
		t.Errorf("FileExists(%q) = false, want true", f)
	}
	if FileExists(dir) {
		t.Error("FileExists(dir) = true, want false")
	}
	if FileExists(filepath.Join(dir, "nope")) {
		t.Error("FileExists(nonexistent) = true, want false")
	}
}
