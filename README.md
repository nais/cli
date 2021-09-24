# nais-cli

See NAIS doc for usage instructions: [nais-cli](https://doc.nais.io/cli/) 

## Local Development

* Be sure to run your local cluster, recommend: [minkube](https://minikube.sigs.k8s.io/docs/start/).

Start minikube with a version < 1.22,
reason: [Feature removals](https://kubernetes.io/blog/2021/07/14/upcoming-changes-in-kubernetes-1-22/).

```
minikube start --kubernetes-version=v1.21.4
```

* Apply liberator CRDs.

```
kubectl apply -f path/to/liberator/crd/bases
```

* Create a `test` cluster.

```
kubectl create namespace test
```

* Create a [secret](https://doc.nais.io/persistence/kafka/#application-config) containing this data.

```yaml
apiVersion: v1
data:
  KAFKA_BROKERS: ...
  KAFKA_CA: ...
  KAFKA_CERTIFICATE: ...
  KAFKA_CREDSTORE_PASSWORD: ...
  KAFKA_PRIVATE_KEY: ...
  KAFKA_SCHEMA_REGISTRY: ...
  KAFKA_SCHEMA_REGISTRY_PASSWORD: ...
  KAFKA_SCHEMA_REGISTRY_USER: ...
  client.keystore.p12: ...
  client.truststore.jks: ...
kind: Secret
metadata:
  annotations:
    aivenator.aiven.nais.io/protected: "true"
    aivenator.aiven.nais.io/with-time-limit: "true"
    kafka.aiven.nais.io/pool: nav-test
    kafka.aiven.nais.io/serviceUser: service-user
  finalizers:
    - aivenator.aiven.nais.io/finalizer
  labels:
    team: test
    type: aivenator.aiven.nais.io
  name: test-user
  namespace: test
type: Opaque
```

```
kubectl apply -f path/to/secret
```

* Generate executable program and test your changes.

```
make nais-cli
```
