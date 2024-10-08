package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"

	"github.com/platformsh/cli/tests/integration/mocks"
)

func TestProjectList(t *testing.T) {
	authServer := mocks.APITokenServer(t)
	defer authServer.Close()

	myUserID := "my-user-id"
	otherUserID := "other-user-id"
	vendor := "platformsh"
	orgs := map[string]org{
		"org-id-1": {
			ID:    "org-id-1",
			Name:  "org-1",
			Label: "Org 1",
			Owner: myUserID,
			Links: makeHALLinks("self=/organizations/org-id-1"),
		},
		"org-id-2": {
			ID:    "org-id-2",
			Name:  "org-2",
			Label: "Org 2",
			Owner: otherUserID,
			Links: makeHALLinks("self=/organizations/org-id-2"),
		},
	}
	projects := map[string]project{
		"project-id-1": {
			ID:           "project-id-1",
			Organization: "org-id-1",
			Vendor:       vendor,
			Title:        "Project 1",
			Region:       "region-1",
			Links:        makeHALLinks("self=/projects/project-id-1"),
		},
		"project-id-2": {
			ID:           "project-id-2",
			Organization: "org-id-2",
			Vendor:       vendor,
			Title:        "Project 2",
			Region:       "region-2",
			Links:        makeHALLinks("self=/projects/project-id-2"),
		},
		"project-id-3": {
			ID:           "project-id-3",
			Organization: "org-id-2",
			Vendor:       vendor,
			Title:        "Project 3",
			Region:       "region-2",
			Links:        makeHALLinks("self=/projects/project-id-3"),
		},
		"project-other-vendor": {
			ID:           "project-other-vendor",
			Organization: "org-other-vendor",
			Vendor:       "acme",
			Title:        "Other Vendor's Project",
			Links:        makeHALLinks("self=/projects/project-other-vendor"),
			Region:       "region-1",
		},
	}
	userGrants := []*userGrant{
		{
			ResourceID:     "org-id-1",
			ResourceType:   "organization",
			OrganizationID: "org-id-1",
			UserID:         orgs["org-id-1"].Owner,
			Permissions:    []string{"admin"},
		},
		{
			ResourceID:     "project-id-1",
			ResourceType:   "project",
			OrganizationID: "org-id-1",
			UserID:         orgs["org-id-1"].Owner,
			Permissions:    []string{"admin"},
		},
		{
			ResourceID:     "project-id-2",
			ResourceType:   "project",
			OrganizationID: "org-id-2",
			UserID:         orgs["org-id-2"].Owner,
			Permissions:    []string{"admin"},
		},
		{
			ResourceID:     "project-id-2",
			ResourceType:   "project",
			OrganizationID: "org-id-2",
			UserID:         myUserID,
			Permissions:    []string{"viewer", "development:admin"},
		},
		{
			ResourceID:     "project-id-3",
			ResourceType:   "project",
			OrganizationID: "org-id-2",
			UserID:         myUserID,
			Permissions:    []string{"viewer", "development:contributor"},
		},
	}

	apiServer := projectListServer(t, myUserID, orgs, projects, userGrants)
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
+--------------+-----------+----------+--------------+
| ID           | Title     | Region   | Organization |
+--------------+-----------+----------+--------------+
| project-id-1 | Project 1 | region-1 | org-1        |
| project-id-2 | Project 2 | region-2 | org-2        |
| project-id-3 | Project 3 | region-2 | org-2        |
+--------------+-----------+----------+--------------+
`, "\n"), run("pro", "-v"))

	assert.Equal(t, strings.TrimLeft(`
ID	Title	Region	Organization
project-id-1	Project 1	region-1	org-1
project-id-2	Project 2	region-2	org-2
project-id-3	Project 3	region-2	org-2
`, "\n"), run("pro", "-v", "--format", "plain"))

	assert.Equal(t, strings.TrimLeft(`
ID,Organization ID
project-id-1,org-id-1
project-id-2,org-id-2
project-id-3,org-id-2
`, "\n"), run("pro", "-v", "--format", "csv", "--columns", "id,organization_id"))

	assert.Equal(t, strings.TrimLeft(`
ID	Title	Region	Organization
project-id-1	Project 1	region-1	org-1
`, "\n"), run("pro", "-v", "--format", "plain", "--my"))

	assert.Equal(t, strings.TrimLeft(`
project-id-1
project-id-2
project-id-3
`, "\n"), run("pro", "-v", "--pipe"))
}

func projectListServer(t *testing.T, myUserID string, orgs map[string]org, projects map[string]project, userGrants []*userGrant) *httptest.Server { //nolint:lll
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if testing.Verbose() {
			t.Log(req)
		}
		switch {
		case req.Method == http.MethodGet &&
			regexp.MustCompile(`^/users/[a-z0-9-]+/extended-access$`).MatchString(req.URL.Path):
			parts := strings.SplitN(req.URL.Path, "/", 4)
			userID := parts[2]
			require.NoError(t, req.ParseForm())
			require.Equal(t, "project", req.Form.Get("filter[resource_type]"))
			var (
				projectGrants = make([]*userGrant, 0, len(userGrants))
				projectIDs    = make(map[string]struct{})
				orgIDs        = make(map[string]struct{})
			)
			for _, g := range userGrants {
				if g.ResourceType == "project" && g.UserID == userID {
					projectGrants = append(projectGrants, g)
					projectIDs[g.ResourceID] = struct{}{}
					orgIDs[g.OrganizationID] = struct{}{}
				}
			}
			ret := struct {
				Items []*userGrant `json:"items"`
				Links halLinks     `json:"_links"`
			}{Items: projectGrants, Links: makeHALLinks(
				"ref:projects:0=/ref/projects?in="+strings.Join(maps.Keys(projectIDs), ","),
				"ref:organizations:0=/ref/organizations?in="+strings.Join(maps.Keys(orgIDs), ","),
			)}
			_ = json.NewEncoder(w).Encode(ret)
			return
		case req.Method == http.MethodGet && req.URL.Path == "/users/me":
			_ = json.NewEncoder(w).Encode(map[string]string{"id": myUserID})
		case req.Method == http.MethodGet && req.URL.Path == "/ref/organizations":
			require.NoError(t, req.ParseForm())
			ids := strings.Split(req.Form.Get("in"), ",")
			refs := make(map[string]*orgRef, len(ids))
			for _, id := range ids {
				if o, ok := orgs[id]; ok {
					refs[id] = o.asRef()
				} else {
					refs[id] = nil
				}
			}
			_ = json.NewEncoder(w).Encode(refs)
		case req.Method == http.MethodGet && req.URL.Path == "/ref/projects":
			require.NoError(t, req.ParseForm())
			ids := strings.Split(req.Form.Get("in"), ",")
			refs := make(map[string]*projectRef, len(ids))
			for _, id := range ids {
				if p, ok := projects[id]; ok {
					refs[id] = p.asRef()
				} else {
					refs[id] = nil
				}
			}
			_ = json.NewEncoder(w).Encode(refs)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}
