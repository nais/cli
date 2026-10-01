package command

import (
	"context"

	"github.com/nais/cli/internal/cloudsql"
	"github.com/nais/cli/internal/cloudsql/command/flag"
	"github.com/nais/cli/internal/labels"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func listCommand(parentFlags *flag.CloudSQL) *naistrix.Command {
	flags := &flag.List{CloudSQL: parentFlags}

	return &naistrix.Command{
		Name:        "list",
		Title:       "List Cloud SQL instances for a team.",
		Description: "List Google Cloud SQL instances owned by a team.",
		Flags:       flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			labelFilters, err := labels.ParseFilters(flags.Labels)
			if err != nil {
				return err
			}

			environments := []string(nil)
			if flags.Environment != "" {
				environments = []string{string(flags.Environment)}
			}

			ret, err := cloudsql.GetTeamCloudSQLInstances(ctx, flags.Team, environments, labelFilters)
			if err != nil {
				return err
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(ret)
			}

			if len(ret) == 0 {
				out.Println("Team has no Cloud SQL instances.")
				return nil
			}

			return out.Table().Render(ret)
		},
	}
}
