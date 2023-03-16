package internal

import (
	"encoding/json"
	"os"
	"path"
)

type stateEntry struct {
	Updates struct {
		LastChecked   int          `json:"last_checked"`
		LatestRelease *ReleaseInfo `json:"latest_release"`
	} `json:"updates"`
}

func getState(statePath string) (*stateEntry, error) {
	data, err := os.ReadFile(statePath)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	state := &stateEntry{}
	if err := json.Unmarshal(data, state); err != nil {
		return nil, err
	}

	return state, nil
}

func setState(statePath string, state *stateEntry) error {
	if err := os.MkdirAll(path.Dir(statePath), 0o755); err != nil {
		return err
	}

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return os.WriteFile(statePath, data, 0o600)
}
