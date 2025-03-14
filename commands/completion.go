package commands

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/platformsh/cli/internal/config"
)

func newCompletionCommand(cnf *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:           "completion",
		Short:         "Print the completion script for your shell",
		Args:          cobra.MaximumNArgs(1),
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			completionArgs := []string{"_completion", "-g", "--program", cnf.Application.Executable}
			if len(args) > 0 {
				completionArgs = append(completionArgs, "--shell-type", args[0])
			}
			var b bytes.Buffer
			c := makeLegacyCLIWrapper(cnf, &b, cmd.ErrOrStderr(), cmd.InOrStdin())

			if err := c.Exec(cmd.Context(), completionArgs...); err != nil {
				exitWithError(err)
			}

			pharPath := c.PharPath()

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
