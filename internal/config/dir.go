package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

// TempDir returns the path to a user-specific temporary directory, suitable for caches.
//
// It creates the temporary directory if it does not already exist.
//
// The directory can be specified in the {ENV_PREFIX}TMP environment variable.
//
// This does not use os.TempDir, as on Linux/Unix systems that usually returns a
// global /tmp directory, which could conflict with other users. It also does not
// use os.MkdirTemp, as the CLI usually needs a stable (not random) directory
// path. It therefore uses os.UserCacheDir which in turn will use XDG_CACHE_HOME
// or the home directory.
func (c *Config) TempDir() (string, error) {
	if c.tempDir != "" {
		return c.tempDir, nil
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
	c.tempDir = path

	return path, nil
}

// WritableUserDir returns the path to a writable user-level directory.
// Deprecated: unless backwards compatibility is desired, TempDir is preferable.
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

// HomeDir returns the user's home directory, which can be overridden with the {ENV_PREFIX}HOME variable.
func (c *Config) HomeDir() (string, error) {
	d := os.Getenv(c.Application.EnvPrefix + "HOME")
	if d != "" {
		return d, nil
	}
	return os.UserHomeDir()
}
