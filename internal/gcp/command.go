package gcp

import (
	"context"

	"github.com/urfave/cli/v3"
)

func LoginCommand(ctx context.Context, cmd *cli.Command) error {
	return Login(ctx)
}
