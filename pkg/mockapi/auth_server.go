package mockapi

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

var ValidAPITokens = []string{"api-token-1"}
var accessTokens = []string{"access-token-1"}

// NewAuthServer creates a new mock authentication server.
// The caller must call Close() on the server when finished.
func NewAuthServer(t *testing.T) *httptest.Server {
	mux := chi.NewRouter()
	if testing.Verbose() {
		mux.Use(middleware.DefaultLogger)
	}

	mux.Post("/oauth2/token", func(w http.ResponseWriter, req *http.Request) {
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
	})

	mux.Get("/ssh/authority", func(w http.ResponseWriter, _ *http.Request) {
		pks, err := publicKeys()
		require.NoError(t, err)
		data := struct {
			Authorities []string `json:"authorities"`
		}{}
		for _, k := range pks {
			sshPubKey, err := ssh.NewPublicKey(k)
			require.NoError(t, err)
			data.Authorities = append(data.Authorities, string(ssh.MarshalAuthorizedKey(sshPubKey)))
		}
		_ = json.NewEncoder(w).Encode(data)
	})

	mux.Post("/ssh", func(w http.ResponseWriter, req *http.Request) {
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
			ValidAfter:  uint64(time.Now().Add(-1 * time.Second).Unix()), //nolint:gosec // G115
			ValidBefore: uint64(time.Now().Add(time.Minute).Unix()),      //nolint:gosec // G115
			Permissions: ssh.Permissions{
				Extensions: extensions,
			},
		}
		err = cert.SignCert(rand.Reader, signer)
		require.NoError(t, err)
		_ = json.NewEncoder(w).Encode(struct {
			Cert string `json:"certificate"`
		}{string(ssh.MarshalAuthorizedKey(cert))})
	})

	return httptest.NewServer(mux)
}

// publicKeys returns the server's public keys, e.g. for SSH certificate generation.
func publicKeys() ([]crypto.PublicKey, error) {
	pub, _, err := keyPair()
	if err != nil {
		return nil, err
	}

	return []crypto.PublicKey{pub}, nil
}

var (
	privateKey crypto.PrivateKey
	publicKey  crypto.PublicKey
)

func keyPair() (crypto.PublicKey, crypto.PrivateKey, error) {
	if privateKey == nil || publicKey == nil {
		pub, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, nil, err
		}
		privateKey = priv
		publicKey = pub
	}
	return publicKey, privateKey, nil
}

var signer ssh.Signer

func sshSigner() (ssh.Signer, error) {
	if signer != nil {
		return signer, nil
	}
	_, priv, err := keyPair()
	if err != nil {
		return nil, err
	}
	s, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		return nil, err
	}
	signer = s
	return s, nil
}
