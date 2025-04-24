package aiven_services

import (
	"fmt"
	"strings"

	aiven_nais_io_v1 "github.com/nais/liberator/pkg/apis/aiven.nais.io/v1"
)

type KafkaPool string

func KafkaPoolFromString(pool string) (KafkaPool, error) {
	if !strings.Contains(pool, "-") {
		return "", fmt.Errorf("invalid pool: %v", pool)
	}
	return KafkaPool(pool), nil
}

func (p KafkaPool) String() string {
	return string(p)
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
	return k.Name() == other.Name()
}
