package main

import (
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/platformsh/cli/commands"
)

func initViper() {
	viper.SetEnvPrefix("platformsh_cli")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	log.SetOutput(color.Error)
}

func main() {
	cobra.OnInitialize(initViper)
	if err := commands.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
