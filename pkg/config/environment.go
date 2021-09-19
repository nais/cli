package config

import (
	"fmt"
	"github.com/nais/nais-cli/pkg/common"
	"github.com/nais/nais-cli/pkg/consts"
	v1 "k8s.io/api/core/v1"
	"strings"
	"time"
)

const (
	KafkaSchemaRegistryEnvName = "kafka-secret.env"
)

func NewEnvConfig(secret *v1.Secret, dest string) Config {
	return &KafkaEnvironment{
		Envs:       "",
		Secret:     secret,
		PrefixPath: dest,
		RequiredFiles: map[string]string{
			consts.KafkaCertificate:         consts.KafkaCertificateCrtFile,
			consts.KafkaPrivateKey:          consts.KafkaPrivateKeyPemFile,
			consts.KafkaCa:                  consts.KafkaCACrtFile,
			consts.KafkaClientKeystoreP12:   consts.KafkaClientKeyStoreP12File,
			consts.KafkaClientTruststoreJks: consts.KafkaClientTruststoreJksFile,
		},
	}
}

type KafkaEnvironment struct {
	Envs          string
	Secret        *v1.Secret
	PrefixPath    string
	RequiredFiles map[string]string
}

func (k *KafkaEnvironment) Init() {
	k.Envs += fmt.Sprintf("# nais-cli %s\n# .env\n", time.Now().Truncate(time.Minute))
}

func (k *KafkaEnvironment) Finit() error {
	if err := k.write(); err != nil {
		return err
	}
	return nil
}

func (k *KafkaEnvironment) write() error {
	if err := common.WriteToFile(k.PrefixPath, KafkaSchemaRegistryEnvName, []byte(k.Envs)); err != nil {
		return fmt.Errorf("could not write envs to file: %s", err)
	}
	return nil
}

func (k *KafkaEnvironment) Set(key string, value []byte, destination string) {
	if destination == "" {
		k.Envs += fmt.Sprintf("%s=%s\n", key, string(value))
	} else {
		k.Envs += fmt.Sprintf("%s=%s\n", key, destination)
	}
}

func (k *KafkaEnvironment) Generate() error {
	err := common.RequiredSecretDataExists(k.RequiredFiles, k.Secret.Data, KafkaCatConfigName)
	if err != nil {
		return err
	}

	for key, value := range k.Secret.Data {
		if err := k.toFile(key, value); err != nil {
			return fmt.Errorf("could not write to file for key: %s\n %s", key, err)
		}
		k.toEnv(key, value)
	}
	return nil
}

func (k *KafkaEnvironment) toEnv(key string, value []byte) {
	if key == consts.KafkaBrokers {
		k.Set(key, value, "")
	}
	if key == consts.KafkaCredStorePassword {
		k.Set(key, value, "")
	}
	if strings.HasPrefix(key, consts.KafkaSchemaRegistry) {
		k.Set(key, value, "")
	}
}

func (k *KafkaEnvironment) toFile(key string, value []byte) error {
	path := k.PrefixPath
	requiredFile := k.RequiredFiles[key]
	if requiredFile != "" {
		if err := common.WriteToFile(path, requiredFile, value); err != nil {
			return err
		}
		k.Set(key, value, common.Destination(path, requiredFile))
	}
	return nil
}
