package file

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopyIfChanged(t *testing.T) {
	cases := []struct {
		name        string
		initialData []byte
		sourceData  []byte
		expectWrite bool
	}{
		{"File does not exist", nil, []byte("new data"), true},
		{"File matches source", []byte("same data"), []byte("same data"), false},
		{"File content differs", []byte("old data"), []byte("new data"), true},
		{"File size differs", []byte("short"), []byte("much longer data"), true},
		{"Empty source", []byte("existing data"), []byte{}, true},
	}

	tmpDir := t.TempDir()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			destFile := filepath.Join(tmpDir, "testfile")

			if c.initialData != nil {
				require.NoError(t, os.WriteFile(destFile, c.initialData, 0o600))
				time.Sleep(time.Millisecond * 5)
			}

			var modTimeBeforeCopy time.Time
			stat, err := os.Stat(destFile)
			if c.initialData == nil {
				require.True(t, os.IsNotExist(err))
			} else {
				require.NoError(t, err)
				modTimeBeforeCopy = stat.ModTime()
			}

			err = CopyIfChanged(destFile, c.sourceData, 0o600)
			require.NoError(t, err)

			statAfterCopy, err := os.Stat(destFile)
			require.NoError(t, err)
			if c.expectWrite {
				assert.Greater(t, statAfterCopy.ModTime().Truncate(time.Millisecond), modTimeBeforeCopy.Truncate(time.Millisecond))
			} else {
				assert.Equal(t, modTimeBeforeCopy.Truncate(time.Millisecond), statAfterCopy.ModTime().Truncate(time.Millisecond))
			}

			data, err := os.ReadFile(destFile)
			require.NoError(t, err)

			assert.Equal(t, data, c.sourceData)
		})
	}
}
