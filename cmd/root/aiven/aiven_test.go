package aiven

import (
	"github.com/nais/cli/cmd"
	"github.com/nais/cli/pkg/consts"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

// create
func TestAivenConfigCreateMissingArguments(t *testing.T) {
	err := createCmd.Execute()
	assert.EqualError(t, err, "missing required arguments: username, namespace")
}

func TestAivenConfigCreateNoValidKafkaPool(t *testing.T) {
	setEnvironment("no-pool", consts.AllConfigurationType)
	createCmd.SetArgs([]string{"username", "namespace"})
	err := createCmd.Execute()
	assert.EqualError(t, err, "valid values for '-pool': nav-dev | nav-prod | nav-integration-test | nav-infrastructure")
}

// get
func TestAivenConfigGetMissingArguments(t *testing.T) {
	err := getCmd.Execute()
	assert.EqualError(t, err, "missing required arguments: secret-name, namespace")
}

func TestAivenConfigGetNoValidConfigFlag(t *testing.T) {
	setEnvironment(KafkaNavIntegrationTest, "non-flag")
	getCmd.SetArgs([]string{"secret-name", "namespace"})
	err := getCmd.Execute()
	assert.EqualError(t, err, "valid values for '--config': java, kcat, .env, all")
}

// tidy doesn't make sense to test here.

func setEnvironment(kafkaPool, configFlag string) {
	viper.Set(cmd.PoolFlag, kafkaPool)
	viper.Set(cmd.ExpireFlag, 1)
	viper.Set(cmd.SecretNameFlag, "secret")
	viper.Set(cmd.ConfigFlag, configFlag)
}
