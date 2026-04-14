package aiven_services

import (
	aiven_nais_io_v1 "github.com/nais/liberator/pkg/apis/aiven.nais.io/v1"
)

type KafkaPool string

func (p KafkaPool) String() string {
	return string(p)
}

type Kafka struct {
	pool       KafkaPool
	secretName string
}

func (k *Kafka) Name() string {
	return "kafka"
}

func (k *Kafka) Setup(setup *ServiceSetup) {
	k.pool = setup.Pool
	k.secretName = setup.SecretName
}

func (k *Kafka) Apply(aivenApplicationSpec *aiven_nais_io_v1.AivenApplicationSpec, _ string) {
	aivenApplicationSpec.Kafka = &aiven_nais_io_v1.KafkaSpec{
		Pool:       k.pool.String(),
		SecretName: k.secretName,
	}
}

func (k *Kafka) Generate(generator SecretGenerator) error {
	return generator.CreateKafkaConfigs()
}

func (k *Kafka) Is(other Service) bool {
	return k.Name() == other.Name()
}
