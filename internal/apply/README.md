# nais apply

Apply Valkey and OpenSearch resources to a Nais environment.

## Usage

```shell
nais alpha apply <config-file> --environment <env> --team <team>
```

## Arguments

- `config-file` — path to a YAML manifest file (`.yaml` or `.yml`)

## Flags

- `--environment` (`-e`) — target environment, e.g. `dev` or `prod` (required, tab-completion supported)
- `--team` (`-t`) — team slug (required)
- `--allow-ignored-fields` — warn instead of failing when a manifest contains fields that `nais apply` ignores (e.g. `metadata.namespace`, `metadata.annotations`)
- `--wait` — wait for applied resources to become ready before returning. Currently supported for `Application` resources; other kinds (Valkey, OpenSearch) are skipped
- `--timeout` — maximum time to wait for resources to become ready when `--wait` is set (default `10m`). Examples: `30s`, `5m`, `10m`

## Waiting for readiness

With `--wait`, after a successful apply `nais apply` polls the Nais API until each
wait-capable resource has converged to a healthy, steady state, and exits
non-zero if the `--timeout` is reached first (useful for failing CI jobs).

For an `Application`, "ready" means the rollout has converged to a single
instance group where all desired instances are ready and the application state is
`RUNNING`. Because the apply endpoint returns no rollout handle, correlation is
best-effort: `nais apply` waits for a new rollout (an instance group created
at/after the apply) to converge. If no new rollout appears shortly after a
no-op apply and the application is already healthy, it is reported as already up
to date.

## Manifest format

`nais apply` uses a stripped-down, nais-native manifest. It looks like a
Kubernetes CRD but hides cluster plumbing: there is no `apiVersion`, and
`metadata` only carries a `name`. Fields such as `metadata.namespace`,
`metadata.annotations`, owner references, generation and timestamps are not
part of the format — including them fails the apply unless
`--allow-ignored-fields` is passed.

The envelope is:

```yaml
version: v1          # currently always v1
kind: Valkey         # the resource kind
metadata:
  name: my-instance  # the only allowed metadata field
spec:
  ...
```

Resources that have a dedicated nais-api mutation (Valkey, OpenSearch) are
created or updated through that mutation. Other kinds are converted back into a
native CRD and sent to the generic apply endpoint.

Multiple resources can be placed in the same file separated by `---`, and the
native format may be mixed with regular Kubernetes CRDs (see below).

### Regular Kubernetes CRDs

Full Kubernetes CRDs (identified by `apiVersion`) are also accepted. They are
forwarded to the apply endpoint untouched — all metadata (namespace, labels,
annotations, ...) is preserved — and are **never** sent to a mutation. The
`--allow-ignored-fields` rule does not apply to them.

```yaml
apiVersion: nais.io/v1alpha1
kind: Application
metadata:
  name: testapp
  namespace: examples
  labels:
    team: examples
spec:
  image: europe-north1-docker.pkg.dev/nais-io/nais/images/testapp:40000
  replicas:
    min: 1
    max: 1
```

### Valkey

```yaml
version: v1
kind: Valkey
metadata:
  name: my-cache
spec:
  tier: SingleNode              # SingleNode | HighAvailability
  memory: "4GB"                 # 1GB | 4GB | 8GB | 14GB | 28GB | 56GB | 112GB | 200GB
  maxMemoryPolicy: allkeys-lru  # optional; see below
  databases: 16                 # optional
  notifyKeyspaceEvents: "Ex"    # optional
  persistence:                  # optional; parsed but not yet sent to the API
    disabled: false
```

**maxMemoryPolicy values:** `allkeys-lfu`, `allkeys-lru`, `allkeys-random`,
`noeviction`, `volatile-lfu`, `volatile-lru`, `volatile-random`, `volatile-ttl`

### OpenSearch

```yaml
version: v1
kind: OpenSearch
metadata:
  name: my-index
spec:
  tier: SingleNode          # SingleNode | HighAvailability
  memory: "4GB"             # 2GB | 4GB | 8GB | 16GB | 32GB | 64GB
  version: "2"              # "1" | "2" | "2.19" | "3.3"
  storageGB: 50
```

### Multi-resource file

```yaml
version: v1
kind: Valkey
metadata:
  name: sessions
spec:
  tier: HighAvailability
  memory: "28GB"
---
version: v1
kind: OpenSearch
metadata:
  name: search
spec:
  tier: SingleNode
  memory: "8GB"
  version: "2.19"
  storageGB: 100
```

## Examples

```shell
nais alpha apply nais.yaml --environment dev --team myteam
nais alpha apply nais.prod.yaml --environment prod --team myteam
```
