package command

import (
	"context"

	"github.com/nais/cli/internal/aiven"
	"github.com/nais/cli/internal/aiven/command/flag"
	nais_kafka "github.com/nais/liberator/pkg/apis/kafka.nais.io/v1"
	"github.com/nais/naistrix"
)

func grantAccessTopic(parentFlags *flag.GrantAccess) *naistrix.Command {
	grantAccessTopicFlags := &flag.GrantAccessTopic{GrantAccess: parentFlags, Access: "read"}

	return &naistrix.Command{
		Name:        "topic",
		Title:       "Grant a user's service-user access to a Kafka Topic.",
		Description: "This command is deprecated. Use 'nais kafka credentials' instead. It adds an ACL entry for a user on a Kafka Topic with the specified access level.",
		Flags:       grantAccessTopicFlags,
		Deprecated:  naistrix.DeprecatedWithReplacement([]string{"kafka", "credentials"}),
		Args: []naistrix.Argument{
			{Name: "username"},
			{Name: "topic"},
		},
		ValidateFunc: func(context.Context, *naistrix.Arguments) error {
			if err := grantAccessTopicFlags.UsesRemovedFlags(); err != nil {
				return err
			}
			_, err := grantAccessTopicFlags.RequiredTeam()
			return err
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			access := grantAccessTopicFlags.Access
			namespace := grantAccessTopicFlags.Team
			topicName := args.Get("topic")
			username := args.Get("username")

			if err := aiven.ValidAclPermission(access); err != nil {
				return err
			}

			newAcl := nais_kafka.TopicACL{
				Team:        namespace,
				Application: username,
				Access:      access,
			}
			accessResult, err := aiven.GrantAccessToTopic(ctx, namespace, topicName, string(grantAccessTopicFlags.Environment), newAcl)
			if err != nil {
				return err
			}

			if accessResult.AlreadyAdded {
				out.Printf("ACL entry already exists for '%s/%s' on topic %s/%s.",
					newAcl.Application, newAcl.Access, namespace, topicName,
				)
				return nil
			}

			out.Printf("ACL added for '%s', with access '%s' on topic '%s/%s'.",
				newAcl.Application, newAcl.Access, namespace, topicName,
			)
			return nil
		},
	}
}
