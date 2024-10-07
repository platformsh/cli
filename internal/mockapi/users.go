package mockapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func (h *Handler) handleUsersMe(w http.ResponseWriter, _ *http.Request) {
	_ = json.NewEncoder(w).Encode(h.store.myUser)
}

func (h *Handler) handleUserRefs(w http.ResponseWriter, req *http.Request) {
	require.NoError(h.t, req.ParseForm())
	ids := strings.Split(req.Form.Get("in"), ",")
	userRefs := make(map[string]UserRef, len(ids))
	for _, id := range ids {
		userRefs[id] = UserRef{ID: id, Email: id + "@example.com", Username: id}
	}
	_ = json.NewEncoder(w).Encode(userRefs)
}

func (h *Handler) handleUserExtendedAccess(w http.ResponseWriter, req *http.Request) {
	h.store.RLock()
	defer h.store.RUnlock()
	userID := chi.URLParam(req, "id")
	require.NoError(h.t, req.ParseForm())
	require.Equal(h.t, "project", req.Form.Get("filter[resource_type]"))
	var (
		projectGrants = make([]*UserGrant, 0, len(h.store.userGrants))
		projectIDs    = make(uniqueMap)
		orgIDs        = make(uniqueMap)
	)
	for _, g := range h.store.userGrants {
		if g.ResourceType == "project" && g.UserID == userID {
			projectGrants = append(projectGrants, g)
			projectIDs[g.ResourceID] = struct{}{}
			orgIDs[g.OrganizationID] = struct{}{}
		}
	}
	ret := struct {
		Items []*UserGrant `json:"items"`
		Links HalLinks     `json:"_links"`
	}{Items: projectGrants, Links: MakeHALLinks(
		"ref:projects:0=/ref/projects?in="+strings.Join(projectIDs.keys(), ","),
		"ref:organizations:0=/ref/organizations?in="+strings.Join(orgIDs.keys(), ","),
	)}
	_ = json.NewEncoder(w).Encode(ret)
}

type uniqueMap map[string]struct{}

func (m uniqueMap) keys() []string {
	r := make([]string, 0, len(m))
	for k := range m {
		r = append(r, k)
	}
	return r
}
