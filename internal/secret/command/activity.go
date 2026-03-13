package command

import (
	"context"

	activityutil "github.com/nais/cli/internal/activity"
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

			activityTypes, err := activityutil.ParseActivityTypes(f.ActivityType)
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
