package tidy

import (
	"context"

	"github.com/nais/cli/internal/aiven"
	"github.com/urfave/cli/v3"
)

func Action(ctx context.Context, cmd *cli.Command) error {
	return aiven.TidyLocalSecrets()
}
