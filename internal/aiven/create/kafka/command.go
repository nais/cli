package kafka

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/aiven"
	"github.com/nais/cli/internal/aiven/aiven_services"
	"github.com/nais/cli/internal/aiven/create"
	"github.com/nais/cli/internal/k8s"
	"github.com/nais/cli/internal/metric"
)

type Flags struct {
	create.Flags
	Pool aiven_services.KafkaPool
}

func Run(ctx context.Context, args create.Arguments, flags Flags) error {
	service := &aiven_services.Kafka{}
	aivenConfig := aiven.Setup(
		ctx,
		k8s.SetupControllerRuntimeClient(),
		service,
		args.Username, args.Namespace, flags.Secret, flags.Expire,
		&aiven_services.ServiceSetup{
			Pool: flags.Pool,
		},
	)
	aivenApp, err := aivenConfig.GenerateApplication()
	if err != nil {
		metric.CreateAndIncreaseCounter(ctx, "aiven_create_generating_aivenapplication_error_total")
		return fmt.Errorf("an error occurred generating 'AivenApplication': %v", err)

	}

	fmt.Printf("Use the following command to generate configuration secrets:\n\tnais aiven get %v %v %v\n", service.Name(), aivenApp.Spec.SecretName, aivenApp.Namespace)

	return nil
}
