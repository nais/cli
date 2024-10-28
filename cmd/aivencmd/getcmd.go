package aivencmd

import (
	ctx "context"
	"fmt"

	"github.com/nais/cli/pkg/aiven"
	"github.com/nais/cli/pkg/aiven/aiven_services"
	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/otel"
)

func getCommand() *cli.Command {

	return &cli.Command{
		Name:      "get",
		Usage:     "Generate preferred config format to '/tmp' folder",
		ArgsUsage: "service username namespace",
		Before: func(context *cli.Context) error {
			myMetric := otel.GetMeterProvider()
			counter, err := myMetric.Meter("Aiven").Int64Counter("Aiven-get")
			if err != nil {
				return fmt.Errorf("metrics provider")
			}
			counter.Add(ctx.Background(), 1)

			if context.Args().Len() < 3 {
				return fmt.Errorf("missing required arguments: service, secret, namespace")
			}

			_, err = aiven_services.FromString(context.Args().Get(0))
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
				return fmt.Errorf("retrieve secret and generating config: %w", err)
			}

			return nil
		},
	}
}
