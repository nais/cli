# nais-cli - A NAIS command-line interface

See NAIS doc for usage instructions: [nais-cli](https://docs.nais.io/how-to-guides/nais-cli/install/)

## Local Development

- Be sure to run your local cluster, recommend: [minkube](https://minikube.sigs.k8s.io/docs/start/).

Start minikube with a version < 1.22,
reason: [Feature removals](https://kubernetes.io/blog/2021/07/14/upcoming-changes-in-kubernetes-1-22/).

- Create a `test` cluster.

```
minikube start --kubernetes-version=v1.21.4
```

- Apply liberator CRDs.

```
kubectl apply -f path/to/liberator/crd/bases
```

- Create a `test` ns.

```
kubectl create namespace test
```

- Generate executable program and test your changes.

```
make nais-cli
bin/nais --version
```

## Instrumentation

We use [otel](https://opentelemetry.io) for instrumentation and record the user interaction with the CLI to see if we can 
optimize the user experience.

We respect your privacy and do not collect any data if you set the environment variable `DO_NOT_TRACK`
according to the [Do Not Track](https://consoledonottrack.com).

