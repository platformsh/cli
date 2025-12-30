package commands

import (
	"bytes"
	"cmp"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/fatih/color"
	"github.com/platformsh/platformify/commands"
	"github.com/platformsh/platformify/vendorization"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/upsun/whatsun/pkg/digest"
	"github.com/upsun/whatsun/pkg/files"
	"gopkg.in/yaml.v3"

	"github.com/upsun/cli/internal/legacy"

	"github.com/upsun/cli/internal/api"
	"github.com/upsun/cli/internal/auth"
	"github.com/upsun/cli/internal/config"
	_init "github.com/upsun/cli/internal/init"
)

func newInitCommand(cnf *config.Config, assets *vendorization.VendorAssets) *cobra.Command {
	var (
		useAI       bool
		initOptions = &_init.Options{}
	)
	cmd := &cobra.Command{
		Use:     "init [flags]",
		Aliases: []string{"project:init", "ify"},
		Short:   "Generate configuration for a project",
		Long:    initCommandHelp(cnf, false),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInitCommand(cmd, args, useAI, initOptions, assets)
		},
	}

	cmd.Flags().BoolVar(&useAI, "ai", false, "Use AI configuration")
	cmd.Flags().StringVar(&initOptions.ExtraContext, "context", "",
		"Add extra context for AI configuration")
	cmd.Flags().BoolVar(&initOptions.OnlyShowDigest, "digest", false,
		"Only show the repository digest (the AI configuration input), without sending it")

	cmd.SetHelpFunc(func(_ *cobra.Command, _ []string) {
		internalCmd := innerProjectInitCommand(cnf)
		fmt.Println(internalCmd.HelpPage(cnf))
	})

	return cmd
}

func initCommandHelp(cnf *config.Config, short bool) string {
	var s strings.Builder

	bold := color.New(color.Bold)
	fmt.Fprintf(&s,
		"This command will generate a %s that you can build on.\n", bold.Sprint("starter configuration"))
	fmt.Fprintln(&s, "You can use AI, or follow a step-by-step setup guide.")

	fmt.Fprintln(&s, "\nUsing AI will send a sanitized repository digest to OpenAI for automated analysis.")

	if !short {
		fmt.Fprintln(&s, "You can review the digest at any time by running:",
			color.GreenString(cnf.Application.Executable+" init --digest"))
	}

	return s.String()
}

// runInitCommand handles the main logic for the init command.
func runInitCommand(
	cmd *cobra.Command, args []string, useAI bool, initOptions *_init.Options, assets *vendorization.VendorAssets,
) error {
	var path = "."
	if len(args) > 0 {
		path = args[0]
	}
	gitRoot, err := findGitRoot(path)
	if err != nil {
		return err
	}
	if gitRoot == "" {
		// TODO suggest creating a repository
		return fmt.Errorf("this command can only be run inside a Git repository")
	}

	if initOptions.OnlyShowDigest {
		dg, err := generateDigest(cmd.Context(), gitRoot)
		if err != nil {
			return err
		}
		return yaml.NewEncoder(cmd.OutOrStdout()).Encode(dg)
	}

	runNonAIConfig := func() error {
		return commands.Platformify(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), assets)
	}

	cnf := config.FromContext(cmd.Context())

	// TODO check if this is needed
	cnf.API.AIServiceURL = cmp.Or(os.Getenv(cnf.Application.EnvPrefix+"API_AI_URL"), cnf.API.AIServiceURL)

	legacyCLIClient, err := auth.NewLegacyCLIClient(cmd.Context(),
		makeLegacyCLIWrapper(cnf, cmd.OutOrStdout(), cmd.ErrOrStderr(), cmd.InOrStdin()))
	if err != nil {
		return err
	}

	msg, canUse := canUseAI(cnf)
	if !canUse {
		if useAI {
			return fmt.Errorf("cannot use AI: %s", msg)
		}
		return runNonAIConfig()
	}

	var stderr = cmd.ErrOrStderr()

	fmt.Fprintln(stderr, color.CyanString("Creating %s configuration", cnf.Service.Name))
	fmt.Fprintln(stderr)

	var isInteractive = !viper.GetBool("no-interaction")

	debugLogf("Checking selected organization")
	org, err := handleOrganizations(cmd.Context(), cnf, legacyCLIClient, initOptions)
	if err != nil {
		return err
	}
	if org != nil {
		debugLogf("Selected organization: %s (%s)", org.Label, org.ID)
		if initOptions.ProjectID != "" {
			debugLogf("Selected project: %s", initOptions.ProjectID)
		}
	}

	// Ask the user if they want to configure their project using AI (if not explicitly set).
	if !cmd.Flags().Changed("ai") && (org == nil || org.Type == api.OrgTypeFlexible) {
		if !isInteractive {
			return fmt.Errorf("specifying --ai=false or --ai=true is required in non-interactive mode")
		}
		fmt.Fprintln(stderr, initCommandHelp(cnf, true))

		res, err := choose(stderr, "How would you like to configure your project?", []string{
			"With AI (automatic)",
			"Without AI (guided)",
		})
		if err != nil {
			return err
		}
		useAI = res == "With AI (automatic)"
	}

	if !useAI {
		return runNonAIConfig()
	}

	if org != nil && org.Type != api.OrgTypeFlexible {
		fmt.Fprintf(stderr, "Selected organization: %s (%s)\n", color.YellowString(org.Label), color.YellowString(org.ID))
		fmt.Fprintln(stderr, color.YellowString(
			"Note: AI configuration is only compatible with `%s` organizations\n", api.OrgTypeFlexible))
	}

	if err := legacyCLIClient.EnsureAuthenticated(cmd.Context()); err != nil {
		return err
	}

	debugLogf("Analyzing repository")
	dg, err := generateDigest(cmd.Context(), gitRoot)
	if err != nil {
		return err
	}

	initOptions.HTTPClient = legacyCLIClient.HTTPClient
	initOptions.AIServiceURL = cnf.API.AIServiceURL
	initOptions.UserAgent = cnf.UserAgent()
	initOptions.IsInteractive = isInteractive
	initOptions.Yes = viper.GetBool("yes")
	initOptions.IsDebug = viper.GetBool("debug")
	initOptions.DebugLogFunc = debugLogf

	return _init.RunAIConfig(cmd.Context(), cnf, dg, gitRoot, initOptions, cmd.OutOrStdout(), cmd.ErrOrStderr())
}

// handleOrganizations manages organization selection and validation.
// It modifies initOptions.OrganizationID and initOptions.ProjectID.
func handleOrganizations(
	ctx context.Context, cnf *config.Config, legacyCLIClient *auth.LegacyCLIClient, initOptions *_init.Options,
) (*api.Organization, error) {
	if !cnf.API.EnableOrganizations {
		return nil, nil
	}

	apiClient, err := api.NewClient(cnf.API.BaseURL, legacyCLIClient.HTTPClient)
	if err != nil {
		return nil, err
	}

	silentLegacyWrapper := makeLegacyCLIWrapper(cnf, nil, nil, nil)
	currentOrgID, currentProjectID := getCurrentOrganizationAndProjectID(ctx, silentLegacyWrapper)

	if currentOrgID == "" {
		return nil, nil
	}

	if err := legacyCLIClient.EnsureAuthenticated(ctx); err != nil {
		return nil, err
	}

	org, err := apiClient.GetOrganization(ctx, currentOrgID)
	if err != nil {
		return nil, err
	}

	initOptions.OrganizationID = currentOrgID
	initOptions.ProjectID = currentProjectID
	return org, nil
}

// canUseAI validates that AI can be used with the current configuration.
func canUseAI(cnf *config.Config) (msg string, canUseAI bool) {
	if !cnf.API.EnableOrganizations {
		return "using AI requires Organizations to be enabled", false
	}
	if cnf.API.AIServiceURL == "" {
		return "using AI requires the service URL to be set", false
	}
	return "", true
}

// findGitRoot finds the closest parent directory containing a .git folder.
// Returns an absolute path to the git root, or empty string if not in a git repo.
func findGitRoot(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}

	for {
		gitPath := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached the root directory.
			return "", nil
		}
		dir = parent
	}
}

// generateDigest creates a digest of the project files.
func generateDigest(ctx context.Context, path string) (*digest.Digest, error) {
	fsys, err := files.LocalFS(path)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	digestConfig, err := _init.DefaultDigestConfig()
	if err != nil {
		return nil, err
	}

	digester, err := digest.NewDigester(fsys, digestConfig)
	if err != nil {
		return nil, err
	}

	return digester.GetDigest(ctx)
}

func getCurrentOrganizationAndProjectID(ctx context.Context, wrapper *legacy.CLIWrapper) (orgID, projectID string) {
	var buf = &bytes.Buffer{}
	wrapper.Stderr = nil
	wrapper.Stdout = buf
	wrapper.Stdin = nil

	if err := wrapper.Exec(ctx, "project:info", "id", "--no-interaction"); err != nil {
		return
	}
	projectID = strings.TrimSpace(buf.String())
	buf.Reset()

	if err := wrapper.Exec(ctx, "organization:info", "id", "--no-interaction"); err != nil {
		return
	}
	orgID = strings.TrimSpace(buf.String())

	return
}

// choose asks the user to select between options.
// TODO refactor this to a shared internal package
func choose(stderr io.Writer, message string, options []string) (result string, err error) {
	var renderer survey.Renderer
	renderer.WithStdio(terminal.Stdio{Err: stderr})
	prompt := &survey.Select{
		Renderer: renderer,
		Message:  message,
		Options:  options,
		Default:  options[0],
	}
	err = survey.AskOne(prompt, &result, survey.WithValidator(survey.Required))
	return
}
