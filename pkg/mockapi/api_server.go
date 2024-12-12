// Package mockapi provides mocks of the HTTP API for use in integration tests.
package mockapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/require"
)

type Handler struct {
	*chi.Mux

	t *testing.T

	store
}

func NewHandler(t *testing.T) *Handler {
	h := &Handler{t: t}
	h.Mux = chi.NewRouter()

	if testing.Verbose() {
		h.Mux.Use(middleware.DefaultLogger)
	}

	h.Mux.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			authHeader := req.Header.Get("Authorization")
			require.NotEmpty(t, authHeader)
			require.True(t, strings.HasPrefix(authHeader, "Bearer "))
			next.ServeHTTP(w, req)
		})
	})

	h.Mux.Get("/users/me", h.handleUsersMe)
	h.Mux.Get("/users/{user_id}/extended-access", h.handleUserExtendedAccess)
	h.Mux.Get("/ref/users", h.handleUserRefs)
	h.Mux.Post("/me/verification", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"state": false, "type": ""})
	})

	h.Mux.Get("/organizations", h.handleListOrgs)
	h.Mux.Post("/organizations", h.handleCreateOrg)
	h.Mux.Get("/organizations/{organization_id}", h.handleGetOrg)
	h.Mux.Patch("/organizations/{organization_id}", h.handlePatchOrg)
	h.Mux.Get("/users/{user_id}/organizations", h.handleListOrgs)
	h.Mux.Get("/ref/organizations", h.handleOrgRefs)

	h.Mux.Post("/organizations/{organization_id}/subscriptions", h.handleCreateSubscription)
	h.Mux.Get("/subscriptions/{subscription_id}", h.handleGetSubscription)
	h.Mux.Get("/organizations/{organization_id}/subscriptions/can-create", h.handleCanCreateSubscriptions)
	h.Mux.Get("/organizations/{organization_id}/setup/options", func(w http.ResponseWriter, _ *http.Request) {
		type options struct {
			Plans   []string `json:"plans"`
			Regions []string `json:"regions"`
		}
		_ = json.NewEncoder(w).Encode(options{[]string{"development"}, []string{"test-region"}})
	})
	h.Mux.Get("/organizations/{organization_id}/subscriptions/estimate", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"total": "$1,000 USD"})
	})

	h.Mux.Get("/projects/{project_id}", h.handleGetProject)
	h.Mux.Patch("/projects/{project_id}", h.handlePatchProject)
	h.Mux.Get("/projects/{project_id}/environments", h.handleListEnvironments)
	h.Mux.Get("/projects/{project_id}/environments/{environment_id}", h.handleGetEnvironment)
	h.Mux.Get("/projects/{project_id}/environments/{environment_id}/backups", h.handleListBackups)
	h.Mux.Post("/projects/{project_id}/environments/{environment_id}/backups", h.handleCreateBackup)
	h.Mux.Get("/projects/{project_id}/environments/{environment_id}/deployments/current", h.handleGetCurrentDeployment)
	h.Mux.Get("/projects/{project_id}/user-access", h.handleProjectUserAccess)
	h.Mux.Get("/ref/projects", h.handleProjectRefs)

	h.Mux.Get("/regions", h.handleListRegions)

	return h
}
