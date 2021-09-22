package consts

const (
	KafkaCertificateCrtFile      = "kafka-certificate.crt"
	KafkaPrivateKeyPemFile       = "kafka-private-key.pem"
	KafkaCACrtFile               = "kafka-ca.pem"
	KafkaClientKeyStoreP12File   = "client.keystore.p12"
	KafkaClientTruststoreJksFile = "client.truststore.jks"

	KafkaCertificatePathKey        = "KAFKA_CERTIFICATE_PATH"
	KafkaPrivateKeyPathKey         = "KAFKA_PRIVATE_KEY_PATH"
	KafkaCAPathKey                 = "KAFKA_CA_PATH"
	KafkaKeystorePathKey           = "KAFKA_KEYSTORE_PATH"
	KafkaTruststorePathKey         = "KAFKA_TRUSTSTORE_PATH"
	KafkaCertificateKey            = "KAFKA_CERTIFICATE"
	KafkaPrivateKeyKey             = "KAFKA_PRIVATE_KEY"
	KafkaCAKey                     = "KAFKA_CA"
	KafkaBrokersKey                = "KAFKA_BROKERS"
	KafkaSchemaRegistryKey         = "KAFKA_SCHEMA_REGISTRY"
	KafkaSchemaRegistryUserKey     = "KAFKA_SCHEMA_REGISTRY_USER"
	KafkaSchemaRegistryPasswordKey = "KAFKA_SCHEMA_REGISTRY_PASSWORD"
	KafkaCredStorePasswordKey      = "KAFKA_CREDSTORE_PASSWORD"
)
