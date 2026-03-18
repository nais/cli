package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/job"
	"github.com/nais/cli/internal/job/command/flag"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func list(parentFlags *flag.Job) *naistrix.Command {
	flags := &flag.List{Job: parentFlags}

	return &naistrix.Command{
		Name:  "list",
		Title: "List jobs in a team.",
		Flags: flags,
		AutoCompleteFunc: func(_ context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() == 0 {
				if flags.Team == "" {
					return nil, "Please provide team. 'nais config set team <team>', or '--team <team>' flag."
				}
			}
			return nil, ""
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			ret, err := job.GetTeamJobs(ctx, flags.Team, flags.Environment)
			if err != nil {
				return err
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(ret)
			}

			if len(ret) == 0 {
				out.Println("Team has no jobs.")
				return nil
			}

			user, err := naisapi.GetAuthenticatedUser(ctx)
			if err != nil {
				return err
			}

			type entry struct {
				Name        output.Link      `json:"name"`
				Environment string           `json:"environment"`
				Schedule    job.Schedule     `json:"schedule"`
				LastRun     job.LastRunState `heading:"Last Run" json:"last_run"`
				State       job.State        `json:"state"`
				Issues      int              `json:"issues"`
			}

			entries := make([]entry, 0, len(ret))
			for _, j := range ret {
				entries = append(entries, entry{
					Name: output.Link{
						Name: j.Name,
						URL: fmt.Sprintf(
							"https://%s/team/%s/%s/job/%s",
							user.ConsoleHost(),
							flags.Team,
							j.Environment,
							j.Name,
						),
					},
					Environment: j.Environment,
					Schedule:    j.Schedule,
					LastRun:     j.LastRun,
					State:       j.State,
					Issues:      j.Issues,
				})
			}

			return out.Table().Render(entries)
		},
	}
}
