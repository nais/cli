package gcp

import (
	"context"

	"github.com/nais/cli/internal/root"
)

func Run(ctx context.Context, flags *root.Flags) error {
	return Login(ctx, flags)
}
