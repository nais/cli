package command

import (
	"context"

	"github.com/nais/cli/internal/auth"
	"github.com/nais/cli/internal/auth/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func workloadIdentityMetadata(parentFlags *flag.Auth) *naistrix.Command {
	flags := &flag.WorkloadIdentityMetadata{Auth: parentFlags}
	return &naistrix.Command{
		Name:         "workload-identity-metadata",
		Title:        "Fetch the workload identity OIDC metadata for an environment.",
		Description:  "Resolves the workload identity OIDC issuer for the selected environment and fetches its metadata from /.well-known/openid-configuration, printing the JSON to stdout.",
		Flags:        flags,
		ValidateFunc: validation.RequireEnvironment(flags),
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			issuer, err := auth.GetEnvironmentOIDCIssuer(ctx, string(flags.Environment))
			if err != nil {
				return err
			}

			doc, err := auth.FetchOIDCDiscoveryDocument(ctx, issuer)
			if err != nil {
				return err
			}

			return out.JSON(output.JSONWithPrettyOutput()).Render(doc)
		},
	}
}
