package config

import (
	b64 "encoding/base64"
	"fmt"
	"github.com/nais/debuk/pkg/common"
	"github.com/nais/debuk/pkg/consts"
	"io/ioutil"
	"strings"
)

const (
	FilePermission             = 0775
	KafkaSchemaRegistryEnvName = "kafka-secret.env"
)

type KafkaGeneralEnvironment struct {
	Envs string
}

func (k *KafkaGeneralEnvironment) Finit(destination string) error {
	if err := k.WriteConfig(common.Destination(destination, KafkaSchemaRegistryEnvName)); err != nil {
		return err
	}
	return nil
}

func (k *KafkaGeneralEnvironment) WriteConfig(dest string) error {
	if err := ioutil.WriteFile(dest, []byte(k.Envs), FilePermission); err != nil {
		return fmt.Errorf("could not write envs to file: %s", err)
	}
	return nil
}

func (k *KafkaGeneralEnvironment) Set(key, value, destination string) {
	if res, err := b64.StdEncoding.DecodeString(value); err == nil {
		if destination == "" {
			k.Envs += fmt.Sprintf("%s: %s\n", key, string(res))
		} else {
			k.Envs += fmt.Sprintf("%s=%s\n", key, destination)
		}
	}
}

func (k *KafkaGeneralEnvironment) Generate(key, value, dest string) error {
	switch key {
	case consts.KafkaCertificate:
		if err := common.WriteToFile(dest, consts.KafkaCertificateCrtFile, value); err != nil {
			return fmt.Errorf("could not write to file for key: %s\n %s", key, err)
		}
		k.Set(key, value, common.Destination(dest, consts.KafkaCertificateCrtFile))

	case consts.KafkaPrivateKey:
		if err := common.WriteToFile(dest, consts.KafkaPrivateKeyPemFile, value); err != nil {
			return fmt.Errorf("could not write to file for key: %s\n %s", key, err)
		}
		k.Set(key, value, common.Destination(dest, consts.KafkaPrivateKeyPemFile))

	case consts.KafkaCa:
		if err := common.WriteToFile(dest, consts.KafkaCACrtFile, value); err != nil {
			return fmt.Errorf("could not write to file for key: %s\n %s", key, err)
		}
		k.Set(key, value, common.Destination(dest, consts.KafkaCACrtFile))

	case consts.KafkaBrokers:
		k.Set(key, value, "")

	case consts.KafkaCredStorePassword:
		k.Set(key, value, "")

	case consts.KafkaClientKeystoreP12:
		if err := common.WriteToFile(dest, consts.KafkaClientKeyStoreP12File, value); err != nil {
			return fmt.Errorf("could not write to file for key: %s\n %s", k, err)
		}
		k.Set(key, value, common.Destination(dest, consts.KafkaClientKeyStoreP12File))

	case consts.KafkaClientTruststoreJks:
		if err := common.WriteToFile(dest, consts.KafkaClientTruststoreJksFile, value); err != nil {
			return fmt.Errorf("could not write to file for key: %s\n %s", k, err)
		}
		k.Set(key, value, common.Destination(dest, consts.KafkaClientTruststoreJksFile))
	}

	if strings.HasPrefix(key, consts.KafkaSchemaRegistry) {
		k.Set(key, value, "")
	}

	return nil
}
