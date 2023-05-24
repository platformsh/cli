package commands

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/platformsh/platformify/commands"
	"github.com/spf13/cobra"
)

func init() {
	ProjectInitCmd.SetHelpFunc(projectInitHelpFn)
	RootCmd.AddCommand(ProjectInitCmd)
}

var ProjectInitCmd = &cobra.Command{
	Use:           "project:init",
	Short:         "Initialize the needed YAML files for your Platform.sh project",
	Aliases:       []string{"ify"},
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          commands.PlatformifyCmd.RunE,
}

func projectInitHelpFn(cmd *cobra.Command, args []string) {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, ' ', 0)

	c := ProjectInitCommand
	fmt.Fprintln(writer, color.YellowString("Command: ")+c.Name.String())
	fmt.Fprintln(writer, color.YellowString("Description: ")+c.Description.String())
	fmt.Fprintln(writer, "")
	if len(c.Usage) > 0 {
		fmt.Fprintln(writer, color.YellowString("Usage:"))
		for _, usage := range c.Usage {
			fmt.Fprintln(writer, " "+usage)
		}
		fmt.Fprintln(writer, "")
	}
	if c.Definition.Arguments != nil && c.Definition.Arguments.Len() > 0 {
		fmt.Fprintln(writer, color.YellowString("Arguments:"))
		for pair := c.Definition.Arguments.Oldest(); pair != nil; pair = pair.Next() {
			arg := pair.Value
			fmt.Fprintf(writer, "  %s\t%s\n", color.GreenString(arg.Name), arg.Description)
		}
		fmt.Fprintln(writer, "")
	}
	if c.Definition.Options != nil && c.Definition.Options.Len() > 0 {
		fmt.Fprintln(writer, color.YellowString("Options:"))
		for pair := c.Definition.Options.Oldest(); pair != nil; pair = pair.Next() {
			opt := pair.Value
			shortcut := opt.Shortcut
			if shortcut == "" {
				shortcut = "   "
			} else {
				shortcut += ","
			}
			fmt.Fprintf(writer, "  %s %s\t%s\n", color.GreenString(shortcut), color.GreenString(opt.Name), opt.Description)
		}
		fmt.Fprintln(writer, "")
	}
	if len(c.Examples) > 0 {
		fmt.Fprintln(writer, color.YellowString("Examples:"))
		for _, example := range c.Examples {
			fmt.Fprintln(writer, " "+example.Description.String()+":")
			fmt.Fprintln(writer, color.GreenString("   "+RootCmd.Name()+" "+c.Name.String()+" "+example.Commandline))
			fmt.Fprintln(writer, "")
		}
	}

	writer.Flush()

	fmt.Fprintln(cmd.OutOrStdout(), b.String())
}
