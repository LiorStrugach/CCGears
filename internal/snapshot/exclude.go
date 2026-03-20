package snapshot

import "path/filepath"

// DefaultExcludes are names excluded from all copy operations.
var DefaultExcludes = []string{
	".DS_Store",
	"__pycache__",
}

// ShouldExclude returns true if the given basename matches any exclusion pattern.
func ShouldExclude(baseName string, excludes []string) bool {
	for _, pattern := range excludes {
		if baseName == pattern {
			return true
		}
		if matched, _ := filepath.Match(pattern, baseName); matched {
			return true
		}
	}
	return false
}
