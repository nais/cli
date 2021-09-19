package config

import (
	"fmt"
	"github.com/nais/nais-cli/pkg/common"
	"github.com/nais/nais-cli/pkg/consts"
	v1 "k8s.io/api/core/v1"
	"time"
)

const (
	KafkaCatSslCertificateLocation = "ssl.certificate.location"
	KafkaCatSslKeyLocation         = "ssl.key.location"
	KafkaCatSslCaLocation          = "ssl.ca.location"
	KafkaCatBootstrapServers       = "bootstrap.servers"

	KafkaCatConfigName = "kcat.conf"
)

func NewKCatConfig(secret *v1.Secret, dest string) Config {
	return &KCat{
		Config:     "",
		Secret:     secret,
		PrefixPath: dest,
		RequiredFiles: map[string]string{
			consts.KafkaCertificate: consts.KafkaCertificateCrtFile,
			consts.KafkaPrivateKey:  consts.KafkaPrivateKeyPemFile,
			consts.KafkaCa:          consts.KafkaCACrtFile,
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

func (k *KCat) Init() {
	k.Config += fmt.Sprintf("# nais %s\n# kcat -F %s -t %s.your.topic\n", time.Now().Truncate(time.Minute), KafkaCatConfigName, k.Secret.Namespace)
}

func (k *KCat) Finit() error {
	k.Config += "security.protocol=ssl\n"
	if err := k.write(); err != nil {
		return fmt.Errorf("could not write %s to file: %s", KafkaCatConfigName, err)
	}
	return nil
}

func (k *KCat) write() error {
	if err := common.WriteToFile(k.PrefixPath, KafkaCatConfigName, []byte(k.Config)); err != nil {
		return fmt.Errorf("could not write kafka.config to file: %s", err)
	}
	return nil
}

func (k *KCat) Set(key string, value []byte, destination string) {
	if destination == "" {
		k.Config += fmt.Sprintf("%s=%s\n", key, string(value))
	} else {
		k.Config += fmt.Sprintf("%s=%s\n", key, destination)
	}
}

func (k *KCat) Generate() error {
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

func (k *KCat) toFile(key string, value []byte) error {
	path := k.PrefixPath
	requiredFile := k.RequiredFiles[key]
	if requiredFile != "" {
		if err := common.WriteToFile(path, requiredFile, value); err != nil {
			return err
		}
		if k.RequiredLocation[requiredFile] != "" {
			k.Set(k.RequiredLocation[requiredFile], value, common.Destination(path, requiredFile))
		}
	}
	return nil
}

func (k *KCat) toEnv(key string, value []byte) {
	if key == consts.KafkaBrokers {
		k.Set(KafkaCatBootstrapServers, value, "")
	}
}
