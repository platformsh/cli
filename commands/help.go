package commands

import (
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

func init() {
	RootCmd.SetHelpFunc(HelpCmd.Run)
	RootCmd.AddCommand(HelpCmd)
}

var HelpCmd = &cobra.Command{
	Use:                "help",
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			args = []string{"list"}
		} else if !slices.Contains(args, "--help") && !slices.Contains(args, "-h") {
			args = append([]string{"help"}, args...)
		}
		RootCmd.Run(cmd, args)
	},
}
