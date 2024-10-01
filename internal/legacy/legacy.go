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

	"github.com/gofrs/flock"

	"github.com/platformsh/cli/internal/config"
)

//go:embed archives/platform.phar
var phar []byte

var (
	LegacyCLIVersion = "0.0.0"
	PHPVersion       = "0.0.0"
)

var phpPath = fmt.Sprintf("php-%s", PHPVersion)
var pharPath = fmt.Sprintf("phar-%s", LegacyCLIVersion)

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

// fileChanged checks if a file's content differs from the provided bytes.
func fileChanged(filename string, content []byte) (bool, error) {
	stat, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, fmt.Errorf("could not stat file: %w", err)
	}
	if int(stat.Size()) != len(content) {
		return true, nil
	}
	current, err := os.ReadFile(filename)
	if err != nil {
		return false, err
	}
	return !bytes.Equal(current, content), nil
}

// CLIWrapper wraps the legacy CLI
type CLIWrapper struct {
	Stdout             io.Writer
	Stderr             io.Writer
	Stdin              io.Reader
	Config             *config.Config
	Version            string
	CustomPharPath     string
	Debug              bool
	DisableInteraction bool
}

func (c *CLIWrapper) cacheDir() string {
	return path.Join(os.TempDir(), fmt.Sprintf("%s-%s-%s", c.Config.Application.Slug, PHPVersion, LegacyCLIVersion))
}

// Init the CLI wrapper, creating a temporary directory and copying over files
func (c *CLIWrapper) Init() error {
	if _, err := os.Stat(c.cacheDir()); os.IsNotExist(err) {
		c.debugLog("cache directory does not exist, creating: %s", c.cacheDir())
		if err := os.Mkdir(c.cacheDir(), 0o700); err != nil {
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

	if _, err := os.Stat(c.PharPath()); os.IsNotExist(err) {
		if c.CustomPharPath != "" {
			return fmt.Errorf("legacy CLI phar file not found: %w", err)
		}

		c.debugLog("phar file does not exist, copying: %s", c.PharPath())
		if err := copyFile(c.PharPath(), phar); err != nil {
			return fmt.Errorf("could not copy phar file: %w", err)
		}
	}

	// Always write the config.yaml file if it changed.
	configContent, err := config.LoadYAML()
	if err != nil {
		return fmt.Errorf("could not load config for checking: %w", err)
	}
	changed, err := fileChanged(c.ConfigPath(), configContent)
	if err != nil {
		return fmt.Errorf("could not check config file: %w", err)
	}
	if changed {
		if err := copyFile(c.ConfigPath(), configContent); err != nil {
			return fmt.Errorf("could not copy config: %w", err)
		}
	}

	if _, err := os.Stat(c.PHPPath()); os.IsNotExist(err) {
		c.debugLog("PHP binary does not exist, copying: %s", c.PHPPath())
		if err := c.copyPHP(); err != nil {
			return fmt.Errorf("could not copy files: %w", err)
		}
		if err := os.Chmod(c.PHPPath(), 0o700); err != nil {
			return fmt.Errorf("could not make PHP executable: %w", err)
		}
	}

	return nil
}

// Exec a legacy CLI command with the given arguments
func (c *CLIWrapper) Exec(ctx context.Context, args ...string) error {
	args = append([]string{c.PharPath()}, args...)
	cmd := exec.CommandContext(ctx, c.PHPPath(), args...) //nolint:gosec
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
	envPrefix := c.Config.Application.EnvPrefix
	cmd.Env = append(
		cmd.Env,
		"CLI_CONFIG_FILE="+c.ConfigPath(),
		envPrefix+"UPDATES_CHECK=0",
		envPrefix+"MIGRATE_CHECK=0",
		envPrefix+"APPLICATION_PROMPT_SELF_INSTALL=0",
		envPrefix+"WRAPPED=1",
		envPrefix+"APPLICATION_VERSION="+c.Version,
	)
	if c.Debug {
		cmd.Env = append(cmd.Env, envPrefix+"CLI_DEBUG=1")
	}
	if c.DisableInteraction {
		cmd.Env = append(cmd.Env, envPrefix+"NO_INTERACTION=1")
	}
	cmd.Env = append(cmd.Env, fmt.Sprintf(
		"%sUSER_AGENT={APP_NAME_DASH}/%s ({UNAME_S}; {UNAME_R}; PHP %s; WRAPPER %s)",
		envPrefix,
		LegacyCLIVersion,
		PHPVersion,
		c.Version,
	))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not run legacy CLI command: %w", err)
	}

	return nil
}

// PharPath returns the path to the legacy CLI's Phar file.
func (c *CLIWrapper) PharPath() string {
	if c.CustomPharPath != "" {
		return c.CustomPharPath
	}

	return path.Join(c.cacheDir(), pharPath)
}

// ConfigPath returns the path to the YAML config file that will be provided to the legacy CLI.
func (c *CLIWrapper) ConfigPath() string {
	return path.Join(c.cacheDir(), "config.yaml")
}

// debugLog logs a debugging message, if debug is enabled
func (c *CLIWrapper) debugLog(msg string, v ...any) {
	if c.Debug {
		log.Printf(msg, v...)
	}
}
