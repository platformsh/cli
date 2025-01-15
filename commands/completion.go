package commands

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

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
				Debug:              viper.GetBool("debug"),
				DisableInteraction: viper.GetBool("no-interaction"),
				Stdout:             &b,
				Stderr:             cmd.ErrOrStderr(),
				Stdin:              cmd.InOrStdin(),
			}

			if err := c.Exec(cmd.Context(), completionArgs...); err != nil {
				handleLegacyError(err)
			}

			pharPath, err := c.PharPath()
			if err != nil {
				handleLegacyError(err)
			}

			completions := strings.ReplaceAll(
				strings.ReplaceAll(
					b.String(),
					pharPath,
					cnf.Application.Executable,
				),
				filepath.Base(pharPath),
				cnf.Application.Executable,
			)
			fmt.Fprintln(cmd.OutOrStdout(), "#compdef "+cnf.Application.Executable)
			fmt.Fprintln(cmd.OutOrStdout(), completions)
		},
	}
}
