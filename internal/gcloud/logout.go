package gcloud

import (
	"context"

	"github.com/nais/naistrix"
)

func Logout(ctx context.Context, out *naistrix.OutputWriter, verbose bool) error {
	if err := executeGcloud(ctx, verbose, "auth", "application-default", "revoke", "--quiet"); err != nil {
		return err
	}

	if err := executeGcloud(ctx, verbose, "auth", "revoke"); err != nil {
		return err
	}

	if !verbose {
		out.Println("Logged out of gcloud.")
	}

	return nil
}
