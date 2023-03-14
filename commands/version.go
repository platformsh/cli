package commands

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/platformsh/cli/internal/legacy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "0.0.0"
	commit  = "local"
	date    = ""
	builtBy = "local"
)

var VersionCmd = &cobra.Command{
	Use:                "version",
	Short:              "Print the version number of the Platform.sh CLI",
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	Run: func(cmd *cobra.Command, args []string) {
		if strings.Split(version, "-")[0] != strings.Split(legacy.PSHVersion, "-")[0] {
			fmt.Fprintf(
				color.Output,
				"Platform.sh CLI %s (Wrapped legacy CLI %s)\n",
				color.CyanString(version),
				color.CyanString(legacy.PSHVersion),
			)
		} else {
			fmt.Fprintf(color.Output, "Platform.sh CLI %s\n", color.CyanString(version))
		}

		if viper.GetBool("debug") {
			fmt.Fprintf(
				color.Output,
				"Embedded PHP version %s\n",
				color.CyanString(legacy.PHPVersion),
			)
			fmt.Fprintf(
				color.Output,
				"Commit %s (built %s by %s)\n",
				color.CyanString(commit),
				color.CyanString(date),
				color.CyanString(builtBy),
			)
		}
	},
}

func init() {
	RootCmd.AddCommand(VersionCmd)
}
