package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/platformsh/cli/tests/integration/mocks"
)

func TestEnvironmentList(t *testing.T) {
	authServer := mocks.APITokenServer(t)
	defer authServer.Close()

	apiServer := environmentListServer(t)
	defer apiServer.Close()

	// The legacy CLI identifier expects project IDs to be alphanumeric.
	// See: https://github.com/platformsh/legacy-cli/blob/main/src/Service/Identifier.php#L75
	mockProjectID := "abcdefg123456"

	run := func(args ...string) string {
		cmd := command(t, args...)
		cmd.Env = append(
			cmd.Env,
			EnvPrefix+"API_BASE_URL="+apiServer.URL,
			EnvPrefix+"API_AUTH_URL="+authServer.URL,
			EnvPrefix+"TOKEN="+mocks.ValidAPITokens[0],
		)
		if testing.Verbose() {
			cmd.Stderr = os.Stderr
		}

		b, err := cmd.Output()
		require.NoError(t, err)
		return string(b)
	}

	assert.Equal(t, strings.TrimLeft(`
+-----------+---------+----------+-------------+
| ID        | Title   | Status   | Type        |
+-----------+---------+----------+-------------+
| main      | Main    | Active   | production  |
|   staging | Staging | Active   | staging     |
|     dev   | Dev     | Active   | development |
|       fix | Fix     | Inactive | development |
+-----------+---------+----------+-------------+
`, "\n"), run("environment:list", "-v", "-p", mockProjectID))

	assert.Equal(t, strings.TrimLeft(`
ID	Title	Status	Type
main	Main	Active	production
staging	Staging	Active	staging
dev	Dev	Active	development
fix	Fix	Inactive	development
`, "\n"), run("environment:list", "-v", "-p", mockProjectID, "--format", "plain"))

	assert.Equal(t, strings.TrimLeft(`
ID	Title	Status	Type
main	Main	Active	production
staging	Staging	Active	staging
dev	Dev	Active	development
`, "\n"), run("environment:list", "-v", "-p", mockProjectID, "--format", "plain", "--no-inactive"))

	assert.Equal(t, "fix\n",
		run("environment:list", "-v", "-p", mockProjectID, "--pipe", "--status=inactive"))
}

func environmentListServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if testing.Verbose() {
			t.Log(req)
		}
		switch {
		case req.Method == http.MethodGet &&
			regexp.MustCompile(`^/projects/[a-z0-9-]+$`).MatchString(req.URL.Path):
			p := project{
				ID:    path.Base(req.URL.Path),
				Links: makeHALLinks("self="+req.URL.Path, "environments="+req.URL.Path+"/environments"),
			}
			_ = json.NewEncoder(w).Encode(p)
			return
		case req.Method == http.MethodGet &&
			regexp.MustCompile(`^/projects/[a-z0-9-]+/environments$`).MatchString(req.URL.Path):
			makeEnv := func(name, envType, status string, parent any) environment {
				return environment{
					ID:          name,
					Name:        name,
					MachineName: name + "-xyz",
					Title:       strings.ToTitle(name[:1]) + name[1:],
					Parent:      parent,
					Type:        envType,
					Status:      status,
					Links:       makeHALLinks("self=" + req.URL.Path + "/" + url.PathEscape(name)),
				}
			}
			_ = json.NewEncoder(w).Encode([]environment{
				makeEnv("main", "production", "active", nil),
				makeEnv("staging", "staging", "active", "main"),
				makeEnv("dev", "development", "active", "staging"),
				makeEnv("fix", "development", "inactive", "dev"),
			})
			return
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}
