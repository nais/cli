package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/aiven/command/flag"
	"github.com/nais/cli/internal/k8s"
	nais_kafka "github.com/nais/liberator/pkg/apis/kafka.nais.io/v1"
	"github.com/nais/naistrix"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

func grantAccessStream(parentFlags *flag.GrantAccess) *naistrix.Command {
	grantAccessStreamFlags := &flag.GrantAccessStream{GrantAccess: parentFlags}

	return &naistrix.Command{
		Name:  "stream",
		Title: "Grant a user's service-user access to a Kafka Stream.",
		Flags: grantAccessStreamFlags,
		Args: []naistrix.Argument{
			{Name: "user-name"},
			{Name: "stream-name"},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			client := k8s.SetupControllerRuntimeClient()

			var namespace v1.Namespace
			if err := client.Get(ctx, ctrl.ObjectKey{Name: args.Get("namespace")}, &namespace); err != nil {
				return fmt.Errorf("validate namespace: %w", err)
			}

			var stream nais_kafka.Stream
			if err := client.Get(ctx, ctrl.ObjectKey{Name: args.Get("stream-name")}, &stream); err != nil {
				return fmt.Errorf("validate stream: %w", err)
			}

			userName := args.Get("username")
			for _, user := range stream.Spec.AdditionalUsers {
				if user.Username == userName {
					out.Printf("Username '%s' already listed in Stream '%s/%s''s ACLs.", userName, &namespace.Name, stream.Name)
					return nil
				}
			}

			stream.Spec.AdditionalUsers = append(stream.Spec.AdditionalUsers, nais_kafka.AdditionalStreamUser{
				Username: userName,
			})
			err := client.Update(ctx, &stream)
			if err != nil {
				return err
			}

			out.Printf("Username '%s' added to Stream '%s/%s' ACLs.", userName, &namespace.Name, stream.Name)

			return nil
		},
	}
}
