package aiven_config

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	KafkaJavaConfigName = "kafka.properties"

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
	properties += fmt.Sprintf(FileHeader, secret.Namespace, secret.Data[KafkaBrokersKey], filepath.Join(destinationPath, KafkaJavaConfigName))

	envsToFile := map[string]string{
		KeyPassProp:            string(secret.Data[KafkaCredStorePasswordKey]),
		KeyStorePassProp:       string(secret.Data[KafkaCredStorePasswordKey]),
		TrustStorePassProp:     string(secret.Data[KafkaCredStorePasswordKey]),
		KeyStoreLocationProp:   windowsify(filepath.Join(destinationPath, KafkaClientKeyStoreP12File)),
		TrustStoreLocationProp: windowsify(filepath.Join(destinationPath, KafkaClientTruststoreJksFile)),
	}

	for key, value := range envsToFile {
		properties += fmt.Sprintf("%s=%s\n", key, value)
	}

	err := os.WriteFile(filepath.Join(destinationPath, KafkaJavaConfigName), []byte(properties), FilePermission)
	if err != nil {
		return fmt.Errorf("write envs to file: %s", err)
	}

	secretsToFile := map[string][]byte{
		KafkaClientKeyStoreP12File:   secret.Data[KafkaClientKeyStoreP12File],
		KafkaClientTruststoreJksFile: secret.Data[KafkaClientTruststoreJksFile],
	}

	for fileName, value := range secretsToFile {
		err = os.WriteFile(filepath.Join(destinationPath, fileName), value, FilePermission)
		if err != nil {
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
