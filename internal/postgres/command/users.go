package command

import "github.com/urfave/cli/v2"

func users() *cli.Command {
	return &cli.Command{
		Name:        "users",
		Usage:       "Administrate users in your Postgres instance",
		Description: "Command used for listing and adding users to database",
		Subcommands: []*cli.Command{
			usersAdd(),
			usersList(),
		},
	}
}
