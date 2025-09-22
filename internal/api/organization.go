package api

import (
	"context"
	"time"
)

const (
	OrgTypeFlexible = "flexible"
	OrgTypeFixed    = "fixed"
)

type Organization struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	OwnerID      string    `json:"owner_id"`
	Namespace    string    `json:"namespace"`
	Name         string    `json:"name"`
	Label        string    `json:"label"`
	Country      string    `json:"country"`
	Vendor       string    `json:"vendor"`
	Capabilities []string  `json:"capabilities"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Links        HALLinks  `json:"_links"`
}

// GetOrganization gets a single organization by ID.
func (c *Client) GetOrganization(ctx context.Context, id string) (o *Organization, err error) {
	u, err := c.baseURLWithSegments("organizations", id)
	if err != nil {
		return
	}
	err = c.getResource(ctx, u.String(), &o)
	return
}
