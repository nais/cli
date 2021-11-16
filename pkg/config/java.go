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
	JavaConfigName = "kafka.properties"

	KeyPassProp            = "ssl.key.password"
	KeyStorePassProp       = "ssl.keystore.password"
	TrustStorePassProp     = "ssl.truststore.password"
	KeyStoreLocationProp   = "ssl.keystore.location"
	TrustStoreLocationProp = "ssl.truststore.location"

	FileHeader = `# Usage example: kafka-console-consumer.sh --topic %s.your.topic --bootstrap-server %s --consumer.config %s
security.protocol=SSL
ssl.protocol=TLS
ssl.keystore.type=PKCS12
ssl.truststore.type=JKS
`
)

func NewJavaConfig(secret *v1.Secret, dest string) Config {
	return &Java{
		Props:      fmt.Sprintf("# nais-cli %s\n", time.Now().Truncate(time.Minute)),
		Secret:     secret,
		PrefixPath: dest,
		RequiredFiles: map[string]RequiredFile{
			consts.KafkaClientKeyStoreP12File:   {consts.KafkaClientKeyStoreP12File, KeyStoreLocationProp, false},
			consts.KafkaClientTruststoreJksFile: {consts.KafkaClientTruststoreJksFile, TrustStoreLocationProp, false},
		},
	}
}

type Java struct {
	Props         string
	Secret        *v1.Secret
	PrefixPath    string
	RequiredFiles map[string]RequiredFile
}

func (k *Java) WriteConfigToFile() error {
	if err := k.write(); err != nil {
		return fmt.Errorf("could not write %s to file: %s", KafkaSchemaRegistryEnvName, err)
	}
	return nil
}

func (k *Java) write() error {
	if err := common.WriteToFile(k.PrefixPath, JavaConfigName, []byte(k.Props)); err != nil {
		return fmt.Errorf("write envs to file: %s", err)
	}
	return nil
}

func (k *Java) Set(key string, value []byte) {
	k.Props += fmt.Sprintf("%s=%s\n", key, string(value))
}

func (k *Java) SetPath(key, path string) {
	k.Props += fmt.Sprintf("%s=%s\n", key, path)
}

func (k *Java) Generate() (string, error) {
	err := requiredSecretDataExists(k.RequiredFiles, k.Secret.Data, JavaConfigName)
	if err != nil {
		return "", err
	}

	k.Props += fmt.Sprintf(FileHeader, k.Secret.Namespace, k.Secret.Data[consts.KafkaBrokersKey], filepath.Join(k.PrefixPath, JavaConfigName))

	for key, value := range k.Secret.Data {
		if err := k.toFile(key, value); err != nil {
			return "", fmt.Errorf("write to file for key: %s\n %s", key, err)
		}
		k.toEnv(key, value)
	}
	return k.Props, nil
}

func (k *Java) toEnv(key string, value []byte) {
	if key == consts.KafkaCredStorePasswordKey {
		k.Set(KeyPassProp, value)
		k.Set(KeyStorePassProp, value)
		k.Set(TrustStorePassProp, value)
	}
}

func (k *Java) toFile(key string, value []byte) error {
	path := k.PrefixPath
	if requiredFile, ok := k.RequiredFiles[key]; ok {
		if err := common.WriteToFile(path, requiredFile.Filename, value); err != nil {
			return err
		}
		k.SetPath(requiredFile.PathKey, filepath.Join(path, requiredFile.Filename))
	}
	return nil
}
