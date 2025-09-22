package commands

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/fatih/color"
	"github.com/platformsh/platformify/commands"
	"github.com/platformsh/platformify/vendorization"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/platformsh/cli/internal"
	"github.com/platformsh/cli/internal/config"
	"github.com/platformsh/cli/internal/config/alt"
	"github.com/platformsh/cli/internal/legacy"
)

// Execute is the main entrypoint to run the CLI.
func Execute(cnf *config.Config) error {
	assets := &vendorization.VendorAssets{
		Use:          "project:init",
		Binary:       cnf.Application.Executable,
		ConfigFlavor: cnf.Service.ProjectConfigFlavor,
		EnvPrefix:    strings.TrimSuffix(cnf.Service.EnvPrefix, "_"),
		ServiceName:  cnf.Service.Name,
		DocsBaseURL:  cnf.Service.DocsURL,
	}

	ctx := vendorization.WithVendorAssets(config.ToContext(context.Background(), cnf), assets)
	return newRootCommand(cnf, assets).ExecuteContext(ctx)
}

func newRootCommand(cnf *config.Config, assets *vendorization.VendorAssets) *cobra.Command {
	var (
		updateMessageChan = make(chan *internal.ReleaseInfo, 1)
		versionCommand    = newVersionCommand(cnf)
	)
	cmd := &cobra.Command{
		Use:                cnf.Application.Executable,
		Short:              cnf.Application.Name,
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: false,
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		SilenceUsage:       true,
		SilenceErrors:      false,
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			if viper.GetBool("quiet") && !viper.GetBool("debug") && !viper.GetBool("verbose") {
				viper.Set("no-interaction", true)
				cmd.SetErr(io.Discard)
			} else {
				// Ensure the command's output writers can handle colors.
				if cmd.OutOrStdout() == os.Stdout {
					cmd.SetOut(color.Output)
				}
				if cmd.ErrOrStderr() == os.Stderr {
					cmd.SetErr(color.Error)
				}
			}
			if viper.GetBool("yes") {
				viper.Set("no-interaction", true)
			}
			if viper.GetBool("version") {
				versionCommand.Run(cmd, []string{})
				os.Exit(0)
			}
			if cnf.Wrapper.GitHubRepo != "" {
				go func() {
					rel, _ := internal.CheckForUpdate(cnf, config.Version)
					updateMessageChan <- rel
				}()
			}
			if alt.ShouldUpdate(cnf) {
				go func() {
					if err := alt.Update(cmd.Context(), cnf, debugLog); err != nil {
						cmd.PrintErrln("Error updating config:", color.RedString(err.Error()))
					}
				}()
			}
		},
		Run: func(cmd *cobra.Command, _ []string) {
			c := makeLegacyCLIWrapper(cnf, cmd.OutOrStdout(), cmd.ErrOrStderr(), cmd.InOrStdin())
			if err := c.Exec(cmd.Context(), os.Args[1:]...); err != nil {
				exitWithError(err)
			}
		},
		PersistentPostRun: func(cmd *cobra.Command, _ []string) {
			checkShellConfigLeftovers(cmd.ErrOrStderr(), cnf)
			select {
			case rel := <-updateMessageChan:
				printUpdateMessage(cmd.ErrOrStderr(), rel, cnf)
			default:
			}
		},
	}

	cmd.SetHelpFunc(func(innerCmd *cobra.Command, args []string) {
		if innerCmd.Use != cmd.Use {
			// For real (Cobra) commands, print the usage string.
			innerCmd.Print(innerCmd.UsageString())
			return
		}

		// Others will be passed to the legacy CLI's help command.
		if !slices.Contains(args, "--help") && !slices.Contains(args, "-h") {
			args = append([]string{"help"}, args...)
		}
		if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
			args = []string{"help"}
		}

		c := makeLegacyCLIWrapper(cnf, cmd.OutOrStdout(), cmd.ErrOrStderr(), cmd.InOrStdin())
		if err := c.Exec(cmd.Context(), args...); err != nil {
			exitWithError(err)
		}
	})

	cmd.PersistentFlags().BoolP("version", "V", false, fmt.Sprintf("Displays the %s version", cnf.Application.Name))
	cmd.PersistentFlags().Bool("debug", false, "Enable debug logging")
	cmd.PersistentFlags().Bool("no-interaction", false, "Enable non-interactive mode")
	cmd.PersistentFlags().BoolP("yes", "y", false, "Answer yes to all confirmation questions; implies --no-interaction")
	cmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().BoolP("quiet", "q", false,
		"Suppress any messages and errors (stderr), while continuing to display necessary output (stdout)."+
			" This implies --no-interaction. Ignored in verbose mode.",
	)

	validateCmd := commands.NewValidateCommand(assets)
	validateCmd.Use = "app:config-validate"
	validateCmd.Aliases = []string{"validate", "lint"}
	validateCmd.SetHelpFunc(func(_ *cobra.Command, _ []string) {
		internalCmd := innerAppConfigValidateCommand(cnf)
		fmt.Println(internalCmd.HelpPage(cnf))
	})

	// Add subcommands.
	cmd.AddCommand(
		newConfigInstallCommand(),
		newCompletionCommand(cnf),
		newHelpCommand(cnf),
		newInitCommand(cnf, assets),
		newListCommand(cnf),
		validateCmd,
		versionCommand,
	)
	if cnf.Service.ProjectConfigFlavor == "upsun" {
		cmd.AddCommand(newProjectConvertCommand(cnf))
	}

	//nolint:errcheck
	viper.BindPFlags(cmd.PersistentFlags())

	return cmd
}

// checkShellConfigLeftovers checks .zshrc and .bashrc for any leftovers from the legacy CLI
func checkShellConfigLeftovers(w io.Writer, cnf *config.Config) {
	start := fmt.Sprintf("# BEGIN SNIPPET: %s configuration", cnf.Application.Name)
	end := "# END SNIPPET"
	shellConfigSnippet := regexp.MustCompile(regexp.QuoteMeta(start) + "(?s).+?" + regexp.QuoteMeta(end))

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	shellConfigFiles := []string{
		filepath.Join(homeDir, ".zshrc"),
		filepath.Join(homeDir, ".bashrc"),
	}

	for _, shellConfigFile := range shellConfigFiles {
		if _, err := os.Stat(shellConfigFile); err != nil {
			continue
		}

		shellConfig, err := os.ReadFile(shellConfigFile)
		if err != nil {
			continue
		}

		if shellConfigSnippet.Match(shellConfig) {
			fmt.Fprintf(w, "%s Your %s file contains code that is no longer needed for the New %s\n",
				color.YellowString("Notice:"),
				shellConfigFile,
				cnf.Application.Name,
			)
			fmt.Fprintf(w, "%s %s\n", color.YellowString("Please remove the following lines from:"), shellConfigFile)
			fmt.Fprintf(w, "\t%s\n", strings.ReplaceAll(string(shellConfigSnippet.Find(shellConfig)), "\n", "\n\t"))
		}
	}
}

func printUpdateMessage(w io.Writer, newRelease *internal.ReleaseInfo, cnf *config.Config) {
	if newRelease == nil {
		return
	}

	fmt.Fprintf(w, "\n\n%s %s → %s\n",
		color.YellowString(fmt.Sprintf("A new release of the %s is available:", cnf.Application.Name)),
		color.CyanString(config.Version),
		color.CyanString(newRelease.Version),
	)

	executable, err := os.Executable()
	if err == nil && cnf.Wrapper.HomebrewTap != "" && isUnderHomebrew(executable) {
		fmt.Fprintf(
			w,
			"To upgrade, run: brew update && brew upgrade %s\n",
			color.YellowString(cnf.Wrapper.HomebrewTap),
		)
	} else if cnf.Wrapper.GitHubRepo != "" {
		fmt.Fprintf(
			w,
			"To upgrade, follow the instructions at: https://github.com/%s#upgrade\n",
			cnf.Wrapper.GitHubRepo,
		)
	}

	fmt.Fprintf(w, "%s\n\n", color.YellowString(newRelease.URL))
}

func isUnderHomebrew(binary string) bool {
	brewExe, err := exec.LookPath("brew")
	if err != nil {
		return false
	}

	brewPrefixBytes, err := exec.Command(brewExe, "--prefix").Output()
	if err != nil {
		return false
	}

	brewBinPrefix := filepath.Join(strings.TrimSpace(string(brewPrefixBytes)), "bin") + string(filepath.Separator)
	return strings.HasPrefix(binary, brewBinPrefix)
}

func debugLog(format string, v ...any) {
	if !viper.GetBool("debug") {
		return
	}

	prefix := color.New(color.ReverseVideo).Sprintf("DEBUG")
	fmt.Fprintf(color.Error, prefix+" "+strings.TrimSpace(format)+"\n", v...)
}

func exitWithError(err error) {
	var execErr *exec.ExitError
	if errors.As(err, &execErr) {
		exitCode := execErr.ExitCode()
		debugLog(err.Error())
		os.Exit(exitCode)
	}
	if !viper.GetBool("quiet") {
		fmt.Fprintln(color.Error, color.RedString(err.Error()))
	}
	os.Exit(1)
}

func makeLegacyCLIWrapper(cnf *config.Config, stdout, stderr io.Writer, stdin io.Reader) *legacy.CLIWrapper {
	return &legacy.CLIWrapper{
		Config:             cnf,
		Version:            config.Version,
		DebugLogFunc:       debugLog,
		DisableInteraction: viper.GetBool("no-interaction"),
		Stdout:             stdout,
		Stderr:             stderr,
		Stdin:              stdin,
	}
}
