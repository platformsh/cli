package api

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"

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

		EventualProjectID: projectID,
	}

	h.store.mux.Lock()
	defer h.store.mux.Unlock()

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
	}
	_ = json.NewEncoder(w).Encode(sub)
}

func (h *Handler) handleGetSubscription(w http.ResponseWriter, req *http.Request) {
	h.store.mux.Lock()
	defer h.store.mux.Unlock()
	id := chi.URLParam(req, "id")
	sub := h.store.subscriptions[id]
	if sub == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	sub.Status = "active"
	sub.ProjectID = sub.EventualProjectID
	sub.ProjectUI = "http://console.example.com/projects/" + url.PathEscape(sub.EventualProjectID)
	h.store.subscriptions[id] = sub
	_ = json.NewEncoder(w).Encode(sub)
}
