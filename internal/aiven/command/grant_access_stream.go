package command

import (
	"context"

	"github.com/nais/cli/internal/aiven"
	"github.com/nais/cli/internal/aiven/command/flag"
	"github.com/nais/naistrix"
)

func grantAccessStream(parentFlags *flag.GrantAccess) *naistrix.Command {
	grantAccessStreamFlags := &flag.GrantAccessStream{GrantAccess: parentFlags}

	return &naistrix.Command{
		Name:  "stream",
		Title: "Grant a user's service-user access to a Kafka Stream.",
		Flags: grantAccessStreamFlags,
		Args: []naistrix.Argument{
			{Name: "username"},
			{Name: "stream"},
		},
		ValidateFunc: func(context.Context, *naistrix.Arguments) error {
			if err := grantAccessStreamFlags.UsesRemovedFlags(); err != nil {
				return err
			}
			_, err := grantAccessStreamFlags.RequiredTeam()
			return err
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			namespace := grantAccessStreamFlags.Team
			userName := args.Get("username")
			stream := args.Get("stream")

			accessResult, err := aiven.GrantAccessToStream(ctx, namespace, stream, userName, string(grantAccessStreamFlags.Environment))
			if err != nil {
				return err
			}

			if accessResult.AlreadyAdded {
				out.Printf("Username '%s' already exists in Stream '%s/%s' ACLs.", userName, namespace, stream)
				return nil
			}

			out.Printf("Username '%s' added to Stream '%s/%s' ACLs.", userName, namespace, stream)
			return nil
		},
	}
}
