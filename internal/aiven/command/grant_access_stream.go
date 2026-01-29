package command

import (
	"context"

	"github.com/nais/cli/internal/aiven"
	"github.com/nais/cli/internal/aiven/command/flag"
	"github.com/nais/naistrix"
)

func grantAccessStream(parentFlags *flag.GrantAccess) *naistrix.Command {
	grantAccessStreamFlags := &flag.GrantAccessCommon{GrantAccess: parentFlags}

	return &naistrix.Command{
		Name:  "stream",
		Title: "Grant a user's service-user access to a Kafka Stream.",
		Flags: grantAccessStreamFlags,
		Args: []naistrix.Argument{
			{Name: "username"},
			{Name: "stream"},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			namespace := args.Get("namespace")
			userName := args.Get("username")
			stream := args.Get("stream")

			accessResult, err := aiven.GrantAccessToStream(ctx, namespace, stream, userName)
			if err != nil {
				return err
			}

			if !accessResult.Added {
				out.Printf("Username '%s' already listed in Stream '%s/%s' ACLs.", userName, namespace, stream)
				return nil
			}

			out.Printf("Username '%s' added to Stream '%s/%s' ACLs.", userName, namespace, stream)
			return nil
		},
	}
}
