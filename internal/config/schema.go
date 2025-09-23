package config

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config provides YAML configuration for the CLI.
// This includes some translation strings for vendorization or white-label needs.
//
// It is able to parse some of the keys in the legacy CLI's config.yaml file.
// See: https://github.com/platformsh/legacy-cli/blob/main/config.yaml
//
//nolint:lll
type Config struct {
	// Fields only used by to the Go wrapper.
	Wrapper struct {
		HomebrewTap string `yaml:"homebrew_tap,omitempty"` // e.g. "platformsh/tap/platformsh-cli"
		GitHubRepo  string `yaml:"github_repo,omitempty"`  // e.g. "platformsh/cli"
	} `yaml:"wrapper,omitempty"`

	Application struct {
		// Fields required for both the PHP and Go applications.
		Name            string `validate:"required"`                                     // e.g. "Upsun CLI"
		EnvPrefix       string `validate:"required" yaml:"env_prefix"`                   // e.g. "UPSUN_CLI_"
		Executable      string `validate:"required"`                                     // e.g. "upsun"
		Slug            string `validate:"required,ascii"`                               // e.g. "upsun-cli"
		UserConfigDir   string `validate:"required" yaml:"user_config_dir"`              // e.g. ".upsun"
		UserStateFile   string `validate:"omitempty" yaml:"user_state_file,omitempty"`   // defaults to "state.json"
		WritableUserDir string `validate:"omitempty" yaml:"writable_user_dir,omitempty"` // defaults to UserConfigDir
		TempSubDir      string `validate:"omitempty" yaml:"tmp_sub_dir,omitempty"`       // defaults to Slug+"-tmp"
	} `validate:"required"`
	Updates struct {
		Check         bool `validate:"omitempty"`                                 // defaults to true
		CheckInterval int  `validate:"omitempty" yaml:"check_interval,omitempty"` // seconds, defaults to 3600
	} `validate:"omitempty"`

	// Fields only needed by the PHP (legacy) CLI, at least for now.
	API struct {
		BaseURL string `validate:"required,url" yaml:"base_url"`            // e.g. "https://api.upsun.com"
		AuthURL string `validate:"omitempty,url" yaml:"auth_url,omitempty"` // e.g. "https://auth.upsun.com"

		UserAgent string `validate:"omitempty" yaml:"user_agent,omitempty"`       // a template - see UserAgent method
		SessionID string `validate:"omitempty,ascii" yaml:"session_id,omitempty"` // the ID for the authentication session - defaults to "default"

		OAuth2ClientID     string `validate:"omitempty" yaml:"oauth2_client_id,omitempty"`                               // e.g. "upsun-cli"
		OAuth2AuthorizeURL string `validate:"required_without=AuthURL,omitempty,url" yaml:"oauth2_auth_url,omitempty"`   // e.g. "https://auth.upsun.com/oauth2/authorize"
		OAuth2RevokeURL    string `validate:"required_without=AuthURL,omitempty,url" yaml:"oauth2_revoke_url,omitempty"` // e.g. "https://auth.upsun.com/oauth2/revoke"
		OAuth2TokenURL     string `validate:"required_without=AuthURL,omitempty,url" yaml:"oauth2_token_url,omitempty"`  // e.g. "https://auth.upsun.com/oauth2/token"
		CertifierURL       string `validate:"required_without=AuthURL,omitempty,url" yaml:"certifier_url,omitempty"`     // No longer used

		AIServiceURL        string `validate:"omitempty,url" yaml:"ai_url,omitempty"`    // The AI service URL, e.g. "https://ai.upsun.com".
		EnableOrganizations bool   `validate:"omitempty" yaml:"organizations,omitempty"` // Whether the "organizations" feature is enabled.
	} `validate:"required"`
	Detection struct {
		GitRemoteName string   `validate:"required" yaml:"git_remote_name"` // e.g. "upsun"
		SiteDomains   []string `validate:"required" yaml:"site_domains"`    // e.g. ["upsunapp.com", "upsun.app"]
	} `validate:"required"`
	Service struct {
		Name                string `validate:"required"`                                         // e.g. "Upsun"
		EnvPrefix           string `validate:"required" yaml:"env_prefix"`                       // e.g. "PLATFORM_"
		ProjectConfigDir    string `validate:"required" yaml:"project_config_dir"`               // e.g. ".platform"
		ProjectConfigFlavor string `validate:"omitempty" yaml:"project_config_flavor,omitempty"` // default: "platform"
		ConsoleURL          string `validate:"omitempty,url" yaml:"console_url,omitempty"`       // e.g. "https://console.upsun.com"
		DocsURL             string `validate:"omitempty,url" yaml:"docs_url,omitempty"`          // e.g. "https://docs.upsun.com"
	} `validate:"required"`
	SSH struct {
		DomainWildcards []string `validate:"required" yaml:"domain_wildcards"` // e.g. ["*.platform.sh"]
	} `validate:"required"`

	Metadata   Metadata `validate:"omitempty" yaml:"metadata,omitempty"`
	SourceFile string   `yaml:"-"`

	raw             []byte `yaml:"-"`
	tempDir         string `yaml:"-"`
	writableUserDir string `yaml:"-"`
}

// Metadata defines information about the config itself.
type Metadata struct {
	Version      string    `validate:"omitempty,version" yaml:"version,omitempty"`
	UpdatedAt    time.Time `validate:"omitempty" yaml:"updated_at,omitempty"`
	DownloadedAt time.Time `validate:"omitempty" yaml:"downloaded_at,omitempty"`
	URL          string    `validate:"omitempty,url" yaml:"url,omitempty"`
}

// applyDefaults applies defaults to config before parsing.
func (c *Config) applyDefaults() {
	c.Application.UserStateFile = "state.json"
	c.Updates.Check = true
	c.Updates.CheckInterval = 3600
	c.Service.ProjectConfigFlavor = "platform"
}

// applyDynamicDefaults applies defaults to config after parsing and validating.
func (c *Config) applyDynamicDefaults() {
	if c.Application.TempSubDir == "" {
		c.Application.TempSubDir = c.Application.Slug + "-tmp"
	}
	if c.Application.WritableUserDir == "" {
		c.Application.WritableUserDir = c.Application.UserConfigDir
	}
	if c.SourceFile == "" {
		if path := os.Getenv("CLI_CONFIG_FILE"); path != "" {
			c.SourceFile = path
		}
	}
}

// Raw returns the config before it was unmarshalled, or a marshaled version if that is not available.
func (c *Config) Raw() ([]byte, error) {
	if len(c.raw) == 0 {
		b, err := yaml.Marshal(c)
		if err != nil {
			return nil, fmt.Errorf("could not load raw config: %w", err)
		}
		c.raw = b
	}
	return c.raw, nil
}

func (c *Config) UserAgent() string {
	template := c.API.UserAgent
	if template == "" {
		template = "{APP_NAME_DASH} {VERSION} ({OS} {ARCH})"
	}
	replacements := map[string]string{
		"{APP_NAME_DASH}": strings.ReplaceAll(c.Application.Name, " ", "-"),
		"{APP_NAME}":      c.Application.Name,
		"{APP_SLUG}":      c.Application.Slug,
		"{VERSION}":       Version,
		"{OS}":            runtime.GOOS,
		"{ARCH}":          runtime.GOARCH,
	}
	for key, value := range replacements {
		template = strings.ReplaceAll(template, key, value)
	}

	return template
}
