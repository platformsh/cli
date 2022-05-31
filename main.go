package main

import (
	"context"
	"fmt"
	"os"

	"lab.plat.farm/akalipetis/psh-go/legacy"
)

func main() {
	c := &legacy.LegacyCLIWrapper{}
	c.Init()
	defer c.Close()

	if err := c.Exec(context.TODO(), os.Args[1:]...); err != nil {
		fmt.Printf("Could not execute command: %s", err)
		fmt.Println()
		return
	}
}
