package command

import (
	"context"
	"sort"
	"strings"

	"github.com/nais/cli/internal/kafka"
	"github.com/nais/cli/internal/kafka/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func revokeGrant(parentFlags *flag.Kafka) *naistrix.Command {
	return &naistrix.Command{
		Name:        "revoke-grant",
		Title:       "Revoke a user's service-user access to a Kafka topic.",
		Description: "Removes an ACL entry for a user on a Kafka topic with the specified access level.",
		Args: []naistrix.Argument{
			{Name: "username"},
			{Name: "topic"},
			{Name: "access"},
		},
		ValidateFunc: naistrix.ValidateFuncs(
			validation.RequireTeamAndEnvironment(parentFlags),
			func(_ context.Context, args *naistrix.Arguments) error {
				return flag.KafkaTopicGrantAccess(args.Get("access")).Validate()
			},
		),
		AutoCompleteFunc: autoCompleteKafkaGrantArguments(parentFlags),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			subject := kafkaApplicationName(args.Get("username"))
			topicName := args.Get("topic")
			access := flag.KafkaTopicGrantAccess(args.Get("access"))
			grant := gql.KafkaTopicGrantInput{
				Subject:  subject,
				TeamName: parentFlags.Team,
				Access:   gql.KafkaTopicGrantAccess(strings.ToUpper(string(access))),
			}

			if err := kafka.RevokeAccessFromKafkaTopic(ctx, topicName, parentFlags.Team, parentFlags.Environment, grant); err != nil {
				return naistrix.Errorf("Unable to revoke grant: %s", err)
			}

			out.Printf(
				"ACL removed for %q, with access %q on topic \"%s/%s\".\n",
				subject, access, parentFlags.Team, topicName,
			)
			return nil
		},
	}
}

func autoCompleteKafkaGrantArguments(flags *flag.Kafka) naistrix.AutoCompleteFunc {
	return func(ctx context.Context, args *naistrix.Arguments, toComplete string) ([]string, string) {
		if args.Len() > 2 {
			return nil, ""
		}

		if flags.Team == "" || flags.Environment == "" {
			return nil, "Please provide team and environment to auto-complete Kafka grants."
		}

		grants, err := kafka.GetTeamKafkaTopicGrants(ctx, flags.Team, flags.Environment)
		if err != nil {
			return nil, "Unable to fetch Kafka grants."
		}

		if args.Len() == 0 {
			subjects := make([]string, 0, len(grants))
			seen := make(map[string]struct{})
			for _, grant := range grants {
				if _, ok := seen[grant.WorkloadName]; ok {
					continue
				}
				seen[grant.WorkloadName] = struct{}{}
				subjects = append(subjects, grant.WorkloadName)
			}
			sort.Strings(subjects)
			if len(subjects) == 0 {
				return nil, "No Kafka grants found in the selected environment."
			}
			return subjects, "Select a subject with a Kafka grant."
		}

		subject := kafkaApplicationName(args.Get("username"))
		if args.Len() == 1 {
			topics := make([]string, 0, len(grants))
			seen := make(map[string]struct{})
			for _, grant := range grants {
				if grant.WorkloadName != subject {
					continue
				}
				if _, ok := seen[grant.TopicName]; ok {
					continue
				}
				seen[grant.TopicName] = struct{}{}
				topics = append(topics, grant.TopicName)
			}
			sort.Strings(topics)
			if len(topics) == 0 {
				return nil, "No Kafka grants found for this subject."
			}
			return topics, "Select a Kafka topic with a grant for this subject."
		}

		accesses := make([]string, 0, len(grants))
		for _, grant := range grants {
			if grant.WorkloadName == subject && grant.TopicName == args.Get("topic") {
				accesses = append(accesses, strings.ToLower(grant.Access))
			}
		}
		sort.Strings(accesses)
		if len(accesses) == 0 {
			return nil, "No access grants found for this subject on the Kafka topic."
		}

		return accesses, "Select an access level to revoke."
	}
}
