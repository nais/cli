package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/input"
)

func del(parentFlags *flag.App) *naistrix.Command {
	flags := &flag.Delete{
		App: parentFlags,
	}

	return &naistrix.Command{
		Name:        "delete",
		Title:       "Delete an application.",
		Description: "Permanently deletes an application.",
		Flags:       flags,
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			name := args.Get("name")

			environment, err := resolveAppEnvironment(ctx, flags.Team, name, string(flags.Environment))
			if err != nil {
				return err
			}

			out.Warnf("Deleting %s permanently removes the application. Attached resources such as databases, buckets and configuration may be permanently deleted or left orphaned depending on their configuration. Review the application in the Nais Console if you are unsure.\n", name)

			if !flags.Yes {
				expected := environment + "/" + name
				answer, err := input.Input(fmt.Sprintf("Confirm deletion by typing %q", expected))
				if err != nil {
					return err
				}
				if answer != expected {
					return fmt.Errorf("cancelled by user")
				}
			}

			if err := app.DeleteApp(ctx, flags.Team, name, environment); err != nil {
				return err
			}

			out.Printf("Deletion of %s in %s has been started.\n", name, environment)
			return nil
		},
		AutoCompleteFunc: autoCompleteAppNames(parentFlags),
	}
}
