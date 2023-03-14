package legacy

import (
	"bytes"

	"github.com/gofrs/flock"

	"context"
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

//go:embed archives/platform.phar
var pshCLI []byte

var PSHVersion string = "0.0.0"
var PHPVersion string = "0.0.0"

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
	Stdout           io.Writer
	Stderr           io.Writer
	Stdin            io.Reader
	Version          string
	CustomPshCliPath string
	Debug            bool
}

func (c *LegacyCLIWrapper) cacheDir() string {
	return path.Join(os.TempDir(), fmt.Sprintf("%s-%s-%s", prefix, PHPVersion, PSHVersion))
}

// Init the CLI wrapper, creating a temporary directory and copying over files
func (c *LegacyCLIWrapper) Init() error {
	if _, err := os.Stat(c.cacheDir()); os.IsNotExist(err) {
		c.debugLog("cache directory does not exist, creating: %s", c.cacheDir())
		if err := os.Mkdir(c.cacheDir(), 0700); err != nil {
			return fmt.Errorf("could not create temporary directory: %w", err)
		}
	}
	fileLock := flock.New(path.Join(c.cacheDir(), ".lock"))
	if err := fileLock.Lock(); err != nil {
		return fmt.Errorf("could not acquire lock: %w", err)
	}
	c.debugLog("lock acquired: %s", fileLock.Path())
	//nolint:errcheck
	defer fileLock.Unlock()

	if _, err := os.Stat(c.PSHPath()); os.IsNotExist(err) {
		if c.CustomPshCliPath != "" {
			return fmt.Errorf("given PSH phar path does not exist: %w", err)
		}

		c.debugLog("PSH .phar file does not exist, copying: %s", c.PSHPath())
		if err := c.copyPSH(); err != nil {
			return fmt.Errorf("could not copy files: %w", err)
		}
	}

	if _, err := os.Stat(c.PHPPath()); os.IsNotExist(err) {
		c.debugLog("PHP binary does not exist, copying: %s", c.PHPPath())
		if err := c.copyPHP(); err != nil {
			return fmt.Errorf("could not copy files: %w", err)
		}
		if err := os.Chmod(c.PHPPath(), 0700); err != nil {
			return fmt.Errorf("could not make PHP executable: %w", err)
		}
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
			if err != nil && c.Debug {
				c.debugLog("could not remove directory: %s", f.Name())
			}
		}
	}

	return nil
}

// Exec a legacy CLI command with the given arguments
func (c *LegacyCLIWrapper) Exec(ctx context.Context, args ...string) error {
	args = append([]string{c.PSHPath()}, args...)
	cmd := exec.CommandContext(ctx, c.PHPPath(), args...)
	if c.Stdin != nil {
		cmd.Stdin = c.Stdin
	} else {
		cmd.Stdin = os.Stdin
	}
	if c.Stdout != nil {
		cmd.Stdout = c.Stdout
	} else {
		cmd.Stdout = os.Stdout
	}
	if c.Stderr != nil {
		cmd.Stderr = c.Stderr
	} else {
		cmd.Stderr = os.Stderr
	}
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, "PLATFORMSH_CLI_UPDATES_CHECK=0")
	cmd.Env = append(cmd.Env, "PLATFORMSH_CLI_MIGRATE_CHECK=0")
	cmd.Env = append(cmd.Env, "PLATFORMSH_CLI_APPLICATION_PROMPT_SELF_INSTALL=0")
	cmd.Env = append(cmd.Env, "PLATFORMSH_CLI_WRAPPED=1")
	if c.Debug {
		cmd.Env = append(cmd.Env, "PLATFORMSH_CLI_DEBUG=1")
	}
	cmd.Env = append(cmd.Env, fmt.Sprintf(
		"PLATFORMSH_CLI_USER_AGENT={APP_NAME_DASH}/%s ({UNAME_S}; {UNAME_R}; PHP %s; WRAPPER psh-go/%s)",
		PSHVersion,
		PHPVersion,
		c.Version,
	))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not run legacy CLI command: %w", err)
	}

	return nil
}

// PSHPath returns the path that the PSH CLI will reside
func (c *LegacyCLIWrapper) PSHPath() string {
	if c.CustomPshCliPath != "" {
		return c.CustomPshCliPath
	}

	return path.Join(c.cacheDir(), pshPath)
}

// copyPSH to destination, if it does not exist
func (c *LegacyCLIWrapper) copyPSH() error {
	// Do not copy the file, if a custom path was given
	if c.CustomPshCliPath != "" {
		return nil
	}

	if err := copyFile(c.PSHPath(), pshCLI); err != nil {
		return fmt.Errorf("could not copy legacy Platform.sh CLI: %w", err)
	}

	return nil
}

// debugLog logs a debugging message, if debug is enabled
func (c *LegacyCLIWrapper) debugLog(msg string, v ...any) {
	if c.Debug {
		log.Printf(msg, v...)
	}
}
