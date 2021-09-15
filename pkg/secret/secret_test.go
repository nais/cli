package secret

import (
	"fmt"
	"github.com/nais/nais-d/pkg/config"
	"github.com/nais/nais-d/pkg/consts"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"strings"
	"testing"
)

func TestConfig(t *testing.T) {

	team := "team"
	secretName := "secret-name"
	content := "c29tZS12YWx1ZQ=="

	data := make(map[string][]byte)

	var envKeys = []string{
		consts.KafkaCertificate,
		consts.KafkaCa,
		consts.KafkaPrivateKey,
		consts.KafkaClientKeystoreP12,
		consts.KafkaClientTruststoreJks,
		consts.KafkaCredStorePassword,
		consts.KafkaSchemaRegistry,
	}

	for _, value := range envKeys {
		data[value] = []byte(content)
	}

	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: team,
		},
		Data: data,
	}

	tempDir, err := ioutil.TempDir(os.TempDir(), "test-")
	assert.NoError(t, err)

	err = Config(secret, tempDir, consts.ALL)
	assert.NoError(t, err)

	// Test kafka.env file created
	KafkaEnvData, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", tempDir, config.KafkaSchemaRegistryEnvName))
	assert.True(t, strings.Contains(string(KafkaEnvData), "client.truststore.jks"))
	assert.True(t, strings.Contains(string(KafkaEnvData), "KAFKA_CREDSTORE_PASSWORD"))
	assert.True(t, strings.Contains(string(KafkaEnvData), "KAFKA_SCHEMA_REGISTRY"))
	assert.True(t, strings.Contains(string(KafkaEnvData), "KAFKA_CERTIFICATE"))
	assert.True(t, strings.Contains(string(KafkaEnvData), "KAFKA_CA"))
	assert.True(t, strings.Contains(string(KafkaEnvData), "AFKA_PRIVATE_KEY"))
	assert.True(t, strings.Contains(string(KafkaEnvData), "client.keystore.p12"))


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
	kcatData, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", tempDir, config.KafkaCatConfigName))
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(kcatData), "ssl.ca.location"))
	assert.True(t, strings.Contains(string(kcatData), "ssl.key.location"))
	assert.True(t, strings.Contains(string(kcatData), "ssl.certificate"))
	assert.True(t, strings.Contains(string(kcatData), "security.protocol"))

	defer os.Remove(tempDir)
}
