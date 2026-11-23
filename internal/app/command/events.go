package command

import (
	"context"

	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/cli/internal/workload"
	events "github.com/nais/cli/internal/workload/command"
	"github.com/nais/naistrix"
)

func appEvents(parentFlags *flag.App) *naistrix.Command {
	flags := &flag.Events{App: parentFlags, Limit: events.DefaultEventLimit}
	return &naistrix.Command{
		Name:        "events",
		Title:       "Show recent events for an application.",
		Description: "Shows workload warnings, restarts, and successful rescaling, newest first. Events are retained for 30 days. Use app issues for active issues or app activity for the audit log.",
		Args: []naistrix.Argument{
			{Name: "name", Prompt: "Name of the application", ChoicesFunc: appNameChoices(parentFlags)},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			query, err := events.ParseEventQuery(flags.Limit, flags.From, flags.To, flags.After)
			if err != nil {
				return err
			}
			name := args.Get("name")
			environment, err := resolveAppEnvironment(ctx, flags.Team, name, string(flags.Environment))
			if err != nil {
				return err
			}
			page, err := workload.GetEvents(ctx, workload.ApplicationEvents, flags.Team, environment, name, query)
			if err != nil {
				return err
			}
			return events.RenderEvents(out, page, query, flags.Output == "json")
		},
	}
}
