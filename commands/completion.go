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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/platformsh/cli/internal/config"
	"github.com/platformsh/cli/internal/legacy"
)

func newCompletionCommand(cnf *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "completion",
		Short: "Print the completion script for your shell",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			completionArgs := []string{"_completion", "-g", "--program", cnf.Application.Executable}
			if len(args) > 0 {
				completionArgs = append(completionArgs, "--shell-type", args[0])
			}
			var b bytes.Buffer
			c := &legacy.CLIWrapper{
				Config:             cnf,
				Version:            version,
				CustomPharPath:     viper.GetString("phar-path"),
				Debug:              viper.GetBool("debug"),
				DisableInteraction: viper.GetBool("no-interaction"),
				Stdout:             &b,
				Stderr:             cmd.ErrOrStderr(),
				Stdin:              cmd.InOrStdin(),
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
				os.Exit(exitCode)
				return
			}

			completions := strings.ReplaceAll(
				strings.ReplaceAll(
					b.String(),
					c.PharPath(),
					cnf.Application.Executable,
				),
				path.Base(c.PharPath()),
				cnf.Application.Executable,
			)
			fmt.Fprintln(cmd.OutOrStdout(), "#compdef "+cnf.Application.Executable)
			fmt.Fprintln(cmd.OutOrStdout(), completions)
		},
	}
}
