package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/symfony-cli/terminal"

	"github.com/platformsh/cli/internal/config"
	"github.com/platformsh/cli/internal/state"
	"github.com/platformsh/cli/internal/version"
)

// ReleaseInfo stores information about a release
type ReleaseInfo struct {
	Version     string    `json:"tag_name"`
	URL         string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
}

// CheckForUpdate checks whether this software has had a newer release on GitHub
func CheckForUpdate(cnf *config.Config, currentVersion string) (*ReleaseInfo, error) {
	if !shouldCheckForUpdate(cnf) {
		return nil, nil
	}

	s, err := state.Load(cnf)
	if err == nil && time.Now().Unix()-s.Updates.LastChecked < int64(cnf.Updates.CheckInterval) {
		// Updates were already checked recently.
		return nil, nil
	}

	defer func() {
		// After checking, save the last check time.
		s.Updates.LastChecked = time.Now().Unix()
		//nolint:errcheck // not being able to set the state should have no impact on the rest of the program
		state.Save(s, cnf)
	}()

	releaseInfo, err := getLatestReleaseInfo(cnf.Wrapper.GitHubRepo)
	if err != nil {
		return nil, fmt.Errorf("could not determine latest release: %w", err)
	}

	cmp, err := version.Compare(releaseInfo.Version, currentVersion)
	if err != nil {
		return nil, fmt.Errorf("could not compare versions: %w", err)
	}
	if cmp > 0 {
		return releaseInfo, nil
	}

	return nil, nil
}

// shouldCheckForUpdate checks updates are not disabled and the environment is a terminal
func shouldCheckForUpdate(cnf *config.Config) bool {
	return config.Version != "0.0.0" &&
		cnf.Wrapper.GitHubRepo != "" &&
		cnf.Updates.Check &&
		os.Getenv(cnf.Application.EnvPrefix+"UPDATES_CHECK") != "0" &&
		!isCI() && terminal.IsTerminal(os.Stdout) && terminal.IsTerminal(os.Stderr)
}

func isCI() bool {
	return os.Getenv("CI") != "" || // GitHub Actions, Travis CI, CircleCI, Cirrus CI, GitLab CI, AppVeyor, CodeShip, dsari
		os.Getenv("BUILD_NUMBER") != "" || // Jenkins, TeamCity
		os.Getenv("RUN_ID") != "" // TaskCluster, dsari
}

// getLatestReleaseInfo from GitHub
func getLatestReleaseInfo(repo string) (*ReleaseInfo, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo), http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var latestRelease ReleaseInfo
	if err := json.Unmarshal(body, &latestRelease); err != nil {
		return nil, err
	}

	return &latestRelease, nil
}
