package secret

import (
	"fmt"
	"github.com/nais/nais-cli/pkg/config"
	"github.com/nais/nais-cli/pkg/consts"
	"github.com/nais/nais-cli/pkg/test"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestGeneratedFilesAndSecretConfiguration(t *testing.T) {

	content := "c29tZS12YWx1ZQ=="

	var envKeys = []string{
		consts.KafkaCertificate,
		consts.KafkaCa,
		consts.KafkaPrivateKey,
		consts.KafkaClientKeystoreP12,
		consts.KafkaClientTruststoreJks,
		consts.KafkaCredStorePassword,
		consts.KafkaSchemaRegistry,
	}

	tempDir := test.SetupDest(t)
	existingSecret := test.SetupSecret(envKeys)

	secret := SetupSecretConfiguration(existingSecret, config.ENV, tempDir)

	kafkaEnvData, err := secret.Config()
	assert.NoError(t, err)

	// Test kafka.env file created
	assert.True(t, strings.Contains(kafkaEnvData, "client.truststore.jks"))
	assert.True(t, strings.Contains(kafkaEnvData, "KAFKA_CREDSTORE_PASSWORD"))
	assert.True(t, strings.Contains(kafkaEnvData, "KAFKA_SCHEMA_REGISTRY"))
	assert.True(t, strings.Contains(kafkaEnvData, "KAFKA_CERTIFICATE"))
	assert.True(t, strings.Contains(kafkaEnvData, "KAFKA_CA"))
	assert.True(t, strings.Contains(kafkaEnvData, "KAFKA_PRIVATE_KEY"))
	assert.True(t, strings.Contains(kafkaEnvData, "client.keystore.p12"))

	var fileKeys = []string{
		consts.KafkaCertificateCrtFile,
		consts.KafkaCACrtFile,
		consts.KafkaPrivateKeyPemFile,
		consts.KafkaClientKeyStoreP12File,
		consts.KafkaClientTruststoreJksFile,
	}

	// Test cert files created
	for _, value := range fileKeys {
		certFileData, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", tempDir, value))
		assert.NoError(t, err)
		assert.Equal(t, content, string(certFileData))
	}

	// Test kcat file created
	secret = SetupSecretConfiguration(existingSecret, config.KCAT, tempDir)
	kcatData, err := secret.Config()
	assert.NoError(t, err)
	assert.True(t, strings.Contains(kcatData, "ssl.ca.location"))
	assert.True(t, strings.Contains(kcatData, "ssl.key.location"))
	assert.True(t, strings.Contains(kcatData, "ssl.certificate"))
	assert.True(t, strings.Contains(kcatData, "security.protocol"))

	defer os.Remove(tempDir)
}
