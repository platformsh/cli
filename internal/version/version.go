package version

import "github.com/Masterminds/semver/v3"

// Compare parses and compares two semantic version numbers.
// It returns -1, 0 or 1, representing whether v1 is less than, equal to or greater than v2.
func Compare(v1, v2 string) (int, error) {
	if v1 == v2 {
		return 0, nil
	}
	version1, err := semver.NewVersion(v1)
	if err != nil {
		return 0, err
	}
	version2, err := semver.NewVersion(v2)
	if err != nil {
		return 0, err
	}
	return version1.Compare(version2), nil
}

// Validate tests if a version number is valid.
func Validate(v string) bool {
	_, err := semver.NewVersion(v)
	return err == nil
}
