package commands

import (
	"fmt"

	"github.com/platformsh/cli/internal/config"
)

func globalOptions(cnf *config.Config) []Option {
	return []Option{
		HelpOption,
		VerboseOption,
		VersionOption,
		YesOption,
		NoInteractionOption(cnf),
		AnsiOption,
		NoAnsiOption,
		NoOption,
		QuietOption,
	}
}

func NoInteractionOption(cnf *config.Config) Option {
	return Option{
		Name:            "--no-interaction",
		Shortcut:        "",
		AcceptValue:     false,
		IsValueRequired: false,
		IsMultiple:      false,
		Description: "Do not ask any interactive questions; accept default values. " +
			"Equivalent to using the environment variable: " +
			fmt.Sprintf("%sNO_INTERACTION=1", cnf.Application.EnvPrefix),
		Default: false,
		Hidden:  false,
	}
}

var (
	HelpOption = Option{
		Name:            "--help",
		Shortcut:        "-h",
		AcceptValue:     false,
		IsValueRequired: false,
		IsMultiple:      false,
		Description:     "Display this help message",
		Default:         false,
		Hidden:          false,
	}
	VerboseOption = Option{
		Name:            "--verbose",
		Shortcut:        "-v|vv|vvv",
		AcceptValue:     false,
		IsValueRequired: false,
		IsMultiple:      false,
		Description:     "Increase the verbosity of messages",
		Default:         false,
		Hidden:          false,
	}
	VersionOption = Option{
		Name:            "--version",
		Shortcut:        "-V",
		AcceptValue:     false,
		IsValueRequired: false,
		IsMultiple:      false,
		Description:     "Display this application version",
		Default:         false,
		Hidden:          false,
	}
	YesOption = Option{
		Name:            "--yes",
		Shortcut:        "-y",
		AcceptValue:     false,
		IsValueRequired: false,
		IsMultiple:      false,
		Description: "Answer \"yes\" to confirmation questions; " +
			"accept the default value for other questions; disable interaction",
		Default: false,
		Hidden:  false,
	}
	AnsiOption = Option{
		Name:            "--ansi",
		Shortcut:        "",
		AcceptValue:     false,
		IsValueRequired: false,
		IsMultiple:      false,
		Description:     "Force ANSI output",
		Default:         false,
		Hidden:          true,
	}
	NoAnsiOption = Option{
		Name:            "--no-ansi",
		Shortcut:        "",
		AcceptValue:     false,
		IsValueRequired: false,
		IsMultiple:      false,
		Description:     "Disable ANSI output",
		Default:         false,
		Hidden:          true,
	}
	NoOption = Option{
		Name:            "--no",
		Shortcut:        "-n",
		AcceptValue:     false,
		IsValueRequired: false,
		IsMultiple:      false,
		Description: "Answer \"no\" to confirmation questions; " +
			"accept the default value for other questions; disable interaction",
		Default: false,
		Hidden:  true,
	}
	QuietOption = Option{
		Name:            "--quiet",
		Shortcut:        "-q",
		AcceptValue:     false,
		IsValueRequired: false,
		IsMultiple:      false,
		Description:     "Do not output any message",
		Default:         false,
		Hidden:          true,
	}
)
