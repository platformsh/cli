package legacy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPHPManager(t *testing.T) {
	tempDir := t.TempDir()

	pm := newPHPManager(tempDir)
	assert.NoError(t, pm.copy())

	assert.FileExists(t, pm.binPath())
}
