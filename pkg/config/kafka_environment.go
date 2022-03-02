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
	KafkaEnvName = "kafka-secret.env"
)

type FileTuple struct {
	Filename string
	PathKey  string
}

func getSecretsToSaveToFile() []FileTuple {
	return []FileTuple{
		{consts.KafkaCertificateCrtFile, consts.KafkaCertificatePathKey},
		{consts.KafkaPrivateKeyPemFile, consts.KafkaPrivateKeyPathKey},
		{consts.KafkaCACrtFile, consts.KafkaCAPathKey},
		{consts.KafkaClientKeyStoreP12File, consts.KafkaKeystorePathKey},
		{consts.KafkaClientTruststoreJksFile, consts.KafkaTruststorePathKey},
	}
}

func getEnvsToSaveToFile() []string {
	return []string{
		consts.KafkaBrokersKey, consts.KafkaCredStorePasswordKey, consts.KafkaSchemaRegistryKey,
		consts.KafkaSchemaRegistryPasswordKey, consts.KafkaSchemaRegistryUserKey, consts.KafkaCertificatePathKey,
		consts.KafkaPrivateKeyPathKey, consts.KafkaCAPathKey,
	}
}

func WriteKafkaEnvConfigToFile(secret *v1.Secret, destinationPath string) error {
	return writeConfigToFile(secret, destinationPath, KafkaEnvName, getEnvsToSaveToFile(), getSecretsToSaveToFile())
}

func writeConfigToFile(secret *v1.Secret, destinationPath, destinationFilename string, envsToSave []string, secretFilesToSave []FileTuple) error {
	envsToFile := fmt.Sprintf("# nais-cli %s .env\n", time.Now().Truncate(time.Minute))
	for _, secretToFile := range secretFilesToSave {
		err := common.WriteToFile(destinationPath, secretToFile.Filename, secret.Data[secretToFile.PathKey])
		if err != nil {
			return err
		}

		envsToFile += fmt.Sprintf("%s=\"%s\"\n", secretToFile.PathKey, filepath.Join(destinationPath, secretToFile.Filename))
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
