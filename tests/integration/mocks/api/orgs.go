package api

import (
	"encoding/json"
	"net/http"
	"path"
	"slices"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func (h *Handler) handleOrgRefs(w http.ResponseWriter, req *http.Request) {
	require.NoError(h.t, req.ParseForm())
	ids := strings.Split(req.Form.Get("in"), ",")
	refs := make(map[string]*OrgRef, len(ids))
	for _, id := range ids {
		if o, ok := h.store.orgs[id]; ok {
			refs[id] = o.AsRef()
		} else {
			refs[id] = nil
		}
	}
	_ = json.NewEncoder(w).Encode(refs)
}

func (h *Handler) handleListOrgs(w http.ResponseWriter, _ *http.Request) {
	h.store.mux.RLock()
	defer h.store.mux.RUnlock()
	var (
		orgs     = make([]*Org, 0, len(h.store.orgs))
		ownerIDs = make(uniqueMap)
	)
	for _, o := range h.store.orgs {
		orgs = append(orgs, o)
		ownerIDs[o.Owner] = struct{}{}
	}
	slices.SortFunc(orgs, func(a, b *Org) int { return strings.Compare(a.ID, b.ID) })
	_ = json.NewEncoder(w).Encode(struct {
		Items []*Org   `json:"items"`
		Links HalLinks `json:"_links"`
	}{
		Items: orgs,
		Links: MakeHALLinks("ref:users:0=/ref/users?in=" + strings.Join(ownerIDs.keys(), ",")),
	})
}

func (h *Handler) handleGetOrg(w http.ResponseWriter, req *http.Request) {
	h.store.mux.RLock()
	defer h.store.mux.RUnlock()
	var org *Org

	// TODO why doesn't Chi decode this?
	orgID := chi.URLParam(req, "id")
	if strings.HasPrefix(orgID, "name%3D") {
		name := strings.TrimPrefix(orgID, "name%3D")
		for _, o := range h.store.orgs {
			if o.Name == name {
				org = o
				break
			}
		}
	} else {
		org = h.store.orgs[path.Base(req.URL.Path)]
	}

	if org == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_ = json.NewEncoder(w).Encode(org)
}
