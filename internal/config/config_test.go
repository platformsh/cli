package config_test

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/platformsh/cli/internal/config"
)

//go:embed test-data/valid-config.yaml
var validConfig string

func TestFromYAML(t *testing.T) {
	cases := []struct {
		name               string
		config             string
		shouldContainError string
	}{
		{
			"missing_values",
			"application: {name: Test CLI}",
			`Error:Field validation for 'EnvPrefix' failed on the 'required' tag`,
		},
		{"complete", validConfig, ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := config.FromYAML([]byte(c.config))
			if c.shouldContainError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), c.shouldContainError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
