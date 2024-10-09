package integration

import (
	"bytes"
	"io"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/platformsh/cli/tests/integration/mocks"
	"github.com/platformsh/cli/tests/integration/mocks/api"
)

func TestAppList(t *testing.T) {
	authServer := mocks.APITokenServer(t)
	defer authServer.Close()

	apiHandler := api.NewHandler(t)

	apiServer := httptest.NewServer(apiHandler)
	defer apiServer.Close()

	apiHandler.SetProjects([]*api.Project{{
		ID: mockProjectID,
		Links: api.MakeHALLinks("self=/projects/"+mockProjectID,
			"environments=/projects/"+mockProjectID+"/environments"),
		DefaultBranch: "main",
	}})

	main := makeEnv(mockProjectID, "main", "production", "active", nil)
	main.SetCurrentDeployment(&api.Deployment{
		WebApps: map[string]api.App{
			"app": {Name: "app", Type: "golang:1.23", Size: "AUTO"},
		},
		Services: map[string]api.App{},
		Routes:   map[string]any{},
		Workers: map[string]api.Worker{
			"app--worker1": {
				App:    api.App{Name: "app--worker1", Type: "golang:1.23", Size: "AUTO"},
				Worker: api.WorkerInfo{Commands: api.Commands{Start: "sleep 60"}},
			},
		},
		Links: api.MakeHALLinks("self=/projects/" + mockProjectID + "/environments/main/deployment/current"),
	})

	envs := []*api.Environment{
		main,
		makeEnv(mockProjectID, "staging", "staging", "active", "main"),
		makeEnv(mockProjectID, "dev", "development", "active", "staging"),
		makeEnv(mockProjectID, "fix", "development", "inactive", "dev"),
	}

	apiHandler.SetEnvironments(envs)

	authenticatedCommand := func(args ...string) *exec.Cmd {
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
		return cmd
	}

	run := func(args ...string) string {
		b, err := authenticatedCommand(args...).Output()
		require.NoError(t, err)
		return string(b)
	}

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

	servicesCmd := authenticatedCommand("services", "-p", mockProjectID, "-e", "main")
	stdErrBuf := bytes.Buffer{}
	servicesCmd.Stderr = &stdErrBuf
	if testing.Verbose() {
		servicesCmd.Stderr = io.MultiWriter(&stdErrBuf, os.Stderr)
	}
	require.NoError(t, servicesCmd.Run())
	assert.Contains(t, stdErrBuf.String(), "No services found")
}
