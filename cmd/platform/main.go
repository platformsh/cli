package main

import (
	"os"

	"github.com/platformsh/cli/commands"
)

func main() {
	if err := commands.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
