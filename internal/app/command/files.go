package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func files(parentFlags *flag.App) *naistrix.Command {
	flags := &flag.Files{
		App: parentFlags,
	}

	return &naistrix.Command{
		Name:        "files",
		Title:       "Show mounted files for an application.",
		Description: "Lists all files mounted into the application from Secrets and ConfigMaps, with their paths and sources. Use 'nais secret view' to inspect secret contents.",
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			if len(flags.Environment) != 1 {
				return fmt.Errorf("exactly one environment must be specified with -e/--environment")
			}

			ret, err := app.GetApplicationFiles(ctx, flags.Team, args.Get("name"), flags.Environment)
			if err != nil {
				return err
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(ret)
			}

			if len(ret) == 0 {
				out.Println("No mounted files found.")
				return nil
			}

			if err := out.Table().Render(ret); err != nil {
				return err
			}

			if app.HasSecretFiles(ret) {
				out.Println("")
				out.Printf("To view secret contents, use 'nais secret view <name> -e %s'.\n", flags.Environment[0])
			}

			return nil
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() == 0 {
				if len(flags.Team) == 0 {
					return nil, "Please provide team to auto-complete application names. 'nais defaults set team <team>', or '--team <team>' flag."
				}
				apps, err := app.GetApplicationNames(ctx, flags.Team, flags.Environment)
				if err != nil {
					return nil, "Unable to fetch application names."
				}
				return apps, "Select an application."
			}
			return nil, ""
		},
	}
}
