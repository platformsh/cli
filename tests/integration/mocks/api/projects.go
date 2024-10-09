package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func (h *Handler) handleProjectRefs(w http.ResponseWriter, req *http.Request) {
	h.store.mux.RLock()
	defer h.store.mux.RUnlock()
	require.NoError(h.t, req.ParseForm())
	ids := strings.Split(req.Form.Get("in"), ",")
	refs := make(map[string]*ProjectRef, len(ids))
	for _, id := range ids {
		if p, ok := h.store.projects[id]; ok {
			refs[id] = p.AsRef()
		} else {
			refs[id] = nil
		}
	}
	_ = json.NewEncoder(w).Encode(refs)
}

func (h *Handler) handleGetProject(w http.ResponseWriter, req *http.Request) {
	h.store.mux.RLock()
	defer h.store.mux.RUnlock()
	projectID := chi.URLParam(req, "id")
	if p, ok := h.store.projects[projectID]; ok {
		_ = json.NewEncoder(w).Encode(p)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func (h *Handler) handleListEnvironments(w http.ResponseWriter, req *http.Request) {
	h.store.mux.RLock()
	defer h.store.mux.RUnlock()
	projectID := chi.URLParam(req, "id")
	var envs []*Environment
	for _, e := range h.store.environments {
		if e.Project == projectID {
			envs = append(envs, e)
		}
	}
	_ = json.NewEncoder(w).Encode(envs)
}

func (h *Handler) handleGetCurrentDeployment(w http.ResponseWriter, req *http.Request) {
	h.store.mux.RLock()
	defer h.store.mux.RUnlock()
	projectID := chi.URLParam(req, "project_id")
	environmentID := chi.URLParam(req, "environment_id")
	var d *Deployment
	for _, e := range h.store.environments {
		if e.Project == projectID && e.ID == environmentID {
			d = e.currentDeployment
		}
	}
	if d == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(d)
}

func (h *Handler) handleListRegions(w http.ResponseWriter, _ *http.Request) {
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
}
