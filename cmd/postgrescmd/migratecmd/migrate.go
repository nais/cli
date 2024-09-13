package migratecmd

import (
	"github.com/nais/cli/pkg/gcp"
	"github.com/urfave/cli/v2"
)

const (
	contextFlagName = "context"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "Command used for migrating to a new Postgres instance",
		Before: func(context *cli.Context) error {
			return gcp.ValidateUserLogin(context.Context, false)
		},
		Subcommands: []*cli.Command{
			setupCommand(),
			promoteCommand(),
		},
	}
}
