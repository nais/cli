package command

import (
	"context"

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
			{Name: "topic"},
		},
		ValidateFunc:     validation.RequireTeamAndEnvironment(flags),
		AutoCompleteFunc: autoCompleteKafkaTopicName(flags.Kafka, 0),
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
