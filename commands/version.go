package commands

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/upsun/cli/internal/config"
	"github.com/upsun/cli/internal/legacy"
)

func newVersionCommand(cnf *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:                "version",
		Short:              "Print the version number of the " + cnf.Application.Name,
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Fprintf(color.Output, "%s %s\n", cnf.Application.Name, color.CyanString(config.Version))

			if viper.GetBool("verbose") {
				fmt.Fprintf(
					color.Output,
					"Embedded PHP version %s\n",
					color.CyanString(legacy.PHPVersion),
				)
				fmt.Fprintf(
					color.Output,
					"Embedded Legacy CLI version %s\n",
					color.CyanString(legacy.LegacyCLIVersion),
				)
				fmt.Fprintf(
					color.Output,
					"Commit %s (built %s by %s)\n",
					color.CyanString(config.Commit),
					color.CyanString(config.Date),
					color.CyanString(config.BuiltBy),
				)
			}
		},
	}
}
