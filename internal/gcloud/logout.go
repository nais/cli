package gcloud

import (
	"context"

	"github.com/nais/naistrix"
)

func Logout(ctx context.Context, out *naistrix.OutputWriter, verbose bool) error {
	if err := executeGcloud(ctx, out, verbose, "auth", "application-default", "revoke", "--quiet"); err != nil {
		return err
	}

	return executeGcloud(ctx, out, verbose, "auth", "revoke")
}
