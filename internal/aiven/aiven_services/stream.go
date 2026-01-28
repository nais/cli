package aiven_services

import (
	kafka_nais_io_v1 "github.com/nais/liberator/pkg/apis/kafka.nais.io/v1"
)

type Stream struct {
	Pool            KafkaPool
	StreamPrefix    string
	AdditionalUsers []string
}

func (s *Stream) Name() string {
	return s.StreamPrefix
}

func (s *Stream) Setup(setup *ServiceSetup) {
	s.Pool = setup.Pool
}

func (s *Stream) Apply(streamSpec *kafka_nais_io_v1.StreamSpec, _ string) {
	streamSpec.Pool = s.Pool.String()
	// streamSpec.AdditionalUsers = append(streamSpec.AdditionalUsers, TODO)
}

func (s *Stream) Is(other Service) bool {
	return s.Name() == other.Name()
}
