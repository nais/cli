package aiven_config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJavaConfigGenerated(t *testing.T) {

	var envKeys = []string{
		KafkaBrokersKey,
		KafkaClientKeyStoreP12File,
		KafkaClientTruststoreJksFile,
		KafkaCredStorePasswordKey,
	}

	tmpDest, err := os.MkdirTemp(os.TempDir(), "test-")
	assert.NoError(t, err)
	err = NewJavaConfig(setupSecret(envKeys), tmpDest)
	assert.NoError(t, err)

	result, err := os.ReadFile(filepath.Join(tmpDest, KafkaJavaConfigName))
	assert.NoError(t, err)

	assert.True(t, strings.Contains(string(result), KafkaClientTruststoreJksFile))
	assert.True(t, strings.Contains(string(result), KafkaClientKeyStoreP12File))
	assert.True(t, strings.Contains(string(result), KeyPassProp))
	assert.True(t, strings.Contains(string(result), KeyStorePassProp))
	assert.True(t, strings.Contains(string(result), TrustStorePassProp))
	assert.True(t, strings.Contains(string(result), KeyStoreLocationProp))

	defer os.Remove(tmpDest)
}
