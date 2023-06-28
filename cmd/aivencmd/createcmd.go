package aivencmd

import (
	"fmt"
	"strings"

	"github.com/nais/cli/pkg/aiven"
	"github.com/nais/cli/pkg/aiven/client"
	"github.com/nais/cli/pkg/aiven/services"
	"github.com/urfave/cli/v2"
)

func createCommand() *cli.Command {
	return &cli.Command{
		Name:      "create",
		Usage:     "Creates a protected and time-limited AivenApplication",
		ArgsUsage: "service username namespace",
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:  "expire",
				Value: 1,
			},
			&cli.StringFlag{
				Name:  "pool",
				Value: "nav-dev",
				Action: func(context *cli.Context, flag string) error {
					service, err := services.ServiceFromString(context.Args().Get(0))
					if err != nil {
						return err
					}

					if !service.Is(&services.Kafka{}) {
						return fmt.Errorf("--pool is only supported for Kafka, not %v", service.Name())
					}

					return nil
				},
			},
			&cli.StringFlag{
				Name: "secret",
			},
			&cli.StringFlag{
				Name: "instance",
				Action: func(context *cli.Context, flag string) error {
					service, err := services.ServiceFromString(context.Args().Get(0))
					if err != nil {
						return err
					}

					if !service.Is(&services.OpenSearch{}) {
						return fmt.Errorf("--intance is only supported for OpenSearch, not %v", service.Name())
					}

					return nil
				},
			},
			&cli.StringFlag{
				Name: "access",
				Action: func(context *cli.Context, flag string) error {
					service, err := services.ServiceFromString(context.Args().Get(0))
					if err != nil {
						return err
					}

					if !service.Is(&services.OpenSearch{}) {
						return fmt.Errorf("--access is only supported for OpenSearch, not %v", service.Name())
					}

					return nil
				},
			},
		},
		Before: func(context *cli.Context) error {
			if context.Args().Len() != 3 {
				return fmt.Errorf("missing required arguments: service, username, namespace")
			}

			_, err := services.ServiceFromString(context.Args().Get(0))
			if err != nil {
				return err
			}

			return nil
		},
		Action: func(context *cli.Context) error {
			service, err := services.ServiceFromString(context.Args().Get(0))
			if err != nil {
				return err
			}

			username := context.Args().Get(1)
			namespace := context.Args().Get(2)

			expire := context.Uint("expire")
			secretName := context.String("secret")
			instance := context.String("instance")

			pool, err := services.KafkaPoolFromString(context.String("pool"))
			if err != nil {
				return fmt.Errorf("valid values for pool: %v", strings.Join(services.KafkaPools, ", "))
			}

			access, err := services.OpenSearchAccessFromString(context.String("access"))
			if err != nil && service.Is(&services.OpenSearch{}) {
				return fmt.Errorf("valid values for access: %v", strings.Join(services.OpenSearchAccesses, ", "))
			}

			aivenConfig := aiven.Setup(client.SetupClient(), service, username, namespace, secretName, instance, pool, access, expire)
			aivenApp, err := aivenConfig.GenerateApplication()
			if err != nil {
				return fmt.Errorf("an error occurred generating 'AivenApplication': %v", err)
			}

			fmt.Printf("use: 'nais aiven get %v %v %v' to generate configuration secrets\n", service.Name(), aivenApp.Spec.SecretName, aivenApp.Namespace)
			return nil
		},
	}
}
