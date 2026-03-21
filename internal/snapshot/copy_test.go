package snapshot

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(data)
}

func TestCopyDir_BasicFiles(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "out")

	writeFile(t, filepath.Join(src, "a.txt"), "alpha")
	writeFile(t, filepath.Join(src, "sub/b.txt"), "beta")

	count, err := CopyDir(src, dst, nil)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
	if got := readFile(t, filepath.Join(dst, "a.txt")); got != "alpha" {
		t.Errorf("a.txt = %q, want %q", got, "alpha")
	}
	if got := readFile(t, filepath.Join(dst, "sub/b.txt")); got != "beta" {
		t.Errorf("sub/b.txt = %q, want %q", got, "beta")
	}
}

func TestCopyDir_Exclusions(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "out")

	writeFile(t, filepath.Join(src, "keep.txt"), "keep")
	writeFile(t, filepath.Join(src, ".DS_Store"), "junk")
	writeFile(t, filepath.Join(src, "__pycache__/mod.pyc"), "cache")
	writeFile(t, filepath.Join(src, "sub/.DS_Store"), "junk2")

	count, err := CopyDir(src, dst, DefaultExcludes)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
	if _, err := os.Stat(filepath.Join(dst, ".DS_Store")); !os.IsNotExist(err) {
		t.Error(".DS_Store should not be copied")
	}
	if _, err := os.Stat(filepath.Join(dst, "__pycache__")); !os.IsNotExist(err) {
		t.Error("__pycache__ should not be copied")
	}
	if _, err := os.Stat(filepath.Join(dst, "sub/.DS_Store")); !os.IsNotExist(err) {
		t.Error("sub/.DS_Store should not be copied")
	}
}

func TestCopyDir_Symlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks require elevated permissions on Windows")
	}
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "out")

	// Create a file and a symlink to it
	writeFile(t, filepath.Join(src, "real.txt"), "content")
	if err := os.Symlink(filepath.Join(src, "real.txt"), filepath.Join(src, "link.txt")); err != nil {
		t.Fatal(err)
	}

	// Create a dir and a symlink to it
	writeFile(t, filepath.Join(src, "realdir/file.txt"), "dirfile")
	if err := os.Symlink(filepath.Join(src, "realdir"), filepath.Join(src, "linkdir")); err != nil {
		t.Fatal(err)
	}

	count, err := CopyDir(src, dst, nil)
	if err != nil {
		t.Fatal(err)
	}
	if count != 4 {
		t.Errorf("count = %d, want 4", count)
	}

	// Symlink to file should be a regular file with the content
	if got := readFile(t, filepath.Join(dst, "link.txt")); got != "content" {
		t.Errorf("link.txt = %q, want %q", got, "content")
	}
	info, _ := os.Lstat(filepath.Join(dst, "link.txt"))
	if info.Mode()&os.ModeSymlink != 0 {
		t.Error("link.txt should be a regular file, not a symlink")
	}

	// Symlink to dir should be a regular dir
	if got := readFile(t, filepath.Join(dst, "linkdir/file.txt")); got != "dirfile" {
		t.Errorf("linkdir/file.txt = %q, want %q", got, "dirfile")
	}
}

func TestCopyDir_DanglingSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks require elevated permissions on Windows")
	}
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "out")

	writeFile(t, filepath.Join(src, "real.txt"), "exists")
	if err := os.Symlink("/nonexistent/path", filepath.Join(src, "dangling")); err != nil {
		t.Fatal(err)
	}

	count, err := CopyDir(src, dst, nil)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1 (dangling symlink skipped)", count)
	}
}

func TestCopyDir_EmptySource(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "out")

	count, err := CopyDir(src, dst, nil)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}
	// Destination should exist as empty dir
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		t.Error("destination dir should be created even if empty")
	}
}

func TestCopyDir_SourceNotExist(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "out")
	_, err := CopyDir("/nonexistent/path", dst, nil)
	if err == nil {
		t.Error("expected error for nonexistent source")
	}
}

func TestCopyDir_Permissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix file permissions not applicable on Windows")
	}
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "out")

	path := filepath.Join(src, "script.sh")
	if err := os.WriteFile(path, []byte("#!/bin/sh"), 0755); err != nil {
		t.Fatal(err)
	}

	_, _ = CopyDir(src, dst, nil)

	info, err := os.Stat(filepath.Join(dst, "script.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0111 == 0 {
		t.Error("executable permission not preserved")
	}
}

func TestCountFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.txt"), "hello")
	writeFile(t, filepath.Join(dir, "sub/b.txt"), "world")
	writeFile(t, filepath.Join(dir, ".DS_Store"), "junk")

	count, size := CountFiles(dir, DefaultExcludes)
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
	if size != 10 {
		t.Errorf("size = %d, want 10", size)
	}
}

func TestRemoveIfExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub/file.txt")
	writeFile(t, path, "data")

	if err := RemoveIfExists(filepath.Join(dir, "sub")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "sub")); !os.IsNotExist(err) {
		t.Error("sub should be removed")
	}

	// No error for nonexistent
	if err := RemoveIfExists(filepath.Join(dir, "nope")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
