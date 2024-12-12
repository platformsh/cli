package mockapi

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func (h *Handler) handleCreateSubscription(w http.ResponseWriter, req *http.Request) {
	var createOptions = struct {
		Region string `json:"project_region"`
		Title  string `json:"project_title"`
	}{}
	err := json.NewDecoder(req.Body).Decode(&createOptions)
	require.NoError(h.t, err)
	id := fmt.Sprint(rand.Int()) //nolint:gosec
	projectID := "p" + id
	sub := Subscription{
		ID:            "s" + id,
		Links:         MakeHALLinks("self=" + "/subscriptions/" + url.PathEscape("s"+id)),
		ProjectRegion: createOptions.Region,
		ProjectTitle:  createOptions.Title,
		Status:        "provisioning",
	}

	h.store.Lock()
	if h.store.subscriptions == nil {
		h.store.subscriptions = make(map[string]*Subscription)
	}
	h.store.subscriptions[sub.ID] = &sub
	if h.store.projects == nil {
		h.store.projects = make(map[string]*Project)
	}
	h.store.projects[projectID] = &Project{
		ID:             projectID,
		Links:          MakeHALLinks("self=/projects/" + projectID),
		Repository:     ProjectRepository{URL: projectID + "@git.example.com:" + projectID + ".git"},
		SubscriptionID: sub.ID,
		Organization:   chi.URLParam(req, "id"),
	}
	h.store.Unlock()

	_ = json.NewEncoder(w).Encode(sub)

	// Imitate "provisioning": wait a little and then activate.
	go func(subID string, projectID string) {
		time.Sleep(time.Second * 2)
		h.store.Lock()
		defer h.store.Unlock()
		sub := h.store.subscriptions[subID]
		sub.Status = "active"
		sub.ProjectID = projectID
		sub.ProjectUI = "http://console.example.com/projects/" + url.PathEscape(projectID)
	}(sub.ID, projectID)
}

func (h *Handler) handleGetSubscription(w http.ResponseWriter, req *http.Request) {
	h.store.RLock()
	defer h.store.RUnlock()
	id := chi.URLParam(req, "id")
	sub := h.store.subscriptions[id]
	if sub == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(sub)
}

func (h *Handler) handleCanCreateSubscriptions(w http.ResponseWriter, req *http.Request) {
	h.store.RLock()
	defer h.store.RUnlock()
	id := chi.URLParam(req, "id")
	cc := h.store.canCreate[id]
	if cc == nil {
		cc = &CanCreateResponse{CanCreate: true}
	}
	_ = json.NewEncoder(w).Encode(cc)
}
