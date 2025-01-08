//go:build darwin || linux
// +build darwin linux

package legacy

import (
	"fmt"
	"path"
)

// copyPHP to destination, if it does not exist
func (c *CLIWrapper) copyPHP() error {
	if err := copyFile(c.PHPPath(), phpCLI); err != nil {
		return fmt.Errorf("could not copy PHP CLI: %w", err)
	}

	return nil
}

// PHPPath returns the path that the PHP CLI will reside
func (c *CLIWrapper) PHPPath() string {
	return path.Join(c.cacheDir(), phpPath)
}

func (c *CLIWrapper) phpSettings() []string {
	return nil
}
