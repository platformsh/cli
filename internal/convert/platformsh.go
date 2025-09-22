package convert

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/viper"
	"github.com/symfony-cli/terminal"
	"github.com/upsun/lib-sun/detector"
	"github.com/upsun/lib-sun/entity"
	"github.com/upsun/lib-sun/readers"
	utils "github.com/upsun/lib-sun/utility"
	"github.com/upsun/lib-sun/writers"
)

// PlatformshToUpsun performs the conversion from Platform.sh config to Upsun config.
func PlatformshToUpsun(path string, stderr io.Writer) error {
	cwd, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("could not normalize project workspace path: %w", err)
	}

	upsunDir := filepath.Join(cwd, ".upsun")
	configPath := filepath.Join(upsunDir, "config.yaml")
	stat, err := os.Stat(configPath)
	if err == nil && !stat.IsDir() {
		fmt.Fprintln(stderr, "The file already exists:", color.YellowString(configPath))
		if !viper.GetBool("yes") {
			if viper.GetBool("no-interaction") {
				return fmt.Errorf("use the -y option to overwrite the file")
			}

			if !terminal.AskConfirmation("Do you want to overwrite it?", true) {
				return nil
			}
		}
	}

	log.Default().SetOutput(stderr)

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
	fmt.Fprintln(stderr, "Removing any `size`, `resources` or `disk` keys.")
	fmt.Fprintln(stderr,
		"Upsun disk sizes are set using Console or the "+color.GreenString("upsun resources:set")+" command.")
	readers.RemoveAllEntry(&metaConfig.Services, "size")
	readers.RemoveAllEntry(&metaConfig.Applications, "size")
	readers.RemoveAllEntry(&metaConfig.Services, "resources")
	readers.RemoveAllEntry(&metaConfig.Applications, "resources")
	readers.RemoveAllEntry(&metaConfig.Applications, "disk")
	readers.RemoveAllEntry(&metaConfig.Services, "disk")

	// Fix storage to match Upsun format
	fmt.Fprintln(stderr, "Replacing mount types (`local` becomes `instance`, and `shared` becomes `storage`).")
	readers.ReplaceAllEntry(&metaConfig.Applications, "local", "instance")
	readers.ReplaceAllEntry(&metaConfig.Applications, "shared", "storage")

	if err := os.MkdirAll(upsunDir, os.ModePerm); err != nil {
		return fmt.Errorf("could not create .upsun directory: %w", err)
	}

	fmt.Fprintln(stderr, "Creating combined configuration file.")
	writers.GenerateUpsunConfigFile(metaConfig, configPath)

	// Move extra config
	fmt.Fprintln(stderr, "Copying additional files if necessary.")
	utils.TransferConfigCustom(cwd, upsunDir)

	fmt.Fprintln(stderr, "Your configuration was successfully converted to the Upsun format.")
	fmt.Fprintln(stderr, "Check the generated files in:", color.GreenString(upsunDir))
	return nil
}
