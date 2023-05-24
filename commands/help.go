package commands

import (
	"context"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.SetHelpCommand(HelpCmd)
}

var HelpCmd = &cobra.Command{
	Use:                "help",
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.SetContext(context.Background())
		cmd, _, e := cmd.Root().Find(args)
		if cmd == nil || e != nil {
			cmd.Printf("Unknown help topic %#q\n", args)
			cobra.CheckErr(cmd.Root().Usage())
		} else {
			cmd.InitDefaultHelpFlag()
			cmd.InitDefaultVersionFlag()
			cmd.HelpFunc()(cmd, args)
		}
	},
}
