package alt_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/platformsh/cli/internal/config/alt"
)

func TestInPath(t *testing.T) {
	tempDir := t.TempDir()
	tempDir, _ = filepath.EvalSymlinks(tempDir)

	homeDir := "/custom/home/directory"
	require.NoError(t, os.Setenv("HOME", homeDir))
	require.NoError(t, os.Setenv("CUSTOM_ENV_VAR", "/custom/path"))
	t.Cleanup(func() {
		_ = os.Unsetenv("PATH")
		_ = os.Unsetenv("HOME")
		_ = os.Unsetenv("CUSTOM_ENV_VAR")
	})

	require.NoError(t, os.Chdir(tempDir))

	if runtime.GOOS == "windows" {
		t.Skip()
	}

	cases := []struct {
		name    string
		dir     string
		pathEnv string
		inPath  bool
	}{
		{
			name:    "double-dot-input",
			dir:     tempDir + "/foo/../foo",
			pathEnv: tempDir + "/foo",
			inPath:  true,
		},
		{
			name:    "double-dot-both",
			dir:     tempDir + "/./foo//../foo",
			pathEnv: tempDir + "/foo/../foo//",
			inPath:  true,
		},
		{
			name:    "home-tilde",
			dir:     homeDir + "/foo/bar/.",
			pathEnv: "/usr/bin:~/foo/bar:/usr/local/bin" + "$HOME/.local/bin",
			inPath:  true,
		},
		{
			name:    "home-variable",
			dir:     homeDir + "/.local/bin",
			pathEnv: "/usr/bin:~/foo/bar:/usr/local/bin:$HOME/.local/bin",
			inPath:  true,
		},
		{
			name:    "home-not-in",
			dir:     homeDir + "/.local/bin",
			pathEnv: "/usr/bin:/usr/local/bin:/usr/.local/bin",
		},
		{
			name:    "custom-variable",
			dir:     "/custom/path/foo",
			pathEnv: `$CUSTOM_ENV_VAR/foo/.:~/.local/bin:/nonexistent/dir`,
			inPath:  true,
		},
		{
			name:    "custom-variable-prefixed",
			dir:     "/prefix/custom/path/foo",
			pathEnv: `~/bin:/prefix/$CUSTOM_ENV_VAR//foo:/bin`,
			inPath:  true,
		},
		{
			name:    "relative",
			dir:     tempDir + "/foo",
			pathEnv: `/bin:./foo/bar/..:/usr/local/bin`,
			inPath:  true,
		},
		{
			name:    "relative-not",
			dir:     tempDir + "/foo",
			pathEnv: `/bin:/foo:/usr/local/bin`,
		},
		{
			name:    "this-dir-as-dot",
			dir:     tempDir + "/x/..",
			pathEnv: `/bin:.:/usr/local/bin`,
			inPath:  true,
		},
		{
			name:    "this-dir-as-empty-entry",
			dir:     tempDir + "/foo/..",
			pathEnv: `/bin::/usr/local/bin`,
			inPath:  true,
		},
		{
			name:    "this-dir-not",
			dir:     tempDir,
			pathEnv: `/bin:/usr/local/bin`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			require.NoError(t, os.Setenv("PATH", c.pathEnv))
			assert.Equal(t, alt.InPath(c.dir), c.inPath)
		})
	}
}
