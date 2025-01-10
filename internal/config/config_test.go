package config_test

import (
	_ "embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/platformsh/cli/internal/config"
)

//go:embed test-data/valid-config.yaml
var validConfig string

func TestFromYAML(t *testing.T) {
	t.Run("missing_values", func(t *testing.T) {
		_, err := config.FromYAML([]byte(`application: {name: Test CLI}`))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), `Error:Field validation for 'EnvPrefix' failed on the 'required' tag`)
	})

	t.Run("complete", func(t *testing.T) {
		cnf, err := config.FromYAML([]byte(validConfig))
		assert.NoError(t, err)

		// Test defaults
		assert.Equal(t, "state.json", cnf.Application.UserStateFile)
		assert.Equal(t, true, cnf.Updates.Check)
		assert.Equal(t, 3600, cnf.Updates.CheckInterval)
		assert.Equal(t, cnf.Application.UserConfigDir, cnf.Application.WritableUserDir)
		assert.Equal(t, "example-cli-tmp", cnf.Application.TempSubDir)
		assert.Equal(t, "platform", cnf.Service.ProjectConfigFlavor)

		homeDir, err := os.UserHomeDir()
		require.NoError(t, err, "the test requires a home directory")

		writableDir, err := cnf.WritableUserDir()
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(homeDir, cnf.Application.WritableUserDir), writableDir)

		_, err = cnf.CacheDir()
		assert.NoError(t, err)
	})
}
