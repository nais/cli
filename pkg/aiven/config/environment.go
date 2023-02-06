package config

import (
	"fmt"
	"github.com/nais/cli/pkg/aiven/consts"
	"github.com/nais/cli/pkg/common"
	v1 "k8s.io/api/core/v1"
	"path/filepath"
	"time"
)

const (
	KafkaEnvName      = "kafka-secret.env"
	OpenSearchEnvName = "opensearch-secret.env"
)

type fileTuple struct {
	Key     string
	PathKey string
}

func WriteOpenSearchEnvConfigToFile(secret *v1.Secret, destinationPath string) error {
	envsToSaveToFile := []string{
		consts.OpenSearchURIKey, consts.OpenSearchPasswordKey, consts.OpenSearchUsernameKey,
	}
	return writeConfigToFile(secret, destinationPath, OpenSearchEnvName, envsToSaveToFile, map[string]fileTuple{})
}

func WriteKafkaEnvConfigToFile(secret *v1.Secret, destinationPath string) error {
	kafkaSecretsToSaveToFile := map[string]fileTuple{
		consts.KafkaCertificateCrtFile:      {consts.KafkaCertificateKey, consts.KafkaCertificatePathKey},
		consts.KafkaPrivateKeyPemFile:       {consts.KafkaPrivateKeyKey, consts.KafkaPrivateKeyPathKey},
		consts.KafkaCACrtFile:               {consts.KafkaCAKey, consts.KafkaCAPathKey},
		consts.KafkaClientKeyStoreP12File:   {consts.KafkaClientKeyStoreP12File, consts.KafkaKeystorePathKey},
		consts.KafkaClientTruststoreJksFile: {consts.KafkaClientTruststoreJksFile, consts.KafkaTruststorePathKey},
	}
	kafkaEnvsToSaveToFile := []string{
		consts.KafkaBrokersKey, consts.KafkaCredStorePasswordKey, consts.KafkaSchemaRegistryKey,
		consts.KafkaSchemaRegistryPasswordKey, consts.KafkaSchemaRegistryUserKey, consts.KafkaCertificateKey,
		consts.KafkaPrivateKeyKey, consts.KafkaCAKey,
	}

	return writeConfigToFile(secret, destinationPath, KafkaEnvName, kafkaEnvsToSaveToFile, kafkaSecretsToSaveToFile)
}

func writeConfigToFile(secret *v1.Secret, destinationPath, destinationFilename string, envsToSave []string, secretFilesToSave map[string]fileTuple) error {
	envsToFile := fmt.Sprintf("# nais-cli %s .env\n", time.Now().Truncate(time.Minute))
	for fileName, tuple := range secretFilesToSave {
		err := common.WriteToFile(destinationPath, fileName, secret.Data[tuple.Key])
		if err != nil {
			return err
		}

		envsToFile += fmt.Sprintf("%s=\"%s\"\n", tuple.PathKey, filepath.Join(destinationPath, fileName))
	}

	for _, key := range envsToSave {
		envsToFile += fmt.Sprintf("%s=\"%s\"\n", key, string(secret.Data[key]))
	}

	err := common.WriteToFile(destinationPath, destinationFilename, []byte(envsToFile))
	if err != nil {
		return err
	}

	return nil
}
