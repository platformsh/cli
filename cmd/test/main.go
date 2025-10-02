package main

import (
	"os"

	"github.com/platformsh/cli/internal/tui"
)

func main() {
	spinr := tui.NewDefault(os.Stderr)
	spinr.Suffix = " Spinning..."
	spinr.Start()
	select {}
}
