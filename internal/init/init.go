package init

import "github.com/upsun/whatsun/pkg/digest"

// Input represents input for the /ai/generate-configuration API.
type Input struct {
	Digest       *digest.Digest
	ExtraContext string `json:"extra_context,omitempty"`

	OrganizationID string `json:"organization_id,omitempty"`
	ProjectID      string `json:"project_id,omitempty"`

	Debug bool `json:"debug,omitempty"`
}

// Output represents output from the /ai/generate-configuration API.
type Output struct {
	ConfigYAML string `json:"config_yaml"`
	Valid      bool   `json:"valid"`
}

var defaultIgnoredFiles = []string{"_www", ".platform", ".upsun"}

func DefaultDigestConfig() (*digest.Config, error) {
	cnf, err := digest.DefaultConfig()
	if err != nil {
		return nil, err
	}
	cnf.IgnoreFiles = defaultIgnoredFiles
	return cnf, nil
}
