package mocks

import (
	"encoding/json"
	"testing"

	"net/http"
	"net/http/httptest"
)

var APITokens = map[string]string{
	"api-token-1": "access-token1",
}

// APITokenServer creates a new mock OAuth 2.0 API token server.
// The caller must call Close() on the server when finished.
func APITokenServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if testing.Verbose() {
			t.Log(req)
		}
		if req.Method == http.MethodPost && req.URL.Path == "/oauth2/token" {
			if err := req.ParseForm(); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			if gt := req.Form.Get("grant_type"); gt != "api_token" {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid grant type: " + gt})
				return
			}
			apiToken := req.Form.Get("api_token")
			if accessToken, ok := APITokens[apiToken]; ok {
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(struct {
					AccessToken string `json:"access_token"`
					ExpiresIn   int    `json:"expires_in"`
					Type        string `json:"token_type"`
				}{AccessToken: accessToken, ExpiresIn: 60, Type: "bearer"})
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid API token"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	}))
}
