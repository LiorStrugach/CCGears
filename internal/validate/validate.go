package validate

import (
	"fmt"
	"os"
	"regexp"
)

// presetNameRe allows 1-48 chars: lowercase alphanumeric, hyphens only in the middle.
var presetNameRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,46}[a-z0-9])?$`)

// PresetName validates a preset name. Returns nil if valid, an error describing the problem otherwise.
func PresetName(name string) error {
	if name == "" {
		return fmt.Errorf("preset name cannot be empty")
	}
	if len(name) > 48 {
		return fmt.Errorf("preset name must be 48 characters or fewer (got %d)", len(name))
	}
	if !presetNameRe.MatchString(name) {
		return fmt.Errorf("preset name must contain only lowercase letters, digits, and hyphens (no leading/trailing hyphens)")
	}
	return nil
}

// DirExists returns true if path is an existing directory.
func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// FileExists returns true if path is an existing regular file.
func FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
