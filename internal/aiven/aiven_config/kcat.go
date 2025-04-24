package aiven_config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	v1 "k8s.io/api/core/v1"
)

const (
	KafkaCatSslCertificateLocation = "ssl.certificate.location"
	KafkaCatSslKeyLocation         = "ssl.key.location"
	KafkaCatSslCaLocation          = "ssl.ca.location"
	KafkaCatBootstrapServers       = "bootstrap.servers"
	KafkaSecurityProtocolLocation  = "security.protocol"

	KafkaCatConfigName = "kcat.conf"
)

func WriteKCatConfigToFile(secret *v1.Secret, destinationPath string) error {
	configFile := fmt.Sprintf("# nais %s\n# kcat -F %s -t %s.your.topic\n", time.Now().Truncate(time.Minute), KafkaCatConfigName, secret.Namespace)
	envsToFile := map[string]string{
		KafkaCatBootstrapServers:       string(secret.Data[KafkaBrokersKey]),
		KafkaSecurityProtocolLocation:  "ssl",
		KafkaCatSslCertificateLocation: filepath.Join(destinationPath, KafkaCertificateCrtFile),
		KafkaCatSslKeyLocation:         filepath.Join(destinationPath, KafkaPrivateKeyPemFile),
		KafkaCatSslCaLocation:          filepath.Join(destinationPath, KafkaCACrtFile),
	}
	for key, value := range envsToFile {
		configFile += fmt.Sprintf("%s=%s\n", key, value)
	}

	err := os.WriteFile(filepath.Join(destinationPath, KafkaCatConfigName), []byte(configFile), FilePermission)
	if err != nil {
		return fmt.Errorf("write to file: %s", err)
	}

	secretsToFile := map[string]string{
		KafkaCertificateKey: KafkaCertificateCrtFile,
		KafkaPrivateKeyKey:  KafkaPrivateKeyPemFile,
		KafkaCAKey:          KafkaCACrtFile,
	}

	for key, fileName := range secretsToFile {
		err = os.WriteFile(filepath.Join(destinationPath, fileName), secret.Data[key], FilePermission)
		if err != nil {
			return err
		}
	}

	return nil
}
