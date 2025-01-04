package alt

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	subDir     = "platform-alt"
	homeSubDir = ".platform-alt"
)

// FindConfigDir finds an appropriate destination directory for an "alt" CLI configuration YAML file.
func FindConfigDir() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	isDir, err := isExistingDirectory(userConfigDir)
	if err != nil {
		return "", err
	}
	if isDir {
		return filepath.Join(userConfigDir, subDir), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, homeSubDir), nil
}

// FindBinDir finds an appropriate destination directory for an "alt" CLI executable.
func FindBinDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}

	var binDir string
	if runtime.GOOS == "windows" {
		binDir = filepath.Join(homeDir, "AppData", "Local", "Programs")
	} else {
		binDir = filepath.Join(homeDir, ".local", "bin")
	}

	isDir, err := isExistingDirectory(binDir)
	if err != nil {
		return "", err
	}
	if isDir {
		return binDir, nil
	}

	return filepath.Join(homeDir, homeSubDir, "bin"), nil
}

// InPath tests if a directory is in the PATH.
func InPath(dir string) (bool, error) {
	normalized, err := normalize(dir)
	if err != nil {
		return false, err
	}
	pathEnv := os.Getenv("PATH")
	pathDirs := strings.Split(pathEnv, string(os.PathListSeparator))
	for _, pathDir := range pathDirs {
		pathNormalized, _err := normalize(pathDir)
		if _err != nil {
			if errors.Is(_err, fs.ErrNotExist) {
				continue
			}
			return false, _err
		}
		if pathNormalized == normalized {
			return true, nil
		}
	}
	return false, nil
}

func isExistingDirectory(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return stat.IsDir(), nil
}

func normalize(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}
	return resolved, nil
}
