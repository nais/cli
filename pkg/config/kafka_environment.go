package config

import (
	"fmt"
	"github.com/nais/cli/pkg/common"
	"github.com/nais/cli/pkg/consts"
	v1 "k8s.io/api/core/v1"
	"path/filepath"
	"time"
)

const (
	KafkaSchemaRegistryEnvName = "kafka-secret.env"
)

func WriteKafkaEnvConfigToFile(secret *v1.Secret, destinationPath string) error {
	kafkaSecretsToSaveToFile := map[string]string{
		consts.KafkaCertificateCrtFile:      consts.KafkaCertificatePathKey,
		consts.KafkaPrivateKeyPemFile:       consts.KafkaPrivateKeyPathKey,
		consts.KafkaCACrtFile:               consts.KafkaCAPathKey,
		consts.KafkaClientKeyStoreP12File:   consts.KafkaKeystorePathKey,
		consts.KafkaClientTruststoreJksFile: consts.KafkaTruststorePathKey,
	}
	kafkaEnvsToSaveToFile := []string{
		consts.KafkaBrokersKey, consts.KafkaCredStorePasswordKey, consts.KafkaSchemaRegistryKey,
		consts.KafkaSchemaRegistryPasswordKey, consts.KafkaSchemaRegistryUserKey, consts.KafkaCertificatePathKey,
		consts.KafkaPrivateKeyPathKey, consts.KafkaCAPathKey,
	}

	return writeConfigToFile(secret, destinationPath, KafkaSchemaRegistryEnvName, kafkaEnvsToSaveToFile, kafkaSecretsToSaveToFile)
}

func writeConfigToFile(secret *v1.Secret, destinationPath, destinationFilename string, envsToSave []string, secretFilesToSave map[string]string) error {
	envsToFile := fmt.Sprintf("# nais-cli %s .env\n", time.Now().Truncate(time.Minute))
	for fileName, pathKey := range secretFilesToSave {
		err := common.WriteToFile(destinationPath, fileName, secret.Data[pathKey])
		if err != nil {
			return err
		}

		envsToFile += fmt.Sprintf("%s=\"%s\"\n", pathKey, filepath.Join(destinationPath, fileName))
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
