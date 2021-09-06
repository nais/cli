package config

import (
	b64 "encoding/base64"
	"fmt"
	"io/ioutil"
	"time"
)

const (
	KafkaCatSslCertificateLocation = "ssl.certificate.location"
	KafkaCatSslKeyLocation         = "ssl.key.location"
	KafkaCatSslCaLocation          = "ssl.ca.location"
	KafkaCateKeyPassword           = "ssl.key.password"
	KafkaCatBootstrapServers       = "bootstrap.servers"

	KafkaCatConfigName = "kafkacat.config"
)

type KafkaCat struct {
	Config string
}

func (k *KafkaCat) Init() {
	k.Config += fmt.Sprintf("# Debuked %s\n# kafkacat -F kafkacat.config\n", time.Now().Truncate(time.Minute))
}

func (k *KafkaCat) Finit(destination string) error {
	k.Config += "security.protocol=ssl\n"
	if err := k.WriteConfig(Destination(destination, KafkaCatConfigName)); err != nil {
		return fmt.Errorf("could not write %s to file: %s", KafkaCatConfigName, err)
	}
	return nil
}

func (k *KafkaCat) WriteConfig(dest string) error {
	if err := ioutil.WriteFile(dest, []byte(k.Config), FilePermission); err != nil {
		return fmt.Errorf("could not write kafka.config to file: %s", err)
	}
	return nil
}

func (k *KafkaCat) Update(location, destination, value string) {
	if res, err := b64.StdEncoding.DecodeString(value); err == nil {
		if destination == "" {
			k.Config += fmt.Sprintf("%s=%s\n", location, string(res))
		} else {
			k.Config += fmt.Sprintf("%s=%s\n", location, destination)
		}
	}
}
