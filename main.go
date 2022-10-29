package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"lab.plat.farm/accounts/psh-go/legacy"
)

var version = ""

func main() {
	versionFlag := false
	flag.BoolVar(&versionFlag, "version", false, "")
	flag.BoolVar(&versionFlag, "v", false, "")

	helpFlag := false
	flag.BoolVar(&helpFlag, "help", false, "")
	flag.BoolVar(&helpFlag, "h", false, "")
	flag.Parse()

	if versionFlag {
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

	if helpFlag {
		if len(flag.Args()) == 0 {
			if err := c.Exec(context.TODO(), "list"); err != nil {
				c.Cleanup()
				log.Fatalf("Could not execute command: %s\n", err)
			}
			return
		}

		if err := c.Exec(context.TODO(), os.Args[1:]...); err != nil {
			c.Cleanup()
			log.Fatalf("Could not execute command: %s\n", err)
		}
		return
	}

	// Catch the completion flag to pass it correctly
	// This catches the command and passes it to
	if flag.Arg(0) == "completion" {
		args := []string{"_completion", "-g", "--program", "platform"}
		if flag.Arg(1) != "" {
			args = append(args, "--shell-type", flag.Arg(1))
		}
		var b bytes.Buffer
		c.Stdout = bufio.NewWriter(&b)

		if err := c.Exec(context.TODO(), args...); err != nil {
			c.Cleanup()
			log.Fatalf("Could not execute command: %s\n", err)
			return
		}
		completions := strings.ReplaceAll(
			strings.ReplaceAll(
				b.String(),
				c.PSHPath(),
				"platform",
			),
			path.Base(c.PSHPath()),
			"platform",
		)
		fmt.Fprintln(os.Stdout, completions)
		return
	}

	if err := c.Exec(context.TODO(), os.Args[1:]...); err != nil {
		c.Cleanup()
		log.Fatalf("Could not execute command: %s\n", err)
		return
	}
}
