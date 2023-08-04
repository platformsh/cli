//go:build darwin || linux
// +build darwin linux

package legacy

import (
	"path"

	"github.com/platformsh/cli/internal/file"
)

// copyPHP to destination, if it does not exist
func (c *CLIWrapper) copyPHP() error {
	return file.CopyIfChanged(c.PHPPath(), phpCLI, 0o755)
}

// PHPPath returns the path that the PHP CLI will reside
func (c *CLIWrapper) PHPPath() string {
	return path.Join(c.cacheDir(), phpPath)
}

func (c *CLIWrapper) phpSettings() []string {
	return nil
}
