package commands

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/fatih/color"
	"github.com/platformsh/cli/internal/legacy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	RootCmd.AddCommand(CompletionCmd)
}

var CompletionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Print the completion script for your shell",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		completionArgs := []string{"_completion", "-g", "--program", "platform"}
		if len(args) > 0 {
			completionArgs = append(completionArgs, "--shell-type", args[0])
		}
		var b bytes.Buffer
		c := &legacy.LegacyCLIWrapper{
			Version:          version,
			CustomPshCliPath: viper.GetString("phar-path"),
			Debug:            viper.GetBool("debug"),
			Stdout:           &b,
			Stderr:           cmd.ErrOrStderr(),
			Stdin:            cmd.InOrStdin(),
		}

		if err := c.Init(); err != nil {
			debugLog("%s\n", color.RedString(err.Error()))
			os.Exit(1)
			return
		}

		if err := c.Exec(cmd.Context(), completionArgs...); err != nil {
			debugLog("%s\n", color.RedString(err.Error()))
			exitCode := 1
			var execErr *exec.ExitError
			if errors.As(err, &execErr) {
				exitCode = execErr.ExitCode()
			}
			//nolint:errcheck
			c.Cleanup()
			os.Exit(exitCode)
			return
		}

		completions := strings.ReplaceAll(
			strings.ReplaceAll(
				b.String(),
				c.PSHPath(),
				"platform",
			),
			path.Base(c.PSHPath()),
			"platform",
		)
		fmt.Fprintln(cmd.OutOrStdout(), "#compdef platform")
		fmt.Fprintln(cmd.OutOrStdout(), completions)
	},
}
