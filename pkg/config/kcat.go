package config

import (
	"fmt"
	"github.com/nais/nais-cli/pkg/common"
	"github.com/nais/nais-cli/pkg/consts"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"time"
)

const (
	KafkaCatSslCertificateLocation = "ssl.certificate.location"
	KafkaCatSslKeyLocation         = "ssl.key.location"
	KafkaCatSslCaLocation          = "ssl.ca.location"
	KafkaCateKeyPassword           = "ssl.key.password"
	KafkaCatBootstrapServers       = "bootstrap.servers"

	KafkaCatConfigName = "kcat.conf"
)

func NewKCatConfig(secret *v1.Secret, dest string) Config {
	return &KCat{
		Config:     "",
		Secret:     secret,
		PrefixPath: dest,
	}
}

type KCat struct {
	Config     string
	Secret     *v1.Secret
	PrefixPath string
}

func (k *KCat) Init() {
	k.Config += fmt.Sprintf("# nais %s\n# kcat -F %s -t %s-your.topic\n", time.Now().Truncate(time.Minute), KafkaCatConfigName, k.Secret.Namespace)
}

func (k *KCat) Finit() error {
	k.Config += "security.protocol=ssl\n"
	if err := k.Write(); err != nil {
		return fmt.Errorf("could not write %s to file: %s", KafkaCatConfigName, err)
	}
	return nil
}

func (k *KCat) Write() error {
	if err := ioutil.WriteFile(common.Destination(k.PrefixPath, KafkaCatConfigName), []byte(k.Config), FilePermission); err != nil {
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
	switch key {
	case consts.KafkaCertificate:
		if err := common.WriteToFile(path, consts.KafkaCertificateCrtFile, value); err != nil {
			return err
		}
		k.Set(KafkaCatSslCertificateLocation, value, common.Destination(path, consts.KafkaCertificateCrtFile))

	case consts.KafkaPrivateKey:
		if err := common.WriteToFile(path, consts.KafkaPrivateKeyPemFile, value); err != nil {
			return err
		}
		k.Set(KafkaCatSslKeyLocation, value, common.Destination(path, consts.KafkaPrivateKeyPemFile))

	case consts.KafkaCa:
		if err := common.WriteToFile(path, consts.KafkaCACrtFile, value); err != nil {
			return err
		}
		k.Set(KafkaCatSslCaLocation, value, common.Destination(path, consts.KafkaCACrtFile))
	}
	return nil
}

func (k *KCat) toEnv(key string, value []byte) {
	if key == consts.KafkaBrokers {
		k.Set(KafkaCatBootstrapServers, value, "")
	}
}
