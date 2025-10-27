package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/internal/aiven"
	"github.com/nais/cli/internal/aiven/aiven_services"
	"github.com/nais/cli/internal/aiven/command/flag"
	"github.com/nais/cli/internal/k8s"
	"github.com/nais/naistrix"
)

func createOpenSearch(parentFlags *flag.Create) *naistrix.Command {
	createOpenSearchFlags := &flag.CreateOpenSearch{Create: parentFlags}

	return &naistrix.Command{
		Name:  "opensearch",
		Title: "Grant a user access to an OpenSearch instance.",
		Args: []naistrix.Argument{
			{Name: "username"},
			{Name: "namespace"},
		},
		AutoCompleteFunc: func(context.Context, *naistrix.Arguments, string) ([]string, string) {
			return aiven_services.OpenSearchAccesses, ""
		},
		Flags: createOpenSearchFlags,
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			access, err := aiven_services.OpenSearchAccessFromString(createOpenSearchFlags.Access)
			if err != nil {
				return fmt.Errorf(
					"valid values for access: %v",
					strings.Join(aiven_services.OpenSearchAccesses, ", "),
				)
			}

			if createOpenSearchFlags.Secret == "" {
				if createOpenSearchFlags.Secret, err = aiven.CreateSecretName(args.Get("username"), args.Get("namespace")); err != nil {
					return fmt.Errorf("creating secret name: %v", err)
				}
			}

			service := &aiven_services.OpenSearch{}
			aivenConfig := aiven.Setup(
				ctx,
				k8s.SetupControllerRuntimeClient(),
				service,
				args.Get("username"),
				args.Get("namespace"),
				createOpenSearchFlags.Expire,
				&aiven_services.ServiceSetup{
					Instance:   createOpenSearchFlags.Instance,
					Access:     access,
					SecretName: createOpenSearchFlags.Secret,
				},
			)
			aivenApp, err := aivenConfig.GenerateApplication(out)
			if err != nil {
				return fmt.Errorf("an error occurred generating 'AivenApplication': %v", err)
			}

			out.Println("Use the following command to generate configuration secrets:")
			out.Printf("\tnais aiven get %v %v %v\n", service.Name(), aivenApp.Spec.OpenSearch.SecretName, aivenApp.Namespace)

			return nil
		},
	}
}
