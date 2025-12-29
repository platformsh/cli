package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"

	"github.com/upsun/cli/internal/config"
	"github.com/upsun/cli/internal/md"
)

type Formatter interface {
	Format(*List, *config.Config) ([]byte, error)
}

type JSONListFormatter struct{}

func (f *JSONListFormatter) Format(list *List, _ *config.Config) ([]byte, error) {
	buff := new(bytes.Buffer)
	encoder := json.NewEncoder(buff)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(list)
	return buff.Bytes(), err
}

type TXTListFormatter struct{}

func (f *TXTListFormatter) Format(list *List, cnf *config.Config) ([]byte, error) {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, ' ', 0)
	fmt.Fprintf(writer, "%s %s\n", list.Application.Name, color.GreenString(list.Application.Version))
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, color.YellowString("Global options:"))
	for _, opt := range globalOptions(cnf) {
		shortcut := opt.Shortcut
		if shortcut == "" {
			shortcut = "  "
		}
		fmt.Fprintf(writer, "  %s\t%s %s\n", color.GreenString(opt.Name), color.GreenString(shortcut), opt.Description)
	}
	fmt.Fprintln(writer)

	writer.Init(&b, 0, 8, 4, ' ', 0)
	if list.DescribesNamespace() {
		fmt.Fprintln(writer, color.YellowString("Available commands for the \"%s\" namespace:", list.Namespace))
	} else {
		fmt.Fprintln(writer, color.YellowString("Available commands:"))
	}

	cmds := make(map[string][]*Command)
	for _, cmd := range list.Commands {
		cmds[cmd.Name.Namespace] = append(cmds[cmd.Name.Namespace], cmd)
	}

	namespaces := make([]string, 0, len(cmds))
	for namespace := range cmds {
		namespaces = append(namespaces, namespace)
	}
	sort.Strings(namespaces)

	for _, namespace := range namespaces {
		if !list.DescribesNamespace() && namespace != "" {
			fmt.Fprintln(writer, color.YellowString("%s\t", namespace))
		}
		for _, cmd := range cmds[namespace] {
			name := color.GreenString(cmd.Name.String())
			if len(cmd.Aliases) > 0 {
				name = name + " (" + strings.Join(cmd.Aliases, ", ") + ")"
			}
			fmt.Fprintf(writer, "  %s\t%s\n", name, cmd.Description.String())
		}
	}
	writer.Flush()

	return b.Bytes(), nil
}

type RawListFormatter struct{}

func (f *RawListFormatter) Format(list *List, _ *config.Config) ([]byte, error) {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 16, ' ', 0)
	for _, cmd := range list.Commands {
		fmt.Fprintf(writer, "%s\t%s\n", cmd.Name.String(), cmd.Description.String())
	}
	writer.Flush()

	return b.Bytes(), nil
}

type MDListFormatter struct{}

func (f *MDListFormatter) Format(list *List, cnf *config.Config) ([]byte, error) {
	b := md.NewBuilder()
	b.H1(list.Application.Name + " " + list.Application.Version)

	cmds := make(map[string][]*Command)
	for _, cmd := range list.Commands {
		cmds[cmd.Name.Namespace] = append(cmds[cmd.Name.Namespace], cmd)
	}

	namespaces := make([]string, 0, len(cmds))
	for namespace := range cmds {
		namespaces = append(namespaces, namespace)
	}
	sort.Strings(namespaces)

	for _, namespace := range namespaces {
		if namespace != "" {
			b.Paragraph(md.Bold(namespace)).Ln()
		}
		for _, cmd := range cmds[namespace] {
			b.ListItem(md.Link(md.Code(cmd.Name.String()), md.Anchor(cmd.Name.String())))
		}
		b.Ln()
	}

	for _, cmd := range list.Commands {
		b.H2(md.Code(cmd.Name.String()))
		b.Paragraph(cmd.Description.String()).Ln()

		if len(cmd.Aliases) > 0 {
			aliases := make([]string, 0, len(cmd.Aliases))
			for _, alias := range cmd.Aliases {
				aliases = append(aliases, md.Code(alias))
			}
			b.Paragraph("Aliases: " + strings.Join(aliases, ", ")).Ln()
		}

		if len(cmd.Usage) > 0 {
			b.H3("Usage")
			for _, usage := range cmd.Usage {
				b.CodeBlock(usage)
			}
			b.Ln()
		}

		if cmd.Help != "" {
			b.Paragraph(cmd.Help.String()).Ln()
		}

		if cmd.Definition.Arguments != nil && cmd.Definition.Arguments.Len() > 0 {
			b.H4("Arguments")
			for pair := cmd.Definition.Arguments.Oldest(); pair != nil; pair = pair.Next() {
				arg := pair.Value
				line := md.Code(arg.Name)
				opts := make([]string, 0, 2)
				if arg.IsRequired {
					opts = append(opts, "required")
				} else {
					opts = append(opts, "optional")
				}
				if arg.IsArray {
					opts = append(opts, "multiple values allowed")
				}
				line += "(" + strings.Join(opts, "; ") + ")"

				b.ListItem(line)
				if arg.Description != "" {
					b.Paragraph("  " + arg.Description.String()).Ln()
				}
			}
		}

		if cmd.Definition.Options != nil && cmd.Definition.Options.Len() > 0 {
			b.H4("Options")
			for pair := cmd.Definition.Options.Oldest(); pair != nil; pair = pair.Next() {
				opt := pair.Value
				line := md.Code(opt.Name)
				if opt.Shortcut != "" {
					line += " (" + md.Code(opt.Shortcut) + ")"
				}
				if opt.AcceptValue {
					line += " (expects a value)"
				}
				b.ListItem(line)
				if opt.Description != "" {
					b.Paragraph("  " + opt.Description.String()).Ln()
				}
			}
		}

		if len(cmd.Examples) > 0 {
			b.H3("Examples")
			for _, example := range cmd.Examples {
				b.ListItem(example.Description.String() + ":")
				b.CodeBlock(cnf.Application.Executable + " " + cmd.Name.String() + " " + example.Commandline).Ln()
			}
		}
	}

	return []byte(b.String()), nil
}
