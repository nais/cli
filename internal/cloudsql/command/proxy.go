package command

import (
	"context"

	"github.com/nais/cli/internal/cloudsql"
	"github.com/nais/cli/internal/cloudsql/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func proxyCommand(parentFlags *flag.CloudSQL) *naistrix.Command {
	flags := &flag.Proxy{
		CloudSQL: parentFlags,
		Port:     5432,
		Host:     "localhost",
	}
	return &naistrix.Command{
		Name:        "proxy",
		Title:       "Create a proxy to a SQL instance.",
		Description: "Allows your user to connect to databases and starts a proxy.",
		Args: []naistrix.Argument{
			{Name: "app_name"},
		},
		Flags:        flags,
		ValidateFunc: validation.RequireTeamAndEnvironment(flags),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return cloudsql.RunProxy(ctx, args.Get("app_name"), flags.Team, string(flags.Environment), flags, out)
		},
	}
}
