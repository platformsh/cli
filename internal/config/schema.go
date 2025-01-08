package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
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
		HomebrewTap string `yaml:"homebrew_tap"` // e.g. "platformsh/tap/platformsh-cli"
		GitHubRepo  string `yaml:"github_repo"`  // e.g. "platformsh/cli"
	}

	Application struct {
		// Fields required for both the PHP and Go applications.
		Name            string `validate:"required"`                           // e.g. "Platform.sh CLI"
		EnvPrefix       string `validate:"required" yaml:"env_prefix"`         // e.g. "PLATFORMSH_CLI_"
		Executable      string `validate:"required"`                           // e.g. "platform"
		Slug            string `validate:"required,ascii"`                     // e.g. "platformsh-cli"
		UserConfigDir   string `validate:"required" yaml:"user_config_dir"`    // e.g. ".platformsh"
		UserStateFile   string `validate:"omitempty" yaml:"user_state_file"`   // defaults to "state.json"
		WritableUserDir string `validate:"omitempty" yaml:"writable_user_dir"` // defaults to UserConfigDir
		TempSubDir      string `validate:"omitempty" yaml:"tmp_sub_dir"`       // defaults to Slug+"-tmp"
	} `validate:"required"`
	Updates struct {
		Check         bool `validate:"omitempty"`                       // defaults to true
		CheckInterval int  `validate:"omitempty" yaml:"check_interval"` // seconds, defaults to 3600
	} `validate:"omitempty"`

	// Fields only needed by the PHP (legacy) CLI, at least for now.
	API struct {
		BaseURL string `validate:"required,url" yaml:"base_url"`  // e.g. "https://api.platform.sh"
		AuthURL string `validate:"omitempty,url" yaml:"auth_url"` // e.g. "https://auth.api.platform.sh"

		OAuth2AuthorizeURL string `validate:"required_without=AuthURL,omitempty,url" yaml:"oauth2_auth_url"`   // e.g. "https://auth.api.platform.sh/oauth2/authorize"
		OAuth2RevokeURL    string `validate:"required_without=AuthURL,omitempty,url" yaml:"oauth2_revoke_url"` // e.g. "https://auth.api.platform.sh/oauth2/revoke"
		OAuth2TokenURL     string `validate:"required_without=AuthURL,omitempty,url" yaml:"oauth2_token_url"`  // e.g. "https://auth.api.platform.sh/oauth2/token"
		CertifierURL       string `validate:"required_without=AuthURL,omitempty,url" yaml:"certifier_url"`     // e.g. "https://ssh.api.platform.sh"
	} `validate:"required"`
	Detection struct {
		GitRemoteName string   `validate:"required" yaml:"git_remote_name"` // e.g. "platform"
		SiteDomains   []string `validate:"required" yaml:"site_domains"`    // e.g. ["platformsh.site", "tst.site"]
	} `validate:"required"`
	Service struct {
		Name                string `validate:"required"`                               // e.g. "Platform.sh"
		EnvPrefix           string `validate:"required" yaml:"env_prefix"`             // e.g. "PLATFORM_"
		ProjectConfigDir    string `validate:"required" yaml:"project_config_dir"`     // e.g. ".platform"
		ProjectConfigFlavor string `validate:"omitempty" yaml:"project_config_flavor"` // default: "platform"
		ConsoleURL          string `validate:"omitempty,url" yaml:"console_url"`       // e.g. "https://console.platform.sh"
		DocsURL             string `validate:"omitempty,url" yaml:"docs_url"`          // e.g. "https://docs.platform.sh"
	} `validate:"required"`
	SSH struct {
		DomainWildcards []string `validate:"required" yaml:"domain_wildcards"` // e.g. ["*.platform.sh"]
	} `validate:"required"`

	cacheDir        string `yaml:"-"`
	writableUserDir string `yaml:"-"`
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
}

// WritableUserDir returns the path to a writable user-level directory.
func (c *Config) WritableUserDir() (string, error) {
	if c.writableUserDir != "" {
		return c.writableUserDir, nil
	}
	hd, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(hd, c.Application.WritableUserDir)
	if err := mkDirIfNotExists(path); err != nil {
		return "", err
	}
	c.writableUserDir = path

	return path, nil
}

// CacheDir returns the path to a cache directory.
func (c *Config) CacheDir() (string, error) {
	if c.cacheDir != "" {
		return c.cacheDir, nil
	}
	ucd, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(ucd, c.Application.TempSubDir)
	if err := mkDirIfNotExists(path); err != nil {
		return "", err
	}
	c.cacheDir = path

	return path, nil
}

func mkDirIfNotExists(path string) error {
	err := os.Mkdir(path, 0o700)
	if errors.Is(err, fs.ErrExist) {
		return nil
	}
	return err
}
