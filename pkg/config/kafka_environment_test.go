package config

import (
	"github.com/nais/cli/pkg/aiven"
	"github.com/nais/cli/pkg/consts"
	"github.com/nais/cli/pkg/test"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestKafkaEnvironmentConfigGenerated(t *testing.T) {

	var envKeys = []string{
		consts.KafkaCAKey,
		consts.KafkaCertificateKey,
		consts.KafkaPrivateKeyKey,
		consts.KafkaClientKeyStoreP12File,
		consts.KafkaClientTruststoreJksFile,
		consts.KafkaCredStorePasswordKey,
		consts.KafkaSchemaRegistryKey,
	}

	tmpDest := test.SetupDest(t)
	kcatConfig, err := NewEnvConfig(test.SetupSecret(envKeys), tmpDest, aiven.Kafka)
	assert.NoError(t, err)

	result, err := kcatConfig.Generate()
	assert.NoError(t, err)

	assert.True(t, strings.Contains(result, consts.KafkaClientTruststoreJksFile))
	assert.True(t, strings.Contains(result, consts.KafkaCredStorePasswordKey))
	assert.True(t, strings.Contains(result, consts.KafkaSchemaRegistryKey))
	assert.True(t, strings.Contains(result, consts.KafkaCertificateKey))
	assert.True(t, strings.Contains(result, consts.KafkaCAKey))
	assert.True(t, strings.Contains(result, consts.KafkaPrivateKeyKey))
	assert.True(t, strings.Contains(result, consts.KafkaClientKeyStoreP12File))

	defer os.Remove(tmpDest)
}

func TestKafkaEnvironmentSecrettMissingRequiredData(t *testing.T) {

	var envKeys = []string{
		consts.KafkaCAKey,
		consts.KafkaCertificateKey,
		consts.KafkaPrivateKeyKey,
		consts.KafkaClientKeyStoreP12File,
		consts.KafkaCredStorePasswordKey,
		consts.KafkaSchemaRegistryKey,
	}

	tmpDest := test.SetupDest(t)
	kcatConfig, err := NewEnvConfig(test.SetupSecret(envKeys), tmpDest, aiven.Kafka)
	assert.NoError(t, err)
	_, err = kcatConfig.Generate()
	assert.EqualError(t, err, "can not generate kafka-secret.env config, secret missing required key: client.truststore.jks")

	defer os.Remove(tmpDest)
}
