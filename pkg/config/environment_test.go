package config

import (
	"github.com/nais/cli/pkg/consts"
	"github.com/nais/cli/pkg/test"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
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
	err := WriteKafkaEnvConfigToFile(test.SetupSecret(envKeys), tmpDest)
	assert.NoError(t, err)

	result, err := ioutil.ReadFile(filepath.Join(tmpDest, KafkaEnvName))
	assert.NoError(t, err)

	assert.True(t, strings.Contains(string(result), consts.KafkaClientTruststoreJksFile))
	assert.True(t, strings.Contains(string(result), consts.KafkaCredStorePasswordKey))
	assert.True(t, strings.Contains(string(result), consts.KafkaSchemaRegistryKey))
	assert.True(t, strings.Contains(string(result), consts.KafkaCertificateKey))
	assert.True(t, strings.Contains(string(result), consts.KafkaCAKey))
	assert.True(t, strings.Contains(string(result), consts.KafkaPrivateKeyKey))
	assert.True(t, strings.Contains(string(result), consts.KafkaClientKeyStoreP12File))

	defer os.Remove(tmpDest)
}
