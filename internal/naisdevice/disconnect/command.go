package disconnect

import (
	"context"

	"github.com/nais/cli/internal/naisdevice"
	"github.com/urfave/cli/v3"
)

func Action(ctx context.Context, cmd *cli.Command) error {
	return naisdevice.Disconnect(ctx)
}
