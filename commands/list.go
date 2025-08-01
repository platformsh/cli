package commands

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/platformsh/cli/internal/config"
)

func newListCommand(cnf *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [flags] [namespace]",
		Short: "Lists commands",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			arguments := []string{"list", "--format=json"}
			if viper.GetBool("all") {
				arguments = append(arguments, "--all")
			}
			if len(args) > 0 {
				arguments = append(arguments, args[0])
			}

			var b bytes.Buffer
			c := makeLegacyCLIWrapper(cnf, &b, cmd.ErrOrStderr(), cmd.InOrStdin())

			if err := c.Exec(cmd.Context(), arguments...); err != nil {
				exitWithError(err)
			}

			var list List
			if err := json.Unmarshal(b.Bytes(), &list); err != nil {
				exitWithError(err)
			}

			// Override the application name and executable with our own config.
			list.Application.Name = cnf.Application.Name
			list.Application.Executable = cnf.Application.Executable

			projectInitCommand := innerProjectInitCommand(cnf)

			if !list.DescribesNamespace() || list.Namespace == projectInitCommand.Name.Namespace {
				list.AddCommand(&projectInitCommand)
			}

			appConfigValidateCommand := innerAppConfigValidateCommand(cnf)

			if !list.DescribesNamespace() || list.Namespace == appConfigValidateCommand.Name.Namespace {
				list.AddCommand(&appConfigValidateCommand)
			}

			// Add ConvSun to command list
			appConfigConvertCommand := innerConvertConfigCommand(cnf)

			if !list.DescribesNamespace() || list.Namespace == appConfigConvertCommand.Name.Namespace {
				list.AddCommand(&appConfigConvertCommand)
			}
			// End of ConvSun

			format := viper.GetString("format")
			raw := viper.GetBool("raw")

			var formatter Formatter
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
					exitWithError(err)
				}
				return
			}

			result, err := formatter.Format(&list, config.FromContext(cmd.Context()))
			if err != nil {
				exitWithError(err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), string(result))
		},
	}

	cmd.Flags().String("format", "txt", "The output format (txt, json, or md) [default: \"txt\"]")
	cmd.Flags().Bool("raw", false, "To output raw command list")
	cmd.Flags().Bool("all", false, "Show all commands, including hidden ones")

	viper.BindPFlags(cmd.Flags()) //nolint:errcheck
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		cmd.Root().Run(cmd.Root(), append([]string{"help", "list"}, args...))
	})

	return cmd
}
