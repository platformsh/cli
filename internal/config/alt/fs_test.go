package alt

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindConfigDir(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("XDG_CONFIG_HOME exists", func(t *testing.T) {
		switch runtime.GOOS {
		case "windows", "darwin", "ios", "plan9":
			t.Skip()
		}
		err := os.Setenv("XDG_CONFIG_HOME", tempDir)
		require.NoError(t, err)
		defer os.Unsetenv("XDG_CONFIG_HOME")

		err = os.Mkdir(filepath.Join(tempDir, subDir), 0o755)
		require.NoError(t, err)

		result, err := FindConfigDir()
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(tempDir, subDir), result)
	})

	t.Run("HOME fallback", func(t *testing.T) {
		err := os.Setenv("HOME", tempDir)
		require.NoError(t, err)
		defer os.Unsetenv("HOME")

		result, err := FindConfigDir()
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(tempDir, homeSubDir), result)
	})
}

func TestFindBinDir(t *testing.T) {
	tempDir := t.TempDir()

	err := os.Setenv("HOME", tempDir)
	require.NoError(t, err)
	defer os.Unsetenv("HOME")

	result, err := FindBinDir()
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(tempDir, homeSubDir, "bin"), result)

	var standardDir string
	if runtime.GOOS == "windows" {
		standardDir = filepath.Join("AppData", "Local", "Programs")
	} else {
		standardDir = filepath.Join(".local", "bin")
	}
	err = os.Setenv("PATH", os.Getenv("PATH")+string(os.PathListSeparator)+filepath.Join(tempDir, standardDir))
	require.NoError(t, err)

	result, err = FindBinDir()
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(tempDir, standardDir), result)
}

func TestFSHelpers(t *testing.T) {
	tempDir := t.TempDir()

	require.NoError(t, writeFile(filepath.Join(tempDir, "test.txt"), []byte("test"), 0, 0o644))
	require.NoError(t, writeFile(filepath.Join(tempDir, "subdir", "test2.txt"), []byte("test2"), 0o755, 0o644))

	dirExists, err := isExistingDirectory(filepath.Join(tempDir, "subdir"))
	assert.NoError(t, err)
	assert.True(t, dirExists)

	dirExists, err = isExistingDirectory(filepath.Join(tempDir, "not-a-subdir"))
	assert.NoError(t, err)
	assert.False(t, dirExists)

	dirExists, err = isExistingDirectory(filepath.Join(tempDir, "test.txt"))
	assert.NoError(t, err)
	assert.False(t, dirExists)
}
