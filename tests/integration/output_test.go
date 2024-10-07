package integration

import (
	"embed"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed data
var dataFS embed.FS

func assertDataEquals(t *testing.T, filename, actual string) bool {
	filepath := path.Join("data", filename)
	expected, err := dataFS.ReadFile(filepath)
	require.NoError(t, err)
	return assert.Equal(t, string(expected), actual, "match content of file: %s", filepath)
}

func TestCommandOutput(t *testing.T) {
	commands := []string{"help", "list", "help list", "help create"}
	for _, c := range commands {
		output, err := command(t, strings.Split(c, " ")...).Output()
		require.NoError(t, err)
		assertDataEquals(t, c+".stdout.txt", string(output))
	}
}
