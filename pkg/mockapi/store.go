package mockapi

import (
	"sync"
)

type store struct {
	sync.RWMutex
	myUser        *User
	orgs          map[string]*Org
	projects      map[string]*Project
	environments  map[string]*Environment
	subscriptions map[string]*Subscription
	userGrants    []*UserGrant

	canCreate map[string]*CanCreateResponse

	projectBackups map[string]map[string]*Backup
}

func (s *store) SetEnvironments(envs []*Environment) {
	s.Lock()
	defer s.Unlock()
	s.environments = make(map[string]*Environment, len(envs))
	for _, e := range envs {
		s.environments[e.ID] = e
	}
}

func (s *store) SetProjects(pros []*Project) {
	s.Lock()
	defer s.Unlock()
	s.projects = make(map[string]*Project, len(pros))
	for _, p := range pros {
		s.projects[p.ID] = p
	}
}

func (s *store) SetOrgs(orgs []*Org) {
	s.Lock()
	defer s.Unlock()
	s.orgs = make(map[string]*Org, len(orgs))
	for _, o := range orgs {
		s.orgs[o.ID] = o
	}
}

func (s *store) SetCanCreate(orgID string, r *CanCreateResponse) {
	s.Lock()
	defer s.Unlock()
	if s.canCreate == nil {
		s.canCreate = make(map[string]*CanCreateResponse)
	}
	s.canCreate[orgID] = r
}

func (s *store) SetUserGrants(grants []*UserGrant) {
	s.Lock()
	defer s.Unlock()
	s.userGrants = grants
}

func (s *store) SetMyUser(u *User) {
	s.myUser = u
}

func (s *store) SetProjectBackups(projectID string, backups []*Backup) {
	s.Lock()
	defer s.Unlock()
	if s.projectBackups == nil {
		s.projectBackups = make(map[string]map[string]*Backup)
	}
	if s.projectBackups[projectID] == nil {
		s.projectBackups[projectID] = make(map[string]*Backup)
	}
	for _, b := range backups {
		s.projectBackups[projectID][b.ID] = b
	}
}

func (s *store) addProjectBackup(projectID string, backup *Backup) {
	s.Lock()
	defer s.Unlock()
	if s.projectBackups == nil {
		s.projectBackups = make(map[string]map[string]*Backup)
	}
	if s.projectBackups[projectID] == nil {
		s.projectBackups[projectID] = make(map[string]*Backup)
	}
	s.projectBackups[projectID][backup.ID] = backup
}
