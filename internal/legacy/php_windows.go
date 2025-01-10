package legacy

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

//go:embed archives/php_windows.zip
var phpCLI []byte

//go:embed archives/cacert.pem
var caCert []byte

// copyPHP to destination, if it does not exist
func (c *CLIWrapper) copyPHP() error {
	dest := path.Join(c.cacheDir(), "php")
	br := bytes.NewReader(phpCLI)
	r, err := zip.NewReader(br, int64(len(phpCLI)))
	if err != nil {
		return fmt.Errorf("could not open zip reader: %w", err)
	}

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("could not open zipped file %s: %w", f.Name, err)
		}
		defer rc.Close()

		fpath := filepath.Join(dest, f.Name[strings.Index(f.Name, string(os.PathSeparator))+1:])
		if f.FileInfo().IsDir() {
			continue
		}

		if lastIndex := strings.LastIndex(fpath, string(os.PathSeparator)); lastIndex > -1 {
			fdir := fpath[:lastIndex]
			if err := os.MkdirAll(fdir, 0755); err != nil {
				return fmt.Errorf("could create parent directory %s: %w", fdir, err)
			}
		}

		f, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("could not open file to unzip %s: %w", fpath, err)
		}
		defer f.Close()

		if _, err := io.Copy(f, rc); err != nil {
			return fmt.Errorf("could not write zipped file %s: %w", fpath, err)
		}
	}

	copyFile(path.Join(c.cacheDir(), "php", "extras", "cacert.pem"), caCert)

	return nil
}

// PHPPath returns the path that the PHP CLI will reside
func (c *CLIWrapper) PHPPath() string {
	return path.Join(c.cacheDir(), "php", "php.exe")
}

func (c *CLIWrapper) phpSettings() []string {
	cacheDir := c.cacheDir()

	return []string{
		"extension=" + filepath.Join(cacheDir, "php", "ext", "php_curl.dll"),
		"extension=" + filepath.Join(cacheDir, "php", "ext", "php_openssl.dll"),
		"openssl.cafile=" + filepath.Join(cacheDir, "php", "extras", "cacert.pem"),
	}
}
