package main

import (
	"context"
	"log"
	"os"

	"lab.plat.farm/accounts/psh-go/legacy"
)

func main() {
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
