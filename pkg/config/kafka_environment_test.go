package config

import (
	"github.com/nais/nais-cli/pkg/consts"
	"github.com/nais/nais-cli/pkg/test"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestKafkaEnvironmentConfigGenerated(t *testing.T) {

	var envKeys = []string{
		consts.KafkaCertificate,
		consts.KafkaCa,
		consts.KafkaPrivateKey,
		consts.KafkaClientKeystoreP12,
		consts.KafkaClientTruststoreJks,
		consts.KafkaCredStorePassword,
		consts.KafkaSchemaRegistry,
	}

	tmpDest := test.SetupDest(t)
	kcatConfig := NewEnvConfig(test.SetupSecret(envKeys), KafkaConfigEnvToFileMap, tmpDest)
	result, err := kcatConfig.Generate()
	assert.NoError(t, err)

	assert.True(t, strings.Contains(result, "client.truststore.jks"))
	assert.True(t, strings.Contains(result, "KAFKA_CREDSTORE_PASSWORD"))
	assert.True(t, strings.Contains(result, "KAFKA_SCHEMA_REGISTRY"))
	assert.True(t, strings.Contains(result, "KAFKA_CERTIFICATE"))
	assert.True(t, strings.Contains(result, "KAFKA_CA"))
	assert.True(t, strings.Contains(result, "KAFKA_PRIVATE_KEY"))
	assert.True(t, strings.Contains(result, "client.keystore.p12"))

	defer os.Remove(tmpDest)
}

func TestKafkaEnvironmentSecrettMissingRequiredData(t *testing.T) {

	var envKeys = []string{
		consts.KafkaCertificate,
		consts.KafkaCa,
		consts.KafkaPrivateKey,
		consts.KafkaClientKeystoreP12,
		consts.KafkaCredStorePassword,
		consts.KafkaSchemaRegistry,
	}

	tmpDest := test.SetupDest(t)
	kcatConfig := NewEnvConfig(test.SetupSecret(envKeys), KafkaConfigEnvToFileMap, tmpDest)
	_, err := kcatConfig.Generate()
	assert.EqualError(t, err, "can not generate kafka-secret.env config, secret missing required key: client.truststore.jks")

	defer os.Remove(tmpDest)
}
