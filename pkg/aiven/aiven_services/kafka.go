package aiven_services

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/nais/liberator/pkg/apis/aiven.nais.io/v1"
)

type KafkaPool int64

const (
	NavDev KafkaPool = iota
	NavProd
	NavIntegrationTest
	NavInfrastructure
)

var KafkaPools = []string{"nav-dev", "nav-prod", "nav-integration-test", "nav-infrastructure"}

func KafkaPoolFromString(pool string) (KafkaPool, error) {
	switch strings.ToLower(pool) {
	case "nav-dev":
		return NavDev, nil
	case "nav-prod":
		return NavProd, nil
	case "nav-integration-test":
		return NavIntegrationTest, nil
	case "nav-infrastructure":
		return NavInfrastructure, nil
	default:
		return -1, fmt.Errorf("unknown pool: %v", pool)
	}
}

func (p KafkaPool) String() string {
	return KafkaPools[p]
}

type Kafka struct {
	pool KafkaPool
}

func (k *Kafka) Name() string {
	return "kafka"
}

func (k *Kafka) Setup(setup *ServiceSetup) {
	k.pool = setup.Pool
}

func (k *Kafka) Apply(aivenApplicationSpec *aiven_nais_io_v1.AivenApplicationSpec, _ string) {
	aivenApplicationSpec.Kafka = &aiven_nais_io_v1.KafkaSpec{
		Pool: k.pool.String(),
	}
}

func (k *Kafka) Generate(generator SecretGenerator) error {
	return generator.CreateKafkaConfigs()
}

func (k *Kafka) Is(other Service) bool {
	return reflect.TypeOf(other) == reflect.TypeOf(k)
}
