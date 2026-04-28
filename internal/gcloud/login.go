package gcloud

import (
	"context"

	"github.com/nais/naistrix"
)

func Login(ctx context.Context, out *naistrix.OutputWriter, verbose bool) error {
	if err := executeGcloud(ctx, verbose, "auth", "login", "--update-adc"); err != nil {
		return err
	}

	if !verbose {
		out.Println("Logged in with gcloud.")
	}

	return nil
}
