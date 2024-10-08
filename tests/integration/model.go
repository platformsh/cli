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
	ID         string            `json:"id"`
	Repository projectRepository `json:"repository"`
	Links      halLinks          `json:"_links"`
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
	Links halLinks `json:"_links"`
}
