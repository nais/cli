package command

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/secret"
	"github.com/nais/cli/internal/secret/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/pterm/pterm"
)

func set(parentFlags *flag.Secret) *naistrix.Command {
	f := &flag.Set{Secret: parentFlags}
	return &naistrix.Command{
		Name:        "set",
		Title:       "Set a key-value pair in a secret.",
		Description: "Set a key-value pair in a secret. If the key already exists, its value is updated. If the key does not exist, it is added. Updating a value will cause a restart of workloads referencing the secret.",
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
				return autoCompleteSecretNames(ctx, f.Team, string(f.Environment), true)
			}
			return nil, ""
		},
		Examples: []naistrix.Example{
			{
				Description: "Set a key-value pair in a secret.",
				Command:     "my-secret --environment dev --key DATABASE_URL --value postgres://localhost/mydb",
			},
			{
				Description: "Read value from stdin (useful for multi-line values or avoiding shell history).",
				Command:     "my-secret --environment dev --key TLS_CERT --value-from-stdin < cert.pem",
			},
			{
				Description: "Upload a file as a secret value. Binary files are automatically Base64-encoded.",
				Command:     "my-secret --environment prod --key keystore.p12 --value-from-file ./keystore.p12",
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

			updated, err := secret.SetValue(ctx, metadata, f.Key, value, encoding)
			if err != nil {
				return fmt.Errorf("setting secret value: %w", err)
			}

			if updated {
				pterm.Success.Printfln("Updated key %q in secret %q in %q", f.Key, metadata.Name, metadata.EnvironmentName)
			} else {
				pterm.Success.Printfln("Added key %q to secret %q in %q", f.Key, metadata.Name, metadata.EnvironmentName)
			}

			return nil
		},
	}
}
