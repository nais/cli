package gcloud

import (
	"context"

	"github.com/nais/naistrix"
)

func PrintAccessToken(ctx context.Context, out *naistrix.OutputWriter) error {
	return executeGcloud(ctx, out, true, "auth", "print-access-token")
}
