package legacy

import (
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
	"github.com/platformsh/cli/internal/file"
)

//go:embed archives/platform.phar
var phar []byte

var (
	LegacyCLIVersion = "0.0.0"
	PHPVersion       = "0.0.0"
)

var phpPath = "php"
var pharPath = "legacy-cli.phar"

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
	return path.Join(os.TempDir(), c.Config.Application.Slug)
}

// Initialize the CLI wrapper, creating a temporary directory and copying over files.
func (c *CLIWrapper) init() error {
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
	defer fileLock.Unlock() //nolint:errcheck

	if err := file.CopyIfChanged(c.PharPath(), phar, 0o644); err != nil {
		return fmt.Errorf("could not copy phar file: %w", err)
	}

	// Always write the config.yaml file if it changed.
	configContent, err := config.LoadYAML()
	if err != nil {
		return fmt.Errorf("could not load config for checking: %w", err)
	}
	if err := file.CopyIfChanged(c.ConfigPath(), configContent, 0o644); err != nil {
		return fmt.Errorf("could not write config: %w", err)
	}

	return c.copyPHP()
}

// Exec a legacy CLI command with the given arguments
func (c *CLIWrapper) Exec(ctx context.Context, args ...string) error {
	cmd := c.makeCmd(ctx, args)
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
		// Cleanup cache directory
		c.debugLog("removing cache directory: %s", c.cacheDir())
		os.RemoveAll(c.cacheDir())
		return fmt.Errorf("could not run legacy CLI command: %w", err)
	}

	return nil
}

// makeCmd makes a legacy CLI command with the given context and arguments.
func (c *CLIWrapper) makeCmd(ctx context.Context, args []string) *exec.Cmd {
	iniSettings := c.phpSettings()
	var cmdArgs = make([]string, 0, len(args)+2+len(iniSettings)*2)
	for _, s := range iniSettings {
		cmdArgs = append(cmdArgs, "-d", s)
	}
	cmdArgs = append(cmdArgs, c.PharPath())
	cmdArgs = append(cmdArgs, args...)
	return exec.CommandContext(ctx, c.PHPPath(), cmdArgs...) //nolint:gosec
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
