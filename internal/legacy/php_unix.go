//go:build darwin || linux
// +build darwin linux

package legacy

import (
	"path/filepath"

	"github.com/platformsh/cli/internal/file"
)

// copyPHP to destination, if it does not exist
func (c *CLIWrapper) copyPHP(cacheDir string) error {
	return file.WriteIfNeeded(c.phpPath(cacheDir), phpCLI, 0o755)
}

// phpPath returns the path to the temporary PHP-CLI binary
func (c *CLIWrapper) phpPath(cacheDir string) string {
	return filepath.Join(cacheDir, "php")
}

func (c *CLIWrapper) phpSettings(_ string) []string {
	return nil
}
