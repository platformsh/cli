package alt

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
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

	var candidates []string
	if runtime.GOOS == "windows" {
		candidates = []string{
			filepath.Join(homeDir, "AppData", "Local", "Programs"),
			filepath.Join(homeDir, ".local", "bin"),
			filepath.Join(homeDir, "bin"),
		}
	} else {
		candidates = []string{
			filepath.Join(homeDir, ".local", "bin"),
			filepath.Join(homeDir, "bin"),
		}
	}

	// Use the first candidate that is in the PATH.
	pathValue := os.Getenv("PATH")
	for _, c := range candidates {
		if inPathValue(c, pathValue) {
			return c, nil
		}
	}

	return filepath.Join(homeDir, homeSubDir, "bin"), nil
}

// isExistingDirectory checks if a path exists and is a directory.
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

// writeFile creates or overwrites a file.
// If dirMode is not 0, the containing directory will be created, if it does not already exist.
func writeFile(path string, content []byte, dirMode, fileMode fs.FileMode) error {
	if dirMode != 0 {
		if err := os.MkdirAll(filepath.Dir(path), dirMode); err != nil {
			return err
		}
	}

	tmpFile := path + ".tmp"
	if err := os.WriteFile(tmpFile, content, fileMode); err != nil {
		return err
	}

	return os.Rename(tmpFile, path)
}
