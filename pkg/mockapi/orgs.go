package mockapi

import (
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

func (h *Handler) handleOrgRefs(w http.ResponseWriter, req *http.Request) {
	h.RLock()
	defer h.RUnlock()
	require.NoError(h.t, req.ParseForm())
	ids := strings.Split(req.Form.Get("in"), ",")
	refs := make(map[string]*OrgRef, len(ids))
	for _, id := range ids {
		if o, ok := h.orgs[id]; ok {
			refs[id] = o.AsRef()
		} else {
			refs[id] = nil
		}
	}
	_ = json.NewEncoder(w).Encode(refs)
}

func (h *Handler) handleListOrgs(w http.ResponseWriter, _ *http.Request) {
	h.RLock()
	defer h.RUnlock()
	var (
		orgs     = make([]*Org, 0, len(h.orgs))
		ownerIDs = make(uniqueMap)
	)
	for _, o := range h.orgs {
		orgs = append(orgs, o)
		ownerIDs[o.Owner] = struct{}{}
	}
	slices.SortFunc(orgs, func(a, b *Org) int { return strings.Compare(a.Name, b.Name) })
	_ = json.NewEncoder(w).Encode(struct {
		Items []*Org   `json:"items"`
		Links HalLinks `json:"_links"`
	}{
		Items: orgs,
		Links: MakeHALLinks("ref:users:0=/ref/users?in=" + strings.Join(ownerIDs.keys(), ",")),
	})
}

func (h *Handler) handleGetOrg(w http.ResponseWriter, req *http.Request) {
	h.RLock()
	defer h.RUnlock()
	var org *Org

	orgID := chi.URLParam(req, "organization_id")
	if strings.HasPrefix(orgID, "name%3D") {
		name := strings.TrimPrefix(orgID, "name%3D")
		for _, o := range h.orgs {
			if o.Name == name {
				org = o
				break
			}
		}
	} else {
		org = h.orgs[path.Base(req.URL.Path)]
	}

	if org == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_ = json.NewEncoder(w).Encode(org)
}

func (h *Handler) handleCreateOrg(w http.ResponseWriter, req *http.Request) {
	h.Lock()
	defer h.Unlock()
	var org Org
	err := json.NewDecoder(req.Body).Decode(&org)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	for _, o := range h.orgs {
		if o.Name == org.Name {
			w.WriteHeader(http.StatusConflict)
			return
		}
	}
	org.ID = ulid.MustNew(ulid.Now(), rand.Reader).String()
	org.Owner = h.myUser.ID
	org.Capabilities = []string{}
	org.Links = MakeHALLinks("self=/organizations/" + url.PathEscape(org.ID))
	h.orgs[org.ID] = &org
	_ = json.NewEncoder(w).Encode(&org)
}

func (h *Handler) handlePatchOrg(w http.ResponseWriter, req *http.Request) {
	h.Lock()
	defer h.Unlock()
	orgID := chi.URLParam(req, "organization_id")
	p, ok := h.orgs[orgID]
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
	h.orgs[orgID] = &patched
	_ = json.NewEncoder(w).Encode(&patched)
}
