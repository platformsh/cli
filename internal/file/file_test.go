package file

import (
	"os"
	"testing"

	"github.com/liamg/memoryfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckHash(t *testing.T) {
	// Temporarily swap to a memory filesystem.
	testableFS = memoryfs.New()
	defer func() {
		testableFS = os.DirFS("/")
	}()
	filesystem := testableFS.(*memoryfs.FS) //nolint:errcheck

	mockContent := "hello world\n"
	mockContentHash := "a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447"
	diffContent := "hello world?\n"
	diffContentHash := "d441ffff4b6663c3f150bda9c519a58c0685e34cf13d26e881d7e004f704eeba"

	cases := []struct {
		name      string
		content   string
		writeHash string

		checkHash  string
		shouldFail bool
	}{
		{
			name:      "static_hash_good",
			writeHash: mockContentHash,
			checkHash: mockContentHash,
		},
		{
			name:       "static_hash_bad",
			writeHash:  diffContentHash,
			checkHash:  mockContentHash,
			shouldFail: true,
		},
		{
			name:      "dynamic_hash_good",
			content:   diffContent,
			checkHash: diffContentHash,
		},
		{
			name:       "dynamic_hash_bad",
			content:    mockContent,
			checkHash:  diffContentHash,
			shouldFail: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			filename := c.name
			require.NoError(t, filesystem.WriteFile(filename, []byte(c.content), 0o644))
			if c.writeHash != "" {
				require.NoError(t, filesystem.WriteFile(filename+HashExt, []byte(c.writeHash), 0o644))
			}
			hashOK, err := CheckHash(filename, c.checkHash)
			assert.NoError(t, err)
			assert.Equal(t, !c.shouldFail, hashOK)
		})
	}
}
