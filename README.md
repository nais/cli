# nais-cli

## Prerequisite

* Authentication & Authorization
    * Connect to [naisdevice](https://doc.nais.io/device/)
    * Tool is used in GCP? please be sure to log in:

```
gcloud auth login --update-adc
```

## Install

First

```
brew tap nais/tap
```

then

```
brew install nais  
```

check

```
nais version
```

You should be able to use

```
nais [commands] [args] [flags]
```

## Commands & Flags

Flags provide modifiers to control how the action command operates.

For help on individual commands, add `--help` short: `-h`.

Available commands:

- aiven
    - get
    - create
    - tidy
- version

### aiven

The aiven kafka debug command is used to create a `aivenApplication` and extract the credentials. The `avien` command
will apply
a [Protected & time-limited](https://doc.nais.io/persistence/kafka/#accessing-topics-from-an-application-on-legacy-infrastructure) `aivenApplication`
in your specified namespace.

This command will give access to personal but time limited credential. These credentials can be used to debug an Aiven
hosted kafka topic. The `aiven get` command extracts the fresh credentials and puts them in `tmp` folder. The
created `aivenApplication` has a default for `expireAt` (days-to-live) and will be set to 1 day.

To gain access be sure to update
your [topic](https://doc.nais.io/persistence/kafka/#creating-topics-and-defining-access) resource & ACLs, add `username`
to `topic`.yaml ACLs and apply to your namespace.

When secrets is extracted you can install and use [kcat](https://github.com/edenhill/kcat) (kcat is the project formerly
known as kafkacat) in preferred way.

#### Required

###### create

* `username` must be passed as **fist** argument after command: Prefix before `@nav.no`, replace `.` with `-`.

* `namespace` must be passed as **second** argument after command: team-namespace (default namespace not supported).

```
nais aiven your-username your-namespace
```

###### get

* `secret-name` must be passed as **fist** argument after command, Secret-name for your aiven application.

* `namespace` must be passed as **second** argument after command, team-namespace (default namespace not supported).

```
nais aiven get your-secret-name your-namespace
```

###### tidy

Remove secret-folders created by tool (or reboot your computer).

```
nais aiven tidy
```

#### Optional

##### create

* `--pool` short `-p` default: `nav-dev`: Preferred kafka pool.

* `--expire` short `-e` default: `1`: Time in days the created secret should be valid.

* `--secret-name` short `-s` default: `namespace-username-(random-id)`: Preferred secret-name instead of the generated.

###### get

* `--dest` short `-d` default: `tmp`: If other than default.

* `--config` short `-c`: default: `all`: Config type, `all || kcat || .env`. `all` generates both .env and kcat config
  files.

###### tidy

* `--root` short `-r` default: `/var/`: other than default `tmp` folder on your system.

### version

```
nais version
```

#### Optional

* `--commit` short `-i` default: `false` : Get detailed information about this `nais` version

## Flows & other details

Read more about what you get with a certain `command` and flow chart about the underlying systems for the various
commands used.

[aiven](doc/AIVEN_README.md)

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
