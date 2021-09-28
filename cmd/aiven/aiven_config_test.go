package aiven

import (
	"github.com/nais/nais-cli/cmd"
	"github.com/nais/nais-cli/pkg/consts"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

// create
func TestAivenConfigCreateMissingArguments(t *testing.T) {
	err := CreateCmd.Execute()
	assert.EqualError(t, err, "missing required arguments: username, namespace")
}

func TestAivenConfigCreateNoValidKafkaPool(t *testing.T) {
	setEnvironment("no-pool", consts.AllConfigurationType, "/temp/")
	CreateCmd.SetArgs([]string{"username", "namespace"})
	err := CreateCmd.Execute()
	assert.EqualError(t, err, "valid values for '-pool': nav-dev | nav-prod | nav-integration-test")
}

func TestAivenConfigCreateNamespaceNotFound(t *testing.T) {
	setEnvironment(KafkaNavIntegrationTest, consts.AllConfigurationType, "/temp/")
	CreateCmd.SetArgs([]string{"username", "namespace"})
	err := CreateCmd.Execute()
	assert.EqualError(t, err, "an error occurred generating 'AivenApplication': get namespace: namespaces \"namespace\" not found")
}

// get
func TestAivenConfigGetMissingArguments(t *testing.T) {
	err := GetCmd.Execute()
	assert.EqualError(t, err, "missing required arguments: secret-name, namespace")
}

func TestAivenConfigGetNoValidConfigFlag(t *testing.T) {
	setEnvironment(KafkaNavIntegrationTest, "non-flag", "/temp/")
	GetCmd.SetArgs([]string{"secret-name", "namespace"})
	err := GetCmd.Execute()
	assert.EqualError(t, err, "valid values for '--config': .env, kcat, all")
}

func TestAivenConfigGetNamespaceNotFound(t *testing.T) {
	setEnvironment(KafkaNavIntegrationTest, consts.AllConfigurationType, "/temp/")
	GetCmd.SetArgs([]string{"secret-name", "namespace"})
	err := GetCmd.Execute()
	assert.EqualError(t, err, "retrieve secret and generating config: validate namespace: get namespace: namespaces \"namespace\" not found")
}

// tidy doesn't make sense to test here.

func setEnvironment(kafkaPool, configFlag, dest string) {
	viper.Set(cmd.DestFlag, dest)
	viper.Set(cmd.PoolFlag, kafkaPool)
	viper.Set(cmd.ExpireFlag, 1)
	viper.Set(cmd.SecretNameFlag, "secret")
	viper.Set(cmd.ConfigFlag, configFlag)
}
