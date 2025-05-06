package create

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/internal/aiven"
	"github.com/nais/cli/internal/aiven/aiven_services"
	"github.com/nais/cli/internal/k8s"
	"github.com/nais/cli/internal/metrics"
	"github.com/urfave/cli/v3"
)

type Flags struct {
	Expire   uint
	Pool     string
	Access   string
	Secret   string
	Instance string
	aiven.Flags
}

type Args struct {
	Service   string
	Username  string
	Namespace string
}

func Validate(ctx context.Context, args Args, flags Flags) error {
	_, err := aiven_services.FromString(args.Service)
	if err != nil {
		return err
	}

	return nil
}

func Action(ctx context.Context, args Args, flags Flags) error {
	service, err := aiven_services.FromString(args.Service)
	if err != nil {
		return err
	}

	if flags.Expire > 30 {
		return fmt.Errorf("expire must be less than %v days", 30)
	}

	pool, err := aiven_services.KafkaPoolFromString(flags.Pool)
	if err != nil {
		metrics.AddOne(ctx, "aiven_create_pool_values_error_total")
		return fmt.Errorf("valid values for pool should specify tenant and environment separated by a dash (-): %v", err)
	}

	access, err := aiven_services.OpenSearchAccessFromString(flags.Access)
	if err != nil && service.Is(&aiven_services.OpenSearch{}) {
		metrics.AddOne(ctx, "aiven_create_access_value_error_total")
		return fmt.Errorf("valid values for access: %v", strings.Join(aiven_services.OpenSearchAccesses, ", "))
	}

	aivenConfig := aiven.Setup(ctx, k8s.SetupControllerRuntimeClient(), service, args.Username, args.Namespace, flags.Secret, flags.Instance, pool, access, flags.Expire)
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
