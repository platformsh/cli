package convert

import (
	_ "embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/platformsh/.upsun/config-ref.yaml
var configRef string

func TestConvert(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.CopyFS(tmpDir, os.DirFS("testdata/platformsh")))
	assert.NoError(t, PlatformshToUpsun(tmpDir, t.Output()))
	assert.FileExists(t, filepath.Join(tmpDir, ".upsun", "config.yaml"))

	b, err := os.ReadFile(filepath.Join(tmpDir, ".upsun", "config.yaml"))
	assert.NoError(t, err)

	assert.Equal(t, configRef, string(b))
}
