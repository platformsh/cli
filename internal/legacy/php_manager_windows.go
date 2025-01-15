package legacy

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/platformsh/cli/internal/file"
)

//go:embed archives/php_windows.zip
var phpCLI []byte

//go:embed archives/cacert.pem
var caCert []byte

func (m *phpManagerPerOS) copy() error {
	destDir := filepath.Join(m.cacheDir, "php")

	r, err := zip.NewReader(bytes.NewReader(phpCLI), int64(len(phpCLI)))
	if err != nil {
		return fmt.Errorf("could not open zip reader: %w", err)
	}

	g := errgroup.Group{}
	g.SetLimit(runtime.NumCPU() * 4)
	for _, f := range r.File {
		g.Go(func() error {
			return copyZipFile(f, destDir)
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}

	if err := file.WriteIfNeeded(filepath.Join(destDir, "extras", "cacert.pem"), caCert, 0o644); err != nil {
		return err
	}

	return nil
}

func (m *phpManagerPerOS) phpPath() string {
	return filepath.Join(m.cacheDir, "php", "php.exe")
}

func (m *phpManagerPerOS) phpSettings() []string {
	return []string{
		"extension=" + filepath.Join(m.cacheDir, "php", "ext", "php_curl.dll"),
		"extension=" + filepath.Join(m.cacheDir, "php", "ext", "php_openssl.dll"),
		"openssl.cafile=" + filepath.Join(m.cacheDir, "php", "extras", "cacert.pem"),
	}
}

// copyZipFile extracts a file from the Zip to the destination directory.
// If the file already exists and has the correct size, it will be skipped.
func copyZipFile(f *zip.File, destDir string) error {
	destPath := filepath.Join(destDir, f.Name)
	if !strings.HasPrefix(destPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", destPath)
	}

	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(destPath, 0755); err != nil {
			return fmt.Errorf("could not create extracted directory %s: %w", destPath, err)
		}
		return nil
	}

	if existingFileInfo, err := os.Lstat(destPath); err == nil && uint64(existingFileInfo.Size()) == f.UncompressedSize64 {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("could not create parent directory for extracted file %s: %w", destPath, err)
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

	if err := file.Write(destPath, b, f.Mode()); err != nil {
		return fmt.Errorf("could not copy extracted file %s: %w", destPath, err)
	}

	return nil
}
