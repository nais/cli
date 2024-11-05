package aivencmd

import (
	"fmt"

	"github.com/nais/cli/pkg/metrics"

	"github.com/nais/cli/pkg/aiven"
	"github.com/nais/cli/pkg/aiven/aiven_services"
	"github.com/urfave/cli/v2"
)

func getCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Generate preferred config format to '/tmp' folder",
		ArgsUsage: "service username namespace",
		Before: func(context *cli.Context) error {
			if context.Args().Len() < 3 {
				metrics.AddOne("aiven_get_arguments_error_total")

				return fmt.Errorf("missing required arguments: service, secret, namespace")
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

			secretName := context.Args().Get(1)
			namespace := context.Args().Get(2)

			err = aiven.ExtractAndGenerateConfig(service, secretName, namespace)
			if err != nil {
				metrics.AddOne("aiven_get_secret_and_config_error_total")
				return fmt.Errorf("retrieve secret and generating config: %w", err)
			}

			return nil
		},
	}
}
