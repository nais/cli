package command

import (
	"context"
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
			{Name: "topic", Prompt: "Name of the Kafka topic to list grants for"},
		},
		ValidateFlagsFunc: validation.RequireTeamAndEnvironmentFlags(flags),
		AutoCompleteFunc:  autoCompleteKafkaTopicName(flags.Kafka, 0),
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

func autoCompleteKafkaTopicName(flags *flag.Kafka, topicArgumentIndex int) naistrix.AutoCompleteFunc {
	return func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
		if args.Len() != topicArgumentIndex {
			return nil, ""
		}
		topics, err := kafka.GetTeamTopics(ctx, flags.Team, string(flags.Environment), nil)
		if err != nil {
			return nil, "Unable to fetch Kafka topics."
		}

		names := make([]string, 0, len(topics))
		for _, topic := range topics {
			names = append(names, topic.Name)
		}
		sort.Strings(names)

		if len(names) == 0 {
			return nil, "No Kafka topics found in the selected environment."
		}

		return names, "Select a Kafka topic."
	}
}
