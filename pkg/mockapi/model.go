package mockapi

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
}

type CanCreateRequiredAction struct {
	Action string `json:"action"`
	Type   string `json:"type"`
}

type CanCreateResponse struct {
	CanCreate bool   `json:"can_create"`
	Message   string `json:"message"`

	RequiredAction *CanCreateRequiredAction `json:"required_action"`
}

type ProjectRepository struct {
	URL string `json:"url"`
}

type Project struct {
	ID            string            `json:"id"`
	Title         string            `json:"title"`
	Region        string            `json:"region"`
	Organization  string            `json:"organization"`
	Vendor        string            `json:"vendor"`
	Repository    ProjectRepository `json:"repository"`
	DefaultBranch string            `json:"default_branch"`
	Links         HalLinks          `json:"_links"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`

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

	currentDeployment *Deployment
}

func (e *Environment) SetCurrentDeployment(d *Deployment) {
	e.currentDeployment = d
}

type Mount struct {
	Source     string `json:"source"`
	SourcePath string `json:"source_path"`
}

type App struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Size string `json:"size"`
	Disk int    `json:"disk"`

	Mounts map[string]Mount `json:"mounts"`
}

type Commands struct {
	Start string `json:"start"`
}

type WorkerInfo struct {
	Commands Commands `json:"commands"`
}

type Worker struct {
	App
	Worker WorkerInfo `json:"worker"`
}

type Deployment struct {
	WebApps  map[string]App    `json:"webapps"`
	Services map[string]App    `json:"services"`
	Workers  map[string]Worker `json:"workers"`

	Routes map[string]any `json:"routes"`

	Links HalLinks `json:"_links"`
}

type Org struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Label        string   `json:"label"`
	Owner        string   `json:"owner_id"`
	Capabilities []string `json:"capabilities"`
	Links        HalLinks `json:"_links"`
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

type ProjectUserGrant struct {
	ProjectID      string    `json:"project_id"`
	OrganizationID string    `json:"organization_id"`
	UserID         string    `json:"user_id"`
	Permissions    []string  `json:"permissions"`
	GrantedAt      time.Time `json:"granted_at"`
	UpdatedAt      time.Time `json:"updated_at"`
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
	ID        string `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
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

type Backup struct {
	ID            string    `json:"id"`
	EnvironmentID string    `json:"environment"`
	Status        string    `json:"status"`
	Safe          bool      `json:"safe"`
	Restorable    bool      `json:"restorable"`
	Automated     bool      `json:"automated"`
	CommitID      string    `json:"commit_id"`
	ExpiresAt     time.Time `json:"expires_at"`

	Links HalLinks `json:"_links"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Activity struct {
	ID   string `json:"id"`
	Type string `json:"type"`

	State             string    `json:"state"`
	Result            string    `json:"result"`
	CompletionPercent int       `json:"completion_percent"`
	CompletedAt       time.Time `json:"completed_at"`
	StartedAt         time.Time `json:"started_at"`

	Project      string   `json:"project"`
	Environments []string `json:"environments"`

	Description string `json:"description"`
	Text        string `json:"text"`

	Payload any `json:"payload"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Variable struct {
	Name           string `json:"name"`
	Value          string `json:"value,omitempty"`
	IsSensitive    bool   `json:"is_sensitive"`
	VisibleBuild   bool   `json:"visible_build"`
	VisibleRuntime bool   `json:"visible_runtime"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Links HalLinks `json:"_links"`
}

type EnvLevelVariable struct {
	Variable

	IsEnabled     bool `json:"is_enabled"`
	Inherited     bool `json:"inherited"`
	IsInheritable bool `json:"is_inheritable"`
}
