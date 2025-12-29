package legacy

import (
	_ "embed"
	"path/filepath"

	"github.com/platformsh/cli/internal/file"
)

//go:embed archives/php_windows.exe
var phpCLI []byte

//go:embed archives/cacert.pem
var caCert []byte

func (m *phpManagerPerOS) copy() error {
	if err := file.WriteIfNeeded(m.binPath(), phpCLI, 0o755); err != nil {
		return err
	}
	// Write cacert.pem for OpenSSL CA bundle (Windows needs this explicitly).
	return file.WriteIfNeeded(filepath.Join(m.cacheDir, "cacert.pem"), caCert, 0o644)
}

func (m *phpManagerPerOS) binPath() string {
	return filepath.Join(m.cacheDir, "php.exe")
}

func (m *phpManagerPerOS) settings() []string {
	return []string{
		"openssl.cafile=" + filepath.Join(m.cacheDir, "cacert.pem"),
	}
}
