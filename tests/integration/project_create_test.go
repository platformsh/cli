package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/platformsh/cli/tests/integration/mocks"
)

type halLink struct {
	HREF string `json:"href"`
}

type halLinks struct {
	Self halLink `json:"self"`
}

type subscription struct {
	ID            string   `json:"id"`
	Links         halLinks `json:"_links"`
	ProjectID     string   `json:"project_id"`
	ProjectRegion string   `json:"project_region"`
	ProjectTitle  string   `json:"project_title"`
	Status        string   `json:"status"`
	ProjectUI     string   `json:"project_ui"`

	eventualProjectID string
}

type subscriptionsCache struct {
	subscriptions map[string]*subscription
	sync.RWMutex
}

func (c *subscriptionsCache) upsert(s *subscription) {
	c.Lock()
	defer c.Unlock()
	if c.subscriptions == nil {
		c.subscriptions = make(map[string]*subscription)
	}
	c.subscriptions[s.ID] = s
}

func (c *subscriptionsCache) get(id string) *subscription {
	c.RLock()
	defer c.RUnlock()
	return c.subscriptions[id]
}

func projectCreateServer(t *testing.T) *httptest.Server {
	var subscriptions subscriptionsCache
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if testing.Verbose() {
			t.Log(req)
		}
		switch {
		case req.Method == http.MethodPost &&
			regexp.MustCompile(`^/organizations/[a-z0-9-]+/subscriptions$`).MatchString(req.URL.Path):
			var createOptions = struct {
				Region string `json:"project_region"`
				Title  string `json:"project_title"`
			}{}
			err := json.NewDecoder(req.Body).Decode(&createOptions)
			require.NoError(t, err)
			id := fmt.Sprint(rand.Int()) //nolint:gosec
			s := subscription{
				ID:                "s" + id,
				Links:             halLinks{Self: halLink{HREF: "/subscriptions/" + url.PathEscape("s"+id)}},
				ProjectRegion:     createOptions.Region,
				ProjectTitle:      createOptions.Title,
				Status:            "provisioning",
				eventualProjectID: "p" + id,
			}
			subscriptions.upsert(&s)
			_ = json.NewEncoder(w).Encode(s)
			return
		case req.URL.Path == "/me/verification":
			_ = json.NewEncoder(w).Encode(map[string]any{"state": false, "type": ""})
			return
		case req.Method == http.MethodGet && regexp.MustCompile(`^/subscriptions/[a-z0-9-]+$`).MatchString(req.URL.Path):
			s := subscriptions.get(path.Base(req.URL.Path))
			if s == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			s.Status = "active"
			s.ProjectID = s.eventualProjectID
			s.ProjectUI = "http://console.example.com/projects/" + url.PathEscape(s.eventualProjectID)
			subscriptions.upsert(s)
			_ = json.NewEncoder(w).Encode(s)
			return
		case req.Method == http.MethodGet && req.URL.Path == "/organizations/name=cli-tests":
			type org struct {
				ID    string   `json:"id"`
				Name  string   `json:"name"`
				Label string   `json:"label"`
				Links halLinks `json:"_links"`
			}
			_ = json.NewEncoder(w).Encode(org{
				ID:    "cli-tests",
				Name:  "cli-tests",
				Label: "CLI Test Organization",
				Links: halLinks{Self: halLink{"/organizations/cli-tests"}},
			})
			return
		case req.Method == http.MethodGet && req.URL.Path == "/organizations/cli-tests/setup/options":
			type options struct {
				Plans   []string `json:"plans"`
				Regions []string `json:"regions"`
			}
			_ = json.NewEncoder(w).Encode(options{[]string{"development"}, []string{"test-region"}})
			return
		case req.Method == http.MethodGet && regexp.MustCompile(`^/projects/[a-z0-9-]+$`).MatchString(req.URL.Path):
			type repoInfo struct {
				URL string `json:"url"`
			}
			type project struct {
				ID         string   `json:"id"`
				Repository repoInfo `json:"repository"`
				Links      halLinks `json:"_links"`
			}
			projectID := path.Base(req.URL.Path)
			_ = json.NewEncoder(w).Encode(project{
				ID:         path.Base(req.URL.Path),
				Links:      halLinks{Self: halLink{req.URL.Path}},
				Repository: repoInfo{projectID + "@git.example.com:" + projectID + ".git"},
			})
			return
		case req.Method == http.MethodGet && req.URL.Path == "/regions":
			type region struct {
				ID             string `json:"id"`
				Label          string `json:"label"`
				SelectionLabel string `json:"selection_label"`
				Available      bool   `json:"available"`
			}
			type regions struct {
				Regions []region `json:"regions"`
			}
			_ = json.NewEncoder(w).Encode(regions{[]region{{
				ID:             "test-region",
				Label:          "Test Region",
				SelectionLabel: "Test Region",
				Available:      true,
			}}})
			return
		case req.Method == http.MethodGet &&
			regexp.MustCompile(`^/organizations/[a-z0-9-]+/subscriptions/estimate$`).MatchString(req.URL.Path):
			_ = json.NewEncoder(w).Encode(map[string]any{"total": "$1,000 USD"})
			return
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestProjectCreate(t *testing.T) {
	authServer := mocks.APITokenServer(t)
	defer authServer.Close()

	apiServer := projectCreateServer(t)
	defer apiServer.Close()

	title := "Test Project Title"
	region := "test-region"

	cmd := command(t, "project:create", "-v", "--region", region, "--title", title, "--org", "cli-tests")
	cmd.Env = append(
		cmd.Env,
		EnvPrefix+"API_BASE_URL="+apiServer.URL,
		EnvPrefix+"API_AUTH_URL="+authServer.URL,
		EnvPrefix+"TOKEN=api-token-1",
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
