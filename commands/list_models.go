package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"

	"github.com/platformsh/cli/internal/config"
)

func innerProjectInitCommand(cnf *config.Config) Command {
	noInteractionOption := NoInteractionOption(cnf)

	return Command{
		Name: CommandName{
			Namespace: "project",
			Command:   "init",
		},
		Usage: []string{
			cnf.Application.Executable + " project:init",
		},
		Aliases: []string{
			"ify",
		},
		Description: "Initialize a project",
		Help:        "",
		Examples: []Example{
			{
				Commandline: "",
				Description: "Create the starter YAML files for your project",
			},
		},
		Definition: Definition{
			Options: map[string]Option{
				"help":           HelpOption,
				"verbose":        VerboseOption,
				"version":        VersionOption,
				"yes":            YesOption,
				"no-interaction": noInteractionOption,
			},
		},
		Hidden: false,
	}
}

func innerAppConfigValidateCommand(cnf *config.Config) Command {
	noInteractionOption := NoInteractionOption(cnf)

	return Command{
		Name: CommandName{
			Namespace: "app",
			Command:   "config-validate",
		},
		Usage: []string{
			cnf.Application.Executable + " app:config-validate",
		},
		Aliases: []string{
			"validate",
		},
		Description: "Validate the config files of a project",
		Help:        "",
		Examples: []Example{
			{
				Commandline: "",
				Description: "Validate the project configuration files in your current directory",
			},
		},
		Definition: Definition{
			Options: map[string]Option{
				"help":           HelpOption,
				"verbose":        VerboseOption,
				"version":        VersionOption,
				"yes":            YesOption,
				"no-interaction": noInteractionOption,
			},
		},
		Hidden: false,
	}
}

type List struct {
	Application Application `json:"application"`
	Commands    []*Command  `json:"commands"`
	Namespace   string      `json:"namespace,omitempty"`
	Namespaces  []Namespace `json:"namespaces,omitempty"`
}

type Application struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Executable string `json:"executable"`
}

type Command struct {
	Name        CommandName `json:"name"`
	Usage       []string    `json:"usage"`
	Aliases     []string    `json:"aliases"`
	Description string      `json:"description"`
	Help        string      `json:"help"`
	Examples    []Example   `json:"examples"`
	Definition  Definition  `json:"definition"`
	Hidden      bool        `json:"hidden"`
}

func (c *Command) HelpPage(cnf *config.Config) string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, ' ', 0)

	fmt.Fprintln(writer, color.YellowString("Command: ")+c.Name.String())
	fmt.Fprintln(writer, color.YellowString("Description: ")+c.Description)
	fmt.Fprintln(writer, "")
	if len(c.Usage) > 0 {
		fmt.Fprintln(writer, color.YellowString("Usage:"))
		for _, usage := range c.Usage {
			fmt.Fprintln(writer, " "+usage)
		}
		fmt.Fprintln(writer, "")
	}
	if len(c.Definition.Arguments) > 0 {
		fmt.Fprintln(writer, color.YellowString("Arguments:"))
		for _, arg := range c.Definition.Arguments {
			fmt.Fprintf(writer, "  %s\t%s\n", color.GreenString(arg.Name), arg.Description)
		}
		fmt.Fprintln(writer, "")
	}
	if len(c.Definition.Options) > 0 {
		fmt.Fprintln(writer, color.YellowString("Options:"))
		for _, opt := range c.Definition.Options {
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
	if c.Help != "" {
		fmt.Fprintln(writer, color.YellowString("Help:"))
		fmt.Fprintln(writer, " "+c.Help)
		fmt.Fprintln(writer, "")
	}
	if len(c.Examples) > 0 {
		fmt.Fprintln(writer, color.YellowString("Examples:"))
		for _, example := range c.Examples {
			fmt.Fprintln(writer, " "+example.Description+":")
			fmt.Fprintln(writer,
				color.GreenString(fmt.Sprintf("   %s %s %s", cnf.Application.Executable, c.Name.String(), example.Commandline)))
			fmt.Fprintln(writer, "")
		}
	}

	writer.Flush()

	return b.String()
}

type CommandName struct {
	Namespace string
	Command   string
}

func (n *CommandName) String() string {
	if n.Namespace == "" {
		return n.Command
	}
	return n.Namespace + ":" + n.Command
}

func (n *CommandName) ContainsNamespace() bool {
	return n.Namespace != ""
}

func (n *CommandName) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.String())
}

func (n *CommandName) UnmarshalJSON(text []byte) error {
	var command string
	err := json.Unmarshal(text, &command)
	if err != nil {
		return err
	}
	names := strings.SplitN(command, ":", 2)
	switch {
	case len(names) == 1:
		n.Command = names[0]
	case len(names) > 1:
		n.Namespace = names[0]
		n.Command = names[1]
	}
	return nil
}

type Example struct {
	Commandline string `json:"commandline"`
	Description string `json:"description"`
}

type Definition struct {
	Arguments map[string]Argument `json:"arguments"`
	Options   map[string]Option   `json:"options"`
}

type Argument struct {
	Name        string `json:"name"`
	IsRequired  bool   `json:"is_required"`
	IsArray     bool   `json:"is_array"`
	Description string `json:"description"`
	Default     any    `json:"default"`
}

type Option struct {
	Name            string `json:"name"`
	Shortcut        string `json:"shortcut"`
	AcceptValue     bool   `json:"accept_value"`
	IsValueRequired bool   `json:"is_value_required"`
	IsMultiple      bool   `json:"is_multiple"`
	Description     string `json:"description"`
	Default         any    `json:"default"`
	Hidden          bool   `json:"hidden"`
}

func (o *Option) GetName() string {
	return strings.TrimPrefix(o.Name, "--")
}

type Namespace struct {
	ID       string   `json:"id"`
	Commands []string `json:"commands"` // the same as Command.Name
}

func (l *List) DescribesNamespace() bool {
	return l.Namespace != ""
}

func (l *List) AddCommand(cmd *Command) {
	for i := range l.Namespaces {
		name := &l.Namespaces[i]
		if name.ID == cmd.Name.Namespace {
			name.Commands = append(name.Commands, cmd.Name.String())
			sort.Strings(name.Commands)
		}
	}

	l.Commands = append(l.Commands, cmd)
	sort.Slice(l.Commands, func(i, j int) bool {
		switch {
		case !l.Commands[i].Name.ContainsNamespace() && l.Commands[j].Name.ContainsNamespace():
			return true
		case l.Commands[i].Name.ContainsNamespace() && !l.Commands[j].Name.ContainsNamespace():
			return false
		default:
			return l.Commands[i].Name.String() < l.Commands[j].Name.String()
		}
	})
}
