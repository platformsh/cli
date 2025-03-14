package config

import (
	"context"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

// LoadYAML reads the configuration file from the environment if specified, falling back to the embedded file.
func LoadYAML() ([]byte, error) {
	if path := os.Getenv("CLI_CONFIG_FILE"); path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
		return b, nil
	}
	return embedded, nil
}

// FromYAML parses YAML configuration.
func FromYAML(b []byte) (*Config, error) {
	c := &Config{}
	c.applyDefaults()
	if err := yaml.Unmarshal(b, c); err != nil {
		return nil, fmt.Errorf("invalid config YAML: %w", err)
	}
	v := validator.New()
	if err := v.Struct(c); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	c.applyDynamicDefaults()
	c.raw = b
	return c, nil
}

type contextKey struct{}

// ToContext adds configuration to a context so it can be later fetched using FromContext.
func ToContext(ctx context.Context, cnf *Config) context.Context {
	return context.WithValue(ctx, contextKey{}, cnf)
}

// FromContext loads configuration that was set using ToContext, and panics if it is not set.
func FromContext(ctx context.Context) *Config {
	v, ok := ctx.Value(contextKey{}).(*Config)
	if !ok {
		panic("Config not set or wrong format")
	}
	return v
}
