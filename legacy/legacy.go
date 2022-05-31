package legacy

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

//go:embed archives/platform.phar
var pshCLI []byte

const (
	pshVersion = "3.79.7"
	phpVersion = "8.1.6"
)

var phpPath = fmt.Sprintf("php-%s", phpVersion)
var pshPath = fmt.Sprintf("psh-%s", pshVersion)

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
	tmpDir string
}

// Init the CLI wrapper, creating a temporary directory and copying over files
func (c *LegacyCLIWrapper) Init() error {
	if c.tmpDir != "" {
		return nil
	}

	var err error
	c.tmpDir, err = ioutil.TempDir("", "psh-go")
	if err != nil {
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

// Close the CLI wrapper, removing the temporary directory that was created
func (c *LegacyCLIWrapper) Close() error {
	if c.tmpDir == "" {
		return nil
	}

	if err := os.RemoveAll(c.tmpDir); err != nil {
		return fmt.Errorf("could not remove temporary directory: %w", err)
	}

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

// PSHPath returns the path that the PHP CLI will reside
func (c *LegacyCLIWrapper) PHPPath() string {
	return path.Join(c.tmpDir, phpPath)
}

// PSHPath returns the path that the PSH CLI will reside
func (c *LegacyCLIWrapper) PSHPath() string {
	return path.Join(c.tmpDir, pshPath)
}

// copyPHP to destination, if it does not exist
func (c *LegacyCLIWrapper) copyPHP() error {
	if err := copyFile(c.PHPPath(), phpCLI); err != nil {
		return fmt.Errorf("could not copy PHP CLI: %w", err)
	}

	return nil
}

// copyPSH to destination, if it does not exist
func (c *LegacyCLIWrapper) copyPSH() error {
	if err := copyFile(c.PSHPath(), pshCLI); err != nil {
		return fmt.Errorf("could not copy legacy Platform.sh CLI: %w", err)
	}

	return nil
}
