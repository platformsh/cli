package tests

import (
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/platformsh/cli/internal/mock"
	"github.com/platformsh/cli/internal/mock/api"
)

func TestAuthInfo(t *testing.T) {
	authServer := mock.NewAuthServer(t)
	defer authServer.Close()

	apiHandler := api.NewHandler(t)
	apiHandler.SetMyUser(&api.User{
		ID:                  "my-user-id",
		Deactivated:         false,
		Namespace:           "ns",
		Username:            "my-username",
		FirstName:           "Foo",
		LastName:            "Bar",
		Email:               "my-user@example.com",
		EmailVerified:       true,
		Picture:             "https://example.com/profile.png",
		Country:             "NO",
		PhoneNumberVerified: true,
		MFAEnabled:          true,
	})

	apiServer := httptest.NewServer(apiHandler)
	defer apiServer.Close()

	run := func(args ...string) string {
		cmd := authenticatedCommand(t, apiServer.URL, authServer.URL, args...)
		if testing.Verbose() {
			cmd.Stderr = os.Stderr
		}

		b, err := cmd.Output()
		require.NoError(t, err)
		return string(b)
	}

	assert.Equal(t, strings.TrimLeft(`
+-----------------------+---------------------+
| Property              | Value               |
+-----------------------+---------------------+
| id                    | my-user-id          |
| first_name            | Foo                 |
| last_name             | Bar                 |
| username              | my-username         |
| email                 | my-user@example.com |
| phone_number_verified | true                |
+-----------------------+---------------------+
`, "\n"), run("auth:info", "-v"))

	assert.Equal(t, "my-user-id\n", run("auth:info", "-P", "id"))
}
