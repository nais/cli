package aiven_services

import (
	"fmt"
	"strings"

	aiven_nais_io_v1 "github.com/nais/liberator/pkg/apis/aiven.nais.io/v1"
)

type OpenSearchAccess int64

const (
	Read OpenSearchAccess = iota
	Write
	ReadWrite
	Admin
)

var OpenSearchAccesses = []string{"read", "write", "readwrite", "admin"}

func OpenSearchAccessFromString(access string) (OpenSearchAccess, error) {
	switch strings.ToLower(access) {
	case OpenSearchAccesses[0]:
		return Read, nil
	case OpenSearchAccesses[1]:
		return Write, nil
	case OpenSearchAccesses[2]:
		return ReadWrite, nil
	case OpenSearchAccesses[3]:
		return Admin, nil
	default:
		return -1, fmt.Errorf("unknown access: %v", access)
	}
}

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
