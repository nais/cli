package postgresCmd

import "github.com/urfave/cli/v2"

func usersCommand() *cli.Command {
	return &cli.Command{
		Name:    "users",
		Aliases: []string{"u"},
		Usage:   "Manage users in your Postgres instance",
		Subcommands: []*cli.Command{
			usersAddCommand(),
			usersListCommand(),
		},
	}
}
