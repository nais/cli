package command

import (
	"context"
	"sort"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/kafka"
	"github.com/nais/cli/internal/kafka/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func Kafka(parentFlags *flags.GlobalFlags) *naistrix.Command {
	flags := &flag.Kafka{GlobalFlags: parentFlags}
	return &naistrix.Command{
		Name:         "kafka",
		Aliases:      []string{"kafkas"},
		Title:        "Interact with Kafka topics.",
		Description:  "Commands for managing Kafka topics and credentials for your team.",
		StickyFlags:  flags,
		ValidateFunc: validation.RequireTeam(flags),
		SubCommands: []*naistrix.Command{
			credentials(flags),
			grantAccess(flags),
			listGrants(flags),
			list(flags),
			revokeGrant(flags),
		},
	}
}

func autoCompleteKafkaTopicName(flags *flag.Kafka, topicArgumentIndex int) naistrix.AutoCompleteFunc {
	return func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
		if args.Len() != topicArgumentIndex {
			return nil, ""
		}
		if flags.Team == "" {
			return nil, "Please provide team to auto-complete Kafka topic names. 'nais defaults set team <team>', or '--team <team>' flag."
		}
		if flags.Environment == "" {
			return nil, "Please provide environment to auto-complete Kafka topic names. '-e, --environment <environment>' flag."
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
