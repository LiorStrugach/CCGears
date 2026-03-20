package snapshot

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// CopyDir recursively copies src to dst, excluding entries matching excludes.
// Symlinks are dereferenced (target content is copied).
// Returns the number of files copied.
func CopyDir(src, dst string, excludes []string) (int, error) {
	src, err := filepath.Abs(src)
	if err != nil {
		return 0, fmt.Errorf("resolving source path: %w", err)
	}

	info, err := os.Stat(src)
	if err != nil {
		return 0, fmt.Errorf("stat source: %w", err)
	}
	if !info.IsDir() {
		return 0, fmt.Errorf("source is not a directory: %s", src)
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return 0, fmt.Errorf("creating destination: %w", err)
	}

	count := 0
	err = filepath.WalkDir(src, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		// Check exclusions on basename
		if ShouldExclude(d.Name(), excludes) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		dstPath := filepath.Join(dst, rel)

		// Handle symlinks
		if d.Type()&fs.ModeSymlink != 0 {
			return copySymlink(path, dstPath, excludes, &count)
		}

		// Regular directory
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		// Regular file
		info, err := d.Info()
		if err != nil {
			return err
		}
		if err := copyFile(path, dstPath, info.Mode()); err != nil {
			return err
		}
		count++
		return nil
	})

	return count, err
}

// copySymlink dereferences a symlink and copies the target.
func copySymlink(path, dstPath string, excludes []string, count *int) error {
	linkTarget, err := os.Readlink(path)
	if err != nil {
		return nil // skip unreadable symlinks
	}
	if !filepath.IsAbs(linkTarget) {
		linkTarget = filepath.Join(filepath.Dir(path), linkTarget)
	}

	targetInfo, err := os.Stat(linkTarget)
	if err != nil {
		return nil // dangling symlink, skip
	}

	if targetInfo.IsDir() {
		n, err := CopyDir(linkTarget, dstPath, excludes)
		*count += n
		return err
	}

	if err := copyFile(linkTarget, dstPath, targetInfo.Mode()); err != nil {
		return err
	}
	*count++
	return nil
}

// copyFile copies a single file preserving permissions.
func copyFile(src, dst string, perm fs.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("creating %s: %w", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copying to %s: %w", dst, err)
	}
	return nil
}

// CountFiles counts files recursively under dir, respecting exclusions.
func CountFiles(dir string, excludes []string) (int, int64) {
	count := 0
	var size int64
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if ShouldExclude(d.Name(), excludes) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.IsDir() {
			count++
			if info, err := d.Info(); err == nil {
				size += info.Size()
			}
		}
		return nil
	})
	return count, size
}

// RemoveIfExists removes a path if it exists. No error if it doesn't exist.
func RemoveIfExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return os.RemoveAll(path)
}
