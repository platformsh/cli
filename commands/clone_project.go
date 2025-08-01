package commands

import (
	"fmt"

	"github.com/platformsh/cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/upsun/clonsun/api"
	"github.com/upsun/lib-sun/entity"
	utils "github.com/upsun/lib-sun/utility"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

func innerCloneProjectCommand(cnf *config.Config) Command {
	noInteractionOption := NoInteractionOption(cnf)

	return Command{
		Name: CommandName{
			Namespace: "project",
			Command:   "clone",
		},
		Usage: []string{
			cnf.Application.Executable + " project:clone",
		},
		Aliases: []string{
			"convert",
		},
		Description: "Clone an existing Platform.sh project environment to a different region and/or to Upsun. Cloning will import and push your codebase, import & export project/environment configurations, import & export your sql databases, import & export your media files. If you selected to clone a project to Upsun the required config.yaml will be automatically generated and added to your project.",
		Help:        "",
		Examples: []Example{
			{
				Commandline: "",
				Description: "Clone an existing Platform.sh project environment to a different region and/or to Upsun. Cloning will import and push your codebase, import & export project/environment configurations, import & export your sql databases, import & export your media files. If you selected to clone a project to Upsun the required config.yaml will be automatically generated and added to your project.",
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

func NewCloneProjectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project:clone",
		Short:   "Clone the current project.",
		Aliases: []string{"clone"},
		//Args:  cobra.ExactArgs(1),
		RunE: runCloneProject,
	}

	return cmd
}

func runCloneProject(cmd *cobra.Command, args []string) error {
	//TODO: Dynamically from console arguments
	projectFrom := entity.MakeProjectContext(
		"platform",
		"zrkopp7gsvzwa",
		"master",
	)

	//TODO: Dynamically from console arguments
	projectTo := entity.MakeProjectContext(
		"upsun",
		"",
		"master",
	)
	projectTo.Region = "eu-3.platform.sh"
	projectTo.OrgEmail = "Mick"

	//TODO: Need on this command ?
	if !utils.IsAuthenticated(projectFrom) {
		fmt.Printf("You are not authenticated, please run: %v login\n", projectFrom.Provider)
	}
	if !utils.IsAuthenticated(projectTo) {
		fmt.Printf("You are not authenticated, please run: %v login\n", projectTo.Provider)
	}

	api.Clone(projectFrom, projectTo)

	return nil
}
