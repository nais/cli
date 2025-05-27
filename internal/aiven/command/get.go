package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/aiven"
	"github.com/nais/cli/internal/aiven/aiven_services"
	"github.com/nais/cli/internal/aiven/command/flag"
	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/metric"
	"github.com/nais/cli/internal/output"
)

func get(_ *flag.Aiven) *cli.Command {
	return cli.NewCommand("get", "Generate preferred config format to '/tmp' folder.",
		cli.WithValidate(cli.ValidateExactArgs(3)),
		cli.WithArgs("service", "username", "namespace"),
		cli.WithAutoComplete(func(ctx context.Context, args []string, toComplete string) ([]string, string) {
			if len(args) == 0 {
				return []string{"kafka", "opensearch"}, "Choose the service you want to get."
			}
			return nil, ""
		}),
		cli.WithRun(func(ctx context.Context, out output.Output, args []string) error {
			service, err := aiven_services.FromString(args[0])
			if err != nil {
				return err
			}

			if err := aiven.ExtractAndGenerateConfig(ctx, service, args[1], args[2]); err != nil {
				metric.CreateAndIncreaseCounter(ctx, "aiven_get_secret_and_config_error_total")
				return fmt.Errorf("retrieve secret and generating config: %w", err)
			}
			return nil
		}),
	)
}
