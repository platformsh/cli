package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	flag "github.com/spf13/pflag"

	"strings"

	"github.com/fatih/color"

	"github.com/platformsh/cli/internal"
	"github.com/platformsh/cli/legacy"
)

var version = "0.0.0"

func main() {
	versionFlag := false
	flag.BoolVarP(&versionFlag, "version", "v", false, "")

	helpFlag := false
	flag.BoolVarP(&helpFlag, "help", "h", false, "")
	flag.CommandLine.ParseErrorsWhitelist.UnknownFlags = true
	flag.Parse()

	exitCode := 0
	defer func() {
		os.Exit(exitCode)
	}()

	// Run the update check in parallel
	updateMessageChan := make(chan *internal.ReleaseInfo)
	go func() {
		rel, _ := internal.CheckForUpdate("platformsh/homebrew-tap", version)
		updateMessageChan <- rel
	}()

	// Defer the check and do not wait for it if the command has finished first
	defer func() {
		select {
		case rel := <-updateMessageChan:
			printUpdateMessage(rel)
		default:
		}
	}()

	if versionFlag {
		fmt.Fprintf(color.Output, "Platform.sh CLI %s (Wrapped legacy CLI %s)\n",
			color.CyanString(version),
			color.GreenString(legacy.PSHVersion),
		)
		return
	}

	c := &legacy.LegacyCLIWrapper{}
	if err := c.Init(); err != nil {
		c.Cleanup()
		debugLog("Could not initialize CLI: %s", err)
		exitCode = 1
		return
	}

	var execErr *exec.ExitError
	if helpFlag {
		if len(flag.Args()) == 0 {
			if err := c.Exec(context.TODO(), "list"); err != nil {
				c.Cleanup()
				debugLog("Could not execute command: %s\n", err)
				exitCode = 1
				if errors.As(err, &execErr) {
					exitCode = execErr.ExitCode()
				}
			}
			return
		}

		if err := c.Exec(context.TODO(), os.Args[1:]...); err != nil {
			c.Cleanup()
			debugLog("Could not execute command: %s\n", err)
			exitCode = 1
			if errors.As(err, &execErr) {
				exitCode = execErr.ExitCode()
			}
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
			debugLog("Could not execute command: %s\n", err)
			exitCode = 1
			if errors.As(err, &execErr) {
				exitCode = execErr.ExitCode()
			}
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
		debugLog("Could not execute command: %s\n", err)
		exitCode = 1
		if errors.As(err, &execErr) {
			exitCode = execErr.ExitCode()
		}
		return
	}
}

func printUpdateMessage(newRelease *internal.ReleaseInfo) {
	if newRelease != nil {
		executable, _ := os.Executable()
		isHomebrew := isUnderHomebrew(executable)
		fmt.Fprintf(color.Error, "\n\n%s %s → %s\n",
			color.YellowString("A new release of the Platform.sh CLI is available:"),
			color.CyanString(version),
			color.CyanString(newRelease.Version),
		)
		if isHomebrew {
			fmt.Fprintf(os.Stderr, "To upgrade, run: %s\n", "brew upgrade platformsh/tap/platformsh-cli")
		}
		fmt.Fprintf(color.Error, "%s\n\n",
			color.YellowString(newRelease.URL))
	}
}

func isUnderHomebrew(pshBinary string) bool {
	brewExe, err := exec.LookPath("brew")
	if err != nil {
		return false
	}

	brewPrefixBytes, err := exec.Command(brewExe, "--prefix").Output()
	if err != nil {
		return false
	}

	brewBinPrefix := filepath.Join(strings.TrimSpace(string(brewPrefixBytes)), "bin") + string(filepath.Separator)
	return strings.HasPrefix(pshBinary, brewBinPrefix)
}

func debugLog(format string, v ...any) {
	if os.Getenv("PLATFORMSH_CLI_DEBUG") != "1" {
		return
	}

	log.Printf(format, v...)
}
