package gcloud

import (
	"context"

	"github.com/nais/naistrix"
)

func Login(ctx context.Context, out naistrix.Output, verbose bool) error {
	return executeGcloud(ctx, out, verbose, "auth", "login", "--update-adc")
}
