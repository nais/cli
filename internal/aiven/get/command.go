package get

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/aiven"
	"github.com/nais/cli/internal/aiven/aiven_services"
	"github.com/nais/cli/internal/metrics"
	"github.com/urfave/cli/v3"
)

func Before(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	if cmd.Args().Len() < 3 {
		metrics.AddOne(ctx, "aiven_get_arguments_error_total")

		return ctx, fmt.Errorf("missing required arguments: service, secret, namespace")
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

	secretName := cmd.Args().Get(1)
	namespace := cmd.Args().Get(2)

	if err = aiven.ExtractAndGenerateConfig(ctx, service, secretName, namespace); err != nil {
		metrics.AddOne(ctx, "aiven_get_secret_and_config_error_total")
		return fmt.Errorf("retrieve secret and generating config: %w", err)
	}

	return nil
}
