package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"lab.plat.farm/accounts/psh-go/legacy"
)

var version = ""

func main() {
	versionFlag := flag.Bool("version", false, "")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Platform.sh CLI %s (Wrapped legacy CLI %s)\n",
			version,
			legacy.PSHVersion,
		)
		return
	}

	c := &legacy.LegacyCLIWrapper{}
	if err := c.Init(); err != nil {
		c.Cleanup()
		log.Fatalf("Could not initialize CLI: %s", err)
		return
	}

	if err := c.Exec(context.TODO(), os.Args[1:]...); err != nil {
		c.Cleanup()
		log.Fatalf("Could not execute command: %s\n", err)
		return
	}
}
