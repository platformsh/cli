package commands

import (
	"github.com/platformsh/platformify/commands"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(PlatformifyCmd)
}

var PlatformifyCmd = &cobra.Command{
	Use:           "project:init",
	Short:         "Initialize the needed YAML files for your Platform.sh project",
	Aliases:       []string{"ify"},
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          commands.PlatformifyCmd.RunE,
}
