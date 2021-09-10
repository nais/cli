package secret

import (
	"fmt"
	"github.com/nais/debuk/pkg/common"
	"github.com/nais/debuk/pkg/config"
	"github.com/nais/debuk/pkg/consts"
	"strings"
)

type Secret struct {
	ApiVersion string            `yaml:"apiVersion"`
	Data       map[string]string `yaml:"data,omitempty"`
}

func (s Secret) ConfigAll(dest string) error {
	kCatConfig := config.KCat{}
	kCatConfig.Init()
	kafkaEnv := config.KafkaGeneralEnvironment{}
	for k, v := range s.Data {

		switch k {
		case consts.KafkaCertificate, consts.KafkaPrivateKey, consts.KafkaCa, consts.KafkaBrokers, consts.KafkaCredStorePassword:
			err := kCatConfig.KcatGenerate(k, v, dest)
			if err != nil {
				return err
			}

		case consts.KafkaClientKeystoreP12:
			if err := common.WriteToFile(dest, consts.KafkaClientKeyStoreP12File, v); err != nil {
				return fmt.Errorf("could not write to file for key: %s\n %s", k, err)
			}

		case consts.KafkaClientTruststoreJks:
			if err := common.WriteToFile(dest, consts.KafkaClientTruststoreJksFile, v); err != nil {
				return fmt.Errorf("could not write to file for key: %s\n %s", k, err)
			}
		}

		if strings.HasPrefix(k, consts.KafkaSchemaRegistry) {
			kafkaEnv.Set(k, v, "")
		}
	}

	if err := kCatConfig.Finit(dest); err != nil {
		return err
	}

	if err := kafkaEnv.Finit(dest); err != nil {
		return err
	}
	return nil
}

func (s Secret) Config(dest, typeConfig string) error {
	switch typeConfig {

	case consts.ENV:
		kafkaEnv := config.KafkaGeneralEnvironment{}
		for k, v := range s.Data {
			err := kafkaEnv.Generate(k, v, dest)
			if err != nil {
				return err
			}
		}
		if err := kafkaEnv.Finit(dest); err != nil {
			return err
		}
	case consts.KCAT:
		kCatConfig := config.KCat{}
		kCatConfig.Init()
		for k, v := range s.Data {
			err := kCatConfig.KcatGenerate(k, v, dest)
			if err != nil {
				return err
			}
		}
		if err := kCatConfig.Finit(dest); err != nil {
			return err
		}
	}
	fmt.Printf("Debuked! Files found here --> %s/*", dest)
	return nil
}

func (s Secret) generateAll(dest string) error {
	err := s.ConfigAll(dest)
	if err != nil {
		return fmt.Errorf("generate all configs: %s", err)
	}
	return nil
}
