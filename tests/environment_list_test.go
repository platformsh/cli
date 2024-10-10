package tests

import (
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/platformsh/cli/internal/mock"
	"github.com/platformsh/cli/internal/mock/api"
)

func TestEnvironmentList(t *testing.T) {
	authServer := mock.NewAuthServer(t)
	defer authServer.Close()

	apiHandler := api.NewHandler(t)
	apiServer := httptest.NewServer(apiHandler)
	defer apiServer.Close()

	apiHandler.SetProjects([]*api.Project{
		{
			ID:    mockProjectID,
			Links: api.MakeHALLinks("self=/projects/"+mockProjectID, "environments=/projects/"+mockProjectID+"/environments"),
		},
	})
	apiHandler.SetEnvironments([]*api.Environment{
		makeEnv(mockProjectID, "main", "production", "active", nil),
		makeEnv(mockProjectID, "staging", "staging", "active", "main"),
		makeEnv(mockProjectID, "dev", "development", "active", "staging"),
		makeEnv(mockProjectID, "fix", "development", "inactive", "dev"),
	})

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

func makeEnv(projectID, name, envType, status string, parent any) *api.Environment {
	return &api.Environment{
		ID:          name,
		Name:        name,
		MachineName: name + "-xyz",
		Title:       strings.ToTitle(name[:1]) + name[1:],
		Parent:      parent,
		Type:        envType,
		Status:      status,
		Project:     projectID,
		Links: api.MakeHALLinks(
			"self=/projects/" + url.PathEscape(projectID) + "/environments/" + url.PathEscape(name),
		),
	}
}
