package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
)

func usersCommand(parentFlags *flag.Postgres) *cli.Command {
	flags := &flag.User{Postgres: parentFlags}
	return cli.NewCommand("users", "Manage users in your SQL instance.",
		cli.WithSubCommands(
			addCommand(flags),
			listCommand(flags),
		),
	)
}

func addCommand(parentFlags *flag.User) *cli.Command {
	userAddFlags := &flag.UserAdd{
		User:      parentFlags,
		Privilege: "select",
	}
	return cli.NewCommand("add", "Add a user to a SQL instance.",
		cli.WithLong("Will grant a user access to tables in public schema."),
		cli.WithArgs("app_name", "username", "password"),
		cli.WithFlag("privilege", "", "The privilege to grant to the user.", &userAddFlags.Privilege),
		cli.WithValidate(cli.ValidateExactArgs(3)),
		cli.WithRun(func(ctx context.Context, out output.Output, args []string) error {
			return postgres.AddUser(ctx, args[0], args[1], args[2], userAddFlags.Context, userAddFlags.Namespace, userAddFlags.Privilege, out)
		}),
	)
}

func listCommand(parentFlags *flag.User) *cli.Command {
	flags := &flag.UserList{User: parentFlags}
	return cli.NewCommand("list", "List users in a SQL instance database.",
		cli.WithArgs("app_name"),
		cli.WithValidate(cli.ValidateExactArgs(1)),
		cli.WithRun(func(ctx context.Context, out output.Output, args []string) error {
			return postgres.ListUsers(ctx, args[0], flags.Context, flags.Namespace, out)
		}),
	)
}
