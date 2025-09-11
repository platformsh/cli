package commands

import (
	"os"
	"path/filepath"

	"github.com/platformsh/cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/upsun/convsun/api"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

func innerConvertConfigCommand(cnf *config.Config) Command {
	noInteractionOption := NoInteractionOption(cnf)

	return Command{
		Name: CommandName{
			Namespace: "project",
			Command:   "convert",
		},
		Usage: []string{
			cnf.Application.Executable + " project:convert",
		},
		Aliases: []string{
			"convert",
		},
		Description: "Locally create your Upsun compatible configuration file based on your existing Platform.sh ones.",
		Help:        "",
		Examples: []Example{
			{
				Commandline: "",
				Description: "Convert the project configuration files in your current directory",
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
			)),
		},
		Hidden: false,
	}
}

func NewConvertConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project:convert ",
		Short:   "Locally create an upsun config",
		Aliases: []string{"convert"},
		//Args:  cobra.ExactArgs(1),
		RunE: runConvertConfig,
	}

	return cmd
}

func runConvertConfig(cmd *cobra.Command, args []string) error {
	rootProject := filepath.Join(".")
	upsunFolder := filepath.Join(rootProject, ".upsun")

	err := os.MkdirAll(upsunFolder, os.ModePerm)
	if err == nil {
		api.Convert(rootProject, upsunFolder)
	}

	return err
}
