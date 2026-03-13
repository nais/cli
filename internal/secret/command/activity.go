package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/secret"
	"github.com/nais/cli/internal/secret/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func activity(parentFlags *flag.Secret) *naistrix.Command {
	f := &flag.Activity{
		Secret: parentFlags,
		Output: "table",
		Limit:  20,
	}

	return &naistrix.Command{
		Name:  "activity",
		Title: "Show activity for a secret.",
		Args:  defaultArgs,
		Flags: f,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			if err := validateArgs(args); err != nil {
				return err
			}

			activityTypes, err := parseActivityTypes(f.ActivityType)
			if err != nil {
				return err
			}

			ret, found, err := secret.GetActivity(ctx, f.Team, args.Get("name"), f.Environment, activityTypes, f.Limit)
			if err != nil {
				return err
			}

			if !found {
				out.Println("Secret not found.")
				return nil
			}

			if len(ret) == 0 {
				out.Println("No activity found for secret.")
				return nil
			}

			if f.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(ret)
			}

			return out.Table().Render(ret)
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() == 0 {
				if f.Team == "" {
					return nil, "Please provide team to auto-complete secret names. 'nais config set team <team>', or '--team <team>' flag."
				}
				return autoCompleteSecretNames(ctx, f.Secret)
			}
			return nil, ""
		},
	}
}

func parseActivityTypes(in []string) ([]gql.ActivityLogActivityType, error) {
	ret := make([]gql.ActivityLogActivityType, 0, len(in))
	allowed := make(map[string]gql.ActivityLogActivityType, len(gql.AllActivityLogActivityType))
	for _, v := range gql.AllActivityLogActivityType {
		allowed[string(v)] = v
	}

	for _, t := range in {
		normalized := strings.ToUpper(strings.TrimSpace(t))
		v, ok := allowed[normalized]
		if !ok {
			return nil, fmt.Errorf("invalid activity type %q", t)
		}
		ret = append(ret, v)
	}

	return ret, nil
}
