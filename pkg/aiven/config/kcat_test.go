package config

import (
	"github.com/nais/cli/pkg/aiven/consts"
	"github.com/nais/cli/pkg/test"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
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
	err := WriteKCatConfigToFile(test.SetupSecret(envKeys), tmpDest)
	assert.NoError(t, err)

	result, err := ioutil.ReadFile(filepath.Join(tmpDest, KafkaCatConfigName))
	assert.NoError(t, err)

	assert.True(t, strings.Contains(string(result), KafkaCatSslCaLocation))
	assert.True(t, strings.Contains(string(result), KafkaCatSslKeyLocation))
	assert.True(t, strings.Contains(string(result), KafkaCatSslCertificateLocation))
	assert.True(t, strings.Contains(string(result), KafkaSecurityProtocolLocation))

	defer os.Remove(tmpDest)
}
