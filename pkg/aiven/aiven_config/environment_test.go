package aiven_config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKafkaEnvironmentConfigGenerated(t *testing.T) {
	var envKeys = []string{
		KafkaCAKey,
		KafkaCertificateKey,
		KafkaPrivateKeyKey,
		KafkaClientKeyStoreP12File,
		KafkaClientTruststoreJksFile,
		KafkaCredStorePasswordKey,
		KafkaSchemaRegistryKey,
	}

	tmpDest, err := os.MkdirTemp(os.TempDir(), "test-")
	assert.NoError(t, err)
	err = WriteKafkaEnvConfigToFile(setupSecret(envKeys), tmpDest)
	assert.NoError(t, err)

	result, err := os.ReadFile(filepath.Join(tmpDest, KafkaEnvName))
	assert.NoError(t, err)

	assert.True(t, strings.Contains(string(result), KafkaClientTruststoreJksFile))
	assert.True(t, strings.Contains(string(result), KafkaCredStorePasswordKey))
	assert.True(t, strings.Contains(string(result), KafkaSchemaRegistryKey))
	assert.True(t, strings.Contains(string(result), KafkaCertificateKey))
	assert.True(t, strings.Contains(string(result), KafkaCAKey))
	assert.True(t, strings.Contains(string(result), KafkaPrivateKeyKey))
	assert.True(t, strings.Contains(string(result), KafkaClientKeyStoreP12File))

	defer os.Remove(tmpDest)
}
