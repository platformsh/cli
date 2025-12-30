package state

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/upsun/cli/internal/config"
)

type State struct {
	Updates struct {
		LastChecked int64 `json:"last_checked"`
	} `json:"updates,omitempty"`

	ConfigUpdates struct {
		LastChecked int64 `json:"last_checked"`
	} `json:"config_updates,omitempty"`
}

// Load reads state from the filesystem.
func Load(cnf *config.Config) (state State, err error) {
	statePath, err := getPath(cnf)
	if err != nil {
		return
	}
	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return
	}
	err = json.Unmarshal(data, &state)
	return
}

// Save writes state to the filesystem.
func Save(state State, cnf *config.Config) error {
	statePath, err := getPath(cnf)
	if err != nil {
		return err
	}

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return os.WriteFile(statePath, data, 0o600)
}

// getPath determines the path to the state JSON file depending on config.
func getPath(cnf *config.Config) (string, error) {
	writableDir, err := cnf.WritableUserDir() //nolint:staticcheck // backwards compatibility is needed for state files
	if err != nil {
		return "", err
	}

	return filepath.Join(writableDir, cnf.Application.UserStateFile), nil
}
