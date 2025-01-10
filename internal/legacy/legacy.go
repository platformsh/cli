package legacy

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gofrs/flock"
	"golang.org/x/sync/errgroup"

	"github.com/platformsh/cli/internal/config"
	"github.com/platformsh/cli/internal/file"
)

//go:embed archives/platform.phar
var phar []byte

var (
	LegacyCLIVersion = "0.0.0"
	PHPVersion       = "0.0.0"
)

const (
	pharBasename   = "legacy-cli.phar"
	configBasename = "config.yaml"
)

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

	_cacheDir string
}

func (c *CLIWrapper) cacheDir() (string, error) {
	if c._cacheDir == "" {
		cd, err := c.Config.CacheDir()
		if err != nil {
			return "", err
		}
		cd = filepath.Join(cd, fmt.Sprintf("legacy-%s-%s", PHPVersion, LegacyCLIVersion))
		if err := os.Mkdir(cd, 0o700); err != nil && !errors.Is(err, fs.ErrExist) {
			return "", err
		}
		c._cacheDir = cd
	}

	return c._cacheDir, nil
}

// Initialize the CLI wrapper, creating a temporary directory and copying over files.
func (c *CLIWrapper) init() error {
	cacheDir, err := c.cacheDir()
	if err != nil {
		return err
	}

	fileLock := flock.New(filepath.Join(cacheDir, ".lock"))
	if err := fileLock.Lock(); err != nil {
		return fmt.Errorf("could not acquire lock: %w", err)
	}
	c.debugLog("lock acquired: %s", fileLock.Path())
	defer fileLock.Unlock() //nolint:errcheck

	g := errgroup.Group{}
	g.Go(func() error {
		if err := file.CopyIfChanged(c.pharPath(cacheDir), phar, 0o644); err != nil {
			return fmt.Errorf("could not copy phar file: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		configContent, err := config.LoadYAML()
		if err != nil {
			return fmt.Errorf("could not load config for checking: %w", err)
		}
		if err := file.CopyIfChanged(filepath.Join(cacheDir, configBasename), configContent, 0o644); err != nil {
			return fmt.Errorf("could not write config: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		return c.copyPHP(cacheDir)
	})

	return g.Wait()
}

// Exec a legacy CLI command with the given arguments
func (c *CLIWrapper) Exec(ctx context.Context, args ...string) error {
	if err := c.init(); err != nil {
		return fmt.Errorf("failed to initialize CLI: %w", err)
	}
	cacheDir, err := c.cacheDir()
	if err != nil {
		return err
	}
	cmd := c.makeCmd(ctx, args, cacheDir)
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
		"CLI_CONFIG_FILE="+filepath.Join(cacheDir, configBasename),
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

// makeCmd makes a legacy CLI command with the given context and arguments.
func (c *CLIWrapper) makeCmd(ctx context.Context, args []string, cacheDir string) *exec.Cmd {
	iniSettings := c.phpSettings(cacheDir)
	var cmdArgs = make([]string, 0, len(args)+2+len(iniSettings)*2)
	for _, s := range iniSettings {
		cmdArgs = append(cmdArgs, "-d", s)
	}
	cmdArgs = append(cmdArgs, c.pharPath(cacheDir))
	cmdArgs = append(cmdArgs, args...)
	return exec.CommandContext(ctx, c.phpPath(cacheDir), cmdArgs...) //nolint:gosec
}

// PharPath returns the path to the legacy CLI's Phar file.
func (c *CLIWrapper) PharPath() (string, error) {
	if c.CustomPharPath != "" {
		return c.CustomPharPath, nil
	}
	cacheDir, err := c.cacheDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(cacheDir, pharBasename), nil
}

func (c *CLIWrapper) pharPath(cacheDir string) string {
	if c.CustomPharPath != "" {
		return c.CustomPharPath
	}

	return filepath.Join(cacheDir, pharBasename)
}

// debugLog logs a debugging message, if debug is enabled
func (c *CLIWrapper) debugLog(msg string, v ...any) {
	if c.Debug {
		log.Printf(msg, v...)
	}
}
