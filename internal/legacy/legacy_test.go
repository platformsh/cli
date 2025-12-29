package legacy

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/upsun/cli/internal/config"
)

func TestLegacyCLI(t *testing.T) {
	if len(phar) == 0 || len(phpCLI) == 0 {
		t.Skip()
	}

	cnf := &config.Config{}
	cnf.Application.Name = "Test CLI"
	cnf.Application.Executable = "platform-test"
	cnf.Application.Slug = "test-cli"
	cnf.Application.EnvPrefix = "TEST_CLI_"
	cnf.Application.TempSubDir = "temp_sub_dir"

	tempDir := t.TempDir()

	_ = os.Setenv(cnf.Application.EnvPrefix+"TMP", tempDir)
	t.Cleanup(func() {
		_ = os.Unsetenv(cnf.Application.EnvPrefix + "TMP")
	})

	stdout := &bytes.Buffer{}
	stdErr := io.Discard
	if testing.Verbose() {
		stdErr = os.Stderr
	}

	testCLIVersion := "1.2.3"

	wrapper := &CLIWrapper{
		Stdout:             stdout,
		Stderr:             stdErr,
		Config:             cnf,
		Version:            testCLIVersion,
		DisableInteraction: true,
	}
	if testing.Verbose() {
		wrapper.DebugLogFunc = t.Logf
	}
	PHPVersion = "6.5.4"
	LegacyCLIVersion = "3.2.1"

	err := wrapper.Exec(context.Background(), "help")
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "Displays help for a command")

	cacheDir, err := wrapper.cacheDir()
	require.NoError(t, err)

	pharPath, err := wrapper.PharPath()
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(cacheDir, "platform-test.phar"), pharPath)

	stdout.Reset()
	err = wrapper.Exec(context.Background(), "--version")
	assert.NoError(t, err)
	assert.Equal(t, "Test CLI "+testCLIVersion, strings.TrimSuffix(stdout.String(), "\n"))
}
