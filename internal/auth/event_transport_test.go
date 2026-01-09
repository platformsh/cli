package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventTransport_RoundTrip(t *testing.T) {
	cases := []struct {
		name              string
		eventName         string
		userAgent         string
		existingUserAgent string
		wantEventHeader   string
		wantUserAgent     string
	}{
		{
			name:            "sets both headers when provided",
			eventName:       "backup:restore",
			userAgent:       "Upsun-CLI/1.0.0",
			wantEventHeader: "backup:restore",
			wantUserAgent:   "Upsun-CLI/1.0.0",
		},
		{
			name:            "sets only event header when user agent is empty",
			eventName:       "project:info",
			userAgent:       "",
			wantEventHeader: "project:info",
			wantUserAgent:   "Go-http-client/1.1", // Go's default User-Agent
		},
		{
			name:            "sets only user agent when event name is empty",
			eventName:       "",
			userAgent:       "Upsun-CLI/1.0.0",
			wantEventHeader: "",
			wantUserAgent:   "Upsun-CLI/1.0.0",
		},
		{
			name:            "does not set headers when both are empty",
			eventName:       "",
			userAgent:       "",
			wantEventHeader: "",
			wantUserAgent:   "Go-http-client/1.1", // Go's default User-Agent
		},
		{
			name:              "does not override existing user agent",
			eventName:         "init",
			userAgent:         "Upsun-CLI/1.0.0",
			existingUserAgent: "Custom-Agent/2.0",
			wantEventHeader:   "init",
			wantUserAgent:     "Custom-Agent/2.0",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var receivedEventHeader, receivedUserAgent string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedEventHeader = r.Header.Get("X-CLI-Event")
				receivedUserAgent = r.Header.Get("User-Agent")
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			transport := &EventTransport{
				Base:      http.DefaultTransport,
				EventName: tc.eventName,
				UserAgent: tc.userAgent,
			}

			client := &http.Client{Transport: transport}

			req, err := http.NewRequest(http.MethodGet, server.URL, http.NoBody)
			require.NoError(t, err)

			if tc.existingUserAgent != "" {
				req.Header.Set("User-Agent", tc.existingUserAgent)
			}

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, tc.wantEventHeader, receivedEventHeader)
			assert.Equal(t, tc.wantUserAgent, receivedUserAgent)
		})
	}
}

func TestWithEventName(t *testing.T) {
	cases := []struct {
		name      string
		eventName string
	}{
		{
			name:      "stores and retrieves event name",
			eventName: "backup:restore",
		},
		{
			name:      "handles empty event name",
			eventName: "",
		},
		{
			name:      "handles command with namespace",
			eventName: "project:info",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = WithEventName(ctx, tc.eventName)

			got := EventNameFromContext(ctx)
			assert.Equal(t, tc.eventName, got)
		})
	}
}

func TestEventNameFromContext_EmptyContext(t *testing.T) {
	ctx := context.Background()
	got := EventNameFromContext(ctx)
	assert.Equal(t, "", got)
}
