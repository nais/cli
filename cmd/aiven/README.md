# aiven

## Available output

After Successful `nais aiven create` and `nais aiven get` commands, a set of files wil be available.

### Configuration

You can specify a configuration `flag` to generate `all | kcat | .env`. Default is `all`

#### all

- client-keystore.p12
- client-truststore.jks
- kafka-ca.pem
- kafka-client-certificate.crt
- kafka-client-private-key.pem
- kafka-secret.env
- kcat.conf

#### .env

- client-keystore.p12
- client-truststore.jks
- kafka-ca.pem
- kafka-client-certificate.crt
- kafka-client-private-key.pem
- kafka-secret.env

##### kafka-secret.env file

```Properties
KAFKA_BROKERS=brokerurl.aivencloud.com:26484
KAFKA_PRIVATE_KEY=/path/to/kafka-client-private-key.pem
client.keystore.p12=/path/to/client-keystore.p12
client.truststore.jks=/path/to/.envs/client-truststore.jks
KAFKA_CA=/path/to/.envs/kafka-ca.pem
KAFKA_CERTIFICATE=/path/to/.envs/kafka-client-certificate.crt
KAFKA_CREDSTORE_PASSWORD=password
KAFKA_SCHEMA_REGISTRY=https://registry-url.aivencloud.com:26487
KAFKA_SCHEMA_REGISTRY_PASSWORD=password
.....
```

#### kcat

- kafka-ca.pem
- kafka-client-certificate.crt
- kafka-client-private-key.pem
- kcat.conf

##### kcat.conf file

```Properties
# nais 2021-09-01 15:26:00
# kcat -F kcat.conf -t namespace.your.topic
ssl.key.location=/path/to/tmp/folder/creds/kafka-client-private-key.pem
ssl.certificate.location=/path/to/tmp/folder/creds/kafka-client-certificate.crt
bootstrap.servers=https://boostrap-server.aivencloud.com:26484
ssl.ca.location=/path/to/tmp/folder/creds/kafka-ca.pem
security.protocol=ssl
....
```

The generated `kcat.conf` can be used with [kcat](https://github.com/edenhill/kcat) to authenticate against the Aiven
hosted topics in GCP.

Read more about [kcat.conf](https://github.com/edenhill/librdkafka/blob/master/CONFIGURATION.md) configurable
properties.

You can refer to generated config with -F flag:

```
kcat -F path/to/kcat.conf -t namespace.your.topic
```

Alternatively, you can specify the same settings directly on the command line:

```
kcat \
    -b https://boostrap-server.aivencloud.com:26484 \
    -X security.protocol=ssl \
    -X ssl.key.location=service.key \
    -X ssl.certificate.location=service.cert \
    -X ssl.ca.location=ca.pem
```

For more details [aiven-kcat](https://help.aiven.io/en/articles/2607674-using-kafkacat)

# Flow

![aiven command under the hood](../../doc/nais-cli.png)