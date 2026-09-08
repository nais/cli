package command

import (
	"context"
	"strings"

	"github.com/nais/cli/internal/kafka"
	"github.com/nais/cli/internal/kafka/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
)

func grantAccess(parentFlags *flag.Kafka) *naistrix.Command {
	grantAccessTopicFlags := &flag.GrantAccess{
		Kafka:  parentFlags,
		Access: "read",
	}
	return &naistrix.Command{
		Name:        "grant-access",
		Title:       "Grant a user's service-user access to a Kafka Topic.",
		Description: "It adds an ACL entry for a user on a Kafka Topic with the specified access level.",
		Flags:       grantAccessTopicFlags,
		Args: []naistrix.Argument{
			{Name: "username"},
			{Name: "topic"},
		},
		ValidateFunc: naistrix.ValidateFuncs(
			validation.RequireTeamAndEnvironment(grantAccessTopicFlags),
			func(context.Context, *naistrix.Arguments) error {
				return grantAccessTopicFlags.Access.Validate()
			},
		),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			topicName := args.Get("topic")
			subject := kafkaApplicationName(args.Get("username"))
			grant := gql.KafkaTopicGrantInput{
				Subject:  subject,
				TeamName: grantAccessTopicFlags.Team,
				Access:   gql.KafkaTopicGrantAccess(strings.ToUpper(string(grantAccessTopicFlags.Access))),
			}

			if err := kafka.GrantAccessToKafkaTopic(ctx, topicName, grantAccessTopicFlags.Team, grantAccessTopicFlags.Environment, grant); err != nil {
				return naistrix.Errorf("Unable to grant access: %s", err)
			}

			out.Printf(
				"ACL added for %q, with access %q on topic \"%s/%s\".\n",
				subject, grantAccessTopicFlags.Access, grantAccessTopicFlags.Team, topicName,
			)
			return nil
		},
	}
}
