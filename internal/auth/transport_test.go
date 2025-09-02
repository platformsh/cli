package auth

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// mockRefresher implements both the refresher and oauth2.TokenSource interfaces for testing
type mockRefresher struct {
	token *oauth2.Token
}

func (m *mockRefresher) refreshToken() error {
	m.token = &oauth2.Token{
		AccessToken: "valid",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(time.Hour),
	}
	return nil
}

func (m *mockRefresher) invalidateToken() error {
	m.token = &oauth2.Token{
		AccessToken: "",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(-time.Hour),
	}

	return nil
}

func (m *mockRefresher) Token() (*oauth2.Token, error) {
	if m.token == nil || !m.token.Valid() {
		if err := m.refreshToken(); err != nil {
			return nil, err
		}
	}
	return m.token, nil
}

func TestTransport_RoundTrip_RetryOn401(t *testing.T) {
	// Create a mock server that initially returns 401, then 200
	responseCodes := []int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read and validate the request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		// Check that we have the expected POST body
		assert.Equal(t, "test-body-content", string(body))

		if r.Header.Get("Authorization") != "Bearer valid" {
			w.WriteHeader(http.StatusUnauthorized)
			if _, err := w.Write([]byte(`{"error": "unauthorized"}`)); err != nil {
				require.NoError(t, err)
			}
			responseCodes = append(responseCodes, http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"success": true}`)); err != nil {
			require.NoError(t, err)
		}
		responseCodes = append(responseCodes, http.StatusOK)
	}))
	defer server.Close()

	// Create mock refresher with token sequence: first invalid, then valid
	mockRef := &mockRefresher{
		token: &oauth2.Token{
			AccessToken: "invalid",
			TokenType:   "Bearer",
			Expiry:      time.Now().Add(time.Hour),
		},
	}

	// Create our Transport with the mock refresher
	transport := &Transport{
		base: &oauth2.Transport{
			Source: mockRef,
			Base:   http.DefaultTransport,
		},
		refresher: mockRef,
	}

	// Create HTTP client with our transport
	client := &http.Client{Transport: transport}

	// Make a POST request with body content
	requestBody := "test-body-content"
	req, err := http.NewRequest("POST", server.URL, bytes.NewBufferString(requestBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify we got a successful response after retry
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	responseBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, `{"success": true}`, string(responseBody))

	// Assert the response codes (401 first and then a 200)
	assert.Equal(t, []int{http.StatusUnauthorized, http.StatusOK}, responseCodes)
}
