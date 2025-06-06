package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/aiven"
	"github.com/nais/cli/internal/aiven/aiven_services"
	"github.com/nais/cli/internal/aiven/command/flag"
	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/k8s"
	"github.com/nais/cli/internal/metric"
)

func createKafka(parentFlags *flag.Create) *cli.Command {
	createKafkaFlags := &flag.CreateKafka{Create: parentFlags, Pool: "nav-dev"}

	return &cli.Command{
		Name:  "kafka",
		Short: "Grant a user access to a Kafka topic.",
		Flags: createKafkaFlags,
		Args: []cli.Argument{
			{Name: "username", Required: true},
			{Name: "namespace", Required: true},
		},
		ValidateFunc: cli.ValidateExactArgs(2),
		RunFunc: func(ctx context.Context, out cli.Output, args []string) error {
			pool, err := aiven_services.KafkaPoolFromString(createKafkaFlags.Pool)
			if err != nil {
				return fmt.Errorf("valid values for pool should specify tenant and environment separated by a dash (-): %v", err)
			}

			service := &aiven_services.Kafka{}

			aivenConfig := aiven.Setup(
				ctx,
				k8s.SetupControllerRuntimeClient(),
				service,
				args[0],
				args[1],
				createKafkaFlags.Secret,
				createKafkaFlags.Expire,
				&aiven_services.ServiceSetup{
					Pool: pool,
				},
			)
			aivenApp, err := aivenConfig.GenerateApplication(out)
			if err != nil {
				metric.CreateAndIncreaseCounter(ctx, "aiven_create_generating_aivenapplication_error_total")
				return fmt.Errorf("an error occurred generating 'AivenApplication': %v", err)

			}

			out.Printf("Use the following command to generate configuration secrets:\n\tnais aiven get %v %v %v\n", service.Name(), aivenApp.Spec.SecretName, aivenApp.Namespace)

			return nil
		},
	}
}
