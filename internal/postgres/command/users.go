package command

import "github.com/urfave/cli/v3"

func users() *cli.Command {
	return &cli.Command{
		Name:        "users",
		Usage:       "Administrate users in your Postgres instance",
		Description: "Command used for listing and adding users to database",
		Commands: []*cli.Command{
			usersAdd(),
			usersList(),
		},
	}
}
