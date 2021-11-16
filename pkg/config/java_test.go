package config

import (
	"github.com/nais/cli/pkg/consts"
	"github.com/nais/cli/pkg/test"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestJavaConfigGenerated(t *testing.T) {

	var envKeys = []string{
		consts.KafkaBrokersKey,
		consts.KafkaClientKeyStoreP12File,
		consts.KafkaClientTruststoreJksFile,
		consts.KafkaCredStorePasswordKey,
	}

	tmpDest := test.SetupDest(t)
	javaConfig := NewJavaConfig(test.SetupSecret(envKeys), tmpDest)

	result, err := javaConfig.Generate()
	assert.NoError(t, err)

	assert.True(t, strings.Contains(result, consts.KafkaClientTruststoreJksFile))
	assert.True(t, strings.Contains(result, consts.KafkaClientKeyStoreP12File))
	assert.True(t, strings.Contains(result, KeyPassProp))
	assert.True(t, strings.Contains(result, KeyStorePassProp))
	assert.True(t, strings.Contains(result, TrustStorePassProp))
	assert.True(t, strings.Contains(result, KeyStoreLocationProp))

	defer os.Remove(tmpDest)
}

func TestJavaSecretMissingRequiredData(t *testing.T) {

	var envKeys = []string{
		consts.KafkaCAKey,
		consts.KafkaCertificateKey,
		consts.KafkaPrivateKeyKey,
		consts.KafkaClientKeyStoreP12File,
		consts.KafkaCredStorePasswordKey,
		consts.KafkaSchemaRegistryKey,
	}

	tmpDest := test.SetupDest(t)
	javaConfig := NewJavaConfig(test.SetupSecret(envKeys), tmpDest)
	_, err := javaConfig.Generate()
	assert.EqualError(t, err, "can not generate kafka.properties config, secret missing required key: client.truststore.jks")

	defer os.Remove(tmpDest)
}
