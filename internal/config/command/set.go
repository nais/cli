package command

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/nais/cli/internal/config"
	"github.com/nais/cli/internal/config/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
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

			// Count the number of value sources provided
			sources := 0
			if f.Value != "" {
				sources++
			}
			if f.ValueFromStdin {
				sources++
			}
			if f.ValueFromFile != "" {
				sources++
			}
			if sources == 0 {
				return fmt.Errorf("--value, --value-from-stdin, or --value-from-file is required")
			}
			if sources > 1 {
				return fmt.Errorf("--value, --value-from-stdin, and --value-from-file are mutually exclusive")
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
			{
				Description: "Upload a file as a config value. Binary files are automatically Base64-encoded.",
				Command:     "my-config --environment prod --key keystore.p12 --value-from-file ./keystore.p12",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			metadata := metadataFromArgs(args, f.Team, string(f.Environment))

			var value string
			encoding := gql.ValueEncodingPlainText

			switch {
			case f.ValueFromFile != "":
				data, err := os.ReadFile(f.ValueFromFile)
				if err != nil {
					return fmt.Errorf("reading file %q: %w", f.ValueFromFile, err)
				}
				if utf8.Valid(data) {
					value = string(data)
				} else {
					value = base64.StdEncoding.EncodeToString(data)
					encoding = gql.ValueEncodingBase64
				}
			case f.ValueFromStdin:
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("reading from stdin: %w", err)
				}
				value = strings.TrimSuffix(string(data), "\n")
			default:
				value = f.Value
			}

			// Check the value size early to give a clear error message.
			const maxValueSize = 1 << 20 // 1 MiB
			if len(value) > maxValueSize {
				return fmt.Errorf("value too large (%d bytes); maximum size is 1 MiB", len(value))
			}

			updated, err := config.SetValue(ctx, metadata, f.Key, value, encoding)
			if err != nil {
				return fmt.Errorf("setting config value: %w", err)
			}

			if updated {
				out.Successf("Updated key %q in config %q in %q\n", f.Key, metadata.Name, metadata.EnvironmentName)
			} else {
				out.Successf("Added key %q to config %q in %q\n", f.Key, metadata.Name, metadata.EnvironmentName)
			}

			return nil
		},
	}
}
