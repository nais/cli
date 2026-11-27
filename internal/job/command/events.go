package command

import (
	"context"

	"github.com/nais/cli/internal/job/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/cli/internal/workload"
	events "github.com/nais/cli/internal/workload/command"
	"github.com/nais/naistrix"
)

func jobEvents(parentFlags *flag.Job) *naistrix.Command {
	flags := &flag.Events{Job: parentFlags, Limit: events.DefaultEventLimit}
	return &naistrix.Command{
		Name:        "events",
		Title:       "Show recent events for a job.",
		Description: "Shows workload warnings and restarts, newest first. Events are retained for 30 days and can span multiple job runs. Use job issues for active issues or job activity for the audit log.",
		Args: []naistrix.Argument{
			{Name: "name", Prompt: "Name of the job"},
		},
		Flags: flags,
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, prefix string) ([]string, string) {
			return events.CompleteWorkloadName(ctx, args, prefix, gql.SearchTypeJob, flags.Team, string(flags.Environment))
		},
		ValidateFlagsFunc: validation.RequireEnvironmentFlag(flags),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			query, err := events.ParseEventQuery(flags.Limit, flags.From, flags.To, flags.After)
			if err != nil {
				return err
			}
			page, err := workload.GetEvents(ctx, workload.JobEvents, flags.Team, string(flags.Environment), args.Get("name"), query)
			if err != nil {
				return err
			}
			return events.RenderEvents(out, page, query, flags.Output == "json")
		},
	}
}
