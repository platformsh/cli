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

func TestProjectCreate(t *testing.T) {
	authServer := mockapi.NewAuthServer(t)
	defer authServer.Close()

	apiHandler := mockapi.NewHandler(t)
	apiHandler.SetOrgs([]*mockapi.Org{
		makeOrg("cli-test-id", "cli-tests", "CLI Test Org", "my-user-id"),
	})

	apiServer := httptest.NewServer(apiHandler)
	defer apiServer.Close()

	title := "Test Project Title"
	region := "test-region"

	cmd := authenticatedCommand(t, apiServer.URL, authServer.URL,
		"project:create", "-v", "--region", region, "--title", title, "--org", "cli-tests")

	var stdErrBuf bytes.Buffer
	var stdOutBuf bytes.Buffer
	cmd.Stderr = &stdErrBuf
	if testing.Verbose() {
		cmd.Stderr = io.MultiWriter(&stdErrBuf, os.Stderr)
	}
	cmd.Stdout = &stdOutBuf
	t.Log("Running:", cmd)
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
