package config

import (
	b64 "encoding/base64"
	"fmt"
	"io/ioutil"
)

const (
	FilePermission = 0775

	KafkaSchemaRegistryEnvName = "kafka-schema-registry.env"
)

type KafkaGeneralEnvironment struct {
	Envs string
}

func (k *KafkaGeneralEnvironment) Finit(destination string) error {
	if err := k.WriteConfig(Destination(destination, KafkaSchemaRegistryEnvName)); err != nil {
		return err
	}
	return nil
}

func (k *KafkaGeneralEnvironment) Set(key, value string) {
	if res, err := b64.StdEncoding.DecodeString(value); err == nil {
		k.Envs += fmt.Sprintf("%s: %s\n", key, string(res))
	}
}

func (k *KafkaGeneralEnvironment) WriteConfig(dest string) error {
	if err := ioutil.WriteFile(dest, []byte(k.Envs), FilePermission); err != nil {
		return fmt.Errorf("could not write envs to file: %s", err)
	}
	return nil
}
