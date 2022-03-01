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
