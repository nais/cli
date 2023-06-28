package aiven_config

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"os"
	"path/filepath"
	"time"
)

const (
	KafkaEnvName          = "kafka-secret.env"
	OpenSearchEnvName     = "opensearch-secret.env"
	OpenSearchURIKey      = "OPEN_SEARCH_URI"
	OpenSearchUsernameKey = "OPEN_SEARCH_USERNAME"
	OpenSearchPasswordKey = "OPEN_SEARCH_PASSWORD"
)

type fileTuple struct {
	Key     string
	PathKey string
}

func WriteOpenSearchEnvConfigToFile(secret *v1.Secret, destinationPath string) error {
	envsToSaveToFile := []string{
		OpenSearchURIKey, OpenSearchPasswordKey, OpenSearchUsernameKey,
	}
	return writeConfigToFile(secret, destinationPath, OpenSearchEnvName, envsToSaveToFile, map[string]fileTuple{})
}

func WriteKafkaEnvConfigToFile(secret *v1.Secret, destinationPath string) error {
	kafkaSecretsToSaveToFile := map[string]fileTuple{
		KafkaCertificateCrtFile:      {KafkaCertificateKey, KafkaCertificatePathKey},
		KafkaPrivateKeyPemFile:       {KafkaPrivateKeyKey, KafkaPrivateKeyPathKey},
		KafkaCACrtFile:               {KafkaCAKey, KafkaCAPathKey},
		KafkaClientKeyStoreP12File:   {KafkaClientKeyStoreP12File, KafkaKeystorePathKey},
		KafkaClientTruststoreJksFile: {KafkaClientTruststoreJksFile, KafkaTruststorePathKey},
	}
	kafkaEnvsToSaveToFile := []string{
		KafkaBrokersKey, KafkaCredStorePasswordKey, KafkaSchemaRegistryKey,
		KafkaSchemaRegistryPasswordKey, KafkaSchemaRegistryUserKey, KafkaCertificateKey,
		KafkaPrivateKeyKey, KafkaCAKey,
	}

	return writeConfigToFile(secret, destinationPath, KafkaEnvName, kafkaEnvsToSaveToFile, kafkaSecretsToSaveToFile)
}

func writeConfigToFile(secret *v1.Secret, destinationPath, destinationFilename string, envsToSave []string, secretFilesToSave map[string]fileTuple) error {
	envsToFile := fmt.Sprintf("# nais-cli %s .env\n", time.Now().Truncate(time.Minute))
	for fileName, tuple := range secretFilesToSave {
		err := os.WriteFile(filepath.Join(destinationPath, fileName), secret.Data[tuple.Key], FilePermission)
		if err != nil {
			return err
		}

		envsToFile += fmt.Sprintf("%s=\"%s\"\n", tuple.PathKey, filepath.Join(destinationPath, fileName))
	}

	for _, key := range envsToSave {
		envsToFile += fmt.Sprintf("%s=\"%s\"\n", key, string(secret.Data[key]))
	}

	err := os.WriteFile(filepath.Join(destinationPath, destinationFilename), []byte(envsToFile), FilePermission)
	if err != nil {
		return err
	}

	return nil
}
