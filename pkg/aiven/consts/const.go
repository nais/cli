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
	KafkaSchemaRegistryPasswordKey = "KAFKA_SCHEMA_REGISTRY_PASSWORD"
	KafkaSchemaRegistryUserKey     = "KAFKA_SCHEMA_REGISTRY_USER"
	KafkaCredStorePasswordKey      = "KAFKA_CREDSTORE_PASSWORD"

	OpenSearchURIKey      = "OPEN_SEARCH_URI"
	OpenSearchUsernameKey = "OPEN_SEARCH_USERNAME"
	OpenSearchPasswordKey = "OPEN_SEARCH_PASSWORD"
)
