package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/secret"
	"github.com/nais/cli/internal/secret/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
	"github.com/pterm/pterm"
)

// Entry represents a key-value pair in a secret. When values are not fetched,
// the Value field is empty and omitted from JSON output.
type Entry struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

type SecretDetail struct {
	Name         string              `json:"name"`
	Environment  string              `json:"environment"`
	Data         []Entry             `json:"data"`
	LastModified secret.LastModified `json:"lastModified"`
	ModifiedBy   string              `json:"modifiedBy,omitempty"`
	Workloads    []string            `json:"workloads,omitempty"`
}

func get(parentFlags *flag.Secret) *naistrix.Command {
	f := &flag.Get{Secret: parentFlags}
	return &naistrix.Command{
		Name:        "get",
		Title:       "Get details about a secret.",
		Description: "This command shows details about a secret, including its keys, workloads using it, and last modification info. Use --with-values to also fetch and display the actual secret values (access is logged for auditing).",
		Flags:       f,
		Args:        defaultArgs,
		ValidateFunc: func(_ context.Context, args *naistrix.Arguments) error {
			if err := validateSingleEnvironmentFlagUsage(); err != nil {
				return err
			}
			if err := validateArgs(args); err != nil {
				return err
			}
			if f.Reason != "" && !f.WithValues {
				return fmt.Errorf("--reason can only be used together with --with-values")
			}
			if f.WithValues && f.Reason != "" && len(f.Reason) < 10 {
				return fmt.Errorf("reason must be at least 10 characters")
			}
			return nil
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() == 0 {
				return autoCompleteSecretNames(ctx, f.Team, string(f.Environment), true)
			}
			return nil, ""
		},
		Examples: []naistrix.Example{
			{
				Description: "Get details for a secret named my-secret in environment dev.",
				Command:     "my-secret --environment dev",
			},
			{
				Description: "Get details including secret values (will prompt for reason).",
				Command:     "my-secret --environment dev --with-values",
			},
			{
				Description: "Get details including secret values with reason provided inline.",
				Command:     "my-secret --environment dev --with-values --reason \"Debugging production issue #1234\"",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			environment, err := resolveSecretEnvironment(ctx, f.Team, args.Get("name"), string(f.Environment))
			if err != nil {
				return err
			}

			return runGetCommand(ctx, args, out, f.Team, environment, f.Output, f.WithValues, f.Reason)
		},
	}
}

func runGetCommand(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter, team, environment string, outputFormat flag.Output, withValues bool, reason string) error {
	metadata := metadataFromArgs(args, team, environment)

	existing, err := secret.Get(ctx, metadata)
	if err != nil {
		return fmt.Errorf("fetching secret: %w", err)
	}

	entries := make([]Entry, len(existing.Keys))
	for i, k := range existing.Keys {
		entries[i] = Entry{Key: k}
	}

	if withValues {
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

		valueMap := make(map[string]string, len(values))
		for _, v := range values {
			valueMap[v.Name] = v.Value
		}

		for i := range entries {
			entries[i].Value = valueMap[entries[i].Key]
		}
	}

	if outputFormat == "json" {
		detail := SecretDetail{
			Name:         existing.Name,
			Environment:  existing.TeamEnvironment.Environment.Name,
			Data:         entries,
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

	pterm.DefaultSection.Println("Data")
	if len(entries) > 0 {
		var data [][]string
		if withValues {
			secretEntries := make([]secret.Entry, len(entries))
			for i, e := range entries {
				secretEntries[i] = secret.Entry{Key: e.Key, Value: e.Value}
			}
			data = secret.FormatDataWithValues(secretEntries)
		} else {
			data = secret.FormatData(existing.Keys)
		}
		err = pterm.DefaultTable.
			WithHasHeader().
			WithHeaderRowSeparator("-").
			WithData(data).
			Render()
		if err != nil {
			return fmt.Errorf("rendering data table: %w", err)
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
}
