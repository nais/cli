package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/internal/aiven"
	"github.com/nais/cli/internal/aiven/aiven_services"
	"github.com/nais/cli/internal/aiven/flag"
	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/k8s"
	"github.com/nais/cli/internal/output"
)

func createOpenSearch(parentFlags *flag.Create) *cli.Command {
	createOpenSearchFlags := &flag.CreateOpenSearch{Create: parentFlags}

	return cli.NewCommand("opensearch", "Grant a user access to an OpenSearch instance.",
		cli.WithValidate(cli.ValidateExactArgs(2)),
		cli.WithArgs("username", "namespace"),
		cli.WithFlag("instance", "i", "The name of the OpenSearch `INSTANCE`.", &createOpenSearchFlags.Instance),
		cli.WithFlag("access", "a", "The access `LEVEL`. Available levels: "+strings.Join(aiven_services.OpenSearchAccesses, ", "), &createOpenSearchFlags.Access),
		cli.WithAutoComplete(func(ctx context.Context, args []string, toComplete string) ([]string, string) {
			return aiven_services.OpenSearchAccesses, ""
		}),
		cli.WithRun(func(ctx context.Context, output output.Output, args []string) error {
			access, err := aiven_services.OpenSearchAccessFromString(createOpenSearchFlags.Access)
			if err != nil {
				return fmt.Errorf(
					"valid values for access: %v",
					strings.Join(aiven_services.OpenSearchAccesses, ", "),
				)
			}

			service := &aiven_services.OpenSearch{}
			aivenConfig := aiven.Setup(
				ctx,
				k8s.SetupControllerRuntimeClient(),
				service,
				args[0],
				args[1],
				createOpenSearchFlags.Secret,
				createOpenSearchFlags.Expire,
				&aiven_services.ServiceSetup{
					Instance: createOpenSearchFlags.Instance,
					Access:   access,
				},
			)
			aivenApp, err := aivenConfig.GenerateApplication()
			if err != nil {
				return fmt.Errorf("an error occurred generating 'AivenApplication': %v", err)
			}

			fmt.Printf("Use the following command to generate configuration secrets:\n\tnais aiven get %v %v %v\n", service.Name(), aivenApp.Spec.SecretName, aivenApp.Namespace)

			return nil
		}),
	)
}
