package integration

import (
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/platformsh/cli/tests/integration/mocks"
	"github.com/platformsh/cli/tests/integration/mocks/api"
)

func TestOrgList(t *testing.T) {
	authServer := mocks.APITokenServer(t)
	defer authServer.Close()

	makeOrg := func(id, name, label, owner string) *api.Org {
		return &api.Org{
			ID:    id,
			Name:  name,
			Label: label,
			Owner: owner,
			Links: api.MakeHALLinks("self=/organizations/" + url.PathEscape(id)),
		}
	}

	myUserID := "user-id-1"

	apiHandler := api.NewHandler(t)
	apiHandler.SetMyUser(&api.User{ID: myUserID})
	apiHandler.SetOrgs([]*api.Org{
		makeOrg("org-id-1", "acme", "ACME Inc.", myUserID),
		makeOrg("org-id-2", "four-seasons", "Four Seasons Total Landscaping", myUserID),
		makeOrg("org-id-3", "duff", "Duff Beer", "user-id-2"),
	})

	apiServer := httptest.NewServer(apiHandler)
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
