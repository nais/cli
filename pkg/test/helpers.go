package test

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"testing"
)

func SetupDest(t *testing.T) string {
	tempDir, err := ioutil.TempDir(os.TempDir(), "test-")
	assert.NoError(t, err)
	return tempDir
}

func SetupSecret(envKeys []string) *v1.Secret {
	namespace := "namespace"
	secretName := "secret-name"
	content := "c29tZS12YWx1ZQ=="
	data := make(map[string][]byte)

	for _, value := range envKeys {
		data[value] = []byte(content)
	}

	createdSecret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Data: data,
	}
	return createdSecret
}
