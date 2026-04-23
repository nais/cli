package command

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/nais/cli/internal/config"
	"github.com/nais/cli/internal/config/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

// Entry represents a key-value pair in a config.
type Entry struct {
	Key      string            `json:"key"`
	Value    string            `json:"value"`
	Encoding gql.ValueEncoding `json:"encoding,omitempty"`
}

type ConfigDetail struct {
	Name         string              `json:"name"`
	Environment  string              `json:"environment"`
	Data         []Entry             `json:"data"`
	LastModified config.LastModified `json:"lastModified"`
	ModifiedBy   string              `json:"modifiedBy,omitempty"`
	Workloads    []string            `json:"workloads,omitempty"`
}

func get(parentFlags *flag.Config) *naistrix.Command {
	f := &flag.Get{Config: parentFlags}
	return &naistrix.Command{
		Name:        "get",
		Title:       "Get details about a config.",
		Description: "This command shows details about a config, including its key-value pairs, workloads using it, and last modification info.",
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
			return nil
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() == 0 {
				return autoCompleteConfigNames(ctx, f.Team, string(f.Environment), false)
			}
			return nil, ""
		},
		Examples: []naistrix.Example{
			{
				Description: "Get details for a config named my-config in environment dev.",
				Command:     "my-config --environment dev",
			},
			{
				Description: "Extract a binary value (e.g. keystore) to a file.",
				Command:     "my-config --environment prod --key keystore.p12 --to-file ./keystore.p12",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			opts := getOptions{
				team:         f.Team,
				outputFormat: f.Output,
				toFile:       f.ToFile,
				key:          f.Key,
			}

			if providedEnvironment := string(f.Environment); providedEnvironment != "" {
				opts.environment = providedEnvironment
				return runGetCommand(ctx, args, out, opts)
			}

			environment, err := resolveConfigEnvironment(ctx, f.Team, args.Get("name"), string(f.Environment))
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
	toFile       string
	key          string
}

func runGetCommand(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter, opts getOptions) error {
	metadata := metadataFromArgs(args, opts.team, opts.environment)

	existing, err := config.Get(ctx, metadata)
	if err != nil {
		return fmt.Errorf("fetching config: %w", err)
	}

	entries := make([]Entry, len(existing.Values))
	for i, v := range existing.Values {
		entries[i] = Entry{Key: v.Name, Value: v.Value, Encoding: v.Encoding}
	}

	// Handle --to-file: extract a single key's value to a file
	if opts.toFile != "" {
		var found *Entry
		for i := range entries {
			if entries[i].Key == opts.key {
				found = &entries[i]
				break
			}
		}
		if found == nil {
			return fmt.Errorf("key %q not found in config %q", opts.key, metadata.Name)
		}

		var data []byte
		if found.Encoding == gql.ValueEncodingBase64 {
			data, err = base64.StdEncoding.DecodeString(found.Value)
			if err != nil {
				return fmt.Errorf("decoding base64 value for key %q: %w", opts.key, err)
			}
		} else {
			data = []byte(found.Value)
		}

		if err := os.WriteFile(opts.toFile, data, 0o600); err != nil {
			return fmt.Errorf("writing to file %q: %w", opts.toFile, err)
		}

		out.Successf("Wrote key %q (%d bytes) to %s\n", opts.key, len(data), opts.toFile)
		return nil
	}

	if opts.outputFormat == "json" {
		detail := ConfigDetail{
			Name:         existing.Name,
			Environment:  existing.TeamEnvironment.Environment.Name,
			Data:         entries,
			LastModified: config.LastModified(existing.LastModifiedAt),
		}
		if existing.LastModifiedBy.Email != "" {
			detail.ModifiedBy = existing.LastModifiedBy.Email
		}
		for _, w := range existing.Workloads.Nodes {
			detail.Workloads = append(detail.Workloads, w.GetName())
		}
		return out.JSON(output.JSONWithPrettyOutput()).Render(detail)
	}

	if err := out.Table().Render(config.FormatDetails(metadata, existing)); err != nil {
		return fmt.Errorf("rendering table: %w", err)
	}

	out.Println()
	if len(entries) > 0 {
		if err := out.Table().Render(config.FormatData(existing.Values)); err != nil {
			return fmt.Errorf("rendering data table: %w", err)
		}
	} else {
		out.Infoln("This config has no keys.")
	}

	out.Println()
	if len(existing.Workloads.Nodes) > 0 {
		return out.Table().Render(config.FormatWorkloads(existing))
	}

	out.Infoln("No workloads are using this config.")
	return nil
}
