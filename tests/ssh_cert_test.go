package tests

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/platformsh/cli/internal/mock"
	"github.com/platformsh/cli/internal/mock/api"
)

func TestSSHCerts(t *testing.T) {
	authServer := mock.NewAuthServer(t)
	defer authServer.Close()

	myUserID := "my-user-id"

	apiHandler := api.NewHandler(t)
	apiHandler.SetMyUser(&api.User{ID: myUserID})
	apiServer := httptest.NewServer(apiHandler)
	defer apiServer.Close()

	run := func(args ...string) string {
		cmd := authenticatedCommand(t, apiServer.URL, authServer.URL, args...)
		b, err := cmd.Output()
		require.NoError(t, err)
		return string(b)
	}

	output := run("ssh-cert:info")
	assert.Regexp(t, `(?m)^filename: .+?id_ed25519-cert\.pub$`, output)
	assert.Contains(t, output, "key_id: test-key-id\n")
	assert.Contains(t, output, "key_type: ssh-ed25519-cert-v01@openssh.com\n")
}
