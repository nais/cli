package config

import (
	"github.com/nais/nais-cli/pkg/consts"
	"github.com/nais/nais-cli/pkg/test"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestKcatConfigGenerated(t *testing.T) {

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
	kcatConfig := NewKCatConfig(test.SetupSecret(envKeys), tmpDest)
	result, err := kcatConfig.Generate()
	assert.NoError(t, err)

	assert.True(t, strings.Contains(result, KafkaCatSslCaLocation))
	assert.True(t, strings.Contains(result, KafkaCatSslKeyLocation))
	assert.True(t, strings.Contains(result, KafkaCatSslCertificateLocation))
	assert.True(t, strings.Contains(result, KafkaSecurityProtocolLocation))

	defer os.Remove(tmpDest)
}

func TestKcatSecretMissingRequiredData(t *testing.T) {

	var envKeys = []string{
		consts.KafkaCAKey,
		consts.KafkaCertificateKey,
	}

	tmpDest := test.SetupDest(t)
	kcatConfig := NewKCatConfig(test.SetupSecret(envKeys), tmpDest)
	_, err := kcatConfig.Generate()
	assert.EqualError(t, err, "can not generate kcat.conf config, secret missing required key: KAFKA_PRIVATE_KEY")

	defer os.Remove(tmpDest)
}
