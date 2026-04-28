package command

import (
	"context"
	"strings"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func env(parentFlags *flag.App) *naistrix.Command {
	flags := &flag.EnvVars{
		App: parentFlags,
	}

	return &naistrix.Command{
		Name:        "env",
		Title:       "Show environment variables for an application.",
		Description: "Lists all environment variables configured for the application with their values and sources. Secret values are hidden — use 'nais secret view' to reveal them.",
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			name := args.Get("name")

			environment, err := resolveAppEnvironment(ctx, out, flags.Team, name, string(flags.Environment), flags.Output == "json")
			if err != nil {
				return err
			}

			ret, err := app.GetApplicationEnvVars(ctx, flags.Team, name, environment)
			if err != nil {
				return err
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(ret)
			}

			if len(ret) == 0 {
				out.Println("No environment variables found.")
				return nil
			}

			if err := out.Table().Render(ret); err != nil {
				return err
			}

			if secretNames := app.UniqueSecretNames(ret); len(secretNames) > 0 {
				out.Println("")
				out.Printf("Secret values hidden. Use 'nais secret view <name> -e %s' to reveal.\n", environment)
				out.Printf("Secrets in use: %s\n", strings.Join(secretNames, ", "))
			}

			return nil
		},
		AutoCompleteFunc: autoCompleteAppNames(flags.App),
	}
}
