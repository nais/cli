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
			cleanupCommand(),
			rollbackCommand(),
		},
	}
}

func kubeConfigFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:        contextFlagName,
		Aliases:     []string{"c"},
		Usage:       "The kubeconfig `CONTEXT` to use",
		DefaultText: "The current context in your kubeconfig",
	}
}
