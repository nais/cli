package command

import (
	"context"

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
		Description: "Lists all files mounted into the application from Secrets and Configs, with their paths and sources. Use 'nais secret view' to inspect secret contents.",
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		Flags: flags,
		ValidateFunc: func(context.Context, *naistrix.Arguments) error {
			return requireSingleEnvironment(flags.Environment)
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
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
		AutoCompleteFunc: autoCompleteAppNames(flags.App),
	}
}
