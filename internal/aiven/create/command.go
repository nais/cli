package create

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/nais/cli/internal/aiven"
	"github.com/nais/cli/internal/aiven/aiven_services"
	"github.com/nais/cli/internal/k8s"
	"github.com/nais/cli/internal/metrics"
	"github.com/urfave/cli/v3"
)

func Before(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	if cmd.Args().Len() < 3 {
		metrics.AddOne(ctx, "aiven_create_required_args_error_total")
		return ctx, fmt.Errorf("missing required arguments: service, username, namespace")
	}

	if _, err := aiven_services.FromString(cmd.Args().Get(0)); err != nil {
		return ctx, err
	}

	return ctx, nil
}

func Action(ctx context.Context, cmd *cli.Command) error {
	service, err := aiven_services.FromString(cmd.Args().Get(0))
	if err != nil {
		return err
	}

	username := cmd.Args().Get(1)
	namespace := cmd.Args().Get(2)

	expire := cmd.Uint("expire")
	if expire > uint(math.MaxInt) {
		return fmt.Errorf("--expire must be less than %v", math.MaxInt)
	}
	secretName := cmd.String("secret")
	instance := cmd.String("instance")

	pool, err := aiven_services.KafkaPoolFromString(cmd.String("pool"))
	if err != nil {
		metrics.AddOne(ctx, "aiven_create_pool_values_error_total")
		return fmt.Errorf("valid values for pool should specify tenant and environment separated by a dash (-): %v", err)
	}

	access, err := aiven_services.OpenSearchAccessFromString(cmd.String("access"))
	if err != nil && service.Is(&aiven_services.OpenSearch{}) {
		metrics.AddOne(ctx, "aiven_create_access_value_error_total")
		return fmt.Errorf("valid values for access: %v", strings.Join(aiven_services.OpenSearchAccesses, ", "))
	}

	aivenConfig := aiven.Setup(ctx, k8s.SetupControllerRuntimeClient(), service, username, namespace, secretName, instance, pool, access, expire)
	aivenApp, err := aivenConfig.GenerateApplication()
	if err != nil {
		metrics.AddOne(ctx, "aiven_create_generating_aivenapplication_error_total")
		return fmt.Errorf("an error occurred generating 'AivenApplication': %v", err)
	}

	fmt.Printf("Use the following command to generate configuration secrets:\n\tnais aiven get %v %v %v\n", service.Name(), aivenApp.Spec.SecretName, aivenApp.Namespace)

	return nil
}

func PoolFlagAction(ctx context.Context, cmd *cli.Command, flag string) error {
	service, err := aiven_services.FromString(cmd.Args().Get(0))
	if err != nil {
		return err
	}

	if !service.Is(&aiven_services.Kafka{}) {
		metrics.AddOne(ctx, "aiven_create_pool_error_total")
		return fmt.Errorf("--pool is only supported for Kafka, not %v", service.Name())
	}

	return nil
}

func InstanceFlagAction(ctx context.Context, cmd *cli.Command, flag string) error {
	service, err := aiven_services.FromString(cmd.Args().Get(0))
	if err != nil {
		return err
	}

	if !service.Is(&aiven_services.OpenSearch{}) {
		metrics.AddOne(ctx, "aiven_create_instance_error_total")
		return fmt.Errorf("--instance is only supported for OpenSearch, not %v", service.Name())
	}

	return nil
}

func AccessFlagAction(ctx context.Context, cmd *cli.Command, flag string) error {
	service, err := aiven_services.FromString(cmd.Args().Get(0))
	if err != nil {
		return err
	}

	if !service.Is(&aiven_services.OpenSearch{}) {
		metrics.AddOne(ctx, "aiven_create_access_error_total")
		return fmt.Errorf("--access is only supported for OpenSearch, not %v", service.Name())
	}

	return nil
}
