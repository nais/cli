package connect

import (
	"context"

	"github.com/nais/cli/internal/naisdevice"
)

func Run(ctx context.Context) error {
	return naisdevice.Connect(ctx)
}
