package integration

import (
	"bytes"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/platformsh/cli/tests/integration/mocks"
	"github.com/platformsh/cli/tests/integration/mocks/api"
)

func TestProjectCreate(t *testing.T) {
	authServer := mocks.APITokenServer(t)
	defer authServer.Close()

	apiHandler := api.NewHandler(t)
	apiHandler.MyUserID = "my-user-id"
	apiHandler.SetOrgs([]*api.Org{{
		ID:    "cli-test-id",
		Name:  "cli-tests",
		Label: "CLI Test Organization",
		Links: api.MakeHALLinks("self=" + "/organizations/cli-test-id"),
	}})

	apiServer := httptest.NewServer(apiHandler)
	defer apiServer.Close()

	title := "Test Project Title"
	region := "test-region"

	cmd := command(t, "project:create", "-v", "--region", region, "--title", title, "--org", "cli-tests")
	cmd.Env = append(
		cmd.Env,
		EnvPrefix+"API_BASE_URL="+apiServer.URL,
		EnvPrefix+"API_AUTH_URL="+authServer.URL,
		EnvPrefix+"TOKEN="+mocks.ValidAPITokens[0],
	)

	var stdErrBuf bytes.Buffer
	var stdOutBuf bytes.Buffer
	cmd.Stderr = &stdErrBuf
	if testing.Verbose() {
		cmd.Stderr = io.MultiWriter(&stdErrBuf, os.Stderr)
	}
	cmd.Stdout = &stdOutBuf
	require.NoError(t, cmd.Run())

	// stdout should contain the project ID.
	projectID := strings.TrimSpace(stdOutBuf.String())
	assert.NotEmpty(t, projectID)

	// stderr should contain various messages.
	stderr := stdErrBuf.String()

	assert.Contains(t, stderr, "The estimated monthly cost of this project is: $1,000 USD")
	assert.Contains(t, stderr, "Region: "+region)
	assert.Contains(t, stderr, "Project ID: "+projectID)
	assert.Contains(t, stderr, "Project title: "+title)
}
