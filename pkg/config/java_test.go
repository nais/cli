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

func TestJavaConfigGenerated(t *testing.T) {

	var envKeys = []string{
		consts.KafkaBrokersKey,
		consts.KafkaClientKeyStoreP12File,
		consts.KafkaClientTruststoreJksFile,
		consts.KafkaCredStorePasswordKey,
	}

	tmpDest := test.SetupDest(t)
	err := NewJavaConfig(test.SetupSecret(envKeys), tmpDest)
	assert.NoError(t, err)

	result, err := ioutil.ReadFile(filepath.Join(tmpDest, JavaConfigName))
	assert.NoError(t, err)

	assert.True(t, strings.Contains(string(result), consts.KafkaClientTruststoreJksFile))
	assert.True(t, strings.Contains(string(result), consts.KafkaClientKeyStoreP12File))
	assert.True(t, strings.Contains(string(result), KeyPassProp))
	assert.True(t, strings.Contains(string(result), KeyStorePassProp))
	assert.True(t, strings.Contains(string(result), TrustStorePassProp))
	assert.True(t, strings.Contains(string(result), KeyStoreLocationProp))

	defer os.Remove(tmpDest)
}
