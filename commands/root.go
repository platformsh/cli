package commands

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/platformsh/cli/internal"
	"github.com/platformsh/cli/internal/legacy"
)

func init() {
	RootCmd.PersistentFlags().BoolP("version", "V", false, "Displays the Platform.sh CLI version")
	RootCmd.PersistentFlags().String("phar-path", "", "Uses a local .phar file for the Legacy Platform.sh CLI")
	RootCmd.PersistentFlags().Bool("debug", false, "Enable debug logging")

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
	//nolint:errcheck
	viper.BindPFlags(RootCmd.PersistentFlags())
	viper.SetEnvPrefix("platformsh_cli")
	log.SetOutput(color.Error)
}

var (
	shellConfigSnippet = regexp.MustCompile("# BEGIN SNIPPET: Platform.sh CLI configuration(?s).+?# END SNIPPET")
	updateMessageChan  = make(chan *internal.ReleaseInfo, 1)
)

var RootCmd = &cobra.Command{
	Use:                "platform",
	Short:              "Platform.sh CLI",
	Args:               cobra.ArbitraryArgs,
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if viper.GetBool("version") {
			VersionCmd.Run(cmd, []string{})
			os.Exit(0)
		}
		go func() {
			rel, _ := internal.CheckForUpdate("platformsh/cli", version)
			updateMessageChan <- rel
		}()
	},
	Run: func(cmd *cobra.Command, args []string) {
		c := &legacy.CLIWrapper{
			Version:          version,
			CustomPshCliPath: viper.GetString("phar-path"),
			Debug:            viper.GetBool("debug"),
			Stdout:           cmd.OutOrStdout(),
			Stderr:           cmd.ErrOrStderr(),
			Stdin:            cmd.InOrStdin(),
		}
		if err := c.Init(); err != nil {
			debugLog("%s\n", color.RedString(err.Error()))
			os.Exit(1)
			return
		}

		if err := c.Exec(cmd.Context(), args...); err != nil {
			debugLog("%s\n", color.RedString(err.Error()))
			exitCode := 1
			var execErr *exec.ExitError
			if errors.As(err, &execErr) {
				exitCode = execErr.ExitCode()
			}
			//nolint:errcheck
			c.Cleanup()
			os.Exit(exitCode)
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		checkShellConfigLeftovers()
		select {
		case rel := <-updateMessageChan:
			printUpdateMessage(rel)
		default:
		}
	},
}

// checkShellConfigLeftovers checks .zshrc and .bashrc for any leftovers from the legacy CLI
func checkShellConfigLeftovers() {
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
			fmt.Fprintf(color.Error, "%s Your %s file contains code that is no longer needed for the New Platform.sh CLI\n",
				color.YellowString("Warning:"),
				shellConfigFile,
			)
			fmt.Fprintf(color.Error, "%s %s\n", color.YellowString("Please remove the following lines from:"), shellConfigFile)
			fmt.Fprintf(color.Error, "\t%s\n", strings.ReplaceAll(string(shellConfigSnippet.Find(shellConfig)), "\n", "\n\t"))
		}
	}
}

func printUpdateMessage(newRelease *internal.ReleaseInfo) {
	if newRelease == nil {
		return
	}

	fmt.Fprintf(color.Error, "\n\n%s %s → %s\n",
		color.YellowString("A new release of the Platform.sh CLI is available:"),
		color.CyanString(version),
		color.CyanString(newRelease.Version),
	)

	executable, err := os.Executable()
	if err == nil && isUnderHomebrew(executable) {
		fmt.Fprintf(color.Error, "To upgrade, run: %s\n", "brew upgrade platformsh/tap/platformsh-cli")
	}

	fmt.Fprintf(color.Error, "%s\n\n", color.YellowString(newRelease.URL))
}

func isUnderHomebrew(pshBinary string) bool {
	brewExe, err := exec.LookPath("brew")
	if err != nil {
		return false
	}

	brewPrefixBytes, err := exec.Command(brewExe, "--prefix").Output()
	if err != nil {
		return false
	}

	brewBinPrefix := filepath.Join(strings.TrimSpace(string(brewPrefixBytes)), "bin") + string(filepath.Separator)
	return strings.HasPrefix(pshBinary, brewBinPrefix)
}

func debugLog(format string, v ...any) {
	if !viper.GetBool("debug") {
		return
	}

	log.Printf(format, v...)
}
