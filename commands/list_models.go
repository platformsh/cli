package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	orderedmap "github.com/wk8/go-ordered-map/v2"

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
			cnf.Application.Executable + " init",
		},
		Aliases: []string{
			"init", "ify",
		},
		Description: "Initialize a project",
		Help:        CleanString(initCommandHelp(cnf, false)),
		Examples: []Example{
			{
				Commandline: "",
				Description: "Create the starter YAML file(s) for your project",
			},
			{
				Commandline: "--ai=false",
				Description: "Disable AI mode",
			},
		},
		Definition: Definition{
			Arguments: &orderedmap.OrderedMap[string, Argument]{},
			Options: orderedmap.New[string, Option](orderedmap.WithInitialData[string, Option](
				orderedmap.Pair[string, Option]{
					Key: "ai",
					Value: Option{
						Name:        "--ai",
						AcceptValue: true,
						Description: "Use AI configuration",
						Default:     Any{any: false},
					},
				},
				orderedmap.Pair[string, Option]{
					Key: "digest",
					Value: Option{
						Name:        "--digest",
						AcceptValue: false,
						Description: "Only show the repository digest (the AI configuration input), without sending it",
						Default:     Any{any: false},
					},
				},
				orderedmap.Pair[string, Option]{
					Key:   HelpOption.GetName(),
					Value: HelpOption,
				},
				orderedmap.Pair[string, Option]{
					Key:   VerboseOption.GetName(),
					Value: VerboseOption,
				},
				orderedmap.Pair[string, Option]{
					Key:   VersionOption.GetName(),
					Value: VersionOption,
				},
				orderedmap.Pair[string, Option]{
					Key:   YesOption.GetName(),
					Value: YesOption,
				},
				orderedmap.Pair[string, Option]{
					Key:   noInteractionOption.GetName(),
					Value: noInteractionOption,
				},
			)),
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
			"lint",
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
			Arguments: &orderedmap.OrderedMap[string, Argument]{},
			Options: orderedmap.New[string, Option](orderedmap.WithInitialData[string, Option](
				orderedmap.Pair[string, Option]{
					Key:   HelpOption.GetName(),
					Value: HelpOption,
				},
				orderedmap.Pair[string, Option]{
					Key:   VerboseOption.GetName(),
					Value: VerboseOption,
				},
				orderedmap.Pair[string, Option]{
					Key:   VersionOption.GetName(),
					Value: VersionOption,
				},
				orderedmap.Pair[string, Option]{
					Key:   YesOption.GetName(),
					Value: YesOption,
				},
				orderedmap.Pair[string, Option]{
					Key:   noInteractionOption.GetName(),
					Value: noInteractionOption,
				},
			)),
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
	Description CleanString `json:"description"`
	Help        CleanString `json:"help"`
	Examples    []Example   `json:"examples"`
	Definition  Definition  `json:"definition"`
	Hidden      bool        `json:"hidden"`
}

// indentLines adds a number of spaces to each line of the given text.
func indentLines(text string, n int) string {
	indent := strings.Repeat(" ", n)
	return indent + strings.ReplaceAll(text, "\n", "\n"+indent)
}

func (c *Command) HelpPage(cnf *config.Config) string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, ' ', 0)

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
	if c.Help != "" {
		fmt.Fprintln(writer, color.YellowString("Help:"))
		fmt.Fprintln(writer, indentLines(c.Help.String(), 1))
		fmt.Fprintln(writer, "")
	}
	if len(c.Examples) > 0 {
		fmt.Fprintln(writer, color.YellowString("Examples:"))
		for _, example := range c.Examples {
			fmt.Fprintln(writer, indentLines(example.Description.String()+":", 1))
			usage := cnf.Application.Executable + " " + c.Name.String()
			if len(c.Usage) > 0 {
				usage = c.Usage[0]
			}
			fmt.Fprintln(writer,
				color.GreenString(fmt.Sprintf("   %s %s", usage, example.Commandline)))
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

type CleanString string

func (s CleanString) String() string {
	return string(s)
}

func (s *CleanString) MarshalJSON() ([]byte, error) {
	buff := new(bytes.Buffer)
	encoder := json.NewEncoder(buff)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(s.String())
	return buff.Bytes(), err
}

var (
	//nolint:lll
	regexColor = regexp.MustCompile(`<((?:fg=(?P<fg>\w+);?)?(?:bg=(?P<bg>\w+);?)?(?:options=(?P<options>\w+);?)?)?>(?P<label>.*?)</>`)
	regexTag   = regexp.MustCompile(`<.*?>`)
)

func (s *CleanString) UnmarshalJSON(text []byte) error {
	var str string
	err := json.Unmarshal(text, &str)
	if err != nil {
		return err
	}

	match := regexColor.FindStringSubmatch(str)
	if len(match) != 0 {
		res := make(map[string]string)
		for i, name := range regexColor.SubexpNames() {
			if i != 0 && name != "" {
				res[name] = match[i]
			}
		}

		atrs := make([]color.Attribute, 0, 2)
		switch res["fg"] {
		case "white":
			atrs = append(atrs, color.FgWhite)
		case "red":
			atrs = append(atrs, color.FgRed)
		case "yellow":
			atrs = append(atrs, color.FgYellow)
		}
		switch res["bg"] {
		case "white":
			atrs = append(atrs, color.BgWhite)
		case "red":
			atrs = append(atrs, color.BgRed)
		case "yellow":
			atrs = append(atrs, color.BgYellow)
		}
		colorStr := color.New(atrs...).SprintFunc()

		str = regexColor.ReplaceAllString(str, colorStr(res["label"]))
	}

	// Remove all remain tags like <comment></comment> and <info></info>
	str = regexTag.ReplaceAllString(str, "")

	*s = CleanString(str)
	return nil
}

type Example struct {
	Commandline string      `json:"commandline"`
	Description CleanString `json:"description"`
}

type Definition struct {
	Arguments *orderedmap.OrderedMap[string, Argument] `json:"arguments"`
	Options   *orderedmap.OrderedMap[string, Option]   `json:"options"`
}

type Argument struct {
	Name        string      `json:"name"`
	IsRequired  YesNo       `json:"is_required"`
	IsArray     YesNo       `json:"is_array"`
	Description CleanString `json:"description"`
	Default     Any         `json:"default"`
}

type Option struct {
	Name            string      `json:"name"`
	Shortcut        string      `json:"shortcut"`
	AcceptValue     YesNo       `json:"accept_value"`
	IsValueRequired YesNo       `json:"is_value_required"`
	IsMultiple      YesNo       `json:"is_multiple"`
	Description     CleanString `json:"description"`
	Default         Any         `json:"default"`
	Hidden          bool        `json:"hidden"`
}

func (o *Option) GetName() string {
	return strings.TrimPrefix(o.Name, "--")
}

type YesNo bool

func (y YesNo) String() string {
	if y {
		return "yes"
	}
	return "no"
}

type Any struct {
	any
}

func (a *Any) String() string {
	if a.any == nil {
		return "NULL"
	}
	switch t := a.any.(type) {
	case bool:
		return fmt.Sprintf("%t", a.any)
	case float32, float64:
		s := a.any.(float64) //nolint:errcheck
		if s == math.Trunc(s) {
			return fmt.Sprintf("%d", int64(s))
		}
		return fmt.Sprintf("%f", s)
	case string:
		s := a.any.(string) //nolint:errcheck
		return fmt.Sprintf("'%s'", s)
	case []any, []string, []int, []float64:
		return "array ()"
	default:
		panic(fmt.Sprintf("options: unsupported type: %T", t))
	}
}

func (a Any) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.any)
}

func (a *Any) UnmarshalJSON(text []byte) error {
	return json.Unmarshal(text, &a.any)
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
