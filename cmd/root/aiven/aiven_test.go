package aiven

import (
	"github.com/nais/cli/cmd"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

// create
func TestAivenConfigCreateMissingArguments(t *testing.T) {
	err := createCmd.Execute()
	assert.EqualError(t, err, "missing required arguments: service, username, namespace")
}

func TestAivenConfigCreateNoValidKafkaPool(t *testing.T) {
	setEnvironment("no-pool")
	createCmd.SetArgs([]string{"kafka", "username", "namespace"})
	err := createCmd.Execute()
	assert.EqualError(t, err, "valid values for '-pool': nav-dev | nav-prod | nav-integration-test | nav-infrastructure")
}

// get
func TestAivenConfigGetMissingArguments(t *testing.T) {
	err := getCmd.Execute()
	assert.EqualError(t, err, "missing required arguments: service, secret-name, namespace")
}

// tidy doesn't make sense to test here.

func setEnvironment(kafkaPool string) {
	viper.Set(cmd.PoolFlag, kafkaPool)
	viper.Set(cmd.ExpireFlag, 1)
	viper.Set(cmd.SecretNameFlag, "secret")
}
