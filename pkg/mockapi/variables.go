package mockapi

import (
	"encoding/json"
	"net/http"
	"net/url"
	"slices"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) handleListProjectVariables(w http.ResponseWriter, req *http.Request) {
	h.store.RLock()
	defer h.store.RUnlock()
	projectID := chi.URLParam(req, "project_id")
	variables := h.store.projectVariables[projectID]
	// Sort variables in descending order by created date.
	slices.SortFunc(variables, func(a, b *Variable) int { return -timeCompare(a.CreatedAt, b.CreatedAt) })
	_ = json.NewEncoder(w).Encode(variables)
}

func (h *Handler) handleGetProjectVariable(w http.ResponseWriter, req *http.Request) {
	h.store.RLock()
	defer h.store.RUnlock()
	projectID := chi.URLParam(req, "project_id")
	variableName, _ := url.PathUnescape(chi.URLParam(req, "name"))
	for _, v := range h.store.projectVariables[projectID] {
		if variableName == v.Name {
			_ = json.NewEncoder(w).Encode(v)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func (h *Handler) handleListEnvLevelVariables(w http.ResponseWriter, req *http.Request) {
	h.store.RLock()
	defer h.store.RUnlock()
	projectID := chi.URLParam(req, "project_id")
	environmentID, _ := url.PathUnescape(chi.URLParam(req, "environment_id"))
	variables := h.store.envLevelVariables[projectID][environmentID]
	// Sort variables in descending order by created date.
	slices.SortFunc(variables, func(a, b *EnvLevelVariable) int { return -timeCompare(a.CreatedAt, b.CreatedAt) })
	_ = json.NewEncoder(w).Encode(variables)
}

func (h *Handler) handleGetEnvLevelVariable(w http.ResponseWriter, req *http.Request) {
	h.store.RLock()
	defer h.store.RUnlock()
	projectID := chi.URLParam(req, "project_id")
	environmentID, _ := url.PathUnescape(chi.URLParam(req, "environment_id"))
	variableName, _ := url.PathUnescape(chi.URLParam(req, "name"))
	for _, v := range h.store.envLevelVariables[projectID][environmentID] {
		if variableName == v.Name {
			_ = json.NewEncoder(w).Encode(v)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}
