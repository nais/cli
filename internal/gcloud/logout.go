package gcloud

import (
	"context"

	"github.com/nais/cli/pkg/cli"
)

func Logout(ctx context.Context, out cli.Output) error {
	if err := executeGcloud(ctx, out, "auth", "application-default", "revoke", "--quiet"); err != nil {
		return err
	}

	return executeGcloud(ctx, out, "auth", "revoke")
}
