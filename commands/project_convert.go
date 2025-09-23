package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	orderedmap "github.com/wk8/go-ordered-map/v2"

	"github.com/platformsh/cli/internal/config"
	"github.com/platformsh/cli/internal/convert"
)

// innerProjectConvertCommand returns the Command struct for the convert config command.
func innerProjectConvertCommand(cnf *config.Config) Command {
	noInteractionOption := NoInteractionOption(cnf)
	providerOption := Option{
		Name:            "--provider",
		Shortcut:        "-p",
		IsValueRequired: false,
		Default:         Any{"platformsh"},
		Description:     "The provider from which to convert the configuration. Currently, only 'platformsh' is supported.",
	}

	return Command{
		Name: CommandName{
			Namespace: "project",
			Command:   "convert",
		},
		Usage: []string{
			cnf.Application.Executable + " convert",
		},
		Aliases: []string{
			"convert",
		},
		Description: "Generate an Upsun compatible configuration based on the configuration from another provider.",
		Help:        "",
		Examples: []Example{
			{
				Commandline: "--provider=platformsh",
				Description: "Convert the Platform.sh project configuration files in your current directory",
			},
		},
		Definition: Definition{
			Arguments: &orderedmap.OrderedMap[string, Argument]{},
			Options: orderedmap.New[string, Option](orderedmap.WithInitialData[string, Option](
				orderedmap.Pair[string, Option]{
					Key:   HelpOption.GetName(),
					Value: HelpOption,
				},
				orderedmap.Pair[string, Option]{
					Key:   VerboseOption.GetName(),
					Value: VerboseOption,
				},
				orderedmap.Pair[string, Option]{
					Key:   VersionOption.GetName(),
					Value: VersionOption,
				},
				orderedmap.Pair[string, Option]{
					Key:   YesOption.GetName(),
					Value: YesOption,
				},
				orderedmap.Pair[string, Option]{
					Key:   noInteractionOption.GetName(),
					Value: noInteractionOption,
				},
				orderedmap.Pair[string, Option]{
					Key:   "provider",
					Value: providerOption,
				},
			)),
		},
		Hidden: false,
	}
}

// newProjectConvertCommand creates the cobra command for converting config.
func newProjectConvertCommand(cnf *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project:convert",
		Short:   "Generate locally Upsun configuration from another provider",
		Aliases: []string{"convert"},
		RunE:    runProjectConvert,
	}

	cmd.Flags().StringP(
		"provider",
		"p",
		"platformsh",
		"The provider from which to convert the configuration. Currently, only 'platformsh' is supported.",
	)

	_ = viper.BindPFlag("provider", cmd.Flags().Lookup("provider"))
	cmd.SetHelpFunc(func(_ *cobra.Command, _ []string) {
		internalCmd := innerProjectConvertCommand(cnf)
		fmt.Println(internalCmd.HelpPage(cnf))
	})
	return cmd
}

// runProjectConvert is the entry point for the convert config command.
func runProjectConvert(cmd *cobra.Command, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get current working directory: %w", err)
	}

	if viper.GetString("provider") == "platformsh" {
		return convert.PlatformshToUpsun(cwd, cmd.ErrOrStderr())
	}

	return fmt.Errorf("only the 'platformsh' provider is currently supported")
}
