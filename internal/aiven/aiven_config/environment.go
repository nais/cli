package aiven_config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
)

const (
	KafkaEnvName = "kafka-secret.env"
)

type fileTuple struct {
	Key     string
	PathKey string
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
	var envsToFile strings.Builder
	envsToFile.WriteString(fmt.Sprintf("# nais-cli %s .env\n", time.Now().Truncate(time.Minute)))
	for fileName, tuple := range secretFilesToSave {
		err := os.WriteFile(filepath.Join(destinationPath, fileName), secret.Data[tuple.Key], FilePermission)
		if err != nil {
			return err
		}

		envsToFile.WriteString(fmt.Sprintf("%s=\"%s\"\n", tuple.PathKey, filepath.Join(destinationPath, fileName)))
	}

	for _, key := range envsToSave {
		envsToFile.WriteString(fmt.Sprintf("%s=\"%s\"\n", key, string(secret.Data[key])))
	}

	err := os.WriteFile(filepath.Join(destinationPath, destinationFilename), []byte(envsToFile.String()), FilePermission)
	if err != nil {
		return err
	}

	return nil
}
