package alt

import (
	"os"
	"path/filepath"
	"strings"
)

// InPath tests if a directory is in the PATH.
func InPath(dir string) bool {
	return inPathValue(dir, os.Getenv("PATH"))
}

func inPathValue(dir, path string) bool {
	homeDir, _ := os.UserHomeDir()
	normalized := normalizePathEntry(dir, homeDir)
	for _, e := range filepath.SplitList(path) {
		if normalizePathEntry(e, homeDir) == normalized {
			return true
		}
	}
	return false
}

func normalizePathEntry(path, homeDir string) string {
	if homeDir != "" && strings.HasPrefix(path, "~") {
		path = homeDir + path[1:]
	}
	if path == "" {
		path = "."
	}
	path = filepath.Clean(os.ExpandEnv(path))
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	return path
}
