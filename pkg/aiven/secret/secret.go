package secret

import (
	b64 "encoding/base64"
	"fmt"
	"github.com/nais/debuk/pkg/aiven/application"
	"github.com/nais/debuk/pkg/config"
	"io/ioutil"
	"strings"
)

const (
	KafkaCertificateCrtFile      = "kafka-certificate.crt"
	KafkaPrivateKeyPemFile       = "kafka-private-key.pem"
	KafkaCACrtFile               = "kafka-ca.cert"
	KafkaClientKeyStoreP12File   = "client-keystore.p12"
	KafkaClientTruststoreJksFile = "client-truststore.jks"

	KafkaCertificate         = "KAFKA_CERTIFICATE"
	KafkaPrivateKey          = "KAFKA_PRIVATE_KEY"
	KafkaCa                  = "KAFKA_CA"
	KafkaBrokers             = "KAFKA_BROKERS"
	KafkaCredStorePassword   = "KAFKA_CREDSTORE_PASSWORD"
	KafkaSchemaRegistry      = "KAFKA_SCHEMA_REGISTRY"
	KafkaClientKeystoreP12   = "client.keystore.p12"
	KafkaClientTruststoreJks = "client.truststore.jks"
)

type Secret struct {
	ApiVersion string            `yaml:"apiVersion"`
	Data       map[string]string `yaml:"data,omitempty"`
}

func (s Secret) GenerateConfiguration(dest, username string) error {
	kafkaCatConfig := config.KafkaCat{}
	kafkaCatConfig.Init()
	kafkaEnv := config.KafkaGeneralEnvironment{}
	for k, v := range s.Data {

		switch k {
		case KafkaCertificate:
			if err := writeToFile(dest, KafkaCertificateCrtFile, v); err != nil {
				return fmt.Errorf("could not write to file for key: %s\n %s", k, err)
			}
			kafkaCatConfig.Update(config.KafkaCatSslCertificateLocation, config.Destination(dest, KafkaCertificateCrtFile), v)

		case KafkaPrivateKey:
			if err := writeToFile(dest, KafkaPrivateKeyPemFile, v); err != nil {
				return fmt.Errorf("could not write to file for key: %s\n %s", k, err)
			}
			kafkaCatConfig.Update(config.KafkaCatSslKeyLocation, config.Destination(dest, KafkaPrivateKeyPemFile), v)

		case KafkaCa:
			if err := writeToFile(dest, KafkaCACrtFile, v); err != nil {
				return fmt.Errorf("could not write to file for key: %s\n %s", k, err)
			}
			kafkaCatConfig.Update(config.KafkaCatSslCaLocation, config.Destination(dest, KafkaCACrtFile), v)

		case KafkaClientKeystoreP12:
			if err := writeToFile(dest, KafkaClientKeyStoreP12File, v); err != nil {
				return fmt.Errorf("could not write to file for key: %s\n %s", k, err)
			}

		case KafkaClientTruststoreJks:
			if err := writeToFile(dest, KafkaClientTruststoreJksFile, v); err != nil {
				return fmt.Errorf("could not write to file for key: %s\n %s", k, err)
			}

		case KafkaBrokers:
			kafkaCatConfig.Update(config.KafkaCatBootstrapServers, "", v)

		case KafkaCredStorePassword:
			kafkaCatConfig.Update(config.KafkaCateKeyPassword, "", v)
		}

		if strings.HasPrefix(k, KafkaSchemaRegistry) {
			kafkaEnv.Set(k, v)
		}
	}

	if err := kafkaCatConfig.Finit(dest); err != nil {
		return err
	}

	if err := kafkaEnv.Finit(dest); err != nil {
		return err
	}
	return nil
}

func writeToFile(dest, filename, value string) error {
	if res, err := b64.StdEncoding.DecodeString(value); err == nil {
		err = ioutil.WriteFile(config.Destination(dest, filename), res, application.FilePermission)
		if err != nil {
			return err
		}
	}
	return nil
}
