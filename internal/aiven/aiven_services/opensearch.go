package aiven_services

import (
	"fmt"

	aiven_nais_io_v1 "github.com/nais/liberator/pkg/apis/aiven.nais.io/v1"
)

type OpenSearchAccess int64

var OpenSearchAccesses = []string{"read", "write", "readwrite", "admin"}

func (p OpenSearchAccess) String() string {
	return OpenSearchAccesses[p]
}

type OpenSearch struct {
	instance   string
	access     OpenSearchAccess
	secretName string
}

func (o *OpenSearch) Name() string {
	return "opensearch"
}

func (o *OpenSearch) Setup(setup *ServiceSetup) {
	o.instance = setup.Instance
	o.access = setup.Access
	o.secretName = setup.SecretName
}

func (o *OpenSearch) Apply(aivenApplicationSpec *aiven_nais_io_v1.AivenApplicationSpec, namespace string) {
	fullyQualifiedInstanceName := fmt.Sprintf("opensearch-%s-%s", namespace, o.instance)
	aivenApplicationSpec.OpenSearch = &aiven_nais_io_v1.OpenSearchSpec{
		Instance:   fullyQualifiedInstanceName,
		Access:     o.access.String(),
		SecretName: o.secretName,
	}
}

func (o *OpenSearch) Generate(generator SecretGenerator) error {
	return generator.CreateOpenSearchConfigs()
}

func (o *OpenSearch) Is(other Service) bool {
	return o.Name() == other.Name()
}
