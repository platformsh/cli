package legacy

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/gofrs/flock"
	"golang.org/x/sync/errgroup"

	"github.com/upsun/cli/internal/config"
	"github.com/upsun/cli/internal/file"
)

//go:embed archives/platform.phar
var phar []byte

var (
	LegacyCLIVersion = "0.0.0"
	PHPVersion       = "0.0.0"
)

const configBasename = "config.yaml"

// CLIWrapper wraps the legacy CLI
type CLIWrapper struct {
	Stdout             io.Writer
	Stderr             io.Writer
	Stdin              io.Reader
	Config             *config.Config
	Version            string
	Debug              bool
	DisableInteraction bool
	ForceColor         bool
	DebugLogFunc       func(string, ...any)

	initOnce  sync.Once
	_cacheDir string
}

func (c *CLIWrapper) debug(msg string, args ...any) {
	if c.DebugLogFunc != nil {
		c.DebugLogFunc(msg, args...)
	}
}

func (c *CLIWrapper) cacheDir() (string, error) {
	if c._cacheDir == "" {
		cd, err := c.Config.TempDir()
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

// runInitOnce runs the init method, only once for this object.
func (c *CLIWrapper) runInitOnce() error {
	var err error
	c.initOnce.Do(func() { err = c.init() })
	return err
}

// init initializes the CLI wrapper, creating a temporary directory and copying over files.
func (c *CLIWrapper) init() error {
	preInit := time.Now()

	cacheDir, err := c.cacheDir()
	if err != nil {
		return err
	}

	preLock := time.Now()
	fileLock := flock.New(filepath.Join(cacheDir, ".lock"))
	if err := fileLock.Lock(); err != nil {
		return fmt.Errorf("could not acquire lock: %w", err)
	}
	c.debug("lock acquired (%s): %s", time.Since(preLock), fileLock.Path())
	defer fileLock.Unlock() //nolint:errcheck

	g := errgroup.Group{}
	g.Go(func() error {
		if err := file.WriteIfNeeded(c.pharPath(cacheDir), phar, 0o644); err != nil {
			return fmt.Errorf("could not copy phar file: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		configContent, err := c.Config.Raw()
		if err != nil {
			return fmt.Errorf("could not load config for checking: %w", err)
		}
		if err := file.WriteIfNeeded(filepath.Join(cacheDir, configBasename), configContent, 0o644); err != nil {
			return fmt.Errorf("could not write config: %w", err)
		}
		return nil
	})

	g.Go(newPHPManager(cacheDir).copy)

	if err := g.Wait(); err != nil {
		return err
	}

	c.debug("Initialized PHP CLI (%s)", time.Since(preInit))

	return nil
}

// Exec a legacy CLI command with the given arguments
func (c *CLIWrapper) Exec(ctx context.Context, args ...string) error {
	if err := c.runInitOnce(); err != nil {
		return fmt.Errorf("failed to initialize PHP CLI: %w", err)
	}
	cacheDir, err := c.cacheDir()
	if err != nil {
		return err
	}
	cmd := c.makeCmd(ctx, args, cacheDir)
	cmd.Stdin = c.Stdin
	cmd.Stdout = c.Stdout
	cmd.Stderr = c.Stderr
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
	if c.DisableInteraction {
		cmd.Env = append(cmd.Env, envPrefix+"NO_INTERACTION=1")
	}
	if c.ForceColor {
		cmd.Env = append(cmd.Env, "CLICOLOR_FORCE=1")
	}
	cmd.Env = append(cmd.Env, fmt.Sprintf(
		"%sUSER_AGENT={APP_NAME_DASH}/%s ({UNAME_S}; {UNAME_R}; PHP %s; WRAPPER %s)",
		envPrefix,
		LegacyCLIVersion,
		PHPVersion,
		c.Version,
	))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not run PHP CLI command: %w", err)
	}

	return nil
}

// makeCmd makes a legacy CLI command with the given context and arguments.
func (c *CLIWrapper) makeCmd(ctx context.Context, args []string, cacheDir string) *exec.Cmd {
	phpMgr := newPHPManager(cacheDir)
	settings := phpMgr.settings()
	var cmdArgs = make([]string, 0, len(args)+2+len(settings)*2)
	for _, s := range settings {
		cmdArgs = append(cmdArgs, "-d", s)
	}
	cmdArgs = append(cmdArgs, c.pharPath(cacheDir))
	cmdArgs = append(cmdArgs, args...)
	return exec.CommandContext(ctx, phpMgr.binPath(), cmdArgs...) //nolint:gosec
}

// PharPath returns the path to the legacy CLI's Phar file.
func (c *CLIWrapper) PharPath() (string, error) {
	cacheDir, err := c.cacheDir()
	if err != nil {
		return "", err
	}

	return c.pharPath(cacheDir), nil
}

func (c *CLIWrapper) pharPath(cacheDir string) string {
	if customPath := os.Getenv(c.Config.Application.EnvPrefix + "PHAR_PATH"); customPath != "" {
		return customPath
	}

	return filepath.Join(cacheDir, c.Config.Application.Executable+".phar")
}
