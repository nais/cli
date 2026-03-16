package command

import (
	"context"
	"fmt"
	"sort"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/opensearch"
	"github.com/nais/cli/internal/opensearch/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func credentials(parentFlags *flag.OpenSearch) *naistrix.Command {
	flags := &flag.Credentials{OpenSearch: parentFlags}
	return &naistrix.Command{
		Name:        "credentials",
		Title:       "Create temporary credentials for an OpenSearch instance.",
		Description: "Creates temporary credentials for accessing an OpenSearch instance. The credentials are printed to stdout as environment variables.",
		Flags:       flags,
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		ValidateFunc: func(ctx context.Context, args *naistrix.Arguments) error {
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
			if args.Len() == 0 {
				instances, err := opensearch.GetAll(ctx, flags.Team)
				if err != nil {
					return nil, "Unable to fetch OpenSearch instances."
				}
				environment := string(flags.Environment)
				var names []string
				for _, instance := range instances {
					if environment != "" && instance.TeamEnvironment.Environment.Name != environment {
						continue
					}
					names = append(names, instance.Name)
				}
				sort.Strings(names)
				return names, "Select an OpenSearch instance."
			}
			return nil, ""
		},
		Examples: []naistrix.Example{
			{
				Description: "Create read credentials for an OpenSearch instance named my-opensearch in environment dev, valid for 1 day.",
				Command:     "my-opensearch --environment dev --permission READ --ttl 1d",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			creds, err := opensearch.CreateCredentials(
				ctx,
				flags.Team,
				string(flags.Environment),
				args.Get("name"),
				gql.AivenPermission(flags.Permission),
				flags.TTL,
			)
			if err != nil {
				return fmt.Errorf("creating OpenSearch credentials: %w", err)
			}

			out.Println(fmt.Sprintf("OPEN_SEARCH_URI=%q", creds.Uri))
			out.Println(fmt.Sprintf("OPEN_SEARCH_HOST=%q", creds.Host))
			out.Println(fmt.Sprintf(`OPEN_SEARCH_PORT="%d"`, creds.Port))
			out.Println(fmt.Sprintf("OPEN_SEARCH_USERNAME=%q", creds.Username))
			out.Println(fmt.Sprintf("OPEN_SEARCH_PASSWORD=%q", creds.Password))
			return nil
		},
	}
}

func isValidAivenPermission(permission gql.AivenPermission) bool {
	for _, p := range gql.AllAivenPermission {
		if p == permission {
			return true
		}
	}
	return false
}
