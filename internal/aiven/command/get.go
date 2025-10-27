package command

import (
	"context"
	"errors"
	"fmt"

	"github.com/nais/cli/internal/aiven"
	"github.com/nais/cli/internal/aiven/aiven_services"
	"github.com/nais/cli/internal/aiven/command/flag"
	"github.com/nais/cli/internal/metric"
	"github.com/nais/naistrix"
)

func get(_ *flag.Aiven) *naistrix.Command {
	return &naistrix.Command{
		Name:  "get",
		Title: "Generate preferred config format to '/tmp' folder.",
		Args: []naistrix.Argument{
			{Name: "service"},
			{Name: "username"},
			{Name: "namespace"},
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, toComplete string) ([]string, string) {
			if args.Len() == 0 {
				return []string{"kafka", "opensearch"}, "Choose the service you want to get."
			}
			return nil, ""
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			service, err := aiven_services.FromString(args.Get("service"))
			if err != nil {
				return err
			}

			username := args.Get("username")
			namespace := args.Get("namespace")
			if err := aiven.ExtractAndGenerateConfig(ctx, service, username, namespace, out); err != nil {
				metric.CreateAndIncreaseCounter(ctx, "aiven_get_secret_and_config_error_total")

				switch {
				case errors.Is(err, aiven.ErrUnsuitableSecret):
					return naistrix.Errorf(`The secret we found for username %q is not suitable for this command.
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
