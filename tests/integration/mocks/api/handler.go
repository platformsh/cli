package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	h.Mux.Get("/users/me", h.handleUsersMe)
	h.Mux.Get("/users/{id}/extended-access", h.handleUserExtendedAccess)
	h.Mux.Get("/ref/users", h.handleUserRefs)
	h.Mux.Post("/me/verification", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"state": false, "type": ""})
	})

	h.Mux.Get("/organizations", h.handleListOrgs)
	h.Mux.Get("/organizations/{id}", h.handleGetOrg)
	h.Mux.Get("/users/{id}/organizations", h.handleListOrgs)
	h.Mux.Get("/ref/organizations", h.handleOrgRefs)

	h.Mux.Post("/organizations/{id}/subscriptions", h.handleCreateSubscription)
	h.Mux.Get("/subscriptions/{id}", h.handleGetSubscription)
	h.Mux.Get("/organizations/{id}/setup/options", func(w http.ResponseWriter, _ *http.Request) {
		type options struct {
			Plans   []string `json:"plans"`
			Regions []string `json:"regions"`
		}
		_ = json.NewEncoder(w).Encode(options{[]string{"development"}, []string{"test-region"}})
	})
	h.Mux.Get("/organizations/{id}/subscriptions/estimate", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"total": "$1,000 USD"})
	})

	h.Mux.Get("/projects/{id}", h.handleGetProject)
	h.Mux.Get("/projects/{id}/environments", h.handleListEnvironments)
	h.Mux.Get("/ref/projects", h.handleProjectRefs)

	h.Mux.Get("/regions", h.handleListRegions)

	return h
}
