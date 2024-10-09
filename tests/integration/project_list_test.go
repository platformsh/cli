package integration

import (
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/platformsh/cli/tests/integration/mocks"
	"github.com/platformsh/cli/tests/integration/mocks/api"
)

func TestProjectList(t *testing.T) {
	authServer := mocks.APITokenServer(t)
	defer authServer.Close()

	myUserID := "my-user-id"
	otherUserID := "other-user-id"
	vendor := "platformsh"

	apiHandler := api.NewHandler(t)
	apiHandler.SetMyUser(&api.User{ID: myUserID})
	apiServer := httptest.NewServer(apiHandler)
	defer apiServer.Close()

	apiHandler.SetOrgs([]*api.Org{
		{
			ID:    "org-id-1",
			Name:  "org-1",
			Label: "Org 1",
			Owner: myUserID,
			Links: api.MakeHALLinks("self=/organizations/org-id-1"),
		},
		{
			ID:    "org-id-2",
			Name:  "org-2",
			Label: "Org 2",
			Owner: otherUserID,
			Links: api.MakeHALLinks("self=/organizations/org-id-2"),
		},
	})
	apiHandler.SetProjects([]*api.Project{
		{
			ID:           "project-id-1",
			Organization: "org-id-1",
			Vendor:       vendor,
			Title:        "Project 1",
			Region:       "region-1",
			Links:        api.MakeHALLinks("self=/projects/project-id-1"),
		},
		{
			ID:           "project-id-2",
			Organization: "org-id-2",
			Vendor:       vendor,
			Title:        "Project 2",
			Region:       "region-2",
			Links:        api.MakeHALLinks("self=/projects/project-id-2"),
		},
		{
			ID:           "project-id-3",
			Organization: "org-id-2",
			Vendor:       vendor,
			Title:        "Project 3",
			Region:       "region-2",
			Links:        api.MakeHALLinks("self=/projects/project-id-3"),
		},
		{
			ID:           "project-other-vendor",
			Organization: "org-other-vendor",
			Vendor:       "acme",
			Title:        "Other Vendor's Project",
			Links:        api.MakeHALLinks("self=/projects/project-other-vendor"),
			Region:       "region-1",
		},
	})
	apiHandler.SetUserGrants([]*api.UserGrant{
		{
			ResourceID:     "org-id-1",
			ResourceType:   "organization",
			OrganizationID: "org-id-1",
			UserID:         myUserID,
			Permissions:    []string{"admin"},
		},
		{
			ResourceID:     "project-id-1",
			ResourceType:   "project",
			OrganizationID: "org-id-1",
			UserID:         myUserID,
			Permissions:    []string{"admin"},
		},
		{
			ResourceID:     "project-id-2",
			ResourceType:   "project",
			OrganizationID: "org-id-2",
			UserID:         "user-id-2",
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
	})

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
