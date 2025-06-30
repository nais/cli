package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/pkg/cli/v2"
	"github.com/nais/cli/v2/internal/aiven"
	"github.com/nais/cli/v2/internal/aiven/aiven_services"
	"github.com/nais/cli/v2/internal/aiven/command/flag"
	"github.com/nais/cli/v2/internal/k8s"
)

func createOpenSearch(parentFlags *flag.Create) *cli.Command {
	createOpenSearchFlags := &flag.CreateOpenSearch{Create: parentFlags}

	return &cli.Command{
		Name:  "opensearch",
		Title: "Grant a user access to an OpenSearch instance.",
		Args: []cli.Argument{
			{Name: "username"},
			{Name: "namespace"},
		},
		AutoCompleteFunc: func(ctx context.Context, args []string, toComplete string) ([]string, string) {
			return aiven_services.OpenSearchAccesses, ""
		},
		Flags: createOpenSearchFlags,
		RunFunc: func(ctx context.Context, out cli.Output, args []string) error {
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
			aivenApp, err := aivenConfig.GenerateApplication(out)
			if err != nil {
				return fmt.Errorf("an error occurred generating 'AivenApplication': %v", err)
			}

			out.Println("Use the following command to generate configuration secrets:")
			out.Printf("\tnais aiven get %v %v %v\n", service.Name(), aivenApp.Spec.SecretName, aivenApp.Namespace)

			return nil
		},
	}
}
