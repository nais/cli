package command

import (
	"context"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/naistrix"
)

func list(parentFlags *flag.App) *naistrix.Command {
	flags := &flag.ListApps{
		App: parentFlags,
	}

	return &naistrix.Command{
		Name:  "list",
		Title: "List applications in a team.",
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			type application struct {
				Name            string `json:"name"`
				Environment     string `json:"environment"`
				State           string `json:"state"`
				Vulnerabilities int    `json:"vulnerabilities"`
				Issues          int    `heading:"Issues" json:"issues"`
			}
			ret, err := app.GetTeamApplications(ctx, flags.Team)
			if err != nil {
				return err
			}

			return out.Table().Render(ret)
		},
	}
}
