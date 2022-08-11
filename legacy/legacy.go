package legacy

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"
)

//go:embed archives/platform.phar
var pshCLI []byte

var PSHVersion string
var PHPVersion string

const prefix = "psh-go"

var phpPath = fmt.Sprintf("php-%s", PHPVersion)
var pshPath = fmt.Sprintf("psh-%s", PSHVersion)

// copyFile from the given bytes to destination
func copyFile(destination string, fin []byte) error {
	if _, err := os.Stat(destination); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not stat file: %w", err)
	}

	fout, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer fout.Close()

	r := bytes.NewReader(fin)

	if _, err := io.Copy(fout, r); err != nil {
		return fmt.Errorf("could copy file: %w", err)
	}

	return nil
}

// LegacyCLIWrapper wraps the legacy CLI
type LegacyCLIWrapper struct {
}

func (c *LegacyCLIWrapper) cacheDir() string {
	return path.Join(os.TempDir(), fmt.Sprintf("%s-%s-%s", prefix, PHPVersion, PSHVersion))
}

// Init the CLI wrapper, creating a temporary directory and copying over files
func (c *LegacyCLIWrapper) Init() error {
	if err := os.Mkdir(c.cacheDir(), 0700); err != nil {
		if os.IsExist(err) {
			log.Printf("cache directory already exists: %s", c.cacheDir())
			return nil
		}

		return fmt.Errorf("could not create temporary directory: %w", err)
	}

	if err := c.copyPHP(); err != nil {
		return fmt.Errorf("could not copy files: %w", err)
	}
	if err := c.copyPSH(); err != nil {
		return fmt.Errorf("could not copy files: %w", err)
	}
	if err := os.Chmod(c.PHPPath(), 0700); err != nil {
		return fmt.Errorf("could not make PHP executable: %w", err)
	}

	return nil
}

// Cleanup the CLI wrapper, removing the cache directory that was created and any other related directory
func (c *LegacyCLIWrapper) Cleanup() error {
	files, err := os.ReadDir(os.TempDir())
	if err != nil {
		return fmt.Errorf("could not list temporary directory: %w", err)
	}

	for _, f := range files {
		if strings.HasPrefix(f.Name(), prefix) {
			err := os.RemoveAll(path.Join(os.TempDir(), f.Name()))
			if err != nil {
				log.Printf("could not remove directory: %s", f.Name())
			}
		}
	}

	w, _ := os.Open("")
	template.Must(template.New("php.ini").Parse("")).Execute(w, map[string]string{"PSHDir": c.cacheDir()})
	return nil
}

// Exec a legacy CLI command with the given arguments
func (c *LegacyCLIWrapper) Exec(ctx context.Context, args ...string) error {
	args = append([]string{c.PSHPath()}, args...)
	cmd := exec.CommandContext(ctx, c.PHPPath(), args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not run legacy CLI command: %w", err)
	}

	return nil
}

// PSHPath returns the path that the PSH CLI will reside
func (c *LegacyCLIWrapper) PSHPath() string {
	return path.Join(c.cacheDir(), pshPath)
}

// copyPSH to destination, if it does not exist
func (c *LegacyCLIWrapper) copyPSH() error {
	if err := copyFile(c.PSHPath(), pshCLI); err != nil {
		return fmt.Errorf("could not copy legacy Platform.sh CLI: %w", err)
	}

	return nil
}
