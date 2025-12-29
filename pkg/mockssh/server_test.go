package mockssh_test

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"

	"github.com/upsun/cli/pkg/mockapi"
	"github.com/upsun/cli/pkg/mockssh"
)

func TestServer(t *testing.T) {
	authServer := mockapi.NewAuthServer(t)
	defer authServer.Close()

	sshServer, err := mockssh.NewServer(t, authServer.URL+"/ssh/authority")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = sshServer.Stop()
	})

	tempDir := t.TempDir()
	tempDir, _ = filepath.EvalSymlinks(tempDir)
	sshServer.CommandHandler = mockssh.ExecHandler(tempDir, []string{})

	cert := getTestSSHAuth(t, authServer.URL)

	// Create the SSH client configuration
	address := fmt.Sprintf("127.0.0.1:%d", sshServer.Port())
	config := &ssh.ClientConfig{
		User: "test",
		Auth: []ssh.AuthMethod{ssh.PublicKeys(cert)},
		HostKeyCallback: func(_ string, remote net.Addr, key ssh.PublicKey) error {
			if remote.String() != address {
				return fmt.Errorf("unexpected address: %s", remote.String())
			}
			if bytes.Equal(sshServer.HostKey().Marshal(), key.Marshal()) {
				return nil
			}
			return fmt.Errorf("host key mismatch")
		},
	}

	client, err := ssh.Dial("tcp", address, config)
	require.NoError(t, err)
	defer client.Close()

	session, err := client.NewSession()
	require.NoError(t, err)
	defer session.Close()

	stdOutBuffer := &bytes.Buffer{}
	session.Stdout = stdOutBuffer

	require.NoError(t, session.Run("pwd"))
	assert.Equal(t, tempDir, strings.TrimRight(stdOutBuffer.String(), "\n"))

	session2, err := client.NewSession()
	require.NoError(t, err)
	defer session2.Close()
	err = session2.Run("false")
	assert.Error(t, err)
	var exitErr *ssh.ExitError
	assert.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 1, exitErr.ExitStatus())
}

func getTestSSHAuth(t *testing.T, authServerURL string) ssh.Signer {
	t.Helper()

	// Generate a keypair
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	s, err := ssh.NewSignerFromKey(priv)
	require.NoError(t, err)

	b, err := json.Marshal(struct{ Key string }{string(ssh.MarshalAuthorizedKey(s.PublicKey()))})
	require.NoError(t, err)
	resp, err := http.DefaultClient.Post(authServerURL+"/ssh", "application/json", bytes.NewReader(b))
	require.NoError(t, err)
	defer resp.Body.Close()

	var rs struct{ Certificate string }
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&rs))

	parsed, _, _, _, err := ssh.ParseAuthorizedKey([]byte(rs.Certificate)) //nolint: dogsled
	require.NoError(t, err)

	cert, _ := parsed.(*ssh.Certificate)
	certSigner, err := ssh.NewCertSigner(cert, s)
	require.NoError(t, err)

	return certSigner
}
