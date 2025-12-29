package commands

import (
	"github.com/spf13/cobra"

	"github.com/upsun/cli/internal/config"
)

func newHelpCommand(_ *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:                "help",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		Run: func(cmd *cobra.Command, args []string) {
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
}
