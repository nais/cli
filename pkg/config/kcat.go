package config

import (
	"fmt"
	"github.com/nais/nais-d/pkg/common"
	"github.com/nais/nais-d/pkg/consts"
	"io/ioutil"
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

type KCat struct {
	Config string
}

func (k *KCat) Init() {
	k.Config += fmt.Sprintf("# nais-d %s\n# kcat -F %s\n", time.Now().Truncate(time.Minute), KafkaCatConfigName)
}

func (k *KCat) Finit(destination string) error {
	k.Config += "security.protocol=ssl\n"
	if err := k.WriteConfig(common.Destination(destination, KafkaCatConfigName)); err != nil {
		return fmt.Errorf("could not write %s to file: %s", KafkaCatConfigName, err)
	}
	return nil
}

func (k *KCat) WriteConfig(dest string) error {
	if err := ioutil.WriteFile(dest, []byte(k.Config), FilePermission); err != nil {
		return fmt.Errorf("could not write kafka.config to file: %s", err)
	}
	return nil
}

func (k *KCat) Update(key string, value []byte, destination string) {
	if destination == "" {
		k.Config += fmt.Sprintf("%s=%s\n", key, string(value))
	} else {
		k.Config += fmt.Sprintf("%s=%s\n", key, destination)
	}
}

func (k *KCat) KcatGenerate(key string, value []byte, dest string) error {
	switch key {
	case consts.KafkaCertificate:
		if err := common.WriteToFile(dest, consts.KafkaCertificateCrtFile, value); err != nil {
			return fmt.Errorf("could not write to file for key: %s\n %s", key, err)
		}
		k.Update(KafkaCatSslCertificateLocation, value, common.Destination(dest, consts.KafkaCertificateCrtFile))

	case consts.KafkaPrivateKey:
		if err := common.WriteToFile(dest, consts.KafkaPrivateKeyPemFile, value); err != nil {
			return fmt.Errorf("could not write to file for key: %s\n %s", key, err)
		}
		k.Update(KafkaCatSslKeyLocation, value, common.Destination(dest, consts.KafkaPrivateKeyPemFile))

	case consts.KafkaCa:
		if err := common.WriteToFile(dest, consts.KafkaCACrtFile, value); err != nil {
			return fmt.Errorf("could not write to file for key: %s\n %s", key, err)
		}
		k.Update(KafkaCatSslCaLocation, value, common.Destination(dest, consts.KafkaCACrtFile))

	case consts.KafkaBrokers:
		k.Update(KafkaCatBootstrapServers, value, "")

	case consts.KafkaCredStorePassword:
		k.Update(KafkaCateKeyPassword, value, "")
	}
	return nil
}
