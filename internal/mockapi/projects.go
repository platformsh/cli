package mockapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func (h *Handler) handleProjectRefs(w http.ResponseWriter, req *http.Request) {
	h.store.RLock()
	defer h.store.RUnlock()
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
	h.store.RLock()
	defer h.store.RUnlock()
	projectID := chi.URLParam(req, "id")
	if p, ok := h.store.projects[projectID]; ok {
		_ = json.NewEncoder(w).Encode(p)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func (h *Handler) handlePatchProject(w http.ResponseWriter, req *http.Request) {
	h.store.Lock()
	defer h.store.Unlock()
	projectID := chi.URLParam(req, "id")
	p, ok := h.store.projects[projectID]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	patched := *p
	err := json.NewDecoder(req.Body).Decode(&patched)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	patched.UpdatedAt = time.Now()
	h.store.projects[projectID] = &patched
	_ = json.NewEncoder(w).Encode(&patched)
}

func (h *Handler) handleListEnvironments(w http.ResponseWriter, req *http.Request) {
	h.store.RLock()
	defer h.store.RUnlock()
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
	h.store.RLock()
	defer h.store.RUnlock()
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

func (h *Handler) handleProjectUserAccess(w http.ResponseWriter, req *http.Request) {
	h.store.RLock()
	defer h.store.RUnlock()
	projectID := chi.URLParam(req, "project_id")
	require.NoError(h.t, req.ParseForm())
	var (
		projectGrants = make([]*ProjectUserGrant, 0, len(h.store.userGrants))
		userIDs       = make(uniqueMap)
		orgIDs        = make(uniqueMap)
	)
	for _, g := range h.store.userGrants {
		if g.ResourceType == "project" && g.ResourceID == projectID {
			projectGrants = append(projectGrants, &ProjectUserGrant{
				ProjectID:      g.ResourceID,
				OrganizationID: g.OrganizationID,
				UserID:         g.UserID,
				Permissions:    g.Permissions,
				GrantedAt:      g.GrantedAt,
				UpdatedAt:      g.UpdatedAt,
			})
			userIDs[g.UserID] = struct{}{}
			orgIDs[g.OrganizationID] = struct{}{}
		}
	}
	ret := struct {
		Items []*ProjectUserGrant `json:"items"`
		Links HalLinks            `json:"_links"`
	}{Items: projectGrants, Links: MakeHALLinks(
		"ref:users:0=/ref/users?in="+strings.Join(userIDs.keys(), ","),
		"ref:organizations:0=/ref/organizations?in="+strings.Join(orgIDs.keys(), ","),
	)}
	_ = json.NewEncoder(w).Encode(ret)
}
