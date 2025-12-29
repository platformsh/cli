package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/symfony-cli/terminal"

	"github.com/upsun/cli/commands"
	"github.com/upsun/cli/internal/config"
)

func main() {
	// Load configuration.
	cnfYAML, err := config.LoadYAML()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	cnf, err := config.FromYAML(cnfYAML)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// When Cobra starts, load Viper config from the environment.
	cobra.OnInitialize(func() {
		viper.SetEnvPrefix(strings.TrimSuffix(cnf.Application.EnvPrefix, "_"))
		viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
		viper.AutomaticEnv()

		if os.Getenv(cnf.Application.EnvPrefix+"NO_INTERACTION") == "1" {
			viper.Set("no-interaction", true)
		}
		if viper.GetBool("no-interaction") {
			terminal.Stdin.SetInteractive(false)
		}
	})

	if err := commands.Execute(cnf); err != nil {
		os.Exit(1)
	}
}
