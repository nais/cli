package command

import (
	"context"
	"errors"
	"fmt"

	"github.com/nais/cli/pkg/cli/v2"
	"github.com/nais/cli/v2/internal/aiven"
	"github.com/nais/cli/v2/internal/aiven/aiven_services"
	"github.com/nais/cli/v2/internal/aiven/command/flag"
	"github.com/nais/cli/v2/internal/metric"
)

func get(_ *flag.Aiven) *cli.Command {
	return &cli.Command{
		Name:  "get",
		Title: "Generate preferred config format to '/tmp' folder.",
		Args: []cli.Argument{
			{Name: "service"},
			{Name: "username"},
			{Name: "namespace"},
		},
		AutoCompleteFunc: func(ctx context.Context, args []string, toComplete string) ([]string, string) {
			if len(args) == 0 {
				return []string{"kafka", "opensearch"}, "Choose the service you want to get."
			}
			return nil, ""
		},
		RunFunc: func(ctx context.Context, out cli.Output, args []string) error {
			service, err := aiven_services.FromString(args[0])
			if err != nil {
				return err
			}

			username := args[1]
			namespace := args[2]
			if err := aiven.ExtractAndGenerateConfig(ctx, service, username, namespace, out); err != nil {
				metric.CreateAndIncreaseCounter(ctx, "aiven_get_secret_and_config_error_total")

				switch {
				case errors.Is(err, aiven.ErrUnsuitableSecret):
					return cli.Errorf(`The secret we found for username %q is not suitable for this command.
Most likely it was not created by using %v, please refer to %v for instructions on how to create one.
`, username, "`nais aiven create`", fmt.Sprintf("`nais aiven create %s --help`", service.Name()))
				default:
					return fmt.Errorf("retrieve secret and generating config: %w", err)
				}
			}
			return nil
		},
	}
}
