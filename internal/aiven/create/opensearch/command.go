package opensearch

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/aiven"
	"github.com/nais/cli/internal/aiven/aiven_services"
	"github.com/nais/cli/internal/aiven/create"
	"github.com/nais/cli/internal/k8s"
)

type Flags struct {
	*create.Flags
	Instance string
	Access   aiven_services.OpenSearchAccess
}

func Run(ctx context.Context, args create.Arguments, flags *Flags) error {
	service := &aiven_services.OpenSearch{}
	aivenConfig := aiven.Setup(
		ctx,
		k8s.SetupControllerRuntimeClient(),
		service,
		args.Username, args.Namespace, flags.Secret, flags.Expire,
		&aiven_services.ServiceSetup{
			Instance: flags.Instance,
			Access:   flags.Access,
		},
	)
	aivenApp, err := aivenConfig.GenerateApplication()
	if err != nil {
		return fmt.Errorf("an error occurred generating 'AivenApplication': %v", err)
	}

	fmt.Printf("Use the following command to generate configuration secrets:\n\tnais aiven get %v %v %v\n", service.Name(), aivenApp.Spec.SecretName, aivenApp.Namespace)

	return nil
}
