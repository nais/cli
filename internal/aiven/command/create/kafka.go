package create

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/aiven"
	"github.com/nais/cli/internal/aiven/aiven_services"
	"github.com/nais/cli/internal/aiven/flag"
	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/k8s"
	"github.com/nais/cli/internal/metric"
	"github.com/nais/cli/internal/output"
)

func createKafka(parentFlags *flag.Create) *cli.Command {
	createKafkaFlags := &flag.CreateKafka{Create: parentFlags, Pool: "nav-dev"}

	return cli.NewCommand("kafka", "Grant a user access to a Kafka topic.",
		cli.WithFlag("pool", "p", "The `NAME` of the pool to create the Kafka instance in.", &createKafkaFlags.Pool),
		cli.WithFlag("test", "t", "Create a test Kafka topic with the given `NAME`.", &createKafkaFlags.Test),
		cli.WithArgs("username", "namespace"),
		cli.WithValidate(cli.ValidateExactArgs(2)),
		cli.WithRun(func(ctx context.Context, output output.Output, args []string) error {
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
			aivenApp, err := aivenConfig.GenerateApplication()
			if err != nil {
				metric.CreateAndIncreaseCounter(ctx, "aiven_create_generating_aivenapplication_error_total")
				return fmt.Errorf("an error occurred generating 'AivenApplication': %v", err)

			}

			fmt.Printf("Use the following command to generate configuration secrets:\n\tnais aiven get %v %v %v\n", service.Name(), aivenApp.Spec.SecretName, aivenApp.Namespace)

			return nil
		}),
	)
}
