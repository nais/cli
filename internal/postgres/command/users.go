package command

import (
	"context"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/cli/internal/postgres/users/add"
	"github.com/nais/cli/internal/postgres/users/list"
)

func usersCommand(parentFlags *flag.Postgres) *cli.Command {
	flags := &flag.User{Postgres: parentFlags}
	userAddFlags := &flag.UserAdd{
		User:      flags,
		Privilege: "select",
	}
	userListFlags := &flag.UserList{User: flags}
	return cli.NewCommand("users", "Manage users in your SQL instance.",
		cli.WithSubCommands(
			cli.NewCommand("add", "Add a user to a SQL instance.",
				cli.WithLong("Will grant a user access to tables in public schema."),
				cli.WithArgs("app_name", "username", "password"),
				cli.WithFlag("privilege", "", "The privilege to grant to the user.", &userAddFlags.Privilege),
				cli.WithValidate(cli.ValidateExactArgs(3)),
				cli.WithRun(func(ctx context.Context, output output.Output, args []string) error {
					return add.Run(
						ctx,
						add.Arguments{
							ApplicationName: args[0],
							Username:        args[1],
							Password:        args[2],
						},
						userAddFlags,
					)
				}),
			),
			cli.NewCommand("list", "List users in a SQL instance database.",
				cli.WithArgs("app_name"),
				cli.WithValidate(cli.ValidateExactArgs(1)),
				cli.WithRun(func(ctx context.Context, output output.Output, args []string) error {
					return list.Run(ctx, args[0], userListFlags)
				}),
			),
		),
	)
}
