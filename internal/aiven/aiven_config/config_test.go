package aiven_config

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func setupSecret(envKeys []string) *v1.Secret {
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
