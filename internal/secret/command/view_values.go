package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/secret/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
	"github.com/pterm/pterm"
)

type SecretValueRow struct {
	Key   string `heading:"Key" json:"key"`
	Value string `heading:"Value" json:"value"`
}

func viewValues(parentFlags *flag.Secret) *naistrix.Command {
	f := &flag.ViewValues{Secret: parentFlags}
	return &naistrix.Command{
		Name:        "view-values",
		Title:       "View secret values.",
		Description: "View the actual values of a secret. This requires team membership and a reason for access. The access is logged for auditing purposes.",
		Flags:       f,
		Args:        defaultArgs,
		ValidateFunc: func(_ context.Context, args *naistrix.Arguments) error {
			if err := validation.CheckEnvironment(string(f.Environment)); err != nil {
				return err
			}
			if err := validateArgs(args); err != nil {
				return err
			}
			if f.Reason != "" && len(f.Reason) < 10 {
				return fmt.Errorf("reason must be at least 10 characters")
			}
			return nil
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() == 0 {
				return autoCompleteSecretNames(ctx, parentFlags)
			}
			return nil, ""
		},
		Examples: []naistrix.Example{
			{
				Description: "View secret values (will prompt for reason).",
				Command:     "my-secret --environment dev",
			},
			{
				Description: "Provide reason inline.",
				Command:     "my-secret --environment dev --reason \"Debugging production issue #1234\"",
			},
			{
				Description: "Output as JSON.",
				Command:     "my-secret --environment dev --reason \"Automated backup\" --output json",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			metadata := metadataFromArgs(args, f.Team, string(f.Environment))

			reason := f.Reason
			if reason == "" {
				pterm.Warning.Println("Viewing secret values is logged for auditing purposes.")
				result, _ := pterm.DefaultInteractiveTextInput.
					WithDefaultText("Reason for accessing secret values (min 10 chars)").
					Show()
				if len(result) < 10 {
					return fmt.Errorf("reason must be at least 10 characters")
				}
				reason = result
			}

			values, err := naisapi.ViewSecretValues(ctx, metadata.TeamSlug, metadata.EnvironmentName, metadata.Name, reason)
			if err != nil {
				return fmt.Errorf("viewing secret values: %w", err)
			}

			if len(values) == 0 {
				out.Infoln("Secret has no values")
				return nil
			}

			rows := make([]SecretValueRow, len(values))
			for i, v := range values {
				rows[i] = SecretValueRow{
					Key:   v.Name,
					Value: v.Value,
				}
			}

			if f.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(rows)
			}

			return out.Table().Render(rows)
		},
	}
}
