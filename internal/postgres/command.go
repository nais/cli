package postgres

import (
	"context"

	"github.com/nais/cli/internal/gcp"
	"github.com/urfave/cli/v3"
)

func Before(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	_, err := gcp.ValidateAndGetUserLogin(ctx, false)
	return ctx, err
}
