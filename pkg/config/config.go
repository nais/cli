package config

import "github.com/nais/nais-cli/pkg/consts"

type Config interface {
	Finit() error
	Init()
	Set(key string, value []byte, destination string)
	Generate() (string, error)
}

const (
	ENV  = ".env"
	KCAT = "kcat"
	ALL  = "all"
)

var KCatEnvToFileMap = map[string]string{
	consts.KafkaCertificate: consts.KafkaCertificateCrtFile,
	consts.KafkaPrivateKey:  consts.KafkaPrivateKeyPemFile,
	consts.KafkaCa:          consts.KafkaCACrtFile,
}

var KafkaConfigEnvToFileMap = map[string]string{
	consts.KafkaCertificate:         consts.KafkaCertificateCrtFile,
	consts.KafkaPrivateKey:          consts.KafkaPrivateKeyPemFile,
	consts.KafkaCa:                  consts.KafkaCACrtFile,
	consts.KafkaClientKeystoreP12:   consts.KafkaClientKeyStoreP12File,
	consts.KafkaClientTruststoreJks: consts.KafkaClientTruststoreJksFile,
}
