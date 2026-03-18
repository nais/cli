package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/kafka"
	"github.com/nais/cli/internal/kafka/command/flag"
	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func list(parentFlags *flag.Kafka) *naistrix.Command {
	flags := &flag.List{Kafka: parentFlags}

	return &naistrix.Command{
		Name:  "list",
		Title: "List Kafka topics in a team.",
		Flags: flags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			ret, err := kafka.GetTeamTopics(ctx, flags.Team, flags.Environment)
			if err != nil {
				return err
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(ret)
			}

			if len(ret) == 0 {
				out.Println("Team has no Kafka topics.")
				return nil
			}

			user, err := naisapi.GetAuthenticatedUser(ctx)
			if err != nil {
				return err
			}

			type entry struct {
				Name        output.Link `json:"name"`
				Environment string      `json:"environment"`
			}

			entries := make([]entry, 0, len(ret))
			for _, topic := range ret {
				entries = append(entries, entry{
					Name: output.Link{
						Name: topic.Name,
						URL: fmt.Sprintf(
							"https://%s/team/%s/%s/kafka/%s",
							user.ConsoleHost(),
							flags.Team,
							topic.Environment,
							topic.Name,
						),
					},
					Environment: topic.Environment,
				})
			}

			return out.Table().Render(entries)
		},
	}
}
