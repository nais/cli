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
```
