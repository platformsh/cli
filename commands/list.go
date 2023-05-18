package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/platformsh/cli/internal/legacy"
)

func init() {
	RootCmd.AddCommand(ListCmd)
}

func init() {
	ListCmd.Flags().String("format", "txt", "The output format (txt, json, or md) [default: \"txt\"]")
	ListCmd.Flags().Bool("raw", false, "To output raw command list")
	ListCmd.Flags().Bool("all", false, "Show all commands, including hidden ones")

	viper.BindPFlags(ListCmd.Flags()) //nolint:errcheck
}

var ListCmd = &cobra.Command{
	Use:   "list [flags] [namespace]",
	Short: "Lists commands",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var b bytes.Buffer
		c := &legacy.CLIWrapper{
			Version:          version,
			CustomPshCliPath: viper.GetString("phar-path"),
			Debug:            viper.GetBool("debug"),
			Stdout:           &b,
			Stderr:           cmd.ErrOrStderr(),
			Stdin:            cmd.InOrStdin(),
		}
		if err := c.Init(); err != nil {
			debugLog("%s\n", color.RedString(err.Error()))
			os.Exit(1)
			return
		}

		arguments := []string{"list", "--format=json"}
		if viper.GetBool("all") {
			arguments = append(arguments, "--all")
		}
		if len(args) > 0 {
			arguments = append(arguments, args[0])
		}
		if err := c.Exec(cmd.Context(), arguments...); err != nil {
			debugLog("%s\n", color.RedString(err.Error()))
			exitCode := 1
			var execErr *exec.ExitError
			if errors.As(err, &execErr) {
				exitCode = execErr.ExitCode()
			}
			//nolint:errcheck
			c.Cleanup()
			os.Exit(exitCode)
			return
		}

		var list List
		if err := json.Unmarshal(b.Bytes(), &list); err != nil {
			debugLog("%s\n", color.RedString(err.Error()))
			os.Exit(1)
			return
		}

		if !list.DescribesNamespace() || list.Namespace == ProjectInitCommand.Name.Namespace {
			list.AddCommand(&ProjectInitCommand)
		}

		format := viper.GetString("format")
		raw := viper.GetBool("raw")

		var formatter Formatter[*List]
		switch format {
		case "json":
			formatter = &JSONListFormatter{}
		case "md":
			formatter = &MDListFormatter{}
		case "txt":
			if raw {
				formatter = &RawListFormatter{}
			} else {
				formatter = &TXTListFormatter{}
			}
		default:
			c.Stdout = cmd.OutOrStdout()
			arguments := []string{"list", "--format=" + format}
			if err := c.Exec(cmd.Context(), arguments...); err != nil {
				debugLog("%s\n", color.RedString(err.Error()))
				exitCode := 1
				var execErr *exec.ExitError
				if errors.As(err, &execErr) {
					exitCode = execErr.ExitCode()
				}
				//nolint:errcheck
				c.Cleanup()
				os.Exit(exitCode)
			}
			return
		}

		result, err := formatter.Format(&list)
		if err != nil {
			debugLog("%s\n", color.RedString(err.Error()))
			os.Exit(1)
			return
		}

		fmt.Fprintln(cmd.OutOrStdout(), string(result))
	},
}
