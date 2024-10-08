package integration

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

var _validatedCommand string

func getCommandName(t *testing.T) string {
	if _validatedCommand != "" {
		return _validatedCommand
	}
	candidate := os.Getenv("TEST_CLI_PATH")
	if candidate == "" {
		t.Skip("enable by setting TEST_CLI_PATH (or use `make integration-test`)")
	}
	versionCmd := exec.Command(candidate, "--version")
	versionCmd.Env = testEnv()
	output, err := versionCmd.Output()
	require.NoError(t, err, "running '--version' must succeed under the CLI at: %s", candidate)
	require.Contains(t, string(output), "Platform Test CLI ")
	if testing.Verbose() {
		log.Printf("Validated CLI command %s", candidate)
	}
	_validatedCommand = candidate
	return _validatedCommand
}

func command(t *testing.T, args ...string) *exec.Cmd {
	cmd := exec.Command(getCommandName(t), args...) //nolint:gosec
	cmd.Env = testEnv()
	cmd.Dir = os.TempDir()
	return cmd
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
	)
}
