package legacy

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/platformsh/cli/internal/file"
)

//go:embed archives/php_windows.zip
var phpCLI []byte

//go:embed archives/cacert.pem
var caCert []byte

// copyPHP to destination, if it does not exist
func (c *CLIWrapper) copyPHP(cacheDir string) error {
	destDir := filepath.Join(cacheDir, "php")

	r, err := zip.NewReader(bytes.NewReader(phpCLI), int64(len(phpCLI)))
	if err != nil {
		return fmt.Errorf("could not open zip reader: %w", err)
	}

	g := errgroup.Group{}
	g.SetLimit(runtime.GOMAXPROCS(0) * 2)
	for _, f := range r.File {
		g.Go(func() error {
			return copyZipFile(f, destDir)
		})
	}

	g.Go(func() error {
		return file.CopyIfChanged(filepath.Join(destDir, "extras", "cacert.pem"), caCert, 0o644)
	})

	return g.Wait()
}

// phpPath returns the path to the temporary PHP-CLI binary
func (c *CLIWrapper) phpPath(cacheDir string) string {
	return filepath.Join(cacheDir, "php", "php.exe")
}

func copyZipFile(f *zip.File, destDir string) error {
	before := time.Now()
	absPath := filepath.Join(destDir, f.Name)
	if !strings.HasPrefix(absPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", absPath)
	}

	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(absPath, 0755); err != nil {
			return fmt.Errorf("could not create extracted directory %s: %w", absPath, err)
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return fmt.Errorf("could not create parent directory for extracted file %s: %w", absPath, err)
	}

	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("could not open file in zip archive %s: %w", f.Name, err)
	}
	defer rc.Close()

	b, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("could not extract zipped file %s: %w", f.Name, err)
	}

	if err := file.CopyIfChanged(absPath, b, f.Mode()); err != nil {
		return fmt.Errorf("could not write extracted file %s: %w", absPath, err)
	}

	log.Printf("took %s to extract %s", time.Since(before), f.Name)

	return nil
}

func (c *CLIWrapper) phpSettings(cacheDir string) []string {
	return []string{
		"extension=" + filepath.Join(cacheDir, "php", "ext", "php_curl.dll"),
		"extension=" + filepath.Join(cacheDir, "php", "ext", "php_openssl.dll"),
		"openssl.cafile=" + filepath.Join(cacheDir, "php", "extras", "cacert.pem"),
	}
}
