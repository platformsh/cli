//go:build darwin || linux

package legacy

import (
	"path/filepath"

	"github.com/platformsh/cli/internal/file"
)

func (m *phpManagerPerOS) copy() error {
	return file.WriteIfNeeded(m.binaryPath(), phpCLI, 0o755)
}

func (m *phpManagerPerOS) binaryPath() string {
	return filepath.Join(m.cacheDir, "php")
}

func (m *phpManagerPerOS) iniSettings() []string {
	return nil
}
