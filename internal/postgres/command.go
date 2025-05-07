package postgres

import (
	"context"

	"github.com/nais/cli/internal/gcp"
	"github.com/nais/cli/internal/root"
)

type Flags struct {
	root.Flags
	Namespace string
	Context   string
}

// TODO: Do something with this
func Before(ctx context.Context) (context.Context, error) {
	_, err := gcp.ValidateAndGetUserLogin(ctx, false)
	return ctx, err
}
