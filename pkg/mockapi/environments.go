package mockapi

import (
	"crypto/rand"
	"encoding/json"
	"net/http"
	"slices"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

func (h *Handler) handleListEnvironments(w http.ResponseWriter, req *http.Request) {
	h.store.RLock()
	defer h.store.RUnlock()
	projectID := chi.URLParam(req, "project_id")
	var envs []*Environment
	for _, e := range h.store.environments {
		if e.Project == projectID {
			envs = append(envs, e)
		}
	}
	_ = json.NewEncoder(w).Encode(envs)
}

func (h *Handler) handleGetEnvironment(w http.ResponseWriter, req *http.Request) {
	h.store.RLock()
	defer h.store.RUnlock()
	projectID := chi.URLParam(req, "project_id")
	environmentID := chi.URLParam(req, "environment_id")
	for id, e := range h.store.environments {
		if e.Project == projectID && id == environmentID {
			_ = json.NewEncoder(w).Encode(e)
			break
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func (h *Handler) handlePatchEnvironment(w http.ResponseWriter, req *http.Request) {
	env := h.findEnvironment(chi.URLParam(req, "project_id"), chi.URLParam(req, "environment_id"))
	if env == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	h.store.Lock()
	defer h.store.Unlock()

	patched := *env
	err := json.NewDecoder(req.Body).Decode(&patched)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	patched.UpdatedAt = time.Now()
	h.store.environments[patched.ID] = &patched
	_ = json.NewEncoder(w).Encode(&patched)
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

func (h *Handler) handleCreateBackup(w http.ResponseWriter, req *http.Request) {
	projectID := chi.URLParam(req, "project_id")
	environmentID := chi.URLParam(req, "environment_id")
	var options = struct {
		Safe bool `json:"safe"`
	}{}
	require.NoError(h.t, json.NewDecoder(req.Body).Decode(&options))
	backup := &Backup{
		ID:            ulid.MustNew(ulid.Now(), rand.Reader).String(),
		EnvironmentID: environmentID,
		Status:        "CREATED",
		Safe:          options.Safe,
		Restorable:    true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	h.addProjectBackup(projectID, backup)
	_ = json.NewEncoder(w).Encode(backup)
}

func (h *Handler) handleListBackups(w http.ResponseWriter, req *http.Request) {
	h.store.RLock()
	defer h.store.RUnlock()
	projectID := chi.URLParam(req, "project_id")
	environmentID := chi.URLParam(req, "environment_id")
	var backups []*Backup
	if projectBackups, ok := h.store.projectBackups[projectID]; ok {
		for _, b := range projectBackups {
			if b.EnvironmentID == environmentID {
				backups = append(backups, b)
			}
		}
	}
	// Sort backups in descending order by created date.
	slices.SortFunc(backups, func(a, b *Backup) int { return -timeCompare(a.CreatedAt, b.CreatedAt) })
	_ = json.NewEncoder(w).Encode(backups)
}

func timeCompare(a, b time.Time) int {
	if a.Equal(b) {
		return 0
	}
	if a.Before(b) {
		return -1
	}
	return 1
}
