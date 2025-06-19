package command

import (
	"context"

	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/cli/pkg/cli"
)

func usersCommand(parentFlags *flag.Postgres) *cli.Command {
	flags := &flag.User{Postgres: parentFlags}
	return &cli.Command{
		Name:        "users",
		Title:       "Manage users in your SQL instance.",
		StickyFlags: flags,
		SubCommands: []*cli.Command{
			addCommand(flags),
			listCommand(flags),
		},
	}
}

func addCommand(parentFlags *flag.User) *cli.Command {
	userAddFlags := &flag.UserAdd{
		User:      parentFlags,
		Privilege: "select",
	}
	return &cli.Command{
		Name:        "add",
		Title:       "Add a user to a SQL instance.",
		Description: "Will grant a user access to tables in public schema.",
		Args: []cli.Argument{
			{Name: "app_name", Required: true},
			{Name: "username", Required: true},
			{Name: "password", Required: true},
		},
		ValidateFunc: cli.ValidateExactArgs(3),
		Flags:        userAddFlags,
		RunFunc: func(ctx context.Context, out cli.Output, args []string) error {
			return postgres.AddUser(ctx, args[0], args[1], args[2], userAddFlags.Context, userAddFlags.Namespace, userAddFlags.Privilege, out)
		},
	}
}

func listCommand(parentFlags *flag.User) *cli.Command {
	flags := &flag.UserList{User: parentFlags}
	return &cli.Command{
		Name:  "list",
		Title: "List users in a SQL instance database.",
		Args: []cli.Argument{
			{Name: "app_name", Required: true},
		},
		ValidateFunc: cli.ValidateExactArgs(1),
		Flags:        flags,
		RunFunc: func(ctx context.Context, out cli.Output, args []string) error {
			return postgres.ListUsers(ctx, args[0], flags.Context, flags.Namespace, out)
		},
	}
}
