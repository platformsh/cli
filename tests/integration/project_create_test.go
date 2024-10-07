package integration

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectCreate(t *testing.T) {
	if os.Getenv(EnvPrefix+"TOKEN") == "" {
		t.Skip("enable by setting " + EnvPrefix + "TOKEN")
	}
	region := "eu-3.platform.sh"
	if v := os.Getenv("TEST_REGION"); v != "" {
		region = v
	}
	org := "cli-tests"
	if v := os.Getenv("TEST_ORG"); v != "" {
		region = v
	}
	cmd := command(t, "project:create", "--region="+region, "--title=Automated_Test", "--org="+org, "--no-set-remote")
	var stdErrBuf bytes.Buffer
	var stdOutBuf bytes.Buffer
	cmd.Stderr = &stdErrBuf
	cmd.Stdout = &stdOutBuf
	if err := cmd.Run(); err != nil {
		require.NoError(t, err, "stderr output: "+stdErrBuf.String())
	}

	// stdout should contain the project ID.
	assert.Greater(t, stdOutBuf.Len(), 10)
	stdout := stdOutBuf.String()
	projectID := stdout

	// stderr should contain various messages.
	stderr := stdErrBuf.String()
	assert.Contains(t, stderr, "Project ID: "+projectID)
}
