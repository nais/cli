package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/internal/aiven"
	"github.com/nais/cli/internal/aiven/aiven_services"
	"github.com/nais/cli/internal/aiven/command/flag"
	"github.com/nais/cli/internal/k8s"
	"github.com/nais/cli/internal/metric"
	"github.com/nais/naistrix"
)

func createStream(parentFlags *flag.Create) *naistrix.Command {
	createStreamFlags := &flag.CreateStream{Create: parentFlags, Pool: "nav-dev"}

	return &naistrix.Command{
		Name:  "stream",
		Title: "Grant a user access to a Kafka Stream.",
		Flags: createStreamFlags,
		Args: []naistrix.Argument{
			{Name: "stream-name"},
			{Name: "username"},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			pool, err := aiven_services.KafkaPoolFromString(createStreamFlags.Pool)
			if err != nil {
				return fmt.Errorf("valid values for pool should specify tenant and environment separated by a dash (-): %v", err)
			}

			streamSplit := strings.Split(args.Get("stream-name"), ".")
			if len(streamSplit) < 2 {
				return fmt.Errorf("Name of stream does not follow expected `<namespace>.<app_name>_stream_*` prefix: %v", args.Get("stream-name"))
			}
			namespace := streamSplit[0]

			if createStreamFlags.Secret == "" {
				if createStreamFlags.Secret, err = aiven.CreateSecretName(args.Get("username"), namespace); err != nil {
					return fmt.Errorf("creating secret name: %v", err)
				}
			}

			// service := &aiven_services.Kafka{}
			// TODO: get the Stream's CRD, and add username to `additionalUsers` list
			client := k8s.SetupControllerRuntimeClient()
			client.Client.Get(ctx, objectKey, object)

			return nil
		},
	}
}
