package integration

import (
	"strings"
	"time"
)

type halLink struct {
	HREF string `json:"href"`
}

type halLinks map[string]halLink

// makeHALLinks helps make a list of HAL links out of arguments in the format name=path.
func makeHALLinks(nameAndPath ...string) halLinks {
	links := make(halLinks, len(nameAndPath))
	for _, p := range nameAndPath {
		parts := strings.SplitN(p, "=", 2)
		links[parts[0]] = halLink{parts[1]}
	}
	return links
}

// TODO unify these models with the 'api' package, and/or use OpenAPI or similar to generate them
type subscription struct {
	ID            string   `json:"id"`
	Links         halLinks `json:"_links"`
	ProjectID     string   `json:"project_id"`
	ProjectRegion string   `json:"project_region"`
	ProjectTitle  string   `json:"project_title"`
	Status        string   `json:"status"`
	ProjectUI     string   `json:"project_ui"`

	eventualProjectID string
}

type projectRepository struct {
	URL string `json:"url"`
}

type project struct {
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Region       string            `json:"region"`
	Organization string            `json:"organization"`
	Vendor       string            `json:"vendor"`
	Repository   projectRepository `json:"repository"`
	Links        halLinks          `json:"_links"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`

	SubscriptionID string `json:"-"`
}

func (p *project) asRef() *projectRef {
	return &projectRef{
		ID:             p.ID,
		Region:         p.Region,
		Title:          p.Title,
		Status:         "active",
		OrganizationID: p.Organization,
		SubscriptionID: p.SubscriptionID,
		Vendor:         p.Vendor,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

type environment struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	MachineName string    `json:"machine_name"`
	Title       string    `json:"title"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	Parent      any       `json:"parent"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Links       halLinks  `json:"_links"`
}

type org struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Label string   `json:"label"`
	Owner string   `json:"owner_id"`
	Links halLinks `json:"_links"`
}

func (o *org) asRef() *orgRef {
	return &orgRef{
		ID:    o.ID,
		Name:  o.Name,
		Label: o.Label,
		Owner: o.Owner,
	}
}

type orgRef struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Label string `json:"label"`
	Owner string `json:"owner_id"`
}

type userGrant struct {
	ResourceID     string    `json:"resource_id"`
	ResourceType   string    `json:"resource_type"`
	OrganizationID string    `json:"organization_id"`
	UserID         string    `json:"user_id"`
	Permissions    []string  `json:"permissions"`
	GrantedAt      time.Time `json:"granted_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type projectRef struct {
	ID             string    `json:"id"`
	Region         string    `json:"region"`
	Title          string    `json:"title"`
	Status         string    `json:"status"`
	OrganizationID string    `json:"organization_id"`
	SubscriptionID string    `json:"subscription_id"`
	Vendor         string    `json:"vendor"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
