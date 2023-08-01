package config

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
		HomebrewTap string `yaml:"homebrew_tap"` // e.g. "platformsh/tap/platformsh-cli"
		GitHubRepo  string `yaml:"github_repo"`  // e.g. "platformsh/cli"
	}

	Application struct {
		// Fields required for both the PHP and Go applications.
		Name       string `validate:"required"`                   // e.g. "Platform.sh CLI"
		EnvPrefix  string `validate:"required" yaml:"env_prefix"` // e.g. "PLATFORMSH_CLI_"
		Executable string `validate:"required"`                   // e.g. "platform"
		Slug       string `validate:"required,ascii"`             // e.g. "platformsh-cli"

		// Fields only needed by the PHP (legacy) CLI, at least for now.
		UserConfigDir string `validate:"required" yaml:"user_config_dir"` // e.g. ".platformsh"
	} `validate:"required,dive"`

	// Fields only needed by the PHP (legacy) CLI, at least for now.
	API struct {
		BaseURL string `validate:"required,url" yaml:"base_url"`  // e.g. "https://api.platform.sh"
		AuthURL string `validate:"omitempty,url" yaml:"auth_url"` // e.g. "https://auth.api.platform.sh"

		OAuth2AuthorizeURL string `validate:"required_without=AuthURL,omitempty,url" yaml:"oauth2_auth_url"`   // e.g. "https://auth.api.platform.sh/oauth2/authorize"
		OAuth2RevokeURL    string `validate:"required_without=AuthURL,omitempty,url" yaml:"oauth2_revoke_url"` // e.g. "https://auth.api.platform.sh/oauth2/revoke"
		OAuth2TokenURL     string `validate:"required_without=AuthURL,omitempty,url" yaml:"oauth2_token_url"`  // e.g. "https://auth.api.platform.sh/oauth2/token"
		CertifierURL       string `validate:"required_without=AuthURL,omitempty,url" yaml:"certifier_url"`     // e.g. "https://ssh.api.platform.sh"

		SSHDomainWildcards []string `validate:"required" yaml:"ssh_domain_wildcards"` // e.g. ["*.platform.sh"]
	} `validate:"required,dive"`
	Detection struct {
		GitRemoteName string   `validate:"required" yaml:"git_remote_name"` // e.g. "platform"
		SiteDomains   []string `validate:"required" yaml:"site_domains"`    // e.g. ["platformsh.site", "tst.site"]
	} `validate:"required,dive"`
	Service struct {
		Name             string `validate:"required"`                           // e.g. "Platform.sh"
		Slug             string `validate:"required,ascii"`                     // e.g. "platformsh"
		EnvPrefix        string `validate:"required" yaml:"env_prefix"`         // e.g. "PLATFORM_"
		ProjectConfigDir string `validate:"required" yaml:"project_config_dir"` // e.g. ".platform"
		ConsoleURL       string `validate:"omitempty,url" yaml:"console_url"`   // e.g. "https://console.platform.sh"
	} `validate:"required,dive"`
}
