package gcloud

import (
	"context"

	"github.com/nais/cli/pkg/cli/v2"
)

func Login(ctx context.Context, out cli.Output) error {
	return executeGcloud(ctx, out, "auth", "login", "--update-adc")
}
