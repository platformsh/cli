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
)

func TestConfigInstallCmd(t *testing.T) {
	tempDir := t.TempDir()
	tempBinDir := filepath.Join(tempDir, "bin")
	require.NoError(t, os.Mkdir(tempBinDir, 0o755))
	_ = os.Setenv("HOME", tempDir)
	_ = os.Setenv("XDG_CONFIG_HOME", "")

	remoteConfig := testConfig()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/test-config.yaml" {
			_ = yaml.NewEncoder(w).Encode(remoteConfig)
		}
	}))
	defer server.Close()
	testConfigURL := server.URL + "/test-config.yaml"

	cnf := testConfig()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = config.ToContext(ctx, cnf)

	cmd := newConfigInstallCommand()
	cmd.SetContext(ctx)
	cmd.SetOut(io.Discard)
	_ = cmd.Flags().Set("config-dir", tempDir)
	_ = cmd.Flags().Set("bin-dir", tempBinDir)

	args := []string{testConfigURL}

	stdErrBuf := &bytes.Buffer{}
	cmd.SetErr(stdErrBuf)
	err := cmd.RunE(cmd, args)
	assert.ErrorContains(t, err, "cannot install config for same executable name as this program: test")

	cnf.Application.Executable = "test-cli-executable-host"
	err = cmd.RunE(cmd, args)
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(tempDir, "test-cli-executable.yaml"))
	assert.FileExists(t, filepath.Join(tempBinDir, "test-cli-executable"))
	assert.Contains(t, stdErrBuf.String(), "~/test-cli-executable.yaml")
	assert.Contains(t, stdErrBuf.String(), "~/bin/test-cli-executable")
	assert.Contains(t, stdErrBuf.String(), "Add the following directory to your PATH")

	b, err := os.ReadFile(filepath.Join(tempBinDir, "test-cli-executable"))
	require.NoError(t, err)
	assert.Contains(t, string(b), `"${HOME}/test-cli-executable.yaml"`)
	assert.Contains(t, string(b), `test-cli-executable-host "$@"`)

	_ = os.Setenv("PATH", tempBinDir+":"+os.Getenv("PATH"))
	remoteConfig.Application.Executable = "test-cli-executable2"
	err = cmd.RunE(cmd, args)
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(tempDir, "test-cli-executable2.yaml"))
	assert.FileExists(t, filepath.Join(tempBinDir, "test-cli-executable2"))
	assert.Contains(t, stdErrBuf.String(), "~/test-cli-executable2.yaml")
	assert.Contains(t, stdErrBuf.String(), "~/bin/test-cli-executable2")
	assert.Contains(t, stdErrBuf.String(), "Run the new CLI with: test-cli-executable2")
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
