package gcloud

import (
	"context"

	"github.com/nais/naistrix"
)

func Login(ctx context.Context, out naistrix.Output) error {
	return executeGcloud(ctx, out, "auth", "login", "--update-adc")
}
