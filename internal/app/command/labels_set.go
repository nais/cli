package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	labelspkg "github.com/nais/cli/internal/labels"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func labelsSet(parentFlags *flag.Labels) *naistrix.Command {
	flags := &flag.LabelsSet{Labels: parentFlags}
	return &naistrix.Command{
		Name:        "set",
		Title:       "Set labels for an application.",
		Description: "Sets one or more labels on an application. Use --label KEY=VALUE and repeat the flag for multiple labels.",
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		Flags: flags,
		ValidateFunc: func(_ context.Context, _ *naistrix.Arguments) error {
			if len(flags.LabelSet) == 0 {
				return fmt.Errorf("at least one --label KEY=VALUE must be specified")
			}
			return nil
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			updates, err := labelspkg.ParseAssignments(flags.LabelSet)
			if err != nil {
				return err
			}

			name := args.Get("name")
			environment, err := resolveAppEnvironment(ctx, flags.Team, name, string(flags.Environment))
			if err != nil {
				return err
			}

			ret, err := app.SetApplicationLabels(ctx, flags.Team, name, environment, updates)
			if err != nil {
				return err
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(ret)
			}
			out.Successf("Updated labels for %q in %q\n", name, environment)
			return out.Table().Render(ret)
		},
		AutoCompleteFunc: autoCompleteAppNames(parentFlags.App),
	}
}
