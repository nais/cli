package command

import (
	"context"

	"github.com/nais/cli/v2/internal/postgres"
	"github.com/nais/cli/v2/internal/postgres/command/flag"
	"github.com/nais/naistrix"
)

func usersCommand(parentFlags *flag.Postgres) *naistrix.Command {
	flags := &flag.User{Postgres: parentFlags}
	return &naistrix.Command{
		Name:        "users",
		Title:       "Manage users in your SQL instance.",
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			addCommand(flags),
			listCommand(flags),
		},
	}
}

func addCommand(parentFlags *flag.User) *naistrix.Command {
	userAddFlags := &flag.UserAdd{
		User:      parentFlags,
		Privilege: "select",
	}
	return &naistrix.Command{
		Name:        "add",
		Title:       "Add a user to a SQL instance.",
		Description: "Will grant a user access to tables in public schema.",
		Args: []naistrix.Argument{
			{Name: "app_name"},
			{Name: "username"},
			{Name: "password"},
		},
		Flags: userAddFlags,
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			return postgres.AddUser(ctx, args[0], args[1], args[2], userAddFlags.Context, userAddFlags.Namespace, userAddFlags.Privilege, out)
		},
	}
}

func listCommand(parentFlags *flag.User) *naistrix.Command {
	flags := &flag.UserList{User: parentFlags}
	return &naistrix.Command{
		Name:  "list",
		Title: "List users in a SQL instance database.",
		Args: []naistrix.Argument{
			{Name: "app_name"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			return postgres.ListUsers(ctx, args[0], flags.Context, flags.Namespace, out)
		},
	}
}
