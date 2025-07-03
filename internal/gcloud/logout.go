package gcloud

import (
	"context"

	"github.com/nais/naistrix"
)

func Logout(ctx context.Context, out naistrix.Output) error {
	if err := executeGcloud(ctx, out, "auth", "application-default", "revoke", "--quiet"); err != nil {
		return err
	}

	return executeGcloud(ctx, out, "auth", "revoke")
}
