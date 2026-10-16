package command

import (
	"context"
	"fmt"
	"sort"

	"github.com/nais/cli/internal/kafka"
	"github.com/nais/cli/internal/kafka/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
)

func listGrants(parentFlags *flag.Kafka) *naistrix.Command {
	flags := &flag.ListGrants{Kafka: parentFlags}

	return &naistrix.Command{
		Name:        "list-grants",
		Title:       "List access grants for a Kafka topic.",
		Description: "Shows the workloads that have been granted access to a Kafka topic.",
		Flags:       flags,
		Args: []naistrix.Argument{
			{
				Name:   "topic",
				Prompt: "Name of the Kafka topic to list grants for",
				ChoicesFunc: func(ctx context.Context, _ *naistrix.Arguments) ([]string, error) {
					topics, err := kafka.GetTeamTopics(ctx, flags.Team, string(flags.Environment), nil)
					if err != nil {
						return nil, fmt.Errorf("unable to fetch Kafka topics: %w", err)
					}

					names := make([]string, len(topics))
					for i, topic := range topics {
						names[i] = topic.Name
					}
					sort.Strings(names)
					return names, nil
				},
			},
		},
		ValidateFlagsFunc: validation.RequireTeamAndEnvironmentFlags(flags),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			grants, err := kafka.GetKafkaTopicGrants(ctx, args.Get("topic"), flags.Team, flags.Environment)
			if err != nil {
				return naistrix.Errorf("Unable to list grants: %s", err)
			}

			if flags.Output == "json" {
				return out.JSON(output.JSONWithPrettyOutput()).Render(grants)
			}

			if len(grants) == 0 {
				out.Println("Kafka topic has no access grants.")
				return nil
			}

			return out.Table().Render(grants)
		},
	}
}
