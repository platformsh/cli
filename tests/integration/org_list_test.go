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

func TestOrgList(t *testing.T) {
	authServer := mocks.APITokenServer(t)
	defer authServer.Close()

	apiServer := orgListServer(t)
	defer apiServer.Close()

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
+--------------+--------------------------------+-----------------------+
| Name         | Label                          | Owner email           |
+--------------+--------------------------------+-----------------------+
| acme         | ACME Inc.                      | user-id-1@example.com |
| four-seasons | Four Seasons Total Landscaping | user-id-1@example.com |
| duff         | Duff Beer                      | user-id-2@example.com |
+--------------+--------------------------------+-----------------------+
`, "\n"), run("orgs"))

	assert.Equal(t, strings.TrimLeft(`
Name	Label	Owner email
acme	ACME Inc.	user-id-1@example.com
four-seasons	Four Seasons Total Landscaping	user-id-1@example.com
duff	Duff Beer	user-id-2@example.com
`, "\n"), run("orgs", "--format", "plain"))

	assert.Equal(t, strings.TrimLeft(`
org-id-1,acme
org-id-2,four-seasons
org-id-3,duff
`, "\n"), run("orgs", "--format", "csv", "--columns", "id,name", "--no-header"))
}

func orgListServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if testing.Verbose() {
			t.Log(req)
		}
		switch {
		case req.Method == http.MethodGet && req.URL.Path == "/organizations",
			req.Method == http.MethodGet && regexp.MustCompile(`^/users/[a-z0-9-]+/organizations$`).MatchString(req.URL.Path):
			makeOrg := func(id, name, label, owner string) org {
				return org{
					ID:    id,
					Name:  name,
					Label: label,
					Owner: owner,
					Links: makeHALLinks("self=/organizations/" + url.PathEscape(id)),
				}
			}
			ownerIDs := []string{"user-id-1", "user-id-2"}
			_ = json.NewEncoder(w).Encode(struct {
				Items []org    `json:"items"`
				Links halLinks `json:"_links"`
			}{Items: []org{
				makeOrg("org-id-1", "acme", "ACME Inc.", ownerIDs[0]),
				makeOrg("org-id-2", "four-seasons", "Four Seasons Total Landscaping", ownerIDs[0]),
				makeOrg("org-id-3", "duff", "Duff Beer", ownerIDs[1]),
			}, Links: makeHALLinks("ref:users:0=/ref/users?in=" + strings.Join(ownerIDs, ","))})
			return
		// TODO generalize users server
		case req.Method == http.MethodGet && req.URL.Path == "/users/me":
			_ = json.NewEncoder(w).Encode(map[string]string{"id": "user-id"})
		case req.Method == http.MethodGet && regexp.MustCompile(`^/users/[a-z0-9-]+$`).MatchString(req.URL.Path):
			_ = json.NewEncoder(w).Encode(map[string]string{"id": path.Base(req.URL.Path)})
		case req.Method == http.MethodGet && req.URL.Path == "/ref/users":
			require.NoError(t, req.ParseForm())
			ids := strings.Split(req.Form.Get("in"), ",")
			type userRef struct {
				ID       string `json:"id"`
				Email    string `json:"email"`
				Username string `json:"username"`
			}
			userRefs := make(map[string]userRef, len(ids))
			for _, id := range ids {
				userRefs[id] = userRef{ID: id, Email: id + "@example.com", Username: id}
			}
			_ = json.NewEncoder(w).Encode(userRefs)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}
