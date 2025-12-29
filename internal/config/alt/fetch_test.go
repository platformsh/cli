package alt_test

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/upsun/cli/internal/config"
	"github.com/upsun/cli/internal/config/alt"
)

//go:embed test-config.yaml
var testConfig []byte

func TestFetchConfig(t *testing.T) {
	cases := []struct {
		path                  string
		handler               http.HandlerFunc
		expectConfigURL       string
		expectErrorContaining string
	}{
		{path: "/success", handler: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write(testConfig)
		}},
		{path: "/withExistingURL", handler: func(w http.ResponseWriter, _ *http.Request) {
			cnf, err := config.FromYAML(testConfig)
			require.NoError(t, err)
			cnf.Metadata.URL = "https://example.com"
			_ = yaml.NewEncoder(w).Encode(cnf)
		}, expectConfigURL: "https://example.com"},
		{path: "/error", handler: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}, expectErrorContaining: "received unexpected response code 500"},
		{path: "/invalid", handler: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = fmt.Fprintln(w, "[some invalid config]")
		}, expectErrorContaining: "invalid config YAML"},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, c := range cases {
			if c.path == r.URL.Path {
				c.handler(w, r)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// TODO use test context
	ctx := context.Background()
	ctx = config.ToContext(ctx, &config.Config{})

	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			result, cnfStruct, err := alt.FetchConfig(ctx, server.URL+c.path)
			if c.expectErrorContaining != "" {
				assert.Error(t, err, c.path)
				assert.ErrorContains(t, err, c.expectErrorContaining)
			} else {
				require.NoError(t, err, c.path)
				var decoded config.Config
				require.NoError(t, result.Decode(&decoded))
				assert.NotEmpty(t, result.HeadComment)
				assert.Empty(t, decoded.Wrapper.GitHubRepo)
				assert.Empty(t, decoded.Wrapper.HomebrewTap)
				assert.Equal(t, decoded.Application.Executable, cnfStruct.Application.Executable)
				if c.expectConfigURL != "" {
					assert.Equal(t, c.expectConfigURL, decoded.Metadata.URL)
				} else {
					assert.Equal(t, server.URL+c.path, decoded.Metadata.URL)
				}
				assert.Greater(t, decoded.Metadata.DownloadedAt, time.Now().Add(-time.Second))
			}
		})
	}

	t.Run("invalid_url", func(t *testing.T) {
		_, _, err := alt.FetchConfig(ctx, "http://example.com")
		assert.ErrorContains(t, err, "invalid")

		_, _, err = alt.FetchConfig(ctx, "://example.com")
		assert.ErrorContains(t, err, "missing protocol scheme")

		_, _, err = alt.FetchConfig(ctx, "//example.com")
		assert.ErrorContains(t, err, "invalid")

		_, _, err = alt.FetchConfig(ctx, "/path/to/file")
		assert.ErrorContains(t, err, "invalid")
	})
}
