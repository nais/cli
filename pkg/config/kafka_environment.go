package config

import (
	"fmt"
	"github.com/nais/cli/pkg/common"
	"github.com/nais/cli/pkg/consts"
	v1 "k8s.io/api/core/v1"
	"path/filepath"
	"strings"
	"time"
)

const (
	KafkaSchemaRegistryEnvName = "kafka-secret.env"
)

type RequiredFile struct {
	Filename     string
	PathKey      string
	IncludeInEnv bool
}

func NewEnvConfig(secret *v1.Secret, dest string) Config {
	return &KafkaEnvironment{
		Envs:       fmt.Sprintf("# nais-cli %s .env\n", time.Now().Truncate(time.Minute)),
		Secret:     secret,
		PrefixPath: dest,
		RequiredFiles: map[string]RequiredFile{
			consts.KafkaCertificateKey:          {consts.KafkaCertificateCrtFile, consts.KafkaCertificatePathKey, true},
			consts.KafkaPrivateKeyKey:           {consts.KafkaPrivateKeyPemFile, consts.KafkaPrivateKeyPathKey, true},
			consts.KafkaCAKey:                   {consts.KafkaCACrtFile, consts.KafkaCAPathKey, true},
			consts.KafkaClientKeyStoreP12File:   {consts.KafkaClientKeyStoreP12File, consts.KafkaKeystorePathKey, false},
			consts.KafkaClientTruststoreJksFile: {consts.KafkaClientTruststoreJksFile, consts.KafkaTruststorePathKey, false},
		},
	}
}

type KafkaEnvironment struct {
	Envs          string
	Secret        *v1.Secret
	PrefixPath    string
	RequiredFiles map[string]RequiredFile
}

func (k *KafkaEnvironment) WriteConfigToFile() error {
	if err := k.write(); err != nil {
		return fmt.Errorf("could not write %s to file: %s", KafkaSchemaRegistryEnvName, err)
	}
	return nil
}

func (k *KafkaEnvironment) write() error {
	if err := common.WriteToFile(k.PrefixPath, KafkaSchemaRegistryEnvName, []byte(k.Envs)); err != nil {
		return fmt.Errorf("write envs to file: %s", err)
	}
	return nil
}

func (k *KafkaEnvironment) Set(key string, value []byte) {
	k.Envs += fmt.Sprintf("%s=\"%s\"\n", key, string(value))
}

func (k *KafkaEnvironment) SetPath(key, path string) {
	k.Envs += fmt.Sprintf("%s=%s\n", key, path)
}

func (k *KafkaEnvironment) Generate() (string, error) {
	err := requiredSecretDataExists(k.RequiredFiles, k.Secret.Data, KafkaSchemaRegistryEnvName)
	if err != nil {
		return "", err
	}

	for key, value := range k.Secret.Data {
		if err := k.toFile(key, value); err != nil {
			return "", fmt.Errorf("write to file for key: %s\n %s", key, err)
		}
		k.toEnv(key, value)
	}
	return k.Envs, nil
}

func requiredSecretDataExists(required map[string]RequiredFile, secretData map[string][]byte, filetype string) error {
	for key, _ := range required {
		if _, ok := secretData[key]; !ok {
			return fmt.Errorf("can not generate %s config, secret missing required key: %s", filetype, key)
		}
	}
	return nil
}

func (k *KafkaEnvironment) toEnv(key string, value []byte) {
	if key == consts.KafkaBrokersKey {
		k.Set(key, value)
	}
	if key == consts.KafkaCredStorePasswordKey {
		k.Set(key, value)
	}
	if strings.HasPrefix(key, consts.KafkaSchemaRegistryKey) {
		k.Set(key, value)
	}
}

func (k *KafkaEnvironment) toFile(key string, value []byte) error {
	path := k.PrefixPath
	if requiredFile, ok := k.RequiredFiles[key]; ok {
		if err := common.WriteToFile(path, requiredFile.Filename, value); err != nil {
			return err
		}
		k.SetPath(requiredFile.PathKey, filepath.Join(path, requiredFile.Filename))
		if requiredFile.IncludeInEnv {
			k.Set(key, value)
		}
	}
	return nil
}
