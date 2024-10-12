// Package tests contains integration tests, which run the CLI as a shell command and verify its output.
//
// A TEST_CLI_PATH environment variable can be provided to override the path to a
// CLI executable. It defaults to `platform` in the repository root.
package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/platformsh/cli/internal/mockapi"
)

var _validatedCommand string

// The legacy CLI identifier expects project IDs to be alphanumeric.
// See: https://github.com/platformsh/legacy-cli/blob/main/src/Service/Identifier.php#L75
const mockProjectID = "abcdefg123456"

func getCommandName(t *testing.T) string {
	if _validatedCommand != "" {
		return _validatedCommand
	}
	candidate := os.Getenv("TEST_CLI_PATH")
	if candidate == "" {
		candidate = "platform"
	}
	if !filepath.IsAbs(candidate) {
		c, err := filepath.Abs("../" + candidate)
		require.NoError(t, err)
		candidate = c
	}
	_, err := os.Stat(candidate)
	switch {
	case os.IsNotExist(err) && os.Getenv("TEST_CLI_PATH") == "":
		t.Skipf("skipping integration tests: CLI not found at path: %s", candidate)
	case err != nil:
		require.NoError(t, err)
	case testing.Short():
		t.Skip("skipping integration test due to -short flag")
	}
	versionCmd := exec.Command(candidate, "--version")
	versionCmd.Env = testEnv()
	output, err := versionCmd.Output()
	require.NoError(t, err, "running '--version' must succeed under the CLI at: %s", candidate)
	require.Contains(t, string(output), "Platform Test CLI ")
	t.Logf("Validated CLI command %s", candidate)
	_validatedCommand = candidate
	return _validatedCommand
}

func command(t *testing.T, args ...string) *exec.Cmd {
	cmd := exec.Command(getCommandName(t), args...) //nolint:gosec
	cmd.Env = testEnv()
	cmd.Dir = os.TempDir()
	if testing.Verbose() {
		cmd.Stderr = os.Stderr
	}
	return cmd
}

func authenticatedCommand(t *testing.T, apiURL, authURL string, args ...string) *exec.Cmd {
	cmd := command(t, args...)
	cmd.Env = append(
		cmd.Env,
		EnvPrefix+"API_BASE_URL="+apiURL,
		EnvPrefix+"API_AUTH_URL="+authURL,
		EnvPrefix+"TOKEN="+mockapi.ValidAPITokens[0],
	)
	return cmd
}

// runnerWithAuth returns a function to authenticate and run a CLI command, returning stdout output.
func runnerWithAuth(t *testing.T, apiURL, authURL string) func(args ...string) string {
	return func(args ...string) string {
		cmd := authenticatedCommand(t, apiURL, authURL, args...)
		b, err := cmd.Output()
		require.NoError(t, err)
		return string(b)
	}
}

const EnvPrefix = "TEST_CLI_"

func testEnv() []string {
	configPath, err := filepath.Abs("config.yaml")
	if err != nil {
		panic(err)
	}
	return append(
		os.Environ(),
		"COLUMNS=120",
		"CLI_CONFIG_FILE="+configPath,
		EnvPrefix+"NO_INTERACTION=1",
		EnvPrefix+"VERSION=1.0.0",
		EnvPrefix+"HOME="+os.TempDir(),
		"TZ=UTC",
	)
}
