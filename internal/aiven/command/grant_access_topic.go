package command

import (
	"context"

	"github.com/nais/cli/internal/aiven"
	"github.com/nais/cli/internal/aiven/command/flag"
	nais_kafka "github.com/nais/liberator/pkg/apis/kafka.nais.io/v1"
	"github.com/nais/naistrix"
)

func grantAccessTopic(parentFlags *flag.GrantAccess) *naistrix.Command {
	grantAccessTopicFlags := &flag.GrantAccessCommon{GrantAccess: parentFlags}

	return &naistrix.Command{
		Name:  "topic",
		Title: "Grant a user's service-user access to a Kafka Topic.",
		Flags: grantAccessTopicFlags,
		Args: []naistrix.Argument{
			{Name: "username"},
			{Name: "topic"},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			namespace := args.Get("namespace")
			topicName := args.Get("topic")

			acl := nais_kafka.TopicACL{
				Team:        namespace,
				Application: args.Get("username"),
				Access:      args.Get("access"),
			}

			accessResult, err := aiven.GrantAccessToTopic(ctx, namespace, topicName, acl)
			if err != nil {
				return err
			}

			if !accessResult.Added {
				out.Printf("ACL already exists for team '%s', application '%s', access '%s' on topic '%s/%s'.",
					acl.Team, acl.Application, acl.Access, namespace, topicName,
				)
				return nil
			}

			out.Printf("ACL added for team '%s', application '%s', access '%s' on topic '%s/%s'.",
				acl.Team, acl.Application, acl.Access, namespace, topicName,
			)
			return nil
		},
	}
}
