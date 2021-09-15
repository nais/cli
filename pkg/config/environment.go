package config

import (
	"fmt"
	"github.com/nais/nais-d/pkg/common"
	"github.com/nais/nais-d/pkg/consts"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"strings"
	"time"
)

const (
	FilePermission             = 0775
	KafkaSchemaRegistryEnvName = "kafka-secret.env"
)

func NewEnvConfig(secret *v1.Secret, dest string) Config {
	return &KafkaEnvironment{
		Envs:       "",
		Secret:     secret,
		PrefixPath: dest,
	}
}

type KafkaEnvironment struct {
	Envs       string
	Secret     *v1.Secret
	PrefixPath string
}

func (k *KafkaEnvironment) Init() {
	k.Envs += fmt.Sprintf("# nais-d %s\n# .env\n", time.Now().Truncate(time.Minute))
}

func (k *KafkaEnvironment) Finit() error {
	if err := k.Write(); err != nil {
		return err
	}
	return nil
}

func (k *KafkaEnvironment) Write() error {
	if err := ioutil.WriteFile(common.Destination(k.PrefixPath, KafkaSchemaRegistryEnvName), []byte(k.Envs), FilePermission); err != nil {
		return fmt.Errorf("could not write envs to file: %s", err)
	}
	return nil
}

func (k *KafkaEnvironment) Set(key string, value []byte, destination string) {
	if destination == "" {
		k.Envs += fmt.Sprintf("%s: %s\n", key, string(value))
	} else {
		k.Envs += fmt.Sprintf("%s=%s\n", key, destination)
	}
}

func (k *KafkaEnvironment) Generate() error {
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
	switch key {
	case consts.KafkaCertificate:
		if err := common.WriteToFile(path, consts.KafkaCertificateCrtFile, value); err != nil {
			return err
		}
		k.Set(key, value, common.Destination(path, consts.KafkaCertificateCrtFile))

	case consts.KafkaPrivateKey:
		if err := common.WriteToFile(path, consts.KafkaPrivateKeyPemFile, value); err != nil {
			return err
		}
		k.Set(key, value, common.Destination(path, consts.KafkaPrivateKeyPemFile))

	case consts.KafkaCa:
		if err := common.WriteToFile(path, consts.KafkaCACrtFile, value); err != nil {
			return err
		}
		k.Set(key, value, common.Destination(path, consts.KafkaCACrtFile))

	case consts.KafkaClientKeystoreP12:
		if err := common.WriteToFile(path, consts.KafkaClientKeyStoreP12File, value); err != nil {
			return err
		}
		k.Set(key, value, common.Destination(path, consts.KafkaClientKeyStoreP12File))

	case consts.KafkaClientTruststoreJks:
		if err := common.WriteToFile(path, consts.KafkaClientTruststoreJksFile, value); err != nil {
			return err
		}
		k.Set(key, value, common.Destination(path, consts.KafkaClientTruststoreJksFile))
	}
	return nil
}
