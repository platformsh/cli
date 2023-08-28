package commands

import (
	"fmt"

	"github.com/platformsh/platformify/commands"
	"github.com/spf13/cobra"

	"github.com/platformsh/cli/internal/config"
)

func newProjectInitCommand(cnf *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "project:init",
		Short:         fmt.Sprintf("Creates starter YAML files for your %s project", cnf.Service.Name),
		Aliases:       []string{"ify"},
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          commands.PlatformifyCmd.RunE,
	}
	cmd.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
		innerCmd := innerProjectInitCommand(cnf)
		fmt.Fprintln(cmd.OutOrStdout(), innerCmd.HelpPage(cnf))
	})
	return cmd
}
