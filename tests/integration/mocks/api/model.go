package api

// TODO unify these models with the 'api' package, and/or use OpenAPI or similar to generate them

import (
	"strings"
	"time"
)

type HALLink struct {
	HREF string `json:"href"`
}

type HalLinks map[string]HALLink

// MakeHALLinks helps make a list of HAL links out of arguments in the format name=path.
func MakeHALLinks(nameAndPath ...string) HalLinks {
	links := make(HalLinks, len(nameAndPath))
	for _, p := range nameAndPath {
		parts := strings.SplitN(p, "=", 2)
		links[parts[0]] = HALLink{parts[1]}
	}
	return links
}

type Subscription struct {
	ID            string   `json:"id"`
	Links         HalLinks `json:"_links"`
	ProjectID     string   `json:"project_id"`
	ProjectRegion string   `json:"project_region"`
	ProjectTitle  string   `json:"project_title"`
	Status        string   `json:"status"`
	ProjectUI     string   `json:"project_ui"`

	EventualProjectID string `json:"-"`
}

type ProjectRepository struct {
	URL string `json:"url"`
}

type Project struct {
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Region       string            `json:"region"`
	Organization string            `json:"organization"`
	Vendor       string            `json:"vendor"`
	Repository   ProjectRepository `json:"repository"`
	Links        HalLinks          `json:"_links"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`

	SubscriptionID string `json:"-"`
}

func (p *Project) AsRef() *ProjectRef {
	return &ProjectRef{
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

type Environment struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	MachineName string    `json:"machine_name"`
	Title       string    `json:"title"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	Parent      any       `json:"parent"`
	Project     string    `json:"project"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Links       HalLinks  `json:"_links"`
}

type Org struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Label string   `json:"label"`
	Owner string   `json:"owner_id"`
	Links HalLinks `json:"_links"`
}

func (o *Org) AsRef() *OrgRef {
	return &OrgRef{
		ID:    o.ID,
		Name:  o.Name,
		Label: o.Label,
		Owner: o.Owner,
	}
}

type OrgRef struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Label string `json:"label"`
	Owner string `json:"owner_id"`
}

type UserGrant struct {
	ResourceID     string    `json:"resource_id"`
	ResourceType   string    `json:"resource_type"`
	OrganizationID string    `json:"organization_id"`
	UserID         string    `json:"user_id"`
	Permissions    []string  `json:"permissions"`
	GrantedAt      time.Time `json:"granted_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type UserRef struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type ProjectRef struct {
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

type User struct {
	ID string `json:"id"`

	Deactivated bool `json:"deactivated"`

	Namespace string `json:"namespace"`

	Username      string `json:"username"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Picture       string `json:"picture"`
	Country       string `json:"country"`

	PhoneNumberVerified bool `json:"phone_number_verified"`

	MFAEnabled bool `json:"mfa_enabled"`
	SSOEnabled bool `json:"sso_enabled"`

	ConsentMethod string    `json:"consent_method"`
	ConsentedAt   time.Time `json:"consented_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
