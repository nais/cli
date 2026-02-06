package command

import (
	"context"

	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
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
			dropCommand(flags),
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
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return postgres.AddUser(ctx, args.Get("app_name"), args.Get("username"), args.Get("password"), userAddFlags, out)
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
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return postgres.ListUsers(ctx, args.Get("app_name"), flags, out)
		},
	}
}

func dropCommand(parentFlags *flag.User) *naistrix.Command {
	flags := &flag.UserDrop{User: parentFlags}

	return &naistrix.Command{
		Name:  "drop",
		Title: "Drop a user from a SQL instance database.",
		Args: []naistrix.Argument{
			{Name: "app_name"},
			{Name: "username"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return postgres.DropUser(ctx, args.Get("app_name"), args.Get("username"), flags, out)
		},
	}
}
