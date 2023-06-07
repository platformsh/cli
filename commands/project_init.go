package commands

import (
	"fmt"

	"github.com/platformsh/platformify/commands"
	"github.com/spf13/cobra"
)

func init() {
	ProjectInitCmd.SetHelpFunc(projectInitHelpFn)
	RootCmd.AddCommand(ProjectInitCmd)
}

var ProjectInitCmd = &cobra.Command{
	Use:           "project:init",
	Short:         "Creates Platform.sh-related starter YAML files for your project",
	Aliases:       []string{"ify"},
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          commands.PlatformifyCmd.RunE,
}

func projectInitHelpFn(cmd *cobra.Command, _ []string) {
	fmt.Fprintln(cmd.OutOrStdout(), ProjectInitCommand.HelpPage())
}
