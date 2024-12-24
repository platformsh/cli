package mockapi

import (
	"encoding/json"
	"net/http"
	"slices"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) handleListProjectActivities(w http.ResponseWriter, req *http.Request) {
	h.store.RLock()
	defer h.store.RUnlock()
	projectID := chi.URLParam(req, "project_id")
	var activities = make([]*Activity, 0, len(h.store.activities[projectID]))
	for _, a := range h.store.activities[projectID] {
		activities = append(activities, a)
	}
	// Sort activities in descending order by created date.
	slices.SortFunc(activities, func(a, b *Activity) int { return -timeCompare(a.CreatedAt, b.CreatedAt) })
	_ = json.NewEncoder(w).Encode(activities)
}

func (h *Handler) handleGetProjectActivity(w http.ResponseWriter, req *http.Request) {
	h.store.RLock()
	defer h.store.RUnlock()
	projectID := chi.URLParam(req, "project_id")
	activityID := chi.URLParam(req, "id")
	if projectActivities := h.store.activities[projectID]; projectActivities != nil {
		_ = json.NewEncoder(w).Encode(projectActivities[activityID])
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func (h *Handler) handleListEnvironmentActivities(w http.ResponseWriter, req *http.Request) {
	h.store.RLock()
	defer h.store.RUnlock()
	projectID := chi.URLParam(req, "project_id")
	environmentID := chi.URLParam(req, "environment_id")
	var activities = make([]*Activity, 0, len(h.store.activities[projectID]))
	for _, a := range h.store.activities[projectID] {
		if slices.Contains(a.Environments, environmentID) {
			activities = append(activities, a)
		}
	}
	// Sort activities in descending order by created date.
	slices.SortFunc(activities, func(a, b *Activity) int { return -timeCompare(a.CreatedAt, b.CreatedAt) })
	_ = json.NewEncoder(w).Encode(activities)
}

func (h *Handler) handleGetEnvironmentActivity(w http.ResponseWriter, req *http.Request) {
	h.store.RLock()
	defer h.store.RUnlock()
	projectID := chi.URLParam(req, "project_id")
	activityID := chi.URLParam(req, "id")
	if projectActivities := h.store.activities[projectID]; projectActivities != nil {
		environmentID := chi.URLParam(req, "environment_id")
		a := projectActivities[activityID]
		if a == nil || !slices.Contains(a.Environments, environmentID) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(a)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}
