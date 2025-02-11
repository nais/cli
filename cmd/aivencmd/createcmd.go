package aivencmd

import (
	"fmt"
	"strings"

	"github.com/nais/cli/pkg/aiven"
	"github.com/nais/cli/pkg/aiven/aiven_services"
	"github.com/nais/cli/pkg/k8s"
	"github.com/nais/cli/pkg/metrics"
	"github.com/urfave/cli/v2"
)

func createCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		Usage:     "Creates a protected and time-limited AivenApplication",
		ArgsUsage: "service username namespace",
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:    "expire",
				Aliases: []string{"e"},
				Value:   1,
			},
			&cli.StringFlag{
				Name:    "pool",
				Aliases: []string{"p"},
				Value:   "nav-dev",
				Action: func(context *cli.Context, flag string) error {
					service, err := aiven_services.FromString(context.Args().Get(0))
					if err != nil {
						return err
					}

					if !service.Is(&aiven_services.Kafka{}) {
						metrics.AddOne("aiven_create_pool_error_total")
						return fmt.Errorf("--pool is only supported for Kafka, not %v", service.Name())
					}

					return nil
				},
			},
			&cli.StringFlag{
				Name:    "secret",
				Aliases: []string{"s"},
			},
			&cli.StringFlag{
				Name:    "instance",
				Aliases: []string{"i"},
				Action: func(context *cli.Context, flag string) error {
					service, err := aiven_services.FromString(context.Args().Get(0))
					if err != nil {
						return err
					}

					if !service.Is(&aiven_services.OpenSearch{}) {
						metrics.AddOne("aiven_create_instance_error_total")
						return fmt.Errorf("--instance is only supported for OpenSearch, not %v", service.Name())
					}

					return nil
				},
			},
			&cli.StringFlag{
				Name:    "access",
				Aliases: []string{"a"},
				Action: func(context *cli.Context, flag string) error {
					service, err := aiven_services.FromString(context.Args().Get(0))
					if err != nil {
						return err
					}

					if !service.Is(&aiven_services.OpenSearch{}) {
						metrics.AddOne("aiven_create_access_error_total")
						return fmt.Errorf("--access is only supported for OpenSearch, not %v", service.Name())
					}

					return nil
				},
			},
		},
		Before: func(context *cli.Context) error {
			if context.Args().Len() < 3 {
				metrics.AddOne("aiven_create_required_args_error_total")
				return fmt.Errorf("missing required arguments: service, username, namespace")
			}

			_, err := aiven_services.FromString(context.Args().Get(0))
			if err != nil {
				return err
			}

			return nil
		},
		Action: func(context *cli.Context) error {
			service, err := aiven_services.FromString(context.Args().Get(0))
			if err != nil {
				return err
			}

			username := context.Args().Get(1)
			namespace := context.Args().Get(2)

			expire := context.Uint("expire")
			secretName := context.String("secret")
			instance := context.String("instance")

			pool, err := aiven_services.KafkaPoolFromString(context.String("pool"))
			if err != nil {
				metrics.AddOne("aiven_create_pool_values_error_total")
				return fmt.Errorf("valid values for pool: %v", strings.Join(aiven_services.KafkaPools, ", "))
			}

			access, err := aiven_services.OpenSearchAccessFromString(context.String("access"))
			if err != nil && service.Is(&aiven_services.OpenSearch{}) {
				metrics.AddOne("aiven_create_access_value_error_total")
				return fmt.Errorf("valid values for access: %v", strings.Join(aiven_services.OpenSearchAccesses, ", "))
			}

			aivenConfig := aiven.Setup(k8s.SetupControllerRuntimeClient(), service, username, namespace, secretName, instance, pool, access, expire)
			aivenApp, err := aivenConfig.GenerateApplication()
			if err != nil {
				metrics.AddOne("aiven_create_generating_aivenapplication_error_total")
				return fmt.Errorf("an error occurred generating 'AivenApplication': %v", err)
			}

			fmt.Printf("Use the following command to generate configuration secrets:\n\tnais aiven get %v %v %v\n", service.Name(), aivenApp.Spec.SecretName, aivenApp.Namespace)

			return nil
		},
	}
}
