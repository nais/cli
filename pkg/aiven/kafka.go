package aiven

import (
	"fmt"
	"strings"
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
	case KafkaPools[0]:
		return NavDev, nil
	case KafkaPools[1]:
		return NavProd, nil
	case KafkaPools[2]:
		return NavIntegrationTest, nil
	case KafkaPools[3]:
		return NavInfrastructure, nil
	default:
		return -1, fmt.Errorf("unknown pool: %v", pool)
	}
}

func (p KafkaPool) String() string {
	return KafkaPools[p]
}
