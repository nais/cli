package command

import (
	"context"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func labelsList(parentFlags *flag.Labels) *naistrix.Command {
	flags := &flag.LabelsList{Labels: parentFlags}
	return &naistrix.Command{
		Name:        "list",
		Title:       "List labels for an application.",
		Description: "Lists labels configured on an application in a specific environment.",
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			name := args.Get("name")
			environment, err := resolveAppEnvironment(ctx, out, flags.Team, name, string(flags.Environment), flags.Output == "json")
			if err != nil {
				return err
			}

			ret, err := app.GetApplicationLabels(ctx, flags.Team, name, environment)
			if err != nil {
				return err
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(ret)
			}
			if len(ret) == 0 {
				out.Println("No labels found.")
				return nil
			}
			return out.Table().Render(ret)
		},
		AutoCompleteFunc: autoCompleteAppNames(parentFlags.App),
	}
}
