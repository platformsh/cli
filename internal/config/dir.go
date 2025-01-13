package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

// HomeDir returns the user's home directory, which can be overridden with the {ENV_PREFIX}HOME variable.
func (c *Config) HomeDir() (string, error) {
	d := os.Getenv(c.Application.EnvPrefix + "HOME")
	if d != "" {
		return d, nil
	}
	return os.UserHomeDir()
}

// WritableUserDir returns the path to a writable user-level directory.
func (c *Config) WritableUserDir() (string, error) {
	if c.writableUserDir != "" {
		return c.writableUserDir, nil
	}
	hd, err := c.HomeDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(hd, c.Application.WritableUserDir)
	if err := os.MkdirAll(path, 0o700); err != nil {
		return "", err
	}
	c.writableUserDir = path

	return path, nil
}

// TempDir returns the path to a user-specific temporary directory.
//
// It creates the temporary directory if it does not already exist.
//
// It does not use os.TempDir on Linux as that usually returns /tmp which could
// conflict with other users. It also does not use os.MkdirTemp as the CLI
// usually needs a stable (not random) directory path. It therefore uses
// os.UserCacheDir which in turn will use XDG_CACHE_HOME or the home directory.
func (c *Config) TempDir() (string, error) {
	if c.cacheDir != "" {
		return c.cacheDir, nil
	}
	d := os.Getenv(c.Application.EnvPrefix + "TMP")
	if d == "" {
		ucd, err := os.UserCacheDir()
		if err != nil {
			return "", err
		}
		d = ucd
	}

	// Windows already has a user-specific temporary directory.
	if runtime.GOOS == "windows" {
		osTemp := os.TempDir()
		if strings.HasPrefix(osTemp, d) {
			d = osTemp
		}
	}

	path := filepath.Join(d, c.Application.TempSubDir)

	// If the subdirectory cannot be created due to a read-only filesystem, fall back to /tmp.
	if err := os.MkdirAll(path, 0o700); err != nil {
		if !errors.Is(err, syscall.EROFS) {
			return "", err
		}
		path = filepath.Join(os.TempDir(), c.Application.TempSubDir)
		if err := os.MkdirAll(path, 0o700); err != nil {
			return "", err
		}
	}
	c.cacheDir = path

	return path, nil
}
