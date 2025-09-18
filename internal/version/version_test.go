package version_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/platformsh/cli/internal/version"
)

func TestCompare(t *testing.T) {
	cases := []struct {
		v1   string
		v2   string
		cmp  int
		fail bool
	}{
		// Normal comparisons
		{v1: "1.0.0", v2: "1.0.0"},
		{v1: "1.0.1", v2: "1.0.0", cmp: 1},
		{v1: "2.1.0", v2: "2.1.1", cmp: -1},
		{v1: "0.0.0", v2: "0.0.0"},
		{v1: "0.1.0", v2: "0.0.1", cmp: 1},
		{v1: "0.0.1", v2: "0.0.2", cmp: -1},
		{v1: "1.0.0", v2: "", fail: true},

		// Quasi-semver
		{v1: "1", v2: "1"},
		{v1: "1", v2: "2", cmp: -1},
		{v1: "2", v2: "1", cmp: 1},
		{v1: "1", v2: "0.0.1", cmp: 1},
		{v1: "1.0", v2: "1.0"},
		{v1: "1.0", v2: "2.0", cmp: -1},
		{v1: "1.0.1", v2: "2", cmp: -1},
		{v1: "1.0.0", v2: "1.0"},
		{v1: "1.01", v2: "1.2", cmp: -1},

		// Suffixes
		{v1: "1.0.1-dev", v2: "1.0.1", cmp: -1},
		{v1: "1.0.1+context", v2: "1.0.1"},
		{v1: "3.0-beta.1", v2: "3.0-beta.2", cmp: -1},
		{v1: "3.0-beta.3", v2: "3.0-beta.2", cmp: 1},
		{v1: "3.0.1-beta.3", v2: "3.0.1-beta.2", cmp: 1},
		{v1: "3.0.1-beta.3", v2: "3.0.2-beta.2", cmp: -1},
		{v1: "1.0.0-beta.9", v2: "1.0.0-beta.10", cmp: -1},
		{v1: "1.0.0-beta.10", v2: "1.0.0-beta.2", cmp: 1},
		{v1: "1.0.0-beta.1", v2: "1.0.0-alpha.1", cmp: 1},
		{v1: "1.0.0+001", v2: "1.0.0+002"},
		{v1: "3.0-beta.03", v2: "3.0-beta.2", fail: true},
		{v1: "1.0.1_invalid", v2: "1.0.1", fail: true},
		{v1: "1.0.1", v2: "2.0.0_invalid", fail: true},

		// Prefixes
		{v1: "v1.0.1", v2: "1.0.1"},
		{v1: "v1.0.1", v2: "1.0.2", cmp: -1},
		{v1: "v1.0.2", v2: "v1.0.1", cmp: 1},
	}
	for _, c := range cases {
		comment := c.v1 + " <=> " + c.v2
		result, err := version.Compare(c.v1, c.v2)
		if c.fail {
			assert.Error(t, err, comment)
		} else {
			assert.NoError(t, err, comment)
			assert.Equal(t, c.cmp, result, comment)
		}
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		v     string
		valid bool
	}{
		{"0", true},
		{"1", true},
		{"v0", true},
		{"1.0.0", true},
		{"v1.0.0", true},
		{"1.0.0+build-info", true},
		{"1.0.0-dev-suffix", true},
		{"1.0.0-dev-suffix.with.numbers.1", true},
		{"v1.0.0-dev-suffix.with.numbers.1", true},
		{"2024.0.1", true},

		{"1.01", true},
	}
	for _, c := range cases {
		assert.Equal(t, c.valid, version.Validate(c.v), c.v)
	}
}
