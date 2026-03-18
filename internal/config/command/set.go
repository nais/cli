package command

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nais/cli/internal/config"
	"github.com/nais/cli/internal/config/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/pterm/pterm"
)

func set(parentFlags *flag.Config) *naistrix.Command {
	f := &flag.Set{Config: parentFlags}
	return &naistrix.Command{
		Name:        "set",
		Title:       "Set a key-value pair in a config.",
		Description: "Set a key-value pair in a config. If the key already exists, its value is updated. If the key does not exist, it is added. Updating a value will cause a restart of workloads referencing the config.",
		Flags:       f,
		Args:        defaultArgs,
		ValidateFunc: func(_ context.Context, args *naistrix.Arguments) error {
			if err := validateSingleEnvironmentFlagUsage(); err != nil {
				return err
			}
			if err := validation.CheckEnvironment(string(f.Environment)); err != nil {
				return err
			}
			if err := validateArgs(args); err != nil {
				return err
			}
			if f.Key == "" {
				return fmt.Errorf("--key is required")
			}
			if f.Value == "" && !f.ValueFromStdin {
				return fmt.Errorf("--value or --value-from-stdin is required")
			}
			if f.Value != "" && f.ValueFromStdin {
				return fmt.Errorf("--value and --value-from-stdin are mutually exclusive")
			}
			return nil
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() == 0 {
				return autoCompleteConfigNames(ctx, f.Team, string(f.Environment), true)
			}
			return nil, ""
		},
		Examples: []naistrix.Example{
			{
				Description: "Set a key-value pair in a config.",
				Command:     "my-config --environment dev --key DATABASE_HOST --value db.example.com",
			},
			{
				Description: "Read value from stdin (useful for multi-line values).",
				Command:     "my-config --environment dev --key CONFIG_FILE --value-from-stdin < config.yaml",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			metadata := metadataFromArgs(args, f.Team, string(f.Environment))

			value := f.Value
			if f.ValueFromStdin {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("reading from stdin: %w", err)
				}
				value = strings.TrimSuffix(string(data), "\n")
			}

			updated, err := config.SetValue(ctx, metadata, f.Key, value)
			if err != nil {
				return fmt.Errorf("setting config value: %w", err)
			}

			if updated {
				pterm.Success.Printfln("Updated key %q in config %q in %q", f.Key, metadata.Name, metadata.EnvironmentName)
			} else {
				pterm.Success.Printfln("Added key %q to config %q in %q", f.Key, metadata.Name, metadata.EnvironmentName)
			}

			return nil
		},
	}
}
