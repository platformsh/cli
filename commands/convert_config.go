package commands

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/platformsh/cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/symfony-cli/terminal"
	"github.com/upsun/lib-sun/detector"
	"github.com/upsun/lib-sun/entity"
	"github.com/upsun/lib-sun/readers"
	utils "github.com/upsun/lib-sun/utility"
	"github.com/upsun/lib-sun/writers"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// innerConvertConfigCommand returns the Command struct for the convert config command.
func innerConvertConfigCommand(cnf *config.Config) Command {
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
			cnf.Application.Executable + " project:convert",
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

// newConvertConfigCommand creates the cobra command for converting config.
func newConvertConfigCommand(cnf *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project:convert",
		Short:   "Generate locally Upsun configuration from another provider",
		Aliases: []string{"convert"},
		RunE:    runConvertConfig,
	}

	cmd.Flags().StringP(
		"provider",
		"p",
		"platformsh",
		"The provider from which to convert the configuration. Currently, only 'platformsh' is supported.",
	)

	viper.BindPFlag("provider", cmd.Flags().Lookup("provider"))
	cmd.SetHelpFunc(func(_ *cobra.Command, _ []string) {
		internalCmd := innerConvertConfigCommand(cnf)
		fmt.Println(internalCmd.HelpPage(cnf))
	})
	return cmd
}

// runConvertConfig is the entry point for the convert config command.
func runConvertConfig(cmd *cobra.Command, args []string) error {
	if viper.GetString("provider") != "platformsh" {
		return fmt.Errorf("only the 'platformsh' provider is currently supported")
	}
	return runPlatformShConvert(cmd)
}

// runPlatformShConvert performs the conversion from Platform.sh config to Upsun config.
func runPlatformShConvert(cmd *cobra.Command) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get current working directory: %w", err)
	}

	cwd, err = filepath.Abs(filepath.Clean(cwd))
	if err != nil {
		return fmt.Errorf("could not normalize project workspace path: %w", err)
	}

	// Disable log for lib-sun
	log.Default().SetOutput(io.Discard)

	// Find config files
	configFiles, err := detector.FindConfig(cwd)
	if err != nil {
		return fmt.Errorf("could not detect configuration files: %w", err)
	}

	// Read PSH application config files
	var metaConfig entity.MetaConfig
	readers.ReadApplications(&metaConfig, configFiles[entity.PSH_APPLICATION], cwd)
	readers.ReadPlatforms(&metaConfig, configFiles[entity.PSH_PLATFORM], cwd)
	if metaConfig.Applications.IsZero() {
		return fmt.Errorf("no Platform.sh applications found")
	}

	// Read PSH services and routes config files
	readers.ReadServices(&metaConfig, configFiles[entity.PSH_SERVICE])
	readers.ReadRoutes(&metaConfig, configFiles[entity.PSH_ROUTE])

	// Remove size and resources entries
	readers.RemoveAllEntry(&metaConfig.Services, "size")
	readers.RemoveAllEntry(&metaConfig.Applications, "size")
	readers.RemoveAllEntry(&metaConfig.Services, "resources")
	readers.RemoveAllEntry(&metaConfig.Applications, "resources")

	// Fix storage to match Upsun format
	readers.ReplaceAllEntry(&metaConfig.Applications, "local", "instance")
	readers.ReplaceAllEntry(&metaConfig.Applications, "shared", "storage")
	readers.RemoveAllEntry(&metaConfig.Applications, "disk")

	upsunDir := filepath.Join(cwd, ".upsun")
	if err := os.MkdirAll(upsunDir, os.ModeDir); err != nil {
		return fmt.Errorf("could not create .upsun directory: %w", err)
	}

	configPath := filepath.Join(upsunDir, "config.yaml")
	stat, err := os.Stat(configPath)
	if err == nil && !stat.IsDir() {
		cmd.Printf("The file %v already exists.\n", configPath)
		if !viper.GetBool("yes") {
			if viper.GetBool("no-interaction") {
				return fmt.Errorf("use the -y option to overwrite the file")
			}

			if !terminal.AskConfirmation("Do you want to overwrite it?", true) {
				return nil
			}
		}
	}
	writers.GenerateUpsunConfigFile(metaConfig, configPath)

	// Move extra config
	utils.TransferConfigCustom(cwd, upsunDir)

	cmd.Println("Your configuration was successfully converted to the Upsun format.")
	cmd.Printf("Check the generated files in %v\n", upsunDir)
	return nil
}
