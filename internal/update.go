package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-isatty"

	"github.com/platformsh/cli/internal/config"
	"github.com/platformsh/cli/internal/state"
)

var versionRegex = regexp.MustCompile(`^(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)(-(?P<preRelease>.+))?$`)

// ReleaseInfo stores information about a release
type ReleaseInfo struct {
	Version     string    `json:"tag_name"`
	URL         string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
}

// Version contains parsed information about a SemVer version
type Version struct {
	VersionParts    [3]int
	PreReleaseParts []string
}

// CompareVersions and see which version is greater
func CompareVersions(a, b *Version) int {
	// Compare Major, Minor and Patch versions
	for i := 0; i < 3; i++ {
		if a.VersionParts[i] > b.VersionParts[i] {
			return 1
		}
		if a.VersionParts[i] < b.VersionParts[i] {
			return -1
		}
	}

	// Start comparing identifiers
	for i := 0; ; i++ {
		// Check that there are identifiers left
		if len(a.PreReleaseParts) <= i && len(b.PreReleaseParts) <= i {
			return 0
		}

		// Shorter takes precedence
		if len(b.PreReleaseParts) <= i {
			return 1
		}
		if len(a.PreReleaseParts) <= i {
			return -1
		}

		aPart := a.PreReleaseParts[i]
		bPart := b.PreReleaseParts[i]
		aInt, aErr := strconv.Atoi(aPart)
		bInt, bErr := strconv.Atoi(bPart)

		// Try comparing integers first
		if aErr == nil && bErr == nil {
			if aInt > bInt {
				return 1
			}
			if aInt < bInt {
				return -1
			}
			// Integer wins string
		} else if aErr == nil {
			return 1
		} else if bErr == nil {
			return -1
			// Compare strings
		} else if cmp := strings.Compare(aPart, bPart); cmp != 0 {
			return cmp
		}
	}
}

// ParseVersion from a string, returning a Version or error if it's not SemVer
func ParseVersion(version string) (*Version, error) {
	if !versionRegex.MatchString(version) {
		return nil, fmt.Errorf("version does not match SemVer: %s", version)
	}

	result := versionRegex.FindStringSubmatch(version)
	major, _ := strconv.Atoi(result[versionRegex.SubexpIndex("major")])
	minor, _ := strconv.Atoi(result[versionRegex.SubexpIndex("minor")])
	patch, _ := strconv.Atoi(result[versionRegex.SubexpIndex("patch")])
	preRelease := result[versionRegex.SubexpIndex("preRelease")]
	var preReleaseParts []string
	if preRelease != "" {
		preReleaseParts = strings.Split(preRelease, ".")
	}

	return &Version{
		VersionParts:    [3]int{major, minor, patch},
		PreReleaseParts: preReleaseParts,
	}, nil
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

	currentVersionParsed, err := ParseVersion(currentVersion)
	if err != nil {
		return nil, err
	}

	latestVersionParsed, err := ParseVersion(releaseInfo.Version)
	if err != nil {
		return nil, err
	}
	if CompareVersions(latestVersionParsed, currentVersionParsed) == 1 {
		return releaseInfo, nil
	}

	return nil, nil
}

// shouldCheckForUpdate checks updates are not disabled and the environment is a terminal
func shouldCheckForUpdate(cnf *config.Config) bool {
	if cnf.Wrapper.GitHubRepo == "" || !cnf.Updates.Check || os.Getenv(cnf.Application.EnvPrefix+"UPDATES_CHECK") == "0" {
		return false
	}

	return !isCI() && isTerminal(os.Stdout) && isTerminal(os.Stderr)
}

func isCI() bool {
	return os.Getenv("CI") != "" || // GitHub Actions, Travis CI, CircleCI, Cirrus CI, GitLab CI, AppVeyor, CodeShip, dsari
		os.Getenv("BUILD_NUMBER") != "" || // Jenkins, TeamCity
		os.Getenv("RUN_ID") != "" // TaskCluster, dsari
}

func isTerminal(f *os.File) bool {
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
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
