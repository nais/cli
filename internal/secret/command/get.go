package command

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/secret"
	"github.com/nais/cli/internal/secret/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
	"github.com/pterm/pterm"
)

// Entry represents a key-value pair in a secret. When values are not fetched,
// the Value field is empty and omitted from JSON output.
type Entry struct {
	Key      string            `json:"key"`
	Value    string            `json:"value,omitempty"`
	Encoding gql.ValueEncoding `json:"encoding,omitempty"`
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
			if providedEnvironment := string(f.Environment); providedEnvironment != "" {
				if err := validation.CheckEnvironment(providedEnvironment); err != nil {
					return err
				}
			}
			if err := validateArgs(args); err != nil {
				return err
			}
			if f.ToFile != "" && f.Key == "" {
				return fmt.Errorf("--to-file requires --key to specify which key to extract")
			}
			if f.Key != "" && f.ToFile == "" {
				return fmt.Errorf("--key is only used with --to-file")
			}
			if f.Reason != "" && !f.WithValues && f.ToFile == "" {
				return fmt.Errorf("--reason can only be used together with --with-values or --to-file")
			}
			if (f.WithValues || f.ToFile != "") && f.Reason != "" && len(f.Reason) < 10 {
				return fmt.Errorf("reason must be at least 10 characters")
			}
			return nil
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() == 0 {
				return autoCompleteSecretNames(ctx, f.Team, string(f.Environment), false)
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
			{
				Description: "Extract a binary value (e.g. keystore) to a file.",
				Command:     "my-secret --environment prod --key keystore.p12 --to-file ./keystore.p12 --reason \"Need keystore for local testing\"",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			opts := getOptions{
				team:         f.Team,
				outputFormat: f.Output,
				withValues:   f.WithValues || f.ToFile != "",
				reason:       f.Reason,
				toFile:       f.ToFile,
				key:          f.Key,
			}

			if providedEnvironment := string(f.Environment); providedEnvironment != "" {
				opts.environment = providedEnvironment
				return runGetCommand(ctx, args, out, opts)
			}

			environment, err := resolveSecretEnvironment(ctx, f.Team, args.Get("name"), string(f.Environment))
			if err != nil {
				return err
			}

			opts.environment = environment
			return runGetCommand(ctx, args, out, opts)
		},
	}
}

type getOptions struct {
	team         string
	environment  string
	outputFormat flag.Output
	withValues   bool
	reason       string
	toFile       string
	key          string
}

func runGetCommand(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter, opts getOptions) error {
	metadata := metadataFromArgs(args, opts.team, opts.environment)

	existing, err := secret.Get(ctx, metadata)
	if err != nil {
		return fmt.Errorf("fetching secret: %w", err)
	}

	entries := make([]Entry, len(existing.Keys))
	for i, k := range existing.Keys {
		entries[i] = Entry{Key: k}
	}

	if opts.withValues {
		reason := opts.reason
		if reason == "" {
			pterm.Warning.Println("Viewing secret values is logged for auditing purposes.")
			result, err := pterm.DefaultInteractiveTextInput.
				WithDefaultText("Reason for accessing secret values (min 10 chars)").
				Show()
			if err != nil {
				return fmt.Errorf("prompting for reason: %w", err)
			}
			if len(result) < 10 {
				return fmt.Errorf("reason must be at least 10 characters")
			}
			reason = result
		}

		values, err := naisapi.ViewSecretValues(ctx, metadata.TeamSlug, metadata.EnvironmentName, metadata.Name, reason)
		if err != nil {
			return fmt.Errorf("viewing secret values: %w", err)
		}

		type valueInfo struct {
			value    string
			encoding gql.ValueEncoding
		}
		valueMap := make(map[string]valueInfo, len(values))
		for _, v := range values {
			valueMap[v.Name] = valueInfo{value: v.Value, encoding: v.Encoding}
		}

		for i := range entries {
			if info, ok := valueMap[entries[i].Key]; ok {
				entries[i].Value = info.value
				entries[i].Encoding = info.encoding
			}
		}

		// Handle --to-file: extract a single key's value to a file
		if opts.toFile != "" {
			info, ok := valueMap[opts.key]
			if !ok {
				return fmt.Errorf("key %q not found in secret %q", opts.key, metadata.Name)
			}

			var data []byte
			if info.encoding == gql.ValueEncodingBase64 {
				data, err = base64.StdEncoding.DecodeString(info.value)
				if err != nil {
					return fmt.Errorf("decoding base64 value for key %q: %w", opts.key, err)
				}
			} else {
				data = []byte(info.value)
			}

			if err := os.WriteFile(opts.toFile, data, 0o600); err != nil {
				return fmt.Errorf("writing to file %q: %w", opts.toFile, err)
			}

			pterm.Success.Printfln("Wrote key %q (%d bytes) to %s", opts.key, len(data), opts.toFile)
			return nil
		}
	}

	if opts.outputFormat == "json" {
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
		if opts.withValues {
			secretEntries := make([]secret.Entry, len(entries))
			for i, e := range entries {
				secretEntries[i] = secret.Entry{Key: e.Key, Value: e.Value, Encoding: e.Encoding}
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
