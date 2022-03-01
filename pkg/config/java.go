package config

import (
	"fmt"
	"github.com/nais/cli/pkg/common"
	"github.com/nais/cli/pkg/consts"
	v1 "k8s.io/api/core/v1"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	JavaConfigName = "kafka.properties"

	KeyPassProp            = "ssl.key.password"
	KeyStorePassProp       = "ssl.keystore.password"
	TrustStorePassProp     = "ssl.truststore.password"
	KeyStoreLocationProp   = "ssl.keystore.location"
	TrustStoreLocationProp = "ssl.truststore.location"

	FileHeader = `# Usage example: kafka-console-consumer.sh --topic %s.your.topic --bootstrap-server %s --consumer.config %s
security.protocol=SSL
ssl.protocol=TLS
ssl.keystore.type=PKCS12
ssl.truststore.type=JKS
`
)

func NewJavaConfig(secret *v1.Secret, destinationPath string) error {
	properties := fmt.Sprintf("# nais-cli %s\n", time.Now().Truncate(time.Minute))
	properties += fmt.Sprintf(FileHeader, secret.Namespace, secret.Data[consts.KafkaBrokersKey], filepath.Join(destinationPath, JavaConfigName))

	envsToFile := map[string]string{
		KeyPassProp:            string(secret.Data[consts.KafkaCredStorePasswordKey]),
		KeyStorePassProp:       string(secret.Data[consts.KafkaCredStorePasswordKey]),
		TrustStorePassProp:     string(secret.Data[consts.KafkaCredStorePasswordKey]),
		KeyStoreLocationProp:   windowsify(filepath.Join(destinationPath, consts.KafkaClientKeyStoreP12File)),
		TrustStoreLocationProp: windowsify(filepath.Join(destinationPath, consts.KafkaClientTruststoreJksFile)),
	}

	for key, value := range envsToFile {
		properties += fmt.Sprintf("%s=%s\n", key, value)
	}

	if err := common.WriteToFile(destinationPath, JavaConfigName, []byte(properties)); err != nil {
		return fmt.Errorf("write envs to file: %s", err)
	}

	secretsToFile := map[string][]byte{
		consts.KafkaClientKeyStoreP12File:   secret.Data[consts.KafkaClientKeyStoreP12File],
		consts.KafkaClientTruststoreJksFile: secret.Data[consts.KafkaClientTruststoreJksFile],
	}
	for fileName, value := range secretsToFile {
		if err := common.WriteToFile(destinationPath, fileName, []byte(value)); err != nil {
			return fmt.Errorf("write to file: %s", err)
		}
	}

	return nil
}

func windowsify(path string) string {
	if runtime.GOOS == "windows" {
		return strings.ReplaceAll(path, "/", "\\")
	}
	return path
}
