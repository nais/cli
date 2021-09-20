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
		consts.KafkaCertificate,
		consts.KafkaCa,
		consts.KafkaPrivateKey,
		consts.KafkaClientKeystoreP12,
		consts.KafkaClientTruststoreJks,
		consts.KafkaCredStorePassword,
		consts.KafkaSchemaRegistry,
	}

	tmpDest := test.SetupDest(t)
	kcatConfig := NewKCatConfig(test.SetupSecret(envKeys), KCatEnvToFileMap, tmpDest)
	result, err := kcatConfig.Generate()
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.True(t, strings.Contains(result, "ssl.ca.location"))
	assert.True(t, strings.Contains(result, "ssl.key.location"))
	assert.True(t, strings.Contains(result, "ssl.certificate"))
	assert.True(t, strings.Contains(result, "security.protocol"))

	defer os.Remove(tmpDest)
}

func TestKcatSecretMissingRequiredData(t *testing.T) {

	var envKeys = []string{
		consts.KafkaCertificate,
		consts.KafkaCa,
	}

	tmpDest := test.SetupDest(t)
	kcatConfig := NewKCatConfig(test.SetupSecret(envKeys), KCatEnvToFileMap, tmpDest)
	_, err := kcatConfig.Generate()
	assert.EqualError(t, err, "can not generate kcat.conf config, secret missing required key: KAFKA_PRIVATE_KEY")

	defer os.Remove(tmpDest)
}
