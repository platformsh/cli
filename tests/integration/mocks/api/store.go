package api

import (
	"sync"
)

type store struct {
	mux           sync.RWMutex
	orgs          map[string]*Org
	projects      map[string]*Project
	environments  map[string]*Environment
	subscriptions map[string]*Subscription
	userGrants    []*UserGrant
}

func (s *store) SetEnvironments(envs []*Environment) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.environments = make(map[string]*Environment, len(envs))
	for _, e := range envs {
		s.environments[e.ID] = e
	}
}

func (s *store) SetProjects(pros []*Project) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.projects = make(map[string]*Project, len(pros))
	for _, p := range pros {
		s.projects[p.ID] = p
	}
}

func (s *store) SetOrgs(orgs []*Org) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.orgs = make(map[string]*Org, len(orgs))
	for _, o := range orgs {
		s.orgs[o.ID] = o
	}
}

func (s *store) SetUserGrants(grants []*UserGrant) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.userGrants = grants
}
