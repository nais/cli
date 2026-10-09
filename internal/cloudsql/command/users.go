package command

import (
	"context"

	"github.com/nais/cli/internal/cloudsql"
	"github.com/nais/cli/internal/cloudsql/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func usersCommand(parentFlags *flag.CloudSQL) *naistrix.Command {
	flags := &flag.User{CloudSQL: parentFlags}
	return &naistrix.Command{
		Name:         "users",
		Title:        "Manage users in your SQL instance.",
		Description:  "Commands for adding, listing, and dropping users in a Cloud SQL instance.",
		StickyFlags:  flags,
		ValidateFunc: validation.RequireTeamAndEnvironment(flags),
		SubCommands: []*naistrix.Command{
			addCommand(flags),
			dropCommand(flags),
			listUsersCommand(flags),
		},
	}
}

func addCommand(parentFlags *flag.User) *naistrix.Command {
	flags := &flag.UserAdd{
		User:      parentFlags,
		Privilege: "select",
	}
	return &naistrix.Command{
		Name:        "add",
		Title:       "Add a user to a SQL instance.",
		Description: "Will grant a user access to tables in public schema.",
		Args: []naistrix.Argument{
			{Name: "app_name", Prompt: "Name of the application whose Cloud SQL database should have a user added"},
			{Name: "username", Prompt: "Username of the new database user"},
			{Name: "password", Prompt: "Password for the new database user"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return cloudsql.AddUser(ctx, args.Get("app_name"), flags.Team, string(flags.Environment), args.Get("username"), args.Get("password"), flags, out)
		},
	}
}

func listUsersCommand(parentFlags *flag.User) *naistrix.Command {
	flags := &flag.UserList{User: parentFlags}
	return &naistrix.Command{
		Name:        "list",
		Title:       "List users in a SQL instance database.",
		Description: "List all users in a Cloud SQL instance database for a given application.",
		Args: []naistrix.Argument{
			{Name: "app_name", Prompt: "Name of the application whose Cloud SQL database users you want to list"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return cloudsql.ListUsers(ctx, args.Get("app_name"), flags.Team, string(flags.Environment), flags, out)
		},
	}
}

func dropCommand(parentFlags *flag.User) *naistrix.Command {
	flags := &flag.UserDrop{User: parentFlags}
	return &naistrix.Command{
		Name:        "drop",
		Title:       "Drop a user from a SQL instance database.",
		Description: "Remove a user from a Cloud SQL instance database.",
		Args: []naistrix.Argument{
			{Name: "app_name", Prompt: "Name of the application whose Cloud SQL database should have a user removed"},
			{Name: "username", Prompt: "Username of the database user to remove"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return cloudsql.DropUser(ctx, args.Get("app_name"), flags.Team, string(flags.Environment), args.Get("username"), flags, out)
		},
	}
}
