package command

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/cli/internal/valkey"
	"github.com/nais/cli/internal/valkey/command/flag"
	"github.com/nais/naistrix"
)

func credentials(parentFlags *flag.Valkey) *naistrix.Command {
	flags := &flag.Credentials{Valkey: parentFlags}
	return &naistrix.Command{
		Name:        "credentials",
		Title:       "Create temporary credentials for a Valkey instance.",
		Description: "Creates temporary credentials for accessing a Valkey instance. The credentials are printed to stdout as environment variables.",
		Flags:       flags,
		Args:        defaultArgs,
		ValidateFunc: func(ctx context.Context, args *naistrix.Arguments) error {
			if err := validateSingleEnvironmentFlagUsage(); err != nil {
				return err
			}
			if err := validation.CheckEnvironment(string(flags.Environment)); err != nil {
				return err
			}
			if err := validateArgs(args); err != nil {
				return err
			}
			if flags.Permission == "" {
				return fmt.Errorf("permission is required, set using --permission/-p flag (READ, WRITE, READWRITE, ADMIN)")
			}
			if !isValidAivenPermission(gql.AivenPermission(flags.Permission)) {
				return fmt.Errorf("invalid permission %q, must be one of: %v", flags.Permission, gql.AllAivenPermission)
			}
			if flags.TTL == "" {
				return fmt.Errorf("ttl is required, set using --ttl flag (e.g. '1d', '7d')")
			}
			return nil
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() != 0 {
				return nil, ""
			}
			return autoCompleteValkeyNames(ctx, flags.Team, string(flags.Environment), true)
		},
		Examples: []naistrix.Example{
			{
				Description: "Create read credentials for a Valkey instance named my-valkey in environment dev, valid for 1 day.",
				Command:     "my-valkey --environment dev --permission READ --ttl 1d",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			creds, err := valkey.CreateCredentials(
				ctx,
				flags.Team,
				string(flags.Environment),
				args.Get("name"),
				gql.AivenPermission(flags.Permission),
				flags.TTL,
			)
			if err != nil {
				return fmt.Errorf("creating Valkey credentials: %w", err)
			}

			out.Println(fmt.Sprintf("VALKEY_URI=%q", creds.Uri))
			out.Println(fmt.Sprintf("VALKEY_HOST=%q", creds.Host))
			out.Println(fmt.Sprintf("VALKEY_PORT=%q", strconv.Itoa(int(creds.Port))))
			out.Println(fmt.Sprintf("VALKEY_USERNAME=%q", creds.Username))
			out.Println(fmt.Sprintf("VALKEY_PASSWORD=%q", creds.Password))
			return nil
		},
	}
}

func isValidAivenPermission(permission gql.AivenPermission) bool {
	return slices.Contains(gql.AllAivenPermission, permission)
}
