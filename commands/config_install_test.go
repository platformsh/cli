package commands

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/platformsh/cli/internal/config"
	"github.com/platformsh/cli/internal/config/alt"
)

func TestConfigInstallCmd(t *testing.T) {
	tempDir := t.TempDir()

	// Ensure filesystem functions looking for UserHomeDir or UserConfigDir return the test directory.
	homeEnv := os.Getenv("HOME")
	require.NoError(t, os.Setenv("HOME", tempDir))
	require.NoError(t, os.Unsetenv("XDG_CONFIG_HOME"))
	require.NoError(t, os.Unsetenv("TEST_HOME"))
	t.Cleanup(func() {
		_ = os.Setenv("HOME", homeEnv)
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/test-config.yaml" {
			cnf := testConfig()
			_ = yaml.NewEncoder(w).Encode(cnf)
		}
	}))
	defer server.Close()
	testConfigURL := server.URL + "/test-config.yaml"

	cnf := testConfig()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = config.ToContext(ctx, cnf)

	cmd := configInstallCommand
	cmd.SetContext(ctx)
	cmd.SetOut(io.Discard)

	args := []string{testConfigURL}

	stdErrBuf := &bytes.Buffer{}
	cmd.SetErr(stdErrBuf)
	err := cmd.RunE(cmd, args)
	assert.ErrorContains(t, err, "cannot install config for same executable name as this program: test")

	cnf.Application.Executable = "test-cli-executable-host"
	err = cmd.RunE(cmd, args)
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(tempDir, alt.HomeSubDir, "test-cli-executable.yaml"))
	assert.FileExists(t, filepath.Join(tempDir, alt.HomeSubDir, "bin", "test-cli-executable"))
	assert.Contains(t, stdErrBuf.String(), filepath.Join("~", alt.HomeSubDir, "test-cli-executable.yaml"))
	assert.Contains(t, stdErrBuf.String(), filepath.Join("~", alt.HomeSubDir, "bin", "test-cli-executable"))

	b, err := os.ReadFile(filepath.Join(tempDir, alt.HomeSubDir, "bin", "test-cli-executable"))
	require.NoError(t, err)
	assert.Contains(t, string(b), filepath.Join(tempDir, alt.HomeSubDir, "test-cli-executable.yaml"))
	assert.Contains(t, string(b), `test-cli-executable-host "$@"`)
}

func testConfig() *config.Config {
	cnf := &config.Config{}
	cnf.Application.Name = "Test CLI"
	cnf.Application.Executable = "test-cli-executable" // Not "test" as that is usually a real binary
	cnf.Application.EnvPrefix = "TEST_"
	cnf.Application.Slug = "test-cli"
	cnf.Application.UserConfigDir = ".test-cli"
	cnf.API.BaseURL = "https://localhost"
	cnf.API.AuthURL = "https://localhost"
	cnf.Detection.GitRemoteName = "platform"
	cnf.Service.Name = "Test"
	cnf.Service.EnvPrefix = "TEST_"
	cnf.Service.ProjectConfigDir = ".test"
	cnf.SSH.DomainWildcards = []string{"*"}
	return cnf
}
