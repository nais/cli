package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/secret"
	"github.com/nais/cli/internal/secret/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
	"github.com/pterm/pterm"
)

type SecretDetail struct {
	Name         string              `json:"name"`
	Environment  string              `json:"environment"`
	Keys         []string            `json:"keys"`
	LastModified secret.LastModified `json:"lastModified,omitempty"`
	ModifiedBy   string              `json:"modifiedBy,omitempty"`
	Workloads    []string            `json:"workloads,omitempty"`
}

func get(parentFlags *flag.Secret) *naistrix.Command {
	f := &flag.Get{Secret: parentFlags}
	return &naistrix.Command{
		Name:        "get",
		Title:       "Get details about a secret.",
		Description: "This command shows details about a secret, including its keys, workloads using it, and last modification info.",
		Flags:       f,
		Args:        defaultArgs,
		ValidateFunc: func(_ context.Context, args *naistrix.Arguments) error {
			if err := validation.CheckEnvironment(string(f.Environment)); err != nil {
				return err
			}
			return validateArgs(args)
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() == 0 {
				return autoCompleteSecretNames(ctx, parentFlags)
			}
			return nil, ""
		},
		Examples: []naistrix.Example{
			{
				Description: "Get details for a secret named my-secret in environment dev.",
				Command:     "my-secret --environment dev",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			metadata := metadataFromArgs(args, f.Team, string(f.Environment))

			existing, err := secret.Get(ctx, metadata)
			if err != nil {
				return fmt.Errorf("fetching secret: %w", err)
			}

			if f.Output == "json" {
				detail := SecretDetail{
					Name:         existing.Name,
					Environment:  existing.TeamEnvironment.Environment.Name,
					Keys:         existing.Keys,
					LastModified: secret.LastModified(existing.LastModifiedAt),
				}
				if existing.LastModifiedBy.Email != "" {
					detail.ModifiedBy = existing.LastModifiedBy.Email
				}
				for _, w := range existing.Workloads.Nodes {
					detail.Workloads = append(detail.Workloads, w.GetName())
				}
				return out.JSON(output.JSONWithPrettyOutput()).Render(detail)
			}

			pterm.DefaultSection.Println("Secret details")
			err = pterm.DefaultTable.
				WithHasHeader().
				WithHeaderRowSeparator("-").
				WithData(secret.FormatDetails(metadata, existing)).
				Render()
			if err != nil {
				return fmt.Errorf("rendering table: %w", err)
			}

			pterm.DefaultSection.Println("Keys")
			if len(existing.Keys) > 0 {
				err = pterm.DefaultTable.
					WithHasHeader().
					WithHeaderRowSeparator("-").
					WithData(secret.FormatKeys(existing)).
					Render()
				if err != nil {
					return fmt.Errorf("rendering keys table: %w", err)
				}
			} else {
				pterm.Info.Println("This secret has no keys.")
			}

			if len(existing.Workloads.Nodes) > 0 {
				pterm.DefaultSection.Println("Workloads using this secret")
				return pterm.DefaultTable.
					WithHasHeader().
					WithHeaderRowSeparator("-").
					WithData(secret.FormatWorkloads(existing)).
					Render()
			}

			pterm.Info.Println("No workloads are using this secret.")
			return nil
		},
	}
}
