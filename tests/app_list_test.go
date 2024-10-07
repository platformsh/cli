package tests

import (
	"bytes"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/platformsh/cli/internal/mockapi"
)

func TestAppList(t *testing.T) {
	authServer := mockapi.NewAuthServer(t)
	defer authServer.Close()

	apiHandler := mockapi.NewHandler(t)

	apiServer := httptest.NewServer(apiHandler)
	defer apiServer.Close()

	apiHandler.SetProjects([]*mockapi.Project{{
		ID: mockProjectID,
		Links: mockapi.MakeHALLinks("self=/projects/"+mockProjectID,
			"environments=/projects/"+mockProjectID+"/environments"),
		DefaultBranch: "main",
	}})

	main := makeEnv(mockProjectID, "main", "production", "active", nil)
	main.SetCurrentDeployment(&mockapi.Deployment{
		WebApps: map[string]mockapi.App{
			"app": {Name: "app", Type: "golang:1.23", Size: "AUTO"},
		},
		Services: map[string]mockapi.App{},
		Routes:   map[string]any{},
		Workers: map[string]mockapi.Worker{
			"app--worker1": {
				App:    mockapi.App{Name: "app--worker1", Type: "golang:1.23", Size: "AUTO"},
				Worker: mockapi.WorkerInfo{Commands: mockapi.Commands{Start: "sleep 60"}},
			},
		},
		Links: mockapi.MakeHALLinks("self=/projects/" + mockProjectID + "/environments/main/deployment/current"),
	})

	envs := []*mockapi.Environment{
		main,
		makeEnv(mockProjectID, "staging", "staging", "active", "main"),
		makeEnv(mockProjectID, "dev", "development", "active", "staging"),
		makeEnv(mockProjectID, "fix", "development", "inactive", "dev"),
	}

	apiHandler.SetEnvironments(envs)

	run := runnerWithAuth(t, apiServer.URL, authServer.URL)

	assert.Equal(t, strings.TrimLeft(`
Name	Type
app	golang:1.23
`, "\n"), run("apps", "-p", mockProjectID, "-e", ".", "--refresh", "--format", "tsv"))

	assert.Equal(t, strings.TrimLeft(`
+--------------+-------------+-------------------+
| Name         | Type        | Commands          |
+--------------+-------------+-------------------+
| app--worker1 | golang:1.23 | start: 'sleep 60' |
+--------------+-------------+-------------------+
`, "\n"), run("workers", "-v", "-p", mockProjectID, "-e", "."))

	servicesCmd := authenticatedCommand(t, apiServer.URL, authServer.URL,
		"services", "-p", mockProjectID, "-e", "main")
	stdErrBuf := bytes.Buffer{}
	servicesCmd.Stderr = &stdErrBuf
	if testing.Verbose() {
		servicesCmd.Stderr = io.MultiWriter(&stdErrBuf, os.Stderr)
	}
	require.NoError(t, servicesCmd.Run())
	assert.Contains(t, stdErrBuf.String(), "No services found")
}
