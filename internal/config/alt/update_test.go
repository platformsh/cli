package alt_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/upsun/cli/internal/config"
	"github.com/upsun/cli/internal/config/alt"
	"github.com/upsun/cli/internal/state"
)

func TestUpdate(t *testing.T) {
	tempDir := t.TempDir()

	// Copy test config to a temporary directory, and fake its modification time.
	testConfigFilename := filepath.Join(tempDir, "config.yaml")
	err := os.WriteFile(testConfigFilename, testConfig, 0o600)
	require.NoError(t, err)

	cnf, err := config.FromYAML(testConfig)
	require.NoError(t, err)

	// Set up state so that it stays in a temporary directory.
	err = os.Setenv(cnf.Application.EnvPrefix+"HOME", tempDir)
	require.NoError(t, err)

	// Set up the config to be updated via a test HTTP server.
	remoteConfig := testConfig
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/config.yaml" {
			_, _ = w.Write(remoteConfig)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cnf.SourceFile = testConfigFilename
	cnf.Updates.CheckInterval = 1
	cnf.Metadata.URL = server.URL + "/config.yaml"

	// TODO use test context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = config.ToContext(ctx, cnf)

	var lastLogged string
	logger := func(msg string, args ...any) {
		lastLogged = fmt.Sprintf(msg, args...)
	}

	assert.True(t, alt.ShouldUpdate(cnf))

	err = alt.Update(ctx, cnf, logger)
	assert.NoError(t, err)
	assert.Contains(t, lastLogged, "Config file updated recently")

	hourAgo := time.Now().Add(-time.Hour)
	require.NoError(t, os.Chtimes(testConfigFilename, hourAgo, hourAgo))

	err = alt.Update(ctx, cnf, logger)
	assert.NoError(t, err)
	assert.Contains(t, lastLogged, "Automatically updated config file")

	err = alt.Update(ctx, cnf, logger)
	assert.NoError(t, err)
	assert.Contains(t, lastLogged, "Config updates checked recently")

	// Reset the LastChecked time and file modified time.
	resetTimes := func() {
		s, err := state.Load(cnf)
		require.NoError(t, err)
		s.ConfigUpdates.LastChecked = 0
		require.NoError(t, state.Save(s, cnf))
		require.NoError(t, os.Chtimes(testConfigFilename, hourAgo, hourAgo))
	}
	resetTimes()

	remoteConfig = append(remoteConfig, []byte("\nmetadata: {version: 1.0.1}")...)
	cnf.Metadata.Version = "invalid"
	err = alt.Update(ctx, cnf, logger)
	assert.ErrorContains(t, err, "could not compare config versions")
	resetTimes()
	cnf.Metadata.Version = "1.0.1"
	err = alt.Update(ctx, cnf, logger)
	assert.NoError(t, err)
	assert.Contains(t, lastLogged, "Config is already up to date (version 1.0.1)")

	resetTimes()

	updated := time.Now()
	cnf.Metadata.Version = ""
	cnf.Metadata.UpdatedAt = updated
	remoteConfig = testConfig
	remoteConfig = append(remoteConfig,
		[]byte(fmt.Sprintf("\nmetadata: {updated_at: %s}", updated.Add(-time.Minute).Format(time.RFC3339)))...)
	err = alt.Update(ctx, cnf, logger)
	assert.NoError(t, err)
	assert.Contains(t, lastLogged, "Config is already up to date")
}

func TestShouldUpdate(t *testing.T) {
	testConfigFilename := "/tmp/mock/path/to/config.yaml"

	cnf, err := config.FromYAML(testConfig)
	require.NoError(t, err)

	cnf.Updates.Check = true
	cnf.SourceFile = testConfigFilename
	cnf.Metadata.URL = "https://example.com/config.yaml"
	assert.True(t, alt.ShouldUpdate(cnf))

	cnf.Updates.Check = false
	assert.False(t, alt.ShouldUpdate(cnf))

	cnf.Updates.Check = true
	cnf.SourceFile = ""
	assert.False(t, alt.ShouldUpdate(cnf))

	cnf.SourceFile = testConfigFilename
	cnf.Metadata.URL = ""
	assert.False(t, alt.ShouldUpdate(cnf))
}
