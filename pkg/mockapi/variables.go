package mockapi

import (
	"encoding/json"
	"net/http"
	"net/url"
	"slices"
	"time"

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

func (h *Handler) handleCreateProjectVariable(w http.ResponseWriter, req *http.Request) {
	h.store.Lock()
	defer h.store.Unlock()

	projectID := chi.URLParam(req, "project_id")

	newVar := Variable{}
	if err := json.NewDecoder(req.Body).Decode(&newVar); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	newVar.CreatedAt = time.Now()
	newVar.UpdatedAt = time.Now()
	newVar.Links = MakeHALLinks(
		"self=/projects/"+projectID+"/variables/"+newVar.Name,
		"#edit=/projects/"+projectID+"/variables/"+newVar.Name,
	)

	for _, v := range h.store.projectVariables[projectID] {
		if newVar.Name == v.Name {
			w.WriteHeader(http.StatusConflict)
			return
		}
	}

	if h.store.projectVariables[projectID] == nil {
		h.store.projectVariables = make(map[string][]*Variable)
	}
	h.store.projectVariables[projectID] = append(h.store.projectVariables[projectID], &newVar)

	_ = json.NewEncoder(w).Encode(map[string]any{
		"_embedded": map[string]any{"entity": newVar},
	})
}

func (h *Handler) handlePatchProjectVariable(w http.ResponseWriter, req *http.Request) {
	h.store.Lock()
	defer h.store.Unlock()

	projectID := chi.URLParam(req, "project_id")
	variableName, _ := url.PathUnescape(chi.URLParam(req, "name"))
	var key = -1
	for k, v := range h.store.projectVariables[projectID] {
		if v.Name == variableName {
			key = k
			break
		}
	}
	if key == -1 {
		w.WriteHeader(http.StatusNotFound)
	}
	patched := *h.store.projectVariables[projectID][key]
	err := json.NewDecoder(req.Body).Decode(&patched)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	patched.UpdatedAt = time.Now()
	h.store.projectVariables[projectID][key] = &patched
	_ = json.NewEncoder(w).Encode(&patched)
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

func (h *Handler) handleCreateEnvLevelVariable(w http.ResponseWriter, req *http.Request) {
	h.store.Lock()
	defer h.store.Unlock()

	projectID := chi.URLParam(req, "project_id")
	environmentID, _ := url.PathUnescape(chi.URLParam(req, "environment_id"))

	newVar := EnvLevelVariable{}
	if err := json.NewDecoder(req.Body).Decode(&newVar); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	newVar.CreatedAt = time.Now()
	newVar.UpdatedAt = time.Now()
	newVar.Links = MakeHALLinks(
		"self=/projects/"+projectID+"/environments/"+environmentID+"/variables/"+newVar.Name,
		"#edit=/projects/"+projectID+"/environments/"+environmentID+"/variables/"+newVar.Name,
	)

	for _, v := range h.store.envLevelVariables[projectID][environmentID] {
		if newVar.Name == v.Name {
			w.WriteHeader(http.StatusConflict)
			return
		}
	}

	if h.store.envLevelVariables == nil {
		h.store.envLevelVariables = make(map[string]map[string][]*EnvLevelVariable)
	}
	if h.store.envLevelVariables[projectID] == nil {
		h.store.envLevelVariables[projectID] = make(map[string][]*EnvLevelVariable)
	}
	h.store.envLevelVariables[projectID][environmentID] = append(
		h.store.envLevelVariables[projectID][environmentID],
		&newVar,
	)

	_ = json.NewEncoder(w).Encode(map[string]any{
		"_embedded": map[string]any{"entity": newVar},
	})
}

func (h *Handler) handlePatchEnvLevelVariable(w http.ResponseWriter, req *http.Request) {
	h.store.Lock()
	defer h.store.Unlock()
	projectID := chi.URLParam(req, "project_id")
	environmentID, _ := url.PathUnescape(chi.URLParam(req, "environment_id"))
	variableName, _ := url.PathUnescape(chi.URLParam(req, "name"))
	var key = -1
	for k, v := range h.store.envLevelVariables[projectID][environmentID] {
		if variableName == v.Name {
			key = k
			break
		}
	}
	if key == -1 {
		w.WriteHeader(http.StatusNotFound)
	}
	patched := *h.store.envLevelVariables[projectID][environmentID][key]
	err := json.NewDecoder(req.Body).Decode(&patched)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	patched.UpdatedAt = time.Now()
	h.store.envLevelVariables[projectID][environmentID][key] = &patched
	_ = json.NewEncoder(w).Encode(&patched)
}
