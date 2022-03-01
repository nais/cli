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

func WriteKCatConfigToFile(secret *v1.Secret, destinationPath string) error {
	configFile := fmt.Sprintf("# nais %s\n# kcat -F %s -t %s.your.topic\n", time.Now().Truncate(time.Minute), KafkaCatConfigName, secret.Namespace)
	envsToFile := map[string]string{
		KafkaCatBootstrapServers:       string(secret.Data[consts.KafkaBrokersKey]),
		"ssl":                          KafkaSecurityProtocolLocation,
		KafkaCatSslCertificateLocation: filepath.Join(destinationPath, consts.KafkaCertificateCrtFile),
		KafkaCatSslKeyLocation:         filepath.Join(destinationPath, consts.KafkaPrivateKeyPemFile),
		KafkaCatSslCaLocation:          filepath.Join(destinationPath, consts.KafkaCACrtFile),
	}
	for key, value := range envsToFile {
		configFile += fmt.Sprintf("%s=%s\n", key, value)

	}

	if err := common.WriteToFile(destinationPath, KafkaCatConfigName, []byte(configFile)); err != nil {
		return fmt.Errorf("write to file: %s", err)
	}

	secretsToFile := map[string]string{
		consts.KafkaCertificateKey: consts.KafkaCertificateCrtFile,
		consts.KafkaPrivateKeyKey:  consts.KafkaPrivateKeyPemFile,
		consts.KafkaCAKey:          consts.KafkaCACrtFile,
	}

	for fileName, valueKey := range secretsToFile {
		if err := common.WriteToFile(destinationPath, fileName, secret.Data[valueKey]); err != nil {
			return err
		}
	}

	return nil
}
