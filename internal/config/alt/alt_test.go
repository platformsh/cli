package alt_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/upsun/cli/internal/config/alt"
)

func TestAlt(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	tempDir := t.TempDir()

	configNode := &yaml.Node{}
	require.NoError(t, yaml.Unmarshal(testConfig, configNode))

	binPath := filepath.Join(tempDir, "bin", "test-binary")
	configPath := filepath.Join(tempDir, "config", "config.yaml")

	a := alt.New(
		binPath,
		"Generated for test",
		"example-target",
		configPath,
		configNode,
	)
	assert.NoError(t, a.GenerateAndSave())

	assert.FileExists(t, binPath)
	assert.FileExists(t, configPath)

	binContent, err := os.ReadFile(binPath)
	assert.NoError(t, err)
	assert.Equal(t, `#!/bin/sh
# Generated for test
export CLI_CONFIG_FILE="`+configPath+`"
example-target "$@"
`, string(binContent))

	require.NoError(t, os.Setenv("XDG_CONFIG_HOME", tempDir+"/config"))
	defer os.Unsetenv("XDG_CONFIG_HOME")

	assert.NoError(t, a.GenerateAndSave())
	binContent, err = os.ReadFile(binPath)
	assert.NoError(t, err)
	assert.Equal(t, `#!/bin/sh
# Generated for test
export CLI_CONFIG_FILE="${XDG_CONFIG_HOME}/config.yaml"
example-target "$@"
`, string(binContent))

	_ = os.Unsetenv("XDG_CONFIG_HOME")
	require.NoError(t, os.Setenv("HOME", tempDir))
	defer os.Unsetenv("HOME")

	assert.NoError(t, a.GenerateAndSave())
	binContent, err = os.ReadFile(binPath)
	assert.NoError(t, err)
	assert.Equal(t, `#!/bin/sh
# Generated for test
export CLI_CONFIG_FILE="${HOME}/config/config.yaml"
example-target "$@"
`, string(binContent))
}
