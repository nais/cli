package aiven_config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKcatConfigGenerated(t *testing.T) {
	envKeys := []string{
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
	err = WriteKCatConfigToFile(setupSecret(envKeys), tmpDest)
	assert.NoError(t, err)

	result, err := os.ReadFile(filepath.Join(tmpDest, KafkaCatConfigName))
	assert.NoError(t, err)

	assert.True(t, strings.Contains(string(result), KafkaCatSslCaLocation))
	assert.True(t, strings.Contains(string(result), KafkaCatSslKeyLocation))
	assert.True(t, strings.Contains(string(result), KafkaCatSslCertificateLocation))
	assert.True(t, strings.Contains(string(result), KafkaSecurityProtocolLocation))

	defer os.Remove(tmpDest)
}
