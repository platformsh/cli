package mockapi

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

var ValidAPITokens = []string{"api-token-1"}
var accessTokens = []string{"access-token-1"}

// NewAuthServer creates a new mock authentication server.
// The caller must call Close() on the server when finished.
func NewAuthServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if testing.Verbose() {
			t.Log(req)
		}
		if req.Method == http.MethodPost && req.URL.Path == "/oauth2/token" {
			require.NoError(t, req.ParseForm())
			if gt := req.Form.Get("grant_type"); gt != "api_token" {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid grant type: " + gt})
				return
			}
			apiToken := req.Form.Get("api_token")
			if slices.Contains(ValidAPITokens, apiToken) {
				_ = json.NewEncoder(w).Encode(struct {
					AccessToken string `json:"access_token"`
					ExpiresIn   int    `json:"expires_in"`
					Type        string `json:"token_type"`
				}{AccessToken: accessTokens[0], ExpiresIn: 60, Type: "bearer"})
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid API token"})
			return
		}

		if req.Method == http.MethodPost && req.URL.Path == "/ssh" {
			var options struct {
				PublicKey string `json:"key"`
			}
			err := json.NewDecoder(req.Body).Decode(&options)
			require.NoError(t, err)
			key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(options.PublicKey))
			require.NoError(t, err)
			signer, err := sshSigner()
			require.NoError(t, err)
			extensions := make(map[string]string)

			// Add standard ssh options
			extensions["permit-X11-forwarding"] = ""
			extensions["permit-agent-forwarding"] = ""
			extensions["permit-port-forwarding"] = ""
			extensions["permit-pty"] = ""
			extensions["permit-user-rc"] = ""
			cert := &ssh.Certificate{
				Key:         key,
				Serial:      0,
				CertType:    ssh.UserCert,
				KeyId:       "test-key-id",
				ValidAfter:  uint64(time.Now().Add(-1 * time.Second).Unix()),
				ValidBefore: uint64(time.Now().Add(time.Minute).Unix()),
				Permissions: ssh.Permissions{
					Extensions: extensions,
				},
			}
			err = cert.SignCert(rand.Reader, signer)
			require.NoError(t, err)
			_ = json.NewEncoder(w).Encode(struct {
				Cert string `json:"certificate"`
			}{string(ssh.MarshalAuthorizedKey(cert))})
			require.NoError(t, err)
			return
		}

		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	}))
}

var signer ssh.Signer // TODO reuse to validate SSH connection

func sshSigner() (ssh.Signer, error) {
	if signer != nil {
		return signer, nil
	}
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	s, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		return nil, err
	}
	signer = s
	return s, nil
}
