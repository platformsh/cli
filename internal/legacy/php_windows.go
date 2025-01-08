package legacy

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/platformsh/cli/internal/file"
)

//go:embed archives/php_windows.zip
var phpCLI []byte

//go:embed archives/php_windows.zip.sha256
var phpCLIHash string

//go:embed archives/cacert.pem
var caCert []byte

// copyPHP to destination, if it does not exist
func (c *CLIWrapper) copyPHP(cacheDir string) error {
	destDir := filepath.Join(cacheDir, "php")
	hashPath := filepath.Join(destDir, "hash")
	hashOK, err := file.CheckHash(hashPath, phpCLIHash)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if hashOK {
		return nil
	}
	br := bytes.NewReader(phpCLI)
	r, err := zip.NewReader(br, int64(len(phpCLI)))
	if err != nil {
		return fmt.Errorf("could not open zip reader: %w", err)
	}

	for _, f := range r.File {
		if err := copyZipFile(f, destDir); err != nil {
			return err
		}
	}

	if err := os.WriteFile(filepath.Join(destDir, "extras", "cacert.pem"), caCert, 0o644); err != nil {
		return err
	}

	return file.CopyIfChanged(hashPath, []byte(phpCLIHash), 0o644)
}

// phpPath returns the path to the temporary PHP-CLI binary
func (c *CLIWrapper) phpPath(cacheDir string) string {
	return filepath.Join(cacheDir, "php", "php.exe")
}

func copyZipFile(f *zip.File, destDir string) error {
	absPath := filepath.Join(destDir, f.Name)
	if !strings.HasPrefix(absPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", absPath)
	}

	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(absPath, 0755); err != nil {
			return fmt.Errorf("could create extracted directory %s: %w", absPath, err)
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return fmt.Errorf("could create parent directory for extracted file %s: %w", absPath, err)
	}

	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("could not open file in zip archive %s: %w", f.Name, err)
	}
	defer rc.Close()

	destFile, err := os.OpenFile(absPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return fmt.Errorf("could not open destination for extracted file %s: %w", absPath, err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, rc); err != nil {
		return fmt.Errorf("could not write extracted file %s: %w", absPath, err)
	}

	return nil
}

func (c *CLIWrapper) phpSettings(cacheDir string) []string {
	return []string{
		"extension=" + filepath.Join(cacheDir, "php", "ext", "php_curl.dll"),
		"extension=" + filepath.Join(cacheDir, "php", "ext", "php_openssl.dll"),
		"openssl.cafile=" + filepath.Join(cacheDir, "php", "extras", "cacert.pem"),
	}
}
