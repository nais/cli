package aiven_services

import (
	"fmt"
	"strings"

	"github.com/nais/liberator/pkg/apis/aiven.nais.io/v1"
)

type ServiceSetup struct {
	Instance string
	Pool     KafkaPool
	Access   OpenSearchAccess
}

type Service interface {
	Name() string
	Setup(setup *ServiceSetup)
	Apply(aivenApplicationSpec *aiven_nais_io_v1.AivenApplicationSpec, namespace string)
	Generate(generator SecretGenerator) error
	Is(other Service) bool
}

type SecretGenerator interface {
	CreateKafkaConfigs() error
	CreateOpenSearchConfigs() error
}

var Services = map[string]Service{}

func init() {
	for _, svc := range []Service{
		&Kafka{},
		&OpenSearch{},
	} {
		Services[svc.Name()] = svc
	}
}

func FromString(service string) (Service, error) {
	svc, ok := Services[strings.ToLower(service)]
	if !ok {
		return nil, fmt.Errorf("unknown service: %v", service)
	}
	return svc, nil
}
