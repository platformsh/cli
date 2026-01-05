package commands

import (
	"github.com/spf13/cobra"

	"github.com/upsun/cli/internal/config"
)

func newHelpCommand(_ *config.Config) *cobra.Command {
	return &cobra.Command{
		Use: "help",
		// Disable flag parsing so flags like --format are preserved for the legacy CLI.
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			foundCmd, _, e := cmd.Root().Find(args)
			if foundCmd == nil || e != nil || foundCmd == cmd.Root() {
				// Unknown command or root: delegate to root's HelpFunc for legacy CLI.
				cmd.Root().HelpFunc()(cmd.Root(), args)
			} else {
				// Known Go-native command: use its built-in help.
				foundCmd.InitDefaultHelpFlag()
				foundCmd.InitDefaultVersionFlag()
				foundCmd.HelpFunc()(foundCmd, args)
			}
		},
	}
}
