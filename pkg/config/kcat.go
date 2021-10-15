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
	KafkaCatSslCertificateLocation = "ssl.certificate.location"
	KafkaCatSslKeyLocation         = "ssl.key.location"
	KafkaCatSslCaLocation          = "ssl.ca.location"
	KafkaCatBootstrapServers       = "bootstrap.servers"
	KafkaSecurityProtocolLocation  = "security.protocol"

	KafkaCatConfigName = "kcat.conf"
)

func NewKCatConfig(secret *v1.Secret, dest string) Config {
	return &KCat{
		Config:     fmt.Sprintf("# nais %s\n# kcat -F %s -t %s.your.topic\n", time.Now().Truncate(time.Minute), KafkaCatConfigName, secret.Namespace),
		Secret:     secret,
		PrefixPath: dest,
		RequiredFiles: map[string]string{
			consts.KafkaCertificateKey: consts.KafkaCertificateCrtFile,
			consts.KafkaPrivateKeyKey:  consts.KafkaPrivateKeyPemFile,
			consts.KafkaCAKey:          consts.KafkaCACrtFile,
		},
		RequiredLocation: map[string]string{
			consts.KafkaCertificateCrtFile: KafkaCatSslCertificateLocation,
			consts.KafkaPrivateKeyPemFile:  KafkaCatSslKeyLocation,
			consts.KafkaCACrtFile:          KafkaCatSslCaLocation,
		},
	}
}

type KCat struct {
	Config           string
	Secret           *v1.Secret
	PrefixPath       string
	RequiredFiles    map[string]string
	RequiredLocation map[string]string
}

func (k *KCat) WriteConfigToFile() error {
	if err := k.write(); err != nil {
		return fmt.Errorf("write %s to file: %s", KafkaCatConfigName, err)
	}
	return nil
}

func (k *KCat) write() error {
	if err := common.WriteToFile(k.PrefixPath, KafkaCatConfigName, []byte(k.Config)); err != nil {
		return fmt.Errorf("write to file: %s", err)
	}
	return nil
}

func (k *KCat) Set(key string, value []byte) {
	k.Config += fmt.Sprintf("%s=%s\n", key, string(value))
}

func (k *KCat) SetPath(key, path string) {
	k.Config += fmt.Sprintf("%s=%s\n", key, path)
}

func (k *KCat) Generate() (string, error) {
	err := common.RequiredSecretDataExists(k.RequiredFiles, k.Secret.Data, KafkaCatConfigName)
	if err != nil {
		return "", err
	}

	for key, value := range k.Secret.Data {
		if err := k.toFile(key, value); err != nil {
			return "", fmt.Errorf("write to file for key: %s\n %s", key, err)
		}
		k.toEnv(key, value)
	}
	k.Config += fmt.Sprintf("%s=ssl\n", KafkaSecurityProtocolLocation)
	return k.Config, nil
}

func (k *KCat) toFile(key string, value []byte) error {
	path := k.PrefixPath
	requiredFile := k.RequiredFiles[key]
	if requiredFile != "" {
		if err := common.WriteToFile(path, requiredFile, value); err != nil {
			return err
		}
		if k.RequiredLocation[requiredFile] != "" {
			k.SetPath(k.RequiredLocation[requiredFile], filepath.Join(path, requiredFile))
		}
	}
	return nil
}

func (k *KCat) toEnv(key string, value []byte) {
	if key == consts.KafkaBrokersKey {
		k.Set(KafkaCatBootstrapServers, value)
	}
}
